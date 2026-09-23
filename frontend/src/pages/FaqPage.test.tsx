import { createRef } from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { I18nProvider, useI18n } from '../i18n'
import { FaqPage } from './FaqPage'

function FaqHarness() {
  const { locale, setLocale } = useI18n()
  return <><button onClick={() => setLocale(locale === 'en' ? 'fr' : 'en')}>Switch locale</button><FaqPage headingRef={createRef<HTMLHeadingElement>()} /></>
}

beforeEach(() => {
  window.localStorage.clear()
  document.documentElement.lang = 'en'
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('FaqPage', () => {
  it('uses native headed disclosures to explain eligibility, selection, scheduling, retries, and actions without fetching data', () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    render(<I18nProvider><FaqPage headingRef={createRef<HTMLHeadingElement>()} /></I18nProvider>)

    expect(screen.getByRole('heading', { level: 1, name: 'Portfolio questions, answered.' })).toBeVisible()
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(6)
    const timing = screen.getByText('How often does Findur update my portfolio?').closest('details')
    expect(timing).not.toHaveAttribute('open')
    fireEvent.click(screen.getByText('How often does Findur update my portfolio?'))
    expect(timing).toHaveAttribute('open')
    expect(screen.getByText(/active connection.*open.*investment account.*completed, available holdings sync/)).toBeInTheDocument()
    expect(screen.getByText(/Deposit accounts, lines of credit, closed or unavailable accounts/)).toBeInTheDocument()
    expect(screen.getByText(/cash-only account may qualify/)).toBeInTheDocument()
    expect(screen.getByText(/none is included automatically.*up to five accounts/)).toBeInTheDocument()
    expect(screen.getByText(/about 24 hours later.*not a guaranteed completion time/)).toBeInTheDocument()
    expect(screen.getByText(/newest 50 accumulated activities/)).toBeInTheDocument()
    expect(screen.getByText(/1, 2, 4, 8, 16, and 32 minutes.*later retries remaining 32 minutes apart/)).toBeInTheDocument()
    expect(screen.getByText(/provider asks Findur to wait longer/)).toBeInTheDocument()
    expect(screen.getByText(/“Check again”.*“Try again”.*neither forces the provider/)).toBeInTheDocument()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('provides the same factual guidance in French with expansion-safe native markup', () => {
    window.localStorage.setItem('findur-locale', 'fr')
    render(<I18nProvider><FaqPage headingRef={createRef<HTMLHeadingElement>()} /></I18nProvider>)

    expect(screen.getByRole('heading', { level: 1, name: 'Vos questions sur le portefeuille.' })).toBeVisible()
    expect(screen.getByText(/connexion active.*être ouvert.*compte de placement.*synchronisation des avoirs terminée/)).toBeInTheDocument()
    expect(screen.getByText(/comptes de dépôt, les marges de crédit, les comptes fermés ou indisponibles/)).toBeInTheDocument()
    expect(screen.getByText(/aucun compte n’est inclus automatiquement.*jusqu’à cinq comptes/)).toBeInTheDocument()
    expect(screen.getByText(/environ 24 heures plus tard.*ne garantit donc pas l’heure d’achèvement/)).toBeInTheDocument()
    expect(screen.getByText(/50 activités cumulées les plus récentes/)).toBeInTheDocument()
    expect(screen.getByText(/1, 2, 4, 8, 16 et 32 minutes.*tentatives suivantes restent espacées de 32 minutes/)).toBeInTheDocument()
    expect(screen.getByText(/« Vérifier à nouveau ».*« Réessayer ».*ne force la synchronisation/)).toBeInTheDocument()
    expect(document.querySelectorAll('details')).toHaveLength(6)
  })

  it('keeps an open disclosure mounted across a locale switch', () => {
    render(<I18nProvider><FaqHarness /></I18nProvider>)
    const disclosure = screen.getByRole('heading', { level: 2, name: 'Which accounts can I choose?' }).closest('details') as HTMLDetailsElement
    fireEvent.click(disclosure.querySelector('summary') as HTMLElement)
    expect(disclosure.open).toBe(true)

    fireEvent.click(screen.getByRole('button', { name: 'Switch locale' }))

    expect(screen.getByRole('heading', { level: 2, name: 'Quels comptes puis-je choisir?' }).closest('details')).toBe(disclosure)
    expect(disclosure.open).toBe(true)
  })

  it('exposes focusable disclosure summaries with a visible state indicator', () => {
    render(<I18nProvider><FaqPage headingRef={createRef<HTMLHeadingElement>()} /></I18nProvider>)
    const disclosure = screen.getByRole('heading', { level: 2, name: 'Which accounts can I choose?' }).closest('details') as HTMLDetailsElement
    const summary = disclosure.querySelector('summary') as HTMLElement
    const indicator = summary.querySelector('.faq-disclosure-indicator')

    summary.focus()
    expect(summary).toHaveFocus()
    expect(indicator).toHaveTextContent('+')
    expect(indicator).toHaveTextContent('−')
    fireEvent.click(summary)
    expect(disclosure.open).toBe(true)
  })
})
