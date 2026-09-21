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
    expect(screen.getByText('Private 18+ demo')).toBeInTheDocument()
    const loginLinks = screen.getAllByRole('link', { name: 'Log in' })
    expect(loginLinks).toHaveLength(2)
    expect(loginLinks.every((link) => link.getAttribute('href') === '/connect')).toBe(true)
    expect(screen.getAllByRole('navigation')).toHaveLength(1)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('navigates to About, updates metadata, and focuses the route heading', async () => {
    render(<App />)

    fireEvent.click(screen.getByRole('link', { name: /how findur works/i }))

    const heading = screen.getByRole('heading', { level: 1, name: /more to investing/i })
    await waitFor(() => expect(heading).toHaveFocus())
    expect(window.location.pathname).toBe('/about')
    expect(document.title).toBe('About Findur')
    expect(document.querySelector('meta[name="description"]')).toHaveAttribute(
      'content',
      expect.stringContaining('conversation starters'),
    )
    expect(screen.getByText(/mix of holdings, diversification, recent activity, and account coverage/)).toBeVisible()
    expect(screen.getByRole('heading', { name: 'The numbers never get the last word.' })).toBeVisible()
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
      'Investir, c’est bien plus que le solde.',
    )
    expect(screen.getAllByRole('navigation')).toHaveLength(1)

    unmount()
    render(<App />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Investir, c’est bien plus que le solde.',
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
      'There’s more to investing than the balance.',
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

    fireEvent.click(screen.getAllByRole('link', { name: 'Log in' })[0])

    const heading = screen.getByRole('heading', { level: 1, name: 'Log in to Findur.' })
    await waitFor(() => expect(heading).toHaveFocus())
    expect(window.location.pathname).toBe('/connect')
    expect(screen.getByText(/No account is included by default/)).toBeVisible()
    expect(screen.getByText(/connecting never publishes your portfolio/i)).toBeVisible()
    expect(screen.getByText(/Findur never receives or stores your brokerage credentials/)).toBeVisible()
    expect(screen.getByText(/connection status, and masked account details/)).toBeVisible()
    expect(screen.getByText(/signal building, profile previews, and Discovery stay off until you explicitly include at least one account/)).toBeVisible()
    expect(screen.getByText(/Confirming selected accounts starts private analysis and profile previews/)).toBeVisible()
    expect(screen.getByText(/What a match may see remains a separate choice/)).toBeVisible()
    expect(screen.getByText(/Findur is read-only/)).toBeVisible()
    expect(screen.getByText(/deletes your complete account from app-controlled active storage/)).toBeVisible()
	await waitFor(() => expect(screen.getByRole('button', { name: 'Continue with SnapTrade' })).toBeEnabled())
	expect(fetchMock).toHaveBeenCalledWith('/api/auth/status', expect.objectContaining({ cache: 'no-store', credentials: 'same-origin' }))
  })

	it('rejects a forged success URL and trusts only server session status', async () => {
		const fetchMock = vi.fn().mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 }))
			.mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })))
		vi.stubGlobal('fetch', fetchMock)
		window.history.replaceState(null, '', '/connect/result?status=success')
		const { unmount } = render(<App />)
		expect(screen.getByRole('status')).toHaveTextContent('Checking your secure session…')
		expect(window.location.pathname).toBe('/onboarding/accounts')
		expect(window.location.search).toBe('')
		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByText('Choose accounts before anything else.')).not.toBeInTheDocument()
		unmount()
		window.history.replaceState(null, '', '/onboarding/accounts')
		render(<App />)
		const portfolioHeading = await screen.findByRole('heading', { level: 1, name: 'Choose what Findur may use.' })
		await waitFor(() => expect(portfolioHeading).toHaveFocus())
		expect(window.location.pathname).toBe('/onboarding/accounts')
	})

	it('gates private content before mounting and exposes only the three private destinations', async () => {
		let resolveStatus!: (response: Response) => void
		vi.stubGlobal('fetch', vi.fn().mockReturnValue(new Promise<Response>((resolve) => { resolveStatus = resolve })))
		window.history.replaceState(null, '', '/portfolio')
		render(<App />)
		expect(screen.getByRole('status')).toHaveTextContent('Checking your secure session…')
		expect(screen.queryByText('Choose accounts before anything else.')).not.toBeInTheDocument()
		resolveStatus(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
		const heading = await screen.findByRole('heading', { level: 1, name: 'Your portfolio profile is taking shape.' })
		await waitFor(() => expect(heading).toHaveFocus())
		for (const name of ['Discovery', 'Portfolio', 'Profile']) expect(screen.getAllByRole('link', { name })).toHaveLength(1)
		expect(screen.getAllByRole('navigation')).toHaveLength(1)
		fireEvent.click(screen.getByRole('link', { name: 'Profile' }))
		const profile = screen.getByRole('heading', { level: 1, name: 'Profile setup comes later.' })
		await waitFor(() => expect(profile).toHaveFocus())
		expect(window.location.pathname).toBe('/profile')
		window.history.back()
		window.dispatchEvent(new PopStateEvent('popstate'))
		await waitFor(() => expect(screen.getByRole('heading', { level: 1, name: 'Your portfolio profile is taking shape.' })).toHaveFocus())
	})

	it('renders one integrated account selection list without render-driven repeats', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
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

		expect(await screen.findByText('Retirement (•••• 8443)')).toBeVisible()
		expect(screen.getByRole('heading', { level: 1, name: 'Choose what Findur may use.' })).toBeVisible()
		expect(screen.getAllByText('Retirement (•••• 8443)')).toHaveLength(1)
		expect(screen.getByRole('heading', { name: 'Connection setup' })).toBeInTheDocument()
		expect(screen.getByText('1 · Connect', { selector: 'li' })).toBeVisible()
		expect(screen.getByText('2 · Choose accounts', { selector: 'li' })).toHaveAttribute('aria-current', 'step')
		expect(screen.getByText('3 · Portfolio showcase', { selector: 'li' })).toBeVisible()
		expect(screen.getByText('4 · Preferences & privacy', { selector: 'li' })).toBeVisible()
		expect(screen.getByText('5 · Preview & discover', { selector: 'li' })).toBeVisible()
		expect(screen.getByText('3 · Portfolio')).toBeInTheDocument()
		expect(screen.getByText(/Pick the accounts you’d like to include/)).toBeVisible()
		expect(screen.getByText(/You can change this anytime/)).toBeVisible()
		const status = screen.getByRole('group', { name: 'Your accounts' })
		const account = screen.getByText('Retirement (•••• 8443)')
		expect(status.compareDocumentPosition(account) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
		expect(document.body).not.toHaveTextContent('Q6542138443')
		expect(document.body).not.toHaveTextContent('$')
		expect(screen.queryByRole('heading', { level: 3, name: 'Synthetic Broker' })).not.toBeInTheDocument()
		expect(screen.getAllByText(/Synthetic Broker ·/)).toHaveLength(2)
		expect(screen.getAllByText('Ready')).toHaveLength(2)
		expect(screen.getByText('0 / 2')).toBeVisible()
		expect(screen.queryByText(/Ready to choose/)).not.toBeInTheDocument()
		expect(screen.getByText(/Synthetic Broker · Category unavailable/)).toBeVisible()
		expect(document.body).not.toHaveTextContent('· unknown')
		expect(screen.getByText('Everyday account')).toBeVisible()
		expect(screen.queryByText(/sync mode|inclusion readiness|confirmation scope|permitted private purposes|purge|recalculation/i)).not.toBeInTheDocument()
		expect(screen.queryByRole('heading', { name: 'Connection status' })).not.toBeInTheDocument()
		expect(screen.queryByRole('navigation', { name: 'Account navigation' })).not.toBeInTheDocument()
		expect(screen.queryByRole('link', { name: 'findur' })).not.toBeInTheDocument()

		fireEvent.click(screen.getAllByRole('radio', { name: 'FR' })[0])
		const frenchHeading = await screen.findByRole('heading', { level: 1, name: 'Choisissez ce que Findur peut utiliser.' })
		await waitFor(() => expect(frenchHeading).toHaveFocus())
		expect(screen.getByText(/Enregistrés dans votre profil : 0 sur 2 comptes affichés/)).toBeVisible()
		fireEvent.click(screen.getAllByRole('radio', { name: 'Sombre' })[0])
		await waitFor(() => expect(document.documentElement).toHaveAttribute('data-theme', 'dark'))
		expect(fetchMock).toHaveBeenCalledTimes(3)
	})

	it('keeps the draft separate, reviews changes compactly, and continues after a committed save', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		document.cookie = 'findur_csrf=inclusion-csrf; Path=/'
		const replaceState = vi.spyOn(window.history, 'replaceState')
		const inventory = {
			state: 'ready', generation: 4, updatedAt: '2026-09-20T12:00:00Z', connections: [{
				id: 'connection-1', brokerageLabel: 'Synthetic Broker', status: 'active', syncMode: 'realtime', available: true, eligible: true,
				accounts: [
					{ id: 'account-1', category: 'investment', type: 'Margin', maskedLabel: 'Retirement (•••• 8443)', available: true, eligible: true, selectable: true, usabilityReason: 'ready', syncState: 'complete' },
					{ id: 'account-2', category: 'deposit', type: 'Checking', maskedLabel: 'Daily cash (•••• 1000)', available: true, eligible: false, selectable: false, usabilityReason: 'unsupported_category', syncState: 'complete' },
				],
			}],
		}
		let resolveSave!: (response: Response) => void
		const saveResponse = new Promise<Response>((resolve) => { resolveSave = resolve })
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(new Response(JSON.stringify(inventory), { status: 200 }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(new Response(JSON.stringify({ version: 0, committed: [] }), { status: 200 }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return saveResponse
			return Promise.resolve(new Response(null, { status: 404 }))
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		const retirement = await screen.findByRole('checkbox', { name: /Retirement/ })
		const disabled = screen.getByRole('checkbox', { name: /Daily cash/ })
		expect(retirement).not.toBeChecked()
		expect(disabled).toBeDisabled()
		expect(screen.getByText(/Unavailable because Findur supports investment accounts only/)).toBeVisible()
		expect(screen.getByText(/Saved to your profile: 0 of 2 accounts shown/)).toBeVisible()
		const review = screen.getByRole('button', { name: 'Review my choices' })
		expect(review).toBeDisabled()
		fireEvent.click(retirement)
		expect(screen.getByText('1 / 2')).toBeVisible()
		expect(screen.getByText(/Selected after saving: 1 of 2 accounts shown/)).toBeVisible()
		expect(review).toBeEnabled()
		fireEvent.click(review)
		const dialog = screen.getByRole('dialog', { name: 'Ready to save your choices?' })
		expect(dialog).toHaveTextContent('Your profile will include: 1 of 2')
		expect(dialog).toHaveTextContent('Adding')
		expect(dialog).toHaveTextContent('Retirement (•••• 8443)')
		expect(dialog).toHaveTextContent('Synthetic Broker')
		expect(dialog).toHaveTextContent('You can change these accounts anytime')
		expect(dialog).not.toHaveTextContent(/provider|balance|position|purpose|lifecycle|purge/i)
		fireEvent.click(screen.getByRole('button', { name: 'Back' }))
		await waitFor(() => expect(review).toHaveFocus())
		expect(retirement).toBeChecked()
		fireEvent.click(review)
		await waitFor(() => expect(screen.getByRole('button', { name: 'Back' })).toHaveFocus())
		fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' })
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
		await waitFor(() => expect(review).toHaveFocus())
		expect(retirement).toBeChecked()
		fireEvent.click(review)
		const save = screen.getByRole('button', { name: 'Save my choices' })
		fireEvent.click(save)
		expect(screen.getByText('Saving your account choices…')).toBeVisible()
		expect(review).toBeDisabled()
		expect(screen.getByText('Saving your account choices…').closest('.account-inclusion')).toHaveAttribute('aria-busy', 'true')
		await act(async () => {
			resolveSave(new Response(JSON.stringify({ version: 1, committed: ['account-1'], change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'committed', additions: ['account-1'], removals: [] } }), { status: 200 }))
		})
		await waitFor(() => expect(window.location.pathname).toBe('/onboarding/portfolio'))
		expect(replaceState).toHaveBeenLastCalledWith(null, '', '/onboarding/portfolio')
		expect(screen.getByRole('heading', { level: 1, name: 'Your portfolio profile is taking shape.' })).toBeVisible()
		expect(screen.getByRole('status')).toHaveTextContent('Account choices saved. Your portfolio profile is ready for the next step.')
		expect(screen.getByRole('link', { name: 'Edit included accounts' })).toHaveAttribute('href', '/onboarding/accounts')
		const post = fetchMock.mock.calls.find(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(post?.[1]).toEqual(expect.objectContaining({
			cache: 'no-store', credentials: 'same-origin', body: JSON.stringify({ accountIds: ['account-1'] }),
			headers: expect.objectContaining({ 'X-CSRF-Token': 'inclusion-csrf', 'X-Inclusion-Version': '0' }),
		}))
	})

	it('omits closed and unactionable temporary accounts from initial setup', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([
			inclusionAccount('account-closed', 'Closed account (•••• 1111)', false, 'account_closed'),
			inclusionAccount('account-loading', 'New account (•••• 2222)', false, 'sync_pending'),
			inclusionAccount('account-unavailable', 'Unavailable account (•••• 3333)', false, 'connection_unavailable'),
			inclusionAccount('account-ready', 'Ready account (•••• 4444)', true, 'ready'),
		])
		installInclusionFetch(inventory, { version: 7, committed: [] })
		render(<App />)

		expect(await screen.findByRole('checkbox', { name: /Ready account/ })).not.toBeChecked()
		expect(screen.queryByRole('checkbox', { name: /New account/ })).not.toBeInTheDocument()
		expect(screen.queryByRole('checkbox', { name: /Unavailable account/ })).not.toBeInTheDocument()
		expect(screen.queryByText('Closed account (•••• 1111)')).not.toBeInTheDocument()
		expect(screen.queryByText('Still getting this account ready.')).not.toBeInTheDocument()
		expect(screen.queryByText('We cannot reach this connection right now.')).not.toBeInTheDocument()
		expect(screen.getByText(/Saved to your profile: 0 of 1 accounts shown/)).toBeVisible()
	})

	it('keeps an existing durable selection available to review without a just-saved notice', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([
			inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready'),
			inclusionAccount('account-2', 'Growth (•••• 2002)', true, 'ready'),
		])
		installInclusionFetch(inventory, { version: 4, committed: ['account-1'] })
		render(<App />)

		expect(await screen.findByRole('checkbox', { name: /Retirement/ })).toBeChecked()
		expect(window.location.pathname).toBe('/onboarding/accounts')
		expect(screen.getByRole('button', { name: 'Review my choices' })).toBeEnabled()
		expect(screen.queryByText(/Account choices saved/)).not.toBeInTheDocument()
	})

	it('keeps account setup visible for recovery when durable choices exist', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = { state: 'unavailable', generation: 4, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		installInclusionFetch(inventory, { version: 4, committed: ['account-1'] })
		render(<App />)

		expect(await screen.findByRole('button', { name: 'Try again' })).toBeVisible()
		expect(window.location.pathname).toBe('/onboarding/accounts')
		expect(screen.queryByText(/Account choices saved/)).not.toBeInTheDocument()
	})

	it('reconstructs failed additions for a scoped retry while keeping removals effective', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
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

		expect(await screen.findByText(/Selected after saving: 1 of 2 accounts shown/)).toBeVisible()
		expect(screen.getByText(/Saved to your profile: 0 of 2 accounts shown/)).toBeVisible()
		expect(screen.getAllByText('Ready')).toHaveLength(2)
		expect(screen.getByRole('checkbox', { name: /Removed/ })).not.toBeChecked()
		expect(screen.queryByRole('button', { name: 'Reload inclusion status' })).not.toBeInTheDocument()
		const retry = screen.getByRole('button', { name: 'Review my choices' })
		expect(retry).toBeEnabled()
		expect(screen.getByRole('alert')).toHaveTextContent('We couldn’t save every choice. Review them and try again.')
		fireEvent.click(retry)
		expect(screen.getByRole('dialog', { name: 'Ready to save your choices?' })).toHaveTextContent('Retirement (•••• 8443)')
	})

	it('offers one clear retry for a durable pending addition with the same scoped draft', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		document.cookie = 'findur_csrf=pending-csrf; Path=/'
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		const pending = {
			version: 1,
			committed: [],
			change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'pending', additions: ['account-1'], removals: [] },
		}
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(jsonResponseBody(pending))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(jsonResponseBody({ version: 2, committed: ['account-1'], change: { ...pending.change, status: 'committed' } }))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		expect(await screen.findByText(/Selected after saving: 1 of 1 accounts shown/)).toBeVisible()
		expect(screen.queryByRole('button', { name: 'Reload inclusion status' })).not.toBeInTheDocument()
		const pendingChoice = screen.getByRole('checkbox', { name: /Retirement/ })
		expect(pendingChoice).toBeEnabled()
		fireEvent.click(pendingChoice)
		expect(screen.getByText('0 / 1')).toBeVisible()
		fireEvent.click(pendingChoice)
		expect(screen.getByText('1 / 1')).toBeVisible()
		const retry = screen.getByRole('button', { name: 'Review my choices' })
		expect(screen.getByText('Finishing your update…')).toBeVisible()
		fireEvent.click(retry)
		expect(screen.getByRole('dialog', { name: 'Ready to save your choices?' })).toBeVisible()
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		await waitFor(() => expect(window.location.pathname).toBe('/onboarding/portfolio'))
		const post = fetchMock.mock.calls.find(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(post?.[1]).toEqual(expect.objectContaining({
			body: JSON.stringify({ accountIds: ['account-1'] }),
			headers: expect.objectContaining({ 'X-Inclusion-Version': '1', 'X-CSRF-Token': 'pending-csrf' }),
		}))
	})


	it.each([
		['pending', 'Finishing your update…', 'status'],
		['failed', 'We couldn’t save every choice. Review them and try again.', 'alert'],
	] as const)('keeps a %s POST result on setup with one accurate announcement', async (status, message, role) => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		const change = { id: '87b24961-b51e-4db8-9226-f198f6518a89', status, additions: ['account-1'], removals: [], ...(status === 'failed' ? { failureReason: 'provider_unavailable' } : {}) }
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(jsonResponseBody({ version: 0, committed: [] }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(jsonResponseBody({ version: 1, committed: [], change }))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Retirement/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		expect(await screen.findByRole(role)).toHaveTextContent(message)
		expect(window.location.pathname).toBe('/onboarding/accounts')
		expect(screen.getByRole('checkbox', { name: /Retirement/ })).toBeChecked()
		expect(screen.getByText('Ready')).toBeVisible()
	})

	it('reconciles an ambiguous failed request as success when durable choices match', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		let getCalls = 0
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') {
				getCalls += 1
				return Promise.resolve(jsonResponseBody(getCalls === 1 ? { version: 0, committed: [] } : { version: 1, committed: ['account-1'] }))
			}
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.reject(new TypeError('connection lost'))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Retirement/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		await waitFor(() => expect(window.location.pathname).toBe('/onboarding/portfolio'))
		expect(screen.getByRole('status')).toHaveTextContent('Account choices saved')
		expect(getCalls).toBe(2)
	})

	it('reports a reconciled durable pending change instead of masking it as a request failure', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		let getCalls = 0
		const pending = { version: 1, committed: [], change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'pending', additions: ['account-1'], removals: [] } }
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') {
				getCalls += 1
				return Promise.resolve(jsonResponseBody(getCalls === 1 ? { version: 0, committed: [] } : pending))
			}
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.reject(new TypeError('connection lost'))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Retirement/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		expect(await screen.findByRole('status')).toHaveTextContent('Finishing your update…')
		expect(screen.queryByRole('alert')).not.toBeInTheDocument()
	})

	it('refreshes stale durable state before showing an unchanged save failure and retrying', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-1', 'Retirement (•••• 8443)', true, 'ready')])
		let getCalls = 0
		let postCalls = 0
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') {
				getCalls += 1
				return Promise.resolve(jsonResponseBody({ version: getCalls === 1 ? 0 : 4, committed: [] }))
			}
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') {
				postCalls += 1
				return postCalls === 1
					? Promise.resolve(new Response(JSON.stringify({ code: 'conflict' }), { status: 409 }))
					: Promise.resolve(jsonResponseBody({ version: 5, committed: [], change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'pending', additions: ['account-1'], removals: [] } }))
			}
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		const account = await screen.findByRole('checkbox', { name: /Retirement/ })
		fireEvent.click(account)
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		expect(await screen.findByRole('alert')).toHaveTextContent('We couldn’t save every choice. Review them and try again.')
		expect(account).not.toBeChecked()
		expect(window.location.pathname).toBe('/onboarding/accounts')

		fireEvent.click(account)
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		expect(await screen.findByRole('status')).toHaveTextContent('Finishing your update…')
		const posts = fetchMock.mock.calls.filter(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(posts[1]?.[1]).toEqual(expect.objectContaining({ headers: expect.objectContaining({ 'X-Inclusion-Version': '4' }) }))
	})

	it('keeps a committed unavailable account enabled and presents removal as an ordinary edit', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([
			inclusionAccount('account-unavailable', 'Unavailable account (•••• 3003)', false, 'connection_unavailable'),
			inclusionAccount('account-ready', 'Ready account (•••• 4004)', true, 'ready'),
			inclusionAccount('account-ready-two', 'Second ready account (•••• 5005)', true, 'ready'),
		])
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(jsonResponseBody({ version: 0, committed: [] }))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(jsonResponseBody({ version: 1, committed: ['account-unavailable'], change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'failed', additions: ['account-ready'], removals: [], failureReason: 'provider_unavailable' } }))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Ready account/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		await screen.findByRole('alert')
		const unavailable = screen.getByRole('checkbox', { name: /Unavailable account/ })
		const ready = screen.getByRole('checkbox', { name: /^Ready account/ })
		const secondReady = screen.getByRole('checkbox', { name: /Second ready account/ })
		const selectAll = screen.getByRole('checkbox', { name: /^Select all accounts/ })
		expect(unavailable).toBeChecked()
		expect(unavailable).toBeEnabled()
		expect(unavailable.closest('label')).not.toHaveClass('account-choice--disabled')
		expect(selectAll).toBePartiallyChecked()
		fireEvent.click(selectAll)
		expect(selectAll).toBeChecked()
		expect(ready).toBeChecked()
		expect(secondReady).toBeChecked()
		expect(unavailable).toBeChecked()
		fireEvent.click(selectAll)
		expect(selectAll).not.toBeChecked()
		expect(ready).not.toBeChecked()
		expect(secondReady).not.toBeChecked()
		expect(unavailable).toBeChecked()
		fireEvent.click(unavailable)
		expect(screen.getByText('Unavailable')).toBeVisible()
		expect(screen.getByRole('button', { name: 'Review my choices' })).toBeDisabled()
		fireEvent.click(ready)
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		const dialog = screen.getByRole('dialog', { name: 'Ready to save your choices?' })
		expect(dialog).toHaveTextContent('Removing')
		expect(dialog).toHaveTextContent('Unavailable account (•••• 3003)')
		expect(dialog).not.toHaveTextContent(/delete|purge|irreversible|retriev/i)
	})

	it('shows a recovery action instead of selection controls when ready accounts are all closed', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-closed', 'Closed account (•••• 1111)', false, 'account_closed')])
		installInclusionFetch(inventory, { version: 0, committed: [] })
		render(<App />)

		expect(await screen.findByText('No accounts are available to add to your profile right now.')).toBeVisible()
		expect(screen.getByRole('button', { name: 'Reconnect or manage accounts' })).toBeVisible()
		expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
		expect(screen.queryByRole('button', { name: 'Review my choices' })).not.toBeInTheDocument()
	})

	it('disables Select All when every visible account is ineligible', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-cash', 'Daily cash (•••• 1111)', false, 'unsupported_category', 'deposit')])
		installInclusionFetch(inventory, { version: 0, committed: [] })
		render(<App />)

		expect(await screen.findByRole('checkbox', { name: /^Select all accounts/ })).toBeDisabled()
		expect(screen.getByRole('checkbox', { name: /Daily cash/ })).toBeDisabled()
	})

	it('keeps a durable failure visible when filtering leaves the chooser with no rows', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([inclusionAccount('account-closed', 'Closed account (•••• 1111)', false, 'account_closed')])
		installInclusionFetch(inventory, {
			version: 1,
			committed: [],
			change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'failed', additions: ['account-closed'], removals: [], failureReason: 'provider_unavailable' },
		})
		render(<App />)

		expect(await screen.findByText('No accounts are available to add to your profile right now.')).toBeVisible()
		expect(screen.getByRole('alert')).toHaveTextContent('We couldn’t save every choice. Review them and try again.')
	})

	it('does not resubmit a hidden failed addition when saving a visible choice', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const inventory = inclusionInventory([
			inclusionAccount('account-closed', 'Closed account (•••• 1111)', false, 'account_closed'),
			inclusionAccount('account-ready', 'Ready account (•••• 2222)', true, 'ready'),
		])
		const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(inventory))
			if (path === '/api/portfolio/inclusion' && init?.method === 'GET') return Promise.resolve(jsonResponseBody({
				version: 1,
				committed: [],
				change: { id: '87b24961-b51e-4db8-9226-f198f6518a89', status: 'failed', additions: ['account-closed'], removals: [], failureReason: 'provider_unavailable' },
			}))
			if (path === '/api/portfolio/inclusion' && init?.method === 'POST') return Promise.resolve(jsonResponseBody({
				version: 2,
				committed: [],
				change: { id: '66d954cb-5e83-4da1-8843-1d195d62a876', status: 'pending', additions: ['account-ready'], removals: [] },
			}))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('checkbox', { name: /Ready account/ }))
		fireEvent.click(screen.getByRole('button', { name: 'Review my choices' }))
		fireEvent.click(screen.getByRole('button', { name: 'Save my choices' }))
		await screen.findByText('Finishing your update…')

		const post = fetchMock.mock.calls.find(([path, init]) => path === '/api/portfolio/inclusion' && init?.method === 'POST')
		expect(post?.[1]).toEqual(expect.objectContaining({ body: JSON.stringify({ accountIds: ['account-ready'] }) }))
	})

	it('shares one in-flight initial inventory request across a development remount', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
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
		expect(await screen.findByText('We could not find any connected accounts.')).toBeVisible()
		expect(fetchMock.mock.calls.filter(([path]) => path === '/api/portfolio/inventory')).toHaveLength(1)
	})

	it('does not share an old in-flight inventory request with a session after logout', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
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
		window.history.replaceState(null, '', '/onboarding/accounts')
		render(<App />)
		expect(await screen.findByText('We could not find any connected accounts.')).toBeVisible()
		expect(inventoryCalls).toBe(2)
	})

	it('lets a pending observer explicitly check persisted status without polling', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const pending = { state: 'pending', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		const ready = { state: 'empty', generation: 1, updatedAt: '2026-09-20T12:00:01Z', connections: [] }
		let inventoryCalls = 0
		const fetchMock = vi.fn().mockImplementation((path: string) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inclusion') return Promise.resolve(jsonResponseBody({ version: 0, committed: [] }))
			if (path === '/api/portfolio/inventory') {
				inventoryCalls += 1
				return Promise.resolve(jsonResponseBody(inventoryCalls === 1 ? pending : ready))
			}
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		const check = await screen.findByRole('button', { name: 'Check again' })
		expect(inventoryCalls).toBe(1)
		fireEvent.click(check)
		expect(await screen.findByText('We could not find any connected accounts.')).toBeVisible()
		expect(inventoryCalls).toBe(2)
	})

	it('offers reconnection when no connected accounts are found', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/auth/status'
			? jsonResponseBody({ authorizationAvailable: true, authenticated: true })
			: emptyInventory())))
		render(<App />)

		expect(await screen.findByText('We could not find any connected accounts.')).toBeVisible()
		expect(screen.getByRole('button', { name: 'Reconnect accounts' })).toBeVisible()
		expect(screen.getAllByText('Account setup')[0]).toBeVisible()
		expect(screen.queryByText('Connection complete')).not.toBeInTheDocument()
		expect(screen.getByText('1 · Connect', { selector: 'li' })).toHaveAttribute('aria-current', 'step')
	})

	it.each([
		['disabled', 'Your connection needs attention before we can load your accounts.', 'Reconnect accounts'],
		['unauthorized', 'Your connection needs to be renewed before we can load your accounts.', 'Reconnect accounts'],
		['rate_limited', 'Your accounts are not ready yet. Check again in a moment.', 'Try again'],
		['malformed', 'We couldn’t load your accounts.', 'Try again'],
	] as const)('renders the %s categorical recovery branch', async (state, message, action) => {
		window.history.replaceState(null, '', '/onboarding/accounts')
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
		expect(screen.getAllByText('Account setup')[0]).toBeVisible()
		expect(screen.queryByText('Connection complete')).not.toBeInTheDocument()
		expect(screen.getByText('1 · Connect', { selector: 'li' })).toHaveAttribute('aria-current', 'step')
		expect(screen.getByText('2 · Choose accounts', { selector: 'li' })).not.toHaveAttribute('aria-current')
		if (state === 'disabled') {
			expect(screen.getByText('Synthetic Broker')).toBeVisible()
			expect(screen.getByText('Needs repair')).toBeVisible()
		}
	})

	it('uses the explicit defended retry action for a recoverable inventory state', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		document.cookie = 'findur_csrf=inventory-csrf; Path=/'
		const unavailable = { state: 'unavailable', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		const ready = { state: 'empty', generation: 2, updatedAt: '2026-09-20T12:01:00Z', connections: [] }
		const fetchMock = vi.fn().mockImplementation((path: string) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated: true }))
			if (path === '/api/portfolio/inclusion') return Promise.resolve(jsonResponseBody({ version: 0, committed: [] }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(unavailable))
			if (path === '/api/portfolio/inventory/retry') return Promise.resolve(jsonResponseBody(ready))
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		fireEvent.click(await screen.findByRole('button', { name: 'Try again' }))
		expect(await screen.findByText('We could not find any connected accounts.')).toBeVisible()
		expect(fetchMock.mock.calls).toContainEqual(['/api/portfolio/inventory/retry', expect.objectContaining({
			method: 'POST', cache: 'no-store', credentials: 'same-origin', headers: { 'X-CSRF-Token': 'inventory-csrf' },
		})])
	})

	it('treats retry 403 as session-defense recovery instead of a provider outage', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		document.cookie = 'findur_csrf=inventory-csrf; Path=/'
		const unavailable = { state: 'unavailable', generation: 1, updatedAt: '2026-09-20T12:00:00Z', connections: [] }
		let authenticated = true
		const fetchMock = vi.fn().mockImplementation((path: string) => {
			if (path === '/api/auth/status') return Promise.resolve(jsonResponseBody({ authorizationAvailable: true, authenticated }))
			if (path === '/api/portfolio/inclusion') return Promise.resolve(jsonResponseBody({ version: 0, committed: [] }))
			if (path === '/api/portfolio/inventory') return Promise.resolve(jsonResponseBody(unavailable))
			if (path === '/api/portfolio/inventory/retry') {
				authenticated = false
				return Promise.resolve(new Response(JSON.stringify({ code: 'forbidden' }), { status: 403 }))
			}
			throw new Error(`unexpected request ${path}`)
		})
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)
		fireEvent.click(await screen.findByRole('button', { name: 'Try again' }))
		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument()
	})

	it('formats a rate-limit retry timestamp in the selected French locale', async () => {
		window.localStorage.setItem('findur-locale', 'fr')
		window.history.replaceState(null, '', '/onboarding/accounts')
		const retryAt = '2026-09-20T12:01:00Z'
		const inventory = { state: 'rate_limited', generation: 1, updatedAt: '2026-09-20T12:00:00Z', retryAt, connections: [] }
		vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/auth/status'
			? new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })
			: new Response(JSON.stringify(inventory), { status: 200 }))))
		render(<App />)
		const expected = new Date(retryAt).toLocaleString('fr-CA')
		expect(await screen.findByText(expected)).toBeVisible()
		expect(screen.getByText(/Réessayer après/)).toBeVisible()
	})

	it('routes an inventory session 401 to authentication recovery without offering retry', async () => {
		window.history.replaceState(null, '', '/onboarding/accounts')
		const fetchMock = vi.fn().mockImplementation((path: string) => Promise.resolve(path === '/api/portfolio/inventory'
			? new Response(JSON.stringify({ code: 'unauthenticated' }), { status: 401 })
			: new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 })))
		vi.stubGlobal('fetch', fetchMock)
		render(<App />)

		await waitFor(() => expect(window.location.pathname).toBe('/connect'))
		expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument()
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
		const action = screen.getByRole('button', { name: 'Continue with SnapTrade' })
		expect(action).toBeDisabled()
		expect(await screen.findByText(/Login is unavailable right now/)).toBeVisible()
		expect(action).toHaveAccessibleDescription(/Login is unavailable right now/)
	})

  it('preserves consent route and focus while locale and theme change', async () => {
    window.history.replaceState(null, '', '/connect')
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: false }), { status: 200 })))
    render(<App />)
    fireEvent.click(screen.getAllByRole('radio', { name: 'FR' })[0])
    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('Connectez-vous à Findur.')
    expect(screen.getByText(/Aucun compte n’est inclus par défaut/)).toBeVisible()
    expect(screen.getByText(/Findur ne reçoit ni ne conserve jamais vos identifiants de courtage/)).toBeVisible()
    expect(screen.getByText(/l’état de la connexion et les renseignements masqués/)).toBeVisible()
    expect(screen.getByText(/jusqu’à ce que vous incluiez explicitement au moins un compte/)).toBeVisible()
    expect(screen.getByText(/La confirmation des comptes choisis lance l’analyse privée/)).toBeVisible()
    expect(screen.getByText(/Ce qu’un match peut voir reste un choix distinct/)).toBeVisible()
    expect(screen.getByText(/Findur est en lecture seule/)).toBeVisible()
    expect(screen.getByText(/supprime votre compte complet du stockage actif/)).toBeVisible()
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
