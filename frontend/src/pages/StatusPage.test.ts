import { describe, expect, it } from 'vitest'
import { classifyStatus } from '../status'

const sha = '0123456789abcdef0123456789abcdef01234567'
const otherSha = '89abcdef0123456789abcdef0123456789abcdef'

describe('status result classification', () => {
  it('accepts only matching full revisions', () => {
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: sha }, sha))
      .toEqual({ kind: 'match', backendSha: sha })
  })

  it('distinguishes mismatch, malformed, stale, and unavailable results', () => {
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: otherSha }, sha).kind)
      .toBe('mismatch')
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: 'main' }, sha).kind)
      .toBe('malformed')
    expect(classifyStatus({ ok: false, status: 409 }, {}, sha).kind).toBe('stale')
    expect(classifyStatus({ ok: false, status: 503 }, {}, sha).kind).toBe('unavailable')
  })
})
