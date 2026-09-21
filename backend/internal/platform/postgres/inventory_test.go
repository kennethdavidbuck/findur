package postgres_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestInventoryRepositoryClaimsPublishesAndIsolatesOwners(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	attempt := fixture.createAttempt(t, 91, fixture.now.Add(10*60*1e9))
	fixture.claimCallback(t, attempt)
	owner := uuid.New()
	final := finalization(attempt.StateHash, owner, fixture.now, 92)
	if err := fixture.repository.FinalizeCallback(fixture.ctx, final); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewInventoryRepository(fixture.pool)

	claims := concurrentPrepare(t, repository, fixture, owner, false)
	if claims != 1 {
		t.Fatalf("bootstrap claims = %d, want 1", claims)
	}
	connections := []portfolio.Connection{{
		ID: "connection-1", BrokerageLabel: "Synthetic Broker", Status: "active", SyncMode: "delayed", Available: true, Eligible: true,
		Accounts: []portfolio.Account{{ID: "account-1", Category: "investment", Type: "Margin", MaskedLabel: "Retirement (•••• 8443)", Available: true, Eligible: true, SyncState: "complete"}},
	}}
	snapshot, published, err := repository.Finalize(fixture.ctx, owner, 1, portfolio.StateReady, nil, connections, fixture.now)
	if err != nil || !published || snapshot.State != portfolio.StateReady || len(snapshot.Connections) != 1 || len(snapshot.Connections[0].Accounts) != 1 {
		t.Fatalf("snapshot=%+v published=%v err=%v", snapshot, published, err)
	}
	revisit, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || revisit.Claimed || revisit.State != portfolio.StateReady || len(revisit.Connections) != 1 {
		t.Fatalf("revisit=%+v err=%v", revisit, err)
	}

	if claims := concurrentPrepare(t, repository, fixture, owner, true); claims != 1 {
		t.Fatalf("retry claims = %d, want 1", claims)
	}
	if _, published, err := repository.Finalize(fixture.ctx, owner, 1, portfolio.StateEmpty, nil, nil, fixture.now); err != nil || published {
		t.Fatalf("stale completion published=%v err=%v", published, err)
	}
	failed, published, err := repository.Finalize(fixture.ctx, owner, 2, portfolio.StateUnavailable, nil, nil, fixture.now)
	if err != nil || !published || failed.State != portfolio.StateUnavailable || len(failed.Connections) != 1 {
		t.Fatalf("failed snapshot=%+v published=%v err=%v", failed, published, err)
	}

	third, err := repository.Prepare(fixture.ctx, owner, true, fixture.now)
	if err != nil || !third.Claimed || third.Generation != 3 {
		t.Fatalf("third=%+v err=%v", third, err)
	}
	bad := []portfolio.Connection{{ID: "bad", BrokerageLabel: "Broker", Status: "invented", SyncMode: "unknown"}}
	if _, _, err := repository.Finalize(fixture.ctx, owner, 3, portfolio.StateReady, nil, bad, fixture.now); err == nil {
		t.Fatal("invalid publication did not roll back")
	}
	pending, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || pending.State != portfolio.StatePending || len(pending.Connections) != 1 {
		t.Fatalf("rolled-back state=%+v err=%v", pending, err)
	}

	other := uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO users (id,origin) VALUES ($1,'oauth')`, other); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO provider_authorizations
		(user_id,provider,access_token_encrypted,envelope_version) VALUES ($1,$2,$3,1)`, other, auth.SnapTradeProvider, bytes.Repeat([]byte{3}, 24)); err != nil {
		t.Fatal(err)
	}
	otherState, err := repository.Prepare(fixture.ctx, other, false, fixture.now)
	if err != nil || !otherState.Claimed || otherState.Generation != 1 || len(otherState.Connections) != 0 {
		t.Fatalf("other owner state=%+v err=%v", otherState, err)
	}
	retryAt := fixture.now.Add(time.Minute)
	rateLimited, published, err := repository.Finalize(fixture.ctx, other, 1, portfolio.StateRateLimited, &retryAt, nil, fixture.now)
	if err != nil || !published || rateLimited.RetryAt == nil {
		t.Fatalf("rate limited=%+v published=%v err=%v", rateLimited, published, err)
	}
	earlyRetry, err := repository.Prepare(fixture.ctx, other, true, fixture.now.Add(30*time.Second))
	if err != nil || earlyRetry.Claimed || earlyRetry.Generation != 1 {
		t.Fatalf("early retry=%+v err=%v", earlyRetry, err)
	}
	safeRetry, err := repository.Prepare(fixture.ctx, other, true, retryAt)
	if err != nil || !safeRetry.Claimed || safeRetry.Generation != 2 {
		t.Fatalf("safe retry=%+v err=%v", safeRetry, err)
	}

	var inclusionColumns int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM information_schema.columns
		WHERE table_name LIKE 'portfolio_inventory_%' AND column_name LIKE '%' || 'inclu' || 'd%'`).Scan(&inclusionColumns); err != nil {
		t.Fatal(err)
	}
	if inclusionColumns != 0 {
		t.Fatalf("inventory schema contains %d committed-inclusion columns", inclusionColumns)
	}
}

func TestInventoryRepositoryReclaimsExpiredClaimAfterHardRestart(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	attempt := fixture.createAttempt(t, 91, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, attempt)
	owner := uuid.New()
	if err := fixture.repository.FinalizeCallback(fixture.ctx, finalization(attempt.StateHash, owner, fixture.now, 92)); err != nil {
		t.Fatal(err)
	}

	firstProcess := postgresadapter.NewInventoryRepository(fixture.pool)
	first, err := firstProcess.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !first.Claimed || first.Generation != 1 || first.ClaimExpiresAt == nil {
		t.Fatalf("first claim=%+v err=%v", first, err)
	}

	secondProcess := postgresadapter.NewInventoryRepository(fixture.pool)
	withinLease, err := secondProcess.Prepare(fixture.ctx, owner, false, first.ClaimExpiresAt.Add(-time.Nanosecond))
	if err != nil || withinLease.Claimed || withinLease.Generation != 1 {
		t.Fatalf("within lease=%+v err=%v", withinLease, err)
	}
	if claims := concurrentPrepareAt(t, secondProcess, fixture, owner, false, *first.ClaimExpiresAt); claims != 1 {
		t.Fatalf("expired claim reclaims = %d, want 1", claims)
	}
	if _, published, err := firstProcess.Finalize(fixture.ctx, owner, 1, portfolio.StateEmpty, nil, nil, *first.ClaimExpiresAt); err != nil || published {
		t.Fatalf("stale process completion published=%v err=%v", published, err)
	}
	current, published, err := secondProcess.Finalize(fixture.ctx, owner, 2, portfolio.StateEmpty, nil, nil, *first.ClaimExpiresAt)
	if err != nil || !published || current.State != portfolio.StateEmpty || current.Generation != 2 {
		t.Fatalf("reclaimed completion=%+v published=%v err=%v", current, published, err)
	}
}

func TestInventoryRepositoryPublishesDisabledRowsAndReauthorizationInvalidatesHead(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	firstAttempt := fixture.createAttempt(t, 101, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, firstAttempt)
	owner := uuid.New()
	first := finalization(firstAttempt.StateHash, owner, fixture.now, 102)
	if err := fixture.repository.FinalizeCallback(fixture.ctx, first); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	claim, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("claim=%+v err=%v", claim, err)
	}
	disabledRows := []portfolio.Connection{{ID: "disabled-connection", BrokerageLabel: "Synthetic Broker", Status: "disabled", SyncMode: "unknown"}}
	disabled, published, err := repository.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateDisabled, nil, disabledRows, fixture.now)
	if err != nil || !published || disabled.State != portfolio.StateDisabled || len(disabled.Connections) != 1 {
		t.Fatalf("disabled=%+v published=%v err=%v", disabled, published, err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inclusion_state (user_id,version,lifecycle_generation,updated_at) VALUES ($1,0,7,$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}

	secondAttempt := fixture.createAttempt(t, 103, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, secondAttempt)
	second := finalization(secondAttempt.StateHash, owner, fixture.now.Add(time.Minute), 104)
	second.Subject = first.Subject
	if err := fixture.repository.FinalizeCallback(fixture.ctx, second); err != nil {
		t.Fatal(err)
	}
	var inclusionLifecycle int64
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1`, owner).Scan(&inclusionLifecycle); err != nil {
		t.Fatal(err)
	}
	if inclusionLifecycle != 8 {
		t.Fatalf("reauthorization inclusion lifecycle=%d, want 8", inclusionLifecycle)
	}
	if stale, published, err := repository.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, disabledRows, fixture.now.Add(time.Minute)); err != nil || published || stale.Generation <= claim.Generation {
		t.Fatalf("old bearer completion snapshot=%+v published=%v err=%v", stale, published, err)
	}
	bootstrap, err := repository.Prepare(fixture.ctx, owner, false, fixture.now.Add(time.Minute))
	if err != nil || !bootstrap.Claimed || bootstrap.Generation != claim.Generation+2 || len(bootstrap.Connections) != 0 || !bytes.Equal(bootstrap.EncryptedToken, second.AccessToken) {
		t.Fatalf("bootstrap=%+v err=%v", bootstrap, err)
	}
	if current, published, err := repository.Finalize(fixture.ctx, owner, bootstrap.Generation, portfolio.StateEmpty, nil, nil, fixture.now.Add(time.Minute)); err != nil || !published || current.State != portfolio.StateEmpty {
		t.Fatalf("new bearer completion snapshot=%+v published=%v err=%v", current, published, err)
	}
	var versions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_versions WHERE user_id=$1`, owner).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if versions != 1 {
		t.Fatalf("inventory versions=%d, want only the new bearer version", versions)
	}
}

func TestInventoryRepositoryPersistsPartialFailureRowsAndRetryTiming(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	for index, state := range []portfolio.State{portfolio.StateRateLimited, portfolio.StateMalformed, portfolio.StateUnavailable} {
		owner := uuid.New()
		if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO users (id,origin) VALUES ($1,'oauth')`, owner); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO provider_authorizations
			(user_id,provider,access_token_encrypted,envelope_version) VALUES ($1,$2,$3,1)`, owner, auth.SnapTradeProvider, bytes.Repeat([]byte{byte(index + 1)}, 24)); err != nil {
			t.Fatal(err)
		}
		claim, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
		if err != nil || !claim.Claimed {
			t.Fatalf("state=%s claim=%+v err=%v", state, claim, err)
		}
		rows := []portfolio.Connection{{ID: "partial-connection", BrokerageLabel: "Synthetic Broker", Status: portfolio.ConnectionStatusUnavailable, SyncMode: portfolio.SyncModeDelayed}}
		var retryAt *time.Time
		if state == portfolio.StateRateLimited {
			value := fixture.now.Add(time.Minute)
			retryAt = &value
		}
		published, accepted, err := repository.Finalize(fixture.ctx, owner, claim.Generation, state, retryAt, rows, fixture.now)
		if err != nil || !accepted || published.State != state || len(published.Connections) != 1 {
			t.Fatalf("state=%s snapshot=%+v accepted=%v err=%v", state, published, accepted, err)
		}
		revisit, err := repository.Prepare(fixture.ctx, owner, false, fixture.now.Add(30*time.Second))
		if err != nil || revisit.Claimed || revisit.State != state || len(revisit.Connections) != 1 {
			t.Fatalf("state=%s revisit=%+v err=%v", state, revisit, err)
		}
		if state == portfolio.StateRateLimited {
			if revisit.RetryAt == nil || !revisit.RetryAt.Equal(*retryAt) {
				t.Fatalf("rate limited retryAt=%v want=%v", revisit.RetryAt, retryAt)
			}
			early, err := repository.Prepare(fixture.ctx, owner, true, fixture.now.Add(30*time.Second))
			if err != nil || early.Claimed {
				t.Fatalf("early rate-limit retry=%+v err=%v", early, err)
			}
		}
	}
}

func concurrentPrepare(t *testing.T, repository *postgresadapter.InventoryRepository, fixture *repositoryFixture, owner uuid.UUID, retry bool) int {
	t.Helper()
	return concurrentPrepareAt(t, repository, fixture, owner, retry, fixture.now)
}

func concurrentPrepareAt(t *testing.T, repository *postgresadapter.InventoryRepository, fixture *repositoryFixture, owner uuid.UUID, retry bool, now time.Time) int {
	t.Helper()
	const workers = 12
	start := make(chan struct{})
	results := make(chan portfolio.Preparation, workers)
	errors := make(chan error, workers)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			result, err := repository.Prepare(fixture.ctx, owner, retry, now)
			results <- result
			errors <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	claims := 0
	for result := range results {
		if result.Claimed {
			claims++
		}
	}
	return claims
}
