import { useEffect, useRef, useState, type RefObject } from 'react'
import type { components } from '../generated/api'
import { useI18n } from '../i18n'
import { getPortfolioInclusion, InventorySessionExpiredError } from '../inventory'
import { getPortfolioShowcase, type PortfolioShowcase } from '../showcase'

type Dataset = components['schemas']['ShowcaseDataset']
type Locale = 'en' | 'fr'
const preparationRefreshIntervalMs = 5_000
const ordinarySyncIntervalMs = 24 * 60 * 60 * 1_000

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
    lastSync: 'Last sync',
    nextSync: 'Next sync',
    balancePending: 'Balance information is still syncing.',
    balanceEmpty: 'No balance information is available.',
    balanceFailure: 'Balance information couldn’t be refreshed.',
    positionsEmpty: 'No positions were reported.',
    activitiesEmpty: 'No recent activity was reported.',
    unavailableDataset: 'This data is unavailable right now.',
    expiredDataset: 'This data has expired, so its saved values are hidden until fresh data is available.',
    diagnosticReasons: {
      no_accounts_returned: 'No accounts were returned.',
      no_supported_accounts: 'The returned accounts are not supported.',
      connection_disabled: 'The connection needs repair.',
      authorization_required: 'Authorization must be renewed for this data.',
      provider_unavailable: 'This data could not be refreshed. Saved values remain visible when safe.',
      sync_pending: 'This data is still syncing.',
      unknown: 'Findur could not determine why this data is unavailable.',
    },
    recordedRows: 'recorded rows',
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
    lastSync: 'Dernière synchronisation',
    nextSync: 'Prochaine synchronisation',
    balancePending: 'Les renseignements sur le solde sont encore en cours de synchronisation.',
    balanceEmpty: 'Aucun renseignement sur le solde n’est disponible.',
    balanceFailure: 'Les renseignements sur le solde n’ont pas pu être actualisés.',
    positionsEmpty: 'Aucune position n’a été déclarée.',
    activitiesEmpty: 'Aucune activité récente n’a été déclarée.',
    unavailableDataset: 'Ces données sont indisponibles pour le moment.',
    expiredDataset: 'Ces données ont expiré; leurs valeurs enregistrées sont masquées jusqu’à ce que de nouvelles données soient disponibles.',
    diagnosticReasons: {
      no_accounts_returned: 'Aucun compte n’a été retourné.',
      no_supported_accounts: 'Les comptes retournés ne sont pas pris en charge.',
      connection_disabled: 'La connexion doit être réparée.',
      authorization_required: 'L’autorisation doit être renouvelée pour ces données.',
      provider_unavailable: 'Ces données n’ont pas pu être actualisées. Les valeurs enregistrées restent visibles lorsqu’elles sont utilisables.',
      sync_pending: 'Ces données sont encore en cours de synchronisation.',
      unknown: 'Findur n’a pas pu déterminer pourquoi ces données sont indisponibles.',
    },
    recordedRows: 'lignes enregistrées',
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
  const focusedHeading = useRef(false)
  const recoveryFocusPending = useRef(false)

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
    if (data && recoveryFocusPending.current) {
      recoveryFocusPending.current = false
      focusedHeading.current = true
      requestAnimationFrame(() => headingRef.current?.focus({ preventScroll: true }))
    } else if (data && !focusedHeading.current) {
      focusedHeading.current = true
      requestAnimationFrame(() => headingRef.current?.focus())
    }
  }, [data, headingRef])

  if (failed) {
    return <section className="portfolio-showcase portfolio-showcase--state">
      <p className="showcase-eyebrow">{text.privateView}</p>
      <h1 ref={headingRef} tabIndex={-1}>{text.title}</h1>
      <div className="showcase-recovery" role="alert">
        <p>{text.failed}</p>
        <button className="action action--secondary" onClick={() => { recoveryFocusPending.current = true; setFailed(false); setData(null); setReload((value) => value + 1) }}>{text.retry}</button>
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
  const diagnostics = data.accounts.flatMap((account) => [account.balances.context.diagnostic, account.positions.context.diagnostic, account.activities.context.diagnostic]).filter((diagnostic) => diagnostic !== undefined)
  const reconnectRequired = diagnostics.some((diagnostic) => diagnostic.recommendedAction === 'reconnect')
  const initialSyncPending = diagnostics.some((diagnostic) => diagnostic.reason === 'sync_pending')

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
          {!preparing && <button className="text-link" onClick={onEdit}>{text.edit} →</button>}
        </div>
        : <>
          {initialSyncPending && <div className="showcase-preparing" role="status">
            <p>{text.preparing}</p>
            <button className={`text-link${refreshing ? ' text-link--refreshing' : ''}`} disabled={refreshing} onClick={refreshShowcase}>{text.checkAgain} →</button>
          </div>}
          {reconnectRequired && <div className="showcase-recovery" aria-label={locale === 'fr' ? 'Actions de récupération des données' : 'Data recovery actions'}>
            <button className="action action--secondary" onClick={onReconnect}>{text.reconnect}</button>
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
              <BalanceSummary dataset={account.balances} locale={locale} />
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
                  <Evidence title={text.balances} dataset={account.balances} rows={account.balances.balances} columns={columns.balances} locale={locale} emptyMessage={text.balanceEmpty} />
                  <Evidence title={text.positions} dataset={account.positions} rows={account.positions.positions} columns={columns.positions} locale={locale} emptyMessage={text.positionsEmpty} />
                  <Evidence title={text.activities} dataset={account.activities} rows={account.activities.activities} columns={columns.activities} locale={locale} emptyMessage={text.activitiesEmpty} activity />
                </div>
              </section>
            })}
          </section>
        </>}
    </section>
  </div>
}

function BalanceSummary({ dataset, locale }: { dataset: Dataset; locale: Locale }) {
  const text = copy[locale]
  if (dataset.context.freshness === 'expired') return <div className="money"><p className="balance-state">{text.expiredDataset}</p></div>
  if (dataset.balances.length === 0 && dataset.context.diagnostic?.reason === 'sync_pending') return <div className="money"><p className="balance-state balance-state--info">{text.balancePending}</p></div>
  if (dataset.balances.length === 0 && dataset.context.diagnostic) return <div className="money"><p className={`balance-state balance-state--${diagnosticTone(dataset.context.diagnostic.reason)}`}>{text.balanceFailure}</p></div>
  if (dataset.context.freshness === 'unavailable') return <div className="money"><p className="balance-state">{text.balanceFailure}</p></div>
  if (dataset.balances.length === 0) return <div className="money"><p className="balance-state">{text.balanceEmpty}</p></div>
  return <div className="money">
    {dataset.balances.map((row, index) => <div key={`${row.currency}-${index}`}>
      <b>{formatMoney(row.cash, row.currency, locale)}</b>
      <span>{text.balances.toLowerCase()} · {row.currency}</span>
    </div>)}
  </div>
}

function Evidence({ title, dataset, rows, columns, locale, emptyMessage, activity = false }: { title: string; dataset: Dataset; rows: Array<Record<string, unknown>>; columns: readonly string[]; locale: Locale; emptyMessage: string; activity?: boolean }) {
  const text = copy[locale]
  const context = dataset.context
  const nextSync = nextSyncAt(context)

  return <section className="dataset">
    <header>
      <h3>{title}</h3>
    </header>
    <dl className="dataset-context">
      <div><dt>{text.source}</dt><dd>{context.source}</dd></div>
      <div><dt>{text.coverage}</dt><dd>{activity ? text.activityCoverage : context.coverage}</dd></div>
      <div><dt>{text.currencies}</dt><dd>{context.currency || '—'}</dd></div>
      <div><dt>{text.lastSync}</dt><dd>{context.publishedAt ? <time dateTime={context.publishedAt}>{formatTimestamp(context.publishedAt, locale)}</time> : '—'}</dd></div>
      <div><dt>{text.nextSync}</dt><dd>{nextSync ? <time dateTime={nextSync}>{formatTimestamp(nextSync, locale)}</time> : '—'}</dd></div>
    </dl>
    {activity && <p className="activity-cadence">△ {text.activityCadence}</p>}
    {context.diagnostic && context.diagnostic.reason !== 'sync_pending' && <div className={`showcase-note showcase-note--diagnostic showcase-note--${diagnosticTone(context.diagnostic.reason)}`}>
      <p>{text.diagnosticReasons[context.diagnostic.reason]}</p>
    </div>}
    {context.freshness === 'expired' || context.freshness === 'unavailable' && !context.diagnostic
      ? <div className="showcase-note">
        <p>{context.freshness === 'expired' ? text.expiredDataset : text.unavailableDataset}</p>
      </div>
      : rows.length === 0
        ? context.diagnostic ? null : <p className="showcase-note">{emptyMessage}</p>
        : <>
          <p className="dataset-summary"><strong>{rows.length}</strong> {text.recordedRows}</p>
          <div className="table-wrap" role="region" aria-label={`${title} ${text.table}`} tabIndex={0}>
            <table className="data">
              <caption><strong>{rows.length}</strong> {text.recordedRows}</caption>
              <thead><tr>{columns.map((column) => <th key={column} scope="col">{columnLabel(column, locale)}</th>)}</tr></thead>
              <tbody>{rows.map((row, rowIndex) => <tr key={rowIndex}>{columns.map((column) => <td key={column}>{formatValue(column, row[column], row.currency, locale)}</td>)}</tr>)}</tbody>
            </table>
          </div>
        </>}
  </section>
}

function nextSyncAt(context: Dataset['context']) {
  if (context.diagnostic) return context.diagnostic.retryAt
  if (!context.publishedAt) return undefined
  const publishedAt = Date.parse(context.publishedAt)
  return Number.isNaN(publishedAt) ? undefined : new Date(publishedAt + ordinarySyncIntervalMs).toISOString()
}

function diagnosticTone(reason: components['schemas']['ResourceDiagnostic']['reason']) {
  if (reason === 'sync_pending') return 'info'
  if (reason === 'provider_unavailable') return 'warning'
  if (reason === 'authorization_required' || reason === 'connection_disabled' || reason === 'unknown') return 'error'
  return 'neutral'
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
