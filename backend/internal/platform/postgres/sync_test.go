package postgres_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestSyncRepositoryLeasesGloballyRetriesAndPublishesAccumulatedData(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 241, "sync-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	inclusions := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := inclusions.PrepareInclusion(fixture.ctx, owner, 0, "initial-sync", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	initialData := completeRepositoryAccountData(fixture.now)
	initialData.Activities.Rows = activityRange(fixture.now.Add(-time.Hour), 0, 50)
	if _, accepted, err := inclusions.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": initialData}, "", nil, fixture.now); err != nil || !accepted {
		t.Fatalf("initial publication accepted=%v err=%v", accepted, err)
	}
	staleAt := fixture.now.Add(-25 * time.Hour)
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_account_sync_state SET
		last_success_at=$3,balances_success_at=$3,positions_success_at=$3,activities_success_at=$3,updated_at=$3
		WHERE user_id=$1 AND account_id=$2`, owner, "account-1", staleAt); err != nil {
		t.Fatal(err)
	}

	first := postgresadapter.NewSyncRepository(fixture.pool)
	second := postgresadapter.NewSyncRepository(fixture.pool)
	start := make(chan struct{})
	results := make(chan bool, 2)
	claims := make(chan uuid.UUID, 2)
	var group sync.WaitGroup
	for _, repository := range []*postgresadapter.SyncRepository{first, second} {
		group.Add(1)
		go func(repo *postgresadapter.SyncRepository) {
			defer group.Done()
			<-start
			claim, acquired, err := repo.AcquireWorkerLease(fixture.ctx, fixture.now, time.Minute)
			if err != nil {
				t.Errorf("worker lease: %v", err)
			}
			results <- acquired
			claims <- claim
		}(repository)
	}
	close(start)
	group.Wait()
	close(results)
	close(claims)
	acquiredCount := 0
	var workerClaim uuid.UUID
	for acquired := range results {
		if acquired {
			acquiredCount++
		}
	}
	for claim := range claims {
		if claim != uuid.Nil {
			workerClaim = claim
		}
	}
	if acquiredCount != 1 {
		t.Fatalf("global worker leases=%d, want 1", acquiredCount)
	}
	if err := first.ReleaseWorkerLease(fixture.ctx, workerClaim, fixture.now); err != nil {
		t.Fatal(err)
	}

	accountClaim, err := first.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute)
	if err != nil || accountClaim == nil {
		t.Fatalf("account claim=%+v err=%v", accountClaim, err)
	}
	if accountClaim.Resource != portfolio.AccountResourceBalances {
		t.Fatalf("first resource=%q, want balances", accountClaim.Resource)
	}
	if duplicate, err := second.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute); err != nil || duplicate != nil {
		t.Fatalf("duplicate claim=%+v err=%v", duplicate, err)
	}
	var oldBalance, oldPosition, oldActivity uuid.UUID
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT balance.version_id,position.version_id,activity.version_id FROM portfolio_balance_heads balance JOIN portfolio_position_heads position USING(user_id,account_id) JOIN portfolio_activity_heads activity USING(user_id,account_id) WHERE balance.user_id=$1 AND balance.account_id=$2`, owner, "account-1").Scan(&oldBalance, &oldPosition, &oldActivity); err != nil {
		t.Fatal(err)
	}
	providerRetryAt := fixture.now.Add(11 * time.Minute)
	accepted, err := first.FinishSync(fixture.ctx, *accountClaim, nil, "rate_limited", &providerRetryAt, fixture.now)
	if err != nil || accepted {
		t.Fatalf("failed finish accepted=%v err=%v", accepted, err)
	}
	var balance, position, activity uuid.UUID
	var nextAttempt time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT balance.version_id,position.version_id,activity.version_id,sync.next_attempt_at FROM portfolio_balance_heads balance JOIN portfolio_position_heads position USING(user_id,account_id) JOIN portfolio_activity_heads activity USING(user_id,account_id) JOIN portfolio_account_sync_state sync USING(user_id,account_id) WHERE balance.user_id=$1 AND balance.account_id=$2`, owner, "account-1").Scan(&balance, &position, &activity, &nextAttempt); err != nil {
		t.Fatal(err)
	}
	if balance != oldBalance || position != oldPosition || activity != oldActivity || !nextAttempt.Equal(providerRetryAt) {
		t.Fatalf("failure changed heads or retry: heads=%v/%v/%v retry=%v", balance, position, activity, nextAttempt)
	}
	var balanceReason, balanceAction string
	var balanceRetryAt time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT balances_failure_reason,balances_failure_action,balances_retry_at
		FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id=$2`, owner, "account-1").Scan(&balanceReason, &balanceAction, &balanceRetryAt); err != nil {
		t.Fatal(err)
	}
	if balanceReason != string(portfolio.DiagnosticProviderUnavailable) || balanceAction != string(portfolio.DiagnosticActionWait) || !balanceRetryAt.Equal(providerRetryAt) {
		t.Fatalf("balance diagnostic=%s/%s retry=%v", balanceReason, balanceAction, balanceRetryAt)
	}
	degraded, err := postgresadapter.NewShowcaseRepository(fixture.pool, func() time.Time { return fixture.now }).GetShowcase(fixture.ctx, owner)
	if err != nil || len(degraded.Accounts) != 1 || degraded.Accounts[0].Balances.Context.Diagnostic == nil || degraded.Accounts[0].Balances.Context.Diagnostic.LastSuccessfulAt == nil {
		t.Fatalf("degraded showcase=%+v err=%v", degraded, err)
	}
	if diagnostic := degraded.Accounts[0].Balances.Context.Diagnostic; diagnostic.Reason != portfolio.DiagnosticProviderUnavailable || diagnostic.RecommendedAction != portfolio.DiagnosticActionWait || diagnostic.RetryAt == nil || !diagnostic.RetryAt.Equal(providerRetryAt) {
		t.Fatalf("balance diagnostic projection=%+v", diagnostic)
	}
	if early, err := first.ClaimDue(fixture.ctx, providerRetryAt.Add(-time.Nanosecond), 24*time.Hour, time.Minute); err != nil || early != nil {
		t.Fatalf("early=%+v err=%v", early, err)
	}
	preservedPositionRetryAt := providerRetryAt.Add(5 * time.Minute)
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_account_sync_state SET
		positions_failure_reason=$3,
		positions_failure_action=$4,
		positions_retry_at=$5
		WHERE user_id=$1 AND account_id=$2`, owner, "account-1", portfolio.DiagnosticProviderUnavailable, portfolio.DiagnosticActionRetry, preservedPositionRetryAt); err != nil {
		t.Fatal(err)
	}

	retry, err := first.ClaimDue(fixture.ctx, providerRetryAt, 24*time.Hour, time.Minute)
	if err != nil || retry == nil {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	refreshed := completeRepositoryAccountData(providerRetryAt)
	refreshed.Activities.Rows = activityRange(providerRetryAt, 25, 75)
	if accepted, err := first.FinishSync(fixture.ctx, *retry, &refreshed, "", nil, providerRetryAt); err != nil || !accepted {
		t.Fatalf("balance refresh accepted=%v err=%v", accepted, err)
	}
	var clearedBalanceReason, clearedBalanceAction, untouchedPositionReason, untouchedPositionAction *string
	var clearedBalanceRetryAt, untouchedPositionRetryAt *time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		balances_failure_reason,
		balances_failure_action,
		balances_retry_at,
		positions_failure_reason,
		positions_failure_action,
		positions_retry_at
		FROM portfolio_account_sync_state
		WHERE user_id=$1
		AND account_id=$2`, owner, "account-1").Scan(&clearedBalanceReason, &clearedBalanceAction, &clearedBalanceRetryAt, &untouchedPositionReason, &untouchedPositionAction, &untouchedPositionRetryAt); err != nil {
		t.Fatal(err)
	}
	if clearedBalanceReason != nil || clearedBalanceAction != nil || clearedBalanceRetryAt != nil ||
		untouchedPositionReason == nil || *untouchedPositionReason != string(portfolio.DiagnosticProviderUnavailable) ||
		untouchedPositionAction == nil || *untouchedPositionAction != string(portfolio.DiagnosticActionRetry) ||
		untouchedPositionRetryAt == nil || !untouchedPositionRetryAt.Equal(preservedPositionRetryAt) {
		t.Fatalf("resource-scoped clear balance=%v/%v/%v position=%v/%v/%v", clearedBalanceReason, clearedBalanceAction, clearedBalanceRetryAt, untouchedPositionReason, untouchedPositionAction, untouchedPositionRetryAt)
	}

	positionClaim, err := first.ClaimDue(fixture.ctx, providerRetryAt, 24*time.Hour, time.Minute)
	if err != nil || positionClaim == nil || positionClaim.Resource != portfolio.AccountResourcePositions {
		t.Fatalf("position claim=%+v err=%v", positionClaim, err)
	}
	if accepted, err := first.FinishSync(fixture.ctx, *positionClaim, nil, "provider_unavailable", nil, providerRetryAt); err != nil || accepted {
		t.Fatalf("position failure accepted=%v err=%v", accepted, err)
	}
	positionRetryAt := providerRetryAt.Add(time.Minute)
	var positionReason, positionAction string
	var positionDiagnosticRetry time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT positions_failure_reason,positions_failure_action,positions_retry_at
		FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id=$2`, owner, "account-1").Scan(&positionReason, &positionAction, &positionDiagnosticRetry); err != nil {
		t.Fatal(err)
	}
	if positionReason != string(portfolio.DiagnosticProviderUnavailable) || positionAction != string(portfolio.DiagnosticActionRetry) || !positionDiagnosticRetry.Equal(positionRetryAt) {
		t.Fatalf("position diagnostic=%s/%s retry=%v", positionReason, positionAction, positionDiagnosticRetry)
	}
	positionRetry, err := first.ClaimDue(fixture.ctx, positionRetryAt, 24*time.Hour, time.Minute)
	if err != nil || positionRetry == nil || positionRetry.Resource != portfolio.AccountResourcePositions {
		t.Fatalf("position retry=%+v err=%v", positionRetry, err)
	}
	if accepted, err := first.FinishSync(fixture.ctx, *positionRetry, &refreshed, "", nil, positionRetryAt); err != nil || !accepted {
		t.Fatalf("position refresh accepted=%v err=%v", accepted, err)
	}

	activityClaim, err := first.ClaimDue(fixture.ctx, positionRetryAt, 24*time.Hour, time.Minute)
	if err != nil || activityClaim == nil || activityClaim.Resource != portfolio.AccountResourceActivities {
		t.Fatalf("activity claim=%+v err=%v", activityClaim, err)
	}
	if accepted, err := first.FinishSync(fixture.ctx, *activityClaim, &refreshed, "", nil, positionRetryAt); err != nil || !accepted {
		t.Fatalf("activity refresh accepted=%v err=%v", accepted, err)
	}

	var uniqueActivities, balanceVersions, positionVersions, activityVersions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_account_activities WHERE user_id=$1 AND account_id=$2),
		(SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1 AND account_id=$2),
		(SELECT count(*) FROM portfolio_position_versions WHERE user_id=$1 AND account_id=$2),
		(SELECT count(*) FROM portfolio_activity_versions WHERE user_id=$1 AND account_id=$2)`, owner, "account-1").Scan(&uniqueActivities, &balanceVersions, &positionVersions, &activityVersions); err != nil {
		t.Fatal(err)
	}
	if uniqueActivities != 75 || balanceVersions != 2 || positionVersions != 2 || activityVersions != 2 {
		t.Fatalf("activities=%d versions=%d/%d/%d", uniqueActivities, balanceVersions, positionVersions, activityVersions)
	}
	showcase, err := postgresadapter.NewShowcaseRepository(fixture.pool, func() time.Time { return positionRetryAt }).GetShowcase(fixture.ctx, owner)
	if err != nil || len(showcase.Accounts) != 1 || len(showcase.Accounts[0].Activities.Activities) != 50 {
		t.Fatalf("showcase=%+v err=%v", showcase, err)
	}
	activities := showcase.Accounts[0].Activities.Activities
	newestDate := providerRetryAt.Add(74 * time.Minute)
	oldestVisibleDate := providerRetryAt.Add(25 * time.Minute)
	if activities[0].TradeDate == nil || !activities[0].TradeDate.Equal(newestDate) || activities[len(activities)-1].TradeDate == nil || !activities[len(activities)-1].TradeDate.Equal(oldestVisibleDate) {
		t.Fatalf("activity window first=%v last=%v", activities[0].TradeDate, activities[len(activities)-1].TradeDate)
	}
}

func TestSyncRepositoryPausesReauthorizationAndFreshCallbackResumesPendingClaim(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 247, "refresh-paused-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	prepared, err := postgresadapter.NewInclusionRepository(fixture.pool).PrepareInclusion(fixture.ctx, owner, 0, "paused-refresh", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_account_sync_state
		SET next_attempt_at=$3,failure_count=3 WHERE user_id=$1 AND account_id=$2`,
		owner, "account-1", fixture.now.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	credential, found, err := postgresadapter.NewCredentialRepository(fixture.pool).ReadCredential(fixture.ctx, owner)
	if err != nil || !found {
		t.Fatalf("credential found=%v err=%v", found, err)
	}
	if err := postgresadapter.NewCredentialRepository(fixture.pool).RequireReauthorization(fixture.ctx, owner, nil, credential.Version, fixture.now); err != nil {
		t.Fatal(err)
	}
	syncRepository := postgresadapter.NewSyncRepository(fixture.pool)
	if claim, err := syncRepository.ClaimDue(fixture.ctx, fixture.now, 24*time.Hour, time.Minute); err != nil || claim != nil {
		t.Fatalf("paused claim=%+v err=%v", claim, err)
	}

	attempt := fixture.createAttempt(t, 248, fixture.now.Add(10*time.Minute))
	fixture.claimCallback(t, attempt)
	fresh := finalization(attempt.StateHash, owner, fixture.now.Add(time.Minute), 249)
	fresh.Subject = "refresh-paused-owner"
	if err := fixture.repository.FinalizeCallback(fixture.ctx, fresh); err != nil {
		t.Fatal(err)
	}
	var nextAttempt *time.Time
	var failures int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT next_attempt_at,failure_count FROM portfolio_account_sync_state
		WHERE user_id=$1 AND account_id=$2`, owner, "account-1").Scan(&nextAttempt, &failures); err != nil {
		t.Fatal(err)
	}
	if nextAttempt != nil || failures != 0 {
		t.Fatalf("fresh grant retained sync backoff: next=%v failures=%d", nextAttempt, failures)
	}
	claim, err := syncRepository.ClaimDue(fixture.ctx, fixture.now.Add(time.Minute), 24*time.Hour, time.Minute)
	if err != nil || claim == nil || claim.ChangeID == nil || *claim.ChangeID != prepared.ChangeID {
		t.Fatalf("resumed claim=%+v err=%v", claim, err)
	}
}

func TestSyncRepositoryRepairsIncludedNullInitializationWithCompleteCycle(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 250, "null-initialization-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	inclusions := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := inclusions.PrepareInclusion(fixture.ctx, owner, 0, "null-initialization", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	data := completeRepositoryAccountData(fixture.now)
	if _, accepted, err := inclusions.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"account-1": data}, "", nil, fixture.now); err != nil || !accepted {
		t.Fatalf("initial publication accepted=%v err=%v", accepted, err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_account_sync_state
		SET initialized_at=NULL WHERE user_id=$1 AND account_id='account-1'`, owner); err != nil {
		t.Fatal(err)
	}

	repository := postgresadapter.NewSyncRepository(fixture.pool)
	cycleAt := fixture.now.Add(time.Minute)
	for _, resource := range []portfolio.AccountResource{portfolio.AccountResourceBalances, portfolio.AccountResourcePositions, portfolio.AccountResourceActivities} {
		claim, err := repository.ClaimDue(fixture.ctx, cycleAt, 24*time.Hour, time.Minute)
		if err != nil || claim == nil || claim.Resource != resource {
			t.Fatalf("resource=%s claim=%+v err=%v", resource, claim, err)
		}
		fresh := completeRepositoryAccountData(cycleAt)
		if accepted, err := repository.FinishSync(fixture.ctx, *claim, &fresh, "", nil, cycleAt); err != nil || !accepted {
			t.Fatalf("resource=%s accepted=%v err=%v", resource, accepted, err)
		}
	}
	var initializedAt *time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT initialized_at FROM portfolio_account_sync_state
		WHERE user_id=$1 AND account_id='account-1'`, owner).Scan(&initializedAt); err != nil {
		t.Fatal(err)
	}
	if initializedAt == nil || !initializedAt.Equal(cycleAt) {
		t.Fatalf("initializedAt=%v", initializedAt)
	}
}

func activityRange(start time.Time, first, end int) []portfolio.Activity {
	activities := make([]portfolio.Activity, 0, end-first)
	for index := first; index < end; index++ {
		tradeDate := start.Add(time.Duration(index) * time.Minute)
		activities = append(activities, portfolio.Activity{
			ID:        fmt.Sprintf("activity-%03d", index),
			Type:      "BUY",
			Currency:  "USD",
			TradeDate: &tradeDate,
		})
	}
	return activities
}

func TestSyncRepositoryResumesRetryableFirstInclusionAtIncompleteResource(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 242, "pending-sync-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{{ID: "account-1", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "First (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	inclusions := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := inclusions.PrepareInclusion(fixture.ctx, owner, 0, "pending-initial-sync", []string{"account-1"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepared=%+v err=%v", prepared, err)
	}
	partial := completeRepositoryAccountData(fixture.now)
	partial.Positions = portfolio.PositionDataset{}
	partial.Activities = portfolio.ActivityDataset{}
	retryAt := fixture.now.Add(5 * time.Minute)
	pending, accepted, err := inclusions.FinalizeInclusion(
		fixture.ctx,
		owner,
		prepared.ChangeID,
		prepared.Version,
		prepared.InventoryGeneration,
		prepared.LifecycleGeneration,
		map[string]portfolio.AccountData{"account-1": partial},
		"rate_limited",
		&retryAt,
		fixture.now,
	)
	if err != nil || accepted || pending.Change == nil || pending.Change.Status != portfolio.InclusionPending {
		t.Fatalf("pending=%+v accepted=%v err=%v", pending, accepted, err)
	}

	syncRepository := postgresadapter.NewSyncRepository(fixture.pool)
	if early, err := syncRepository.ClaimDue(fixture.ctx, retryAt.Add(-time.Nanosecond), 24*time.Hour, time.Minute); err != nil || early != nil {
		t.Fatalf("early=%+v err=%v", early, err)
	}
	positionClaim, err := syncRepository.ClaimDue(fixture.ctx, retryAt, 24*time.Hour, time.Minute)
	if err != nil || positionClaim == nil || positionClaim.Resource != portfolio.AccountResourcePositions || positionClaim.ChangeID == nil || *positionClaim.ChangeID != prepared.ChangeID {
		t.Fatalf("position claim=%+v err=%v", positionClaim, err)
	}
	complete := completeRepositoryAccountData(retryAt)
	if accepted, err := syncRepository.FinishSync(fixture.ctx, *positionClaim, &complete, "", nil, retryAt); err != nil || !accepted {
		t.Fatalf("position accepted=%v err=%v", accepted, err)
	}
	activityClaim, err := syncRepository.ClaimDue(fixture.ctx, retryAt, 24*time.Hour, time.Minute)
	if err != nil || activityClaim == nil || activityClaim.Resource != portfolio.AccountResourceActivities {
		t.Fatalf("activity claim=%+v err=%v", activityClaim, err)
	}
	if accepted, err := syncRepository.FinishSync(fixture.ctx, *activityClaim, &complete, "", nil, retryAt); err != nil || !accepted {
		t.Fatalf("activity accepted=%v err=%v", accepted, err)
	}
	committed, err := inclusions.GetInclusion(fixture.ctx, owner)
	if err != nil || len(committed.Committed) != 1 || committed.Committed[0] != "account-1" || committed.Change == nil || committed.Change.Status != portfolio.InclusionCommitted {
		t.Fatalf("committed=%+v err=%v", committed, err)
	}
	var balanceVersions int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1 AND account_id=$2`, owner, "account-1").Scan(&balanceVersions); err != nil {
		t.Fatal(err)
	}
	if balanceVersions != 1 {
		t.Fatalf("balance versions=%d, successful resource was repeated", balanceVersions)
	}
}
