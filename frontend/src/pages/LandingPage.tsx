import type { RefObject } from 'react'
import { PublicLink } from '../components/PublicLayout'
import { useI18n } from '../i18n'

type LandingPageProps = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onNavigate: (route: '/' | '/about') => void
}

export function LandingPage({ headingRef, onNavigate }: LandingPageProps) {
  const { messages } = useI18n()
  const content = messages.home

  return (
    <>
      <section className="hero section-pad">
        <div className="container hero-grid">
          <div className="hero-copy">
            <p className="eyebrow">{content.eyebrow}</p>
            <h1 ref={headingRef} tabIndex={-1}>
              {content.titleBefore} <strong>{content.titleAccent}</strong>
            </h1>
            <p className="hero-intro">{content.intro}</p>
            <div className="hero-actions">
              <PublicLink className="action action--primary" href="/about" onNavigate={onNavigate}>
                {content.aboutAction} <span aria-hidden="true">→</span>
              </PublicLink>
              <span className="demo-marker"><span aria-hidden="true">◇</span> {content.demoLabel}</span>
            </div>
          </div>

          <figure className="constellation" aria-labelledby="constellation-caption">
            <p className="visually-hidden">{content.visualSummary}</p>
            <div className="constellation-field" aria-hidden="true">
              <span className="orbit orbit--one" />
              <span className="orbit orbit--two" />
              <span className="connector connector--one" />
              <span className="connector connector--two" />
              <span className="connector connector--three" />
              <span className="node node--foundation">{content.nodes.foundation}</span>
              <span className="node node--curiosity">{content.nodes.curiosity}</span>
              <span className="node node--patience">{content.nodes.patience}</span>
              <span className="node node--perspective">{content.nodes.perspective}</span>
            </div>
            <figcaption id="constellation-caption">
              <strong>{content.visualCaption}</strong>
              <span>{content.visualNote}</span>
            </figcaption>
          </figure>
        </div>
      </section>

      <section className="process section-pad" aria-labelledby="process-heading">
        <div className="container">
          <div className="section-heading">
            <p className="eyebrow">{content.howEyebrow}</p>
            <h2 id="process-heading">{content.howTitle}</h2>
          </div>
          <ol className="step-grid">
            {content.steps.map((step) => (
              <li key={step.number}>
                <span className="step-number">{step.number}</span>
                <h3>{step.title}</h3>
                <p>{step.body}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      <section className="boundary section-pad" aria-labelledby="boundary-heading">
        <div className="container boundary-grid">
          <p className="eyebrow">{content.boundaryEyebrow}</p>
          <div>
            <h2 id="boundary-heading">{content.boundaryTitle}</h2>
            <p>{content.boundaryBody}</p>
            <PublicLink className="text-link" href="/about" onNavigate={onNavigate}>
              {content.boundaryLink} <span aria-hidden="true">→</span>
            </PublicLink>
          </div>
        </div>
      </section>
    </>
  )
}
