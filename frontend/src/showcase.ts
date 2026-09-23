import type { components } from './generated/api'
import { InventorySessionExpiredError } from './inventory'

export type PortfolioShowcase = components['schemas']['PortfolioShowcase']

export async function getPortfolioShowcase(): Promise<PortfolioShowcase> {
  const response = await fetch('/api/portfolio/showcase', { cache: 'no-store', credentials: 'same-origin' })
  if (response.status === 401) throw new InventorySessionExpiredError()
  if (!response.ok) throw new Error('unavailable')
  const body: unknown = await response.json()
  if (!isShowcase(body)) throw new Error('malformed')
  return body
}
function isShowcase(value: unknown): value is PortfolioShowcase {
  return isRecord(value) && Array.isArray(value.accounts) && value.accounts.every((account) => isRecord(account)
    && typeof account.connectionId === 'string'
    && typeof account.label === 'string'
    && typeof account.brokerage === 'string'
    && knownString(syncModes, account.syncMode)
    && isDataset(account.balances)
    && isDataset(account.positions)
    && isDataset(account.activities))
}

function isDataset(value: unknown) {
  return isRecord(value)
    && isContext(value.context)
    && Array.isArray(value.balances) && value.balances.every(isBalance)
    && Array.isArray(value.positions) && value.positions.every(isPosition)
    && Array.isArray(value.activities) && value.activities.every(isActivity)
}

function isContext(value: unknown) {
  return isRecord(value)
    && typeof value.source === 'string'
    && typeof value.coverage === 'string'
    && typeof value.currency === 'string'
    && knownString(freshnessStates, value.freshness)
    && (value.diagnostic === undefined || isDiagnostic(value.diagnostic))
    && optionalString(value.observedAt)
    && optionalString(value.retrievedAt)
    && optionalString(value.publishedAt)
}

function isDiagnostic(value: unknown) {
  return isRecord(value)
    && Object.keys(value).every((key) => diagnosticFields.has(key))
    && knownString(diagnosticReasons, value.reason)
    && knownString(diagnosticActions, value.recommendedAction)
    && diagnosticActionsByReason[value.reason].has(value.recommendedAction)
    && optionalDateTime(value.retryAt)
    && optionalDateTime(value.lastSuccessfulAt)
}

function isBalance(value: unknown) {
  return isRecord(value) && typeof value.currency === 'string' && optionalString(value.cash) && optionalString(value.buyingPower)
}

function isPosition(value: unknown) {
  return isRecord(value)
    && typeof value.symbol === 'string'
    && typeof value.kind === 'string'
    && typeof value.currency === 'string'
    && optionalString(value.units)
    && optionalString(value.price)
    && optionalString(value.costBasis)
}

function isActivity(value: unknown) {
  return isRecord(value)
    && typeof value.type === 'string'
    && typeof value.currency === 'string'
    && optionalString(value.tradeDate)
    && optionalString(value.amount)
    && optionalString(value.fee)
    && optionalString(value.price)
    && optionalString(value.units)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function optionalString(value: unknown) {
  return value === undefined || typeof value === 'string'
}

function optionalDateTime(value: unknown) {
  if (value === undefined) return true
  if (typeof value !== 'string') return false
  const match = rfc3339DateTime.exec(value)
  if (!match || !Number.isFinite(Date.parse(value))) return false
  const year = Number(match[1]); const month = Number(match[2]); const day = Number(match[3])
  return day <= new Date(Date.UTC(year, month, 0)).getUTCDate()
}

function knownString(values: Set<string>, value: unknown): value is string {
  return typeof value === 'string' && values.has(value)
}

const syncModes = new Set(['realtime', 'delayed', 'unknown'])
const freshnessStates = new Set(['current', 'stale_usable', 'expired', 'unavailable'])
const diagnosticReasons = new Set(['no_accounts_returned', 'no_supported_accounts', 'connection_disabled', 'authorization_required', 'provider_unavailable', 'sync_pending', 'unknown'])
const diagnosticActions = new Set(['none', 'wait', 'retry', 'reconnect'])
const diagnosticActionsByReason: Record<string, Set<string>> = {
  no_accounts_returned: new Set(['retry']),
  no_supported_accounts: new Set(['none']),
  connection_disabled: new Set(['reconnect']),
  authorization_required: new Set(['reconnect']),
  provider_unavailable: new Set(['retry', 'wait']),
  sync_pending: new Set(['wait']),
  unknown: new Set(['retry']),
}
const diagnosticFields = new Set(['reason', 'recommendedAction', 'retryAt', 'lastSuccessfulAt'])
const rfc3339DateTime = /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/
