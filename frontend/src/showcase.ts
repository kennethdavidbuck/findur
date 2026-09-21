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
    && optionalString(value.observedAt)
    && optionalString(value.retrievedAt)
    && optionalString(value.publishedAt)
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

function knownString(values: Set<string>, value: unknown) {
  return typeof value === 'string' && values.has(value)
}

const syncModes = new Set(['realtime', 'delayed', 'unknown'])
const freshnessStates = new Set(['current', 'stale_usable', 'expired', 'unavailable'])
