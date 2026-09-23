import { afterEach, describe, expect, it, vi } from 'vitest'
import { InventorySessionExpiredError } from './inventory'
import { getPortfolioShowcase } from './showcase'

afterEach(() => vi.unstubAllGlobals())

describe('getPortfolioShowcase', () => {
	it.each(['no_accounts_returned', 'no_supported_accounts', 'connection_disabled', 'authorization_required', 'provider_unavailable', 'sync_pending', 'unknown'])('accepts the bounded %s diagnostic', async (reason) => {
		const dataset = {
			context: { source: 'SnapTrade', coverage: 'included account', currency: '', freshness: 'unavailable', diagnostic: { reason, recommendedAction: reason === 'authorization_required' || reason === 'connection_disabled' ? 'reconnect' : reason === 'sync_pending' ? 'wait' : reason === 'no_supported_accounts' ? 'none' : 'retry', retryAt: '2026-09-23T12:00:00Z', lastSuccessfulAt: '2026-09-22T12:00:00Z' } },
			balances: [], positions: [], activities: [],
		}
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
			accounts: [{ connectionId: 'connection-1', label: 'Masked account', brokerage: 'Broker', syncMode: 'realtime', balances: dataset, positions: dataset, activities: dataset }],
		}), { status: 200 })))

		await expect(getPortfolioShowcase()).resolves.toBeDefined()
	})

	it('accepts provider-directed waiting without exposing an incompatible action', async () => {
		const dataset = { context: { source: 'SnapTrade', coverage: 'included account', currency: '', freshness: 'unavailable', diagnostic: { reason: 'provider_unavailable', recommendedAction: 'wait' } }, balances: [], positions: [], activities: [] }
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ accounts: [{ connectionId: 'connection-1', label: 'Masked account', brokerage: 'Broker', syncMode: 'realtime', balances: dataset, positions: dataset, activities: dataset }] }), { status: 200 })))

		await expect(getPortfolioShowcase()).resolves.toBeDefined()
	})

	it('rejects unknown reasons and unsafe diagnostic fields', async () => {
		const dataset = {
			context: { source: 'SnapTrade', coverage: 'included account', currency: '', freshness: 'unavailable', diagnostic: { reason: 'raw_provider_failure', recommendedAction: 'retry', rawError: 'secret' } },
			balances: [], positions: [], activities: [],
		}
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
			accounts: [{ connectionId: 'connection-1', label: 'Masked account', brokerage: 'Broker', syncMode: 'realtime', balances: dataset, positions: dataset, activities: dataset }],
		}), { status: 200 })))

		await expect(getPortfolioShowcase()).rejects.toThrow('malformed')
	})

	it.each([
		{ description: 'incompatible reason and action', diagnostic: { reason: 'sync_pending', recommendedAction: 'retry' } },
		{ description: 'non-date retry timestamp', diagnostic: { reason: 'provider_unavailable', recommendedAction: 'retry', retryAt: 'tomorrow' } },
		{ description: 'impossible success timestamp', diagnostic: { reason: 'provider_unavailable', recommendedAction: 'retry', lastSuccessfulAt: '2026-02-30T12:00:00Z' } },
	])('rejects $description diagnostics', async ({ diagnostic }) => {
		const dataset = { context: { source: 'SnapTrade', coverage: 'included account', currency: '', freshness: 'unavailable', diagnostic }, balances: [], positions: [], activities: [] }
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
			accounts: [{ connectionId: 'connection-1', label: 'Masked account', brokerage: 'Broker', syncMode: 'realtime', balances: dataset, positions: dataset, activities: dataset }],
		}), { status: 200 })))

		await expect(getPortfolioShowcase()).rejects.toThrow('malformed')
	})
  it('rejects malformed nested financial data', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      accounts: [{ label: 'Masked account', brokerage: 'Broker', syncMode: 'realtime', balances: null }],
    }), { status: 200 })))

    await expect(getPortfolioShowcase()).rejects.toThrow('malformed')
  })

  it('uses an owner-private no-store request and identifies an expired session', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 401 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(getPortfolioShowcase()).rejects.toBeInstanceOf(InventorySessionExpiredError)
    expect(fetchMock).toHaveBeenCalledWith('/api/portfolio/showcase', { cache: 'no-store', credentials: 'same-origin' })
  })
})
