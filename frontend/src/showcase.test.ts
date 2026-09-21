import { afterEach, describe, expect, it, vi } from 'vitest'
import { InventorySessionExpiredError } from './inventory'
import { getPortfolioShowcase } from './showcase'

afterEach(() => vi.unstubAllGlobals())

describe('getPortfolioShowcase', () => {
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
