package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestShowcaseRepositoryReadsOnlyOwnerCommittedHeadsAndFailsClosed(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 181, "showcase-owner")
	publishInclusionInventory(t, fixture, owner, []portfolio.Account{
		{ID: "included", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Included (•••• 1001)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
		{ID: "excluded", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Excluded (•••• 1002)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete},
	})

	cash, units, amount := "100.123400", "2.5000", "4.2500"
	tradeDate := fixture.now.Add(-24 * time.Hour)
	data := portfolio.AccountData{
		Balances:   portfolio.BalanceDataset{RetrievedAt: fixture.now, Rows: []portfolio.Balance{{Currency: "CAD", Cash: &cash}, {Currency: "USD"}}},
		Positions:  portfolio.PositionDataset{ObservedAt: fixture.now, RetrievedAt: fixture.now, Rows: []portfolio.Position{{InstrumentID: "instrument", Symbol: "FND", Kind: "equity", Currency: "USD", Units: &units}}},
		Activities: portfolio.ActivityDataset{RetrievedAt: fixture.now, Rows: []portfolio.Activity{{ID: "activity", Type: "DIVIDEND", TradeDate: &tradeDate, Currency: "CAD", Amount: &amount}}},
	}
	inclusion := postgresadapter.NewInclusionRepository(fixture.pool)
	prepared, err := inclusion.PrepareInclusion(fixture.ctx, owner, 0, "showcase", []string{"included"}, fixture.now)
	if err != nil || !prepared.Claimed {
		t.Fatalf("prepare=%+v err=%v", prepared, err)
	}
	repository := postgresadapter.NewShowcaseRepository(fixture.pool, func() time.Time { return fixture.now })
	pendingShowcase, err := repository.GetShowcase(fixture.ctx, owner)
	if err != nil || len(pendingShowcase.Accounts) != 1 || pendingShowcase.Accounts[0].Label != "Included (•••• 1001)" || pendingShowcase.Accounts[0].Balances.Context.Freshness != portfolio.FreshnessUnavailable {
		t.Fatalf("pending selection was not immediately visible: showcase=%+v err=%v", pendingShowcase, err)
	}
	if _, accepted, err := inclusion.FinalizeInclusion(fixture.ctx, owner, prepared.ChangeID, prepared.Version, prepared.InventoryGeneration, prepared.LifecycleGeneration, map[string]portfolio.AccountData{"included": data}, "", nil, fixture.now); err != nil || !accepted {
		t.Fatalf("accepted=%v err=%v", accepted, err)
	}

	other := inclusionOwnerWithInventory(t, fixture, 183, "other-showcase-owner")
	publishInclusionInventory(t, fixture, other, []portfolio.Account{{ID: "included", Category: portfolio.AccountCategoryInvestment, Type: "Margin", MaskedLabel: "Other owner (•••• 9999)", Available: true, Eligible: true, Selectable: true, UsabilityReason: portfolio.UsabilityReady, SyncState: portfolio.AccountSyncStateComplete}})
	otherPrepared, err := inclusion.PrepareInclusion(fixture.ctx, other, 0, "other-showcase", []string{"included"}, fixture.now)
	if err != nil || !otherPrepared.Claimed {
		t.Fatalf("other prepare=%+v err=%v", otherPrepared, err)
	}
	if _, accepted, err := inclusion.FinalizeInclusion(fixture.ctx, other, otherPrepared.ChangeID, otherPrepared.Version, otherPrepared.InventoryGeneration, otherPrepared.LifecycleGeneration, map[string]portfolio.AccountData{"included": completeRepositoryAccountData(fixture.now)}, "", nil, fixture.now); err != nil || !accepted {
		t.Fatalf("other accepted=%v err=%v", accepted, err)
	}

	historyID := uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_versions (id,user_id,account_id,inclusion_version,retrieved_at,published_at) VALUES ($1,$2,'included',99,$3,$3)`, historyID, owner, fixture.now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_rows (version_id,row_number,currency,cash) VALUES ($1,0,'USD',999999)`, historyID); err != nil {
		t.Fatal(err)
	}

	showcase, err := repository.GetShowcase(fixture.ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(showcase.Accounts) != 1 || showcase.Accounts[0].Label != "Included (•••• 1001)" {
		t.Fatalf("owner-scoped accounts=%+v", showcase.Accounts)
	}
	account := showcase.Accounts[0]
	if account.ConnectionID != "connection" {
		t.Fatalf("connection ID=%q", account.ConnectionID)
	}
	if account.Balances.Context.Currency != "CAD, USD" || len(account.Balances.Balances) != 2 || account.Balances.Balances[0].Cash == nil || *account.Balances.Balances[0].Cash != cash || account.Balances.Balances[0].BuyingPower != nil || account.Balances.Balances[1].Cash != nil {
		t.Fatalf("current balance head=%+v", account.Balances)
	}
	if account.Positions.Context.Currency != "USD" || len(account.Positions.Positions) != 1 || account.Positions.Positions[0].Units == nil || *account.Positions.Positions[0].Units != units || account.Positions.Positions[0].Price != nil || account.Activities.Context.Currency != "CAD" || len(account.Activities.Activities) != 1 || account.Activities.Activities[0].Amount == nil || *account.Activities.Activities[0].Amount != amount || account.Activities.Activities[0].Fee != nil {
		t.Fatalf("typed datasets positions=%+v activities=%+v", account.Positions, account.Activities)
	}
	var ownerBalanceVersion, otherBalanceVersion uuid.UUID
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT version_id FROM portfolio_balance_heads WHERE user_id=$1 AND account_id='included'`, owner).Scan(&ownerBalanceVersion); err != nil {
		t.Fatal(err)
	}
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT version_id FROM portfolio_balance_heads WHERE user_id=$1 AND account_id='included'`, other).Scan(&otherBalanceVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_balance_heads SET version_id=$3 WHERE user_id=$1 AND account_id=$2`, owner, "included", otherBalanceVersion); err != nil {
		t.Fatal(err)
	}
	showcase, err = repository.GetShowcase(fixture.ctx, owner)
	if err != nil || showcase.Accounts[0].Balances.Context.Freshness != portfolio.FreshnessUnavailable || len(showcase.Accounts[0].Balances.Balances) != 0 {
		t.Fatalf("foreign head pointer did not fail closed: showcase=%+v err=%v", showcase, err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_balance_heads SET version_id=$3 WHERE user_id=$1 AND account_id=$2`, owner, "included", ownerBalanceVersion); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.pool.Exec(fixture.ctx, `DELETE FROM portfolio_position_heads WHERE user_id=$1 AND account_id='included'`, owner); err != nil {
		t.Fatal(err)
	}
	showcase, err = repository.GetShowcase(fixture.ctx, owner)
	if err != nil || showcase.Accounts[0].Positions.Context.Freshness != portfolio.FreshnessUnavailable || len(showcase.Accounts[0].Positions.Positions) != 0 {
		t.Fatalf("missing head did not fail closed: showcase=%+v err=%v", showcase, err)
	}

	for _, test := range []struct {
		name, state string
		reason      portfolio.ResourceDiagnosticReason
		action      portfolio.ResourceDiagnosticAction
	}{
		{name: "pending waits", state: string(portfolio.AccountSyncStatePending), reason: portfolio.DiagnosticSyncPending, action: portfolio.DiagnosticActionWait},
		{name: "unavailable retries", state: string(portfolio.AccountSyncStateUnavailable), reason: portfolio.DiagnosticProviderUnavailable, action: portfolio.DiagnosticActionRetry},
		{name: "unknown retries without invented cause", state: string(portfolio.AccountSyncStateUnknown), reason: portfolio.DiagnosticUnknown, action: portfolio.DiagnosticActionRetry},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_accounts SET sync_state=$3 WHERE user_id=$1 AND account_id=$2`, owner, "included", test.state); err != nil {
				t.Fatal(err)
			}
			showcase, err = repository.GetShowcase(fixture.ctx, owner)
			diagnostic := showcase.Accounts[0].Balances.Context.Diagnostic
			if err != nil || showcase.Accounts[0].Balances.Context.Freshness != portfolio.FreshnessUnavailable || len(showcase.Accounts[0].Balances.Balances) != 0 || diagnostic == nil || diagnostic.Reason != test.reason || diagnostic.RecommendedAction != test.action {
				t.Fatalf("state=%q showcase=%+v diagnostic=%+v err=%v", test.state, showcase, diagnostic, err)
			}
		})
	}

	if _, err := fixture.pool.Exec(fixture.ctx, `DELETE FROM portfolio_included_accounts WHERE user_id=$1 AND account_id='included'`, owner); err != nil {
		t.Fatal(err)
	}
	showcase, err = repository.GetShowcase(fixture.ctx, owner)
	if err != nil || len(showcase.Accounts) != 0 {
		t.Fatalf("removed account remained visible: showcase=%+v err=%v", showcase, err)
	}
}
