package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	provideradapter "github.com/kennethdavidbuck/findur/backend/internal/platform/provider"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestScheduledInventoryRepairsTwentyTwoConnectionLegacyShapeWithoutBrowserRequest(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	owner := inclusionOwnerWithInventory(t, fixture, 236, "legacy-twenty-two-connection-owner")
	bootstrap, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !bootstrap.Claimed {
		t.Fatalf("bootstrap=%+v err=%v", bootstrap, err)
	}
	legacyConnections := make([]portfolio.Connection, 22)
	for index := range legacyConnections {
		legacyConnections[index] = portfolio.Connection{ID: uuid.NewString(), BrokerageLabel: "Legacy synthetic broker", Status: portfolio.ConnectionStatusUnavailable, SyncMode: portfolio.SyncModeUnknown}
		if index < 15 {
			legacyConnections[index].Accounts = scheduledInventoryConnections(uuid.NewString())[0].Accounts
		}
	}
	if _, accepted, err := repository.Finalize(fixture.ctx, owner, bootstrap.Generation, portfolio.StateUnavailable, nil, legacyConnections, fixture.now); err != nil || !accepted {
		t.Fatalf("legacy publication accepted=%v err=%v", accepted, err)
	}

	connections := make([]map[string]any, 22)
	accounts := make([]map[string]any, 22)
	for index := range connections {
		connectionID := uuid.NewString()
		connections[index] = map[string]any{
			"id": connectionID, "disabled": false,
			"brokerage":           map[string]any{"display_name": "Synthetic Broker"},
			"data_freshness_mode": map[string]any{"institution": "realtime", "snaptrade": "delayed"},
		}
		accounts[index] = map[string]any{
			"id": uuid.NewString(), "brokerage_authorization": connectionID,
			"name": "Synthetic investment", "number": "••••0001", "institution_name": "Synthetic Broker",
			"account_category": "INVESTMENT", "raw_type": "Margin", "status": "open",
			"sync_status": map[string]any{"holdings": map[string]any{"initial_sync_completed": true}},
			"balance":     map[string]any{"total": map[string]any{"amount": "125000.25", "currency": "CAD"}},
		}
	}
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		paths = append(paths, request.URL.Path)
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/authorizations" {
			_ = json.NewEncoder(response).Encode(connections)
			return
		}
		if request.URL.Path == "/accounts" {
			_ = json.NewEncoder(response).Encode(accounts)
			return
		}
		http.NotFound(response, request)
	}))
	defer server.Close()
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	providerClient, err := provideradapter.NewInventoryClient(baseURL, server.Client(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	service, err := portfolio.NewService(repository, providerClient, inventoryCredentialReader{token: "synthetic-access-token"}, time.Now, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	processed, err := service.RefreshDue(fixture.ctx, 24*time.Hour, time.Minute)
	if err != nil || !processed {
		t.Fatalf("processed=%v err=%v", processed, err)
	}
	if !slices.Equal(paths, []string{"/authorizations", "/accounts"}) {
		t.Fatalf("provider paths=%v", paths)
	}
	var generation int64
	var status string
	var connectionCount, accountCount int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT inventory.head_generation,inventory.current_status,
		(SELECT count(*) FROM portfolio_inventory_connections WHERE user_id=$1 AND generation=inventory.head_generation),
		(SELECT count(*) FROM portfolio_inventory_accounts WHERE user_id=$1 AND generation=inventory.head_generation)
		FROM portfolio_inventory_state inventory WHERE inventory.user_id=$1`, owner).Scan(&generation, &status, &connectionCount, &accountCount); err != nil {
		t.Fatal(err)
	}
	if generation != 2 || status != string(portfolio.StateReady) || connectionCount != 22 || accountCount != 22 {
		t.Fatalf("generation=%d status=%s connections=%d accounts=%d", generation, status, connectionCount, accountCount)
	}
	var totalAmount, totalCurrency string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT total_balance_amount,total_balance_currency
		FROM portfolio_inventory_accounts
		WHERE user_id=$1 AND generation=$2
		ORDER BY account_id
		LIMIT 1`, owner, generation).Scan(&totalAmount, &totalCurrency); err != nil {
		t.Fatal(err)
	}
	if totalAmount != "125000.25" || totalCurrency != "CAD" {
		t.Fatalf("total=%s currency=%s", totalAmount, totalCurrency)
	}
}

type inventoryCredentialReader struct{ token string }

func (r inventoryCredentialReader) Read(ctx context.Context, _ uuid.UUID, read func(context.Context, string) error) error {
	return read(ctx, r.token)
}

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

func TestInventoryRepositoryPublishesDisabledRowsAndReturningLoginPreservesHead(t *testing.T) {
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
	if stale, published, err := repository.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, disabledRows, fixture.now.Add(time.Minute)); err != nil || published || stale.Generation != claim.Generation {
		t.Fatalf("old bearer completion snapshot=%+v published=%v err=%v", stale, published, err)
	}
	bootstrap, err := repository.Prepare(fixture.ctx, owner, false, fixture.now.Add(time.Minute))
	if err != nil || bootstrap.Claimed || bootstrap.Generation != claim.Generation || len(bootstrap.Connections) != 1 {
		t.Fatalf("bootstrap=%+v err=%v", bootstrap, err)
	}
	var versions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_versions WHERE user_id=$1`, owner).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if versions != 1 {
		t.Fatalf("inventory versions=%d, want preserved first inventory version", versions)
	}
}

func TestInventoryRepositoryRetainsButDoesNotServeFailedLegacyRows(t *testing.T) {
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
		if err != nil || !accepted || published.State != state || len(published.Connections) != 0 {
			t.Fatalf("state=%s snapshot=%+v accepted=%v err=%v", state, published, accepted, err)
		}
		var retainedRows int
		if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_connections WHERE user_id=$1`, owner).Scan(&retainedRows); err != nil || retainedRows != 1 {
			t.Fatalf("state=%s retainedRows=%d err=%v", state, retainedRows, err)
		}
		revisit, err := repository.Prepare(fixture.ctx, owner, false, fixture.now.Add(30*time.Second))
		if err != nil || revisit.Claimed || revisit.State != state || len(revisit.Connections) != 0 {
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

func TestInventoryRepositoryScheduledClaimsDueWorkAndRetainsLastGoodHead(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	owner := scheduledInventoryOwner(t, fixture, repository, 211, "scheduled-inventory-owner")
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions
		SET published_at=$2 WHERE user_id=$1 AND generation=1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
	if err != nil || claim == nil || claim.Owner != owner || claim.Generation != 2 {
		t.Fatalf("exact-boundary claim=%+v err=%v", claim, err)
	}
	failures := []struct {
		name          string
		delay         time.Duration
		providerDelay time.Duration
	}{
		{name: "first", delay: time.Minute, providerDelay: 30 * time.Second},
		{name: "second", delay: 2 * time.Minute},
		{name: "third", delay: 4 * time.Minute},
		{name: "fourth", delay: 8 * time.Minute},
		{name: "fifth", delay: 16 * time.Minute},
		{name: "sixth", delay: 32 * time.Minute},
		{name: "capped", delay: 32 * time.Minute},
		{name: "provider retry wins", delay: 45 * time.Minute, providerDelay: 45 * time.Minute},
	}
	failedAt := fixture.now
	for index, test := range failures {
		var providerRetryAt *time.Time
		if test.providerDelay > 0 {
			value := failedAt.Add(test.providerDelay)
			providerRetryAt = &value
		}
		snapshot, accepted, err := repository.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateUnavailable, providerRetryAt, nil, failedAt)
		if err != nil || !accepted || snapshot.State != portfolio.StateReady || len(snapshot.Connections) != 1 {
			t.Fatalf("%s retained snapshot=%+v accepted=%v err=%v", test.name, snapshot, accepted, err)
		}
		var failureCount int
		var retryAt time.Time
		if err := fixture.pool.QueryRow(fixture.ctx, `SELECT failure_count,retry_at FROM portfolio_inventory_state WHERE user_id=$1`, owner).Scan(&failureCount, &retryAt); err != nil {
			t.Fatal(err)
		}
		if failureCount != index+1 || !retryAt.Equal(failedAt.Add(test.delay)) {
			t.Fatalf("%s failureCount=%d retryAt=%v", test.name, failureCount, retryAt)
		}
		if early, err := repository.ClaimDue(fixture.ctx, retryAt.Add(-time.Nanosecond), 24*time.Hour, time.Minute); err != nil || early != nil {
			t.Fatalf("%s early retry claim=%+v err=%v", test.name, early, err)
		}
		claim, err = repository.ClaimDue(fixture.ctx, retryAt, 24*time.Hour, time.Minute)
		if err != nil || claim == nil || claim.Generation != int64(index+3) {
			t.Fatalf("%s retry claim=%+v err=%v", test.name, claim, err)
		}
		failedAt = retryAt
	}
	connections := scheduledInventoryConnections("account")
	if _, accepted, err := repository.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateReady, nil, connections, failedAt); err != nil || !accepted {
		t.Fatalf("retry publication accepted=%v err=%v", accepted, err)
	}
	var failureCount int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT failure_count FROM portfolio_inventory_state WHERE user_id=$1`, owner).Scan(&failureCount); err != nil || failureCount != 0 {
		t.Fatalf("cleared failureCount=%d err=%v", failureCount, err)
	}
}

func TestInventoryRepositoryScheduledClaimsNeverSyncedAndLegacyFailure(t *testing.T) {
	fixture := newRepositoryFixture(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)

	t.Run("never synced", func(t *testing.T) {
		fixture.reset(t)
		owner := inclusionOwnerWithInventory(t, fixture, 205, "never-synced-inventory-owner")
		claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
		if err != nil || claim == nil || claim.Owner != owner || claim.Generation != 1 {
			t.Fatalf("claim=%+v err=%v", claim, err)
		}
	})

	t.Run("legacy unavailable", func(t *testing.T) {
		fixture.reset(t)
		owner := inclusionOwnerWithInventory(t, fixture, 207, "legacy-failed-inventory-owner")
		bootstrap, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
		if err != nil || !bootstrap.Claimed {
			t.Fatalf("bootstrap=%+v err=%v", bootstrap, err)
		}
		if _, accepted, err := repository.Finalize(fixture.ctx, owner, bootstrap.Generation, portfolio.StateUnavailable, nil, nil, fixture.now); err != nil || !accepted {
			t.Fatalf("legacy failure accepted=%v err=%v", accepted, err)
		}
		claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
		if err != nil || claim == nil || claim.Owner != owner || claim.Generation != 2 {
			t.Fatalf("claim=%+v err=%v", claim, err)
		}
	})
}

func TestFreshAuthorizationClearsInventoryBackoffAndMakesOldHeadDue(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	owner := scheduledInventoryOwner(t, fixture, repository, 228, "inventory-authorization-resume-owner")
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions
		SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_state
		SET retry_at=$2,failure_count=4 WHERE user_id=$1`, owner, fixture.now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	credentials := postgresadapter.NewCredentialRepository(fixture.pool)
	credential, found, err := credentials.ReadCredential(fixture.ctx, owner)
	if err != nil || !found {
		t.Fatalf("credential found=%v err=%v", found, err)
	}
	if err := credentials.RequireReauthorization(fixture.ctx, owner, nil, credential.Version, fixture.now); err != nil {
		t.Fatal(err)
	}
	if claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute); err != nil || claim != nil {
		t.Fatalf("inactive authorization claim=%+v err=%v", claim, err)
	}

	attempt := fixture.createAttempt(t, 229, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, attempt)
	fresh := finalization(attempt.StateHash, owner, fixture.now.Add(time.Minute), 230)
	fresh.Subject = "inventory-authorization-resume-owner"
	if err := fixture.repository.FinalizeCallback(fixture.ctx, fresh); err != nil {
		t.Fatal(err)
	}
	var retryAt *time.Time
	var failureCount int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT retry_at,failure_count
		FROM portfolio_inventory_state WHERE user_id=$1`, owner).Scan(&retryAt, &failureCount); err != nil {
		t.Fatal(err)
	}
	if retryAt != nil || failureCount != 0 {
		t.Fatalf("retryAt=%v failureCount=%d", retryAt, failureCount)
	}
	claim, err := repository.ClaimDue(fixture.ctx, fresh.CompletedAt, 24*time.Hour, time.Minute)
	if err != nil || claim == nil {
		t.Fatalf("resumed inventory claim=%+v err=%v", claim, err)
	}
}

func TestInventoryRepositoryScheduledClaimExclusionsAndConcurrency(t *testing.T) {
	fixture := newRepositoryFixture(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)

	tests := []struct {
		name  string
		block func(*testing.T, uuid.UUID)
	}{
		{
			name: "pending inclusion",
			block: func(t *testing.T, owner uuid.UUID) {
				inclusion := postgresadapter.NewInclusionRepository(fixture.pool)
				prepared, err := inclusion.PrepareInclusion(fixture.ctx, owner, 0, "pending-inventory-exclusion", []string{"account"}, fixture.now)
				if err != nil || !prepared.Claimed {
					t.Fatalf("pending inclusion=%+v err=%v", prepared, err)
				}
			},
		},
		{
			name: "active resource claim",
			block: func(t *testing.T, owner uuid.UUID) {
				_, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_account_sync_state
					(user_id,account_id,claim_id,claim_expires_at,updated_at) VALUES ($1,'account',$2,$3,$4)
					ON CONFLICT (user_id,account_id) DO UPDATE SET claim_id=EXCLUDED.claim_id,claim_expires_at=EXCLUDED.claim_expires_at`, owner, uuid.New(), fixture.now.Add(time.Minute), fixture.now)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "inactive authorization",
			block: func(t *testing.T, owner uuid.UUID) {
				if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE provider_authorizations SET lifecycle_status='reauthorization-required' WHERE user_id=$1`, owner); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture.reset(t)
			owner := scheduledInventoryOwner(t, fixture, repository, byte(220+index*2), "inventory-exclusion-"+test.name)
			if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
				t.Fatal(err)
			}
			test.block(t, owner)
			claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
			if err != nil || claim != nil {
				t.Fatalf("blocked claim=%+v err=%v", claim, err)
			}
		})
	}

	t.Run("concurrent workers yield one claim", func(t *testing.T) {
		fixture.reset(t)
		owner := scheduledInventoryOwner(t, fixture, repository, 230, "inventory-concurrent-owner")
		if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
			t.Fatal(err)
		}
		const workers = 12
		start := make(chan struct{})
		results := make(chan *portfolio.ScheduledInventoryClaim, workers)
		errors := make(chan error, workers)
		var group sync.WaitGroup
		for range workers {
			group.Add(1)
			go func() {
				defer group.Done()
				<-start
				claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
				results <- claim
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
		for claim := range results {
			if claim != nil {
				claims++
			}
		}
		if claims != 1 {
			t.Fatalf("claims=%d, want 1", claims)
		}
	})
}

func TestInventoryClaimBlocksInclusionAndStaleScheduledResultCannotPublish(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	inventory := postgresadapter.NewInventoryRepository(fixture.pool)
	owner := scheduledInventoryOwner(t, fixture, inventory, 232, "inventory-inclusion-fence-owner")
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	claim, err := inventory.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
	if err != nil || claim == nil {
		t.Fatalf("inventory claim=%+v err=%v", claim, err)
	}
	inclusion := postgresadapter.NewInclusionRepository(fixture.pool)
	if _, err := inclusion.PrepareInclusion(fixture.ctx, owner, 0, "blocked-by-inventory", []string{"account"}, fixture.now); !errors.Is(err, portfolio.ErrInclusionConflict) {
		t.Fatalf("active inventory claim inclusion error=%v", err)
	}

	afterLease := fixture.now.Add(time.Minute)
	prepared, err := inclusion.PrepareInclusion(fixture.ctx, owner, 0, "after-inventory-expired", []string{"account"}, afterLease)
	if err != nil || !prepared.Claimed {
		t.Fatalf("post-expiry inclusion=%+v err=%v", prepared, err)
	}
	if _, accepted, err := inventory.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateReady, nil, scheduledInventoryConnections("new-account"), afterLease); err != nil || accepted {
		t.Fatalf("stale inventory accepted=%v err=%v", accepted, err)
	}
	var versions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_versions WHERE user_id=$1`, owner).Scan(&versions); err != nil || versions != 1 {
		t.Fatalf("inventory versions=%d err=%v", versions, err)
	}
}

func TestScheduledInventoryRejectsExpiredAndLifecycleStaleResults(t *testing.T) {
	fixture := newRepositoryFixture(t)
	tests := []struct {
		name   string
		marker byte
		mutate func(*testing.T, uuid.UUID) time.Time
	}{
		{
			name:   "expired lease",
			marker: 220,
			mutate: func(_ *testing.T, _ uuid.UUID) time.Time { return fixture.now.Add(time.Minute) },
		},
		{
			name:   "authorization rotated",
			marker: 221,
			mutate: func(t *testing.T, owner uuid.UUID) time.Time {
				attempt := fixture.createAttempt(t, 222, fixture.now.Add(10*time.Minute))
				fixture.claimCallback(t, attempt)
				fresh := finalization(attempt.StateHash, owner, fixture.now.Add(30*time.Second), 223)
				fresh.Subject = "scheduled-stale-authorization-rotated"
				if err := fixture.repository.FinalizeCallback(fixture.ctx, fresh); err != nil {
					t.Fatal(err)
				}
				return fresh.CompletedAt
			},
		},
		{
			name:   "owner deactivated",
			marker: 224,
			mutate: func(t *testing.T, owner uuid.UUID) time.Time {
				if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE users SET active=false WHERE id=$1`, owner); err != nil {
					t.Fatal(err)
				}
				return fixture.now.Add(30 * time.Second)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture.reset(t)
			repository := postgresadapter.NewInventoryRepository(fixture.pool)
			subject := "scheduled-stale-" + strings.ReplaceAll(test.name, " ", "-")
			owner := scheduledInventoryOwner(t, fixture, repository, test.marker, subject)
			if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions
				SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
				t.Fatal(err)
			}
			claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
			if err != nil || claim == nil || claim.AuthorizationGeneration == 0 {
				t.Fatalf("claim=%+v err=%v", claim, err)
			}
			finalizedAt := test.mutate(t, owner)
			if _, accepted, err := repository.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateReady, nil, scheduledInventoryConnections("new-account"), finalizedAt); err != nil || accepted {
				t.Fatalf("accepted=%v err=%v", accepted, err)
			}
			var versions int
			if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_versions WHERE user_id=$1`, owner).Scan(&versions); err != nil || versions != 1 {
				t.Fatalf("versions=%d err=%v", versions, err)
			}
		})
	}
}

func TestScheduledInventoryPublicationPreservesStableDependenciesAndRemovesIneligibleInclusion(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInventoryRepository(fixture.pool)
	owner := scheduledInventoryOwner(t, fixture, repository, 234, "inventory-identity-owner")
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inclusion_state (user_id,version,lifecycle_generation,updated_at) VALUES ($1,1,1,$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_included_accounts (user_id,account_id,inclusion_version,included_at) VALUES ($1,'account',1,$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_account_sync_state (user_id,account_id,updated_at) VALUES ($1,'account',$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	balanceVersion, positionVersion, activityVersion := uuid.New(), uuid.New(), uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_versions
		(id,user_id,account_id,inclusion_version,retrieved_at,published_at) VALUES ($2,$1,'account',1,$3,$3)`, owner, balanceVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_heads (user_id,account_id,version_id) VALUES ($1,'account',$2)`, owner, balanceVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_position_versions
		(id,user_id,account_id,inclusion_version,observed_at,retrieved_at,published_at) VALUES ($2,$1,'account',1,$3,$3,$3)`, owner, positionVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_position_heads (user_id,account_id,version_id) VALUES ($1,'account',$2)`, owner, positionVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_activity_versions
		(id,user_id,account_id,inclusion_version,retrieved_at,published_at) VALUES ($2,$1,'account',1,$3,$3)`, owner, activityVersion, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_activity_heads (user_id,account_id,version_id) VALUES ($1,'account',$2)`, owner, activityVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
	if err != nil || claim == nil {
		t.Fatalf("claim=%+v err=%v", claim, err)
	}
	connections := scheduledInventoryConnections("account")
	connections = append(connections, portfolio.Connection{
		ID: "new-connection", BrokerageLabel: "New synthetic broker", Status: portfolio.ConnectionStatusActive, SyncMode: portfolio.SyncModeDelayed, Available: true, Eligible: true,
		Accounts: []portfolio.Account{{ID: "new-account", Category: portfolio.AccountCategoryInvestment, Type: "Cash", MaskedLabel: "New synthetic account", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}},
	})
	if _, accepted, err := repository.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateReady, nil, connections, fixture.now); err != nil || !accepted {
		t.Fatalf("expanded publication accepted=%v err=%v", accepted, err)
	}
	var identities, membership, syncRows int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_account_identities WHERE user_id=$1),
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id='account')`, owner).Scan(&identities, &membership, &syncRows); err != nil {
		t.Fatal(err)
	}
	if identities != 2 || membership != 1 || syncRows != 1 {
		t.Fatalf("identities=%d membership=%d syncRows=%d", identities, membership, syncRows)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1 AND generation=2`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	missingClaim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
	if err != nil || missingClaim == nil {
		t.Fatalf("missing-account claim=%+v err=%v", missingClaim, err)
	}
	if _, accepted, err := repository.FinalizeScheduled(fixture.ctx, *missingClaim, portfolio.StateReady, nil, scheduledInventoryConnections("new-account"), fixture.now); err != nil || !accepted {
		t.Fatalf("missing-account publication accepted=%v err=%v", accepted, err)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_account_identities WHERE user_id=$1),
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id='account')`, owner).Scan(&identities, &membership, &syncRows); err != nil {
		t.Fatal(err)
	}
	if identities != 2 || membership != 0 || syncRows != 1 {
		t.Fatalf("retained identities=%d membership=%d syncRows=%d", identities, membership, syncRows)
	}
	var financialVersions, financialHeads int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1 AND account_id='account')
			+ (SELECT count(*) FROM portfolio_position_versions WHERE user_id=$1 AND account_id='account')
			+ (SELECT count(*) FROM portfolio_activity_versions WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1 AND account_id='account')
			+ (SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1 AND account_id='account')
			+ (SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1 AND account_id='account')`, owner).Scan(&financialVersions, &financialHeads); err != nil {
		t.Fatal(err)
	}
	if financialVersions != 3 || financialHeads != 3 {
		t.Fatalf("retained financial versions=%d heads=%d", financialVersions, financialHeads)
	}
	var inclusionVersion, lifecycleGeneration int64
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT version,lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1`, owner).Scan(&inclusionVersion, &lifecycleGeneration); err != nil {
		t.Fatal(err)
	}
	if inclusionVersion != 2 || lifecycleGeneration != 2 {
		t.Fatalf("inclusion version=%d lifecycle=%d", inclusionVersion, lifecycleGeneration)
	}
}

func TestScheduledInventoryRemovesPresentButIneligibleAccounts(t *testing.T) {
	fixture := newRepositoryFixture(t)
	tests := []struct {
		name   string
		marker byte
		mutate func([]portfolio.Connection)
	}{
		{
			name:   "account unavailable",
			marker: 225,
			mutate: func(connections []portfolio.Connection) {
				connections[0].Accounts[0].Available = false
				connections[0].Accounts[0].Selectable = false
				connections[0].Accounts[0].UsabilityReason = portfolio.UsabilityAccountUnavailable
			},
		},
		{
			name:   "sync unavailable",
			marker: 226,
			mutate: func(connections []portfolio.Connection) {
				connections[0].Accounts[0].SyncState = portfolio.AccountSyncStateUnavailable
				connections[0].Accounts[0].Selectable = false
			},
		},
		{
			name:   "connection disabled",
			marker: 227,
			mutate: func(connections []portfolio.Connection) {
				connections[0].Status = portfolio.ConnectionStatusDisabled
				connections[0].Available = false
				connections[0].Accounts[0].Selectable = false
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture.reset(t)
			repository := postgresadapter.NewInventoryRepository(fixture.pool)
			owner := scheduledInventoryOwner(t, fixture, repository, test.marker, "present-ineligible-"+strings.ReplaceAll(test.name, " ", "-"))
			if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inclusion_state (user_id,version,lifecycle_generation,updated_at) VALUES ($1,1,1,$2)`, owner, fixture.now); err != nil {
				t.Fatal(err)
			}
			if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_included_accounts (user_id,account_id,inclusion_version,included_at) VALUES ($1,'account',1,$2)`, owner, fixture.now); err != nil {
				t.Fatal(err)
			}
			if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_account_sync_state (user_id,account_id,updated_at) VALUES ($1,'account',$2)`, owner, fixture.now); err != nil {
				t.Fatal(err)
			}
			if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_versions SET published_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-24*time.Hour)); err != nil {
				t.Fatal(err)
			}
			claim, err := repository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
			if err != nil || claim == nil {
				t.Fatalf("claim=%+v err=%v", claim, err)
			}
			connections := scheduledInventoryConnections("account")
			test.mutate(connections)
			if _, accepted, err := repository.FinalizeScheduled(fixture.ctx, *claim, portfolio.StateReady, nil, connections, fixture.now); err != nil || !accepted {
				t.Fatalf("accepted=%v err=%v", accepted, err)
			}
			var membership, identities, syncRows int
			if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
				(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1 AND account_id='account'),
				(SELECT count(*) FROM portfolio_account_identities WHERE user_id=$1 AND account_id='account'),
				(SELECT count(*) FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id='account')`, owner).Scan(&membership, &identities, &syncRows); err != nil {
				t.Fatal(err)
			}
			if membership != 0 || identities != 1 || syncRows != 1 {
				t.Fatalf("membership=%d identities=%d syncRows=%d", membership, identities, syncRows)
			}
		})
	}
}

func scheduledInventoryOwner(t *testing.T, fixture *repositoryFixture, repository *postgresadapter.InventoryRepository, marker byte, subject string) uuid.UUID {
	t.Helper()
	owner := inclusionOwnerWithInventory(t, fixture, marker, subject)
	claim, err := repository.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("initial claim=%+v err=%v", claim, err)
	}
	if _, accepted, err := repository.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, scheduledInventoryConnections("account"), fixture.now); err != nil || !accepted {
		t.Fatalf("initial publication accepted=%v err=%v", accepted, err)
	}
	return owner
}

func scheduledInventoryConnections(accountID string) []portfolio.Connection {
	return []portfolio.Connection{{
		ID: "connection", BrokerageLabel: "Synthetic Broker", Status: portfolio.ConnectionStatusActive, SyncMode: portfolio.SyncModeRealtime, Available: true, Eligible: true,
		Accounts: []portfolio.Account{{ID: accountID, Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Synthetic account", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}},
	}}
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
