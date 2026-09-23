import type { MouseEvent, ReactNode } from 'react'
import { PreferenceControls } from './PublicLayout'
import { useI18n } from '../i18n'
import { useAuthenticatedPreferences } from '../authenticated-preferences'

export type ProtectedRoute = '/discovery' | '/portfolio' | '/profile' | '/faq'

type Props = {
  children: ReactNode
  route: ProtectedRoute
  setup?: boolean
  loggingOut: boolean
  logoutFailed: boolean
  onNavigate: (route: ProtectedRoute) => void
  onLogout: () => void
}

const destinations: Array<{ route: ProtectedRoute; key: 'portfolio' | 'profile' | 'faq' }> = [
  { route: '/portfolio', key: 'portfolio' },
  { route: '/profile', key: 'profile' },
  { route: '/faq', key: 'faq' },
]

function Navigation({ route, onNavigate }: Pick<Props, 'route' | 'onNavigate'>) {
  const { messages } = useI18n()
  const handleClick = (event: MouseEvent<HTMLAnchorElement>, next: ProtectedRoute) => {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(next)
  }
  return (
    <nav className="authenticated-nav authenticated-nav--adaptive" aria-label={messages.authenticated.navigation}>
      {destinations.map((destination) => (
        <a key={destination.route} href={destination.route} aria-current={route === destination.route ? 'page' : undefined} onClick={(event) => handleClick(event, destination.route)}>
          {messages.authenticated[destination.key]}
        </a>
      ))}
    </nav>
  )
}

function PreferenceSaveStatus() {
  const { messages } = useI18n()
  const preferences = useAuthenticatedPreferences()
  if (preferences?.state === 'saving') return <p className="visually-hidden" role="status">{messages.preferencesSaving}</p>
  if (preferences?.state === 'retry') return <p className="visually-hidden" aria-live="polite">{messages.preferencesSaveFailed}</p>
  return null
}

export function AuthenticatedLayout({ children, route, setup = false, loggingOut, logoutFailed, onNavigate, onLogout }: Props) {
  const { messages } = useI18n()
  return (
    <div className={`authenticated-shell${setup ? ' authenticated-shell--setup' : ''}`}>
      <a className="skip-link" href="#private-content">{messages.skipLink}</a>
      <header className="authenticated-header">
        {setup ? <span className="wordmark wordmark--static" aria-label="findur">find<span>ur</span></span>
          : <a className="wordmark" href="/portfolio" onClick={(event) => { event.preventDefault(); onNavigate('/portfolio') }}>find<span>ur</span></a>}
        <PreferenceControls compact />
        <div className="logout-control">
          <button className="action action--secondary" type="button" disabled={loggingOut} onClick={onLogout}>
            {loggingOut ? messages.authenticated.loggingOut : logoutFailed ? messages.authenticated.retryLogout : messages.authenticated.logout}
          </button>
          {logoutFailed && <p role="alert">{messages.authenticated.logoutFailed}</p>}
        </div>
      </header>
      {!setup && <Navigation route={route} onNavigate={onNavigate} />}
      <PreferenceSaveStatus />
      <main id="private-content" className="authenticated-content">{children}</main>
    </div>
  )
}
