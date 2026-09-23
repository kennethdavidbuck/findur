import { afterEach, describe, expect, it, vi } from 'vitest'
import { getPortfolioInventory, resetInitialInventoryRequest } from './inventory'

afterEach(() => {
  resetInitialInventoryRequest()
  vi.unstubAllGlobals()
})

describe('getPortfolioInventory diagnostics', () => {
  it.each(['no_accounts_returned', 'no_supported_accounts', 'connection_disabled', 'authorization_required', 'provider_unavailable', 'sync_pending', 'unknown'])('accepts the bounded %s reason', async (reason) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({
      state: 'ready', generation: 1, updatedAt: '2026-09-23T12:00:00Z',
      connections: [{
        id: 'connection', brokerageLabel: 'Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true, accounts: [],
        diagnostic: { reason, recommendedAction: reason === 'connection_disabled' || reason === 'authorization_required' ? 'reconnect' : reason === 'sync_pending' ? 'wait' : reason === 'no_supported_accounts' ? 'none' : 'retry' },
      }],
    })))

    await expect(getPortfolioInventory()).resolves.toBeDefined()
  })

  it('accepts provider-directed waiting', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({
      state: 'ready', generation: 1, updatedAt: '2026-09-23T12:00:00Z',
      connections: [{ id: 'connection', brokerageLabel: 'Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true, accounts: [], diagnostic: { reason: 'provider_unavailable', recommendedAction: 'wait' } }],
    })))

    await expect(getPortfolioInventory()).resolves.toBeDefined()
  })

  it('rejects unknown and unsafe diagnostic data', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({
      state: 'ready', generation: 1, updatedAt: '2026-09-23T12:00:00Z',
      connections: [{
        id: 'connection', brokerageLabel: 'Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true, accounts: [],
        diagnostic: { reason: 'provider_unavailable', recommendedAction: 'retry', rawProviderError: 'secret' },
      }],
    })))

    await expect(getPortfolioInventory()).rejects.toThrow('malformed')
  })

  it.each([
    { reason: 'authorization_required', recommendedAction: 'retry' },
    { reason: 'provider_unavailable', recommendedAction: 'retry', retryAt: 'soon' },
    { reason: 'provider_unavailable', recommendedAction: 'retry', lastSuccessfulAt: '2026-02-30T12:00:00Z' },
  ])('rejects incompatible or malformed diagnostic $reason/$recommendedAction', async (diagnostic) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({
      state: 'ready', generation: 1, updatedAt: '2026-09-23T12:00:00Z',
      connections: [{
        id: 'connection', brokerageLabel: 'Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true, accounts: [], diagnostic,
      }],
    })))

    await expect(getPortfolioInventory()).rejects.toThrow('malformed')
  })
})

function response(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
}
