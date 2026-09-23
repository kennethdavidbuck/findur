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

    const declineStatus = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/__fixture/oidc/deny-next',{method:'POST'}).then(response=>done(response.status),error=>done(String(error)))`, args: [] }),
    })
    assert.equal(declineStatus, 204, 'integration provider is armed for one declined authorization')
    await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `document.querySelector('form[action="/api/auth/snaptrade/authorize"] button').click()`, args: [] }),
    })
    let denialEvidence
    for (let attempt = 0; attempt < 60; attempt += 1) {
      denialEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const alert=document.querySelector('[role="alert"]');return {path:location.pathname,search:location.search,text:alert?.innerText||'',focused:document.activeElement===alert,retry:[...document.querySelectorAll('button')].some(button=>button.textContent.trim()==='Try SnapTrade again'&&!button.disabled)}`, args: [] }),
      })
      if (denialEvidence.retry && denialEvidence.focused) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(denialEvidence.path, '/connect')
    assert.equal(denialEvidence.search, '', 'declined authorization detail is removed from browser history')
    assert.match(denialEvidence.text, /You didn’t give Findur access\. Nothing was changed/)
    assert.equal(denialEvidence.focused, true, 'declined authorization guidance receives focus')
    assert.equal(denialEvidence.retry, true, 'declined authorization offers a useful retry')

    await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const form=document.createElement('form');form.method='post';form.action='/api/auth/snaptrade/authorize';form.enctype='text/plain';const input=document.createElement('input');input.name='returnTo';input.value='/portfolio';form.append(input);document.body.append(form);form.submit()`, args: [] }),
    })
    let invalidStartEvidence
    for (let attempt = 0; attempt < 40; attempt += 1) {
      invalidStartEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const alert=document.querySelector('[role="alert"]');return {path:location.pathname,search:location.search,text:alert?.innerText||'',raw:document.body.innerText.includes('invalid_request')}`, args: [] }),
      })
      if (invalidStartEvidence.text) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(invalidStartEvidence.path, '/connect')
    assert.equal(invalidStartEvidence.search, '')
    assert.match(invalidStartEvidence.text, /We couldn’t begin that sign-in safely\. Nothing was changed/)
    assert.equal(invalidStartEvidence.raw, false, 'browser initiation failures never render a raw error code')

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
      if (oauthEvidence.path === '/onboarding/accounts' && ['Choose what Findur may use.', 'Choisissez ce que Findur peut utiliser.'].includes(oauthEvidence.heading)) break
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
    if (oauthEvidence.heading === 'Choisissez ce que Findur peut utiliser.') {
      await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `document.querySelector('.authenticated-header .choice-group input[value="en"]')?.click()`, args: [] }),
      })
      for (let attempt = 0; attempt < 40; attempt += 1) {
        oauthEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ script: `return { path: location.pathname, search: location.search, heading: document.querySelector('h1')?.innerText }`, args: [] }),
        })
        if (oauthEvidence.heading === 'Choose what Findur may use.') break
        await new Promise((resolve) => setTimeout(resolve, 100))
      }
    }
    assert.deepEqual(oauthEvidence, { path: '/onboarding/accounts', search: '', heading: 'Choose what Findur may use.' })
    await webdriver(`/session/${sessionId}/window/rect`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ width: 1032, height: 900 }),
    })
    const setupLayout = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const header=document.querySelector('.authenticated-header').getBoundingClientRect();const eyebrow=document.querySelector('.portfolio-inventory > .eyebrow').getBoundingClientRect();const progress=document.querySelector('#setup-progress-title').getBoundingClientRect();return {headerBottom:header.bottom,eyebrowTop:eyebrow.top,progressWidth:progress.width,progressHeight:progress.height}`, args: [] }),
    })
    assert.ok(setupLayout.headerBottom <= setupLayout.eyebrowTop, `1032px setup controls clear the OAuth return text; observed: ${JSON.stringify(setupLayout)}`)
    assert.ok(setupLayout.progressWidth <= 1 && setupLayout.progressHeight <= 1, `the accessible setup label is not visually rendered; observed: ${JSON.stringify(setupLayout)}`)
    let inventoryEvidence
    let nextInventoryCheckAt = 0
    for (let attempt = 0; attempt < 60; attempt += 1) {
      const mayCheckAgain = Date.now() >= nextInventoryCheckAt
      inventoryEvidence = await webdriver(`/session/${sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const mayCheckAgain=arguments[0];const text=document.body.innerText;const check=[...document.querySelectorAll('button')].find(button=>button.textContent.trim()==='Check again'&&!button.disabled);const clicked=Boolean(mayCheckAgain&&check);if(clicked)check.click();return {text,clicked,focused:document.activeElement===document.querySelector('h1'),chooserCount:document.querySelectorAll('.account-selection').length,inventoryCount:document.querySelectorAll('.inventory-connections').length,accountLabelCount:text.split('Healthy Realtime — Full Data (•••• X001)').length-1}`, args: [mayCheckAgain] }),
      })
      if (inventoryEvidence.clicked) nextInventoryCheckAt = Date.now() + 3000
      if (inventoryEvidence.text.includes('Healthy Realtime — Full Data (•••• X001)') && inventoryEvidence.chooserCount === 1) break
      await new Promise((resolve) => setTimeout(resolve, 250))
    }
    assert.equal(inventoryEvidence.focused, true, 'the account-choice heading retains meaningful focus')
    assert.equal(inventoryEvidence.chooserCount, 1, `one account chooser is rendered; observed: ${JSON.stringify(inventoryEvidence)}`)
    assert.equal(inventoryEvidence.inventoryCount, 0, 'the raw inventory is not duplicated beside the chooser')
    assert.equal(inventoryEvidence.accountLabelCount, 1, `the account is rendered once; observed: ${JSON.stringify(inventoryEvidence)}`)
    assert.match(inventoryEvidence.text, /Pick the accounts you’d like to include\. You can change this anytime\./)
    assert.match(inventoryEvidence.text, /Delayed Account — Empty Positions \(•••• X002\)/)
    assert.match(inventoryEvidence.text, /Cash Only — No Positions \(•••• CASH\)/)
    assert.match(inventoryEvidence.text, /Saved to your profile:\s*0\s*of\s*28\s*accounts shown/)
    for (const forbidden of ['SANDBOX-001', 'SANDBOX-002', 'SANDBOX-CASH', '25000', '12500', '5000']) assert.doesNotMatch(inventoryEvidence.text, new RegExp(forbidden), `${forbidden} is not rendered`)
    const browserCookies = await webdriver(`/session/${sessionId}/cookie`)
    assert.ok(browserCookies.some((cookie) => cookie.name === 'findur_session' && cookie.httpOnly && cookie.secure), 'opaque secure session cookie is issued')
    assert.ok(browserCookies.some((cookie) => cookie.name === 'findur_csrf' && !cookie.httpOnly && cookie.secure), 'separate secure CSRF cookie is issued')
    assert.ok(!browserCookies.some((cookie) => cookie.name === 'findur_oauth_attempt'), 'attempt cookie is expired')
  })
}
