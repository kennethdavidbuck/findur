const fetchBounded = (url, init = {}) => fetch(url, {
  ...init,
  signal: AbortSignal.timeout(15_000),
})

export async function withWebDriverSession(browserUrl, verify) {
  if (!browserUrl) throw new Error('browser URL is required')

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
    await verify({ webdriver, sessionId })
  } finally {
    await webdriver(`/session/${sessionId}`, { method: 'DELETE' })
  }
}
