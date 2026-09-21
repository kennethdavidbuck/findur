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

	rows, err := tx.Query(ctx, `SELECT a.account_id,a.masked_label,c.brokerage_label,c.sync_mode,c.status,c.available,a.available,a.sync_state
 FROM portfolio_included_accounts i JOIN portfolio_inventory_state s ON s.user_id=i.user_id
 JOIN portfolio_inventory_accounts a ON a.user_id=i.user_id AND a.account_id=i.account_id AND a.generation=s.head_generation
 JOIN portfolio_inventory_connections c ON c.user_id=a.user_id AND c.generation=a.generation AND c.connection_id=a.connection_id
 WHERE i.user_id=$1 ORDER BY a.masked_label`, owner)
	if err != nil {
		return portfolio.Showcase{}, err
	}
	type accountHead struct {
		id                                    string
		account                               portfolio.ShowcaseAccount
		connectionStatus, accountSyncState    string
		connectionAvailable, accountAvailable bool
	}
	accounts := make([]accountHead, 0)
	for rows.Next() {
		var head accountHead
		if err := rows.Scan(&head.id, &head.account.Label, &head.account.Brokerage, &head.account.SyncMode, &head.connectionStatus, &head.connectionAvailable, &head.accountAvailable, &head.accountSyncState); err != nil {
			rows.Close()
			return portfolio.Showcase{}, err
		}
		accounts = append(accounts, head)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return portfolio.Showcase{}, err
	}
	rows.Close()

	now := r.clock()
	result := portfolio.Showcase{Accounts: make([]portfolio.ShowcaseAccount, 0, len(accounts))}
	for _, head := range accounts {
		if head.connectionStatus != string(portfolio.ConnectionStatusActive) || !head.connectionAvailable || !head.accountAvailable || head.accountSyncState != string(portfolio.AccountSyncStateComplete) || head.account.SyncMode == portfolio.SyncModeUnknown {
			head.account.Balances = unavailable("SnapTrade", "included account")
			head.account.Positions = unavailable("SnapTrade", "included account")
			head.account.Activities = unavailable("SnapTrade", "included account; last 30 days, up to 500 rows")
		} else {
			head.account.Balances = r.balances(ctx, tx, owner, head.id, head.account.SyncMode, now)
			head.account.Positions = r.positions(ctx, tx, owner, head.id, head.account.SyncMode, now)
			head.account.Activities = r.activities(ctx, tx, owner, head.id, head.account.SyncMode, now)
		}
		result.Accounts = append(result.Accounts, head.account)
	}
	if err := tx.Commit(ctx); err != nil {
		return portfolio.Showcase{}, err
	}
	return result, nil
}
func datasetContext(source, coverage, currency string, observed *time.Time, retrieved, published *time.Time, mode portfolio.SyncMode, activities bool, now time.Time) portfolio.DatasetContext {
	return portfolio.DatasetContext{Source: source, Coverage: coverage, Currency: currency, ObservedAt: observed, RetrievedAt: retrieved, PublishedAt: published, Freshness: portfolio.ClassifyFreshness(mode, activities, observed, retrieved, now)}
}
func unavailable(source, coverage string) portfolio.ShowcaseDataset {
	return portfolio.ShowcaseDataset{Context: portfolio.DatasetContext{Source: source, Coverage: coverage, Freshness: portfolio.FreshnessUnavailable}}
}
func (r *ShowcaseRepository) balances(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time) portfolio.ShowcaseDataset {
	var observed *time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at FROM portfolio_balance_heads h JOIN portfolio_balance_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account")
	}
	rows, err := tx.Query(ctx, `SELECT currency,cash::text,buying_power::text FROM portfolio_balance_rows WHERE version_id=$1 ORDER BY row_number`, version)
	if err != nil {
		return unavailable("SnapTrade", "included account")
	}
	defer rows.Close()
	d := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account", "", observed, &retrieved, &published, mode, false, now)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var x portfolio.ShowcaseBalance
		if err := rows.Scan(&x.Currency, &x.Cash, &x.BuyingPower); err != nil {
			return unavailable("SnapTrade", "included account")
		}
		currencies[x.Currency] = struct{}{}
		d.Balances = append(d.Balances, x)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account")
	}
	d.Context.Currency = currencySummary(currencies)
	return d
}
func (r *ShowcaseRepository) positions(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time) portfolio.ShowcaseDataset {
	var observed time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at FROM portfolio_position_heads h JOIN portfolio_position_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account")
	}
	rows, err := tx.Query(ctx, `SELECT symbol,kind,currency,units::text,price::text,cost_basis::text FROM portfolio_position_rows WHERE version_id=$1 ORDER BY row_number`, version)
	if err != nil {
		return unavailable("SnapTrade", "included account")
	}
	defer rows.Close()
	d := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account", "", &observed, &retrieved, &published, mode, false, now)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var x portfolio.ShowcasePosition
		if err := rows.Scan(&x.Symbol, &x.Kind, &x.Currency, &x.Units, &x.Price, &x.CostBasis); err != nil {
			return unavailable("SnapTrade", "included account")
		}
		currencies[x.Currency] = struct{}{}
		d.Positions = append(d.Positions, x)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account")
	}
	d.Context.Currency = currencySummary(currencies)
	return d
}
func (r *ShowcaseRepository) activities(ctx context.Context, tx pgx.Tx, owner uuid.UUID, account string, mode portfolio.SyncMode, now time.Time) portfolio.ShowcaseDataset {
	var observed *time.Time
	var retrieved, published time.Time
	var version uuid.UUID
	err := tx.QueryRow(ctx, `SELECT v.id,v.observed_at,v.retrieved_at,v.published_at FROM portfolio_activity_heads h JOIN portfolio_activity_versions v ON v.id=h.version_id AND v.user_id=h.user_id AND v.account_id=h.account_id WHERE h.user_id=$1 AND h.account_id=$2`, owner, account).Scan(&version, &observed, &retrieved, &published)
	if err != nil {
		return unavailable("SnapTrade", "included account; last 30 days, up to 500 rows")
	}
	rows, err := tx.Query(ctx, `SELECT activity_type,trade_date,currency,amount::text,fee::text,price::text,units::text FROM portfolio_activity_rows WHERE version_id=$1 ORDER BY row_number LIMIT 500`, version)
	if err != nil {
		return unavailable("SnapTrade", "included account; last 30 days, up to 500 rows")
	}
	defer rows.Close()
	d := portfolio.ShowcaseDataset{Context: datasetContext("SnapTrade", "included account; last 30 days, up to 500 rows", "", observed, &retrieved, &published, mode, true, now)}
	currencies := map[string]struct{}{}
	for rows.Next() {
		var x portfolio.ShowcaseActivity
		if err := rows.Scan(&x.Type, &x.TradeDate, &x.Currency, &x.Amount, &x.Fee, &x.Price, &x.Units); err != nil {
			return unavailable("SnapTrade", "included account; last 30 days, up to 500 rows")
		}
		currencies[x.Currency] = struct{}{}
		d.Activities = append(d.Activities, x)
	}
	if rows.Err() != nil {
		return unavailable("SnapTrade", "included account; last 30 days, up to 500 rows")
	}
	d.Context.Currency = currencySummary(currencies)
	return d
}

func currencySummary(values map[string]struct{}) string {
	currencies := make([]string, 0, len(values))
	for currency := range values {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	return strings.Join(currencies, ", ")
}
