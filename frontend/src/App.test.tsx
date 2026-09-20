import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { App } from './App'

type MediaChangeListener = (event: MediaQueryListEvent) => void

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
			.mockResolvedValueOnce(new Response(JSON.stringify({ authorizationAvailable: true, authenticated: true }), { status: 200 }))
		vi.stubGlobal('fetch', fetchMock)
		window.history.replaceState(null, '', '/connect/result?status=success')
		const { unmount } = render(<App />)
		expect(screen.getByRole('heading', { level: 1, name: 'Checking your secure session…' })).toBeVisible()
		expect(window.location.pathname).toBe('/connect/result')
		expect(window.location.search).toBe('')
		expect(await screen.findByRole('heading', { level: 1, name: 'Please restart the secure connection.' })).toHaveFocus()
		unmount()
		window.history.replaceState(null, '', '/connect/result')
		render(<App />)
		expect(await screen.findByRole('heading', { level: 1, name: 'Your private session is ready.' })).toHaveFocus()
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
