import { describe, expect, it } from 'vitest'
import { classifyStatus } from '../status'

const sha = '0123456789abcdef0123456789abcdef01234567'
const otherSha = '89abcdef0123456789abcdef0123456789abcdef'

describe('status result classification', () => {
  it('reports a ready API even when the revisions differ', () => {
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: sha }))
      .toEqual({ kind: 'ready', backendSha: sha })
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: otherSha }))
      .toEqual({ kind: 'ready', backendSha: otherSha })
  })

  it('reports unavailable for a non-ready response', () => {
    expect(classifyStatus({ ok: false, status: 503 }, {}).kind).toBe('unavailable')
    expect(classifyStatus({ ok: false, status: 409 }, {}).kind).toBe('unavailable')
  })

  it('keeps a malformed API revision diagnostic unavailable without changing readiness', () => {
    expect(classifyStatus({ ok: true, status: 200 }, { status: 'ready', buildSha: 'main' }))
      .toEqual({ kind: 'ready', backendSha: undefined })
  })
})
