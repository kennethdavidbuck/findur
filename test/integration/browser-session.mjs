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
    if (state.path === '/portfolio') break
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  assert.deepEqual(state, { path: '/portfolio', heading: 'Choose accounts before anything else.', focused: true })
}

export async function verifyBrowserSession({ browserUrl, publicOrigin }) {
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
      let secondHeading
      for (let attempt = 0; attempt < 40; attempt += 1) {
        secondHeading = await second.webdriver(`/session/${second.sessionId}/execute/sync`, {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ script: `return document.querySelector('h1')?.innerText`, args: [] }),
        })
        if (secondHeading) break
        await new Promise((resolve) => setTimeout(resolve, 100))
      }
      assert.equal(secondHeading, 'Choose accounts before anything else.', 'another browser remains authenticated')

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
    })
  })
}
