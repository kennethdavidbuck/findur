import { useCallback, useEffect, useRef, useState } from 'react'
import { useAuthorizationStatus } from './auth-status'
import { AuthenticatedLayout, type ProtectedRoute } from './components/AuthenticatedLayout'
import { PublicLayout } from './components/PublicLayout'
import { I18nProvider, useI18n } from './i18n'
import { AboutPage } from './pages/AboutPage'
import { ConsentPage } from './pages/ConsentPage'
import { LandingPage } from './pages/LandingPage'
import { PortfolioPage } from './pages/PortfolioPage'
import { StatusPage } from './pages/StatusPage'
import { resetInitialInventoryRequest } from './inventory'
import { endCurrentSession } from './session'
import { ThemeProvider } from './theme'

type PublicRoute = '/' | '/about' | '/connect' | '/__status'
type AppRoute = PublicRoute | ProtectedRoute | '/connect/result'

function routeFromPath(pathname: string): AppRoute {
  if (pathname === '/__status' || pathname === '/__status/') return '/__status'
  if (pathname === '/connect' || pathname === '/connect/') return '/connect'
  if (pathname === '/connect/result' || pathname === '/connect/result/') return '/connect/result'
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

function PublicApp({ route, onNavigate }: { route: PublicRoute; onNavigate: (route: AppRoute, replace?: boolean) => void }) {
  const headingRef = useRef<HTMLHeadingElement>(null)
  const mounted = useRef(false)
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

  if (route === '/__status') return <StatusPage />
  const navigatePublic = (next: '/' | '/about' | '/connect') => onNavigate(next)
  return (
    <PublicLayout route={route} onNavigate={navigatePublic}>
      {route === '/about' ? <AboutPage headingRef={headingRef} onNavigate={navigatePublic} />
        : route === '/connect' ? <ConsentPage headingRef={headingRef} onNavigate={navigatePublic} />
          : <LandingPage headingRef={headingRef} onNavigate={navigatePublic} />}
    </PublicLayout>
  )
}

function ProtectedApp({ requestedRoute, onNavigate }: { requestedRoute: ProtectedRoute | '/connect/result'; onNavigate: (route: AppRoute, replace?: boolean) => void }) {
  const authorization = useAuthorizationStatus()
  const { messages } = useI18n()
  const headingRef = useRef<HTMLHeadingElement>(null)
  const [loggingOut, setLoggingOut] = useState(false)
  const [logoutFailed, setLogoutFailed] = useState(false)
  const route: ProtectedRoute = requestedRoute === '/connect/result' ? '/portfolio' : requestedRoute
  const reconnect = useCallback(() => onNavigate('/connect'), [onNavigate])
  const recoverSession = useCallback(() => onNavigate('/connect', true), [onNavigate])

  useEffect(() => {
    if (authorization.resolving) return
    if (!authorization.status.authenticated) onNavigate('/connect', true)
    else if (requestedRoute === '/connect/result') onNavigate('/portfolio', true)
  }, [authorization, onNavigate, requestedRoute])

  useEffect(() => {
    if (!authorization.resolving && authorization.status.authenticated) headingRef.current?.focus()
  }, [authorization.resolving, authorization.status, route])

  useEffect(() => {
    const title = messages.authenticated[route.slice(1) as 'discovery' | 'portfolio' | 'profile']
    document.title = `${title} — Findur`
    document.querySelector<HTMLMetaElement>('meta[name="description"]')?.setAttribute('content', messages.authenticated.metaDescription)
  }, [messages, route])

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
    <AuthenticatedLayout route={route} loggingOut={loggingOut} logoutFailed={logoutFailed} onNavigate={onNavigate} onLogout={() => { void logout() }}>
      {route === '/portfolio' ? <PortfolioPage headingRef={headingRef} onReconnect={reconnect} onSessionExpired={recoverSession} /> : <section className="private-placeholder">
        <p className="eyebrow">{messages.authenticated.privateEyebrow}</p>
        <h1 ref={headingRef} tabIndex={-1}>{messages.authenticated[`${route.slice(1)}Title` as 'discoveryTitle' | 'portfolioTitle' | 'profileTitle']}</h1>
        <p className="large-copy">{messages.authenticated[`${route.slice(1)}Body` as 'discoveryBody' | 'portfolioBody' | 'profileBody']}</p>
      </section>}
    </AuthenticatedLayout>
  )
}

function RoutedApp() {
  const [route, setRoute] = useState<AppRoute>(normalizePath)
  useEffect(() => {
    const onPopState = () => setRoute(normalizePath())
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])
  const navigate = (next: AppRoute, replace = false) => {
    if (next === route) return
    window.history[replace ? 'replaceState' : 'pushState'](null, '', next)
    setRoute(next)
  }
  return route === '/connect/result' || route === '/discovery' || route === '/portfolio' || route === '/profile'
    ? <ProtectedApp requestedRoute={route} onNavigate={navigate} />
    : <PublicApp route={route} onNavigate={navigate} />
}

export function App() {
  return <I18nProvider><ThemeProvider><RoutedApp /></ThemeProvider></I18nProvider>
}
