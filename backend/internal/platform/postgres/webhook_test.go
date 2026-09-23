package postgres_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestWebhookRepositoryAppliesLifecycleAtomicallyAndIdempotently(t *testing.T) {
	fixture := newRepositoryFixture(t)
	tests := []struct {
		name            string
		event           portfolio.WebhookEvent
		wantConnections int
		wantAccounts    int
		wantState       portfolio.State
	}{
		{name: "connection broken", event: portfolio.WebhookEvent{Type: portfolio.WebhookConnectionBroken, ConnectionID: "connection"}, wantConnections: 1, wantAccounts: 1, wantState: portfolio.StateDisabled},
		{name: "connection deleted", event: portfolio.WebhookEvent{Type: portfolio.WebhookConnectionDeleted, ConnectionID: "connection"}, wantConnections: 0, wantAccounts: 0, wantState: portfolio.StateEmpty},
		{name: "account removed", event: portfolio.WebhookEvent{Type: portfolio.WebhookAccountRemoved, AccountID: "account"}, wantConnections: 1, wantAccounts: 0, wantState: portfolio.StateEmpty},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture.reset(t)
			subject := "webhook-removal-" + test.name
			owner := webhookOwnerWithInventory(t, fixture, byte(20+index), subject)
			seedWebhookInclusionAndHistory(t, fixture, owner)
			repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
			service, err := portfolio.NewWebhookService(repository, func() time.Time { return fixture.now.Add(time.Minute) })
			if err != nil {
				t.Fatal(err)
			}
			test.event.Subject = subject
			changed, err := service.Apply(fixture.ctx, test.event)
			if err != nil || !changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
			assertWebhookRemovalState(t, fixture, owner)
			assertCurrentInventoryRows(t, fixture, owner, test.wantConnections, test.wantAccounts)
			assertWebhookInventoryStates(t, fixture, owner, test.wantState, portfolio.StateReady)

			changed, err = service.Apply(fixture.ctx, test.event)
			if err != nil || changed {
				t.Fatalf("duplicate changed=%v err=%v", changed, err)
			}
			assertWebhookRemovalState(t, fixture, owner)
			assertCurrentInventoryRows(t, fixture, owner, test.wantConnections, test.wantAccounts)
			assertWebhookInventoryStates(t, fixture, owner, test.wantState, portfolio.StateReady)
		})
	}
}

func assertWebhookInventoryStates(t *testing.T, fixture *repositoryFixture, owner uuid.UUID, wantCurrent, wantHistorical portfolio.State) {
	t.Helper()
	var current, historical portfolio.State
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT state.current_status,version.status
		FROM portfolio_inventory_state state
		JOIN portfolio_inventory_versions version
			ON version.user_id=state.user_id
			AND version.generation=state.head_generation
		WHERE state.user_id=$1`, owner).Scan(&current, &historical); err != nil {
		t.Fatal(err)
	}
	if current != wantCurrent || historical != wantHistorical {
		t.Fatalf("current state=%s historical state=%s; want %s/%s", current, historical, wantCurrent, wantHistorical)
	}
}

func assertCurrentInventoryRows(t *testing.T, fixture *repositoryFixture, owner uuid.UUID, wantConnections, wantAccounts int) {
	t.Helper()
	var connections, accounts int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_inventory_connections WHERE user_id=$1 AND generation=1),
		(SELECT count(*) FROM portfolio_inventory_accounts WHERE user_id=$1 AND generation=1)`, owner).Scan(&connections, &accounts); err != nil {
		t.Fatal(err)
	}
	if connections != wantConnections || accounts != wantAccounts {
		t.Fatalf("current inventory connections=%d accounts=%d; want %d/%d", connections, accounts, wantConnections, wantAccounts)
	}
}

func TestWebhookRepositoryCreatesOnlySafeProvisionalRows(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-provisional"
	owner := webhookOwnerWithInventory(t, fixture, 40, subject)
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	service, err := portfolio.NewWebhookService(repository, func() time.Time { return fixture.now.Add(time.Minute) })
	if err != nil {
		t.Fatal(err)
	}
	event := portfolio.WebhookEvent{Subject: subject, Type: portfolio.WebhookNewAccountAvailable, ConnectionID: "new-connection", AccountID: "new-account"}
	changed, err := service.Apply(fixture.ctx, event)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	var status, syncMode, category, syncState, reason string
	var connectionAvailable, connectionEligible, accountAvailable, accountEligible, selectable bool
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT connection.status,connection.sync_mode,connection.available,connection.eligible,
		account.category,account.sync_state,account.available,account.eligible,account.selectable,account.usability_reason
		FROM portfolio_inventory_state state
		JOIN portfolio_inventory_connections connection ON connection.user_id=state.user_id AND connection.generation=state.head_generation
		JOIN portfolio_inventory_accounts account ON account.user_id=connection.user_id AND account.generation=connection.generation AND account.connection_id=connection.connection_id
		WHERE state.user_id=$1 AND account.account_id='new-account'`, owner).Scan(
		&status, &syncMode, &connectionAvailable, &connectionEligible,
		&category, &syncState, &accountAvailable, &accountEligible, &selectable, &reason,
	); err != nil {
		t.Fatal(err)
	}
	if status != "unavailable" || syncMode != "unknown" || connectionAvailable || connectionEligible || category != "unknown" || syncState != "unknown" || accountAvailable || accountEligible || selectable || reason != "provisional_category" {
		t.Fatalf("unsafe provisional row: status=%s sync=%s connection=%v/%v category=%s account_sync=%s account=%v/%v selectable=%v reason=%s", status, syncMode, connectionAvailable, connectionEligible, category, syncState, accountAvailable, accountEligible, selectable, reason)
	}
	changed, err = service.Apply(fixture.ctx, event)
	if err != nil || changed {
		t.Fatalf("duplicate changed=%v err=%v", changed, err)
	}
	preparation, err := postgresadapter.NewInventoryRepository(fixture.pool).Prepare(fixture.ctx, owner, false, fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	var projected *portfolio.Account
	for _, connection := range preparation.Connections {
		for index := range connection.Accounts {
			if connection.Accounts[index].ID == "new-account" {
				projected = &connection.Accounts[index]
			}
		}
	}
	if projected == nil || projected.Selectable || projected.MaskedLabel != "Account details unavailable" {
		t.Fatalf("projected provisional account=%+v", projected)
	}
}

func TestWebhookRepositoryAddsConnectionAndAcceptsExpiredClaim(t *testing.T) {
	tests := []struct {
		name         string
		expiredClaim bool
		event        portfolio.WebhookEvent
		wantState    portfolio.State
	}{
		{name: "connection added", event: portfolio.WebhookEvent{Type: portfolio.WebhookConnectionAdded, ConnectionID: "new-connection"}, wantState: portfolio.StateReady},
		{name: "expired claim", expiredClaim: true, event: portfolio.WebhookEvent{Type: portfolio.WebhookConnectionBroken, ConnectionID: "connection"}, wantState: portfolio.StateDisabled},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newRepositoryFixture(t)
			fixture.reset(t)
			subject := "webhook-state-" + test.name
			owner := webhookOwnerWithInventory(t, fixture, byte(70+index), subject)
			if test.expiredClaim {
				if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_state
					SET current_generation=current_generation+1,
						current_status='pending',
						claim_expires_at=$2
					WHERE user_id=$1`, owner, fixture.now.Add(-time.Second)); err != nil {
					t.Fatal(err)
				}
			}
			service, err := portfolio.NewWebhookService(postgresadapter.NewInventoryLifecycleRepository(fixture.pool), func() time.Time { return fixture.now })
			if err != nil {
				t.Fatal(err)
			}
			test.event.Subject = subject
			changed, err := service.Apply(fixture.ctx, test.event)
			if err != nil || !changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
			assertWebhookInventoryStates(t, fixture, owner, test.wantState, portfolio.StateReady)
			if test.event.Type == portfolio.WebhookConnectionAdded {
				var status string
				var available, eligible bool
				if err := fixture.pool.QueryRow(fixture.ctx, `SELECT status,available,eligible
					FROM portfolio_inventory_connections
					WHERE user_id=$1 AND generation=1 AND connection_id='new-connection'`, owner).Scan(&status, &available, &eligible); err != nil {
					t.Fatal(err)
				}
				if status != "unavailable" || available || eligible {
					t.Fatalf("status=%s available=%v eligible=%v", status, available, eligible)
				}
			}
		})
	}
}

func TestWebhookRepositoryRejectsUnavailableOwners(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *repositoryFixture, uuid.UUID)
	}{
		{name: "known owner without inventory state"},
		{name: "known owner without inventory head", setup: func(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) {
			inventory := postgresadapter.NewInventoryRepository(fixture.pool)
			claim, err := inventory.Prepare(fixture.ctx, owner, false, fixture.now)
			if err != nil || !claim.Claimed {
				t.Fatalf("claim=%+v err=%v", claim, err)
			}
			if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_state
				SET claim_expires_at=$2
				WHERE user_id=$1`, owner, fixture.now.Add(-time.Second)); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newRepositoryFixture(t)
			fixture.reset(t)
			subject := "webhook-unavailable-owner-" + test.name
			owner := inclusionOwnerWithInventory(t, fixture, byte(80+index), subject)
			if test.setup != nil {
				test.setup(t, fixture, owner)
			}
			repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
			_, err := repository.AddConnection(fixture.ctx, subject, "new-connection", fixture.now)
			if !errors.Is(err, portfolio.ErrInventoryBusy) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestWebhookRepositoryRejectsInactiveOwner(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-inactive"
	owner := webhookOwnerWithInventory(t, fixture, 90, subject)
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE users SET active=false WHERE id=$1`, owner); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	changed, err := repository.SetConnectionStatus(fixture.ctx, subject, "connection", portfolio.ConnectionStatusDisabled, fixture.now)
	if err != nil || changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	var status string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT status
		FROM portfolio_inventory_connections
		WHERE user_id=$1 AND generation=1 AND connection_id='connection'`, owner).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "active" {
		t.Fatalf("status=%s", status)
	}
}

func TestWebhookRepositoryReassociatesAccountWithoutLosingMetadata(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-reassociation"
	owner := webhookOwnerWithInventory(t, fixture, 95, subject)
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inventory_connections
		(user_id,generation,connection_id,brokerage_label,status,sync_mode,available,eligible)
		VALUES ($1,1,'replacement-connection','Replacement Broker','active','realtime',true,true)`, owner); err != nil {
		t.Fatal(err)
	}
	eventTime := fixture.now.Add(time.Minute)
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	changed, err := repository.AddAccount(fixture.ctx, subject, "replacement-connection", "account", eventTime)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	var inventoryConnection, identityConnection, category, accountType, label, syncState, reason string
	var available, eligible, selectable bool
	var firstSeen, lastSeen time.Time
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
			account.connection_id,
			identity.connection_id,
			account.category,
			account.account_type,
			account.masked_label,
			account.sync_state,
			account.usability_reason,
			account.available,
			account.eligible,
			account.selectable,
			identity.first_seen_at,
			identity.last_seen_at
		FROM portfolio_inventory_accounts account
		JOIN portfolio_account_identities identity
			ON identity.user_id=account.user_id
			AND identity.account_id=account.account_id
		WHERE account.user_id=$1 AND account.generation=1 AND account.account_id='account'`, owner).Scan(
		&inventoryConnection, &identityConnection, &category, &accountType, &label, &syncState, &reason,
		&available, &eligible, &selectable, &firstSeen, &lastSeen,
	); err != nil {
		t.Fatal(err)
	}
	if inventoryConnection != "replacement-connection" || identityConnection != "replacement-connection" ||
		category != "investment" || accountType != "Margin" || label != "Synthetic account" || syncState != "complete" || reason != "ready" ||
		!available || !eligible || !selectable || !firstSeen.Equal(fixture.now) || !lastSeen.Equal(eventTime) {
		t.Fatalf("account reassociation metadata changed: inventory=%s identity=%s category=%s type=%s label=%s sync=%s reason=%s flags=%v/%v/%v first=%s last=%s",
			inventoryConnection, identityConnection, category, accountType, label, syncState, reason, available, eligible, selectable, firstSeen, lastSeen)
	}
}

func TestWebhookRepositoryFixesConnectionIdempotently(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-fixed"
	owner := webhookOwnerWithInventory(t, fixture, 45, subject)
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	service, err := portfolio.NewWebhookService(repository, func() time.Time { return fixture.now.Add(time.Minute) })
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := service.Apply(fixture.ctx, portfolio.WebhookEvent{Subject: subject, Type: portfolio.WebhookConnectionBroken, ConnectionID: "connection"}); err != nil || !changed {
		t.Fatalf("broken changed=%v err=%v", changed, err)
	}
	assertBrokenAccountState(t, fixture, owner)
	assertWebhookInventoryStates(t, fixture, owner, portfolio.StateDisabled, portfolio.StateReady)
	fixed := portfolio.WebhookEvent{Subject: subject, Type: portfolio.WebhookConnectionFixed, ConnectionID: "connection"}
	if changed, err := service.Apply(fixture.ctx, fixed); err != nil || !changed {
		t.Fatalf("fixed changed=%v err=%v", changed, err)
	}
	if changed, err := service.Apply(fixture.ctx, fixed); err != nil || changed {
		t.Fatalf("duplicate fixed changed=%v err=%v", changed, err)
	}
	var status string
	var available, eligible bool
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT status,available,eligible FROM portfolio_inventory_connections WHERE user_id=$1 AND generation=1 AND connection_id='connection'`, owner).Scan(&status, &available, &eligible); err != nil {
		t.Fatal(err)
	}
	if status != "active" || !available || !eligible {
		t.Fatalf("status=%s available=%v eligible=%v", status, available, eligible)
	}
	assertBrokenAccountState(t, fixture, owner)
	assertWebhookInventoryStates(t, fixture, owner, portfolio.StateReady, portfolio.StateReady)
}

func assertBrokenAccountState(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) {
	t.Helper()
	var available, eligible, selectable bool
	var reason string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT available,eligible,selectable,usability_reason
		FROM portfolio_inventory_accounts
		WHERE user_id=$1 AND generation=1 AND account_id='account'`, owner).Scan(&available, &eligible, &selectable, &reason); err != nil {
		t.Fatal(err)
	}
	if available || eligible || selectable || reason != "connection_disabled" {
		t.Fatalf("account available=%v eligible=%v selectable=%v reason=%s", available, eligible, selectable, reason)
	}
}

func TestWebhookRepositoryPreservesActiveInventoryClaim(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-busy"
	owner := webhookOwnerWithInventory(t, fixture, 50, subject)
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE portfolio_inventory_state
		SET current_generation=current_generation+1,current_status='pending',claim_expires_at=$2
		WHERE user_id=$1`, owner, fixture.now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	service, err := portfolio.NewWebhookService(repository, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Apply(fixture.ctx, portfolio.WebhookEvent{Subject: subject, Type: portfolio.WebhookConnectionBroken, ConnectionID: "connection"})
	if !errors.Is(err, portfolio.ErrInventoryBusy) {
		t.Fatalf("error=%v", err)
	}
	var status string
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT status FROM portfolio_inventory_connections WHERE user_id=$1 AND generation=1 AND connection_id='connection'`, owner).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "active" {
		t.Fatalf("connection status=%s", status)
	}
}

func TestWebhookRepositoryPreservesActiveInitialInventoryClaim(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-initial-busy"
	owner := inclusionOwnerWithInventory(t, fixture, 55, subject)
	inventory := postgresadapter.NewInventoryRepository(fixture.pool)
	claim, err := inventory.Prepare(fixture.ctx, owner, false, fixture.now)
	if err != nil || !claim.Claimed {
		t.Fatalf("claim=%+v err=%v", claim, err)
	}
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	service, err := portfolio.NewWebhookService(repository, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Apply(fixture.ctx, portfolio.WebhookEvent{Subject: subject, Type: portfolio.WebhookConnectionAdded, ConnectionID: "connection"})
	if !errors.Is(err, portfolio.ErrInventoryBusy) {
		t.Fatalf("error=%v", err)
	}
	var connections int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT count(*) FROM portfolio_inventory_connections WHERE user_id=$1`, owner).Scan(&connections); err != nil {
		t.Fatal(err)
	}
	if connections != 0 {
		t.Fatalf("connections=%d", connections)
	}
}

func TestWebhookRepositoryRollsBackFailedProvisionalMutation(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	const subject = "webhook-rollback"
	owner := webhookOwnerWithInventory(t, fixture, 60, subject)
	repository := postgresadapter.NewInventoryLifecycleRepository(fixture.pool)
	if _, err := repository.AddAccount(fixture.ctx, subject, strings.Repeat("x", 129), "new-account", fixture.now.Add(time.Minute)); err == nil {
		t.Fatal("oversized connection unexpectedly succeeded")
	}
	var connections, accounts, identities int
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_inventory_connections WHERE user_id=$1),
		(SELECT count(*) FROM portfolio_inventory_accounts WHERE user_id=$1),
		(SELECT count(*) FROM portfolio_account_identities WHERE user_id=$1)`, owner).Scan(&connections, &accounts, &identities); err != nil {
		t.Fatal(err)
	}
	if connections != 1 || accounts != 1 || identities != 1 {
		t.Fatalf("partial mutation connections=%d accounts=%d identities=%d", connections, accounts, identities)
	}
}

func webhookOwnerWithInventory(t *testing.T, fixture *repositoryFixture, marker byte, subject string) uuid.UUID {
	t.Helper()
	inventory := postgresadapter.NewInventoryRepository(fixture.pool)
	return scheduledInventoryOwner(t, fixture, inventory, marker, subject)
}

func seedWebhookInclusionAndHistory(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) {
	t.Helper()
	versionID := uuid.New()
	claimID := uuid.New()
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_inclusion_state (user_id,version,lifecycle_generation,updated_at) VALUES ($1,1,1,$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_included_accounts (user_id,account_id,inclusion_version,included_at) VALUES ($1,'account',1,$2)`, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_account_sync_state
		(user_id,account_id,claim_id,claim_expires_at,claimed_inclusion_version,claimed_lifecycle_generation,claimed_inventory_generation,claimed_resource,updated_at)
		VALUES ($1,'account',$2,$3,1,1,1,'balances',$4)`, owner, claimID, fixture.now.Add(time.Minute), fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_versions
		(id,user_id,account_id,inclusion_version,retrieved_at,published_at)
		VALUES ($1,$2,'account',1,$3,$3)`, versionID, owner, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `INSERT INTO portfolio_balance_heads (user_id,account_id,version_id) VALUES ($1,'account',$2)`, owner, versionID); err != nil {
		t.Fatal(err)
	}
}

func assertWebhookRemovalState(t *testing.T, fixture *repositoryFixture, owner uuid.UUID) {
	t.Helper()
	var membership, identities, syncRows, balanceVersions, balanceHeads int
	var version, lifecycle int64
	var claimID *uuid.UUID
	if err := fixture.pool.QueryRow(fixture.ctx, `SELECT
		(SELECT count(*) FROM portfolio_included_accounts WHERE user_id=$1),
		(SELECT count(*) FROM portfolio_account_identities WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_account_sync_state WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_balance_versions WHERE user_id=$1 AND account_id='account'),
		(SELECT count(*) FROM portfolio_balance_heads WHERE user_id=$1 AND account_id='account'),
		state.version,state.lifecycle_generation,sync.claim_id
		FROM portfolio_inclusion_state state
		JOIN portfolio_account_sync_state sync ON sync.user_id=state.user_id AND sync.account_id='account'
		WHERE state.user_id=$1`, owner).Scan(&membership, &identities, &syncRows, &balanceVersions, &balanceHeads, &version, &lifecycle, &claimID); err != nil {
		t.Fatal(err)
	}
	if membership != 0 || identities != 1 || syncRows != 1 || balanceVersions != 1 || balanceHeads != 1 || version != 2 || lifecycle != 2 || claimID != nil {
		t.Fatalf("membership=%d identities=%d sync=%d balance_versions=%d balance_heads=%d version=%d lifecycle=%d claim=%v", membership, identities, syncRows, balanceVersions, balanceHeads, version, lifecycle, claimID)
	}
}
