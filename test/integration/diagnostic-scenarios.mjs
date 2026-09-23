const jsonHeaders = { 'Content-Type': 'application/json' }
const accountIDPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/
const resourcePaths = {
  balances: 'balances',
  positions: 'positions/all',
  activities: 'activities',
}

export async function installFindurFailureScenario(wiremockURL, { accountID, resource, status = 503 }) {
  if (!accountIDPattern.test(accountID)) throw new Error('diagnostic scenario requires a synthetic UUID account ID')
  const resourcePath = resourcePaths[resource]
  if (!resourcePath) throw new Error(`unsupported diagnostic scenario resource: ${resource}`)
  if (!Number.isInteger(status) || status < 400 || status > 599) throw new Error('diagnostic scenario status must be an HTTP error')
  const response = await fetch(`${wiremockURL}/__admin/mappings`, {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({
      priority: 1,
      request: { method: 'GET', urlPath: `/accounts/${accountID}/${resourcePath}` },
      response: { status, headers: jsonHeaders, jsonBody: { ignored: 'browser-safe normalization discards this synthetic body' } },
      metadata: { findurDiagnosticScenario: true, accountID, resource },
    }),
  })
  if (!response.ok) throw new Error(`install diagnostic scenario: ${response.status}`)
  return (await response.json()).id
}

export async function removeDiagnosticScenario(wiremockURL, mappingID) {
  const response = await fetch(`${wiremockURL}/__admin/mappings/${mappingID}`, { method: 'DELETE' })
  if (!response.ok) throw new Error(`remove diagnostic scenario: ${response.status}`)
}

export async function resetDiagnosticScenarios(wiremockURL) {
  const response = await fetch(`${wiremockURL}/__admin/mappings/remove-by-metadata`, {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ matchesJsonPath: { expression: '$.findurDiagnosticScenario', equalTo: 'true' } }),
  })
  if (!response.ok) throw new Error(`reset diagnostic scenarios: ${response.status}`)
}
