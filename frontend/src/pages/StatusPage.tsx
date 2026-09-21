import { useEffect, useState } from 'react'
import { frontendBuildSha } from '../build'
import { useI18n } from '../i18n'
import { classifyStatus, type Probe, type StatusResult } from '../status'

export function StatusPage() {
  const [result, setResult] = useState<StatusResult>({ kind: 'checking' })
  const { messages } = useI18n()

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    let timedOut = false
    const timeout = window.setTimeout(() => {
      timedOut = true
      controller.abort()
      if (active) setResult({ kind: 'unavailable' })
    }, 5_000)
    fetch('/api/readyz', {
      cache: 'no-store',
      credentials: 'same-origin',
      signal: controller.signal,
    })
      .then(async (response) => classifyStatus(response, (await response.json()) as Probe))
      .then((nextResult) => {
        window.clearTimeout(timeout)
        if (active && !timedOut) setResult(nextResult)
      })
      .catch(() => {
        window.clearTimeout(timeout)
        if (active) setResult({ kind: 'unavailable' })
      })
    return () => {
      active = false
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [])

  const labels: Record<StatusResult['kind'], string> = messages.status.states

  return (
    <main className="status-page" id="main-content">
      <p className="eyebrow">{messages.status.eyebrow}</p>
      <h1 tabIndex={-1}>{messages.status.title}</h1>
      <p role="status" data-result={result.kind}>{labels[result.kind]}</p>
      <dl>
        <dt>{messages.status.frontendRevision}</dt>
        <dd><code>{frontendBuildSha}</code></dd>
        <dt>{messages.status.apiRevision}</dt>
        <dd><code>{result.kind === 'ready' && result.backendSha ? result.backendSha : messages.status.notAvailable}</code></dd>
      </dl>
    </main>
  )
}
