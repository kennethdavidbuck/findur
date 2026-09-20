import { afterEach, describe, expect, it, vi } from 'vitest'
import { clearProtectedBrowserData, csrfToken, endCurrentSession } from './session'

afterEach(() => {
  window.localStorage.clear()
  window.sessionStorage.clear()
  vi.unstubAllGlobals()
})

describe('session browser cleanup', () => {
  it('returns an empty token for malformed cookie encoding', () => {
    document.cookie = 'findur_csrf=%E0%A4%A; Path=/'
    expect(csrfToken()).toBe('')
  })

  it('treats 401 as ended and removes every non-preference storage key including empty', async () => {
    window.localStorage.setItem('findur-locale', 'fr')
    window.localStorage.setItem('findur-theme', 'dark')
    window.localStorage.setItem('', 'protected')
    window.localStorage.setItem('protected', 'secret')
    window.sessionStorage.setItem('protected', 'secret')
    vi.stubGlobal('indexedDB', undefined)
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 401 })))

    await endCurrentSession()

    expect(window.localStorage.getItem('findur-locale')).toBe('fr')
    expect(window.localStorage.getItem('findur-theme')).toBe('dark')
    expect(window.localStorage.getItem('')).toBeNull()
    expect(window.localStorage.getItem('protected')).toBeNull()
    expect(window.sessionStorage.length).toBe(0)
  })

  it('requires exactly 204 for a normal logout response', async () => {
    window.localStorage.setItem('protected', 'secret')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 200 })))
    await expect(endCurrentSession()).rejects.toThrow('logout failed')
    expect(window.localStorage.getItem('protected')).toBe('secret')
  })

  it('rejects cleanup when Cache Storage does not confirm deletion', async () => {
    vi.stubGlobal('indexedDB', undefined)
    vi.stubGlobal('caches', { keys: vi.fn().mockResolvedValue(['protected']), delete: vi.fn().mockResolvedValue(false) })
    await expect(clearProtectedBrowserData()).rejects.toThrow('cache cleanup failed')
  })

  it('rejects cleanup when IndexedDB deletion fails', async () => {
    const request = { error: null } as unknown as IDBOpenDBRequest
    const database = {
      databases: vi.fn().mockResolvedValue([{ name: 'protected', version: 1 }]),
      deleteDatabase: vi.fn(() => {
        queueMicrotask(() => request.onerror?.call(request, new Event('error')))
        return request
      }),
    }
    vi.stubGlobal('indexedDB', database)
    await expect(clearProtectedBrowserData()).rejects.toThrow('indexedDB cleanup failed')
  })
})
