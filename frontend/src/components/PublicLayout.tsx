import type { MouseEvent, ReactNode } from 'react'
import { Radio, RadioGroup } from 'react-aria-components'
import { type Locale, useI18n } from '../i18n'
import { type ThemePreference, useTheme } from '../theme'

type PublicRoute = '/' | '/about' | '/connect'

type PublicLayoutProps = {
  children: ReactNode
  route: PublicRoute
  onNavigate: (route: PublicRoute) => void
}

type PublicLinkProps = {
  children: ReactNode
  className?: string
  current?: boolean
  href: PublicRoute
  onNavigate: (route: PublicRoute) => void
}

export function PublicLink({ children, className, current, href, onNavigate }: PublicLinkProps) {
  const handleClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onNavigate(href)
  }

  return (
    <a className={className} href={href} aria-current={current ? 'page' : undefined} onClick={handleClick}>
      {children}
    </a>
  )
}

export function PreferenceControls({ compact = false }: { compact?: boolean }) {
  const { locale, messages, setLocale } = useI18n()
  const { preference, setPreference } = useTheme()

  return (
    <div className={`preference-controls${compact ? ' preference-controls--compact' : ''}`}>
      <RadioGroup
        aria-label={messages.languageLabel}
        className="choice-group"
        orientation="horizontal"
        value={locale}
        onChange={(value) => setLocale(value as Locale)}
      >
        <Radio className="choice" value="en">EN</Radio>
        <Radio className="choice" value="fr">FR</Radio>
      </RadioGroup>
      <RadioGroup
        aria-label={messages.themeLabel}
        className="choice-group"
        orientation="horizontal"
        value={preference}
        onChange={(value) => setPreference(value as ThemePreference)}
      >
        <Radio className="choice choice--wide" value="system">{messages.themes.system}</Radio>
        <Radio className="choice" value="light">{messages.themes.light}</Radio>
        <Radio className="choice" value="dark">{messages.themes.dark}</Radio>
      </RadioGroup>
    </div>
  )
}

export function PublicLayout({ children, route, onNavigate }: PublicLayoutProps) {
  const { messages } = useI18n()

  return (
    <div className="site-shell">
      <a className="skip-link" href="#main-content">{messages.skipLink}</a>
      <header className="public-header">
        <div className="container header-inner">
          <PublicLink className="wordmark" href="/" current={route === '/'} onNavigate={onNavigate}>
            find<span>ur</span>
          </PublicLink>
          <nav className="public-nav" aria-label={messages.nav.primary}>
            <PublicLink href="/" current={route === '/'} onNavigate={onNavigate}>
              {messages.nav.home}
            </PublicLink>
            <PublicLink href="/about" current={route === '/about'} onNavigate={onNavigate}>
              {messages.nav.about}
            </PublicLink>
          </nav>
          <div className="owner-entry">
            <PublicLink className="owner-button" href="/connect" current={route === '/connect'} onNavigate={onNavigate}>
              {messages.owner.action}
            </PublicLink>
          </div>
        </div>
      </header>

      <main id="main-content">{children}</main>

      <footer className="public-footer">
        <div className="container footer-grid">
          <div>
            <PublicLink className="wordmark wordmark--footer" href="/" current={route === '/'} onNavigate={onNavigate}>
              find<span>ur</span>
            </PublicLink>
            <p className="footer-statement">{messages.footer.statement}</p>
          </div>
          <div className="footer-preferences">
            <p className="mono-label">{messages.footer.preferences}</p>
            <PreferenceControls compact />
          </div>
        </div>
        <div className="container footer-boundary">{messages.footer.boundary}</div>
      </footer>
    </div>
  )
}
