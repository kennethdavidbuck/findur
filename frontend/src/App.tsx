import { useEffect, useRef, useState } from 'react'
import { PublicLayout } from './components/PublicLayout'
import { I18nProvider, useI18n } from './i18n'
import { AboutPage } from './pages/AboutPage'
import { LandingPage } from './pages/LandingPage'
import { StatusPage } from './pages/StatusPage'
import { ThemeProvider } from './theme'

type PublicRoute = '/' | '/about' | '/__status'

function routeFromPath(pathname: string): PublicRoute {
  if (pathname === '/__status' || pathname === '/__status/') return '/__status'
  return pathname === '/about' || pathname === '/about/' ? '/about' : '/'
}

function normalizePublicPath() {
  const route = routeFromPath(window.location.pathname)
  const canonicalPath = route
  if (window.location.pathname !== canonicalPath) {
    window.history.replaceState(null, '', canonicalPath)
  }
  return route
}

function PublicApp() {
  const [route, setRoute] = useState<PublicRoute>(normalizePublicPath)
  const headingRef = useRef<HTMLHeadingElement>(null)
  const mounted = useRef(false)
  const { messages } = useI18n()

  useEffect(() => {
    const handlePopState = () => setRoute(normalizePublicPath())
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  useEffect(() => {
    if (route === '/__status') {
      document.title = messages.status.metaTitle
      return
    }
    const metadata = route === '/about' ? messages.meta.about : messages.meta.home
    document.title = metadata.title
    const description = document.querySelector<HTMLMetaElement>('meta[name="description"]')
    description?.setAttribute('content', metadata.description)
  }, [messages, route])

  useEffect(() => {
    if (mounted.current) {
      headingRef.current?.focus()
    } else {
      mounted.current = true
    }
  }, [route])

  const navigate = (nextRoute: PublicRoute) => {
    if (nextRoute === route) return
    window.history.pushState(null, '', nextRoute)
    setRoute(nextRoute)
  }

  const navigatePublic = (nextRoute: '/' | '/about') => navigate(nextRoute)

  return (
    route === '/__status' ? <StatusPage /> :
    <PublicLayout route={route} onNavigate={navigatePublic}>
      {route === '/about' ? (
        <AboutPage headingRef={headingRef} onNavigate={navigatePublic} />
      ) : (
        <LandingPage headingRef={headingRef} onNavigate={navigatePublic} />
      )}
    </PublicLayout>
  )
}

export function App() {
  return (
    <I18nProvider>
      <ThemeProvider>
        <PublicApp />
      </ThemeProvider>
    </I18nProvider>
  )
}
