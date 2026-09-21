import fs from 'node:fs/promises'
import { withWebDriverSession } from './webdriver.mjs'

const browserUrl = process.env.BROWSER_URL
const origin = 'http://127.0.0.1:8080'

await withWebDriverSession(browserUrl, async ({ webdriver, sessionId }) => {
  await webdriver(`/session/${sessionId}/window/rect`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ width: 1440, height: 1100 }),
  })
  await webdriver(`/session/${sessionId}/url`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url: `${origin}/connect` }),
  })
  let authorizationReady = false
  for (let attempt = 0; attempt < 50; attempt += 1) {
    const ready = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `const b=document.querySelector('form[action="/api/auth/snaptrade/authorize"] button'); return Boolean(b&&!b.disabled)`, args: [] }),
    })
    if (ready) {
      authorizationReady = true
      break
    }
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  if (!authorizationReady) throw new Error('authorization control did not become ready for visual capture')
  await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `document.querySelector('form[action="/api/auth/snaptrade/authorize"] button').click()`, args: [] }),
  })
  let chooserLoaded = false
  for (let attempt = 0; attempt < 80; attempt += 1) {
    const loaded = await webdriver(`/session/${sessionId}/execute/sync`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ script: `return location.pathname==='/onboarding/accounts' && Boolean(document.querySelector('.account-inclusion'))`, args: [] }),
    })
    if (loaded) {
      chooserLoaded = true
      break
    }
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  if (!chooserLoaded) throw new Error('account chooser did not become ready for visual capture')
  const desktopMetrics = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const h=document.querySelector('.portfolio-inventory h1');const g=document.querySelector('.connection-setup-grid');return {path:location.pathname,width:innerWidth,scrollWidth:document.documentElement.scrollWidth,h1:getComputedStyle(h).fontSize,grid:g.getBoundingClientRect().toJSON(),bodyText:document.body.innerText}`, args: [] }),
  })
  const desktop = await webdriver(`/session/${sessionId}/screenshot`)
  await fs.writeFile('/tmp/findur-chooser-desktop.png', Buffer.from(desktop, 'base64'))

  await webdriver(`/session/${sessionId}/window/rect`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ width: 390, height: 844 }),
  })
  await new Promise((resolve) => setTimeout(resolve, 300))
  const mobileMetrics = await webdriver(`/session/${sessionId}/execute/sync`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script: `const h=document.querySelector('.portfolio-inventory h1');const rail=document.querySelector('.setup-progress');const layers=document.querySelector('.setup-layers');return {width:innerWidth,scrollWidth:document.documentElement.scrollWidth,h1:getComputedStyle(h).fontSize,rail:getComputedStyle(rail).display,layers:getComputedStyle(layers).display}`, args: [] }),
  })
  const mobile = await webdriver(`/session/${sessionId}/screenshot`)
  await fs.writeFile('/tmp/findur-chooser-mobile.png', Buffer.from(mobile, 'base64'))
  console.log(JSON.stringify({ desktopMetrics, mobileMetrics }))
})
