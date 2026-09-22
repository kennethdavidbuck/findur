import { expect, it, vi } from 'vitest'
import { preferenceStorage } from './storage'

it('keeps page-lifetime preferences when browser storage is unavailable', () => {
  const key = 'storage-fallback-test'
  vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('blocked') })
  vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked') })

  preferenceStorage.set(key, 'dark')

  expect(preferenceStorage.get(key)).toBe('dark')
})
