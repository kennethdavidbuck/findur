import assert from 'node:assert/strict'
import { verifyBrowserOAuth } from './browser-oauth.mjs'
import { verifyBrowserSession } from './browser-session.mjs'
import { verifyBrowserStatus } from './browser-status.mjs'

const base = process.env.BASE_URL
const expectedSha = process.env.EXPECTED_SHA
const browser = process.env.BROWSER_URL
const wiremock = process.env.WIREMOCK_URL
const request = (path, init = {}) => fetch(`${base}${path}`, {
  ...init,
  signal: init.signal ?? AbortSignal.timeout(15_000),
})

const resetJournal = await fetch(`${wiremock}/__admin/requests`, { method: 'DELETE', signal: AbortSignal.timeout(15_000) })
assert.ok(resetJournal.ok, 'WireMock request journal reset for this run')

const health = await request('/api/healthz')
assert.equal(health.status, 200)
assert.deepEqual(await health.json(), { status: 'ok', buildSha: expectedSha })

const ready = await request('/api/readyz')
assert.equal(ready.status, 200)
assert.deepEqual(await ready.json(), { status: 'ready', buildSha: expectedSha })

const authorization = await request('/api/auth/snaptrade/authorize', {
  method: 'POST',
  redirect: 'manual',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ returnTo: '/portfolio' }),
})
assert.equal(authorization.status, 303)
assert.equal(authorization.headers.get('cache-control'), 'no-store')
const authorizationLocation = new URL(authorization.headers.get('location'))
assert.equal(authorizationLocation.origin, 'http://127.0.0.1:8080')
assert.equal(authorizationLocation.pathname, '/api/__fixture/oidc/authorize')
assert.equal(authorizationLocation.searchParams.get('scope'), 'openid read')
assert.equal(authorizationLocation.searchParams.get('redirect_uri'), 'http://127.0.0.1:8080/api/auth/snaptrade/callback')
assert.equal(authorizationLocation.searchParams.get('response_type'), 'code')
assert.doesNotMatch(authorizationLocation.toString(), /synthetic-oauth-client-secret/, 'client secret never enters the browser redirect')
assert.equal(authorizationLocation.searchParams.get('code_challenge_method'), 'S256')
for (const parameter of ['state', 'nonce', 'code_challenge']) assert.ok(authorizationLocation.searchParams.get(parameter))
const attemptCookie = authorization.headers.getSetCookie()[0]
assert.match(attemptCookie, /^findur_oauth_attempt=/)
assert.match(attemptCookie, /; Path=\/api\/auth\/snaptrade\/callback/)
const cookieMaxAge = Number(attemptCookie.match(/; Max-Age=(\d+)/)?.[1])
assert.ok(cookieMaxAge >= 1 && cookieMaxAge <= 600, `bounded cookie Max-Age: ${cookieMaxAge}`)
assert.match(attemptCookie, /; HttpOnly/)
assert.match(attemptCookie, /; Secure/)
assert.match(attemptCookie, /; SameSite=Lax/)
assert.doesNotMatch(attemptCookie, /Domain=/i)

const statusRoute = await request('/__status')
assert.equal(statusRoute.status, 200)
const html = await statusRoute.text()
assert.match(html, new RegExp(`<meta name="findur-build-sha" content="${expectedSha}"`))
const assetPath = html.match(/<script[^>]+src="([^"]+)"/)?.[1]
assert.ok(assetPath, 'SPA module asset is present')
const assetResponse = await request(assetPath)
assert.equal(assetResponse.ok, true, 'frontend module request succeeds')
assert.match(
  assetResponse.headers.get('content-type') || '',
  /^(application|text)\/(javascript|x-javascript|ecmascript)(?:;|$)/i,
  'frontend module has a JavaScript-compatible content type',
)
const asset = await assetResponse.text()
assert.doesNotMatch(asset, /<(!doctype|html|body)(?:[\s>])/i, 'frontend module is not an HTML fallback')
assert.ok(asset.includes(expectedSha), 'frontend bundle independently contains the expected SHA')

const fallback = await request('/not-a-real-public-route')
assert.match(await fallback.text(), /<div id="root"><\/div>/)

const unsafe = await request('/api/__fixture/proxy/unsafe-json', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ text: '<script>&"' }),
})
assert.equal(unsafe.status, 200)
assert.equal(unsafe.headers.get('cache-control'), 'no-store')

const callback = await request('/api/__fixture/proxy/callback?code=a%2Bb%2Fc%3D&state=synthetic-state')
assert.equal(callback.status, 204)

const cookieResponse = await request('/api/__fixture/proxy/cookies/set')
const cookies = cookieResponse.headers.getSetCookie()
assert.equal(cookies.length, 2)
assert.ok(cookies.every((cookie) => !/domain=/i.test(cookie)), 'cookies remain host-only')
const cookieHeader = cookies.map((cookie) => cookie.split(';', 1)[0]).join('; ')
const replay = await request('/api/__fixture/proxy/cookies/replay', { headers: { Cookie: cookieHeader } })
assert.equal(replay.status, 200)

const cached = await request('/api/__fixture/proxy/cached')
assert.equal(cached.headers.get('cache-control'), 'private, max-age=60')
assert.equal(cached.headers.get('etag'), 'synthetic-etag')

const failure = await request('/api/__fixture/proxy/failure')
assert.equal(failure.status, 429)
assert.equal(failure.headers.get('retry-after'), '7')

const provider = await request('/api/__fixture/provider/accounts')
assert.equal(provider.status, 200)
assert.deepEqual(await provider.json(), { accounts: [] })
assert.equal(
  (await request('/api/__fixture/provider/accounts', {
    headers: {
      Authorization: 'Bearer browser-token-must-be-ignored',
      clientId: 'forbidden',
      consumerKey: 'forbidden',
      userId: 'forbidden',
      userSecret: 'forbidden',
      timestamp: 'forbidden',
      Signature: 'forbidden',
    },
  })).status,
  200,
  'the backend replaces browser authentication with its server-held bearer and strips Commercial fields',
)
assert.equal(
  (await request('/api/__fixture/provider/not-allowlisted')).status,
  404,
  'provider paths are allowlisted',
)

await verifyBrowserStatus({ browserUrl: browser, publicOrigin: base, expectedSha })
await verifyBrowserOAuth({ browserUrl: browser, oauthOrigin: 'http://127.0.0.1:8080' })
await verifyBrowserSession({ browserUrl: browser, publicOrigin: 'http://127.0.0.1:8080', wiremockUrl: wiremock })

const journal = await (await fetch(`${wiremock}/__admin/requests`, { signal: AbortSignal.timeout(15_000) })).json()
const inventoryEvents = journal.requests
  .filter(({ request }) => request.url === '/authorizations' || request.url === '/authorizations/87b24961-b51e-4db8-9226-f198f6518a89/accounts')
const inventoryRequests = inventoryEvents.map(({ request }) => request)
assert.ok(inventoryRequests.filter(({ url }) => url === '/authorizations').length >= 7, 'success and every categorical fixture executed through WireMock')
assert.ok(inventoryRequests.some(({ url }) => url.endsWith('/accounts')), 'successful bootstrap reaches the scoped account operation')
for (let index = 0; index < inventoryRequests.length; index += 1) {
  if (inventoryRequests[index].url.endsWith('/accounts')) assert.equal(inventoryRequests[index - 1]?.url, '/authorizations', 'connections are requested before accounts')
}
for (const providerRequest of inventoryRequests) {
  assert.equal(providerRequest.headers.Authorization, 'Bearer synthetic-access-token')
  for (const forbidden of ['clientId', 'consumerKey', 'userId', 'userSecret', 'timestamp', 'Signature']) assert.equal(providerRequest.headers[forbidden], undefined)
}
const accountDataRequests = journal.requests
  .map(({ request }) => request)
  .filter(({ url }) => url.startsWith('/accounts/917c8734-8470-4a3e-a18f-57c3f2ee6631/'))
assert.deepEqual(
  accountDataRequests.map(({ url }) => new URL(url, 'http://wiremock').pathname).sort(),
  [
    '/accounts/917c8734-8470-4a3e-a18f-57c3f2ee6631/activities',
    '/accounts/917c8734-8470-4a3e-a18f-57c3f2ee6631/balances',
    '/accounts/917c8734-8470-4a3e-a18f-57c3f2ee6631/positions/all',
  ],
  'inclusion calls each required allowlisted dataset exactly once',
)
for (const providerRequest of accountDataRequests) {
  assert.equal(providerRequest.headers.Authorization, 'Bearer synthetic-access-token')
  for (const forbidden of ['clientId', 'consumerKey', 'userId', 'userSecret', 'timestamp', 'Signature']) assert.equal(providerRequest.headers[forbidden], undefined)
}

console.log('integration contracts passed')
