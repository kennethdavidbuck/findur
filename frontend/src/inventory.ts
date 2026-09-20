import type { components } from './generated/api'
import { csrfToken } from './session'

export type PortfolioInventory = components['schemas']['PortfolioInventory']

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

async function requestInventory(path: string, init: RequestInit): Promise<PortfolioInventory> {
  const response = await fetch(path, { ...init, cache: 'no-store', credentials: 'same-origin' })
  if (response.status === 401) throw new InventorySessionExpiredError()
  if (response.status === 403) throw new InventorySessionDefenseError()
  if (!response.ok) throw new Error('unavailable')
  const body: unknown = await response.json()
  if (!isInventory(body)) throw new Error('malformed')
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
    && Array.isArray(connection.accounts)
    && connection.accounts.every((account) => isRecord(account)
      && typeof account.id === 'string'
      && knownString(accountCategories, account.category)
      && typeof account.type === 'string'
      && typeof account.maskedLabel === 'string'
      && typeof account.available === 'boolean'
      && typeof account.eligible === 'boolean'
      && knownString(accountSyncStates, account.syncState)))
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function knownString(values: Set<string>, value: unknown): value is string {
  return typeof value === 'string' && values.has(value)
}

const inventoryStates = new Set(['pending', 'ready', 'empty', 'disabled', 'unauthorized', 'rate_limited', 'unavailable', 'malformed'])
const connectionStatuses = new Set(['active', 'disabled', 'unavailable'])
const syncModes = new Set(['realtime', 'delayed', 'unknown'])
const accountCategories = new Set(['investment', 'deposit', 'credit', 'unknown'])
const accountSyncStates = new Set(['complete', 'pending', 'unavailable', 'unknown'])
