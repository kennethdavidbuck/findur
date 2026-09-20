import { useEffect, type RefObject } from 'react'
import { useI18n } from '../i18n'
import { PublicLink } from '../components/PublicLayout'
import { useAuthorizationStatus } from '../auth-status'

type Props = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about' | '/connect') => void
}

export function AuthorizationResultPage({ headingRef, onNavigate }: Props) {
  const { messages } = useI18n()
	const authorization = useAuthorizationStatus()
	const result = authorization.resolving ? messages.authorizationResult.resolving : authorization.status.authenticated ? messages.authorizationResult.success : messages.authorizationResult.restart
	useEffect(() => { if (!authorization.resolving) headingRef.current?.focus() }, [authorization.resolving, headingRef])
  return (
    <section className="consent-page section-pad">
      <div className="container consent-panel" role="status">
        <p className="eyebrow">{messages.authorizationResult.eyebrow}</p>
        <h1 ref={headingRef} tabIndex={-1}>{result.title}</h1>
        <p className="large-copy">{result.body}</p>
		{!authorization.resolving && <PublicLink className="action action--primary" href={result.href} onNavigate={onNavigate}>{result.action}</PublicLink>}
      </div>
    </section>
  )
}
