import { useEffect, useState } from 'react'

export type AuthorizationStatus = {
  authorizationAvailable: boolean
  authenticated: boolean
}

type StatusState = { resolving: true; status: null } | { resolving: false; status: AuthorizationStatus }

export function useAuthorizationStatus(): StatusState {
  const [state, setState] = useState<StatusState>({ resolving: true, status: null })

  useEffect(() => {
    const controller = new AbortController()
    let disposed = false
    const timeout = window.setTimeout(() => controller.abort(), 10_000)
    void fetch('/api/auth/status', { cache: 'no-store', credentials: 'same-origin', signal: controller.signal })
      .then(async (response) => {
        if (!response.ok) throw new Error('authorization status unavailable')
        const value = await response.json() as Partial<AuthorizationStatus>
        if (typeof value.authorizationAvailable !== 'boolean' || typeof value.authenticated !== 'boolean') {
          throw new Error('authorization status malformed')
        }
        window.clearTimeout(timeout)
        setState({ resolving: false, status: value as AuthorizationStatus })
      })
      .catch((error: unknown) => {
        window.clearTimeout(timeout)
        if (disposed && error instanceof DOMException && error.name === 'AbortError') return
        setState({ resolving: false, status: { authorizationAvailable: false, authenticated: false } })
      })
    return () => {
      disposed = true
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [])

  return state
}
