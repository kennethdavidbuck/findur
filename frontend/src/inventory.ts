import type { components } from './generated/api'
import { csrfToken } from './session'

export type PortfolioInventory = components['schemas']['PortfolioInventory']
export type PortfolioInclusion = components['schemas']['PortfolioInclusion']

export class InventorySessionExpiredError extends Error {
  constructor() {
    super('inventory session expired')
    this.name = 'InventorySessionExpiredError'
  }
}

export class InventorySessionDefenseError extends Error {
  constructor() {
    super('inventory session defense failed')
    this.name = 'InventorySessionDefenseError'
  }
}

let initialInventoryRequest: Promise<PortfolioInventory> | null = null

export function resetInitialInventoryRequest() {
  initialInventoryRequest = null
}

export async function getPortfolioInventory(): Promise<PortfolioInventory> {
	if (!initialInventoryRequest) {
		const request = requestInventory('/api/portfolio/inventory', { method: 'GET' })
		initialInventoryRequest = request
		void request.finally(() => {
			if (initialInventoryRequest === request) initialInventoryRequest = null
		}).catch(() => undefined)
	}
	return initialInventoryRequest
}

export async function checkPortfolioInventory(): Promise<PortfolioInventory> {
	return requestInventory('/api/portfolio/inventory', { method: 'GET' })
}

export async function retryPortfolioInventory(): Promise<PortfolioInventory> {
  return requestInventory('/api/portfolio/inventory/retry', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken() },
  })
}

export async function getPortfolioInclusion(): Promise<PortfolioInclusion> {
  return requestInclusion('/api/portfolio/inclusion', { method: 'GET' })
}

export async function confirmPortfolioInclusion(version: number, accountIds: string[]): Promise<PortfolioInclusion> {
  return requestInclusion('/api/portfolio/inclusion', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken(),
      'X-Inclusion-Version': String(version),
      'Idempotency-Key': crypto.randomUUID(),
    },
    body: JSON.stringify({ accountIds }),
  })
}

async function requestInventory(path: string, init: RequestInit): Promise<PortfolioInventory> {
  const response = await fetch(path, { ...init, cache: 'no-store', credentials: 'same-origin' })
  if (response.status === 401) throw new InventorySessionExpiredError()
  if (response.status === 403) throw new InventorySessionDefenseError()
  if (!response.ok) throw new Error('unavailable')
  const body: unknown = await response.json()
  if (!isInventory(body)) throw new Error('malformed')
  return body
}

async function requestInclusion(path: string, init: RequestInit): Promise<PortfolioInclusion> {
  const response = await fetch(path, { ...init, cache: 'no-store', credentials: 'same-origin' })
  if (response.status === 401) throw new InventorySessionExpiredError()
  if (response.status === 403) throw new InventorySessionDefenseError()
  if (response.status === 409) throw new Error('conflict')
  if (!response.ok) throw new Error('unavailable')
  const body: unknown = await response.json()
  if (!isInclusion(body)) throw new Error('malformed')
  return body
}

function isInventory(value: unknown): value is PortfolioInventory {
  if (!isRecord(value) || !knownString(inventoryStates, value.state) || !Number.isSafeInteger(value.generation) || typeof value.updatedAt !== 'string' || !Array.isArray(value.connections)) return false
  if (value.retryAt !== undefined && typeof value.retryAt !== 'string') return false
  return value.connections.every((connection) => isRecord(connection)
    && typeof connection.id === 'string'
    && typeof connection.brokerageLabel === 'string'
    && knownString(connectionStatuses, connection.status)
    && knownString(syncModes, connection.syncMode)
    && typeof connection.available === 'boolean'
    && typeof connection.eligible === 'boolean'
    && (connection.diagnostic === undefined || isDiagnostic(connection.diagnostic))
    && Array.isArray(connection.accounts)
    && connection.accounts.every((account) => isRecord(account)
      && typeof account.id === 'string'
      && knownString(accountCategories, account.category)
      && typeof account.type === 'string'
      && typeof account.maskedLabel === 'string'
      && typeof account.available === 'boolean'
      && typeof account.eligible === 'boolean'
      && typeof account.selectable === 'boolean'
      && knownString(usabilityReasons, account.usabilityReason)
      && knownString(accountSyncStates, account.syncState)))
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

function isInclusion(value: unknown): value is PortfolioInclusion {
  if (!isRecord(value) || !Number.isSafeInteger(value.version) || !Array.isArray(value.committed) || !value.committed.every((id) => typeof id === 'string')) return false
  if (value.change === undefined) return true
  return isRecord(value.change)
    && typeof value.change.id === 'string'
    && knownString(inclusionStatuses, value.change.status)
    && Array.isArray(value.change.additions) && value.change.additions.every((id) => typeof id === 'string')
    && Array.isArray(value.change.removals) && value.change.removals.every((id) => typeof id === 'string')
    && (value.change.failureReason === undefined || knownString(inclusionFailureReasons, value.change.failureReason))
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function knownString(values: Set<string>, value: unknown): value is string {
  return typeof value === 'string' && values.has(value)
}

function optionalDateTime(value: unknown) {
  if (value === undefined) return true
  if (typeof value !== 'string') return false
  const match = rfc3339DateTime.exec(value)
  if (!match || !Number.isFinite(Date.parse(value))) return false
  const year = Number(match[1]); const month = Number(match[2]); const day = Number(match[3])
  return day <= new Date(Date.UTC(year, month, 0)).getUTCDate()
}

const inventoryStates = new Set(['pending', 'ready', 'empty', 'disabled', 'unauthorized', 'rate_limited', 'unavailable', 'malformed'])
const connectionStatuses = new Set(['active', 'disabled', 'unavailable'])
const syncModes = new Set(['realtime', 'delayed', 'unknown'])
const accountCategories = new Set(['investment', 'deposit', 'credit', 'unknown'])
const accountSyncStates = new Set(['complete', 'pending', 'unavailable', 'unknown'])
const usabilityReasons = new Set(['ready', 'provisional_status', 'provisional_category', 'sync_pending', 'connection_disabled', 'connection_unavailable', 'account_closed', 'account_unavailable', 'unsupported_category', 'sync_unavailable'])
const inclusionStatuses = new Set(['pending', 'committed', 'failed'])
const inclusionFailureReasons = new Set(['authorization_required', 'rate_limited', 'provider_unavailable', 'unusable_data', 'stale_guard'])
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
