import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { App } from './App'

type MediaChangeListener = (event: MediaQueryListEvent) => void

const emptyInventory = () => new Response(JSON.stringify({
  state: 'empty', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [],
}), { status: 200, headers: { 'Content-Type': 'application/json' } })

const jsonResponseBody = (body: unknown) => new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })

function inclusionAccount(id: string, maskedLabel: string, selectable: boolean, usabilityReason: string, category = 'investment') {
	return { id, category, type: category === 'investment' ? 'Margin' : 'Checking', maskedLabel, available: true, eligible: selectable, selectable, usabilityReason, syncState: 'complete' }
}

function inclusionInventory(accounts: ReturnType<typeof inclusionAccount>[]) {
	return {
		state: 'ready', generation: 4, updatedAt: '2026-09-20T12:00:00Z', connections: [{
			id: 'connection-1', brokerageLabel: 'Synthetic Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true, accounts,
		}],
	}
}

function installInclusionFetch(inventory: ReturnType<typeof inclusionInventory>, inclusion: unknown) {
	const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
		if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
		if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
		if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(jsonResponseBody(inclusion))
		return Promise.resolve(new Response(null, { status: 404 }))
	})
	vi.stubGlobal('fetch', fetchMock)
	return fetchMock
}

function installMatchMedia(initialDark = false) {
  let matches = initialDark
  const listeners = new Set<MediaChangeListener>()
  const mediaQuery = {
    get matches() {
      return matches
    },
    media: '(prefers-color-scheme: dark)',
    onchange: null,
    addEventListener: (_type: string, listener: MediaChangeListener) => listeners.add(listener),
    removeEventListener: (_type: string, listener: MediaChangeListener) => listeners.delete(listener),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  } as unknown as MediaQueryList

  vi.stubGlobal('matchMedia', vi.fn(() => mediaQuery))

  return {
    setDark(nextMatches: boolean) {
      matches = nextMatches
      const event = { matches, media: mediaQuery.media } as MediaQueryListEvent
      listeners.forEach((listener) => listener(event))
    },
  }
}

beforeEach(() => {
  window.history.replaceState(null, '', '/')
  window.localStorage.clear()
  document.documentElement.lang = 'en'
  document.documentElement.removeAttribute('data-theme')
  document.documentElement.removeAttribute('data-theme-preference')
  document.head.innerHTML = '<meta name="description" content=""><meta name="theme-color" content="">'
  installMatchMedia(false)
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

describe('public site', () => {
  it('renders the static landing page without requesting the backend', () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Find a different pattern in the same sky.',
    )
    expect(screen.getByText('18+ evaluation demo')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Owner access' })).toHaveAttribute('href', '/connect')
    expect(screen.getByRole('link', { name: 'Owner access' })).toHaveAccessibleDescription('Review the secure connection boundary.')
    expect(screen.getByText('Review the secure connection boundary.')).toBeVisible()
    expect(screen.getAllByRole('navigation')).toHaveLength(1)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('navigates to About, updates metadata, and focuses the route heading', async () => {
    render(<App />)

    fireEvent.click(screen.getByRole('link', { name: /explore the idea/i }))

    const heading = screen.getByRole('heading', { level: 1, name: /compatibility can begin/i })
    await waitFor(() => expect(heading).toHaveFocus())
    expect(window.location.pathname).toBe('/about')
    expect(document.title).toBe('About Findur')
    expect(document.querySelector('meta[name="description"]')).toHaveAttribute(
      'content',
      expect.stringContaining('deliberate disclosure'),
    )
    expect(screen.getAllByRole('link', { name: 'About' })[0]).toHaveAttribute('aria-current', 'page')
  })

  it('switches all public copy to French and restores the stored locale', async () => {
    const { unmount } = render(<App />)

    fireEvent.click(screen.getAllByRole('radio', { name: 'FR' })[0])

    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent(
      'Trouvez une autre trajectoire dans le même ciel.',
    )
    expect(document.documentElement.lang).toBe('fr')
    expect(window.localStorage.getItem('findur-locale')).toBe('fr')
    expect(screen.getByRole('link', { name: 'Aller au contenu' })).toBeInTheDocument()
    expect(document.title).toBe('Findur — Des rencontres axées sur le portefeuille')

    fireEvent.click(screen.getAllByRole('link', { name: 'À propos' })[0])
    expect(window.location.pathname).toBe('/about')
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'La compatibilité peut commencer par nos choix.',
    )
    expect(screen.getAllByRole('navigation')).toHaveLength(1)

    unmount()
    render(<App />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'La compatibilité peut commencer par nos choix.',
    )
  })

  it('operates the language selector from the keyboard', async () => {
    render(<App />)
    const english = screen.getByRole('radio', { name: 'EN' })

    english.focus()
    fireEvent.keyDown(english, { key: 'ArrowRight' })

    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent(
      'Trouvez une autre trajectoire dans le même ciel.',
    )
    expect(screen.getByRole('radio', { name: 'FR' })).toBeChecked()
  })

  it('persists an explicit dark theme and overrides the system preference', async () => {
    installMatchMedia(false)
    render(<App />)

    fireEvent.click(screen.getAllByRole('radio', { name: 'Dark' })[0])

    await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'dark'))
    expect(document.documentElement).toHaveAttribute('data-theme-preference', 'dark')
    expect(window.localStorage.getItem('findur-theme')).toBe('dark')
    expect(document.querySelector('meta[name="theme-color"]')).toHaveAttribute('content', '#0B1020')
  })

  it('tracks operating-system changes while System theme is selected', async () => {
    const media = installMatchMedia(false)
    render(<App />)

    expect(document.documentElement).toHaveAttribute('data-theme', 'light')
    media.setDark(true)

    await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'dark'))
    expect(document.documentElement).toHaveAttribute('data-theme-preference', 'system')
  })

  it('restores a stored explicit light theme even when the system is dark', async () => {
    window.localStorage.setItem('findur-theme', 'light')
    installMatchMedia(true)

    render(<App />)

    await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'light'))
    expect(document.documentElement).toHaveAttribute('data-theme-preference', 'light')
  })

  it('renders the About route directly and responds to browser navigation', async () => {
    window.history.replaceState(null, '', '/about')
    render(<App />)

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Compatibility can begin with how we choose.',
    )

    window.history.pushState(null, '', '/')
    window.dispatchEvent(new PopStateEvent('popstate'))

    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
        'Find a different pattern in the same sky.',
      ),
    )
  })

  it('enables staged consent only after the server reports authorization available', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    vi.stubGlobal('fetch', fetchMock)
    render(<App />)

    fireEvent.click(screen.getByRole('link', { name: 'Owner access' }))

    const heading = screen.getByRole('heading', { level: 1, name: /without sharing your brokerage password/i })
    await waitFor(() => expect(heading).toHaveFocus())
    expect(window.location.pathname).toBe('/connect')
    expect(screen.getByText('No account is included by default.')).toBeVisible()
    expect(screen.getByText(/Balances, positions, activities, signal derivation/)).toBeVisible()
    expect(screen.getByText(/Findur never sees or stores/)).toBeVisible()
    expect(screen.getByText(/cannot trade and does not provide financial advice/)).toBeVisible()
    expect(screen.getByText(/deleting your data from app-controlled active storage/)).toBeVisible()
	await waitFor(() => expect(screen.getByRole('button', { name: 'Continue to SnapTrade' })).toBeEnabled())
	expect(fetchMock).toHaveBeenCalledWith('/api/auth/status', expect.objectContaining({ cache: 'no-store', credentials: 'same-origin' }))
  })

	it('rejects a forged success URL and trusts only server session status', async () => {
		const fetchMock = vi.fn().mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 }))
			.mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })))
		vi.stubGlobal('fetch', fetchMock)
		window.history.replaceState(null, '', '/connect/result?status=success')
		const { unmount } = render(<App />)
		expect(screen.getByRole('status')).toHaveTextContent('Checking your secure session…')
		expect(window.location.pathname).toBe('/connect/result')
		expect(window.location.search).toBe('')
		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByText('Choose accounts before anything else.')).not.toBeInTheDocument()
		unmount()
		window.history.replaceState(null, '', '/connect/result')
		render(<App />)
		const portfolioHeading = await screen.findByRole('heading', { level: 1, name: 'Your masked account inventory' })
		await waitFor(() => expect(portfolioHeading).toHaveFocus())
		expect(window.location.pathname).toBe('/portfolio')
	})

	it('gates private content before mounting and exposes only the three private destinations', async () => {
		let resolveStatus!: (response: Response) => void
		vi.stubGlobal('fetch', vi.fn().mockReturnValue(new Promise<Response>((resolve) => { resolveStatus = resolve })))
		window.history.replaceState(null, '', '/portfolio')
		render(<App />)
		expect(screen.getByRole('status')).toHaveTextContent('Checking your secure session…')
		expect(screen.queryByText('Choose accounts before anything else.')).not.toBeInTheDocument()
		resolveStatus(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
		const heading = await screen.findByRole('heading', { level: 1, name: 'Your masked account inventory' })
		await waitFor(() => expect(heading).toHaveFocus())
		for (const name of ['Discovery', 'Portfolio', 'Profile']) expect(screen.getAllByRole('link', { name })).toHaveLength(1)
		expect(screen.getAllByRole('navigation')).toHaveLength(1)
		fireEvent.click(screen.getByRole('link', { name: 'Profile' }))
		const profile = screen.getByRole('heading', { level: 1, name: 'Profile setup comes later.' })
		await waitFor(() => expect(profile).toHaveFocus())
		expect(window.location.pathname).toBe('/profile')
		window.history.back()
		window.dispatchEvent(new PopStateEvent('popstate'))
		await waitFor(() => expect(screen.getByRole('heading', { level: 1, name: 'Your masked account inventory' })).toHaveFocus())
	})

	it('renders connection state before minimized accounts without render-driven repeats', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const inventory = {
			state: 'ready', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [{
				id: 'connection-1', brokerageLabel: 'Synthetic Broker', status: 'active', syncMode: 'delayed', available: true, eligible: true,
				accounts: [
					{ id: 'account-1', category: 'investment', type: 'Margin', maskedLabel: 'Retirement (•••• 8443)', available: true, eligible: true, selectable: true, usabilityReason: 'ready', syncState: 'complete' },
					{ id: 'account-2', category: 'unknown', type: 'unknown', maskedLabel: 'Everyday account', available: true, eligible: false, selectable: true, usabilityReason: 'provisional_category', syncState: 'complete' },
				],
			}],
		}
		const fetchMock = vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/auth/status'
			? new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })
			: path === '/api/portfolio/inclusion'
				? new Response(JSON.stringify({ version: 0, committed: [] }), { status: 200 })
				: new Response(JSON.stringify(inventory), { status: 200 })))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		expect((await screen.findAllByText('Retirement (•••• 8443)'))[0]).toBeVisible()
		expect(screen.getByText(/No account is included by default/)).toBeVisible()
		const status = screen.getByRole('heading', { name: 'Connection status' })
		const account = screen.getAllByText('Retirement (•••• 8443)')[0]
		expect(status.compareDocumentPosition(account) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
		expect(document.body).not.toHaveTextContent('Q6542138443')
		expect(document.body).not.toHaveTextContent('$')
		expect(screen.getAllByText(/Availability:/).length).toBeGreaterThan(0)
		expect(screen.getByText('Later-inclusion eligibility: Eligible')).toBeVisible()
		expect(screen.queryByText('Later-inclusion eligibility: Not eligible')).not.toBeInTheDocument()
		expect(screen.getAllByText('Inclusion readiness: Selectable for confirmation')).toHaveLength(2)
		expect(screen.getAllByText('Selectable provisionally; the provider must confirm investment data.')[0]).toBeVisible()
		expect(screen.getByText('Connection sync mode: Delayed')).toBeVisible()
		expect(screen.getAllByText('Account sync state: Complete')).toHaveLength(2)
		expect(screen.getByText('Category: Category unavailable')).toBeVisible()
		expect(screen.getAllByText('Everyday account')[0]).toBeVisible()

		fireEvent.click(screen.getAllByRole('radio', { name: 'FR' })[0])
		const frenchHeading = await screen.findByRole('heading', { level: 1, name: 'Votre inventaire de comptes masqués' })
		await waitFor(() => expect(frenchHeading).toHaveFocus())
		expect(screen.getByText(/Couverture confirmée: 0 sur 2 comptes connectés/)).toBeVisible()
		fireEvent.click(screen.getAllByRole('radio', { name: 'Sombre' })[0])
		await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'dark'))
		expect(fetchMock).toHaveBeenCalledTimes(3)
	})

	it('keeps inclusion draft separate, explains disabled rows, and confirms through the defended dialog', async () => {
		window.history.replaceState(null, '', '/portfolio')
		document.cookie = 'findur_csrf=inclusion-csrf; Path=/'
		const inventory = {
			state: 'ready', generation: 4, updatedAt: '2026-09-20T12:00:00Z', connections: [{
				id: 'connection-1', brokerageLabel: 'Synthetic Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true,
				accounts: [
					{ id: 'account-1', category: 'investment', type: 'Margin', maskedLabel: 'Retirement (•••• 8443)', available: true, eligible: true, selectable: true, usabilityReason: 'ready', syncState: 'complete' },
					{ id: 'account-2', category: 'deposit', type: 'Checking', maskedLabel: 'Daily cash (•••• 1000)', available: true, eligible: false, selectable: false, usabilityReason: 'unsupported_category', syncState: 'complete' },
				],
			}],
		}
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(new Response(JSON.stringify(inventory), { status: 200 }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(new Response(JSON.stringify({ version: 0, committed: [] }), { status: 200 }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(new Response(JSON.stringify({ version: 1, committed: ['account-1'], change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'committed', additions: ['account-1'], removals: [] } }), { status: 200 }))
			return Promise.resolve(new Response(null, { status: 404 }))
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		const retirement = await screen.findByRole('checkbox', { name: /Retirement/ })
		const disabled = screen.getByRole('checkbox', { name: /Daily cash/ })
		expect(retirement).not.toBeChecked()
		expect(disabled).toBeDisabled()
		expect(screen.getAllByText('Unavailable because Findur supports investment accounts only.').length).toBeGreaterThan(0)
		expect(screen.getByText('Inclusion readiness: Unavailable for inclusion')).toBeVisible()
		expect(screen.getByText(/Committed coverage: 0 of 2 connected accounts/)).toBeVisible()
		fireEvent.click(retirement)
		expect(screen.getByText(/Draft coverage: 1 of 2 connected accounts/)).toBeVisible()
		const review = screen.getByRole('button', { name: 'Review and confirm' })
		fireEvent.click(review)
		expect(screen.getByRole('dialog', { name: 'Confirm included accounts' })).toBeVisible()
		expect(screen.getByText(/separate balances, positions, and recent-activities datasets/)).toBeVisible()
		fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
		await waitFor(() => expect(review).toHaveFocus())
		expect(retirement).toBeChecked()

		fireEvent.click(review)
		fireEvent.click(screen.getByRole('button', { name: 'Confirm selection' }))
		await waitFor(() => expect(screen.getByText(/Currently included: Retirement/)).toBeVisible())
		const post = fetchMock.mock.calls.find(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(post?.[1]).toEqual(expect.objectContaining({
			cache: 'no-store', credentials: 'same-origin', body: JSON.stringify({ accountIds: ['account-1'] }),
			headers: expect.objectContaining({ 'X-CSRF-Token': 'inclusion-csrf', 'X-Inclusion-Version': '0' }),
		}))
	})

	it('applies tri-state Select All only to usable rows and preserves disabled committed coverage', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const inventory = inclusionInventory([
			inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready'),
			inclusionAccount('account-2', 'Growth (•••• 2002)', true, 'ready'),
			inclusionAccount('account-3', 'Legacy cash (•••• 3003)', false, 'unsupported_category', 'deposit'),
		])
		installInclusionFetch(inventory, { version: 4, committed: ['account-3'] })
		render(<App />)

		const selectAll = await screen.findByRole('checkbox', { name: 'Select all usable accounts' })
		const first = screen.getByRole('checkbox', { name: /Retirement/ })
		const second = screen.getByRole('checkbox', { name: /Growth/ })
		const disabledCommitted = screen.getByRole('checkbox', { name: /Legacy cash/ })
		expect(selectAll).not.toBeChecked()
		expect(disabledCommitted).toBeChecked()

		fireEvent.click(selectAll)
		expect(selectAll).toBeChecked()
		expect(first).toBeChecked()
		expect(second).toBeChecked()
		expect(disabledCommitted).toBeChecked()

		fireEvent.click(first)
		expect(selectAll).toBePartiallyChecked()
		fireEvent.click(selectAll)
		expect(selectAll).toBeChecked()
		fireEvent.click(selectAll)
		expect(selectAll).not.toBeChecked()
		expect(first).not.toBeChecked()
		expect(second).not.toBeChecked()
		expect(disabledCommitted).toBeChecked()
	})

	it('reconstructs failed additions for a scoped retry while keeping removals effective', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const inventory = inclusionInventory([
			inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready'),
			inclusionAccount('account-2', 'Removed (•••• 1000)', true, 'ready'),
		])
		installInclusionFetch(inventory, {
			version: 3,
			committed: [],
			change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'failed', additions: ['account-1'], removals: ['account-2'], failureReason: 'provider_unavailable' },
		})
		render(<App />)

		expect(await screen.findByText(/Draft selection: Retirement/)).toBeVisible()
		expect(screen.getByText(/Currently included: None/)).toBeVisible()
		expect(screen.getByRole('checkbox', { name: /Removed/ })).not.toBeChecked()
		const retry = screen.getByRole('button', { name: 'Review scoped retry' })
		expect(retry).toBeEnabled()
		fireEvent.click(retry)
		const dialog = screen.getByRole('dialog', { name: 'Confirm included accounts' })
		expect(dialog).toHaveTextContent('SnapTrade authorization permits provider access')
		expect(dialog).toHaveTextContent('Category: Investment')
		expect(dialog).toHaveTextContent('Current included-account coverage: 0 of 2 connected accounts')
		expect(dialog).toHaveTextContent('Coverage after this confirmation: 1 of 2 connected accounts')
		expect(dialog).toHaveTextContent('Private derivation and matching')
	})

	it('reloads or safely supersedes a durable pending addition with the same scoped draft', async () => {
		window.history.replaceState(null, '', '/portfolio')
		document.cookie = 'findur_csrf=pending-csrf; Path=/'
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		const pending = {
			version: 1,
			committed: [],
			change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'pending', additions: ['account-1'], removals: [] },
		}
		let inclusionGets = 0
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') {
				inclusionGets += 1
				return Promise.resolve(jsonResponseBody(pending))
			}
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(jsonResponseBody({ version: 2, committed: ['account-1'], change: { ...pending.change, status: 'committed' } }))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		expect(await screen.findByText(/Draft selection: Retirement/)).toBeVisible()
		fireEvent.click(screen.getByRole('button', { name: 'Reload inclusion status' }))
		await waitFor(() => expect(inclusionGets).toBe(2))
		const retry = screen.getByRole('button', { name: 'Review scoped retry' })
		fireEvent.click(retry)
		const cancel = await screen.findByRole('button', { name: 'Cancel' })
		await waitFor(() => expect(cancel).toHaveFocus())
		fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' })
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
		await waitFor(() => expect(retry).toHaveFocus())

		fireEvent.click(retry)
		fireEvent.click(screen.getByRole('button', { name: 'Confirm selection' }))
		await waitFor(() => expect(screen.getByText(/Currently included: Retirement/)).toBeVisible())
		const post = fetchMock.mock.calls.find(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(post?.[1]).toEqual(expect.objectContaining({
			body: JSON.stringify({ accountIds: ['account-1'] }),
			headers: expect.objectContaining({ 'X-Inclusion-Version': '1', 'X-CSRF-Token': 'pending-csrf' }),
		}))
	})

	it('uses a labelled danger action and complete purge warning for removals', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		installInclusionFetch(inventory, { version: 1, committed: ['account-1'] })
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Retirement/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review and confirm' }))
		const danger = screen.getByRole('button', { name: 'Remove accounts and confirm selection' })
		expect(danger).toHaveClass('action--danger')
		const consequence = screen.getByText(/Excluding an account is immediate and irreversible in Findur/)
		expect(consequence).toHaveTextContent('source and normalized financial data')
		expect(consequence).toHaveTextContent('derived outputs, signals, matching and ranking inputs or results')
		expect(consequence).toHaveTextContent('caches and rendered or prefetched state')
		expect(consequence).toHaveTextContent('disclosure-controlled previews')
		expect(consequence).toHaveTextContent('Discovery outputs')
		expect(consequence).toHaveTextContent('Failed additions do not restore any removed data')
		expect(consequence).not.toHaveTextContent(/revok/i)
		expect(screen.getByText(/Destructive removal:/)).toBeVisible()
	})

	it('shares one in-flight initial inventory request across a development remount', async () => {
		window.history.replaceState(null, '', '/portfolio')
		let resolveInventory!: (response: Response) => void
		const inventoryResponse = new Promise<Response>((resolve) => { resolveInventory = resolve })
		const fetchMock = vi.fn().mockImplementation((path: string) => path === '/api/auth/status'
			? Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			: inventoryResponse)
		vi.stubGlobal('fetch', fetchMock)
		const first = render(<App />)
		await waitFor(() => expect(fetchMock.mock.calls.filter(([path]) => path === '/api/portfolio/inventory')).toHaveLength(1))
		first.unmount()
		render(<App />)
		await waitFor(() => expect(fetchMock.mock.calls.filter(([path]) => path === '/api/auth/status')).toHaveLength(2))
		resolveInventory(emptyInventory())
		expect(await screen.findByText('No brokerage connections are available.')).toBeVisible()
		expect(fetchMock.mock.calls.filter(([path]) => path === '/api/portfolio/inventory')).toHaveLength(1)
	})

	it('does not share an old in-flight inventory request with a session after logout', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const oldInventoryRequest = new Promise<Response>(() => undefined)
		let inventoryCalls = 0
		const fetchMock = vi.fn().mockImplementation((path: string) => {
			if (path === '/api/auth/status') return Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			if (path === '/api/auth/logout') return Promise.resolve(new Response(null, { status: 204 }))
			if (path === '/api/portfolio/inventory') {
				inventoryCalls += 1
				return inventoryCalls === 1 ? oldInventoryRequest : Promise.resolve(emptyInventory())
			}
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		const first = render(<App />)
		await waitFor(() => expect(inventoryCalls).toBe(1))
		fireEvent.click(screen.getByRole('button', { name: 'Log out' }))
		await waitFor(() => expect(window.location.pathname).toBe('/'))
		first.unmount()
		window.history.replaceState(null, '', '/portfolio')
		render(<App />)
		expect(await screen.findByText('No brokerage connections are available.')).toBeVisible()
		expect(inventoryCalls).toBe(2)
	})

	it('lets a pending observer explicitly check persisted status without polling', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const pending = { state: 'pending', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		const ready = { state: 'empty', generation: 1, updatedAt: '2026-09-20T12:00:01Z', connections: [] }
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify(pending), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify(ready), { status: 200 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		const check = await screen.findByRole('button', { name: 'Check inventory status' })
		expect(fetchMock).toHaveBeenCalledTimes(2)
		fireEvent.click(check)
		expect(await screen.findByText('No brokerage connections are available.')).toBeVisible()
		expect(fetchMock).toHaveBeenCalledTimes(3)
	})

	it.each([
		['disabled', 'Your brokerage connection needs repair before Findur can retrieve accounts.', 'Return to SnapTrade authorization'],
		['unauthorized', 'SnapTrade authorization is no longer valid.', 'Return to SnapTrade authorization'],
		['rate_limited', 'SnapTrade asked Findur to wait before trying again.', 'Retry masked inventory'],
		['malformed', 'SnapTrade returned account information Findur could not safely use.', 'Retry masked inventory'],
	] as const)('renders the %s categorical recovery branch', async (state, message, action) => {
		window.history.replaceState(null, '', '/portfolio')
		const inventory = {
			state, generation: 1, updatedAt: '2026-09-20T12:00:00Z', retryAt: state === 'rate_limited' ? '2026-09-20T12:01:00Z' : undefined,
			connections: state === 'disabled' ? [{ id: 'disabled', brokerageLabel: 'Synthetic Broker', status: 'disabled', syncMode: 'unknown', available: false, eligible: false, accounts: [] }] : [],
		}
		vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/auth/status'
			? new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })
			: new Response(JSON.stringify(inventory), { status: 200 }))))
		render(<App />)
		expect(await screen.findByText(message)).toBeVisible()
		expect(screen.getByRole('button', { name: action })).toBeVisible()
		if (state === 'disabled') {
			expect(screen.getByText('Synthetic Broker')).toBeVisible()
			expect(screen.getByText('Availability: Unavailable')).toBeVisible()
			expect(screen.getByText('Later-inclusion eligibility: Not eligible')).toBeVisible()
		}
	})

	it('uses the explicit defended retry action for a recoverable inventory state', async () => {
		window.history.replaceState(null, '', '/portfolio')
		document.cookie = 'findur_csrf=inventory-csrf; Path=/'
		const unavailable = { state: 'unavailable', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		const ready = { state: 'empty', generation: 2, updatedAt: '2026-09-20T12:01:00Z', connections: [] }
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify(unavailable), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify(ready), { status: 200 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('button', { name: 'Retry masked inventory' }))
		expect(await screen.findByText('No brokerage connections are available.')).toBeVisible()
		expect(fetchMock).toHaveBeenLastCalledWith('/api/portfolio/inventory/retry', expect.objectContaining({
			method: 'POST', cache: 'no-store', credentials: 'same-origin', headers: { 'X-CSRF-Token': 'inventory-csrf' },
		}))
	})

	it('treats retry 403 as session-defense recovery instead of a provider outage', async () => {
		window.history.replaceState(null, '', '/portfolio')
		document.cookie = 'findur_csrf=inventory-csrf; Path=/'
		const unavailable = { state: 'unavailable', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify(unavailable), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ code: 'forbidden' }), { status: 403 }))
			.mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		fireEvent.click(await screen.findByRole('button', { name: 'Retry masked inventory' }))
		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByRole('button', { name: 'Retry masked inventory' })).not.toBeInTheDocument()
	})

	it('formats a rate-limit retry timestamp in the selected French locale', async () => {
		window.localStorage.setItem('findur-locale', 'fr')
		window.history.replaceState(null, '', '/portfolio')
		const retryAt = '2026-09-20T12:01:00Z'
		const inventory = { state: 'rate_limited', generation: 1, updatedAt: '2026-09-20T12:00:00Z', retryAt, connections: [] }
		vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/auth/status'
			? new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })
			: new Response(JSON.stringify(inventory), { status: 200 }))))
		render(<App />)
		const expected = new Date(retryAt).toLocaleString('fr-CA')
		expect(await screen.findByText(expected)).toBeVisible()
		expect(screen.getByText(/Moment sûr pour réessayer/)).toBeVisible()
	})

	it('routes an inventory session 401 to authentication recovery without offering retry', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const fetchMock = vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/portfolio/inventory'
			? new Response(JSON.stringify({ code: 'unauthenticated' }), { status: 401 })
			: new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByRole('button', { name: 'Retry masked inventory' })).not.toBeInTheDocument()
		expect(fetchMock).toHaveBeenCalledWith('/api/portfolio/inventory', expect.objectContaining({ method: 'GET', cache: 'no-store', credentials: 'same-origin' }))
	})

	it('logs out centrally, clears protected storage, and retains only locale and theme', async () => {
		window.history.replaceState(null, '', '/portfolio')
		window.localStorage.setItem('findur-locale', 'fr')
		window.localStorage.setItem('findur-theme', 'dark')
		window.localStorage.setItem('protected-payload', 'secret')
		window.sessionStorage.setItem('private-task', 'secret')
		document.cookie = 'findur_csrf=csrf-token; Path=/'
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(emptyInventory())
			.mockResolvedValueOnce(new Response(null, { status: 204 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		fireEvent.click(await screen.findByRole('button', { name: 'Se déconnecter' }))
		await waitFor(() => expect(window.location.pathname).toBe('/'))
		expect(window.localStorage.getItem('findur-locale')).toBe('fr')
		expect(window.localStorage.getItem('findur-theme')).toBe('dark')
		expect(window.localStorage.getItem('protected-payload')).toBeNull()
		expect(window.sessionStorage.getItem('private-task')).toBeNull()
		expect(fetchMock).toHaveBeenLastCalledWith('/api/auth/logout', expect.objectContaining({ method: 'POST', credentials: 'same-origin', headers: { 'X-CSRF-Token': 'csrf-token' } }))
	})

	it('keeps the authenticated shell retryable when logout fails', async () => {
		window.history.replaceState(null, '', '/portfolio')
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(emptyInventory())
			.mockResolvedValueOnce(new Response(JSON.stringify({ code: 'forbidden' }), { status: 403 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		fireEvent.click(await screen.findByRole('button', { name: 'Log out' }))
		expect(await screen.findByRole('alert')).toHaveTextContent('Your session is still active. Please retry.')
		expect(screen.getByRole('button', { name: 'Retry logout' })).toBeEnabled()
		expect(window.location.pathname).toBe('/portfolio')
	})

	it('cleans up and exits the shell when logout reports an already-ended session', async () => {
		window.history.replaceState(null, '', '/portfolio')
		window.localStorage.setItem('protected-payload', 'secret')
		const fetchMock = vi.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: false, authenticated: true }), { status: 200 }))
			.mockResolvedValueOnce(emptyInventory())
			.mockResolvedValueOnce(new Response(JSON.stringify({ code: 'unauthenticated' }), { status: 401 }))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		fireEvent.click(await screen.findByRole('button', { name: 'Log out' }))
		await waitFor(() => expect(window.location.pathname).toBe('/'))
		expect(window.localStorage.getItem('protected-payload')).toBeNull()
	})

	it('keeps consent closed with localized guidance when the server gate is closed', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: false, authenticated: false }), { status: 200 })))
		window.history.replaceState(null, '', '/connect')
		render(<App />)
		const action = screen.getByRole('button', { name: 'Continue to SnapTrade' })
		expect(action).toBeDisabled()
		expect(await screen.findByText(/Authorization is not available yet/)).toBeVisible()
		expect(action).toHaveAccessibleDescription(/Authorization is not available yet/)
	})

  it('preserves consent route and focus while locale and theme change', async () => {
    window.history.replaceState(null, '', '/connect')
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 })))
    render(<App />)
    fireEvent.click(screen.getAllByRole('radio', { name: 'FR' })[0])
    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('Connectez-vous sans partager votre mot de passe de courtage.')
    expect(screen.getByText('Aucun compte n’est inclus par défaut.')).toBeVisible()
    expect(screen.getByText(/Findur ne peut effectuer aucune opération/)).toBeVisible()
    expect(screen.getByText(/supprimer vos données du stockage actif/)).toBeVisible()
    expect(window.location.pathname).toBe('/connect')
    fireEvent.click(screen.getAllByRole('radio', { name: 'Sombre' })[0])
    await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'dark'))
    expect(window.location.pathname).toBe('/connect')
  })

  it('returns from consent without posting an authorization request', async () => {
    window.history.replaceState(null, '', '/connect')
	const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 })); vi.stubGlobal('fetch', fetchMock)
    render(<App />)
    fireEvent.click(screen.getByRole('link', { name: 'Return home' }))
    expect(window.location.pathname).toBe('/')
	await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1))
	expect(fetchMock).toHaveBeenCalledWith('/api/auth/status', expect.any(Object))
  })

  it('normalizes an unknown public path to the landing route', () => {
    window.history.replaceState(null, '', '/not-a-public-route')

    render(<App />)

    expect(window.location.pathname).toBe('/')
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Find a different pattern in the same sky.',
    )
  })

  it('loads the unlinked status route with one same-origin readiness request', async () => {
	window.history.replaceState(null, '', '/__status')
	const fetchMock = vi.fn().mockResolvedValue(
	  new Response(JSON.stringify({
	    status: 'ready',
	    buildSha: '0123456789abcdef0123456789abcdef01234567',
	  }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
	)
	vi.stubGlobal('fetch', fetchMock)

	render(<App />)

	expect(await screen.findByText('Build identity malformed')).toBeInTheDocument()
	expect(fetchMock).toHaveBeenCalledTimes(1)
	expect(fetchMock).toHaveBeenCalledWith('/api/readyz', expect.objectContaining({
	  cache: 'no-store',
	  credentials: 'same-origin',
	}))
	expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  })


  it('reports API failures categorically without exposing response details', async () => {
	window.history.replaceState(null, '', '/__status')
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('private failure', { status: 503 })))

	render(<App />)

	expect(await screen.findByText('API unavailable')).toBeInTheDocument()
	expect(screen.queryByText('private failure')).not.toBeInTheDocument()
  })

  it('renders the diagnostic page in the selected French locale', async () => {
    window.localStorage.setItem('findur-locale', 'fr')
    window.history.replaceState(null, '', '/__status')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      status: 'ready',
      buildSha: 'invalid',
    }), { status: 200, headers: { 'Content-Type': 'application/json' } })))

    render(<App />)

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('État de la version Findur')
    expect(await screen.findByText('Identité de version non valide')).toBeInTheDocument()
    expect(screen.getByText('Version de l’interface')).toBeInTheDocument()
    expect(document.title).toBe('État de la version — Findur')
  })

  it('bounds a pending readiness request and reports the API unavailable', () => {
    vi.useFakeTimers()
    window.history.replaceState(null, '', '/__status')
    const fetchMock = vi.fn().mockReturnValue(new Promise(() => {}))
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)
    act(() => vi.advanceTimersByTime(5_000))

    expect(screen.getByText('API unavailable')).toBeInTheDocument()
    const request = fetchMock.mock.calls[0][1] as RequestInit
    expect(request.signal?.aborted).toBe(true)
  })
})
