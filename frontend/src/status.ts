import { frontendBuildSha, isFullGitSha } from './build'

export type Probe = { status?: unknown; buildSha?: unknown }
export type StatusResult =
  | { kind: 'checking' }
  | { kind: 'match'; backendSha: string }
  | { kind: 'unavailable' | 'malformed' | 'stale' | 'mismatch'; backendSha?: string }

export function classifyStatus(
  response: Pick<Response, 'ok' | 'status'>,
  payload: Probe,
  frontendSha = frontendBuildSha,
): StatusResult {
  if (response.status === 409) return { kind: 'stale' }
  if (!response.ok || payload.status !== 'ready') return { kind: 'unavailable' }
  if (!isFullGitSha(frontendSha) || !isFullGitSha(payload.buildSha)) {
    return { kind: 'malformed' }
  }
  if (payload.buildSha !== frontendSha) {
    return { kind: 'mismatch', backendSha: payload.buildSha }
  }
  return { kind: 'match', backendSha: payload.buildSha }
}

