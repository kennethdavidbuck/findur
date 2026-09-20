import assert from 'node:assert/strict'

const fetchBounded = (url, init = {}) => fetch(url, {
  ...init,
  signal: AbortSignal.timeout(15_000),
})

export async function verifyBrowserStatus({ browserUrl, publicOrigin, expectedSha }) {
  if (!browserUrl || !publicOrigin || !/^[0-9a-f]{40}$/.test(expectedSha || '')) {
    throw new Error('browser URL, public origin, and a full lowercase expected SHA are required')
  }
  const webdriver = async (path, init) => {
    const response = await fetchBounded(`${browserUrl}${path}`, init)
    const payload = response.status === 204 ? {} : await response.json()
    if (!response.ok) throw new Error(`WebDriver ${path} failed: ${response.status} ${JSON.stringify(payload)}`)
    return payload.value
  }

  const session = await webdriver('/session', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      capabilities: {
        alwaysMatch: {
          browserName: 'chrome',
          'goog:chromeOptions': { args: ['--headless=new', '--no-sandbox', '--disable-dev-shm-usage'] },
        },
      },
    }),
  })
  const sessionId = session.sessionId
  try {
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
            text: document.body.innerText,
            origin: location.origin,
            readyRequests: performance.getEntriesByType('resource').map(entry => entry.name).filter(name => name.includes('/api/readyz'))
          }`,
          args: [],
        }),
      })
      if (evidence.result === 'match') break
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
    assert.equal(evidence.result, 'match', 'browser JavaScript verifies an exact build match')
    assert.match(evidence.text, new RegExp(expectedSha))
    assert.equal(evidence.readyRequests.length, 1, 'status page makes exactly one readiness call')
    assert.equal(new URL(evidence.readyRequests[0]).origin, evidence.origin, 'readiness is same-origin')
  } finally {
    await webdriver(`/session/${sessionId}`, { method: 'DELETE' })
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await verifyBrowserStatus({
    browserUrl: process.env.BROWSER_URL,
    publicOrigin: process.env.PUBLIC_ORIGIN,
    expectedSha: process.env.EXPECTED_SHA,
  })
  console.log('browser exact-build check passed')
}
