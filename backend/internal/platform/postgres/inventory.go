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

// InventoryClaimLease bounds how long a crashed provider operation can block recovery.
const InventoryClaimLease = 30 * time.Second

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

// Prepare returns persisted inventory or atomically claims a bounded bootstrap/retry.
func (r *InventoryRepository) Prepare(ctx context.Context, owner uuid.UUID, retry bool, now time.Time) (portfolio.Preparation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	preparation, err := prepareInventoryClaim(ctx, tx, owner, retry, now)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	if preparation.Claimed {
		if err := loadInventoryAuthorization(ctx, tx, owner, &preparation); err != nil {
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
	claimExpiresAt := now.Add(InventoryClaimLease)
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
	preparation.Generation++
	claimExpiresAt := now.Add(InventoryClaimLease)
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_generation=$2,current_status='pending',retry_at=NULL,claim_expires_at=$3,updated_at=$4 WHERE user_id=$1`, owner, preparation.Generation, claimExpiresAt, now)
	if err != nil {
		return portfolio.Preparation{}, err
	}
	preparation.State, preparation.RetryAt, preparation.ClaimExpiresAt, preparation.UpdatedAt, preparation.Claimed = portfolio.StatePending, nil, &claimExpiresAt, now, true
	return preparation, nil
}

func loadInventoryAuthorization(ctx context.Context, tx pgx.Tx, owner uuid.UUID, preparation *portfolio.Preparation) error {
	err := tx.QueryRow(ctx, `SELECT access_token_encrypted,envelope_version FROM provider_authorizations WHERE user_id=$1 AND provider=$2`, owner, auth.SnapTradeProvider).Scan(&preparation.EncryptedToken, &preparation.TokenVersion)
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
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_status=$3,head_generation=$2,retry_at=$4,claim_expires_at=NULL,updated_at=$5
		WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, owner, generation, state, retryAt, now)
	return err
}

func insertInventoryRows(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connections []portfolio.Connection, now time.Time) error {
	for _, connection := range connections {
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_connections
			(user_id,generation,connection_id,brokerage_label,status,sync_mode,available,eligible)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, owner, generation, connection.ID, connection.BrokerageLabel, connection.Status, connection.SyncMode, connection.Available, connection.Eligible); err != nil {
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
			(user_id,generation,account_id,connection_id,category,account_type,masked_label,available,eligible,sync_state,selectable,usability_reason)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, owner, generation, account.ID, connection.ID, account.Category, account.Type, account.MaskedLabel, account.Available, account.Eligible, account.SyncState, selectable, reason); err != nil {
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
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state SET current_status=$3,retry_at=$4,claim_expires_at=NULL,updated_at=$5
		WHERE user_id=$1 AND current_generation=$2 AND current_status='pending'`, owner, generation, state, retryAt, now)
	return err
}

func loadPreparation(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (portfolio.Preparation, bool, error) {
	result, head, found, err := loadInventoryState(ctx, tx, owner)
	if err != nil || !found {
		return result, found, err
	}
	if head == nil {
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
	if result.State == portfolio.StateReady && accountCount == 0 {
		result.State = portfolio.StateEmpty
	}
	return result, true, nil
}

func loadInventoryState(ctx context.Context, tx pgx.Tx, owner uuid.UUID) (portfolio.Preparation, *int64, bool, error) {
	var result portfolio.Preparation
	var head *int64
	err := tx.QueryRow(ctx, `SELECT current_generation,current_status,head_generation,retry_at,claim_expires_at,updated_at
		FROM portfolio_inventory_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&result.Generation, &result.State, &head, &result.RetryAt, &result.ClaimExpiresAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return portfolio.Preparation{}, nil, false, nil
	}
	return result, head, err == nil, err
}

func loadInventoryConnections(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64) ([]portfolio.Connection, error) {
	rows, err := tx.Query(ctx, `SELECT connection_id,brokerage_label,status,sync_mode,available,eligible
		FROM portfolio_inventory_connections WHERE user_id=$1 AND generation=$2 ORDER BY connection_id`, owner, generation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var connections []portfolio.Connection
	for rows.Next() {
		var connection portfolio.Connection
		if err := rows.Scan(&connection.ID, &connection.BrokerageLabel, &connection.Status, &connection.SyncMode, &connection.Available, &connection.Eligible); err != nil {
			return nil, err
		}
		connections = append(connections, connection)
	}
	return connections, rows.Err()
}

func loadInventoryAccounts(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connectionID string) ([]portfolio.Account, error) {
	rows, err := tx.Query(ctx, `SELECT account.account_id,account.category,account.account_type,account.masked_label,account.available,account.eligible,account.sync_state,account.selectable,account.usability_reason
		FROM portfolio_inventory_accounts account
		JOIN portfolio_inventory_connections connection USING (user_id,generation,connection_id)
		WHERE account.user_id=$1 AND account.generation=$2 AND account.connection_id=$3
		AND `+selectableInventoryAccountSQL+` ORDER BY account.account_id`, owner, generation, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []portfolio.Account
	for rows.Next() {
		var account portfolio.Account
		if err := rows.Scan(&account.ID, &account.Category, &account.Type, &account.MaskedLabel, &account.Available, &account.Eligible, &account.SyncState, &account.Selectable, &account.UsabilityReason); err != nil {
			return nil, err
		}
		// Matching today's predicate is sufficient, including legacy null
		// provider status/category values. Project current semantics without
		// rewriting immutable inventory history.
		account.Selectable = true
		if account.Category == portfolio.AccountCategoryInvestment {
			account.Eligible, account.UsabilityReason = true, portfolio.UsabilityReady
		} else {
			account.Eligible = false
			if account.UsabilityReason != portfolio.UsabilityProvisionalStatus {
				account.UsabilityReason = portfolio.UsabilityProvisionalCategory
			}
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}
