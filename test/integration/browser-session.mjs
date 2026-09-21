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

async function exerciseAccountInclusion(webdriver, sessionId, wiremockUrl) {
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
      body: JSON.stringify({ script: `const text=document.body.innerText;const steps=document.querySelectorAll('.setup-progress li');return document.querySelector('h1')?.innerText==='Choose what Findur may use.'&&steps.length===3&&getComputedStyle(steps[0],'::after').borderLeftWidth==='1px'&&getComputedStyle(steps[2],'::after').borderLeftWidth==='0px'&&text.includes('Individual')&&text.includes('IRA')&&text.includes('Cash Account')`, args: [] }),
    })
    if (chooserReady) break
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  assert.equal(chooserReady, true, 'account chooser is ready with the three-step initial progress rail')

  const selected = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const input=document.querySelector('.select-all input[type="checkbox"]');input?.click();return Boolean(input)`, args: [] }),
  })
  assert.equal(selected, true, 'all three default synthetic accounts can be selected')
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

  let showcaseState
  for (let attempt = 0; attempt < 80; attempt += 1) {
    showcaseState = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const accounts=[...document.querySelectorAll('.account h2')].map(value=>value.innerText);return {path:location.pathname,heading:document.querySelector('h1')?.innerText,focused:document.activeElement===document.querySelector('h1'),accounts:accounts.length===3&&['Individual','IRA','Cash Account'].every(name=>accounts.some(value=>value.includes(name))),tables:document.querySelectorAll('.table-wrap').length,mvpNav:[...document.querySelectorAll('.authenticated-nav a')].map(value=>value.innerText).join('|')}`, args: [] }),
    })
    if (showcaseState.path === '/portfolio' && showcaseState.accounts && showcaseState.tables === 5) break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.deepEqual(showcaseState, { path: '/portfolio', heading: 'Your portfolio', focused: true, accounts: true, tables: 5, mvpNav: 'Portfolio|Profile' }, 'initial save opens the focused Portfolio Showcase with only the MVP navigation')

  const showcase = await webdriver(`/session/${sessionId}/execute/async`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const done=arguments[arguments.length-1];fetch('/api/portfolio/showcase',{credentials:'same-origin',cache:'no-store'}).then(async response=>done({status:response.status,cache:response.headers.get('cache-control'),body:await response.json()}),error=>done({error:String(error)}))`, args: [] }),
  })
  assert.equal(showcase.status, 200)
  assert.equal(showcase.cache, 'private, no-store')
  assert.equal(showcase.body.accounts.length, 3)
  const individual = showcase.body.accounts.find(({ label }) => label.startsWith('Individual'))
  const ira = showcase.body.accounts.find(({ label }) => label.startsWith('IRA'))
  const cash = showcase.body.accounts.find(({ label }) => label.startsWith('Cash Account'))
  assert.equal(individual?.balances.balances[0].cash, '25000', 'the self-directed cash balance survives the full stack')
  assert.equal(individual?.balances.balances[0].buyingPower, '50000', 'the supplied buying power survives the full stack')
  assert.equal(individual?.positions.positions[0].units, '10.50000001', 'exact position precision survives the full stack')
  assert.equal(individual?.activities.activities[0].amount, '-123.45', 'bounded activities reach the Showcase')
  assert.equal(ira?.balances.balances[0].cash, '12500', 'the IRA synthetic is fully selectable')
  assert.deepEqual(ira?.positions.positions, [], 'the IRA empty positions dataset remains complete')
  assert.equal(cash?.balances.balances[0].cash, '5000', 'the Cash Account synthetic is fully selectable')
  assert.deepEqual(cash?.activities.activities, [], 'the Cash Account empty activities dataset remains complete')
  assert.doesNotMatch(JSON.stringify(showcase.body), /03867fbb|7e7dcb86|50bb0405|synthetic-access-token/, 'Showcase response omits IDs, raw payload fields, and credentials')

  const requestsBeforeReload = await (await fetch(`${wiremockUrl}/__admin/requests`, { signal: AbortSignal.timeout(15_000) })).json()
  const providerReadsBeforeReload = requestsBeforeReload.requests.filter(({ request }) => request.url.startsWith('/accounts/')).length
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
  const providerReadsAfterReload = requestsAfterReload.requests.filter(({ request }) => request.url.startsWith('/accounts/')).length
  assert.equal(providerReadsAfterReload, providerReadsBeforeReload, 'rendering and reloading the Showcase makes zero provider calls')

  for (const width of [1440, 390, 320]) {
    await webdriver(`/session/${sessionId}/window/rect`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ width, height: 900 }),
    })
    const layout = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const table=document.querySelector('.table-wrap');const summary=document.querySelector('.dataset-summary');return {innerWidth,documentWidth:document.documentElement.scrollWidth,tableClientWidth:table.clientWidth,tableScrollWidth:table.scrollWidth,summaryLeft:summary.getBoundingClientRect().left,summaryRight:summary.getBoundingClientRect().right,datasetDirection:getComputedStyle(document.querySelector('.dataset>header')).flexDirection}`, args: [] }),
    })
    assert.ok(layout.documentWidth <= layout.innerWidth, `${width}px has no page-level horizontal scrolling`)
    assert.ok(layout.summaryLeft >= 0 && layout.summaryRight <= layout.innerWidth, `${width}px keeps the table summary in the readable flow`)
    if (width <= 390) {
      assert.ok(layout.tableScrollWidth > layout.tableClientWidth, `${width}px confines horizontal overflow to the evidence table`)
      assert.equal(layout.datasetDirection, 'column', `${width}px stacks dataset headings and freshness`)
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
      await exerciseAccountInclusion(second.webdriver, second.sessionId, wiremockUrl)
      await exercisePersonalProfile(second.webdriver, second.sessionId, publicOrigin)
      await exerciseInventoryFixtures(second.webdriver, second.sessionId, wiremockUrl)
    })
  })
}
