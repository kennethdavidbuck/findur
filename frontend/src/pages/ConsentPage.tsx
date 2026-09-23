import { useEffect, useRef, type RefObject } from 'react'
import { Button } from 'react-aria-components'
import { useI18n } from '../i18n'
import { PublicLink } from '../components/PublicLayout'
import { useAuthorizationStatus, type AuthorizationStatus } from '../auth-status'
import type { AuthorizationNotice } from '../authorization-notice'

type ConsentPageProps = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about' | '/connect') => void
  onAuthenticated: () => void
  initialStatus?: AuthorizationStatus
  authorizationNotice?: AuthorizationNotice
}

export function ConsentPage({ headingRef, onNavigate, onAuthenticated, initialStatus, authorizationNotice }: ConsentPageProps) {
  const { messages } = useI18n()
  const consent = messages.consent
	const authorization = useAuthorizationStatus(initialStatus)
	const recoveryRef = useRef<HTMLDivElement>(null)
	const available = !authorization.resolving && authorization.status.authorizationAvailable
	const recovery = !authorization.resolving && authorization.status.reauthorizationRequired ? 'reauthorization_required' : authorizationNotice
	const blockingRecovery = recovery !== undefined && recovery !== 'denied'

  useEffect(() => {
    if (!authorization.resolving && authorization.status.authenticated && !recovery) onAuthenticated()
  }, [authorization, onAuthenticated, recovery])

  useEffect(() => {
    if (authorization.resolving) return
    if (recovery) recoveryRef.current?.focus()
    else if (!authorization.status.authenticated) headingRef.current?.focus()
  }, [authorization, headingRef, recovery])

  if (authorization.resolving || authorization.status.authenticated && !recovery) {
    return <section className="consent-page consent-page--checking section-pad" aria-busy="true">
      <div className="container consent-panel" role="status"><p>{consent.checking}</p></div>
    </section>
  }

  return (
    <section className="consent-page section-pad">
      <div className="container consent-panel">
        <p className="eyebrow">{consent.eyebrow}</p>
        <h1 ref={headingRef} tabIndex={-1}>{consent.title}</h1>
        <p className="large-copy">{consent.intro}</p>

        {recovery && <div className={`consent-recovery${blockingRecovery ? ' consent-recovery--error' : ''}`} id="authorization-recovery" role="alert" tabIndex={-1} ref={recoveryRef}>
          <h2>{consent.recovery[recovery].title}</h2>
          <p>{consent.recovery[recovery].body}</p>
        </div>}

        <p className="consent-reassurance">{consent.reassurance}</p>
        <div className="consent-actions">
          <form action="/api/auth/snaptrade/authorize" method="post" aria-describedby={!available ? "authorization-unavailable" : recovery ? "authorization-recovery" : undefined}>
            <input type="hidden" name="returnTo" value="/portfolio" />
            <Button className="action action--primary" type="submit" isDisabled={!available} aria-describedby={!available ? "authorization-unavailable" : recovery ? "authorization-recovery" : undefined}>{recovery === 'reauthorization_required' ? consent.reconnectAction : recovery ? consent.retryAction : consent.action}</Button>
            {!available && <p className="consent-unavailable" id="authorization-unavailable" role="status">{authorization.resolving ? consent.checking : consent.unavailable}</p>}
          </form>
          <p className="mono-label consent-eligibility">{consent.eligibility}</p>
        </div>
        <section className="consent-summary" aria-labelledby="consent-summary-heading">
          <h2 id="consent-summary-heading">{consent.summaryTitle}</h2>
          <ul>{consent.summary.map((item) => <li key={item}>{item}</li>)}</ul>
        </section>
        <PublicLink className="text-link" href="/" onNavigate={onNavigate}>{consent.back}</PublicLink>
      </div>
    </section>
  )
}
