const memory = new Map<string, string>()
const fallbackKeys = new Set<string>()

// Browser storage is only a best-effort public/startup cache. This adapter keeps
// page-lifetime choices working when storage is unavailable or throws.
export const preferenceStorage = {
  get(key: string): string | null {
    try {
      const value = window.localStorage.getItem(key)
      if (value !== null) {
        memory.set(key, value)
        fallbackKeys.delete(key)
        return value
      }
      if (fallbackKeys.has(key)) return memory.get(key) ?? null
      memory.delete(key)
      return null
    } catch { return memory.get(key) ?? null }
  },
  set(key: string, value: string) {
    memory.set(key, value)
    try {
      window.localStorage.setItem(key, value)
      fallbackKeys.delete(key)
    } catch {
      fallbackKeys.add(key)
    }
  },
}
