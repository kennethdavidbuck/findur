package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

// InventoryRepository owns short claims and atomic immutable publication.
type InventoryRepository struct{ pool *pgxpool.Pool }

const maxInventoryRetryExponent = 5

// Read/admission policy must also constrain legacy snapshots whose selectable
// flag predates brokerage-only filtering. Persisted reasons distinguish open
// and unspecified status from accounts known to be closed or unavailable.
const selectableInventoryAccountSQL = `account.selectable AND account.available
	AND account.sync_state='complete'
	AND ((account.category='investment' AND account.usability_reason IN ('ready','provisional_status'))
		OR (account.category='unknown' AND account.usability_reason IN ('provisional_category','provisional_status')))
	AND connection.status='active' AND connection.available`

// NewInventoryRepository constructs the PostgreSQL inventory repository.
func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

// ClaimDue atomically claims one active user's due inventory generation.
func (r *InventoryRepository) ClaimDue(ctx context.Context, now time.Time, refreshAge, lease time.Duration) (*portfolio.ScheduledInventoryClaim, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	owner, authorizationGeneration, err := selectDueInventoryOwner(ctx, tx, now, refreshAge)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	preparation, found, err := loadPreparation(ctx, tx, owner)
	if err != nil {
		return nil, err
	}
	due, err := inventoryStillDue(ctx, tx, owner, now, refreshAge)
	if err != nil {
		return nil, err
	}
	if !due {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if !found {
		preparation, err = claimInitialInventoryWithLease(ctx, tx, owner, now, lease)
	} else {
		preparation, err = reclaimInventoryWithLease(ctx, tx, owner, preparation, now, lease)
	}
	if err != nil || !preparation.Claimed {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &portfolio.ScheduledInventoryClaim{
		Owner:                   owner,
		Generation:              preparation.Generation,
		AuthorizationGeneration: authorizationGeneration,
	}, nil
}

func inventoryStillDue(ctx context.Context, tx pgx.Tx, owner uuid.UUID, now time.Time, refreshAge time.Duration) (bool, error) {
	var due bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1
		FROM users app_user
		JOIN provider_authorizations provider_auth
			ON provider_auth.user_id=app_user.id
			AND provider_auth.provider=$2
			AND provider_auth.lifecycle_status='active'
		LEFT JOIN portfolio_inventory_state inventory
			ON inventory.user_id=app_user.id
		LEFT JOIN portfolio_inventory_versions head
			ON head.user_id=inventory.user_id
			AND head.generation=inventory.head_generation
			AND head.status IN ('ready','empty','disabled')
		WHERE app_user.id=$1
			AND app_user.active
			AND NOT EXISTS (
				SELECT 1
				FROM portfolio_inclusion_changes change
				WHERE change.user_id=app_user.id AND change.status='pending'
			)
			AND NOT EXISTS (
				SELECT 1
				FROM portfolio_account_sync_state sync
				WHERE sync.user_id=app_user.id AND sync.claim_expires_at>$3
			)
			AND (
				inventory.user_id IS NULL
				OR (inventory.diagnostic_refresh_needed
					AND (inventory.retry_at IS NULL OR inventory.retry_at<=$3))
				OR (inventory.current_status='pending' AND inventory.claim_expires_at<=$3)
				OR ((inventory.head_generation IS NULL OR head.generation IS NULL)
					AND inventory.current_status<>'pending'
					AND (inventory.retry_at IS NULL OR inventory.retry_at<=$3))
				OR (inventory.current_status NOT IN ('pending','ready','empty','disabled')
					AND (inventory.retry_at IS NULL OR inventory.retry_at<=$3))
				OR (inventory.current_status IN ('ready','empty','disabled')
					AND inventory.retry_at IS NOT NULL
					AND inventory.retry_at<=$3)
				OR (head.published_at<=$4
					AND inventory.retry_at IS NULL
					AND inventory.current_status IN ('ready','empty','disabled'))
			)
	)`, owner, auth.SnapTradeProvider, now, now.Add(-refreshAge)).Scan(&due)
	return due, err
}

func selectDueInventoryOwner(ctx context.Context, tx pgx.Tx, now time.Time, refreshAge time.Duration) (uuid.UUID, int64, error) {
	var owner uuid.UUID
	var authorizationGeneration int64
	err := tx.QueryRow(ctx, `SELECT
			app_user.id,
			provider_auth.lifecycle_generation
        FROM users app_user
        JOIN provider_authorizations provider_auth
            ON provider_auth.user_id = app_user.id
            AND provider_auth.provider = $1
            AND provider_auth.lifecycle_status = 'active'
        LEFT JOIN portfolio_inventory_state inventory
            ON inventory.user_id = app_user.id
        LEFT JOIN portfolio_inventory_versions head
            ON head.user_id = inventory.user_id
            AND head.generation = inventory.head_generation
            AND head.status IN ('ready','empty','disabled')
        WHERE app_user.active
            AND NOT EXISTS (
                SELECT 1
                FROM portfolio_inclusion_changes change
                WHERE change.user_id = app_user.id
                    AND change.status = 'pending'
            )
            AND NOT EXISTS (
                SELECT 1
                FROM portfolio_account_sync_state sync
                WHERE sync.user_id = app_user.id
                    AND sync.claim_expires_at > $2
            )
			AND (
				inventory.user_id IS NULL
				OR (inventory.diagnostic_refresh_needed
					AND (inventory.retry_at IS NULL OR inventory.retry_at <= $2))
				OR (inventory.current_status = 'pending' AND inventory.claim_expires_at <= $2)
				OR ((inventory.head_generation IS NULL OR head.generation IS NULL)
					AND inventory.current_status <> 'pending'
					AND (inventory.retry_at IS NULL OR inventory.retry_at <= $2))
                OR (inventory.current_status NOT IN ('pending','ready','empty','disabled')
                    AND (inventory.retry_at IS NULL OR inventory.retry_at <= $2))
				OR (inventory.current_status IN ('ready','empty','disabled')
					AND inventory.retry_at IS NOT NULL
					AND inventory.retry_at <= $2)
				OR (head.published_at <= $3
					AND inventory.retry_at IS NULL
					AND inventory.current_status IN ('ready','empty','disabled'))
            )
        ORDER BY head.published_at NULLS FIRST, inventory.updated_at NULLS FIRST, app_user.id
        FOR UPDATE OF app_user SKIP LOCKED
        LIMIT 1`, auth.SnapTradeProvider, now, now.Add(-refreshAge)).Scan(&owner, &authorizationGeneration)
	return owner, authorizationGeneration, err
}

// Prepare returns persisted inventory or atomically claims a bounded bootstrap/retry.
func (r *InventoryRepository) Prepare(ctx context.Context, owner uuid.UUID, retry bool, now time.Time) (portfolio.Preparation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockActiveOwner(ctx, tx, owner); err != nil {
		return portfolio.Preparation{}, err
	}
	preparation, err := prepareInventoryClaim(ctx, tx, owner, retry, now)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	if preparation.Claimed {
		if err := loadInventoryAuthorization(ctx, tx, owner); err != nil {
			return portfolio.Preparation{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.Preparation{}, err
	}
	return preparation, nil
}

func prepareInventoryClaim(ctx context.Context, tx pgx.Tx, owner uuid.UUID, retry bool, now time.Time) (portfolio.Preparation, error) {
	preparation, found, err := loadPreparation(ctx, tx, owner)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	if !found {
		return claimInitialInventory(ctx, tx, owner, now)
	}
	if claimAllowed(preparation, retry, now) {
		return reclaimInventory(ctx, tx, owner, preparation, now)
	}
	return preparation, nil
}

func claimInitialInventory(ctx context.Context, tx pgx.Tx, owner uuid.UUID, now time.Time) (portfolio.Preparation, error) {
	return claimInitialInventoryWithLease(ctx, tx, owner, now, time.Minute)
}

func claimInitialInventoryWithLease(ctx context.Context, tx pgx.Tx, owner uuid.UUID, now time.Time, lease time.Duration) (portfolio.Preparation, error) {
	claimExpiresAt := now.Add(lease)
	command, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_state (user_id,current_generation,current_status,claim_expires_at,updated_at)
		VALUES ($1,1,'pending',$2,$3) ON CONFLICT (user_id) DO NOTHING`, owner, claimExpiresAt, now)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	if command.RowsAffected() == 1 {
		return portfolio.Preparation{Snapshot: portfolio.Snapshot{State: portfolio.StatePending, Generation: 1, UpdatedAt: now}, Claimed: true, ClaimExpiresAt: &claimExpiresAt}, nil
	}
	preparation, found, err := loadPreparation(ctx, tx, owner)
	if err != nil || !found {
		return portfolio.Preparation{}, err
	}
	return preparation, nil
}

func claimAllowed(preparation portfolio.Preparation, retry bool, now time.Time) bool {
	expiredPending := preparation.State == portfolio.StatePending && preparation.ClaimExpiresAt != nil && !now.Before(*preparation.ClaimExpiresAt)
	retryReady := retry && preparation.State != portfolio.StatePending && (preparation.RetryAt == nil || !now.Before(*preparation.RetryAt))
	return expiredPending || retryReady
}

func reclaimInventory(ctx context.Context, tx pgx.Tx, owner uuid.UUID, preparation portfolio.Preparation, now time.Time) (portfolio.Preparation, error) {
	return reclaimInventoryWithLease(ctx, tx, owner, preparation, now, time.Minute)
}

func reclaimInventoryWithLease(ctx context.Context, tx pgx.Tx, owner uuid.UUID, preparation portfolio.Preparation, now time.Time, lease time.Duration) (portfolio.Preparation, error) {
	preparation.Generation++
	claimExpiresAt := now.Add(lease)
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_generation=$2,current_status='pending',retry_at=NULL,claim_expires_at=$3,updated_at=$4 WHERE user_id=$1`, owner, preparation.Generation, claimExpiresAt, now)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	preparation.State, preparation.RetryAt, preparation.ClaimExpiresAt, preparation.UpdatedAt, preparation.Claimed = portfolio.StatePending, nil, &claimExpiresAt, now, true
	return preparation, nil
}

func loadInventoryAuthorization(ctx context.Context, tx pgx.Tx, owner uuid.UUID) error {
	err := tx.QueryRow(ctx, `SELECT 1
        FROM provider_authorizations
        WHERE user_id = $1
            AND provider = $2
            AND lifecycle_status = 'active'`, owner, auth.SnapTradeProvider).Scan(new(int))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

// Finalize publishes a successful immutable version or stores a safe failure category.
func (r *InventoryRepository) Finalize(ctx context.Context, owner uuid.UUID, generation int64, state portfolio.State, retryAt *time.Time, connections []portfolio.Connection, now time.Time) (portfolio.Snapshot, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	preparation, found, err := loadPreparation(ctx, tx, owner)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	if !found || preparation.Generation != generation || preparation.State != portfolio.StatePending {
		if err := tx.Commit(ctx); err != nil {
			return portfolio.Snapshot{}, false, err
		}
		return preparation.Snapshot, false, nil
	}
	if shouldPublishInventory(state, connections) {
		err = publishInventoryVersion(ctx, tx, owner, generation, state, retryAt, connections, now)
	} else {
		err = recordInventoryFailure(ctx, tx, owner, generation, state, retryAt, now)
	}
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	result, _, err := loadPreparation(ctx, tx, owner)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.Snapshot{}, false, err
	}
	return result.Snapshot, true, nil
}

// FinalizeScheduled publishes a guarded bundle or schedules a bounded retry.
func (r *InventoryRepository) FinalizeScheduled(ctx context.Context, claim portfolio.ScheduledInventoryClaim, state portfolio.State, providerRetryAt *time.Time, connections []portfolio.Connection, now time.Time) (portfolio.Snapshot, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	active, authorizationStatus, authorizationGeneration, err := lockScheduledInventoryOwner(ctx, tx, claim.Owner)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	preparation, found, err := loadPreparation(ctx, tx, claim.Owner)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	authorizationMatches := authorizationStatus == auth.CredentialStatusActive && authorizationGeneration == claim.AuthorizationGeneration
	authorizationAllowsFailure := state == portfolio.StateUnauthorized && authorizationStatus != auth.CredentialStatusActive
	guarded := active && found && preparation.Generation == claim.Generation && preparation.State == portfolio.StatePending &&
		preparation.ClaimExpiresAt != nil && preparation.ClaimExpiresAt.After(now) &&
		(authorizationMatches || authorizationAllowsFailure)
	if guarded {
		var blocked bool
		err = tx.QueryRow(ctx, `SELECT
            EXISTS(SELECT 1 FROM portfolio_inclusion_changes WHERE user_id=$1 AND status='pending')
            OR EXISTS(SELECT 1 FROM portfolio_account_sync_state WHERE user_id=$1 AND claim_expires_at>$2)`, claim.Owner, now).Scan(&blocked)
		guarded = err == nil && !blocked
	}
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	switch {
	case !guarded:
		if found && preparation.Generation == claim.Generation && preparation.State == portfolio.StatePending {
			_, err = tx.Exec(ctx, `UPDATE portfolio_inventory_state
				SET current_status=COALESCE((SELECT status FROM portfolio_inventory_versions
					WHERE user_id=$1 AND generation=head_generation),'unavailable'),
					claim_expires_at=NULL,updated_at=$3
                WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, claim.Owner, claim.Generation, now)
		}
	case shouldPublishInventory(state, connections):
		err = publishInventoryVersion(ctx, tx, claim.Owner, claim.Generation, state, nil, connections, now)
	default:
		failureCount, countErr := preparationFailureCount(ctx, tx, claim.Owner)
		if countErr != nil {
			return portfolio.Snapshot{}, false, countErr
		}
		retryAt := now.Add(time.Minute << min(failureCount, maxInventoryRetryExponent))
		if providerRetryAt != nil && providerRetryAt.After(retryAt) {
			retryAt = *providerRetryAt
		}
		reason, action := inventoryFailureDiagnostic(state, &retryAt)
		_, err = tx.Exec(ctx, `UPDATE portfolio_inventory_state
			SET current_status=COALESCE((SELECT status FROM portfolio_inventory_versions
				WHERE user_id=$1 AND generation=head_generation AND status IN ('ready','empty','disabled')),$3),
				retry_at=$4,claim_expires_at=NULL,failure_count=failure_count+1,
				failure_reason=$6,failure_action=$7,updated_at=$5
			WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, claim.Owner, claim.Generation, state, retryAt, now, reason, action)
	}
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	result, _, err := loadPreparation(ctx, tx, claim.Owner)
	if err != nil {
		return portfolio.Snapshot{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.Snapshot{}, false, err
	}
	return result.Snapshot, guarded, nil
}

func preparationFailureCount(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (int, error) {
	var failures int
	err := tx.QueryRow(ctx, `SELECT failure_count FROM portfolio_inventory_state WHERE user_id=$1`, owner).Scan(&failures)
	return failures, err
}

func shouldPublishInventory(state portfolio.State, connections []portfolio.Connection) bool {
	return state == portfolio.StateReady || state == portfolio.StateEmpty || len(connections) > 0
}

func publishInventoryVersion(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, state portfolio.State, retryAt *time.Time, connections []portfolio.Connection, now time.Time) error {
	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_versions (user_id,generation,status,published_at) VALUES ($1,$2,$3,$4)`, owner, generation, state, now); err != nil {
		return err
	}
	if err := insertInventoryRows(ctx, tx, owner, generation, connections, now); err != nil {
		return err
	}
	if err := removeIneligibleIncludedAccounts(ctx, tx, owner, generation, now); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_status=$3,head_generation=$2,retry_at=$4,
		claim_expires_at=NULL,failure_count=0,failure_reason=NULL,failure_action=NULL,
		diagnostic_refresh_needed=false,updated_at=$5
		WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, owner, generation, state, retryAt, now)
	return err
}

func removeIneligibleIncludedAccounts(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, now time.Time) error {
	rows, err := tx.Query(ctx, `SELECT included.account_id
		FROM portfolio_included_accounts included
		WHERE included.user_id=$1
		AND NOT EXISTS (
			SELECT 1
			FROM portfolio_inventory_accounts account
			JOIN portfolio_inventory_connections connection USING (user_id,generation,connection_id)
			WHERE account.user_id=included.user_id
			AND account.generation=$2
			AND account.account_id=included.account_id
			AND `+selectableInventoryAccountSQL+`
		)
		ORDER BY included.account_id
		FOR UPDATE OF included`, owner, generation)
	if err != nil {
		return err
	}
	var removals []string
	for rows.Next() {
		var accountID string
		if err := rows.Scan(&accountID); err != nil {
			rows.Close()
			return err
		}
		removals = append(removals, accountID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(removals) == 0 {
		return nil
	}

	var version, lifecycleGeneration int64
	if err := tx.QueryRow(ctx, `SELECT version,lifecycle_generation
		FROM portfolio_inclusion_state
		WHERE user_id=$1
		FOR UPDATE`, owner).Scan(&version, &lifecycleGeneration); err != nil {
		return err
	}
	resultVersion := version + 1
	lifecycleGeneration++
	if _, err := tx.Exec(ctx, `DELETE FROM portfolio_included_accounts
		WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE portfolio_account_sync_state
		SET claim_id=NULL,claim_expires_at=NULL,claimed_resource=NULL,claimed_change_id=NULL,updated_at=$3
		WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE portfolio_inclusion_state
		SET version=$2,lifecycle_generation=$3,updated_at=$4
		WHERE user_id=$1`, owner, resultVersion, lifecycleGeneration, now); err != nil {
		return err
	}
	remaining, err := loadCommittedAccountIDs(ctx, tx, owner)
	if err != nil {
		return err
	}
	changeID := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO portfolio_inclusion_changes
		(id,user_id,idempotency_key,expected_version,result_version,inventory_generation,lifecycle_generation,target_account_ids,addition_account_ids,removal_account_ids,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'{}',$9,'committed',$10,$10)`,
		changeID, owner, "system-inventory-"+changeID.String(), version, resultVersion, generation, lifecycleGeneration, remaining, removals, now)
	return err
}

func lockScheduledInventoryOwner(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (bool, string, int64, error) {
	var active bool
	if err := tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1 FOR UPDATE`, owner).Scan(&active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", 0, nil
		}
		return false, "", 0, err
	}
	var status string
	var generation int64
	err := tx.QueryRow(ctx, `SELECT lifecycle_status,lifecycle_generation
		FROM provider_authorizations
		WHERE user_id=$1 AND provider=$2
		FOR UPDATE`, owner, auth.SnapTradeProvider).Scan(&status, &generation)
	if errors.Is(err, pgx.ErrNoRows) {
		return active, "", 0, nil
	}
	return active, status, generation, err
}

func insertInventoryRows(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connections []portfolio.Connection, now time.Time) error {
	for _, connection := range connections {
		var diagnosticReason, diagnosticAction any
		if connection.Diagnostic != nil {
			diagnosticReason = connection.Diagnostic.Reason
			diagnosticAction = connection.Diagnostic.RecommendedAction
		}
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_connections
			(user_id,generation,connection_id,brokerage_label,status,sync_mode,available,eligible,diagnostic_reason,diagnostic_action)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, owner, generation, connection.ID, connection.BrokerageLabel, connection.Status, connection.SyncMode, connection.Available, connection.Eligible, diagnosticReason, diagnosticAction); err != nil {
			return err
		}
		if err := insertAccountRows(ctx, tx, owner, generation, connection, now); err != nil {
			return err
		}
	}
	return nil
}

func insertAccountRows(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connection portfolio.Connection, now time.Time) error {
	for _, account := range connection.Accounts {
		selectable, reason := account.Selectable || account.Eligible, account.UsabilityReason
		if reason == "" {
			reason = portfolio.UsabilityAccountUnavailable
			if account.Eligible {
				reason = portfolio.UsabilityReady
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_accounts
			(user_id,generation,account_id,connection_id,category,account_type,masked_label,available,eligible,sync_state,selectable,usability_reason,total_balance_amount,total_balance_currency)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, owner, generation, account.ID, connection.ID, account.Category, account.Type, account.MaskedLabel, account.Available, account.Eligible, account.SyncState, selectable, reason, account.TotalBalanceAmount, account.TotalBalanceCurrency); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_account_identities (user_id,account_id,connection_id,first_seen_at,last_seen_at)
			VALUES ($1,$2,$3,$4,$4) ON CONFLICT (user_id,account_id) DO UPDATE SET connection_id=EXCLUDED.connection_id,last_seen_at=EXCLUDED.last_seen_at`, owner, account.ID, connection.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func recordInventoryFailure(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, state portfolio.State, retryAt *time.Time, now time.Time) error {
	reason, action := inventoryFailureDiagnostic(state, retryAt)
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_status=$3,retry_at=$4,
		claim_expires_at=NULL,failure_reason=$6,failure_action=$7,updated_at=$5
		WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, owner, generation, state, retryAt, now, reason, action)
	return err
}

func inventoryFailureDiagnostic(state portfolio.State, retryAt *time.Time) (portfolio.ResourceDiagnosticReason, portfolio.ResourceDiagnosticAction) {
	switch state {
	case portfolio.StateUnauthorized:
		return portfolio.DiagnosticAuthorizationRequired, portfolio.DiagnosticActionReconnect
	case portfolio.StateRateLimited, portfolio.StateUnavailable:
		if retryAt != nil {
			return portfolio.DiagnosticProviderUnavailable, portfolio.DiagnosticActionWait
		}
		return portfolio.DiagnosticProviderUnavailable, portfolio.DiagnosticActionRetry
	case portfolio.StatePending:
		return portfolio.DiagnosticSyncPending, portfolio.DiagnosticActionWait
	default:
		return portfolio.DiagnosticUnknown, portfolio.DiagnosticActionRetry
	}
}

func loadPreparation(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (portfolio.Preparation, bool, error) {
	result, head, found, err := loadInventoryState(ctx, tx, owner)
	if err != nil || !found {
		return result, found, err
	}
	if head == nil {
		return result, true, nil
	}
	var trustworthy bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1
		FROM portfolio_inventory_versions
		WHERE user_id=$1
		AND generation=$2
		AND status IN ('ready','empty','disabled'))`, owner, *head).Scan(&trustworthy); err != nil {
		return portfolio.Preparation{}, false, err
	}
	if !trustworthy {
		return result, true, nil
	}
	result.Connections, err = loadInventoryConnections(ctx, tx, owner, *head)
	if err != nil {
		return portfolio.Preparation{}, false, err
	}
	accountCount := 0
	for index := range result.Connections {
		result.Connections[index].Accounts, err = loadInventoryAccounts(ctx, tx, owner, *head, result.Connections[index].ID)
		if err != nil {
			return portfolio.Preparation{}, false, err
		}
		accountCount += len(result.Connections[index].Accounts)
	}
	applyInventoryFailureDiagnostics(&result)
	if result.State == portfolio.StateReady && accountCount == 0 {
		result.State = portfolio.StateEmpty
	}
	return result, true, nil
}

func loadInventoryState(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (portfolio.Preparation, *int64, bool, error) {
	var result portfolio.Preparation
	var head *int64
	var reason *portfolio.ResourceDiagnosticReason
	var action *portfolio.ResourceDiagnosticAction
	err := tx.QueryRow(ctx, `SELECT current_generation,current_status,head_generation,retry_at,claim_expires_at,updated_at,failure_reason,failure_action
		FROM portfolio_inventory_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&result.Generation, &result.State, &head, &result.RetryAt, &result.ClaimExpiresAt, &result.UpdatedAt, &reason, &action)
	if errors.Is(err, pgx.ErrNoRows) {
		return portfolio.Preparation{}, nil, false, nil
	}
	if err == nil && reason != nil && action != nil {
		result.Diagnostic = &portfolio.ResourceDiagnostic{Reason: *reason, RecommendedAction: *action, RetryAt: result.RetryAt}
	}
	return result, head, err == nil, err
}

func loadInventoryConnections(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64) ([]portfolio.Connection, error) {
	rows, err := tx.Query(ctx, `SELECT
			connection.connection_id,
			connection.brokerage_label,
			connection.status,
			connection.sync_mode,
			connection.available,
			connection.eligible,
			connection.diagnostic_reason,
			connection.diagnostic_action,
			version.published_at
		FROM portfolio_inventory_connections connection
		JOIN portfolio_inventory_versions version
			ON version.user_id = connection.user_id
			AND version.generation = connection.generation
		WHERE connection.user_id = $1
			AND connection.generation = $2
		ORDER BY connection.connection_id`, owner, generation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var connections []portfolio.Connection
	for rows.Next() {
		var connection portfolio.Connection
		var reason *portfolio.ResourceDiagnosticReason
		var action *portfolio.ResourceDiagnosticAction
		var publishedAt time.Time
		if err := rows.Scan(&connection.ID, &connection.BrokerageLabel, &connection.Status, &connection.SyncMode, &connection.Available, &connection.Eligible, &reason, &action, &publishedAt); err != nil {
			return nil, err
		}
		if reason != nil && action != nil {
			connection.Diagnostic = &portfolio.ResourceDiagnostic{Reason: *reason, RecommendedAction: *action, LastSuccessfulAt: &publishedAt}
		}
		connection.LastSuccessfulAt = &publishedAt
		connections = append(connections, connection)
	}
	return connections, rows.Err()
}

func applyInventoryFailureDiagnostics(result *portfolio.Preparation) {
	if result.Diagnostic != nil {
		for index := range result.Connections {
			diagnostic := *result.Diagnostic
			diagnostic.LastSuccessfulAt = result.Connections[index].LastSuccessfulAt
			result.Connections[index].Diagnostic = &diagnostic
		}
		return
	}
	var reason portfolio.ResourceDiagnosticReason
	var action portfolio.ResourceDiagnosticAction
	switch result.State {
	case portfolio.StatePending:
		reason, action = portfolio.DiagnosticSyncPending, portfolio.DiagnosticActionWait
	case portfolio.StateUnauthorized:
		reason, action = portfolio.DiagnosticAuthorizationRequired, portfolio.DiagnosticActionReconnect
	case portfolio.StateRateLimited, portfolio.StateUnavailable:
		reason, action = portfolio.DiagnosticProviderUnavailable, portfolio.DiagnosticActionRetry
		if result.RetryAt != nil {
			action = portfolio.DiagnosticActionWait
		}
	case portfolio.StateMalformed:
		reason, action = portfolio.DiagnosticUnknown, portfolio.DiagnosticActionRetry
	default:
		return
	}
	for index := range result.Connections {
		result.Connections[index].Diagnostic = &portfolio.ResourceDiagnostic{
			Reason:            reason,
			RecommendedAction: action,
			RetryAt:           result.RetryAt,
			LastSuccessfulAt:  result.Connections[index].LastSuccessfulAt,
		}
	}
}

func loadInventoryAccounts(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connectionID string) ([]portfolio.Account, error) {
	rows, err := tx.Query(ctx, `SELECT
			account.account_id,
			account.category,
			account.account_type,
			account.masked_label,
			account.available,
			account.eligible,
			account.sync_state,
			account.selectable,
			account.usability_reason,
			account.total_balance_amount,
			account.total_balance_currency
		FROM portfolio_inventory_accounts account
		JOIN portfolio_inventory_connections connection USING (user_id,generation,connection_id)
		WHERE account.user_id=$1 AND account.generation=$2 AND account.connection_id=$3
		AND (
			`+selectableInventoryAccountSQL+`
			OR (
				NOT account.selectable
				AND (
					account.usability_reason IN ('unsupported_category','connection_disabled')
					OR (
						account.usability_reason = 'provisional_category'
						AND account.category = 'unknown'
						AND account.sync_state = 'unknown'
						AND NOT account.available
					)
				)
			)
		)
		ORDER BY account.account_id`, owner, generation, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []portfolio.Account
	for rows.Next() {
		var account portfolio.Account
		if err := rows.Scan(&account.ID, &account.Category, &account.Type, &account.MaskedLabel, &account.Available, &account.Eligible, &account.SyncState, &account.Selectable, &account.UsabilityReason, &account.TotalBalanceAmount, &account.TotalBalanceCurrency); err != nil {
			return nil, err
		}
		if account.Selectable {
			// Matching today's predicate is sufficient, including legacy null
			// provider status/category values. Project current semantics without
			// rewriting immutable inventory history.
			if account.Category == portfolio.AccountCategoryInvestment {
				account.Eligible, account.UsabilityReason = true, portfolio.UsabilityReady
			} else {
				account.Eligible = false
				if account.UsabilityReason != portfolio.UsabilityProvisionalStatus {
					account.UsabilityReason = portfolio.UsabilityProvisionalCategory
				}
			}
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}
