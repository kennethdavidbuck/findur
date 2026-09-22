import { useCallback, useEffect, useLayoutEffect, useRef, useState, type RefObject } from 'react'
import { getAuthorizationStatus, useAuthorizationStatus, type AuthorizationStatus } from './auth-status'
import { AuthenticatedLayout, type ProtectedRoute } from './components/AuthenticatedLayout'
import { PublicLayout } from './components/PublicLayout'
import { I18nProvider, useI18n } from './i18n'
import { AboutPage } from './pages/AboutPage'
import { ConsentPage } from './pages/ConsentPage'
import { LandingPage } from './pages/LandingPage'
import { PortfolioPage } from './pages/PortfolioPage'
import { PortfolioShowcasePage } from './pages/PortfolioShowcasePage'
import { ProfilePage } from './pages/ProfilePage'
import { StatusPage } from './pages/StatusPage'
import { getPortfolioInclusion, InventorySessionExpiredError, resetInitialInventoryRequest, type PortfolioInclusion } from './inventory'
import { endCurrentSession } from './session'
import { ThemeProvider } from './theme'
import { AuthenticatedPreferences } from './authenticated-preferences'

type PublicRoute = '/' | '/about' | '/connect' | '/__status'
type OnboardingRoute = '/onboarding/accounts'
type AccountEditRoute = '/portfolio/accounts'
type AppRoute = PublicRoute | ProtectedRoute | OnboardingRoute | AccountEditRoute
type ConnectHandoff = { id: number; status: AuthorizationStatus }

function routeFromPath(pathname: string): AppRoute {
  if (pathname === '/__status' || pathname === '/__status/') return '/__status'
  if (pathname === '/connect' || pathname === '/connect/') return '/connect'
  if (pathname === '/onboarding/accounts' || pathname === '/onboarding/accounts/' || pathname === '/connect/result' || pathname === '/connect/result/') return '/onboarding/accounts'
  if (pathname === '/portfolio/accounts' || pathname === '/portfolio/accounts/') return '/portfolio/accounts'
  if (pathname === '/discovery' || pathname === '/discovery/') return '/discovery'
  if (pathname === '/portfolio' || pathname === '/portfolio/') return '/portfolio'
  if (pathname === '/profile' || pathname === '/profile/') return '/profile'
  return pathname === '/about' || pathname === '/about/' ? '/about' : '/'
}

function normalizePath(): AppRoute {
  const route = routeFromPath(window.location.pathname)
  if (window.location.pathname !== route || window.location.search !== '') window.history.replaceState(null, '', route)
  return route
}

function PublicApp({ route, onNavigate, connectHandoff }: { route: PublicRoute; onNavigate: (route: AppRoute, replace?: boolean, connectHandoff?: ConnectHandoff) => void; connectHandoff?: ConnectHandoff }) {
  const headingRef = useRef<HTMLHeadingElement>(null)
  const mounted = useRef(false)
  const authorizationRequest = useRef<AbortController | null>(null)
  const authorizationRequestID = useRef(0)
  const { messages } = useI18n()

  useEffect(() => {
    if (route === '/__status') {
      document.title = messages.status.metaTitle
      return
    }
    const metadata = route === '/about' ? messages.meta.about : route === '/connect' ? messages.meta.connect : messages.meta.home
    document.title = metadata.title
    document.querySelector<HTMLMetaElement>('meta[name="description"]')?.setAttribute('content', metadata.description)
  }, [messages, route])

  useEffect(() => {
    if (mounted.current) headingRef.current?.focus()
    else mounted.current = true
  }, [route])

  useEffect(() => () => authorizationRequest.current?.abort(), [route])

  if (route === '/__status') return <StatusPage />
  const navigatePublic = (next: '/' | '/about' | '/connect') => {
    if (next !== '/connect') {
      authorizationRequest.current?.abort()
      authorizationRequest.current = null
      onNavigate(next)
      return
    }
    authorizationRequest.current?.abort()
    const controller = new AbortController()
    const requestID = ++authorizationRequestID.current
    authorizationRequest.current = controller
    void getAuthorizationStatus(controller.signal)
      .then((status) => {
        if (controller.signal.aborted || requestID !== authorizationRequestID.current) return
        authorizationRequest.current = null
        if (status.authenticated) {
          onNavigate('/onboarding/accounts', true)
          return
        }
        onNavigate('/connect', false, { id: requestID, status })
      })
      .catch(() => {
        if (controller.signal.aborted || requestID !== authorizationRequestID.current) return
        authorizationRequest.current = null
        onNavigate('/connect')
      })
  }
  return (
    <PublicLayout route={route} onNavigate={navigatePublic}>
      {route === '/about' ? <AboutPage headingRef={headingRef} onNavigate={navigatePublic} />
        : route === '/connect' ? <ConsentPage key={connectHandoff?.id ?? 'unverified'} headingRef={headingRef} onNavigate={navigatePublic} onAuthenticated={() => onNavigate('/onboarding/accounts', true)} initialStatus={connectHandoff?.status} />
          : <LandingPage headingRef={headingRef} onNavigate={navigatePublic} />}
    </PublicLayout>
  )
}

function ProtectedApp({ requestedRoute, onNavigate }: { requestedRoute: ProtectedRoute | OnboardingRoute | AccountEditRoute; onNavigate: (route: AppRoute, replace?: boolean) => void }) {
  const authorization = useAuthorizationStatus()
  const { messages } = useI18n()
  const headingRef = useRef<HTMLHeadingElement>(null)
  const [loggingOut, setLoggingOut] = useState(false)
  const [logoutFailed, setLogoutFailed] = useState(false)
  const accountSelection = requestedRoute === '/onboarding/accounts' || requestedRoute === '/portfolio/accounts'
  const editingAccounts = requestedRoute === '/portfolio/accounts'
  const route: ProtectedRoute = accountSelection ? '/portfolio' : requestedRoute
  const connectionSetup = requestedRoute === '/onboarding/accounts'
  const onboarding = connectionSetup
  const navigateProtected = useCallback((next: ProtectedRoute) => {
    onNavigate(next)
  }, [onNavigate])
  const reconnect = useCallback(() => onNavigate('/connect'), [onNavigate])
  const recoverSession = useCallback(() => onNavigate('/connect', true), [onNavigate])
  const completeSetup = useCallback(() => {
    onNavigate('/portfolio', true)
  }, [onNavigate])

  useEffect(() => {
    if (authorization.resolving) return
    if (!authorization.status.authenticated) onNavigate('/connect', true)
  }, [authorization, onNavigate, requestedRoute])

  useEffect(() => {
    if (!authorization.resolving && authorization.status.authenticated && route !== '/profile') headingRef.current?.focus()
  }, [authorization.resolving, authorization.status, connectionSetup, requestedRoute, route])

  useEffect(() => {
    const title = accountSelection ? messages.authenticated.inventory.inclusion.setupTitle : messages.authenticated[route.slice(1) as 'discovery' | 'portfolio' | 'profile']
    document.title = `${title} — Findur`
    document.querySelector<HTMLMetaElement>('meta[name="description"]')?.setAttribute('content', messages.authenticated.metaDescription)
  }, [accountSelection, messages, route])

  if (authorization.resolving) return <main className="session-gate" role="status"><p>{messages.authenticated.checking}</p></main>
  if (!authorization.status.authenticated) return <main className="session-gate" role="status"><p>{messages.authenticated.recovering}</p></main>

  const logout = async () => {
    setLoggingOut(true)
    setLogoutFailed(false)
    try {
      await endCurrentSession()
      resetInitialInventoryRequest()
      onNavigate('/', true)
    } catch {
      setLoggingOut(false)
      setLogoutFailed(true)
    }
  }

  return (
    <AuthenticatedPreferences onSessionExpired={recoverSession}><AuthenticatedLayout route={route} setup={onboarding} loggingOut={loggingOut} logoutFailed={logoutFailed} onNavigate={navigateProtected} onLogout={() => { void logout() }}>
      {connectionSetup ? <OnboardingAccountSelection headingRef={headingRef} onComplete={completeSetup} onReconnect={reconnect} onSessionExpired={recoverSession} /> : accountSelection ? <PortfolioPage editing={editingAccounts} headingRef={headingRef} onComplete={completeSetup} onReconnect={reconnect} onSessionExpired={recoverSession} /> : route === '/portfolio' ? <PortfolioShowcasePage headingRef={headingRef} onEdit={() => onNavigate('/portfolio/accounts')} onReconnect={reconnect} onSessionExpired={recoverSession} /> : route === '/profile' ? <ProfilePage headingRef={headingRef} onSessionExpired={recoverSession} /> : <section className="private-placeholder">
        <h1 ref={headingRef} tabIndex={-1}>{messages.authenticated[`${route.slice(1)}Title` as 'discoveryTitle' | 'portfolioTitle' | 'profileTitle']}</h1>
        <p className="large-copy">{messages.authenticated[`${route.slice(1)}Body` as 'discoveryBody' | 'portfolioBody' | 'profileBody']}</p>
      </section>}
    </AuthenticatedLayout></AuthenticatedPreferences>
  )
}

function OnboardingAccountSelection({ headingRef, onComplete, onReconnect, onSessionExpired }: { headingRef: RefObject<HTMLHeadingElement | null>; onComplete: () => void; onReconnect: () => void; onSessionExpired: () => void }) {
  const { messages } = useI18n()
  const [checking, setChecking] = useState(true)
  const [inclusion, setInclusion] = useState<PortfolioInclusion | null>(null)

  useEffect(() => {
    let active = true
    void getPortfolioInclusion().then((inclusion) => {
      if (!active) return
      if (inclusion.committed.length > 0) onComplete()
      else { setInclusion(inclusion); setChecking(false) }
    }).catch((error) => {
      if (!active) return
      if (error instanceof InventorySessionExpiredError) onSessionExpired()
      else setChecking(false)
    })
    return () => { active = false }
  }, [onComplete, onSessionExpired])

  useLayoutEffect(() => {
    if (!checking) headingRef.current?.focus()
  }, [checking, headingRef])

  if (checking) return <main className="session-gate" role="status"><p>{messages.authenticated.checking}</p></main>
  return <PortfolioPage initialInclusion={inclusion ?? undefined} headingRef={headingRef} onComplete={onComplete} onReconnect={onReconnect} onSessionExpired={onSessionExpired} />
}

function RoutedApp() {
  const [{ route, connectHandoff }, setRoute] = useState<{ route: AppRoute; connectHandoff?: ConnectHandoff }>(() => ({ route: normalizePath() }))
  useEffect(() => {
    const onPopState = () => setRoute({ route: normalizePath() })
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])
  const navigate = useCallback((next: AppRoute, replace = false, handoff?: ConnectHandoff) => {
    if (next === route && !handoff) return
    window.history[replace ? 'replaceState' : 'pushState'](null, '', next)
    setRoute({ route: next, connectHandoff: handoff })
  }, [route])
  return route === '/onboarding/accounts' || route === '/portfolio/accounts' || route === '/discovery' || route === '/portfolio' || route === '/profile'
    ? <ProtectedApp requestedRoute={route} onNavigate={navigate} />
    : <PublicApp route={route} onNavigate={navigate} connectHandoff={connectHandoff} />
}

export function App() {
  return <I18nProvider><ThemeProvider><RoutedApp /></ThemeProvider></I18nProvider>
}
