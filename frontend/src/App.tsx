import { useEffect, useRef, useState } from 'react'
import { PublicLayout } from './components/PublicLayout'
import { I18nProvider, useI18n } from './i18n'
import { AboutPage } from './pages/AboutPage'
import { LandingPage } from './pages/LandingPage'
import { ThemeProvider } from './theme'

type PublicRoute = '/' | '/about'

function routeFromPath(pathname: string): PublicRoute {
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

  return (
    <PublicLayout route={route} onNavigate={navigate}>
      {route === '/about' ? (
        <AboutPage headingRef={headingRef} onNavigate={navigate} />
      ) : (
        <LandingPage headingRef={headingRef} onNavigate={navigate} />
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
