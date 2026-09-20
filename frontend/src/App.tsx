import { useEffect, useState } from 'react'
import { Button } from 'react-aria-components'

type ConnectionState = 'checking' | 'connected' | 'unavailable'

type HealthResponse = {
  status: string
}

const healthTimeoutMs = 5_000

async function fetchHealth(): Promise<ConnectionState> {
  const controller = new AbortController()
  const timeout = window.setTimeout(() => controller.abort(), healthTimeoutMs)
  try {
    const response = await fetch('/api/healthz', {
      headers: { Accept: 'application/json' },
      signal: controller.signal,
    })
    if (!response.ok) {
      return 'unavailable'
    }
    const body = (await response.json()) as HealthResponse
    return body.status === 'ok' ? 'connected' : 'unavailable'
  } catch {
    return 'unavailable'
  } finally {
    window.clearTimeout(timeout)
  }
}

export function App() {
  const [connection, setConnection] = useState<ConnectionState>('checking')

  useEffect(() => {
    let active = true
    void fetchHealth().then((state) => {
      if (active) {
        setConnection(state)
      }
    })
    return () => {
      active = false
    }
  }, [])

  const message = {
    checking: 'Checking service connection…',
    connected: 'Service connected',
    unavailable: 'Service temporarily unavailable',
  }[connection]

  return (
    <main>
      <section className="status-card" aria-labelledby="status-heading">
        <p className="eyebrow">Findur</p>
        <h1 id="status-heading">Walking skeleton</h1>
        <div className={`connection connection--${connection}`} role="status" aria-live="polite">
          <span className="connection__dot" aria-hidden="true" />
          {message}
        </div>
        <p className="description">
          The public shell is online. Product features will arrive only after the delivery path is proven.
        </p>
        {connection === 'unavailable' ? (
          <Button
            className="retry"
            onPress={() => {
              setConnection('checking')
              void fetchHealth().then(setConnection)
            }}
          >
            Check again
          </Button>
        ) : null}
      </section>
    </main>
  )
}
