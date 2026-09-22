import { useEffect, useState } from 'react'

export type AuthorizationStatus = {
  authorizationAvailable: boolean
  authenticated: boolean
}

type StatusState = { resolving: true; status: null } | { resolving: false; status: AuthorizationStatus }

export async function getAuthorizationStatus(signal?: AbortSignal): Promise<AuthorizationStatus> {
  const response = await fetch('/api/auth/status', { cache: 'no-store', credentials: 'same-origin', signal })
  if (!response.ok) throw new Error('authorization status unavailable')
  const value = await response.json() as Partial<AuthorizationStatus>
  if (typeof value.authorizationAvailable !== 'boolean' || typeof value.authenticated !== 'boolean') {
    throw new Error('authorization status malformed')
  }
  return value as AuthorizationStatus
}

export function useAuthorizationStatus(initialStatus?: AuthorizationStatus): StatusState {
  const [state, setState] = useState<StatusState>(() => initialStatus
    ? { resolving: false, status: initialStatus }
    : { resolving: true, status: null })

  useEffect(() => {
    if (initialStatus) return
    const controller = new AbortController()
    let disposed = false
    const timeout = window.setTimeout(() => controller.abort(), 10_000)
    void getAuthorizationStatus(controller.signal)
      .then((value) => {
        window.clearTimeout(timeout)
        setState({ resolving: false, status: value })
      })
      .catch(() => {
        window.clearTimeout(timeout)
        if (disposed) return
        setState({ resolving: false, status: { authorizationAvailable: false, authenticated: false } })
      })
    return () => {
      disposed = true
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [initialStatus])

  return state
}
