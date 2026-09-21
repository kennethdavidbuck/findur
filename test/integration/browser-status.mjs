import assert from 'node:assert/strict'
import { withWebDriverSession } from './webdriver.mjs'

export async function verifyBrowserStatus({ browserUrl, publicOrigin }) {
  if (!browserUrl || !publicOrigin) {
    throw new Error('browser URL and public origin are required')
  }

  await withWebDriverSession(browserUrl, async ({ webdriver, sessionId }) => {
    await webdriver(`/session/${sessionId}/url`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: `${publicOrigin.replace(/\/$/, '')}/__status` }),
    })

    let evidence
    for (let attempt = 0; attempt < 40; attempt += 1) {
      evidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          script: `return {
            result: document.querySelector('[data-result]')?.dataset.result,
            origin: location.origin,
            revisions: Array.from(document.querySelectorAll('dl dd code'), ({ textContent }) => textContent),
            readyRequests: performance.getEntriesByType('resource').map(entry => entry.name).filter(name => name.includes('/api/readyz'))
          }`,
          args: [],
        }),
      })
      if (evidence.result === 'ready') break
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
    assert.equal(evidence.result, 'ready', 'browser JavaScript verifies deployment health')
    assert.equal(evidence.revisions.length, 2, 'status page renders separate frontend and API diagnostics')
    for (const revision of evidence.revisions) {
      assert.match(revision || '', /^[0-9a-f]{40}$/, 'each deployment diagnostic is a full revision SHA')
    }
    assert.equal(evidence.readyRequests.length, 1, 'status page makes exactly one readiness call')
    assert.equal(new URL(evidence.readyRequests[0]).origin, evidence.origin, 'readiness is same-origin')
  })
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await verifyBrowserStatus({
    browserUrl: process.env.BROWSER_URL,
    publicOrigin: process.env.PUBLIC_ORIGIN,
  })
  console.log('browser deployment health check passed')
}
