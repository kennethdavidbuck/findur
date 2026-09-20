import type { RefObject } from 'react'
import { Button } from 'react-aria-components'
import { useI18n } from '../i18n'
import { PublicLink } from '../components/PublicLayout'

type ConsentPageProps = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about' | '/connect') => void
}

export function ConsentPage({ headingRef, onNavigate }: ConsentPageProps) {
  const { messages } = useI18n()
  const consent = messages.consent

  return (
    <section className="consent-page section-pad">
      <div className="container consent-panel">
        <p className="eyebrow">{consent.eyebrow}</p>
        <h1 ref={headingRef} tabIndex={-1}>{consent.title}</h1>
        <p className="large-copy">{consent.intro}</p>

        <div className="consent-sections">
          <section aria-labelledby="initial-consent-heading">
            <h2 id="initial-consent-heading">{consent.initialTitle}</h2>
            <p>{consent.initialBody}</p>
            <p className="consent-emphasis">{consent.noDefault}</p>
          </section>
          <section aria-labelledby="later-consent-heading">
            <h2 id="later-consent-heading">{consent.laterTitle}</h2>
            <p>{consent.laterBody}</p>
          </section>
          <section aria-labelledby="disclosure-consent-heading">
            <h2 id="disclosure-consent-heading">{consent.disclosureTitle}</h2>
            <p>{consent.disclosureBody}</p>
          </section>
          <section aria-labelledby="limits-consent-heading">
            <h2 id="limits-consent-heading">{consent.limitsTitle}</h2>
            <ul>{consent.limits.map((limit) => <li key={limit}>{limit}</li>)}</ul>
          </section>
        </div>

        <p className="mono-label consent-eligibility">{consent.eligibility}</p>
        <form action="/api/auth/snaptrade/authorize" method="post" aria-describedby="authorization-unavailable">
          <input type="hidden" name="returnTo" value="/portfolio" />
          <Button className="action action--primary" type="submit" isDisabled aria-describedby="authorization-unavailable">{consent.action}</Button>
          <p className="consent-unavailable" id="authorization-unavailable" role="status">{consent.unavailable}</p>
        </form>
        <PublicLink className="text-link" href="/" onNavigate={onNavigate}>{consent.back}</PublicLink>
      </div>
    </section>
  )
}
