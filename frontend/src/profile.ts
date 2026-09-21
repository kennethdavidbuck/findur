import type { components } from './generated/api'
import { csrfToken } from './session'

export type PersonalProfile = components['schemas']['PersonalProfile']
export type PersonalProfileInput = components['schemas']['PersonalProfileInput']
export type PersonalProfileSnapshot = components['schemas']['PersonalProfileSnapshot']

export class ProfileSessionExpiredError extends Error {}
export class ProfileDefenseError extends Error {}
export class ProfileConflictError extends Error {}
export class ProfileValidationError extends Error {
  constructor(readonly fields: string[]) { super('invalid profile') }
}

export async function getPersonalProfile(): Promise<PersonalProfileSnapshot> {
  return request<PersonalProfileSnapshot>('/api/profile', { method: 'GET' })
}

export async function putPersonalProfile(input: PersonalProfileInput): Promise<PersonalProfile> {
  return request<PersonalProfile>('/api/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken() },
    body: JSON.stringify(input),
  })
}

async function request<T>(path: string, init: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, cache: 'no-store', credentials: 'same-origin' })
  if (response.status === 401) throw new ProfileSessionExpiredError('profile session expired')
  if (response.status === 403) throw new ProfileDefenseError('profile request defense failed')
  if (response.status === 409) throw new ProfileConflictError('profile changed')
  if (response.status === 400) {
    const body: unknown = await response.json()
    throw new ProfileValidationError(isRecord(body) && Array.isArray(body.fields) ? body.fields.filter((field): field is string => typeof field === 'string') : [])
  }
  if (!response.ok) throw new Error('profile unavailable')
  return await response.json() as T
}

function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null }
