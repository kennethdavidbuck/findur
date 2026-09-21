import { isFullGitSha } from './build'

export type Probe = { status?: unknown; buildSha?: unknown }
export type StatusResult =
  | { kind: 'checking' }
  | { kind: 'ready'; backendSha?: string }
  | { kind: 'unavailable' }

export function classifyStatus(
  response: Pick<Response, 'ok' | 'status'>,
  payload: Probe,
): StatusResult {
  if (!response.ok || payload.status !== 'ready') return { kind: 'unavailable' }
  return { kind: 'ready', backendSha: isFullGitSha(payload.buildSha) ? payload.buildSha : undefined }
}
