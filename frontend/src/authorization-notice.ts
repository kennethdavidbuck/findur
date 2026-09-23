export const authorizationNoticeValues = [
  'denied',
  'restart_required',
  'authorization_unavailable',
  'initialization_failed',
  'invalid_request',
] as const

export type AuthorizationNotice = typeof authorizationNoticeValues[number]

export function authorizationNoticeFromSearch(search: string): AuthorizationNotice | undefined {
  const value = new URLSearchParams(search).get('authorization')
  return authorizationNoticeValues.find((notice) => notice === value)
}
