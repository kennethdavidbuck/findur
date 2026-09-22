package postgres_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/migrations"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestAccountInclusionMigrationBackfillsExistingInventory(t *testing.T) {
	fixture := newRepositoryFixture(t)
	_, sourceFile, _, _ := runtime.Caller(0)
	migrationPath, err := filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "../../../db/migrations"))
	if err != nil {
		t.Fatal(err)
	}
	sourceURL := "file://" + filepath.ToSlash(migrationPath)
	databaseURL := fixture.pool.Config().ConnString()
	migrator, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Migrate(5); err != nil {
		t.Fatal(err)
	}
	if sourceErr, databaseErr := migrator.Close(); sourceErr != nil || databaseErr != nil {
		t.Fatalf("close migrator source=%v database=%v", sourceErr, databaseErr)
	}

	owner := uuid.New()
	firstSeen, lastSeen := fixture.now.Add(-time.Hour), fixture.now
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO users (id,origin) VALUES ($1,'oauth')`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inventory_state (user_id,current_generation,current_status,head_generation,updated_at) VALUES ($1,2,'ready',2,$2)`, owner, lastSeen); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inventory_versions (user_id,generation,status,published_at) VALUES ($1,1,'ready',$2),($1,2,'ready',$3)`, owner, firstSeen, lastSeen); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inventory_connections (user_id,generation,connection_id,brokerage_label,status,sync_mode,available,eligible) VALUES
			($1,1,'connection-old','Broker','active','realtime',true,true),
			($1,2,'connection','Broker','active','realtime',true,true),
			($1,2,'connection-disabled','Disabled','disabled','unknown',false,false)`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inventory_accounts (user_id,generation,account_id,connection_id,category,account_type,masked_label,available,eligible,sync_state) VALUES
			($1,1,'ready-account','connection-old','investment','Margin','Ready old',true,true,'complete'),
			($1,2,'ready-account','connection','investment','Margin','Ready',true,true,'complete'),
			($1,2,'provisional-category-account','connection','unknown','Margin','Provisional category',true,false,'complete'),
			($1,2,'provisional-status-account','connection','investment','Margin','Provisional status',true,false,'complete'),
			($1,2,'pending-account','connection','investment','Margin','Pending',true,false,'pending'),
			($1,2,'deposit-account','connection','deposit','Chequing','Deposit',true,false,'complete'),
			($1,2,'unavailable-account','connection','investment','Margin','Unavailable',false,false,'unavailable'),
			($1,2,'disabled-account','connection-disabled','investment','Margin','Disabled',true,true,'complete')`, owner); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Up(fixture.ctx, sourceURL, databaseURL); err != nil {
		t.Fatal(err)
	}

	rows, err := fixture.pool.Query(fixture.ctx, `SELECT account_id,selectable,usability_reason FROM portfolio_inventory_accounts WHERE user_id=$1 AND generation=2 ORDER BY account_id`, owner)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := map[string]struct {
		selectable bool
		reason     string
	}{
		"deposit-account":              {false, "unsupported_category"},
		"disabled-account":             {false, "connection_disabled"},
		"pending-account":              {false, "sync_pending"},
		"provisional-category-account": {true, "provisional_category"},
		"provisional-status-account":   {true, "provisional_status"},
		"ready-account":                {true, "ready"},
		"unavailable-account":          {false, "sync_unavailable"},
	}
	for rows.Next() {
		var accountID, reason string
		var selectable bool
		if err := rows.Scan(&accountID, &selectable, &reason); err != nil {
			t.Fatal(err)
		}
		expected, ok := want[accountID]
		if !ok || selectable != expected.selectable || reason != expected.reason {
			t.Fatalf("backfill account=%q selectable=%v reason=%q expected=%+v", accountID, selectable, reason, expected)
		}
		delete(want, accountID)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(want) != 0 {
		t.Fatalf("missing backfilled accounts: %v", want)
	}
	var connectionID string
	var identityFirst, identityLast time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT connection_id,first_seen_at,last_seen_at FROM portfolio_account_identities WHERE user_id=$1 AND account_id='ready-account'`, owner).Scan(&connectionID, &identityFirst, &identityLast); err != nil {
		t.Fatal(err)
	}
	if connectionID != "connection" || !identityFirst.Equal(firstSeen) || !identityLast.Equal(lastSeen) {
		t.Fatalf("backfilled identity connection=%q first=%v last=%v", connectionID, identityFirst, identityLast)
	}
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	// The historical migration remains unchanged; admission applies the current
	// investment-only rule even though its legacy unknown-category flag is true.
	if _, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "reject-upgraded-unknown", []string{"provisional-category-account"}, fixture.now); !errors.Is(err, portfolio.ErrInvalidAccountSelection) {
		t.Fatalf("legacy unknown-category selection error=%v", err)
	}
	prepared, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "upgraded-selectable", []string{"ready-account", "provisional-status-account"}, fixture.now)
	if err != nil || !prepared.Claimed || len(prepared.Additions) != 2 {
		t.Fatalf("upgraded selectable accounts remained blocked: preparation=%+v err=%v", prepared, err)
	}
}

func TestInclusionAndReauthorizationShareOwnerLockOrder(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	repository := postgresadapter.NewInclusionRepository(fixture.pool)

	prepareOwner := inclusionOwnerWithInventory(t, fixture, 241, "prepare-race-owner")
	publishInclusionInventory(t, fixture, prepareOwner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	prepareAttempt := fixture.createAttempt(t, 243, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, prepareAttempt)
	prepareReauthorization := finalization(prepareAttempt.StateHash, prepareOwner, fixture.now.Add(time.Minute), 245)
	prepareReauthorization.Subject = "prepare-race-owner"

	ctx, cancel := context.WithTimeout(fixture.ctx, 10*time.Second)
	defer cancel()
	start := make(chan struct{})
	type prepareResult struct {
		value portfolio.InclusionPreparation
		err   error
	}
	prepareDone := make(chan prepareResult, 1)
	reauthorizationDone := make(chan error, 1)
	go func() {
		<-start
		value, err := repository.PrepareInclusion(ctx, prepareOwner, 0, "prepare-race", []string{"account-1"}, fixture.now)
		prepareDone <- prepareResult{value: value, err: err}
	}()
	go func() {
		<-start
		reauthorizationDone <- fixture.repository.FinalizeCallback(ctx, prepareReauthorization)
	}()
	close(start)
	prepared := <-prepareDone
	if err := <-reauthorizationDone; err != nil {
		t.Fatalf("concurrent prepare reauthorization: %v", err)
	}
	if prepared.err != nil && !errors.Is(prepared.err, portfolio.ErrInvalidAccountSelection) {
		t.Fatalf("concurrent prepare error=%v", prepared.err)
	}
	if prepared.err == nil {
		if _, accepted, err := repository.FinalizeInclusion(ctx, prepareOwner, prepared.value.ChangeID, prepared.value.Version, prepared.value.InventoryGeneration, prepared.value.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(2*time.Minute)); err != nil || accepted {
			t.Fatalf("pre-reauthorization preparation published stale data accepted=%v err=%v", accepted, err)
		}
	}
	assertNoPublishedInclusionRows(t, fixture, prepareOwner)

	finalizeOwner := inclusionOwnerWithInventory(t, fixture, 251, "finalize-race-owner")
	publishInclusionInventory(t, fixture, finalizeOwner, []portfolio.Account{{ID: "account-2", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Second (•••• 1002)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	finalizePrepared, err := repository.PrepareInclusion(ctx, finalizeOwner, 0, "finalize-race", []string{"account-2"}, fixture.now)
	if err != nil || !finalizePrepared.Claimed {
		t.Fatalf("finalize preparation=%+v err=%v", finalizePrepared, err)
	}
	finalizeAttempt := fixture.createAttempt(t, 253, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, finalizeAttempt)
	finalizeReauthorization := finalization(finalizeAttempt.StateHash, finalizeOwner, fixture.now.Add(3*time.Minute), 255)
	finalizeReauthorization.Subject = "finalize-race-owner"
	type finalizeResult struct {
		accepted bool
		err      error
	}
	finalizeDone := make(chan finalizeResult, 1)
	reauthorizationDone = make(chan error, 1)
	start = make(chan struct{})
	go func() {
		<-start
		_, accepted, err := repository.FinalizeInclusion(ctx, finalizeOwner, finalizePrepared.ChangeID, finalizePrepared.Version, finalizePrepared.InventoryGeneration, finalizePrepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-2": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(2*time.Minute))
		finalizeDone <- finalizeResult{accepted: accepted, err: err}
	}()
	go func() {
		<-start
		reauthorizationDone <- fixture.repository.FinalizeCallback(ctx, finalizeReauthorization)
	}()
	close(start)
	finalized := <-finalizeDone
	if finalized.err != nil {
		t.Fatalf("concurrent finalizer: %v", finalized.err)
	}
	if err := <-reauthorizationDone; err != nil {
		t.Fatalf("concurrent finalize reauthorization: %v", err)
	}
	var lifecycle int64
	if err := fixture.pool.QueryRow(ctx, `SELECT lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1`, finalizeOwner).Scan(&lifecycle); err != nil {
		t.Fatal(err)
	}
	if lifecycle != finalizePrepared.LifecycleGeneration+1 {
		t.Fatalf("finalize race lifecycle=%d old=%d", lifecycle, finalizePrepared.LifecycleGeneration)
	}
	if !finalized.accepted {
		assertNoPublishedInclusionRows(t, fixture, finalizeOwner)
	}
	if _, accepted, err := repository.FinalizeInclusion(ctx, finalizeOwner, finalizePrepared.ChangeID, finalizePrepared.Version, finalizePrepared.InventoryGeneration, finalizePrepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-2": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(4*time.Minute)); err != nil || accepted {
		t.Fatalf("post-reauthorization stale replay accepted=%v err=%v", accepted, err)
	}
}

func TestInclusionRepositoryIsIdempotentOwnerScopedAndRemovalFirst(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 201, "subject-one")
	inventory := postgresadapter.NewInventoryRepository(fixture.pool)
	claim, err := inventory.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("inventory claim=%+v err=%v", claim, err)
	}
	connections := []portfolio.Connection{{ID: "connection", BrokerageLabel: "Broker", Status: portfolio.ConnectionStatusActive, SyncMode: portfolio.SyncModeRealtime, Available: true, Eligible: true, Accounts: []portfolio.Account{
		{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
		{ID: "account-2", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Second (•••• 1002)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
	}}}
	if _, published, err := inventory.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, connections, fixture.now); err != nil || !published {
		t.Fatalf("inventory publication=%v err=%v", published, err)
	}

	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	initial, err := repository.GetInclusion(fixture.ctx, owner)
	if err != nil || initial.Version != 0 || len(initial.Committed) != 0 {
		t.Fatalf("initial=%+v err=%v", initial, err)
	}
	start := make(chan struct{})
	preparations := make(chan portfolio.InclusionPreparation, 2)
	errorsFound := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			result, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "add-first", []string{"account-1"}, fixture.now)
			preparations <- result
			errorsFound <- err
		}()
	}
	close(start)
	group.Wait()
	close(preparations)
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	var prepared portfolio.InclusionPreparation
	claims := 0
	for result := range preparations {
		if result.Version != 1 || len(result.Committed) != 0 {
			t.Fatalf("concurrent result=%+v", result)
		}
		if result.Claimed {
			claims++
			prepared = result
		}
	}
	if claims != 1 {
		t.Fatalf("concurrent idempotent claims=%d, want 1", claims)
	}
	data := map[string]portfolio.AccountData{"account-1": completeRepositoryAccountData(fixture.now)}
	committed, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, data, "", fixture.now)
	if err != nil || !accepted || len(committed.Committed) != 1 || committed.Change == nil || committed.Change.Status != portfolio.InclusionCommitted {
		t.Fatalf("committed=%+v accepted=%v err=%v", committed, accepted, err)
	}
	var balanceHeads, positionHeads, activityHeads int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1 AND account_id='account-1'),
		(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1 AND account_id='account-1'),
		(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1 AND account_id='account-1')`, owner).Scan(&balanceHeads, &positionHeads, &activityHeads); err != nil {
		t.Fatal(err)
	}
	if balanceHeads != 1 || positionHeads != 1 || activityHeads != 1 {
		t.Fatalf("dataset heads balances=%d positions=%d activities=%d", balanceHeads, positionHeads, activityHeads)
	}
	var balanceObserved, activityObserved *time.Time
	var positionObserved, balanceRetrieved, positionRetrieved, activityRetrieved, balancePublished, positionPublished, activityPublished time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT b.observed_at,p.observed_at,a.observed_at,b.retrieved_at,p.retrieved_at,a.retrieved_at,b.published_at,p.published_at,a.published_at
		FROM portfolio_balance_versions b
		JOIN portfolio_position_versions p ON p.user_id=b.user_id AND p.account_id=b.account_id
		JOIN portfolio_activity_versions a ON a.user_id=b.user_id AND a.account_id=b.account_id
		WHERE b.user_id=$1 AND b.account_id='account-1'`, owner).Scan(&balanceObserved, &positionObserved, &activityObserved, &balanceRetrieved, &positionRetrieved, &activityRetrieved, &balancePublished, &positionPublished, &activityPublished); err != nil {
		t.Fatal(err)
	}
	if balanceObserved != nil || activityObserved != nil || !positionObserved.Equal(fixture.now) || !balanceRetrieved.Equal(fixture.now) || !positionRetrieved.Equal(fixture.now) || !activityRetrieved.Equal(fixture.now) || !balancePublished.Equal(fixture.now) || !positionPublished.Equal(fixture.now) || !activityPublished.Equal(fixture.now) {
		t.Fatalf("dataset metadata balanceObserved=%v positionObserved=%v activityObserved=%v", balanceObserved, positionObserved, activityObserved)
	}
	replay, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "add-first", []string{"account-1"}, fixture.now)
	if err != nil || replay.Claimed || replay.Version != 1 || len(replay.Committed) != 1 {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	if _, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "stale", nil, fixture.now); !errors.Is(err, portfolio.ErrInclusionConflict) {
		t.Fatalf("stale error=%v", err)
	}
	if _, err := repository.PrepareInclusion(fixture.ctx, owner, 1, "foreign", []string{"foreign-account"}, fixture.now); !errors.Is(err, portfolio.ErrInvalidAccountSelection) {
		t.Fatalf("foreign error=%v", err)
	}

	older, err := repository.PrepareInclusion(fixture.ctx, owner, 1, "older-add", []string{"account-1", "account-2"}, fixture.now.Add(time.Second))
	if err != nil || !older.Claimed || older.Version != 2 || len(older.Committed) != 1 {
		t.Fatalf("older=%+v err=%v", older, err)
	}
	mixed, err := repository.PrepareInclusion(fixture.ctx, owner, 2, "mixed", []string{"account-2"}, fixture.now.Add(2*time.Second))
	if err != nil || !mixed.Claimed || mixed.Version != 3 || len(mixed.Committed) != 0 || mixed.Change == nil || len(mixed.Change.Removals) != 1 || mixed.LifecycleGeneration != older.LifecycleGeneration+1 {
		t.Fatalf("mixed=%+v err=%v", mixed, err)
	}
	var included, oldVersions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1`, owner).Scan(&included); err != nil {
		t.Fatal(err)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1 AND account_id='account-1')+
		(SELECT count(*) FROM portfolio_position_versions WHERE user_id=$1 AND account_id='account-1')+
		(SELECT count(*) FROM portfolio_activity_versions WHERE user_id=$1 AND account_id='account-1')`, owner).Scan(&oldVersions); err != nil {
		t.Fatal(err)
	}
	if included != 0 || oldVersions != 0 {
		t.Fatalf("removal was not immediate: included=%d versions=%d", included, oldVersions)
	}
	if _, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, older.ChangeID, older.Version, older.InventoryGeneration, older.LifecycleGeneration, map[string]portfolio.AccountData{"account-2": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(3*time.Second)); err != nil || accepted {
		t.Fatalf("older finalizer accepted=%v err=%v", accepted, err)
	}
	var olderStatus, olderFailure string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT status,failure_reason FROM portfolio_inclusion_changes WHERE id=$1`, older.ChangeID).Scan(&olderStatus, &olderFailure); err != nil {
		t.Fatal(err)
	}
	if olderStatus != "failed" || olderFailure != "stale_guard" {
		t.Fatalf("older finalizer status=%q failure=%q", olderStatus, olderFailure)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT (SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1)+(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1)+(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1)+(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1)`, owner).Scan(&included); err != nil {
		t.Fatal(err)
	}
	if included != 0 {
		t.Fatalf("older finalizer recreated removed coverage: rows=%d", included)
	}
	failed, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, mixed.ChangeID, mixed.Version, mixed.InventoryGeneration, mixed.LifecycleGeneration, nil, "provider_unavailable", fixture.now.Add(4*time.Second))
	if err != nil || accepted || len(failed.Committed) != 0 || failed.Change == nil || failed.Change.Status != portfolio.InclusionFailed {
		t.Fatalf("failed=%+v accepted=%v err=%v", failed, accepted, err)
	}

	other := inclusionOwnerWithInventory(t, fixture, 202, "subject-two")
	otherSnapshot, err := repository.GetInclusion(fixture.ctx, other)
	if err != nil || otherSnapshot.Version != 0 || len(otherSnapshot.Committed) != 0 {
		t.Fatalf("other owner observed state: %+v err=%v", otherSnapshot, err)
	}
}

func TestInclusionRepositoryPublishesThreeDatasetsAtomicallyAndCanRecover(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 211, "atomic-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{
		{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
		{ID: "account-2", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Second (•••• 1002)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
	})
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "atomic", []string{"account-1", "account-2"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	first, second := completeRepositoryAccountData(fixture.now), completeRepositoryAccountData(fixture.now)
	badCurrency := "1"
	second.Balances.Rows = []portfolio.Balance{{Currency: "US", Cash: &badCurrency}}
	if _, _, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": first, "account-2": second}, "", fixture.now); err == nil {
		t.Fatal("invalid second account unexpectedly published")
	}
	var rows int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_position_versions WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_activity_versions WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1)`, owner).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("partial dataset publication rows=%d", rows)
	}
	second = completeRepositoryAccountData(fixture.now)
	result, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": first, "account-2": second}, "", fixture.now.Add(time.Second))
	if err != nil || !accepted || len(result.Committed) != 2 {
		t.Fatalf("recovery result=%+v accepted=%v err=%v", result, accepted, err)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1)`, owner).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 8 {
		t.Fatalf("atomic memberships and three heads rows=%d, want 8", rows)
	}
}

func TestInclusionRepositoryPublishesTypedRowsThroughDatasetHeads(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 213, "typed-row-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "typed-rows", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	tradeDate := fixture.now.Add(-24 * time.Hour)
	cash, buyingPower := "100.123456789", "150.5"
	units, price, costBasis := "1.25", "123.456789", "100.01"
	amount, fee, activityPrice, activityUnits := "-10.25", "0.125", "9.75", "2.5"
	data := portfolio.AccountData{
		Balances:   portfolio.BalanceDataset{RetrievedAt: fixture.now, Rows: []portfolio.Balance{{Currency: "USD", Cash: &cash, BuyingPower: &buyingPower}}},
		Positions:  portfolio.PositionDataset{ObservedAt: fixture.now.Add(-time.Minute), RetrievedAt: fixture.now, Rows: []portfolio.Position{{InstrumentID: "instrument-1", Symbol: "AAPL", Kind: "stock", Currency: "USD", Units: &units, Price: &price, CostBasis: &costBasis}}},
		Activities: portfolio.ActivityDataset{RetrievedAt: fixture.now, Rows: []portfolio.Activity{{ID: "activity-1", Type: "BUY", TradeDate: &tradeDate, Currency: "CAD", Amount: &amount, Fee: &fee, Price: &activityPrice, Units: &activityUnits}}},
	}
	if _, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": data}, "", fixture.now); err != nil || !accepted {
		t.Fatalf("accepted=%v err=%v", accepted, err)
	}

	var balanceCurrency, storedCash, storedBuyingPower string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT row.currency,row.cash::text,row.buying_power::text FROM portfolio_balance_heads head JOIN portfolio_balance_rows row ON row.version_id=head.version_id WHERE head.user_id=$1 AND head.account_id='account-1'`, owner).Scan(&balanceCurrency, &storedCash, &storedBuyingPower); err != nil {
		t.Fatal(err)
	}
	if balanceCurrency != "USD" || storedCash != cash || storedBuyingPower != buyingPower {
		t.Fatalf("balance row currency=%q cash=%q buying_power=%q", balanceCurrency, storedCash, storedBuyingPower)
	}
	var instrumentID, symbol, kind, positionCurrency, storedUnits, storedPrice, storedCostBasis string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT row.instrument_id,row.symbol,row.kind,row.currency,row.units::text,row.price::text,row.cost_basis::text FROM portfolio_position_heads head JOIN portfolio_position_rows row ON row.version_id=head.version_id WHERE head.user_id=$1 AND head.account_id='account-1'`, owner).Scan(&instrumentID, &symbol, &kind, &positionCurrency, &storedUnits, &storedPrice, &storedCostBasis); err != nil {
		t.Fatal(err)
	}
	if instrumentID != "instrument-1" || symbol != "AAPL" || kind != "stock" || positionCurrency != "USD" || storedUnits != units || storedPrice != price || storedCostBasis != costBasis {
		t.Fatalf("position row id=%q symbol=%q kind=%q currency=%q units=%q price=%q cost=%q", instrumentID, symbol, kind, positionCurrency, storedUnits, storedPrice, storedCostBasis)
	}
	var activityID, activityType, activityCurrency, storedAmount, storedFee, storedActivityPrice, storedActivityUnits string
	var storedTradeDate time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT row.activity_id,row.activity_type,row.trade_date,row.currency,row.amount::text,row.fee::text,row.price::text,row.units::text FROM portfolio_activity_heads head JOIN portfolio_activity_rows row ON row.version_id=head.version_id WHERE head.user_id=$1 AND head.account_id='account-1'`, owner).Scan(&activityID, &activityType, &storedTradeDate, &activityCurrency, &storedAmount, &storedFee, &storedActivityPrice, &storedActivityUnits); err != nil {
		t.Fatal(err)
	}
	if activityID != "activity-1" || activityType != "BUY" || !storedTradeDate.Equal(tradeDate) || activityCurrency != "CAD" || storedAmount != amount || storedFee != fee || storedActivityPrice != activityPrice || storedActivityUnits != activityUnits {
		t.Fatalf("activity row id=%q type=%q date=%v currency=%q amount=%q fee=%q price=%q units=%q", activityID, activityType, storedTradeDate, activityCurrency, storedAmount, storedFee, storedActivityPrice, storedActivityUnits)
	}
}

func TestReauthorizationFencesPendingInclusionFinalizer(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 221, "reauthorization-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "before-reauthorization", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}

	attempt := fixture.createAttempt(t, 223, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, attempt)
	reauthorized := finalization(attempt.StateHash, owner, fixture.now.Add(time.Minute), 225)
	reauthorized.Subject = "reauthorization-owner"
	if err := fixture.repository.FinalizeCallback(fixture.ctx, reauthorized); err != nil {
		t.Fatal(err)
	}
	var lifecycle int64
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1`, owner).Scan(&lifecycle); err != nil {
		t.Fatal(err)
	}
	if lifecycle != prepared.LifecycleGeneration+1 {
		t.Fatalf("reauthorized lifecycle=%d old=%d", lifecycle, prepared.LifecycleGeneration)
	}
	if _, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(2*time.Minute)); err != nil || accepted {
		t.Fatalf("old finalizer accepted=%v err=%v", accepted, err)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT (SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1)+(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1)+(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1)+(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1)`, owner).Scan(&lifecycle); err != nil {
		t.Fatal(err)
	}
	if lifecycle != 0 {
		t.Fatalf("reauthorization-stale finalizer recreated rows=%d", lifecycle)
	}
}

func TestInclusionRepositorySupersedesDurablePendingChangeWithoutBroadening(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 231, "pending-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	older, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "pending-original", []string{"account-1"}, fixture.now)
	if err != nil || !older.Claimed {
		t.Fatalf("older=%+v err=%v", older, err)
	}
	replacement, err := repository.PrepareInclusion(fixture.ctx, owner, older.Version, "pending-replacement", []string{"account-1"}, fixture.now.Add(time.Second))
	if err != nil || !replacement.Claimed || replacement.Version != older.Version+1 || len(replacement.Additions) != 1 || replacement.Additions[0] != "account-1" || len(replacement.Committed) != 0 {
		t.Fatalf("replacement=%+v err=%v", replacement, err)
	}
	if _, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, older.ChangeID, older.Version, older.InventoryGeneration, older.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(2*time.Second)); err != nil || accepted {
		t.Fatalf("superseded finalizer accepted=%v err=%v", accepted, err)
	}
	result, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, replacement.ChangeID, replacement.Version, replacement.InventoryGeneration, replacement.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": completeRepositoryAccountData(fixture.now)}, "", fixture.now.Add(3*time.Second))
	if err != nil || !accepted || len(result.Committed) != 1 || result.Committed[0] != "account-1" {
		t.Fatalf("replacement result=%+v accepted=%v err=%v", result, accepted, err)
	}
}

func completeRepositoryAccountData(now time.Time) portfolio.AccountData {
	return portfolio.AccountData{
		Balances:   portfolio.BalanceDataset{RetrievedAt: now, Rows: []portfolio.Balance{}},
		Positions:  portfolio.PositionDataset{ObservedAt: now, RetrievedAt: now, Rows: []portfolio.Position{}},
		Activities: portfolio.ActivityDataset{RetrievedAt: now, Rows: []portfolio.Activity{}},
	}
}

func TestInclusionRepositoryRechecksLegacySelectionAndPreservesCommittedAccounts(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 201, "legacy-selection-owner")
	excluded := publishLegacyEligibilityInventory(t, fixture, owner)
	repository := postgresadapter.NewInclusionRepository(fixture.pool)
	for _, id := range excluded {
		t.Run(id, func(t *testing.T) {
			if _, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "reject-"+id, []string{"open-investment", id}, fixture.now); !errors.Is(err, portfolio.ErrInvalidAccountSelection) {
				t.Fatalf("legacy selectable account %q admission error=%v", id, err)
			}
		})
	}
	before, err := repository.GetInclusion(fixture.ctx, owner)
	if err != nil || before.Version != 0 || len(before.Committed) != 0 {
		t.Fatalf("rejected selections changed membership: %+v err=%v", before, err)
	}
	targets := []string{"open-investment", "unspecified-status-investment"}
	prepared, err := repository.PrepareInclusion(fixture.ctx, owner, 0, "admit-investments", targets, fixture.now)
	if err != nil || !prepared.Claimed || len(prepared.Additions) != 2 {
		t.Fatalf("approved investments preparation=%+v err=%v", prepared, err)
	}
	data := map[string]portfolio.AccountData{}
	for _, id := range targets {
		data[id] = completeRepositoryAccountData(fixture.now)
	}
	committed, accepted, err := repository.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, data, "", fixture.now)
	if err != nil || !accepted || !slices.Equal(committed.Committed, targets) {
		t.Fatalf("approved investments committed=%+v accepted=%v err=%v", committed, accepted, err)
	}

	// An existing choice may become ineligible under a later policy/snapshot.
	// Hide it from inventory, but preserve membership and removal-only changes.
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_accounts SET category='unknown' WHERE user_id=$1 AND account_id=ANY($2)`, owner, targets); err != nil {
		t.Fatal(err)
	}
	reloaded, err := postgresadapter.NewInventoryRepository(fixture.pool).Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	for _, connection := range reloaded.Connections {
		if len(connection.Accounts) != 0 {
			t.Fatalf("excluded committed accounts remained in inventory: %+v", connection.Accounts)
		}
	}
	retained, err := repository.PrepareInclusion(fixture.ctx, owner, committed.Version, "retain-existing", targets, fixture.now)
	if err != nil || retained.Claimed || !slices.Equal(retained.Committed, targets) {
		t.Fatalf("existing selection was not preserved: %+v err=%v", retained, err)
	}
	removed, err := repository.PrepareInclusion(fixture.ctx, owner, retained.Version, "remove-existing", []string{}, fixture.now)
	if err != nil || removed.Claimed || len(removed.Committed) != 0 || removed.Change.Status != portfolio.InclusionCommitted {
		t.Fatalf("removal-only cleanup failed: %+v err=%v", removed, err)
	}
	assertNoPublishedInclusionRows(t, fixture, owner)
	if _, err := repository.PrepareInclusion(fixture.ctx, owner, removed.Version, "reject-readmission", targets, fixture.now); !errors.Is(err, portfolio.ErrInvalidAccountSelection) {
		t.Fatalf("excluded accounts could be re-added after removal: %v", err)
	}
}

func publishInclusionInventory(t *testing.T, fixture *repositoryFixture, owner uuid.UUID, accounts []portfolio.Account) {
	t.Helper()
	inventory := postgresadapter.NewInventoryRepository(fixture.pool)
	claim, err := inventory.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("inventory claim=%+v err=%v", claim, err)
	}
	connections := []portfolio.Connection{{ID: "connection", BrokerageLabel: "Broker", Status: portfolio.ConnectionStatusActive, SyncMode: portfolio.SyncModeRealtime, Available: true, Eligible: true, Accounts: accounts}}
	if _, published, err := inventory.Finalize(fixture.ctx, owner, claim.Generation, portfolio.StateReady, nil, connections, fixture.now); err != nil || !published {
		t.Fatalf("inventory publication=%v err=%v", published, err)
	}
}

func assertNoPublishedInclusionRows(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) {
	t.Helper()
	var rows int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_position_heads WHERE user_id=$1)+
		(SELECT count(*) FROM portfolio_activity_heads WHERE user_id=$1)`, owner).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("stale inclusion publication rows=%d", rows)
	}
}

func inclusionOwnerWithInventory(t *testing.T, fixture *repositoryFixture, marker byte, subject string) uuid.UUID {
	t.Helper()
	attempt := fixture.createAttempt(t, marker, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, attempt)
	owner := uuid.New()
	final := finalization(attempt.StateHash, owner, fixture.now, marker+1)
	final.Subject = subject
	if err := fixture.repository.FinalizeCallback(fixture.ctx, final); err != nil {
		t.Fatal(err)
	}
	return owner
}
