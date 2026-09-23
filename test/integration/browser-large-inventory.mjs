import assert from 'node:assert/strict'
import { largeInventoryFixture } from './large-inventory-fixture.mjs'

// These generous smoke budgets catch hangs/regressions, not production latency SLOs.
const loadBudgetMs = 10_000
const interactionBudgetMs = 5_000

export async function exerciseLargeInventory(webdriver, sessionId, origin, wiremockUrl) {
  const fixture = largeInventoryFixture()
  const { expected } = fixture
  const execute = (script, args = [], async = false) => webdriver(`/session/${sessionId}/execute/${async ? 'async' : 'sync'}`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script, args }),
  })
  const navigate = (path) => webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: `${origin}${path}` }),
  })
  const journal = async () => {
    const response = await fetch(`${wiremockUrl}/__admin/requests`, { signal: AbortSignal.timeout(15_000) })
    assert.ok(response.ok)
    return (await response.json()).requests
  }
  const retry = () => execute(`
    const done=arguments[arguments.length-1];const started=performance.now();
    const csrf=decodeURIComponent(document.cookie.split('; ').find(value=>value.startsWith('findur_csrf='))?.split('=',2)[1]||'');
    fetch('/api/portfolio/inventory/retry',{method:'POST',credentials:'same-origin',headers:{'X-CSRF-Token':csrf}})
      .then(async response=>done({status:response.status,body:await response.json(),durationMs:performance.now()-started}),error=>done({error:String(error)}));
  `, [], true)
  const rowState = `const rows=[...document.querySelectorAll('.account-choice input')];const all=document.querySelector('.select-all input');return {rows:rows.length,disabled:rows.filter(row=>row.disabled).length,checked:rows.filter(row=>row.checked).length,allChecked:!!all?.checked,indeterminate:!!all?.indeterminate,elapsedMs:performance.now()}`
  const interact = (script) => execute(`
    const done=arguments[arguments.length-1];const started=performance.now();
    ${script}
    requestAnimationFrame(()=>requestAnimationFrame(()=>done({durationMs:performance.now()-started})));
  `, [], true)
  const mappings = []
  let primaryError
  try {
    for (const [path, jsonBody] of [['/authorizations', fixture.connections], ['/accounts', fixture.accounts]]) {
      const response = await fetch(`${wiremockUrl}/__admin/mappings`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: AbortSignal.timeout(15_000),
        body: JSON.stringify({ priority: 1, request: { method: 'GET', urlPath: path }, response: { status: 200, headers: { 'Content-Type': 'application/json' }, jsonBody: path === '/accounts' ? fixture.accounts.map((account) => ({ ...account, account_category: 'DEPOSIT' })) : jsonBody } }),
      })
      assert.equal(response.status, 201, `large ${path} fixture registered`)
      mappings.push((await response.json()).id)
    }
    // Prove unsupported rows remain visible but cannot be selected through the
    // real provider, persistence, and UI before restoring the mixed fixture.
    const passive = await retry()
    assert.equal(passive.status, 200)
    assert.equal(passive.body.state, 'ready', 'unsupported accounts publish passive explanatory rows')
    assert.equal(passive.body.connections.length, expected.connections)
    const passiveAccounts = passive.body.connections.flatMap((connection) => connection.accounts)
    assert.equal(passiveAccounts.length, expected.passiveOnlyVisible, 'closed, unavailable, and pending rows remain excluded')
    assert.equal(passiveAccounts.filter((account) => account.selectable).length, 0)
    assert.ok(passiveAccounts.every((account) => account.usabilityReason === 'unsupported_category' || account.usabilityReason === 'connection_disabled'))
    await navigate('/portfolio/accounts')
    let passiveRendered
    for (let attempt = 0; attempt < 80; attempt += 1) {
      passiveRendered = await execute(rowState)
      if (passiveRendered.rows === expected.passiveOnlyVisible) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(passiveRendered.rows, expected.passiveOnlyVisible, 'selector explains every permitted passive unsupported account')
    assert.equal(passiveRendered.disabled, expected.passiveOnlyVisible, 'crafted selection is unavailable in the browser')
    const restoredMapping = await fetch(`${wiremockUrl}/__admin/mappings/${mappings[1]}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' }, signal: AbortSignal.timeout(15_000),
      body: JSON.stringify({ id: mappings[1], priority: 1, request: { method: 'GET', urlPath: '/accounts' }, response: { status: 200, headers: { 'Content-Type': 'application/json' }, jsonBody: fixture.accounts } }),
    })
    assert.equal(restoredMapping.status, 200, 'mixed accounts replace the filtered-empty scenario')
    const baseline = new Set((await journal()).map((event) => event.id))
    const loaded = await retry()
    assert.equal(loaded.status, 200, 'large inventory explicit retry succeeds')
    assert.equal(loaded.body.state, 'ready', 'large inventory has no truncation or malformed response')
    assert.equal(loaded.body.connections.length, expected.connections)
    const normalized = loaded.body.connections.flatMap((connection) => connection.accounts)
    assert.equal(normalized.length, expected.normalized, 'only usable investment accounts survive normalization and persistence')
    assert.equal(new Set(normalized.map((account) => account.id)).size, expected.normalized, 'every normalized account remains distinct')
    assert.equal(normalized.filter((account) => account.eligible).length, expected.eligible)
    assert.equal(normalized.filter((account) => account.selectable).length, expected.selectable)
    assert.deepEqual(normalized.map((account) => account.id).sort(), [...expected.accountIds].sort(), 'exact selectable and passive account set matches the independent fixture oracle')
    assert.deepEqual(normalized.filter((account) => account.selectable).map((account) => account.id).sort(), [...expected.selectableAccountIds].sort(), 'exact selectable account set matches the independent fixture oracle')
    const disabledConnection = loaded.body.connections.find((connection) => connection.status === 'disabled')
    assert.equal(disabledConnection?.accounts.length, 20, 'disabled connection accounts remain visible as passive rows')
    assert.ok(disabledConnection.accounts.every((account) => !account.selectable && account.usabilityReason === 'connection_disabled'))
    for (const connection of fixture.connections) {
      const actual = loaded.body.connections.find((candidate) => candidate.id === connection.id)
      assert.equal(actual?.brokerageLabel, connection.brokerage.display_name, 'brokerage label stays attached to its connection')
      const expectedIds = fixture.accounts.filter((account) => account.brokerage_authorization === connection.id && expected.accountIds.includes(account.id)).map((account) => account.id)
      assert.deepEqual(actual.accounts.map((account) => account.id).sort(), expectedIds.sort(), `account ownership preserved for ${connection.brokerage.display_name}`)
    }
    for (const field of ['institution_account_id', 'balance', 'raw_data', 'synthetic-access-token']) {
      assert.doesNotMatch(JSON.stringify(loaded.body), new RegExp(field), `large inventory omits provider field ${field}`)
    }
    assert.ok(loaded.durationMs < loadBudgetMs, `inventory fetch/normalize/persist took ${loaded.durationMs} ms`)

    await navigate('/portfolio/accounts')
    let rendered
    for (let attempt = 0; attempt < 80; attempt += 1) {
      rendered = await execute(rowState)
      if (rendered.rows === expected.visible) break
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
    assert.equal(rendered.rows, expected.visible, 'chooser renders every visible row, including the end of the list')
    assert.equal(rendered.disabled, expected.disabled, 'only passive unsupported and repair-required rows are disabled')
    assert.equal(expected.accounts - rendered.rows, expected.hidden, 'all excluded accounts are removed by the server')
    assert.equal(rendered.checked, 0, 'stress test begins with an empty draft')
    assert.ok(rendered.elapsedMs < loadBudgetMs, `document navigation and render took ${rendered.elapsedMs} ms`)
    const browserLabels = await execute(`return [...document.querySelectorAll('.account-choice')].map(row=>({account:row.querySelector('strong').textContent,brokerage:row.querySelector('small').textContent.split(' · ')[0]}))`)
    const visibleIds = new Set(expected.accountIds)
    const expectedLabels = fixture.accounts.filter((account) => visibleIds.has(account.id)).map((account) => ({ account: `${account.name} (•••• ${account.number.slice(-4)})`, brokerage: account.institution_name }))
    assert.equal(browserLabels.length, expectedLabels.length)
    for (const expectedLabel of expectedLabels) {
      const actual = browserLabels.find((row) => row.account === expectedLabel.account)
      assert.equal(actual?.brokerage, expectedLabel.brokerage, `browser brokerage is correct for ${expectedLabel.account}`)
    }

    const selected = await interact(`document.querySelector('.select-all input').click();`)
    assert.equal((await execute(rowState)).checked, 5, 'bulk selection stops at the explicit account limit')
    const deselected = await interact(`document.querySelector('.select-all input').click();`)
    assert.equal((await execute(rowState)).checked, 0, 'deselect all clears the large draft')
    const element = await execute(`const row=[...document.querySelectorAll('.account-choice input:not(:disabled)')].at(-1);row.scrollIntoView({block:'center',behavior:'instant'});return row;`)
    const toggleStarted = performance.now()
    await webdriver(`/session/${sessionId}/element/${element['element-6066-11e4-a52e-4f735466cecf']}/click`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })
    const toggled = await execute(rowState)
    const individualToggle = { durationMs: performance.now() - toggleStarted }
    assert.equal(toggled.checked, 1, 'last selectable account can be toggled')
    assert.equal(toggled.indeterminate, true, 'select all reflects the individual selection')
    await interact(`document.querySelector('.select-all input').click();`)
    const reviewed = await interact(`document.querySelector('.inclusion-actions button').click();`)
    assert.equal(await execute(`return document.querySelectorAll('.confirmation-dialog li').length`), 5, 'review lists only the bounded selection')
    await interact(`document.querySelector('.dialog-actions button').click();`)
    assert.equal((await execute(rowState)).checked, 5, 'closing review preserves the bounded draft')
    for (const [name, metric] of Object.entries({ selected, deselected, individualToggle, reviewed })) {
      assert.ok(metric.durationMs < interactionBudgetMs, `${name} took ${metric.durationMs} ms`)
    }

    const requests = (await journal()).filter((event) => !baseline.has(event.id)).map((event) => event.request.url)
    assert.deepEqual(requests.sort(), ['/accounts', '/authorizations'], '1000 accounts require exactly two provider requests, including UI review; no per-connection or financial calls')
    console.log(`large inventory performance: ${JSON.stringify({ accounts: expected.accounts, connections: expected.connections, selectable: expected.selectable, visible: expected.visible, providerRequests: requests.length, inventoryFetchAndPersistMs: Math.round(loaded.durationMs), navigationAndRenderMs: Math.round(rendered.elapsedMs), selectAllMs: Math.round(selected.durationMs), deselectAllMs: Math.round(deselected.durationMs), individualToggleMs: Math.round(individualToggle.durationMs), reviewMs: Math.round(reviewed.durationMs) })}`)

  } catch (error) {
    primaryError = error
  } finally {
    const cleanupErrors = []
    const cleanup = async (label, action) => {
      try { await action() } catch (error) { cleanupErrors.push(new Error(`${label}: ${error.message}`, { cause: error })) }
    }
    for (const id of mappings) {
      await cleanup(`remove mapping ${id}`, async () => {
        const response = await fetch(`${wiremockUrl}/__admin/mappings/${id}`, { method: 'DELETE', signal: AbortSignal.timeout(15_000) })
        assert.ok(response.ok, 'large inventory mapping removed')
      })
    }
    // Discard the unsaved browser draft and restore the normal rich mixed-state
    // snapshot. This also preserves the later categorical/rate-limit scenarios.
    await cleanup('discard browser draft', () => navigate('/profile'))
    await cleanup('restore normal inventory', async () => {
      const restored = await retry()
      assert.equal(restored.body?.state, 'ready', 'normal inventory restored after the stress test')
      assert.equal(restored.body.connections.flatMap((connection) => connection.accounts).length, 28)
    })
    if (primaryError) {
      for (const error of cleanupErrors) console.error(error)
    } else if (cleanupErrors.length) primaryError = new AggregateError(cleanupErrors, 'large inventory cleanup failed')
  }
  if (primaryError) throw primaryError
}
