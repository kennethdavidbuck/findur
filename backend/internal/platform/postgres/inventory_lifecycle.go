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

const (
	provisionalBrokerageLabel = "Pending brokerage connection"
	provisionalAccountType    = "Pending account"
	provisionalAccountLabel   = "Account details unavailable"
	ownerArgument             = "owner"
	connectionArgument        = "connection"
	accountArgument           = "account"
	generationArgument        = "generation"
)

// InventoryLifecycleRepository owns guarded lifecycle changes to the current inventory head.
type InventoryLifecycleRepository struct{ pool *pgxpool.Pool }

// NewInventoryLifecycleRepository constructs the PostgreSQL lifecycle repository.
func NewInventoryLifecycleRepository(pool *pgxpool.Pool) *InventoryLifecycleRepository {
	return &InventoryLifecycleRepository{pool: pool}
}

// AddConnection records a safe provisional connection without degrading an existing row.
func (r *InventoryLifecycleRepository) AddConnection(ctx context.Context, subject, connectionID string, now time.Time) (bool, error) {
	return r.mutateInventory(ctx, subject, now, func(tx pgx.Tx, owner uuid.UUID, generation int64) (bool, error) {
		tag, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_connections
			(user_id, generation, connection_id, brokerage_label, status, sync_mode, available, eligible)
			VALUES (@owner, @generation, @connection, @brokerage, 'unavailable', 'unknown', false, false)
			ON CONFLICT (user_id, generation, connection_id) DO NOTHING`, pgx.StrictNamedArgs{
			ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID, "brokerage": provisionalBrokerageLabel,
		})
		return tag.RowsAffected() > 0, err
	})
}

// SetConnectionStatus applies the bounded active/disabled lifecycle projection.
func (r *InventoryLifecycleRepository) SetConnectionStatus(ctx context.Context, subject, connectionID string, status portfolio.ConnectionStatus, now time.Time) (bool, error) {
	return r.mutateInventory(ctx, subject, now, func(tx pgx.Tx, owner uuid.UUID, generation int64) (bool, error) {
		var query string
		switch status {
		case portfolio.ConnectionStatusDisabled:
			connectionTag, err := tx.Exec(ctx, `UPDATE portfolio_inventory_connections
				SET status = 'disabled', available = false, eligible = false,
					diagnostic_reason = 'connection_disabled', diagnostic_action = 'reconnect'
				WHERE user_id = @owner
					AND generation = @generation
					AND connection_id = @connection
					AND (status <> 'disabled' OR available OR eligible
						OR diagnostic_reason IS DISTINCT FROM 'connection_disabled'
						OR diagnostic_action IS DISTINCT FROM 'reconnect')`, pgx.StrictNamedArgs{
				ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID,
			})
			if err != nil {
				return false, err
			}
			accountTag, err := tx.Exec(ctx, `UPDATE portfolio_inventory_accounts
				SET available = false,
					eligible = false,
					selectable = false,
					usability_reason = 'connection_disabled'
				WHERE user_id = @owner
					AND generation = @generation
					AND connection_id = @connection
					AND (available OR eligible OR selectable
						OR usability_reason <> 'connection_disabled')`, pgx.StrictNamedArgs{
				ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID,
			})
			return connectionTag.RowsAffected()+accountTag.RowsAffected() > 0, err
		case portfolio.ConnectionStatusActive:
			query = `UPDATE portfolio_inventory_connections
				SET status = 'active', available = true, eligible = true,
					diagnostic_reason = NULL, diagnostic_action = NULL
				WHERE user_id = @owner
					AND generation = @generation
					AND connection_id = @connection
					AND (status <> 'active' OR NOT available OR NOT eligible
						OR diagnostic_reason IS NOT NULL OR diagnostic_action IS NOT NULL)`
		default:
			return false, errors.New("unsupported connection lifecycle status")
		}
		tag, err := tx.Exec(ctx, query, pgx.StrictNamedArgs{
			ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID,
		})
		return tag.RowsAffected() > 0, err
	})
}

// RemoveConnection removes only the current-head projection.
func (r *InventoryLifecycleRepository) RemoveConnection(ctx context.Context, subject, connectionID string, now time.Time) (bool, error) {
	return r.mutateInventory(ctx, subject, now, func(tx pgx.Tx, owner uuid.UUID, generation int64) (bool, error) {
		tag, err := tx.Exec(ctx, `DELETE FROM portfolio_inventory_connections
			WHERE user_id = @owner
				AND generation = @generation
				AND connection_id = @connection`, pgx.StrictNamedArgs{
			ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID,
		})
		return tag.RowsAffected() > 0, err
	})
}

// AddAccount records a safe non-selectable provisional account and stable identity.
func (r *InventoryLifecycleRepository) AddAccount(ctx context.Context, subject, connectionID, accountID string, now time.Time) (bool, error) {
	return r.mutateInventory(ctx, subject, now, func(tx pgx.Tx, owner uuid.UUID, generation int64) (bool, error) {
		return addProvisionalAccount(ctx, tx, owner, generation, connectionID, accountID, now)
	})
}

// RemoveAccount removes only the current-head projection.
func (r *InventoryLifecycleRepository) RemoveAccount(ctx context.Context, subject, accountID string, now time.Time) (bool, error) {
	return r.mutateInventory(ctx, subject, now, func(tx pgx.Tx, owner uuid.UUID, generation int64) (bool, error) {
		tag, err := tx.Exec(ctx, `DELETE FROM portfolio_inventory_accounts
			WHERE user_id = @owner
				AND generation = @generation
				AND account_id = @account`, pgx.StrictNamedArgs{
			ownerArgument: owner, generationArgument: generation, accountArgument: accountID,
		})
		return tag.RowsAffected() > 0, err
	})
}

type inventoryLifecycleMutation func(pgx.Tx, uuid.UUID, int64) (bool, error)

func (r *InventoryLifecycleRepository) mutateInventory(ctx context.Context, subject string, now time.Time, mutate inventoryLifecycleMutation) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	owner, found, err := lockInventoryLifecycleOwner(ctx, tx, subject)
	if err != nil {
		return false, err
	}
	if !found {
		return false, tx.Commit(ctx)
	}
	generation, found, err := lockCurrentInventoryHead(ctx, tx, owner, now)
	if err != nil {
		return false, err
	}
	if !found {
		return false, portfolio.ErrInventoryBusy
	}
	changed, err := mutate(tx, owner, generation)
	if err != nil {
		return false, err
	}
	if changed {
		if err := removeIneligibleIncludedAccounts(ctx, tx, owner, generation, now); err != nil {
			return false, err
		}
		if err := recomputeCurrentInventoryStatus(ctx, tx, owner, generation, now); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return changed, nil
}

func recomputeCurrentInventoryStatus(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, now time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE portfolio_inventory_state state
		SET current_status = CASE
				WHEN EXISTS (
					SELECT 1
					FROM portfolio_inventory_connections connection
					WHERE connection.user_id = @owner
						AND connection.generation = @generation
				)
				AND NOT EXISTS (
					SELECT 1
					FROM portfolio_inventory_connections connection
					WHERE connection.user_id = @owner
						AND connection.generation = @generation
						AND connection.status <> 'disabled'
				)
				THEN 'disabled'
				WHEN EXISTS (
					SELECT 1
					FROM portfolio_inventory_accounts account
					WHERE account.user_id = @owner
						AND account.generation = @generation
				)
				THEN 'ready'
				ELSE 'empty'
			END,
			claim_expires_at = NULL,
			updated_at = @now
		WHERE state.user_id = @owner
			AND state.head_generation = @generation
			AND (state.current_status <> 'pending' OR state.claim_expires_at <= @now)`, pgx.StrictNamedArgs{
		ownerArgument: owner, generationArgument: generation, "now": now,
	})
	return err
}

func lockInventoryLifecycleOwner(ctx context.Context, tx pgx.Tx, subject string) (uuid.UUID, bool, error) {
	var owner uuid.UUID
	err := tx.QueryRow(ctx, `SELECT identity.user_id
		FROM external_identities identity
		JOIN users owner ON owner.id = identity.user_id
		WHERE identity.provider = @provider
			AND identity.subject = @subject
			AND owner.active
		FOR UPDATE OF owner`, pgx.StrictNamedArgs{
		"provider": auth.SnapTradeProvider,
		"subject":  subject,
	}).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return owner, err == nil, err
}

func lockCurrentInventoryHead(ctx context.Context, tx pgx.Tx, owner uuid.UUID, now time.Time) (int64, bool, error) {
	var generation *int64
	var status portfolio.State
	var claimExpiresAt *time.Time
	err := tx.QueryRow(ctx, `SELECT head_generation, current_status, claim_expires_at
		FROM portfolio_inventory_state
		WHERE user_id = @owner
		FOR UPDATE`, pgx.StrictNamedArgs{ownerArgument: owner}).Scan(&generation, &status, &claimExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if status == portfolio.StatePending && claimExpiresAt != nil && claimExpiresAt.After(now) {
		return 0, false, portfolio.ErrInventoryBusy
	}
	if generation == nil {
		return 0, false, nil
	}
	return *generation, true, nil
}

func addProvisionalAccount(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, connectionID, accountID string, now time.Time) (bool, error) {
	connectionTag, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_connections
		(user_id, generation, connection_id, brokerage_label, status, sync_mode, available, eligible)
		VALUES (@owner, @generation, @connection, @brokerage, 'unavailable', 'unknown', false, false)
		ON CONFLICT (user_id, generation, connection_id) DO NOTHING`, pgx.StrictNamedArgs{
		ownerArgument: owner, generationArgument: generation, connectionArgument: connectionID, "brokerage": provisionalBrokerageLabel,
	})
	if err != nil {
		return false, err
	}
	accountTag, err := tx.Exec(ctx, `INSERT INTO portfolio_inventory_accounts
		(user_id, generation, account_id, connection_id, category, account_type, masked_label,
		 available, eligible, sync_state, selectable, usability_reason)
		VALUES (@owner, @generation, @account, @connection, 'unknown', @account_type, @account_label,
		 false, false, 'unknown', false, 'provisional_category')
		ON CONFLICT (user_id, generation, account_id) DO UPDATE
		SET connection_id = EXCLUDED.connection_id
		WHERE portfolio_inventory_accounts.connection_id IS DISTINCT FROM EXCLUDED.connection_id`, pgx.StrictNamedArgs{
		ownerArgument: owner, generationArgument: generation, accountArgument: accountID, connectionArgument: connectionID,
		"account_type": provisionalAccountType, "account_label": provisionalAccountLabel,
	})
	if err != nil {
		return false, err
	}
	identityTag, err := tx.Exec(ctx, `INSERT INTO portfolio_account_identities
		(user_id, account_id, connection_id, first_seen_at, last_seen_at)
		VALUES (@owner, @account, @connection, @now, @now)
		ON CONFLICT (user_id, account_id) DO UPDATE
		SET connection_id = EXCLUDED.connection_id,
			last_seen_at = EXCLUDED.last_seen_at
		WHERE portfolio_account_identities.connection_id IS DISTINCT FROM EXCLUDED.connection_id`, pgx.StrictNamedArgs{
		ownerArgument: owner, accountArgument: accountID, connectionArgument: connectionID, "now": now,
	})
	if err != nil {
		return false, err
	}
	return connectionTag.RowsAffected()+accountTag.RowsAffected()+identityTag.RowsAffected() > 0, nil
}
