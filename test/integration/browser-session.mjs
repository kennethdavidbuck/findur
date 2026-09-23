import assert from 'node:assert/strict'
import { withWebDriverSession } from './webdriver.mjs'
import { exerciseLargeInventory } from './browser-large-inventory.mjs'
import { removeDiagnosticScenario } from './diagnostic-scenarios.mjs'

async function establishSession(webdriver, sessionId, origin) {
  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/connect` }),
  })
  for (let attempt = 0; attempt < 40; attempt += 1) {
    const ready = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const button=document.querySelector('form[action="/api/auth/snaptrade/authorize"] button'); if(button&&!button.disabled){button.click();return true} return false`, args: [] }),
    })
    if (ready) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  let state
  for (let attempt = 0; attempt < 60; attempt += 1) {
    state = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return {path:location.pathname,heading:document.querySelector('h1')?.innerText,focused:document.activeElement===document.querySelector('h1')}`, args: [] }),
    })
    if (state.path === '/onboarding/accounts' && state.heading === 'Choose what Findur may use.') break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.deepEqual(state, { path: '/onboarding/accounts', heading: 'Choose what Findur may use.', focused: true })
}

async function bypassActiveSessionFromLanding(webdriver, sessionId, origin, expected, message) {
  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/` }),
  })
  const openedConsent = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const link=document.querySelector('.public-header .owner-button[href="/connect"]');link?.click();return Boolean(link)`, args: [] }),
  })
  assert.equal(openedConsent, true, 'the landing-page Login link opens staged consent')
  let state
  for (let attempt = 0; attempt < 40; attempt += 1) {
    state = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const navigation=document.querySelector('.authenticated-nav');return {path:location.pathname,heading:document.querySelector('h1')?.innerText,consentAction:!!document.querySelector('form[action="/api/auth/snaptrade/authorize"] button'),navigationPosition:navigation?getComputedStyle(navigation).position:null}`, args: [] }),
    })
    if (state.path === expected.path && state.heading === expected.heading) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(state, expected, message)
}

async function verifyIncompleteActiveSessionBypassesConsent(webdriver, sessionId, origin) {
  await bypassActiveSessionFromLanding(
    webdriver,
    sessionId,
    origin,
    { path: '/onboarding/accounts', heading: 'Choose what Findur may use.', consentAction: false, navigationPosition: null },
    'an incomplete active session bypasses staged consent and renders account setup',
  )
  await webdriver(`/session/${sessionId}/back`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })
  let returnState
  for (let attempt = 0; attempt < 40; attempt += 1) {
    returnState = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return {path:location.pathname,heading:document.querySelector('h1')?.innerText}`, args: [] }),
    })
    if (returnState.path === '/onboarding/accounts' && returnState.heading === 'Choose what Findur may use.') break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(returnState, { path: '/onboarding/accounts', heading: 'Choose what Findur may use.' }, 'the active-session redirect replaces staged consent in browser history')
}

async function verifyCompletedActiveSessionBypassesConsent(webdriver, sessionId, origin) {
  await bypassActiveSessionFromLanding(
    webdriver,
    sessionId,
    origin,
    { path: '/portfolio', heading: 'Your portfolio', consentAction: false, navigationPosition: 'fixed' },
    'a completed active session bypasses staged consent and renders the authenticated Portfolio shell',
  )
}

async function exerciseInventoryFixtures(webdriver, sessionId, wiremockUrl) {
  const fixtures = [
    { name: 'empty', status: 200, bodyFileName: 'inventory/empty.json', want: 'empty' },
    { name: 'disabled', status: 200, bodyFileName: 'inventory/disabled.json', want: 'disabled' },
    { name: 'malformed', status: 200, bodyFileName: 'inventory/malformed.json', want: 'malformed' },
    { name: 'transient', status: 503, bodyFileName: 'inventory/transient.json', want: 'unavailable' },
    { name: 'rate limited', status: 429, bodyFileName: 'inventory/rate-limited.json', want: 'rate_limited' },
  ]
  for (const fixture of fixtures) {
    const mappingResponse = await fetch(`${wiremockUrl}/__admin/mappings`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: AbortSignal.timeout(15_000),
      body: JSON.stringify({
        priority: 1,
        request: { method: 'GET', urlPath: '/authorizations' },
        response: { status: fixture.status, headers: { 'Content-Type': 'application/json' }, bodyFileName: fixture.bodyFileName },
      }),
    })
    assert.equal(mappingResponse.status, 201, `${fixture.name} WireMock mapping registered`)
    const mapping = await mappingResponse.json()
    try {
      const result = await webdriver(`/session/${sessionId}/execute/async`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const done=arguments[arguments.length-1];const csrf=document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'';fetch('/api/portfolio/inventory/retry',{method:'POST',credentials:'same-origin',headers:{'X-CSRF-Token':decodeURIComponent(csrf)}}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
      })
      assert.equal(result.status, 200, `${fixture.name} inventory retry status`)
      assert.equal(result.cache, 'private, no-store', `${fixture.name} inventory cache policy`)
      assert.equal(result.body.state, fixture.want, `${fixture.name} categorical state`)
      assert.ok(Array.isArray(result.body.connections), `${fixture.name} normalized connections`)
      const serialized = JSON.stringify(result.body)
      for (const forbidden of ['25000', '12500', '5000', 'synthetic-access-token']) assert.doesNotMatch(serialized, new RegExp(forbidden), `${fixture.name} response is minimized`)
      if (fixture.want === 'disabled') {
        assert.equal(result.body.connections[0]?.brokerageLabel, 'Synthetic Disabled Broker')
        assert.equal(result.body.connections[0]?.available, false)
      }
      if (fixture.want === 'rate_limited') assert.ok(Date.parse(result.body.retryAt) > Date.now(), 'headerless 429 receives a safe retry floor')
    } finally {
      const removed = await fetch(`${wiremockUrl}/__admin/mappings/${mapping.id}`, { method: 'DELETE', signal: AbortSignal.timeout(15_000) })
      assert.ok(removed.ok, `${fixture.name} WireMock mapping removed`)
    }
  }
}

async function exerciseAuthorizationRenewal(webdriver, sessionId, origin, wiremockUrl) {
  const mappingResponse = await fetch(`${wiremockUrl}/__admin/mappings`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: AbortSignal.timeout(15_000),
    body: JSON.stringify({
      priority: 1,
      request: { method: 'GET', urlPath: '/authorizations' },
      response: { status: 401, headers: { 'Content-Type': 'application/json' }, jsonBody: { detail: 'synthetic revoked grant' } },
    }),
  })
  assert.equal(mappingResponse.status, 201, 'revoked-grant WireMock mapping registered')
  const mapping = await mappingResponse.json()
  try {
    const unauthorized = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];const csrf=decodeURIComponent(document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'');fetch('/api/portfolio/inventory/retry',{method:'POST',credentials:'same-origin',headers:{'X-CSRF-Token':csrf}}).then(async response=>done({status:response.status,body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    assert.equal(unauthorized.status, 200)
    assert.equal(unauthorized.body.state, 'unauthorized', 'a repeated provider 401 requires permission renewal')
  } finally {
    const removed = await fetch(`${wiremockUrl}/__admin/mappings/${mapping.id}`, { method: 'DELETE', signal: AbortSignal.timeout(15_000) })
    assert.ok(removed.ok, 'revoked-grant WireMock mapping removed')
  }

  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/connect` }),
  })
  let renewal
  for (let attempt = 0; attempt < 40; attempt += 1) {
    renewal = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const alert=document.querySelector('[role="alert"]');const labels=['Reconnect with SnapTrade','Reconnecter avec SnapTrade'];const button=[...document.querySelectorAll('button')].find(value=>labels.includes(value.textContent.trim()));return {path:location.pathname,text:alert?.innerText||'',ready:Boolean(button&&!button.disabled),focused:document.activeElement===alert,csrfPresent:document.cookie.split('; ').some(value=>value.startsWith('findur_csrf='))}`, args: [] }),
    })
    if (renewal.ready) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(renewal.path, '/connect', 'authorization loss returns the browser to Connect')
  assert.match(renewal.text, /Your saved account choices are still here|Vos choix de comptes sont toujours enregistrés/)
  assert.equal(renewal.focused, true, 'permission-renewal guidance receives focus')
  assert.equal(renewal.ready, true, 'permission renewal offers a fresh hosted authorization')
  assert.equal(renewal.csrfPresent, false, 'permission renewal expires the Findur session cookies')

  await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const labels=['Reconnect with SnapTrade','Reconnecter avec SnapTrade'];[...document.querySelectorAll('button')].find(value=>labels.includes(value.textContent.trim())).click()`, args: [] }),
  })
  let recovered
  for (let attempt = 0; attempt < 60; attempt += 1) {
    recovered = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];if(location.pathname!=='/portfolio'){done({path:location.pathname});return}Promise.all([fetch('/api/auth/status',{credentials:'same-origin',cache:'no-store'}).then(value=>value.json()),fetch('/api/portfolio/inclusion',{credentials:'same-origin',cache:'no-store'}).then(value=>value.json())]).then(([status,inclusion])=>done({path:location.pathname,status,inclusion}),error=>done({error:String(error)}))`, args: [] }),
    })
    if (recovered.path === '/portfolio' && recovered.status) break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.equal(recovered.path, '/portfolio', 'successful renewal resolves back to the saved portfolio')
  assert.equal(recovered.status.authenticated, true)
  assert.equal(recovered.status.reauthorizationRequired, false)
  assert.ok(recovered.inclusion.committed.length > 0, 'successful renewal preserves the committed account selection')
}

async function exerciseAccountInclusion(webdriver, sessionId, wiremockUrl, diagnosticScenario, onCommitted) {
  let inventory
  for (let attempt = 0; attempt < 40; attempt += 1) {
    inventory = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/inventory',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    if (inventory.status === 200 && inventory.body.state === 'ready') break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.equal(inventory?.body?.state, 'ready', 'masked inventory is ready before inclusion')

  let initial = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/inclusion',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(initial.status, 200)
  assert.equal(initial.cache, 'private, no-store')
  if (initial.body.committed.length > 0) {
    initial = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];const csrf=decodeURIComponent(document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'');fetch('/api/portfolio/inclusion',{method:'POST',credentials:'same-origin',cache:'no-store',headers:{'Content-Type':'application/json','X-CSRF-Token':csrf,'X-Inclusion-Version':String(${initial.body.version}),'Idempotency-Key':crypto.randomUUID()},body:JSON.stringify({accountIds:[]})}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    assert.equal(initial.status, 200, 'a repeated local run clears prior synthetic inclusion state')
  }
  assert.deepEqual(initial.body.committed, [], 'no account is included by default')

  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: 'http://127.0.0.1:8080/onboarding/accounts' }),
  })
  let chooserReady = false
  for (let attempt = 0; attempt < 40; attempt += 1) {
    chooserReady = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const text=document.body.innerText;const steps=document.querySelectorAll('.setup-progress li');return document.querySelector('h1')?.innerText==='Choose what Findur may use.'&&steps.length===3&&getComputedStyle(steps[0],'::after').borderLeftWidth==='1px'&&getComputedStyle(steps[2],'::after').borderLeftWidth==='0px'&&['Healthy Realtime — Full Data','Delayed Account — Empty Positions','Cash Only — No Positions','Positions — Refresh Failed','Balances — Initial Sync Pending','Activities — Provider Temporarily Unavailable','Unsupported Account Type','Secondary Investment Account','Unavailable because Findur supports investment accounts only'].every(value=>text.includes(value))`, args: [] }),
    })
    if (chooserReady) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(chooserReady, true, 'account chooser renders the healthy and degraded mixed-state demo')

  const selected = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const labels=['Healthy Realtime — Full Data','Delayed Account — Empty Positions','Cash Only — No Positions'];const inputs=[...document.querySelectorAll('.account-choice')].filter(value=>labels.some(label=>value.innerText.includes(label))).map(value=>value.querySelector('input[type="checkbox"]'));inputs.forEach(input=>input?.click());return inputs.length`, args: [] }),
  })
  assert.equal(selected, 3, 'the three rich default accounts can be selected without selecting every demo account')
  let reviewOpened = false
  for (let attempt = 0; attempt < 30; attempt += 1) {
    reviewOpened = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const button=[...document.querySelectorAll('button')].find(value=>value.innerText==='Review my choices');if(button&&!button.disabled){button.click();return true}return false`, args: [] }),
    })
    if (reviewOpened) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(reviewOpened, true, 'account choice review opens')
  const saved = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const button=[...document.querySelectorAll('button')].find(value=>value.innerText==='Save my choices');if(button){button.click();return true}return false`, args: [] }),
  })
  assert.equal(saved, true, 'account choice is saved through the UI')

  let preparingState
  for (let attempt = 0; attempt < 40; attempt += 1) {
    preparingState = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const text=document.body.innerText;return {path:location.pathname,accounts:document.querySelectorAll('.account h2').length,syncing:text.includes('Some portfolio data is still syncing'),unknown:text.includes('Findur could not determine why this data is unavailable')}`, args: [] }),
    })
    if (preparingState.path === '/portfolio' && preparingState.accounts === 3 && preparingState.syncing) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(
    { path: preparingState.path, accounts: preparingState.accounts, unknown: preparingState.unknown },
    { path: '/portfolio', accounts: 3, unknown: false },
    'saved accounts render immediately without an unknown diagnostic',
  )

  let degraded
  for (let attempt = 0; attempt < 360; attempt += 1) {
    degraded = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/showcase',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    const account = degraded?.body?.accounts?.find(({ label }) => label.startsWith('Healthy Realtime — Full Data'))
    if (account?.activities?.context?.diagnostic?.reason === 'provider_unavailable') break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  const degradedAccount = degraded?.body?.accounts?.find(({ label }) => label.startsWith('Healthy Realtime — Full Data'))
  assert.equal(degraded?.status, 200, `degraded Showcase remains readable: ${JSON.stringify(degraded)}`)
  assert.equal(degradedAccount?.activities?.context?.diagnostic?.reason, 'provider_unavailable', 'the worker persists the temporary activities failure')
  assert.equal(degradedAccount?.activities?.context?.diagnostic?.recommendedAction, 'retry')
  assert.ok(degradedAccount?.balances?.balances?.length > 0, 'successful balance evidence remains visible during the resource failure')
  assert.ok(degradedAccount?.positions?.positions?.length > 0, 'successful position evidence remains visible during the resource failure')
  assert.equal(degradedAccount?.balances?.context?.diagnostic, undefined, 'the balance resource remains healthy')
  assert.equal(degradedAccount?.positions?.context?.diagnostic, undefined, 'the positions resource remains healthy')

  await webdriver(`/session/${sessionId}/refresh`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })
  let degradedReloaded = false
  for (let attempt = 0; attempt < 80; attempt += 1) {
    degradedReloaded = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const text=document.body.innerText;return location.pathname==='/portfolio'&&text.includes('This data could not be refreshed')&&text.includes('FDR')`, args: [] }),
    })
    if (degradedReloaded) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(degradedReloaded, true, 'the retained evidence and resource diagnostic survive a browser reload')

  await removeDiagnosticScenario(wiremockUrl, diagnosticScenario.mappingID)

  let recovered
  for (let attempt = 0; attempt < 720; attempt += 1) {
    recovered = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/showcase',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    const account = recovered?.body?.accounts?.find(({ label }) => label.startsWith('Healthy Realtime — Full Data'))
    if (account?.activities?.activities?.length > 0 && account.activities.context.diagnostic === undefined) break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  const recoveredAccount = recovered?.body?.accounts?.find(({ label }) => label.startsWith('Healthy Realtime — Full Data'))
  assert.equal(recoveredAccount?.activities?.context?.diagnostic, undefined, 'the recovered resource diagnostic clears')
  assert.equal(recoveredAccount?.balances?.context?.diagnostic, undefined, 'recovery does not invent a balance diagnostic')
  assert.equal(recoveredAccount?.positions?.context?.diagnostic, undefined, 'recovery does not invent a positions diagnostic')
  assert.deepEqual(recoveredAccount?.balances?.balances, degradedAccount?.balances?.balances, 'retained balances remain unchanged through recovery')
  assert.deepEqual(recoveredAccount?.positions?.positions, degradedAccount?.positions?.positions, 'retained positions remain unchanged through recovery')

  let showcaseState
  for (let attempt = 0; attempt < 360; attempt += 1) {
    showcaseState = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const text=document.body.innerText;const accounts=[...document.querySelectorAll('.account h2')].map(value=>value.innerText);return {path:location.pathname,heading:document.querySelector('h1')?.innerText,focused:document.activeElement===document.querySelector('h1'),accounts:accounts.length===3&&['Healthy Realtime — Full Data','Delayed Account — Empty Positions','Cash Only — No Positions'].every(name=>accounts.some(value=>value.includes(name))),tables:document.querySelectorAll('.table-wrap').length,activity:text.includes('BUY'),expired:text.includes('has expired'),unknown:text.includes('could not determine why'),mvpNav:[...document.querySelectorAll('.authenticated-nav a')].map(value=>value.innerText).join('|')}`, args: [] }),
    })
    if (showcaseState.path === '/portfolio' && showcaseState.accounts && showcaseState.tables === 5 && showcaseState.activity) break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.deepEqual(showcaseState, { path: '/portfolio', heading: 'Your portfolio', focused: true, accounts: true, tables: 5, activity: true, expired: false, unknown: false, mvpNav: 'Portfolio|Profile|FAQ' }, 'the worker completes the Portfolio Showcase with current synthetic data and no unknown diagnostic')

  let inclusionState
  for (let attempt = 0; attempt < 360; attempt += 1) {
    inclusionState = await webdriver(`/session/${sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/inclusion',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
    })
    if (inclusionState.status === 200 && inclusionState.body.change?.status === 'committed') break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.equal(inclusionState?.status, 200, `the completed account inclusion remains readable: ${JSON.stringify(inclusionState)}`)
  assert.equal(inclusionState?.body.change?.status, 'committed', `the first account synchronization completes before measuring Showcase provider calls: ${JSON.stringify(inclusionState)}`)

  const showcase = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/showcase',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(showcase.status, 200)
  assert.equal(showcase.cache, 'private, no-store')
  assert.equal(showcase.body.accounts.length, 3)
  const individual = showcase.body.accounts.find(({ label }) => label.startsWith('Healthy Realtime — Full Data'))
  const ira = showcase.body.accounts.find(({ label }) => label.startsWith('Delayed Account — Empty Positions'))
  const cash = showcase.body.accounts.find(({ label }) => label.startsWith('Cash Only — No Positions'))
  assert.equal(individual?.balances.balances[0].cash, '25000', 'the self-directed cash balance survives the full stack')
  assert.equal(individual?.balances.balances[0].buyingPower, '50000', 'the supplied buying power survives the full stack')
  assert.equal(individual?.positions.positions[0].units, '10.50000001', 'exact position precision survives the full stack')
  assert.equal(individual?.activities.activities[0].amount, '-123.45', 'bounded activities reach the Showcase')
  assert.equal(ira?.balances.balances[0].cash, '12500', 'the delayed empty-positions synthetic is fully selectable')
  assert.deepEqual(ira?.positions.positions, [], 'the delayed account empty positions dataset remains complete')
  assert.equal(cash?.balances.balances[0].cash, '5000', 'the cash-only synthetic is fully selectable')
  assert.deepEqual(cash?.activities.activities, [], 'the cash-only empty activities dataset remains complete')
  assert.doesNotMatch(JSON.stringify(showcase.body), /03867fbb|7e7dcb86|50bb0405|synthetic-access-token/, 'Showcase response omits IDs, raw payload fields, and credentials')

  const requestsBeforeReload = await (await fetch(`${wiremockUrl}/__admin/requests`, { signal: AbortSignal.timeout(15_000) })).json()
  const inventoryReadsBeforeReload = requestsBeforeReload.requests.filter(({ request }) => request.url === '/authorizations' || request.url === '/accounts').length
  await webdriver(`/session/${sessionId}/refresh`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })
  for (let attempt = 0; attempt < 40; attempt += 1) {
    const ready = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return location.pathname==='/portfolio'&&document.querySelectorAll('.table-wrap').length===5`, args: [] }),
    })
    if (ready) break
    await new Promise((resolve) => setTimeout(resolve, 150))
  }
  const requestsAfterReload = await (await fetch(`${wiremockUrl}/__admin/requests`, { signal: AbortSignal.timeout(15_000) })).json()
  const inventoryReadsAfterReload = requestsAfterReload.requests.filter(({ request }) => request.url === '/authorizations' || request.url === '/accounts').length
  assert.equal(inventoryReadsAfterReload, inventoryReadsBeforeReload, 'rendering and reloading the Showcase does not refresh account inventory')

  if (onCommitted) await onCommitted()

  let showcaseRestored = false
  for (let attempt = 0; attempt < 40; attempt += 1) {
    showcaseRestored = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return location.pathname==='/portfolio'&&document.querySelectorAll('.table-wrap').length===5`, args: [] }),
    })
    if (showcaseRestored) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(showcaseRestored, true, 'the completed-session check returns to the rendered Portfolio Showcase')

  for (const width of [1440, 390, 320]) {
    await webdriver(`/session/${sessionId}/window/rect`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ width, height: 900 }),
    })
    const layout = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const table=document.querySelector('.table-wrap');const summary=document.querySelector('.dataset-summary');return {innerWidth,documentWidth:document.documentElement.scrollWidth,tableClientWidth:table.clientWidth,tableScrollWidth:table.scrollWidth,summaryLeft:summary.getBoundingClientRect().left,summaryRight:summary.getBoundingClientRect().right}`, args: [] }),
    })
    assert.ok(layout.documentWidth <= layout.innerWidth, `${width}px has no page-level horizontal scrolling`)
    assert.ok(layout.summaryLeft >= 0 && layout.summaryRight <= layout.innerWidth, `${width}px keeps the table summary in the readable flow`)
    if (width <= 390) {
      assert.ok(layout.tableScrollWidth > layout.tableClientWidth, `${width}px confines horizontal overflow to the evidence table`)
    }
  }

  const switchedToFrench = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const control=[...document.querySelectorAll('label.choice')].find(value=>value.innerText==='FR');control?.click();return Boolean(control)`, args: [] }),
  })
  assert.equal(switchedToFrench, true)
  const frenchLayout = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `return {heading:document.querySelector('h1')?.innerText,innerWidth,documentWidth:document.documentElement.scrollWidth}`, args: [] }),
  })
  assert.equal(frenchLayout.heading, 'Votre portefeuille')
  assert.ok(frenchLayout.documentWidth <= frenchLayout.innerWidth, 'French copy reflows without page-level horizontal scrolling')

  const editOpened = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const button=[...document.querySelectorAll('button')].find(value=>value.innerText.includes('Modifier les comptes inclus'));button?.click();return Boolean(button)`, args: [] }),
  })
  assert.equal(editOpened, true)
  let editor
  for (let attempt = 0; attempt < 40; attempt += 1) {
    editor = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const checked=[...document.querySelectorAll('.account-choice input[type="checkbox"]')].filter(value=>value.checked).length;return {path:location.pathname,checked}`, args: [] }),
    })
    if (editor.path === '/portfolio/accounts' && editor.checked === 3) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(editor, { path: '/portfolio/accounts', checked: 3 }, 'Edit included accounts opens the separate editor with all saved choices')

  const editReviewed = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const button=[...document.querySelectorAll('button')].find(value=>value.innerText==='Vérifier mes choix');button?.click();return Boolean(button)`, args: [] }),
  })
  assert.equal(editReviewed, true, 'saved account choices can be reviewed from the separate editor')
  let editSaved = false
  for (let attempt = 0; attempt < 30; attempt += 1) {
    editSaved = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const button=[...document.querySelectorAll('button')].find(value=>value.innerText==='Enregistrer mes choix');if(button){button.click();return true}return false`, args: [] }),
    })
    if (editSaved) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(editSaved, true, 'saved account choices can be confirmed from the editor')
  let editReturn
  for (let attempt = 0; attempt < 80; attempt += 1) {
    editReturn = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return {path:location.pathname,heading:document.querySelector('h1')?.innerText,accounts:document.querySelectorAll('.account h2').length}`, args: [] }),
    })
    if (editReturn.path === '/portfolio' && editReturn.accounts === 3) break
    await new Promise((resolve) => setTimeout(resolve, 150))
  }
  assert.deepEqual(editReturn, { path: '/portfolio', heading: 'Votre portefeuille', accounts: 3 }, 'a successful edit returns to the Portfolio Showcase')

  const result = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];(async()=>{const csrf=decodeURIComponent(document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'');const read=async()=>{const response=await fetch('/api/portfolio/inclusion',{credentials:'same-origin',cache:'no-store'});return {status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}};const send=async(version,key,accountIds)=>{const response=await fetch('/api/portfolio/inclusion',{method:'POST',credentials:'same-origin',cache:'no-store',headers:{'Content-Type':'application/json','X-CSRF-Token':csrf,'X-Inclusion-Version':String(version),'Idempotency-Key':key},body:JSON.stringify({accountIds})});return {status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}};const added=await read();const foreign=await send(added.body.version,crypto.randomUUID(),['00000000-0000-0000-0000-000000000001']);const removeKey=crypto.randomUUID();const removed=await send(added.body.version,removeKey,[]);const replay=await send(added.body.version,removeKey,[]);const stale=await send(${initial.body.version},crypto.randomUUID(),[]);done({added,foreign,removed,replay,stale})})().catch(error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(result.added.status, 200)
  assert.deepEqual(result.added.body.committed, ['03867fbb-41b4-4a05-8815-c96f94f8ba6b', '50bb0405-5efd-473f-a742-78a82bb1db53', '7e7dcb86-7d52-4f46-8fcf-91d5c9f81629'])
  assert.equal(result.foreign.status, 409)
  assert.equal(result.foreign.body.code, 'invalid_selection')
  assert.doesNotMatch(JSON.stringify(result.foreign.body), /00000000-0000-0000-0000-000000000001/)
  assert.equal(result.removed.status, 200)
  assert.deepEqual(result.removed.body.committed, [], 'removal is committed immediately')
  assert.deepEqual(result.replay.body, result.removed.body, 'identical idempotency replay has one effect')
  assert.equal(result.stale.status, 409)
  assert.equal(result.stale.body.code, 'conflict')
}

async function exercisePersonalProfile(webdriver, sessionId, origin) {
  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/profile` }),
  })
  let loaded
  for (let attempt = 0; attempt < 40; attempt += 1) {
    loaded = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const name=document.querySelector('#displayName');return {path:location.pathname,ready:!!name,name:name?.value??null,headingFocused:document.activeElement===document.querySelector('h1')}`, args: [] }),
    })
    if (loaded.ready && loaded.headingFocused) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(loaded.path, '/profile', 'profile loads directly')
  assert.equal(loaded.ready, true, 'profile form becomes ready')
  assert.equal(loaded.headingFocused, true, 'direct profile load focuses its heading')

  for (const [width, layout] of [[767, 'grid'], [768, 'rail']]) {
    await webdriver(`/session/${sessionId}/window/rect`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ width, height: 900 }),
    })
    const navigation = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const style=getComputedStyle(document.querySelector('.authenticated-nav'));return {position:style.position,direction:style.flexDirection,columns:style.gridTemplateColumns}`, args: [] }),
    })
    assert.equal(navigation.position, 'fixed', `${width}px profile navigation remains fixed`)
    if (layout === 'grid') assert.notEqual(navigation.columns, 'none', '767px profile uses bottom grid navigation')
    else assert.equal(navigation.direction, 'column', '768px profile uses rail navigation')
  }

  const initial = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/profile',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(initial.status, 200)
  assert.equal(initial.cache, 'private, no-store')

  if (!initial.body.profile) {
    assert.equal(loaded.name, '', 'new owner starts with a blank personal profile')
    await editAndSaveProfile(webdriver, sessionId, { displayName: '  Alex Journey  ', biography: '  Created through the composed browser journey.  ', locale: 'en', theme: 'system' })
    const created = await readProfile(webdriver, sessionId)
    assert.equal(created.body.profile.displayName, 'Alex Journey', 'create applies the canonical trimmed display name')
    assert.equal(created.body.profile.biography, 'Created through the composed browser journey.', 'create persists canonical biography text')
    assert.equal(created.body.profile.version, 1, 'first browser save creates version one')

    await webdriver(`/session/${sessionId}/url`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/profile` }),
    })
    const reloaded = await waitForProfileValue(webdriver, sessionId, 'Alex Journey')
    assert.equal(reloaded.biography, 'Created through the composed browser journey.', 'created profile survives a direct reload')
    assert.equal(reloaded.headingFocused, true, 'persisted profile reload preserves route focus')
  }

  const beforeEdit = await readProfile(webdriver, sessionId)
  await editAndSaveProfile(webdriver, sessionId, { displayName: 'Alex Journey Updated', biography: 'Edited and persisted through the composed browser journey.', locale: 'fr', theme: 'dark' })
  const edited = await readProfile(webdriver, sessionId)
  assert.equal(edited.body.profile.displayName, 'Alex Journey Updated')
  assert.equal(edited.body.profile.biography, 'Edited and persisted through the composed browser journey.')
  assert.equal(edited.body.profile.relationshipIntent, 'open-to-long-term')
  assert.equal(edited.body.profile.avatarKey, 'cedar')
  assert.equal(edited.body.profile.locale, 'fr')
  assert.equal(edited.body.profile.theme, 'dark')
  assert.equal(edited.body.profile.version, beforeEdit.body.profile.version + 1, 'edit advances the optimistic version exactly once')

  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/profile` }),
  })
  const persisted = await waitForProfileValue(webdriver, sessionId, 'Alex Journey Updated')
  assert.equal(persisted.biography, 'Edited and persisted through the composed browser journey.', 'edited profile survives reload')
}

async function exerciseFaq(webdriver, sessionId, origin) {
  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}/faq/` }),
  })
  let loaded
  for (let attempt = 0; attempt < 40; attempt += 1) {
    loaded = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const heading=document.querySelector('h1');return {path:location.pathname,heading:heading?.innerText,focused:document.activeElement===heading,title:document.title,nav:[...document.querySelectorAll('.authenticated-nav a')].map(link=>({text:link.innerText,current:link.getAttribute('aria-current')})),questions:document.querySelectorAll('.faq-list details').length}`, args: [] }),
    })
    if (loaded.path === '/faq' && loaded.focused) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(loaded.heading, 'Vos questions sur le portefeuille.', 'FAQ follows the saved French preference')
  assert.equal(loaded.title, 'FAQ — Findur', 'FAQ updates the document title')
  assert.equal(loaded.questions, 6, 'FAQ renders semantic disclosure topics')
  assert.deepEqual(loaded.nav.map(({ text }) => text), ['Portefeuille', 'Profil', 'FAQ'], 'FAQ follows Profile in authenticated navigation')
  assert.equal(loaded.nav[2].current, 'page', 'FAQ is the current destination')

  const summaryElement = await webdriver(`/session/${sessionId}/element`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ using: 'css selector', value: '.faq-list summary' }),
  })
  const summaryElementId = summaryElement['element-6066-11e4-a52e-4f735466cecf']
  await webdriver(`/session/${sessionId}/element/${summaryElementId}/value`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text: '\uE007', value: ['\uE007'] }),
  })
  const opened = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const disclosure=document.querySelector('.faq-list details');const summary=disclosure.querySelector('summary');const indicator=[...summary.querySelectorAll('.faq-disclosure-indicator span')].find(value=>getComputedStyle(value).visibility!=='hidden')?.innerText;return {open:disclosure.open,focused:document.activeElement===summary,text:disclosure.innerText,indicator}`, args: [] }),
  })
  assert.equal(opened.open, true, 'FAQ disclosure opens with native interaction')
  assert.equal(opened.focused, true, 'keyboard activation retains focus on the disclosure summary')
  assert.equal(opened.indicator, '−', 'FAQ disclosure shows its open indicator')
  assert.match(opened.text, /admissibilité sur le serveur/, 'FAQ explains server-enforced account eligibility in French')

  const localeSwitched = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const control=[...document.querySelectorAll('label.choice')].find(value=>value.innerText==='EN');control?.click();return Boolean(control)`, args: [] }),
  })
  assert.equal(localeSwitched, true)
  let localePreserved
  for (let attempt = 0; attempt < 40; attempt += 1) {
    localePreserved = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const disclosure=document.querySelector('.faq-list details');return {open:disclosure.open,question:disclosure.querySelector('h2')?.innerText}`, args: [] }),
    })
    if (localePreserved.question === 'Which accounts can I choose?') break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(localePreserved, { open: true, question: 'Which accounts can I choose?' }, 'locale switch preserves the open FAQ disclosure')
  const frenchRestored = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const control=[...document.querySelectorAll('label.choice')].find(value=>value.innerText==='FR');control?.click();return Boolean(control)`, args: [] }),
  })
  assert.equal(frenchRestored, true)
  let restoredQuestion
  for (let attempt = 0; attempt < 40; attempt += 1) {
    restoredQuestion = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script: `return document.querySelector('.faq-list h2')?.innerText`, args: [] }),
    })
    if (restoredQuestion === 'Quels comptes puis-je choisir?') break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(restoredQuestion, 'Quels comptes puis-je choisir?', 'French expanded content is restored before reflow checks')

  for (const [width, layout] of [[767, 'grid'], [390, 'grid'], [320, 'grid'], [768, 'rail'], [1440, 'rail']]) {
    await webdriver(`/session/${sessionId}/window/rect`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ width, height: 900 }),
    })
    const state = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const nav=document.querySelector('.authenticated-nav');const navStyle=getComputedStyle(nav);const links=[...nav.querySelectorAll('a')].map(link=>link.getBoundingClientRect());const page=document.querySelector('.faq-page').getBoundingClientRect();const list=document.querySelector('.faq-list').getBoundingClientRect();const summary=document.querySelector('.faq-list summary').getBoundingClientRect();const answer=document.querySelector('.faq-answer p').getBoundingClientRect();return {innerWidth,documentWidth:document.documentElement.scrollWidth,pageLeft:page.left,pageWidth:page.width,listLeft:list.left,listWidth:list.width,summaryLeft:summary.left,summaryRight:summary.right,answerLeft:answer.left,answerRight:answer.right,direction:navStyle.flexDirection,columns:navStyle.gridTemplateColumns,linkTops:links.map(link=>link.top),linkWidths:links.map(link=>link.width)}`, args: [] }),
    })
    assert.ok(state.documentWidth <= state.innerWidth, `${width}px FAQ has no page-level horizontal scrolling`)
    assert.ok(state.summaryLeft >= 0 && state.summaryRight <= state.innerWidth, `${width}px expanded French question stays in the viewport`)
    assert.ok(state.answerLeft >= 0 && state.answerRight <= state.innerWidth, `${width}px expanded French answer stays in the viewport`)
    if (layout === 'grid') {
      const columns = state.columns.split(' ')
      assert.equal(columns.length, 3, `${width}px FAQ uses exactly three mobile navigation columns`)
      assert.ok(Math.max(...state.linkWidths) - Math.min(...state.linkWidths) < 1, `${width}px FAQ mobile navigation columns are equal`)
      assert.equal(new Set(state.linkTops).size, 1, `${width}px FAQ mobile navigation remains in one row`)
    } else assert.equal(state.direction, 'column', `${width}px FAQ uses rail navigation`)
    if (width === 1440) {
      assert.ok(Math.abs(state.pageWidth - 1024) < 1, 'desktop FAQ uses the same 64rem page width as Portfolio and Profile')
      assert.ok(Math.abs(state.listWidth - 720) < 1, 'desktop FAQ keeps the shared 45rem reading width')
      assert.ok(Math.abs(state.pageLeft - state.listLeft) < 1, 'desktop FAQ reading content aligns to the shared page left edge')
    }
  }

  const profileOpened = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const link=[...document.querySelectorAll('.authenticated-nav a')].find(value=>value.getAttribute('href')==='/profile');link.click();return Boolean(link)`, args: [] }),
  })
  assert.equal(profileOpened, true)
  await webdriver(`/session/${sessionId}/back`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })
  let restored
  for (let attempt = 0; attempt < 40; attempt += 1) {
    restored = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return {path:location.pathname,focused:document.activeElement===document.querySelector('h1'),current:document.querySelector('.authenticated-nav a[aria-current="page"]')?.getAttribute('href')}`, args: [] }),
    })
    if (restored.path === '/faq' && restored.focused) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.deepEqual(restored, { path: '/faq', focused: true, current: '/faq' }, 'browser history restores FAQ focus and current navigation')
}

async function editAndSaveProfile(webdriver, sessionId, values) {
  await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const values=arguments[0];const set=(selector,value)=>{const element=document.querySelector(selector);const setter=Object.getOwnPropertyDescriptor(Object.getPrototypeOf(element),'value').set;setter.call(element,value);element.dispatchEvent(new Event('input',{bubbles:true}));element.dispatchEvent(new Event('change',{bubbles:true}))};set('#displayName',values.displayName);set('#locationKey','halifax-ns');set('#biography',values.biography);const adult=document.querySelector('#adultAttested');if(!adult.checked)adult.click();document.querySelector('input[name="relationshipIntent"][value="open-to-long-term"]').click();document.querySelector('input[name="avatar"][value="cedar"]').click();document.querySelector('input[name="profile-locale"][value="'+values.locale+'"]').click();document.querySelector('input[name="profile-theme"][value="'+values.theme+'"]').click();const save=document.querySelector('.profile-save-region button[type="submit"]');save.focus();save.click()`, args: [values] }),
  })
  let result
  for (let attempt = 0; attempt < 40; attempt += 1) {
    result = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const save=document.querySelector('.profile-save-region button[type="submit"]');return {saved:document.querySelector('.profile-status')?.textContent?.trim()||'',disabled:save?.disabled,focused:document.activeElement===save,name:document.querySelector('#displayName')?.value}`, args: [] }),
    })
    if (result.saved && !result.disabled) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.match(result.saved, /saved|enregistré/i, 'profile success is announced')
  assert.equal(result.focused, true, 'profile save success preserves button focus')
  assert.equal(result.name, values.displayName.trim(), 'the canonical response replaces the submitted draft')
}

async function readProfile(webdriver, sessionId) {
  return await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/profile',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
  })
}

async function waitForProfileValue(webdriver, sessionId, displayName) {
  let state
  for (let attempt = 0; attempt < 40; attempt += 1) {
    state = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return {name:document.querySelector('#displayName')?.value,biography:document.querySelector('#biography')?.value,headingFocused:document.activeElement===document.querySelector('h1')}`, args: [] }),
    })
    if (state.name === displayName && state.headingFocused) return state
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(state?.name, displayName, 'profile value appears after direct load')
  assert.equal(state?.headingFocused, true, 'profile heading receives focus after direct load')
  return state
}

export async function verifyBrowserSession({ browserUrl, publicOrigin, wiremockUrl, diagnosticScenario }) {
  await withWebDriverSession(browserUrl, async (first) => {
    await establishSession(first.webdriver, first.sessionId, publicOrigin)
    await verifyIncompleteActiveSessionBypassesConsent(first.webdriver, first.sessionId, publicOrigin)
    const seeded = await first.webdriver(`/session/${first.sessionId}/execute/async`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const done=arguments[arguments.length-1];localStorage.setItem('findur-locale','en');localStorage.setItem('findur-theme','dark');localStorage.setItem('protected','secret');sessionStorage.setItem('protected','secret');Promise.all([caches.open('findur-protected').then(cache=>cache.put('/protected',new Response('secret'))),new Promise((resolve,reject)=>{const request=indexedDB.open('findur-protected',1);request.onupgradeneeded=()=>request.result.createObjectStore('protected');request.onerror=()=>reject(request.error);request.onsuccess=()=>{request.result.close();resolve()}})]).then(()=>done(true),error=>done(String(error)))`, args: [] }),
    })
    assert.equal(seeded, true, 'protected browser stores were seeded')

    await withWebDriverSession(browserUrl, async (second) => {
      await establishSession(second.webdriver, second.sessionId, publicOrigin)
      await first.webdriver(`/session/${first.sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `document.querySelector('.logout-control button').click()`, args: [] }),
      })
      let firstState
      for (let attempt = 0; attempt < 40; attempt += 1) {
        firstState = await first.webdriver(`/session/${first.sessionId}/execute/sync`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ script: `return {path:location.pathname,locale:localStorage.getItem('findur-locale'),theme:localStorage.getItem('findur-theme'),protectedLocal:localStorage.getItem('protected'),protectedSession:sessionStorage.getItem('protected')}`, args: [] }),
        })
        if (firstState.path === '/') break
        await new Promise((resolve) => setTimeout(resolve, 100))
      }
      assert.deepEqual(firstState, { path: '/', locale: 'en', theme: 'dark', protectedLocal: null, protectedSession: null })
      const browserStores = await first.webdriver(`/session/${first.sessionId}/execute/async`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ script: `const done=arguments[arguments.length-1];Promise.all([caches.keys(),indexedDB.databases()]).then(([cacheNames,databases])=>done({cacheNames,databaseNames:databases.map(database=>database.name)}),error=>done({error:String(error)}))`, args: [] }),
      })
      assert.deepEqual(browserStores, { cacheNames: [], databaseNames: [] }, 'Cache Storage and IndexedDB contain no protected data after logout')
      const firstCookies = await first.webdriver(`/session/${first.sessionId}/cookie`)
      assert.ok(!firstCookies.some((cookie) => cookie.name === 'findur_session' || cookie.name === 'findur_csrf'), 'logout expires both session cookies')

      await second.webdriver(`/session/${second.sessionId}/url`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${publicOrigin}/portfolio` }),
      })
      let otherSessionHeading
      for (let attempt = 0; attempt < 40; attempt += 1) {
        otherSessionHeading = await second.webdriver(`/session/${second.sessionId}/execute/sync`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script: `return document.querySelector('h1')?.innerText`, args: [] }),
        })
        if (otherSessionHeading) break
        await new Promise((resolve) => setTimeout(resolve, 100))
      }
      assert.equal(otherSessionHeading, 'Your portfolio', 'logging out one browser preserves another active session')

      for (const [width, expectedPosition] of [[767, 'fixed'], [768, 'fixed']]) {
        await second.webdriver(`/session/${second.sessionId}/window/rect`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ width, height: 900 }),
        })
        const layout = await second.webdriver(`/session/${second.sessionId}/execute/sync`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script: `const nav=document.querySelector('.authenticated-nav');const style=getComputedStyle(nav);return {position:style.position,direction:style.flexDirection,columns:style.gridTemplateColumns}`, args: [] }),
        })
        assert.equal(layout.position, expectedPosition)
        if (width === 767) assert.notEqual(layout.columns, 'none', '767px uses bottom grid navigation')
        else assert.equal(layout.direction, 'column', '768px uses rail navigation')
      }
      const desktopHeader = await second.webdriver(`/session/${second.sessionId}/execute/sync`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script: `return getComputedStyle(document.querySelector('.authenticated-header')).position`, args: [] }),
      })
      assert.equal(desktopHeader, 'sticky', 'desktop authenticated header remains visible above the fixed rail')
      await exerciseAccountInclusion(second.webdriver, second.sessionId, wiremockUrl, diagnosticScenario, async () => {
        await verifyCompletedActiveSessionBypassesConsent(second.webdriver, second.sessionId, publicOrigin)
        await exerciseAuthorizationRenewal(second.webdriver, second.sessionId, publicOrigin, wiremockUrl)
      })
      await exercisePersonalProfile(second.webdriver, second.sessionId, publicOrigin)
      await exerciseFaq(second.webdriver, second.sessionId, publicOrigin)
      await exerciseLargeInventory(second.webdriver, second.sessionId, publicOrigin, wiremockUrl)
      await exerciseInventoryFixtures(second.webdriver, second.sessionId, wiremockUrl)
    })
  })
}
