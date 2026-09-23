package postgres

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

// ShowcaseRepository projects committed included accounts and current heads only.
type ShowcaseRepository struct {
	pool  *pgxpool.Pool
	clock func() time.Time
}

type showcaseAccountHead struct {
	id                                    string
	account                               portfolio.ShowcaseAccount
	connectionStatus, accountSyncState    string
	connectionAvailable, accountAvailable bool
	diagnostics                           map[portfolio.AccountResource]*portfolio.ResourceDiagnostic
}

func (h showcaseAccountHead) usable() bool {
	return h.connectionStatus == string(portfolio.ConnectionStatusActive) &&
		h.connectionAvailable && h.accountAvailable &&
		h.accountSyncState == string(portfolio.AccountSyncStateComplete) &&
		h.account.SyncMode != portfolio.SyncModeUnknown
}

// NewShowcaseRepository creates a PostgreSQL-backed owner-scoped showcase reader.
func NewShowcaseRepository(pool *pgxpool.Pool, clock func() time.Time) *ShowcaseRepository {
	return &ShowcaseRepository{pool: pool, clock: clock}
}

// GetShowcase returns committed dataset heads for accounts included by owner.
func (r *ShowcaseRepository) GetShowcase(ctx context.Context, owner uuid.UUID) (portfolio.Showcase, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return portfolio.Showcase{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	accounts, err := loadShowcaseAccounts(ctx, tx, owner)
	if err != nil {
		return portfolio.Showcase{}, err
	}

	now := r.clock()
	result := portfolio.Showcase{Accounts: make([]portfolio.ShowcaseAccount, 0, len(accounts))}
	for _, head := range accounts {
		if !head.usable() {
			reason, action := unusableAccountDiagnostic(head.connectionStatus, head.accountSyncState)
			diagnostic := &portfolio.ResourceDiagnostic{Reason: reason, RecommendedAction: action}
			head.account.Balances = unavailable("SnapTrade", "included account", diagnostic)
			head.account.Positions = unavailable("SnapTrade", "included account", diagnostic)
			head.account.Activities = unavailable("SnapTrade", "included account; newest 50 accumulated activities", diagnostic)
		} else {
			head.account.Balances = r.balances(ctx, tx, owner, head.id, head.account.SyncMode, now, head.diagnostics[portfolio.AccountResourceBalances])
			head.account.Positions = r.positions(ctx, tx, owner, head.id, head.account.SyncMode, now, head.diagnostics[portfolio.AccountResourcePositions])
			head.account.Activities = r.activities(ctx, tx, owner, head.id, head.account.SyncMode, now, head.diagnostics[portfolio.AccountResourceActivities])
		}
		result.Accounts = append(result.Accounts, head.account)
	}

	if err := tx.Commit(ctx); err != nil {
		return portfolio.Showcase{}, err
	}
	return result, nil
}

func unusableAccountDiagnostic(connectionStatus, accountSyncState string) (portfolio.ResourceDiagnosticReason, portfolio.ResourceDiagnosticAction) {
	if connectionStatus == string(portfolio.ConnectionStatusDisabled) {
		return portfolio.DiagnosticAuthorizationRequired, portfolio.DiagnosticActionReconnect
	}
	switch accountSyncState {
	case string(portfolio.AccountSyncStatePending):
		return portfolio.DiagnosticSyncPending, portfolio.DiagnosticActionWait
	case string(portfolio.AccountSyncStateUnavailable):
		return portfolio.DiagnosticProviderUnavailable, portfolio.DiagnosticActionRetry
	default:
		return portfolio.DiagnosticUnknown, portfolio.DiagnosticActionRetry
	}
}

func loadShowcaseAccounts(ctx context.Context, tx pgx.Tx, owner uuid.UUID) ([]showcaseAccountHead, error) {
	rows, err := tx.Query(ctx, `SELECT
			a.account_id,
			a.masked_label,
			c.brokerage_label,
			c.sync_mode,
			c.status,
			c.available,
			a.available,
			a.sync_state,
			sync.balances_failure_reason,
			sync.balances_failure_action,
			sync.balances_retry_at,
			sync.balances_success_at,
			sync.positions_failure_reason,
			sync.positions_failure_action,
			sync.positions_retry_at,
			sync.positions_success_at,
			sync.activities_failure_reason,
			sync.activities_failure_action,
			sync.activities_retry_at,
			sync.activities_success_at
		FROM portfolio_included_accounts i JOIN portfolio_inventory_state s ON s.user_id=i.user_id
		JOIN portfolio_inventory_accounts a ON a.user_id=i.user_id AND a.account_id=i.account_id AND a.generation=s.head_generation
		JOIN portfolio_inventory_connections c ON c.user_id=a.user_id AND c.generation=a.generation AND c.connection_id=a.connection_id
		JOIN portfolio_account_sync_state sync ON sync.user_id=i.user_id AND sync.account_id=i.account_id
		WHERE i.user_id=$1 ORDER BY a.masked_label`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]showcaseAccountHead, 0)
	for rows.Next() {
		var head showcaseAccountHead
		var balanceReason, positionReason, activityReason *portfolio.ResourceDiagnosticReason
		var balanceAction, positionAction, activityAction *portfolio.ResourceDiagnosticAction
		var balanceRetry, positionRetry, activityRetry *time.Time
		var balanceSuccess, positionSuccess, activitySuccess *time.Time
		if err := rows.Scan(
			&head.id,
			&head.account.Label,
			&head.account.Brokerage,
			&head.account.SyncMode,
			&head.connectionStatus,
			&head.connectionAvailable,
			&head.accountAvailable,
			&head.accountSyncState,
			&balanceReason,
			&balanceAction,
			&balanceRetry,
			&balanceSuccess,
			&positionReason,
			&positionAction,
			&positionRetry,
			&positionSuccess,
			&activityReason,
			&activityAction,
			&activityRetry,
			&activitySuccess,
		); err != nil {
			return nil, err
		}
		head.diagnostics = map[portfolio.AccountResource]*portfolio.ResourceDiagnostic{
			portfolio.AccountResourceBalances:   resourceDiagnostic(balanceReason, balanceAction, balanceRetry, balanceSuccess),
			portfolio.AccountResourcePositions:  resourceDiagnostic(positionReason, positionAction, positionRetry, positionSuccess),
			portfolio.AccountResourceActivities: resourceDiagnostic(activityReason, activityAction, activityRetry, activitySuccess),
		}
		accounts = append(accounts, head)
	}
	return accounts, rows.Err()
}

func resourceDiagnostic(reason *portfolio.ResourceDiagnosticReason, action *portfolio.ResourceDiagnosticAction, retryAt, lastSuccessfulAt *time.Time) *portfolio.ResourceDiagnostic {
	if reason == nil {
		if lastSuccessfulAt == nil {
			return &portfolio.ResourceDiagnostic{
				Reason:            portfolio.DiagnosticSyncPending,
				RecommendedAction: portfolio.DiagnosticActionWait,
			}
		}
		return nil
	}
	recommendedAction := portfolio.DiagnosticActionRetry
	if action != nil {
		recommendedAction = *action
	}
	return &portfolio.ResourceDiagnostic{Reason: *reason, RecommendedAction: recommendedAction, RetryAt: retryAt, LastSuccessfulAt: lastSuccessfulAt}
}

func datasetContext(source, coverage, currency string, observed *time.Time, retrieved, published *time.Time, mode portfolio.SyncMode, activities bool, now time.Time, diagnostic *portfolio.ResourceDiagnostic) portfolio.DatasetContext {
	freshness := portfolio.ClassifyFreshness(mode, activities, observed, retrieved, now)
	if diagnostic != nil && freshness == portfolio.FreshnessCurrent {
		freshness = portfolio.FreshnessStale
	}
	return portfolio.DatasetContext{
		Source:      source,
		Coverage:    coverage,
		Currency:    currency,
		ObservedAt:  observed,
		RetrievedAt: retrieved,
		PublishedAt: published,
		Freshness:   freshness,
		Diagnostic:  diagnostic,
	}
}

func unavailable(source, coverage string, diagnostic *portfolio.ResourceDiagnostic) portfolio.ShowcaseDataset {
	if diagnostic == nil {
		diagnostic = &portfolio.ResourceDiagnostic{Reason: portfolio.DiagnosticUnknown, RecommendedAction: portfolio.DiagnosticActionRetry}
	}
	return portfolio.ShowcaseDataset{
		Context: portfolio.DatasetContext{
			Source:     source,
			Coverage:   coverage,
			Freshness:  portfolio.FreshnessUnavailable,
			Diagnostic: diagnostic,
		},
	}
}

func (r *ShowcaseRepository) balances(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time, diagnostic *portfolio.ResourceDiagnostic) portfolio.ShowcaseDataset {
	var observed *time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at
		FROM portfolio_balance_heads h
		JOIN portfolio_balance_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id
		WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).
		Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	rows, err := tx.Query(ctx, `SELECT currency,cash::text,buying_power::text FROM portfolio_balance_rows WHERE version_id=$1 ORDER BY row_number`, version)
	if err != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	defer rows.Close()

	dataset := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account", "", observed, &retrieved, &published, mode, false, now, diagnostic)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var balance portfolio.ShowcaseBalance
		if err := rows.Scan(&balance.Currency, &balance.Cash, &balance.BuyingPower); err != nil {
			return unavailable("SnapTrade", "included account", diagnostic)
		}
		currencies[balance.Currency] = struct{}{}
		dataset.Balances = append(dataset.Balances, balance)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	dataset.Context.Currency = currencySummary(currencies)
	return dataset
}

func (r *ShowcaseRepository) positions(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time, diagnostic *portfolio.ResourceDiagnostic) portfolio.ShowcaseDataset {
	var observed time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at
		FROM portfolio_position_heads h
		JOIN portfolio_position_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id
		WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).
		Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	rows, err := tx.Query(ctx, `SELECT symbol,kind,currency,units::text,price::text,cost_basis::text FROM portfolio_position_rows WHERE version_id=$1 ORDER BY row_number`, version)
	if err != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	defer rows.Close()

	dataset := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account", "", &observed, &retrieved, &published, mode, false, now, diagnostic)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var position portfolio.ShowcasePosition
		if err := rows.Scan(&position.Symbol, &position.Kind, &position.Currency, &position.Units, &position.Price, &position.CostBasis); err != nil {
			return unavailable("SnapTrade", "included account", diagnostic)
		}
		currencies[position.Currency] = struct{}{}
		dataset.Positions = append(dataset.Positions, position)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account", diagnostic)
	}
	dataset.Context.Currency = currencySummary(currencies)
	return dataset
}

func (r *ShowcaseRepository) activities(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time, diagnostic *portfolio.ResourceDiagnostic) portfolio.ShowcaseDataset {
	var observed *time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at
		FROM portfolio_activity_heads h
		JOIN portfolio_activity_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id
		WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).
		Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account; newest 50 accumulated activities", diagnostic)
	}
	rows, err := tx.Query(ctx, `SELECT activity_type,trade_date,currency,amount::text,fee::text,price::text,units::text
		FROM portfolio_account_activities WHERE user_id=$1 AND account_id=$2
		ORDER BY trade_date DESC NULLS LAST,activity_id DESC LIMIT 50`, owner, account)
	if err != nil {
		return unavailable("SnapTrade", "included account; newest 50 accumulated activities", diagnostic)
	}
	defer rows.Close()

	dataset := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account; newest 50 accumulated activities", "", observed, &retrieved, &published, mode, true, now, diagnostic)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var activity portfolio.ShowcaseActivity
		if err := rows.Scan(&activity.Type, &activity.TradeDate, &activity.Currency, &activity.Amount, &activity.Fee, &activity.Price, &activity.Units); err != nil {
			return unavailable("SnapTrade", "included account; newest 50 accumulated activities", diagnostic)
		}
		currencies[activity.Currency] = struct{}{}
		dataset.Activities = append(dataset.Activities, activity)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account; newest 50 accumulated activities", diagnostic)
	}
	dataset.Context.Currency = currencySummary(currencies)
	return dataset
}

func currencySummary(values map[string]struct{}) string {
	currencies := make([]string, 0, len(values))
	for currency := range values {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	return strings.Join(currencies, ", ")
}
