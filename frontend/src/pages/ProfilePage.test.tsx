import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { App } from '../App'

const locations = [{ key: 'halifax-ns', cityEn: 'Halifax', cityFr: 'Halifax', provinceEn: 'Nova Scotia', provinceFr: 'Nouvelle-Écosse' }]

beforeEach(() => {
  window.history.replaceState(null, '', '/profile')
  window.localStorage.clear()
  Object.defineProperty(window, 'matchMedia', { configurable: true, value: vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })) })
})

afterEach(() => { cleanup(); vi.unstubAllGlobals() })

it('loads a blank localized profile and saves the complete draft without changing focus', async () => {
  let resolveSave!: (response: Response) => void
  const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/preferences/display') return Promise.resolve(new Response(null, { status: 204 }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations }))
    if (path === '/api/profile' && init?.method === 'PUT') return new Promise<Response>((resolve) => { resolveSave = resolve })
    return Promise.resolve(new Response(null, { status: 404 }))
  })
  vi.stubGlobal('fetch', fetchMock)
  render(<App />)
  const name = await screen.findByLabelText('Display name')
  await waitFor(() => expect(screen.getByRole('heading', { name: 'Make the profile yours.' })).toHaveFocus())
  fireEvent.change(name, { target: { value: '  Alex  ' } })
  fireEvent.click(screen.getByLabelText('I confirm that I am 18 or older'))
  fireEvent.change(screen.getByLabelText('City'), { target: { value: 'halifax-ns' } })
  const biography = screen.getByLabelText('Short biography')
  fireEvent.change(biography, { target: { value: '  Hello there  ' } })
  const save = screen.getByRole('button', { name: 'Save profile' }); save.focus(); fireEvent.click(save)
  await waitFor(() => {
    expect(name).toBeDisabled()
    expect(biography).toBeDisabled()
    expect(save).toBeDisabled()
  })
  resolveSave(json({ displayName: 'Alex', adultAttestedAt: '2026-09-21T12:00:00Z', locationKey: 'halifax-ns', relationshipIntent: 'long-term', biography: 'Hello there', avatarKey: 'aurora', locale: 'en', theme: 'system', version: 1 }))
  expect(await screen.findByText('Profile saved.')).toBeVisible()
  expect(save).toHaveFocus()
  expect(name).toHaveValue('Alex')
  expect(biography).toHaveValue('Hello there')
  const request = fetchMock.mock.calls.find(([path, init]) => path === '/api/profile' && init?.method === 'PUT')
  expect(JSON.parse(String(request?.[1]?.body))).toMatchObject({ displayName: '  Alex  ', adultAttested: true, expectedVersion: 0 })
})

it('counts and validates Unicode code points instead of UTF-16 code units', async () => {
  const sixtyEmoji = '😀'.repeat(60)
  const fiveHundredEmoji = '😀'.repeat(500)
  const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations }))
    if (path === '/api/profile' && init?.method === 'PUT') return Promise.resolve(json({ displayName: sixtyEmoji, adultAttestedAt: '2026-09-21T12:00:00Z', locationKey: 'halifax-ns', relationshipIntent: 'long-term', biography: fiveHundredEmoji, avatarKey: 'aurora', locale: 'en', theme: 'system', version: 1 }))
    return Promise.resolve(new Response(null, { status: 404 }))
  })
  vi.stubGlobal('fetch', fetchMock)
  render(<App />)
  const name = await screen.findByLabelText('Display name')
  const biography = screen.getByLabelText('Short biography')
  expect(name).not.toHaveAttribute('maxlength')
  expect(biography).not.toHaveAttribute('maxlength')
  fireEvent.change(name, { target: { value: sixtyEmoji } })
  fireEvent.click(screen.getByLabelText('I confirm that I am 18 or older'))
  fireEvent.change(screen.getByLabelText('City'), { target: { value: 'halifax-ns' } })
  fireEvent.change(biography, { target: { value: fiveHundredEmoji } })
  expect(screen.getByText('500/500')).toBeVisible()
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  expect(await screen.findByText('Profile saved.')).toBeVisible()
  expect(fetchMock.mock.calls.filter(([path, init]) => path === '/api/profile' && init?.method === 'PUT')).toHaveLength(1)
})

it('focuses linked client errors before sending an incomplete first profile', async () => {
  const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations }))
    return Promise.resolve(new Response(null, { status: 404 }))
  })
  vi.stubGlobal('fetch', fetchMock)
  render(<App />)
  await screen.findByLabelText('Display name')
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  const summary = await screen.findByRole('alert')
  await waitFor(() => expect(summary).toHaveFocus())
  expect(summary).toHaveTextContent('Display name')
  expect(summary).toHaveTextContent('I confirm that I am 18 or older')
  expect(summary).toHaveTextContent('City')
  expect(summary).toHaveTextContent('Short biography')
  expect(screen.getByRole('link', { name: 'Display name' })).toHaveAttribute('href', '#displayName')
  expect(fetchMock.mock.calls.filter(([path, init]) => path === '/api/profile' && init?.method === 'PUT')).toHaveLength(0)
})

it('synchronizes a saved French and dark profile with the shared providers', async () => {
  vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/preferences/display') return Promise.resolve(json({ locale: 'fr', theme: 'dark', version: 2 }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations, profile: { displayName: 'Alex', adultAttestedAt: '2026-09-21T12:00:00Z', locationKey: 'halifax-ns', relationshipIntent: 'open-to-long-term', biography: 'Bonjour', avatarKey: 'cedar', locale: 'fr', theme: 'dark', version: 2 } }))
    return Promise.resolve(new Response(null, { status: 404 }))
  }))
  render(<App />)
  expect(await screen.findByRole('heading', { name: 'Créez un profil à votre image.' })).toBeVisible()
  await waitFor(() => {
    expect(document.documentElement).toHaveAttribute('lang', 'fr')
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark')
  })
  expect(screen.getByRole('radio', { name: 'Français' })).toBeChecked()
  for (const radio of screen.getAllByRole('radio', { name: 'Sombre' })) expect(radio).toBeChecked()
  expect(screen.getByRole('group', { name: /Langue/ })).toHaveTextContent('Privé')
  expect(screen.getByRole('group', { name: /Portrait/ })).toHaveTextContent('Après un match mutuel')
})

it('keeps the draft and offers recovery when a version conflict occurs', async () => {
  let profileLoads = 0
  let conflicted = false
  const fetchMock = vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/profile' && init?.method === 'GET') {
      profileLoads += 1
      return Promise.resolve(json(profileLoads === 1 ? { locations } : { locations, profile: { displayName: 'Saved elsewhere', adultAttestedAt: '2026-09-21T12:00:00Z', locationKey: 'halifax-ns', relationshipIntent: 'long-term', biography: 'Server copy', avatarKey: 'aurora', locale: 'en', theme: 'system', version: 3 } }))
    }
    if (path === '/api/profile' && init?.method === 'PUT') {
      if (!conflicted) {
        conflicted = true
        return Promise.resolve(new Response(JSON.stringify({ code: 'conflict' }), { status: 409 }))
      }
      return Promise.resolve(json({ displayName: 'Still here', adultAttestedAt: '2026-09-21T12:00:00Z', locationKey: 'halifax-ns', relationshipIntent: 'long-term', biography: 'Still recoverable', avatarKey: 'aurora', locale: 'en', theme: 'system', version: 4 }))
    }
    return Promise.resolve(new Response(null, { status: 404 }))
  })
  vi.stubGlobal('fetch', fetchMock)
  render(<App />)
  const name = await screen.findByLabelText('Display name')
  fireEvent.change(name, { target: { value: 'Still here' } })
  fireEvent.click(screen.getByLabelText('I confirm that I am 18 or older'))
  fireEvent.change(screen.getByLabelText('City'), { target: { value: 'halifax-ns' } })
  fireEvent.change(screen.getByLabelText('Short biography'), { target: { value: 'Still recoverable' } })
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  const recovery = await screen.findByRole('button', { name: 'Reload latest and reapply my entries' })
  await waitFor(() => expect(recovery).toHaveFocus())
  expect(name).toHaveValue('Still here')
  fireEvent.click(recovery)
  expect(await screen.findByText(/entries were reapplied/)).toBeVisible()
  expect(name).toHaveValue('Still here')
  const save = screen.getByRole('button', { name: 'Save profile' })
  await waitFor(() => expect(save).toHaveFocus())
  fireEvent.click(save)
  expect(await screen.findByText('Profile saved.')).toBeVisible()
  const putRequests = fetchMock.mock.calls.filter(([path, init]) => path === '/api/profile' && init?.method === 'PUT')
  expect(JSON.parse(String(putRequests[1]?.[1]?.body))).toMatchObject({ displayName: 'Still here', expectedVersion: 3 })
})

it('links server validation errors, focuses the summary, and preserves the draft', async () => {
  vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations }))
    if (path === '/api/profile' && init?.method === 'PUT') return Promise.resolve(new Response(JSON.stringify({ code: 'invalid_profile', fields: ['displayName', 'biography'] }), { status: 400, headers: { 'Content-Type': 'application/json' } }))
    return Promise.resolve(new Response(null, { status: 404 }))
  }))
  render(<App />)
  const name = await screen.findByLabelText('Display name')
  const biography = screen.getByLabelText('Short biography')
  fireEvent.change(name, { target: { value: 'Still Alex' } })
  fireEvent.change(biography, { target: { value: 'Still a recoverable biography' } })
  fireEvent.click(screen.getByLabelText('I confirm that I am 18 or older'))
  fireEvent.change(screen.getByLabelText('City'), { target: { value: 'halifax-ns' } })
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  const summary = await screen.findByRole('alert')
  await waitFor(() => expect(summary).toHaveFocus())
  expect(summary).toHaveTextContent('Display name')
  expect(summary).toHaveTextContent('Short biography')
  expect(name).toHaveValue('Still Alex')
  expect(biography).toHaveValue('Still a recoverable biography')
  expect(screen.getAllByText('Review this field.')).toHaveLength(2)
})

it('keeps the draft after an unavailable save and returns to connection on session expiry', async () => {
  let saveStatus = 503
  vi.stubGlobal('fetch', vi.fn().mockImplementation((path: string, init?: RequestInit) => {
    if (path === '/api/auth/status') return Promise.resolve(json({ authorizationAvailable: true, authenticated: true, reauthorizationRequired: false }))
    if (path === '/api/profile' && init?.method === 'GET') return Promise.resolve(json({ locations }))
    if (path === '/api/profile' && init?.method === 'PUT') return Promise.resolve(new Response(null, { status: saveStatus }))
    return Promise.resolve(new Response(null, { status: 404 }))
  }))
  render(<App />)
  const name = await screen.findByLabelText('Display name')
  fireEvent.change(name, { target: { value: 'Kept through failure' } })
  fireEvent.click(screen.getByLabelText('I confirm that I am 18 or older'))
  fireEvent.change(screen.getByLabelText('City'), { target: { value: 'halifax-ns' } })
  fireEvent.change(screen.getByLabelText('Short biography'), { target: { value: 'A complete profile that stays put.' } })
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('entries are still here')
  expect(name).toHaveValue('Kept through failure')
  saveStatus = 401
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }))
  await waitFor(() => expect(window.location.pathname).toBe('/connect'))
})

function json(body: unknown) { return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }) }
