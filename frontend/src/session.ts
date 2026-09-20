const retainedLocalStorage = new Set(['findur-locale', 'findur-theme'])

export function csrfToken(): string {
  const prefix = 'findur_csrf='
  const cookie = document.cookie.split(';').map((value) => value.trim()).find((value) => value.startsWith(prefix))
  if (!cookie) return ''
  try {
    return decodeURIComponent(cookie.slice(prefix.length))
  } catch {
    return ''
  }
}

export async function clearProtectedBrowserData(): Promise<void> {
  for (let index = window.localStorage.length - 1; index >= 0; index -= 1) {
    const key = window.localStorage.key(index)
    if (key !== null && !retainedLocalStorage.has(key)) window.localStorage.removeItem(key)
  }
  window.sessionStorage.clear()
  if ('caches' in window) {
    const names = await window.caches.keys()
    const deleted = await Promise.all(names.map((name) => window.caches.delete(name)))
    if (deleted.some((result) => !result)) throw new Error('cache cleanup failed')
  }
  if (typeof indexedDB !== 'undefined') {
    if (!('databases' in indexedDB) || typeof indexedDB.databases !== 'function') throw new Error('indexedDB cleanup is unavailable')
    const databases = await indexedDB.databases()
    await Promise.all(databases.flatMap((database) => database.name !== undefined ? [deleteDatabase(database.name)] : []))
  }
}

function deleteDatabase(name: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.deleteDatabase(name)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error ?? new Error('indexedDB cleanup failed'))
    request.onblocked = () => reject(new Error('indexedDB cleanup blocked'))
  })
}

export async function endCurrentSession(): Promise<void> {
  const response = await fetch('/api/auth/logout', {
    method: 'POST',
    credentials: 'same-origin',
    cache: 'no-store',
    headers: { 'X-CSRF-Token': csrfToken() },
  })
  if (response.status !== 204 && response.status !== 401) throw new Error('logout failed')
  await clearProtectedBrowserData()
}
