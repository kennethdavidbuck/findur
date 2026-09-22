package postgres

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

// InclusionRepository owns removal-first account selection and guarded dataset publication.
type InclusionRepository struct{ pool *pgxpool.Pool }

// NewInclusionRepository constructs the PostgreSQL inclusion repository.
func NewInclusionRepository(pool *pgxpool.Pool) *InclusionRepository {
	return &InclusionRepository{pool: pool}
}

// GetInclusion returns only the specified owner's committed and latest change state.
func (r *InclusionRepository) GetInclusion(ctx context.Context, owner uuid.UUID) (portfolio.InclusionSnapshot, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.InclusionSnapshot{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	snapshot, err := loadInclusionSnapshot(ctx, tx, owner, false)
	if err != nil {
		return portfolio.InclusionSnapshot{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.InclusionSnapshot{}, err
	}
	return snapshot, nil
}

// PrepareInclusion validates ownership, fences the change, and purges removals.
func (r *InclusionRepository) PrepareInclusion(ctx context.Context, owner uuid.UUID, expectedVersion int64, idempotencyKey string, targets []string, now time.Time) (portfolio.InclusionPreparation, error) {
	if idempotencyKey == "" || len(idempotencyKey) > 200 || expectedVersion < 0 {
		return portfolio.InclusionPreparation{}, portfolio.ErrInvalidAccountSelection
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockActiveOwner(ctx, tx, owner); err != nil {
		return portfolio.InclusionPreparation{}, err
	}

	if replay, found, err := loadReplay(ctx, tx, owner, idempotencyKey, expectedVersion, targets); err != nil {
		return portfolio.InclusionPreparation{}, err
	} else if found {
		if err := tx.Commit(ctx); err != nil {
			return portfolio.InclusionPreparation{}, err
		}
		return replay, nil
	}
	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inclusion_state (user_id,version,lifecycle_generation,updated_at) VALUES ($1,0,1,$2) ON CONFLICT (user_id) DO NOTHING`, owner, now); err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	var version, lifecycleGeneration int64
	if err := tx.QueryRow(ctx, `SELECT version,lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&version, &lifecycleGeneration); err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	// A concurrent request with the same key may have committed while this
	// transaction waited for the owner lock. Re-read before applying anything.
	if replay, found, err := loadReplay(ctx, tx, owner, idempotencyKey, expectedVersion, targets); err != nil {
		return portfolio.InclusionPreparation{}, err
	} else if found {
		if err := tx.Commit(ctx); err != nil {
			return portfolio.InclusionPreparation{}, err
		}
		return replay, nil
	}
	if version != expectedVersion {
		return portfolio.InclusionPreparation{}, portfolio.ErrInclusionConflict
	}

	var inventoryGeneration int64
	var inventoryHead *int64
	if err := tx.QueryRow(ctx, `SELECT current_generation,head_generation FROM portfolio_inventory_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&inventoryGeneration, &inventoryHead); err != nil || inventoryHead == nil {
		if errors.Is(err, pgx.ErrNoRows) || inventoryHead == nil {
			return portfolio.InclusionPreparation{}, portfolio.ErrInvalidAccountSelection
		}
		return portfolio.InclusionPreparation{}, err
	}
	committed, err := loadCommittedAccountIDs(ctx, tx, owner)
	if err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	if err := validateSelection(ctx, tx, owner, *inventoryHead, targets, committed); err != nil {
		return portfolio.InclusionPreparation{}, err
	}

	preparation, err := prepareInclusionChange(ctx, tx, owner, idempotencyKey, targets, committed, version, inventoryGeneration, lifecycleGeneration, now)
	if err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	return preparation, nil
}

// prepareInclusionChange requires the caller to hold the owner, inclusion, and
// inventory locks and to have validated the expected version and target accounts.
func prepareInclusionChange(ctx context.Context, tx pgx.Tx, owner uuid.UUID, idempotencyKey string, targets, committed []string, version, inventoryGeneration, lifecycleGeneration int64, now time.Time) (portfolio.InclusionPreparation, error) {
	additions, removals := difference(targets, committed), difference(committed, targets)
	changeID := uuid.New()
	resultVersion := version
	status := portfolio.InclusionCommitted
	if len(additions) > 0 || len(removals) > 0 {
		resultVersion++
		if len(removals) > 0 {
			lifecycleGeneration++
		}
		status = portfolio.InclusionPending
		if len(additions) == 0 {
			status = portfolio.InclusionCommitted
		}
		if _, err := tx.Exec(ctx, `UPDATE portfolio_inclusion_state SET version=$2,lifecycle_generation=$3,updated_at=$4 WHERE user_id=$1`, owner, resultVersion, lifecycleGeneration, now); err != nil {
			return portfolio.InclusionPreparation{}, err
		}
		if len(removals) > 0 {
			if err := purgeExcludedAccounts(ctx, tx, owner, removals); err != nil {
				return portfolio.InclusionPreparation{}, err
			}
		}
	}

	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_inclusion_changes
		(id,user_id,idempotency_key,expected_version,result_version,inventory_generation,lifecycle_generation,target_account_ids,addition_account_ids,removal_account_ids,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)`, changeID, owner, idempotencyKey, version, resultVersion, inventoryGeneration, lifecycleGeneration, targets, additions, removals, status, now); err != nil {
		return portfolio.InclusionPreparation{}, err
	}
	preparation := portfolio.InclusionPreparation{
		InclusionSnapshot: portfolio.InclusionSnapshot{
			Version:   resultVersion,
			Committed: difference(committed, removals),
			Change: &portfolio.InclusionChange{
				ID:        changeID,
				Status:    status,
				Additions: additions,
				Removals:  removals,
			},
		},
		Claimed:             len(additions) > 0,
		ChangeID:            changeID,
		InventoryGeneration: inventoryGeneration,
		LifecycleGeneration: lifecycleGeneration,
		Additions:           additions,
	}
	if preparation.Claimed {
		if err := tx.QueryRow(ctx, `SELECT access_token_encrypted,envelope_version FROM provider_authorizations WHERE user_id=$1 AND provider=$2`, owner, auth.SnapTradeProvider).Scan(&preparation.EncryptedToken, &preparation.TokenVersion); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return portfolio.InclusionPreparation{}, err
		}
	}
	return preparation, nil
}

func purgeExcludedAccounts(ctx context.Context, tx pgx.Tx, owner uuid.UUID, removals []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM portfolio_included_accounts WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM portfolio_balance_versions WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM portfolio_position_versions WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `DELETE FROM portfolio_activity_versions WHERE user_id=$1 AND account_id=ANY($2)`, owner, removals)
	return err
}

// FinalizeInclusion atomically publishes complete guarded additions or records failure.
func (r *InclusionRepository) FinalizeInclusion(ctx context.Context, owner uuid.UUID, changeID uuid.UUID, inclusionVersion, inventoryGeneration, lifecycleGeneration int64, data map[string]portfolio.AccountData, failure string, now time.Time) (portfolio.InclusionSnapshot, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockActiveOwner(ctx, tx, owner); err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}

	var status portfolio.InclusionChangeStatus
	var additions []string
	var changeInventory, changeLifecycle, resultVersion int64
	if err := tx.QueryRow(ctx, `SELECT status,addition_account_ids,inventory_generation,lifecycle_generation,result_version FROM portfolio_inclusion_changes WHERE id=$1 AND user_id=$2 FOR UPDATE`, changeID, owner).Scan(&status, &additions, &changeInventory, &changeLifecycle, &resultVersion); err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}
	if status != portfolio.InclusionPending {
		snapshot, err := loadInclusionSnapshot(ctx, tx, owner, false)
		if err == nil {
			err = tx.Commit(ctx)
		}
		return snapshot, false, err
	}

	var currentVersion, currentLifecycle, currentInventory int64
	var inventoryHead *int64
	guardErr := tx.QueryRow(ctx, `SELECT version,lifecycle_generation FROM portfolio_inclusion_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&currentVersion, &currentLifecycle)
	if guardErr == nil {
		guardErr = tx.QueryRow(ctx, `SELECT current_generation,head_generation FROM portfolio_inventory_state WHERE user_id=$1 FOR UPDATE`, owner).Scan(&currentInventory, &inventoryHead)
	}
	if guardErr != nil && !errors.Is(guardErr, pgx.ErrNoRows) {
		return portfolio.InclusionSnapshot{}, false, guardErr
	}
	guarded := guardErr == nil && inventoryHead != nil &&
		currentVersion == inclusionVersion && resultVersion == inclusionVersion &&
		currentLifecycle == lifecycleGeneration && changeLifecycle == lifecycleGeneration &&
		currentInventory == inventoryGeneration && changeInventory == inventoryGeneration
	if guarded {
		guarded = validateSelection(ctx, tx, owner, *inventoryHead, additions, nil) == nil
	}
	if !guarded {
		failure = "stale_guard"
	}
	if failure == "" && !completeAccountData(additions, data) {
		failure = "unusable_data"
	}

	if failure != "" {
		if _, err := tx.Exec(ctx, `UPDATE portfolio_inclusion_changes SET status='failed',failure_reason=$3,updated_at=$4 WHERE id=$1 AND user_id=$2`, changeID, owner, failure, now); err != nil {
			return portfolio.InclusionSnapshot{}, false, err
		}
	} else if err := publishAccountData(ctx, tx, owner, inclusionVersion, additions, data, now); err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	} else if _, err := tx.Exec(ctx, `UPDATE portfolio_inclusion_changes SET status='committed',failure_reason=NULL,updated_at=$3 WHERE id=$1 AND user_id=$2`, changeID, owner, now); err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}

	snapshot, err := loadInclusionSnapshot(ctx, tx, owner, false)
	if err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.InclusionSnapshot{}, false, err
	}
	return snapshot, failure == "", nil
}

func lockActiveOwner(ctx context.Context, tx pgx.Tx, owner uuid.UUID) error {
	var active bool
	if err := tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1 FOR UPDATE`, owner).Scan(&active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return portfolio.ErrInvalidAccountSelection
		}
		return err
	}
	if !active {
		return portfolio.ErrInvalidAccountSelection
	}
	return nil
}

func loadReplay(ctx context.Context, tx pgx.Tx, owner uuid.UUID, key string, expected int64, targets []string) (portfolio.InclusionPreparation, bool, error) {
	var change portfolio.InclusionChange
	var storedExpected, version int64
	var storedTargets []string
	err := tx.QueryRow(ctx, `SELECT id,status,expected_version,result_version,target_account_ids,addition_account_ids,removal_account_ids,COALESCE(failure_reason,'') FROM portfolio_inclusion_changes WHERE user_id=$1 AND idempotency_key=$2`, owner, key).Scan(&change.ID, &change.Status, &storedExpected, &version, &storedTargets, &change.Additions, &change.Removals, &change.FailureReason)
	if errors.Is(err, pgx.ErrNoRows) {
		return portfolio.InclusionPreparation{}, false, nil
	}
	if err != nil {
		return portfolio.InclusionPreparation{}, false, err
	}
	if storedExpected != expected || !slices.Equal(storedTargets, targets) {
		return portfolio.InclusionPreparation{}, false, portfolio.ErrIdempotencyConflict
	}
	committed, err := loadCommittedAccountIDs(ctx, tx, owner)
	return portfolio.InclusionPreparation{
		InclusionSnapshot: portfolio.InclusionSnapshot{
			Version:   version,
			Committed: committed,
			Change:    &change,
		},
	}, true, err
}

func loadInclusionSnapshot(ctx context.Context, tx pgx.Tx, owner uuid.UUID, lock bool) (portfolio.InclusionSnapshot, error) {
	query := `SELECT version FROM portfolio_inclusion_state WHERE user_id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	var snapshot portfolio.InclusionSnapshot
	err := tx.QueryRow(ctx, query, owner).Scan(&snapshot.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return snapshot, nil
	}
	if err != nil {
		return snapshot, err
	}
	snapshot.Committed, err = loadCommittedAccountIDs(ctx, tx, owner)
	if err != nil {
		return portfolio.InclusionSnapshot{}, err
	}
	var change portfolio.InclusionChange
	err = tx.QueryRow(ctx, `SELECT id,status,addition_account_ids,removal_account_ids,COALESCE(failure_reason,'') FROM portfolio_inclusion_changes WHERE user_id=$1 ORDER BY result_version DESC,created_at DESC,id DESC LIMIT 1`, owner).Scan(&change.ID, &change.Status, &change.Additions, &change.Removals, &change.FailureReason)
	if err == nil {
		snapshot.Change = &change
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return portfolio.InclusionSnapshot{}, err
	}
	return snapshot, nil
}

func loadCommittedAccountIDs(ctx context.Context, tx pgx.Tx, owner uuid.UUID) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT account_id FROM portfolio_included_accounts WHERE user_id=$1 ORDER BY account_id`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

func validateSelection(ctx context.Context, tx pgx.Tx, owner uuid.UUID, generation int64, targets, committed []string) error {
	committedSet := make(map[string]bool, len(committed))
	for _, id := range committed {
		committedSet[id] = true
	}
	for _, id := range targets {
		if committedSet[id] {
			continue
		}
		var selectable bool
		err := tx.QueryRow(ctx, `SELECT account.selectable FROM portfolio_inventory_accounts account
			JOIN portfolio_inventory_connections connection USING (user_id,generation,connection_id)
			WHERE account.user_id=$1 AND account.generation=$2 AND account.account_id=$3
			AND `+selectableInventoryAccountSQL, owner, generation, id).Scan(&selectable)
		if errors.Is(err, pgx.ErrNoRows) || err == nil && !selectable {
			return portfolio.ErrInvalidAccountSelection
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func difference(left, right []string) []string {
	other := make(map[string]bool, len(right))
	for _, value := range right {
		other[value] = true
	}
	result := make([]string, 0)
	for _, value := range left {
		if !other[value] {
			result = append(result, value)
		}
	}
	return result
}

func completeAccountData(additions []string, data map[string]portfolio.AccountData) bool {
	if len(data) != len(additions) {
		return false
	}
	for _, id := range additions {
		value, ok := data[id]
		if !ok {
			return false
		}
		if value.Balances.RetrievedAt.IsZero() || value.Balances.Rows == nil {
			return false
		}
		if value.Positions.ObservedAt.IsZero() || value.Positions.RetrievedAt.IsZero() || value.Positions.Rows == nil {
			return false
		}
		if value.Activities.RetrievedAt.IsZero() || value.Activities.Rows == nil {
			return false
		}
	}
	return true
}

func publishAccountData(ctx context.Context, tx pgx.Tx, owner uuid.UUID, inclusionVersion int64, additions []string, data map[string]portfolio.AccountData, now time.Time) error {
	for _, accountID := range additions {
		value := data[accountID]
		if err := publishBalances(ctx, tx, owner, accountID, inclusionVersion, value.Balances, now); err != nil {
			return err
		}
		if err := publishPositions(ctx, tx, owner, accountID, inclusionVersion, value.Positions, now); err != nil {
			return err
		}
		if err := publishActivities(ctx, tx, owner, accountID, inclusionVersion, value.Activities, now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_included_accounts (user_id,account_id,inclusion_version,included_at)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (user_id,account_id) DO UPDATE
			SET inclusion_version=EXCLUDED.inclusion_version,included_at=EXCLUDED.included_at`, owner, accountID, inclusionVersion, now); err != nil {
			return err
		}
	}
	return nil
}

func publishBalances(ctx context.Context, tx pgx.Tx, owner uuid.UUID, accountID string, inclusionVersion int64, data portfolio.BalanceDataset, now time.Time) error {
	versionID := uuid.New()
	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_balance_versions (id,user_id,account_id,inclusion_version,observed_at,retrieved_at,published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, versionID, owner, accountID, inclusionVersion, data.ObservedAt, data.RetrievedAt, now); err != nil {
		return err
	}
	for index, row := range data.Rows {
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_balance_rows (version_id,row_number,currency,cash,buying_power)
			VALUES ($1,$2,$3,$4,$5)`, versionID, index, row.Currency, row.Cash, row.BuyingPower); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, `INSERT INTO portfolio_balance_heads (user_id,account_id,version_id)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id,account_id) DO UPDATE SET version_id=EXCLUDED.version_id`, owner, accountID, versionID)
	return err
}

func publishPositions(ctx context.Context, tx pgx.Tx, owner uuid.UUID, accountID string, inclusionVersion int64, data portfolio.PositionDataset, now time.Time) error {
	versionID := uuid.New()
	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_position_versions (id,user_id,account_id,inclusion_version,observed_at,retrieved_at,published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, versionID, owner, accountID, inclusionVersion, data.ObservedAt, data.RetrievedAt, now); err != nil {
		return err
	}
	for index, row := range data.Rows {
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_position_rows (version_id,row_number,instrument_id,symbol,kind,currency,units,price,cost_basis)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, versionID, index, row.InstrumentID, row.Symbol, row.Kind, row.Currency, row.Units, row.Price, row.CostBasis); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, `INSERT INTO portfolio_position_heads (user_id,account_id,version_id)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id,account_id) DO UPDATE SET version_id=EXCLUDED.version_id`, owner, accountID, versionID)
	return err
}

func publishActivities(ctx context.Context, tx pgx.Tx, owner uuid.UUID, accountID string, inclusionVersion int64, data portfolio.ActivityDataset, now time.Time) error {
	versionID := uuid.New()
	if _, err := tx.Exec(ctx, `INSERT INTO portfolio_activity_versions (id,user_id,account_id,inclusion_version,observed_at,retrieved_at,published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, versionID, owner, accountID, inclusionVersion, data.ObservedAt, data.RetrievedAt, now); err != nil {
		return err
	}
	for index, row := range data.Rows {
		if _, err := tx.Exec(ctx, `INSERT INTO portfolio_activity_rows (version_id,row_number,activity_id,activity_type,trade_date,currency,amount,fee,price,units)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, versionID, index, row.ID, row.Type, row.TradeDate, row.Currency, row.Amount, row.Fee, row.Price, row.Units); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, `INSERT INTO portfolio_activity_heads (user_id,account_id,version_id)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id,account_id) DO UPDATE SET version_id=EXCLUDED.version_id`, owner, accountID, versionID)
	return err
}
