import { useEffect, useState, type RefObject } from 'react'
import type { components } from '../generated/api'
import { useI18n } from '../i18n'
import { getPortfolioInclusion, InventorySessionExpiredError } from '../inventory'
import { getPortfolioShowcase, type PortfolioShowcase } from '../showcase'

type Dataset = components['schemas']['ShowcaseDataset']
type Locale = 'en' | 'fr'
const preparationRefreshIntervalMs = 5_000

const copy = {
  en: {
    privateView: 'Private evidence view',
    title: 'Your portfolio',
    introduction: 'A private record of the accounts you chose to include. Values remain in their original currencies; this is not a total, score, or performance view.',
    edit: 'Edit included accounts',
    loading: 'Loading your saved portfolio evidence…',
    failed: 'We can’t show your saved portfolio evidence right now.',
    retry: 'Try again',
    reconnect: 'Reconnect',
    empty: 'No included accounts are saved.',
    preparing: 'Your accounts are saved. Some portfolio data is still syncing. Check back in a few minutes.',
    preparingDataset: 'This account data is still syncing.',
    checkAgain: 'Check again',
    includedAccount: 'Included account',
    included: 'Included',
    coverageHeading: (count: number) => `Using ${count} included account${count === 1 ? '' : 's'}`,
    coverageBody: 'Included account coverage is distinct from provider access. Currencies are never silently combined.',
    daily: 'daily',
    balances: 'Balances',
    positions: 'Positions',
    activities: 'Recent activities',
    source: 'Source',
    coverage: 'Coverage',
    currencies: 'Currencies',
    observed: 'Observed',
    retrieved: 'Retrieved',
    published: 'Published',
    current: 'Current',
    stale: 'Stale, still usable',
    expired: 'Expired',
    unavailable: 'Unavailable',
    emptyDataset: 'This complete dataset has no recorded rows.',
    unavailableDataset: 'This dataset is unavailable. Reconnect or try again later.',
    expiredDataset: 'This dataset has expired, so its saved values are hidden. Reconnect to recover it.',
    recordedRows: 'recorded rows',
    separateCurrencies: 'Currencies stay separate',
    activityCoverage: 'Newest 50 accumulated activities',
    activityCadence: 'Activities are bounded records, often daily; they are not real-time orders.',
    table: 'table',
  },
  fr: {
    privateView: 'Vue privée des données',
    title: 'Votre portefeuille',
    introduction: 'Un relevé privé des comptes que vous avez choisi d’inclure. Les valeurs restent dans leur devise d’origine; ce n’est ni un total, ni une note, ni un rendement.',
    edit: 'Modifier les comptes inclus',
    loading: 'Chargement de vos données enregistrées…',
    failed: 'Nous ne pouvons pas afficher vos données enregistrées pour le moment.',
    retry: 'Réessayer',
    reconnect: 'Reconnecter',
    empty: 'Aucun compte inclus enregistré.',
    preparing: 'Vos comptes sont enregistrés. Certaines données du portefeuille sont encore en cours de synchronisation. Revenez dans quelques minutes.',
    preparingDataset: 'Les données de ce compte sont encore en cours de synchronisation.',
    checkAgain: 'Vérifier à nouveau',
    includedAccount: 'Compte inclus',
    included: 'Inclus',
    coverageHeading: (count: number) => `${count} compte${count === 1 ? '' : 's'} inclus`,
    coverageBody: 'La couverture des comptes inclus est distincte de l’accès du fournisseur. Les devises ne sont jamais combinées.',
    daily: 'quotidien',
    balances: 'Soldes',
    positions: 'Positions',
    activities: 'Activités récentes',
    source: 'Source',
    coverage: 'Couverture',
    currencies: 'Devises',
    observed: 'Observé',
    retrieved: 'Récupéré',
    published: 'Publié',
    current: 'À jour',
    stale: 'Ancien, encore utilisable',
    expired: 'Expiré',
    unavailable: 'Indisponible',
    emptyDataset: 'Cet ensemble complet ne contient aucune ligne enregistrée.',
    unavailableDataset: 'Cet ensemble est indisponible. Reconnectez-vous ou réessayez plus tard.',
    expiredDataset: 'Cet ensemble a expiré; ses valeurs enregistrées sont donc masquées. Reconnectez-vous pour le récupérer.',
    recordedRows: 'lignes enregistrées',
    separateCurrencies: 'Les devises restent distinctes',
    activityCoverage: '50 activités cumulées les plus récentes',
    activityCadence: 'Les activités sont des enregistrements bornés, souvent quotidiens; elles ne sont pas des ordres en temps réel.',
    table: 'tableau',
  },
} as const

const columns = {
  balances: ['currency', 'cash', 'buyingPower'],
  positions: ['symbol', 'kind', 'currency', 'units', 'price', 'costBasis'],
  activities: ['type', 'tradeDate', 'currency', 'amount', 'fee'],
} as const

export function PortfolioShowcasePage({ headingRef, onEdit, onReconnect, onSessionExpired }: { headingRef: RefObject<HTMLHeadingElement | null>; onEdit: () => void; onReconnect: () => void; onSessionExpired: () => void }) {
  const { locale } = useI18n()
  const text = copy[locale]
  const [data, setData] = useState<PortfolioShowcase | null>(null)
  const [failed, setFailed] = useState(false)
  const [preparing, setPreparing] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void getPortfolioShowcase()
      .then(async (showcase) => {
        if (!alive) return
        setData(showcase)
        try {
          const inclusion = await getPortfolioInclusion()
          if (alive) setPreparing(inclusion.change?.status === 'pending')
        } catch (error) {
          if (error instanceof InventorySessionExpiredError) onSessionExpired()
        }
      })
      .catch((error) => {
        if (error instanceof InventorySessionExpiredError) onSessionExpired()
        else if (alive) setFailed(true)
      })
      .finally(() => { if (alive) setRefreshing(false) })
    return () => { alive = false }
  }, [onSessionExpired, reload])

  useEffect(() => {
    if (!preparing) return
    const timer = window.setTimeout(
      () => setReload((value) => value + 1),
      preparationRefreshIntervalMs,
    )
    return () => window.clearTimeout(timer)
  }, [preparing, reload])

  useEffect(() => {
    if (data) requestAnimationFrame(() => headingRef.current?.focus())
  }, [data, headingRef])

  if (failed) {
    return <section className="portfolio-showcase portfolio-showcase--state">
      <p className="showcase-eyebrow">{text.privateView}</p>
      <h1 ref={headingRef} tabIndex={-1}>{text.title}</h1>
      <div className="showcase-recovery" role="alert">
        <p>{text.failed}</p>
        <button className="action action--secondary" onClick={() => { setFailed(false); setData(null); setReload((value) => value + 1) }}>{text.retry}</button>
      </div>
    </section>
  }

  if (!data) {
    return <section className="portfolio-showcase portfolio-showcase--state" aria-busy="true">
      <p className="showcase-eyebrow">{text.privateView}</p>
      <h1 ref={headingRef} tabIndex={-1}>{text.title}</h1>
      <p role="status" className="showcase-loading">{text.loading}</p>
    </section>
  }

  const refreshShowcase = () => {
    setRefreshing(true)
    setReload((value) => value + 1)
  }

  return <div className="showcase-layout">
    <section className="portfolio-showcase showcase-reference">
      <header className="topline">
        <div>
          <p className="eyebrow">{text.privateView} · Portfolio Showcase</p>
          <h1 ref={headingRef} tabIndex={-1}>{text.title}</h1>
          <p>{text.introduction}</p>
        </div>
      </header>
      {data.accounts.length === 0
        ? <div className="showcase-empty" role="status">
          <span aria-hidden="true">◇</span>
          <p>{preparing ? text.preparing : text.empty}</p>
          {preparing
            ? <button className={`text-link${refreshing ? ' text-link--refreshing' : ''}`} disabled={refreshing} onClick={() => { setPreparing(false); setData(null); refreshShowcase() }}>{text.checkAgain} →</button>
            : <button className="text-link" onClick={onEdit}>{text.edit} →</button>}
        </div>
        : <>
          {preparing && <div className="showcase-preparing" role="status">
            <p>{text.preparing}</p>
            <button className={`text-link${refreshing ? ' text-link--refreshing' : ''}`} disabled={refreshing} onClick={refreshShowcase}>{text.checkAgain} →</button>
          </div>}
          <section className="coverage" aria-label={text.coverage}>
            <div>
              <strong>{text.coverageHeading(data.accounts.length)}</strong>
              <span>{text.coverageBody}</span>
            </div>
            <button className="text-link" onClick={onEdit}>{text.edit} →</button>
          </section>
          <section className="account-grid" aria-label={locale === 'fr' ? 'Comptes inclus' : 'Included accounts'}>
            {data.accounts.map((account, index) => <article className="account" key={`${account.brokerage}-${account.label}`}>
              <header>
                <div>
                  <h2>{account.label}</h2>
                  <small>{account.brokerage} · {account.syncMode === 'delayed' ? text.daily : account.syncMode}</small>
                </div>
                <span className="included">◆ {text.included}</span>
              </header>
              <BalanceSummary dataset={account.balances} locale={locale} preparing={preparing} />
              <span className="account-index">{text.includedAccount} {String(index + 1).padStart(2, '0')}</span>
            </article>)}
          </section>
          <section className="showcase-ledgers" aria-label={locale === 'fr' ? 'Données des comptes inclus' : 'Included account evidence'}>
            {data.accounts.map((account, index) => {
              const ledgerID = `showcase-account-${index + 1}`
              return <section className="account-ledger" aria-labelledby={ledgerID} key={`ledger-${account.brokerage}-${account.label}`}>
                <header className="ledger-header">
                  <div>
                    <p className="ledger-kicker">{text.includedAccount} {String(index + 1).padStart(2, '0')}</p>
                    <h2 id={ledgerID}>{account.label}</h2>
                    <p>{account.brokerage} · {account.syncMode === 'delayed' ? text.daily : account.syncMode}</p>
                  </div>
                </header>
                <div className="dataset-list">
                  <Evidence title={text.balances} dataset={account.balances} rows={account.balances.balances} columns={columns.balances} locale={locale} onReconnect={onReconnect} preparing={preparing} />
                  {shouldShowPositions(account.positions) && <Evidence title={text.positions} dataset={account.positions} rows={account.positions.positions} columns={columns.positions} locale={locale} onReconnect={onReconnect} preparing={preparing} />}
                  <Evidence title={text.activities} dataset={account.activities} rows={account.activities.activities} columns={columns.activities} locale={locale} onReconnect={onReconnect} preparing={preparing} activity />
                </div>
              </section>
            })}
          </section>
        </>}
    </section>
  </div>
}

function shouldShowPositions(dataset: Dataset) {
  return dataset.positions.length > 0
}

function BalanceSummary({ dataset, locale, preparing }: { dataset: Dataset; locale: Locale; preparing: boolean }) {
  const text = copy[locale]
  if (dataset.context.freshness === 'expired') return <div className="money"><p>{text.expiredDataset}</p></div>
  if (dataset.context.freshness === 'unavailable') return <div className="money"><p>{preparing ? text.preparingDataset : text.unavailableDataset}</p></div>
  if (dataset.balances.length === 0) return <div className="money"><p>{text.emptyDataset}</p></div>
  return <div className="money">
    {dataset.balances.map((row, index) => <div key={`${row.currency}-${index}`}>
      <b>{formatMoney(row.cash, row.currency, locale)}</b>
      <span>{text.balances.toLowerCase()} · {row.currency}</span>
    </div>)}
    <span className="no-total">△ {text.separateCurrencies}</span>
  </div>
}

function Evidence({ title, dataset, rows, columns, locale, onReconnect, preparing, activity = false }: { title: string; dataset: Dataset; rows: Array<Record<string, unknown>>; columns: readonly string[]; locale: Locale; onReconnect: () => void; preparing: boolean; activity?: boolean }) {
  const text = copy[locale]
  const context = dataset.context
  const recorded = context.observedAt ?? context.retrievedAt
  const recordedLabel = context.observedAt ? text.observed : text.retrieved
  const freshness = context.freshness === 'current' ? text.current : context.freshness === 'stale_usable' ? text.stale : context.freshness === 'expired' ? text.expired : text.unavailable

  return <section className="dataset">
    <header>
      <h3>{title}</h3>
      <span className={`freshness freshness--${context.freshness}`}>
        <i aria-hidden="true">{context.freshness === 'current' ? '◆' : context.freshness === 'stale_usable' ? '◇' : '×'}</i>
        {freshness} · {recorded ? <time dateTime={recorded}>{formatTimestamp(recorded, locale)}</time> : '—'}
      </span>
    </header>
    <dl className="dataset-context">
      <div><dt>{text.source}</dt><dd>{context.source}</dd></div>
      <div><dt>{text.coverage}</dt><dd>{activity ? text.activityCoverage : context.coverage}</dd></div>
      <div><dt>{text.currencies}</dt><dd>{context.currency || '—'}</dd></div>
      <div><dt>{recordedLabel}</dt><dd>{recorded ? <time dateTime={recorded}>{formatTimestamp(recorded, locale)}</time> : '—'}</dd></div>
      <div><dt>{text.published}</dt><dd>{context.publishedAt ? <time dateTime={context.publishedAt}>{formatTimestamp(context.publishedAt, locale)}</time> : '—'}</dd></div>
    </dl>
    {activity && <p className="activity-cadence">△ {text.activityCadence}</p>}
    {context.freshness === 'expired' || context.freshness === 'unavailable'
      ? <div className="showcase-note showcase-note--unavailable">
        <p>{context.freshness === 'expired' ? text.expiredDataset : preparing ? text.preparingDataset : text.unavailableDataset}</p>
        {!preparing && <button className="action action--secondary" onClick={onReconnect}>{text.reconnect}</button>}
      </div>
      : rows.length === 0
        ? <p className="showcase-note">{text.emptyDataset}</p>
        : <>
          <p className="dataset-summary"><strong>{rows.length}</strong> {text.recordedRows} · {text.separateCurrencies}</p>
          <div className="table-wrap" role="region" aria-label={`${title} ${text.table}`} tabIndex={0}>
            <table className="data">
              <caption><strong>{rows.length}</strong> {text.recordedRows} · {text.separateCurrencies}</caption>
              <thead><tr>{columns.map((column) => <th key={column} scope="col">{columnLabel(column, locale)}</th>)}</tr></thead>
              <tbody>{rows.map((row, rowIndex) => <tr key={rowIndex}>{columns.map((column) => <td key={column}>{formatValue(column, row[column], row.currency, locale)}</td>)}</tr>)}</tbody>
            </table>
          </div>
        </>}
  </section>
}

const moneyFields = new Set(['cash', 'buyingPower', 'price', 'costBasis', 'amount', 'fee'])

function formatValue(column: string, value: unknown, currency: unknown, locale: Locale) {
  if (moneyFields.has(column)) return formatMoney(value, currency, locale)
  if (column === 'tradeDate' && typeof value === 'string') return formatTimestamp(value, locale)
  return value == null ? '—' : String(value)
}

function formatMoney(value: unknown, currency: unknown, locale: Locale) {
  if (typeof value !== 'string' || typeof currency !== 'string') return '—'
  const match = /^(-?)(\d+)(\.\d+)?$/.exec(value)
  if (!match) return `${currency} ${value}`
  const language = locale === 'fr' ? 'fr-CA' : 'en-CA'
  const integer = new Intl.NumberFormat(language, { useGrouping: true, maximumFractionDigits: 0 }).format(BigInt(match[2]))
  const decimal = new Intl.NumberFormat(language).formatToParts(1.1).find((part) => part.type === 'decimal')?.value ?? '.'
  const amount = `${match[1]}${integer}${match[3] ? `${decimal}${match[3].slice(1)}` : ''}`
  return locale === 'fr' ? `${amount}\u00a0$ ${currency}` : `${currency} $${amount}`
}

function formatTimestamp(value: string, locale: Locale) {
  return new Date(value).toLocaleString(locale === 'fr' ? 'fr-CA' : 'en-CA')
}

function columnLabel(column: string, locale: Locale) {
  const labels: Record<string, readonly [string, string]> = {
    account: ['Account', 'Compte'], currency: ['Currency', 'Devise'], cash: ['Cash', 'Encaisse'], buyingPower: ['Buying power', 'Pouvoir d’achat'],
    symbol: ['Symbol', 'Symbole'], kind: ['Type', 'Type'], units: ['Units', 'Unités'], price: ['Price', 'Prix'], costBasis: ['Cost basis', 'Coût de base'],
    type: ['Activity', 'Activité'], tradeDate: ['Trade date', 'Date d’opération'], amount: ['Amount', 'Montant'], fee: ['Fee', 'Frais'],
  }
  return labels[column]?.[locale === 'fr' ? 1 : 0] ?? column
}
