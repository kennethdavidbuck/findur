import type { RefObject } from 'react'
import { PublicLink } from '../components/PublicLayout'
import { useI18n } from '../i18n'

type AboutPageProps = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about' | '/connect') => void
}

export function AboutPage({ headingRef, onNavigate }: AboutPageProps) {
  const { messages } = useI18n()
  const content = messages.about

  return (
    <>
      <section className="about-hero section-pad">
        <div className="container reading-width">
          <p className="eyebrow">{content.eyebrow}</p>
          <h1 ref={headingRef} tabIndex={-1}>
            {content.titleBefore} <strong>{content.titleAccent}</strong>
          </h1>
          <p className="hero-intro">{content.intro}</p>
        </div>
      </section>

      <section className="about-thesis section-pad" aria-labelledby="thesis-heading">
        <div className="container split-section">
          <p className="eyebrow">{content.thesisLabel}</p>
          <div className="reading-column">
            <h2 id="thesis-heading">{content.thesisTitle}</h2>
            <p className="large-copy">{content.thesisBody}</p>
          </div>
        </div>
      </section>

      <section className="principles section-pad" aria-labelledby="principles-heading">
        <div className="container">
          <div className="section-heading">
            <p className="eyebrow">{content.principlesEyebrow}</p>
            <h2 id="principles-heading">{content.principlesTitle}</h2>
          </div>
          <div className="principle-grid">
            {content.principles.map((principle, index) => (
              <article key={principle.title}>
                <span className="step-number">0{index + 1}</span>
                <h3>{principle.title}</h3>
                <p>{principle.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="status-panel section-pad" aria-labelledby="status-heading">
        <div className="container status-inner">
          <div>
            <p className="eyebrow">{content.statusEyebrow}</p>
            <h2 id="status-heading">{content.statusTitle}</h2>
          </div>
          <div className="reading-column">
            <p>{content.statusBody}</p>
            <PublicLink className="action action--secondary" href="/" onNavigate={onNavigate}>
              <span aria-hidden="true">←</span> {content.backHome}
            </PublicLink>
          </div>
        </div>
      </section>
    </>
  )
}
