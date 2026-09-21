import assert from 'node:assert/strict'
import { withWebDriverSession } from './webdriver.mjs'

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

async function exerciseInventoryFixtures(webdriver, sessionId, wiremockUrl) {
  const fixtures = [
    { name: 'empty', status: 200, bodyFileName: 'inventory/empty.json', want: 'empty' },
    { name: 'disabled', status: 200, bodyFileName: 'inventory/disabled.json', want: 'disabled' },
    { name: 'unauthorized', status: 401, bodyFileName: 'inventory/unauthorized.json', want: 'unauthorized' },
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
      for (const forbidden of ['15363.23', 'RAW-SYMBOL-DO-NOT-RENDER', 'synthetic-access-token']) assert.doesNotMatch(serialized, new RegExp(forbidden), `${fixture.name} response is minimized`)
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

async function exerciseAccountInclusion(webdriver, sessionId) {
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

  const result = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];(async()=>{const csrf=decodeURIComponent(document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'');const read=async()=>{const response=await fetch('/api/portfolio/inclusion',{credentials:'same-origin',cache:'no-store'});return {status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}};const initial=await read();const send=async(version,key,accountIds)=>{const response=await fetch('/api/portfolio/inclusion',{method:'POST',credentials:'same-origin',cache:'no-store',headers:{'Content-Type':'application/json','X-CSRF-Token':csrf,'X-Inclusion-Version':String(version),'Idempotency-Key':key},body:JSON.stringify({accountIds})});return {status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}};const addKey=crypto.randomUUID();const added=await send(initial.body.version,addKey,['917c8734-8470-4a3e-a18f-57c3f2ee6631']);const replay=await send(initial.body.version,addKey,['917c8734-8470-4a3e-a18f-57c3f2ee6631']);const stale=await send(initial.body.version,crypto.randomUUID(),[]);const foreign=await send(added.body.version,crypto.randomUUID(),['00000000-0000-0000-0000-000000000001']);const removed=await send(added.body.version,crypto.randomUUID(),[]);done({initial,added,replay,stale,foreign,removed})})().catch(error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(result.initial.status, 200)
  assert.equal(result.initial.cache, 'private, no-store')
  assert.deepEqual(result.initial.body.committed, [], 'no account is included by default')
  assert.equal(result.added.status, 200)
  assert.equal(result.added.body.change.status, 'committed')
  assert.deepEqual(result.added.body.committed, ['917c8734-8470-4a3e-a18f-57c3f2ee6631'])
  assert.deepEqual(result.replay.body, result.added.body, 'identical idempotency replay has one effect')
  assert.equal(result.stale.status, 409)
  assert.equal(result.stale.body.code, 'conflict')
  assert.equal(result.foreign.status, 409)
  assert.equal(result.foreign.body.code, 'invalid_selection')
  assert.doesNotMatch(JSON.stringify(result.foreign.body), /00000000-0000-0000-0000-000000000001/)
  assert.equal(result.removed.status, 200)
  assert.deepEqual(result.removed.body.committed, [], 'removal is committed immediately')
}

export async function verifyBrowserSession({ browserUrl, publicOrigin, wiremockUrl }) {
  await withWebDriverSession(browserUrl, async (first) => {
    await establishSession(first.webdriver, first.sessionId, publicOrigin)
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
      assert.equal(otherSessionHeading, 'Your portfolio profile is taking shape.', 'logging out one browser preserves another active session')

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
      await exerciseAccountInclusion(second.webdriver, second.sessionId)
      await exerciseInventoryFixtures(second.webdriver, second.sessionId, wiremockUrl)
    })
  })
}
