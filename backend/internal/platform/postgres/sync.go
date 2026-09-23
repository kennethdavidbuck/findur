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

// SyncRepository owns durable account refresh claims and guarded publication.
type SyncRepository struct{ pool *pgxpool.Pool }

const (
	maxSyncRetryExponent           = 5
	syncFailureAuthorization       = "authorization_required"
	syncFailureRateLimited         = "rate_limited"
	syncFailureProviderUnavailable = "provider_unavailable"
)

// NewSyncRepository constructs a PostgreSQL synchronization repository.
func NewSyncRepository(pool *pgxpool.Pool) *SyncRepository { return &SyncRepository{pool: pool} }

// AcquireWorkerLease ensures that only one process drains sync work at a time.
func (r *SyncRepository) AcquireWorkerLease(ctx context.Context, now time.Time, lease time.Duration) (uuid.UUID, bool, error) {
	claimID := uuid.New()
	var accepted uuid.UUID
	err := r.pool.QueryRow(ctx, `INSERT INTO portfolio_sync_worker_lease (singleton,claim_id,claim_expires_at,updated_at)
		VALUES (true,$1,$2,$3) ON CONFLICT (singleton) DO UPDATE SET
		claim_id=EXCLUDED.claim_id,claim_expires_at=EXCLUDED.claim_expires_at,updated_at=EXCLUDED.updated_at
		WHERE portfolio_sync_worker_lease.claim_expires_at IS NULL OR portfolio_sync_worker_lease.claim_expires_at<=$3
		RETURNING claim_id`, claimID, now.Add(lease), now).Scan(&accepted)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return accepted, err == nil, err
}

// ReleaseWorkerLease releases only the caller's current singleton claim.
func (r *SyncRepository) ReleaseWorkerLease(ctx context.Context, claimID uuid.UUID, now time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE portfolio_sync_worker_lease SET claim_id=NULL,claim_expires_at=NULL,updated_at=$2 WHERE singleton AND claim_id=$1`, claimID, now)
	return err
}

// ClaimDue prioritizes pending first-inclusion resources, then scheduled ones.
func (r *SyncRepository) ClaimDue(ctx context.Context, now time.Time, refreshAge, lease time.Duration) (*portfolio.SyncClaim, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	claim, err := selectDueSyncClaim(ctx, tx, now, refreshAge)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	claim.ID = uuid.New()
	if err := persistSyncClaim(ctx, tx, claim, now, lease); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &claim, nil
}

func selectDueSyncClaim(ctx context.Context, tx pgx.Tx, now time.Time, refreshAge time.Duration) (portfolio.SyncClaim, error) {
	claim, err := selectPendingInclusionClaim(ctx, tx, now)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		return claim, err
	}
	return selectScheduledSyncClaim(ctx, tx, now, refreshAge)
}

func selectPendingInclusionClaim(ctx context.Context, tx pgx.Tx, now time.Time) (portfolio.SyncClaim, error) {
	var claim portfolio.SyncClaim
	err := tx.QueryRow(ctx, `SELECT
            sync.user_id,
            sync.account_id,
            change.result_version,
            change.lifecycle_generation,
            inventory.current_generation,
            CASE
                WHEN balance.version_id IS NULL OR sync.balances_success_at IS NULL THEN 'balances'
                WHEN position.version_id IS NULL OR sync.positions_success_at IS NULL THEN 'positions'
                ELSE 'activities'
            END,
            change.id
        FROM portfolio_inclusion_changes change
        JOIN portfolio_inclusion_state inclusion
            ON inclusion.user_id = change.user_id
            AND inclusion.version = change.result_version
            AND inclusion.lifecycle_generation = change.lifecycle_generation
        JOIN portfolio_account_sync_state sync
            ON sync.user_id = change.user_id
            AND sync.account_id = ANY(change.sync_account_ids)
        JOIN portfolio_inventory_state inventory
            ON inventory.user_id = sync.user_id
            AND inventory.current_generation = change.inventory_generation
        JOIN portfolio_inventory_accounts account
            ON account.user_id = sync.user_id
            AND account.account_id = sync.account_id
            AND account.generation = inventory.head_generation
        JOIN portfolio_inventory_connections connection
            ON connection.user_id = account.user_id
            AND connection.generation = account.generation
            AND connection.connection_id = account.connection_id
        JOIN users app_user
            ON app_user.id = sync.user_id
            AND app_user.active
        JOIN provider_authorizations provider_auth
            ON provider_auth.user_id = sync.user_id
            AND provider_auth.provider = $1
            AND provider_auth.lifecycle_status = 'active'
        LEFT JOIN portfolio_balance_heads balance
            ON balance.user_id = sync.user_id
            AND balance.account_id = sync.account_id
        LEFT JOIN portfolio_position_heads position
            ON position.user_id = sync.user_id
            AND position.account_id = sync.account_id
        WHERE change.status = 'pending'
            AND sync.initialized_at IS NULL
            AND `+selectableInventoryAccountSQL+`
            AND (sync.next_attempt_at IS NULL OR sync.next_attempt_at <= $2)
            AND (sync.claim_expires_at IS NULL OR sync.claim_expires_at <= $2)
        ORDER BY change.created_at, sync.updated_at
        FOR UPDATE OF sync SKIP LOCKED
        LIMIT 1`, auth.SnapTradeProvider, now).Scan(
		&claim.Owner, &claim.AccountID, &claim.InclusionVersion, &claim.LifecycleGeneration, &claim.InventoryGeneration, &claim.Resource, &claim.ChangeID)
	return claim, err
}

func selectScheduledSyncClaim(ctx context.Context, tx pgx.Tx, now time.Time, refreshAge time.Duration) (portfolio.SyncClaim, error) {
	var claim portfolio.SyncClaim
	err := tx.QueryRow(ctx, `SELECT
            sync.user_id,
            sync.account_id,
            included.inclusion_version,
            inclusion.lifecycle_generation,
            inventory.current_generation,
            CASE
                WHEN balance.version_id IS NULL
                    OR sync.balances_success_at IS NULL
                    OR sync.balances_success_at <= sync.last_success_at THEN 'balances'
                WHEN position.version_id IS NULL
                    OR sync.positions_success_at IS NULL
                    OR sync.positions_success_at <= sync.last_success_at THEN 'positions'
                ELSE 'activities'
            END,
            NULL::uuid
        FROM portfolio_account_sync_state sync
        JOIN portfolio_included_accounts included
            ON included.user_id = sync.user_id
            AND included.account_id = sync.account_id
        JOIN portfolio_inclusion_state inclusion
            ON inclusion.user_id = sync.user_id
        JOIN portfolio_inventory_state inventory
            ON inventory.user_id = sync.user_id
        JOIN portfolio_inventory_accounts account
            ON account.user_id = sync.user_id
            AND account.account_id = sync.account_id
            AND account.generation = inventory.head_generation
        JOIN portfolio_inventory_connections connection
            ON connection.user_id = account.user_id
            AND connection.generation = account.generation
            AND connection.connection_id = account.connection_id
        JOIN users app_user
            ON app_user.id = sync.user_id
            AND app_user.active
        JOIN provider_authorizations provider_auth
            ON provider_auth.user_id = sync.user_id
            AND provider_auth.provider = $1
            AND provider_auth.lifecycle_status = 'active'
        LEFT JOIN portfolio_balance_heads balance
            ON balance.user_id = sync.user_id
            AND balance.account_id = sync.account_id
        LEFT JOIN portfolio_position_heads position
            ON position.user_id = sync.user_id
            AND position.account_id = sync.account_id
        LEFT JOIN portfolio_activity_heads activity
            ON activity.user_id = sync.user_id
            AND activity.account_id = sync.account_id
		WHERE `+selectableInventoryAccountSQL+`
			AND (sync.next_attempt_at IS NULL OR sync.next_attempt_at <= $2)
			AND (sync.claim_expires_at IS NULL OR sync.claim_expires_at <= $2)
			AND (sync.initialized_at IS NULL
				OR sync.last_success_at IS NULL
				OR sync.last_success_at <= $3
                OR balance.version_id IS NULL
                OR position.version_id IS NULL
                OR activity.version_id IS NULL)
        ORDER BY sync.last_success_at NULLS FIRST, sync.updated_at
        FOR UPDATE OF sync SKIP LOCKED
        LIMIT 1`, auth.SnapTradeProvider, now, now.Add(-refreshAge)).Scan(
		&claim.Owner, &claim.AccountID, &claim.InclusionVersion, &claim.LifecycleGeneration, &claim.InventoryGeneration, &claim.Resource, &claim.ChangeID)
	return claim, err
}

func persistSyncClaim(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, now time.Time, lease time.Duration) error {
	_, err := tx.Exec(ctx, `UPDATE portfolio_account_sync_state SET claim_id=$3,claim_expires_at=$4,
		claimed_inclusion_version=$5,claimed_lifecycle_generation=$6,claimed_inventory_generation=$7,
		claimed_resource=$9,claimed_change_id=$10,updated_at=$2
		WHERE user_id=$1 AND account_id=$8`, claim.Owner, now, claim.ID, now.Add(lease), claim.InclusionVersion, claim.LifecycleGeneration, claim.InventoryGeneration, claim.AccountID, claim.Resource, claim.ChangeID)
	return err
}

// FinishSync publishes one guarded resource or durably schedules its retry.
func (r *SyncRepository) FinishSync(ctx context.Context, claim portfolio.SyncClaim, data *portfolio.AccountData, failure string, providerRetryAt *time.Time, now time.Time) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockActiveOwner(ctx, tx, claim.Owner); err != nil {
		return false, err
	}
	guard, found, err := loadSyncGuard(ctx, tx, claim)
	if err != nil {
		return false, err
	}
	guarded := found && guard.matches(claim)
	if guarded && validateSelection(ctx, tx, claim.Owner, guard.inventoryGeneration, []string{claim.AccountID}, nil) != nil {
		guarded = false
	}
	if failure == "" && !completeSyncData(claim.Resource, data) {
		failure = "unusable_data"
	}

	switch {
	case !guarded:
		err = clearSyncClaim(ctx, tx, claim, now)
	case failure != "":
		err = scheduleSyncRetry(ctx, tx, claim, guard.failureCount, failure, providerRetryAt, now)
	default:
		err = publishScheduledSync(ctx, tx, claim, *data, now)
	}
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return guarded && failure == "" && data != nil, nil
}

type syncGuard struct {
	claimID                                                    *uuid.UUID
	inclusionVersion, lifecycleGeneration, inventoryGeneration int64
	failureCount                                               int
	resource                                                   portfolio.AccountResource
	changeID                                                   *uuid.UUID
}

func (g syncGuard) matches(claim portfolio.SyncClaim) bool {
	return g.claimID != nil && *g.claimID == claim.ID &&
		g.inclusionVersion == claim.InclusionVersion &&
		g.lifecycleGeneration == claim.LifecycleGeneration &&
		g.inventoryGeneration == claim.InventoryGeneration &&
		g.resource == claim.Resource && equalOptionalUUID(g.changeID, claim.ChangeID)
}

func equalOptionalUUID(left, right *uuid.UUID) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func loadSyncGuard(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim) (syncGuard, bool, error) {
	var guard syncGuard
	query := `SELECT sync.claim_id,included.inclusion_version,inclusion.lifecycle_generation,inventory.current_generation,
		sync.failure_count,sync.claimed_resource,sync.claimed_change_id
		FROM portfolio_account_sync_state sync
		JOIN portfolio_included_accounts included
		  ON included.user_id=sync.user_id AND included.account_id=sync.account_id
		JOIN portfolio_inclusion_state inclusion ON inclusion.user_id=sync.user_id
		JOIN portfolio_inventory_state inventory ON inventory.user_id=sync.user_id
		WHERE sync.user_id=$1 AND sync.account_id=$2 FOR UPDATE OF sync`
	if claim.ChangeID != nil {
		query = `SELECT sync.claim_id,change.result_version,change.lifecycle_generation,inventory.current_generation,
			sync.failure_count,sync.claimed_resource,sync.claimed_change_id
			FROM portfolio_account_sync_state sync
			JOIN portfolio_inclusion_changes change
			  ON change.id=$3 AND change.user_id=sync.user_id AND sync.account_id=ANY(change.sync_account_ids) AND change.status='pending'
			JOIN portfolio_inclusion_state inclusion
			  ON inclusion.user_id=change.user_id AND inclusion.version=change.result_version AND inclusion.lifecycle_generation=change.lifecycle_generation
			JOIN portfolio_inventory_state inventory
			  ON inventory.user_id=sync.user_id AND inventory.current_generation=change.inventory_generation
			WHERE sync.user_id=$1 AND sync.account_id=$2 FOR UPDATE OF sync,change`
	}
	args := []any{claim.Owner, claim.AccountID}
	if claim.ChangeID != nil {
		args = append(args, claim.ChangeID)
	}
	err := tx.QueryRow(ctx, query, args...).Scan(
		&guard.claimID, &guard.inclusionVersion, &guard.lifecycleGeneration, &guard.inventoryGeneration,
		&guard.failureCount, &guard.resource, &guard.changeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return syncGuard{}, false, nil
	}
	return guard, err == nil, err
}

func completeSyncData(resource portfolio.AccountResource, data *portfolio.AccountData) bool {
	if data == nil {
		return false
	}
	switch resource {
	case portfolio.AccountResourceBalances:
		return !data.Balances.RetrievedAt.IsZero() && data.Balances.Rows != nil
	case portfolio.AccountResourcePositions:
		return !data.Positions.ObservedAt.IsZero() && !data.Positions.RetrievedAt.IsZero() && data.Positions.Rows != nil
	case portfolio.AccountResourceActivities:
		return !data.Activities.RetrievedAt.IsZero() && data.Activities.Rows != nil
	default:
		return false
	}
}

func clearSyncClaim(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, now time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE portfolio_account_sync_state SET claim_id=NULL,claim_expires_at=NULL,claimed_resource=NULL,claimed_change_id=NULL,updated_at=$3
		WHERE user_id=$1 AND account_id=$2 AND claim_id=$4`, claim.Owner, claim.AccountID, now, claim.ID)
	return err
}

func scheduleSyncRetry(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, failures int, failure string, providerRetryAt *time.Time, now time.Time) error {
	retryAt := now.Add(time.Minute << min(failures, maxSyncRetryExponent))
	if providerRetryAt != nil && providerRetryAt.After(retryAt) {
		retryAt = *providerRetryAt
	}
	reasonColumn, actionColumn, retryColumn, err := diagnosticColumns(claim.Resource)
	if err != nil {
		return err
	}
	reason := syncFailureDiagnosticReason(failure)
	action := syncFailureDiagnosticAction(failure)
	_, err = tx.Exec(ctx, `UPDATE portfolio_account_sync_state SET failure_count=failure_count+1,
		next_attempt_at=$3,`+reasonColumn+`=$6,`+actionColumn+`=$7,`+retryColumn+`=$3,
		claim_id=NULL,claim_expires_at=NULL,claimed_resource=NULL,claimed_change_id=NULL,updated_at=$4
		WHERE user_id=$1 AND account_id=$2 AND claim_id=$5`, claim.Owner, claim.AccountID, retryAt, now, claim.ID, reason, action)
	return err
}

func syncFailureDiagnosticReason(failure string) portfolio.ResourceDiagnosticReason {
	switch failure {
	case syncFailureAuthorization:
		return portfolio.DiagnosticAuthorizationRequired
	case syncFailureRateLimited, syncFailureProviderUnavailable:
		return portfolio.DiagnosticProviderUnavailable
	default:
		return portfolio.DiagnosticUnknown
	}
}

func syncFailureDiagnosticAction(failure string) portfolio.ResourceDiagnosticAction {
	switch failure {
	case syncFailureAuthorization:
		return portfolio.DiagnosticActionReconnect
	case syncFailureRateLimited:
		return portfolio.DiagnosticActionWait
	default:
		return portfolio.DiagnosticActionRetry
	}
}

func diagnosticColumns(resource portfolio.AccountResource) (string, string, string, error) {
	switch resource {
	case portfolio.AccountResourceBalances:
		return "balances_failure_reason", "balances_failure_action", "balances_retry_at", nil
	case portfolio.AccountResourcePositions:
		return "positions_failure_reason", "positions_failure_action", "positions_retry_at", nil
	case portfolio.AccountResourceActivities:
		return "activities_failure_reason", "activities_failure_action", "activities_retry_at", nil
	default:
		return "", "", "", errors.New("unsupported portfolio sync resource")
	}
}

func publishScheduledSync(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, data portfolio.AccountData, now time.Time) error {
	switch claim.Resource {
	case portfolio.AccountResourceBalances:
		if err := publishBalances(ctx, tx, claim.Owner, claim.AccountID, claim.InclusionVersion, data.Balances, now); err != nil {
			return err
		}
	case portfolio.AccountResourcePositions:
		if err := publishPositions(ctx, tx, claim.Owner, claim.AccountID, claim.InclusionVersion, data.Positions, now); err != nil {
			return err
		}
	case portfolio.AccountResourceActivities:
		if err := publishActivities(ctx, tx, claim.Owner, claim.AccountID, claim.InclusionVersion, data.Activities, now); err != nil {
			return err
		}
	default:
		return errors.New("unsupported portfolio sync resource")
	}
	if err := checkpointSyncResource(ctx, tx, claim, now); err != nil {
		return err
	}
	if claim.ChangeID != nil {
		return commitCompletedInclusion(ctx, tx, claim, now)
	}
	return nil
}

func checkpointSyncResource(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, now time.Time) error {
	column := ""
	switch claim.Resource {
	case portfolio.AccountResourceBalances:
		column = "balances_success_at"
	case portfolio.AccountResourcePositions:
		column = "positions_success_at"
	case portfolio.AccountResourceActivities:
		column = "activities_success_at"
	default:
		return errors.New("unsupported portfolio sync resource")
	}
	reasonColumn, actionColumn, retryColumn, err := diagnosticColumns(claim.Resource)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE portfolio_account_sync_state SET `+column+`=$3,
		last_success_at=CASE WHEN
			(CASE WHEN $4='balances' THEN $3 ELSE balances_success_at END)>COALESCE(last_success_at,'-infinity') AND
			(CASE WHEN $4='positions' THEN $3 ELSE positions_success_at END)>COALESCE(last_success_at,'-infinity') AND
			(CASE WHEN $4='activities' THEN $3 ELSE activities_success_at END)>COALESCE(last_success_at,'-infinity')
			THEN $3 ELSE last_success_at END,
		initialized_at=CASE WHEN
			(CASE WHEN $4='balances' THEN $3 ELSE balances_success_at END)>COALESCE(last_success_at,'-infinity') AND
			(CASE WHEN $4='positions' THEN $3 ELSE positions_success_at END)>COALESCE(last_success_at,'-infinity') AND
			(CASE WHEN $4='activities' THEN $3 ELSE activities_success_at END)>COALESCE(last_success_at,'-infinity')
			THEN COALESCE(initialized_at,$3) ELSE initialized_at END,
		`+reasonColumn+`=NULL,`+actionColumn+`=NULL,`+retryColumn+`=NULL,
		next_attempt_at=NULL,failure_count=0,claim_id=NULL,claim_expires_at=NULL,
		claimed_resource=NULL,claimed_change_id=NULL,updated_at=$3
		WHERE user_id=$1 AND account_id=$2 AND claim_id=$5`, claim.Owner, claim.AccountID, now, claim.Resource, claim.ID)
	return err
}

func commitCompletedInclusion(ctx context.Context, tx pgx.Tx, claim portfolio.SyncClaim, now time.Time) error {
	var additions []string
	var incomplete bool
	err := tx.QueryRow(ctx, `SELECT change.addition_account_ids,
		EXISTS(SELECT 1 FROM portfolio_account_sync_state state WHERE state.user_id=change.user_id
			AND state.account_id=ANY(change.sync_account_ids) AND state.initialized_at IS NULL)
		FROM portfolio_inclusion_changes change WHERE change.id=$1 AND change.user_id=$2 AND change.status='pending' FOR UPDATE`,
		*claim.ChangeID, claim.Owner).Scan(&additions, &incomplete)
	if err != nil || incomplete {
		return err
	}
	if err := upsertIncludedAccounts(ctx, tx, claim.Owner, additions, claim.InclusionVersion, now); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE portfolio_inclusion_changes SET status='committed',failure_reason=NULL,updated_at=$3
		WHERE id=$1 AND user_id=$2`, *claim.ChangeID, claim.Owner, now)
	return err
}
