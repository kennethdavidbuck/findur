import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { App } from './App'

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

describe('App', () => {
  it('reports a successful same-origin health check accessibly', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: 'ok' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByRole('status')).toHaveTextContent('Service connected')
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/healthz',
      expect.objectContaining({
        headers: { Accept: 'application/json' },
        signal: expect.anything(),
      }),
    )
  })

  it('reports a failed health check without provider details', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('private provider failure')))

    render(<App />)

    expect(await screen.findByRole('status')).toHaveTextContent('Service temporarily unavailable')
    expect(screen.queryByText(/private provider failure/i)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Check again' })).toBeInTheDocument()
  })

  it('aborts a health check after five seconds', async () => {
    vi.useFakeTimers()
    const fetchMock = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      return new Promise<Response>((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => {
          reject(new DOMException('Aborted', 'AbortError'))
        })
      })
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000)
    })

    expect(screen.getByRole('status')).toHaveTextContent('Service temporarily unavailable')
    const options = fetchMock.mock.calls[0]?.[1]
    expect(options?.signal?.aborted).toBe(true)
  })

  it('recovers when a failed health check is retried successfully', async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error('temporary failure'))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ status: 'ok' }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: 'Check again' }))

    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('Service connected'))
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
