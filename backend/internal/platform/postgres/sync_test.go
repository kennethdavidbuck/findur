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
	if early, err := first.ClaimDue(fixture.ctx, providerRetryAt.Add(-time.Nanosecond), 24*time.Hour, time.Minute); err != nil || early != nil {
		t.Fatalf("early=%+v err=%v", early, err)
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

	positionClaim, err := first.ClaimDue(fixture.ctx, providerRetryAt, 24*time.Hour, time.Minute)
	if err != nil || positionClaim == nil || positionClaim.Resource != portfolio.AccountResourcePositions {
		t.Fatalf("position claim=%+v err=%v", positionClaim, err)
	}
	if accepted, err := first.FinishSync(fixture.ctx, *positionClaim, nil, "provider_unavailable", nil, providerRetryAt); err != nil || accepted {
		t.Fatalf("position failure accepted=%v err=%v", accepted, err)
	}
	positionRetryAt := providerRetryAt.Add(time.Minute)
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
