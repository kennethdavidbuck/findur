package postgres_test

import (
	"bytes"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestInventoryRepositoryFiltersLegacyCachedAccountsWithoutRetry(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 201, "legacy-inventory-owner")
	excluded := publishLegacyEligibilityInventory(t, fixture, owner)

	// All excluded rows really exist in the immutable old snapshot, including
	// obsolete selectable=true flags. Reading must apply today's policy without
	// refreshing provider data or rewriting historical inventory.
	var persistedCount int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_accounts WHERE user_id=$1`, owner).Scan(&persistedCount); err != nil {
		t.Fatal(err)
	}
	if persistedCount != len(excluded)+3 {
		t.Fatalf("persisted accounts=%d, want %d", persistedCount, len(excluded)+3)
	}
	reloaded, err := postgresadapter.NewInventoryRepository(fixture.pool).Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || reloaded.Claimed || reloaded.State != portfolio.StateReady {
		t.Fatalf("cached inventory=%+v err=%v", reloaded, err)
	}
	var ids []string
	for _, connection := range reloaded.Connections {
		for _, account := range connection.Accounts {
			ids = append(ids, account.ID)
			if !account.Selectable {
				t.Fatalf("accepted cached account is not selectable: %+v", account)
			}
			if account.Category == portfolio.AccountCategoryUnknown {
				if account.Eligible || account.UsabilityReason != portfolio.UsabilityProvisionalCategory {
					t.Fatalf("unknown-category account lost provisional semantics: %+v", account)
				}
			} else if !account.Eligible || account.UsabilityReason != portfolio.UsabilityReady {
				t.Fatalf("investment account does not project current eligibility: %+v", account)
			}
		}
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []string{"open-investment", "unknown", "unspecified-status-investment"}) {
		t.Fatalf("cached inventory accounts=%v", ids)
	}
	var historicalEligible bool
	var historicalReason string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT eligible,usability_reason FROM portfolio_inventory_accounts WHERE user_id=$1 AND account_id='unspecified-status-investment'`, owner).Scan(&historicalEligible, &historicalReason); err != nil {
		t.Fatal(err)
	}
	if historicalEligible || historicalReason != string(portfolio.UsabilityProvisionalStatus) {
		t.Fatalf("cached projection rewrote history: eligible=%v reason=%s", historicalEligible, historicalReason)
	}
}

func TestInventoryRepositoryProjectsExcludedOnlyCachedHeadAsEmpty(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 203, "excluded-only-inventory-owner")
	publishLegacyEligibilityInventory(t, fixture, owner)
	if _, err := fixture.pool.Exec(fixture.ctx, `DELETE FROM portfolio_inventory_accounts WHERE user_id=$1 AND account_id IN ('open-investment','unspecified-status-investment','unknown')`, owner); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	empty, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || empty.Claimed || empty.State != portfolio.StateEmpty || empty.Generation != 1 {
		t.Fatalf("excluded-only cached projection=%+v err=%v", empty, err)
	}
	for _, connection := range empty.Connections {
		if len(connection.Accounts) != 0 {
			t.Fatalf("excluded account leaked into empty projection: %+v", connection)
		}
	}
	var historicalStatus string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT current_status FROM portfolio_inventory_state WHERE user_id=$1`, owner).Scan(&historicalStatus); err != nil {
		t.Fatal(err)
	}
	if historicalStatus != string(portfolio.StateReady) {
		t.Fatalf("empty projection rewrote persisted status=%s", historicalStatus)
	}
	for _, state := range []portfolio.State{portfolio.StateUnavailable, portfolio.StateMalformed, portfolio.StateUnauthorized, portfolio.StateRateLimited, portfolio.StateDisabled, portfolio.StatePending} {
		if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_state SET current_status=$2,claim_expires_at=CASE WHEN $2='pending' THEN $3::timestamptz ELSE NULL END WHERE user_id=$1`, owner, state, fixture.now.Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
		cached, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
		if err != nil || cached.Claimed || cached.State != state || cached.Generation != empty.Generation {
			t.Fatalf("categorical state %s changed or triggered refresh: %+v err=%v", state, cached, err)
		}
	}
}

func publishLegacyEligibilityInventory(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) []string {
	t.Helper()
	base := portfolio.Account{Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Synthetic investment", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}
	open, unspecified, provisionalCategory := base, base, base
	open.ID, unspecified.ID, provisionalCategory.ID = "open-investment", "unspecified-status-investment", "unknown"
	unspecified.Eligible, unspecified.UsabilityReason = false, portfolio.UsabilityProvisionalStatus
	provisionalCategory.Category, provisionalCategory.Eligible, provisionalCategory.UsabilityReason = portfolio.AccountCategoryUnknown, false, portfolio.UsabilityProvisionalCategory
	active := portfolio.Connection{ID: "active", BrokerageLabel: "Synthetic Broker", Status: portfolio.ConnectionStatusActive, SyncMode: portfolio.SyncModeRealtime, Available: true, Eligible: true, Accounts: []portfolio.Account{open, unspecified, provisionalCategory}}
	var excluded []string
	add := func(id string, change func(*portfolio.Account)) {
		account := base
		account.ID = id
		change(&account)
		active.Accounts = append(active.Accounts, account)
		excluded = append(excluded, id)
	}
	for _, category := range []portfolio.AccountCategory{portfolio.AccountCategoryDeposit, portfolio.AccountCategoryCredit} {
		add(string(category), func(account *portfolio.Account) { account.Category = category })
	}
	for _, syncState := range []portfolio.AccountSyncState{portfolio.AccountSyncStatePending, portfolio.AccountSyncStateUnavailable, portfolio.AccountSyncStateUnknown} {
		add("sync-"+string(syncState), func(account *portfolio.Account) { account.SyncState = syncState })
	}
	add("unavailable-account", func(account *portfolio.Account) { account.Available = false })
	for _, reason := range []portfolio.UsabilityReason{portfolio.UsabilityAccountClosed, portfolio.UsabilityAccountUnavailable, portfolio.UsabilitySyncUnavailable, portfolio.UsabilityProvisionalCategory} {
		add("reason-"+string(reason), func(account *portfolio.Account) { account.UsabilityReason = reason })
	}
	connections := []portfolio.Connection{active}
	for _, status := range []portfolio.ConnectionStatus{portfolio.ConnectionStatusDisabled, portfolio.ConnectionStatusUnavailable} {
		connection, account := active, base
		connection.ID, connection.Status = string(status), status
		account.ID = "connection-" + string(status)
		connection.Accounts = []portfolio.Account{account}
		connections = append(connections, connection)
		excluded = append(excluded, account.ID)
	}
	connection, account := active, base
	connection.ID, connection.Available = "unavailable-active", false
	account.ID = "unavailable-active-account"
	connection.Accounts = []portfolio.Account{account}
	connections = append(connections, connection)
	excluded = append(excluded, account.ID)

	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	claim, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("legacy inventory claim=%+v err=%v", claim, err)
	}
	if _, accepted, err := repository.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, connections, fixture.now); err != nil || !accepted {
		t.Fatalf("legacy publication accepted=%v err=%v", accepted, err)
	}
	return excluded
}

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
