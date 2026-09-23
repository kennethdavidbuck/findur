import { useEffect, useState } from 'react'

export type AuthorizationStatus = {
  authorizationAvailable: boolean
  authenticated: boolean
  reauthorizationRequired: boolean
}

type StatusState =
  | { resolving: true; status: null; validationKey?: string }
  | { resolving: false; status: AuthorizationStatus; validationKey?: string }

export async function getAuthorizationStatus(signal?: AbortSignal): Promise<AuthorizationStatus> {
  const response = await fetch('/api/auth/status', { cache: 'no-store', credentials: 'same-origin', signal })
  if (!response.ok) throw new Error('authorization status unavailable')
  const value = await response.json() as Partial<AuthorizationStatus>
  if (typeof value.authorizationAvailable !== 'boolean' || typeof value.authenticated !== 'boolean' || typeof value.reauthorizationRequired !== 'boolean') {
    throw new Error('authorization status malformed')
  }
  return {
    authorizationAvailable: value.authorizationAvailable,
    authenticated: value.authenticated,
    reauthorizationRequired: value.reauthorizationRequired,
  }
}

export function useAuthorizationStatus(initialStatus?: AuthorizationStatus, validationKey?: string): StatusState {
  const [state, setState] = useState<StatusState>(() => initialStatus
    ? { resolving: false, status: initialStatus, validationKey }
    : { resolving: true, status: null, validationKey })

  useEffect(() => {
    if (initialStatus) return
    const controller = new AbortController()
    let disposed = false
    const timeout = window.setTimeout(() => controller.abort(), 10_000)
    void getAuthorizationStatus(controller.signal)
      .then((value) => {
        window.clearTimeout(timeout)
        if (disposed) return
        setState({ resolving: false, status: value, validationKey })
      })
      .catch(() => {
        window.clearTimeout(timeout)
        if (disposed) return
        setState({ resolving: false, status: { authorizationAvailable: false, authenticated: false, reauthorizationRequired: false }, validationKey })
      })
    return () => {
      disposed = true
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [initialStatus, validationKey])

  return state
}
