import type { RefObject } from 'react'
import { Button } from 'react-aria-components'
import { useI18n } from '../i18n'
import { PublicLink } from '../components/PublicLayout'
import { useAuthorizationStatus } from '../auth-status'

type ConsentPageProps = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about' | '/connect') => void
}

export function ConsentPage({ headingRef, onNavigate }: ConsentPageProps) {
  const { messages } = useI18n()
  const consent = messages.consent
	const authorization = useAuthorizationStatus()
	const available = !authorization.resolving && authorization.status.authorizationAvailable

  return (
    <section className="consent-page section-pad">
      <div className="container consent-panel">
        <p className="eyebrow">{consent.eyebrow}</p>
        <h1 ref={headingRef} tabIndex={-1}>{consent.title}</h1>
        <p className="large-copy">{consent.intro}</p>

        <p className="consent-reassurance">{consent.reassurance}</p>
        <div className="consent-actions">
          <form action="/api/auth/snaptrade/authorize" method="post" aria-describedby={available ? undefined : "authorization-unavailable"}>
            <input type="hidden" name="returnTo" value="/portfolio" />
            <Button className="action action--primary" type="submit" isDisabled={!available} aria-describedby={available ? undefined : "authorization-unavailable"}>{consent.action}</Button>
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
