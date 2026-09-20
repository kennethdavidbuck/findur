import assert from 'node:assert/strict'
import { withWebDriverSession } from './webdriver.mjs'

export async function verifyBrowserOAuth({ browserUrl, oauthOrigin }) {
  if (!oauthOrigin) throw new Error('OAuth origin is required')
  const oauthUrl = new URL(oauthOrigin)
  if (oauthUrl.protocol !== 'http:' || oauthUrl.hostname !== '127.0.0.1' || oauthUrl.origin !== oauthOrigin.replace(/\/$/, '')) {
    throw new Error('OAuth origin must be an HTTP 127.0.0.1 origin')
  }

  await withWebDriverSession(browserUrl, async ({ webdriver, sessionId }) => {
    await webdriver(`/session/${sessionId}/url`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: `${oauthUrl.origin}/connect` }),
    })
    let actionReady = false
    for (let attempt = 0; attempt < 40; attempt += 1) {
      actionReady = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const button=document.querySelector('form[action="/api/auth/snaptrade/authorize"] button'); return Boolean(button && !button.disabled)`, args: [] }),
      })
      if (actionReady) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(actionReady, true, 'authorization action opens only after the server capability check')
    await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `document.querySelector('form[action="/api/auth/snaptrade/authorize"] button').click()`, args: [] }),
    })
    let oauthEvidence
    for (let attempt = 0; attempt < 60; attempt += 1) {
      oauthEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `return { path: location.pathname, search: location.search, heading: document.querySelector('h1')?.innerText }`, args: [] }),
      })
      if (oauthEvidence.path === '/portfolio' && oauthEvidence.heading === 'Your masked account inventory') break
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
    assert.deepEqual(oauthEvidence, { path: '/portfolio', search: '', heading: 'Your masked account inventory' })
    let inventoryEvidence
    for (let attempt = 0; attempt < 40; attempt += 1) {
      inventoryEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `return {text:document.body.innerText,focused:document.activeElement===document.querySelector('h1')}`, args: [] }),
      })
      if (inventoryEvidence.text.includes('Retirement (•••• 8443)')) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(inventoryEvidence.focused, true, 'the masked inventory heading retains meaningful focus')
    assert.match(inventoryEvidence.text, /Connection status[\s\S]*Retirement \(•••• 8443\)/, 'connection state precedes accounts')
    assert.match(inventoryEvidence.text, /No account is included by default/, 'account inclusion remains a later choice')
    for (const forbidden of ['Q6542138443', '15363.23', 'RAW-SYMBOL-DO-NOT-RENDER']) assert.doesNotMatch(inventoryEvidence.text, new RegExp(forbidden), `${forbidden} is not rendered`)
    const browserCookies = await webdriver(`/session/${sessionId}/cookie`)
    assert.ok(browserCookies.some((cookie) => cookie.name === 'findur_session' && cookie.httpOnly && cookie.secure), 'opaque secure session cookie is issued')
    assert.ok(browserCookies.some((cookie) => cookie.name === 'findur_csrf' && !cookie.httpOnly && cookie.secure), 'separate secure CSRF cookie is issued')
    assert.ok(!browserCookies.some((cookie) => cookie.name === 'findur_oauth_attempt'), 'attempt cookie is expired')
  })
}
