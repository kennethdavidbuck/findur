---
title: 'Handle SnapTrade authorization recovery gracefully'
type: 'bugfix'
created: '2026-09-23'
status: 'done'
route: 'dispatch'
baseline_commit: '37bbc0ab09e13bffea6a11c15e4af36066fb067a'
review_loop_iteration: 0
context:
  - '_bmad-output/specs/spec-browser-safe-error-contract/SPEC.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur safely categorizes SnapTrade denial, invalid callbacks, revoked grants, and disabled connections, but the browser experience is incomplete: callback failures silently bounce through onboarding, authorization-start failures can show raw JSON, and an authenticated user's Reconnect action loops back into the app instead of starting recovery.

**Approach:** Keep recovery in the existing connect and portfolio surfaces. Explain the outcome in plain, reassuring language and offer exactly the useful next action. A proven app-wide SnapTrade authorization loss ends the current Findur session and returns the browser to Connect with a SnapTrade reconnect action. A disabled individual brokerage connection remains an in-session, connection-specific repair flow.

## Boundaries & Constraints

**Always:** Use bounded server categories and authoritative session/authorization state. English and French copy says what happened, what remains safe, and what to do next. End the current Findur session and expire its browser cookies when app-wide SnapTrade authorization is proven unusable; retain saved choices for a later successful login. Preserve callback single-use, state, nonce, PKCE, cookie, no-store, refresh, and safe-logging protections. Announce and focus blocking recovery accessibly.

**Never:** Render SnapTrade descriptions, OAuth terms, token language, provider codes, raw JSON, secrets, or financial data. Treat a disabled individual brokerage connection as an app-wide authorization loss, automatically retry an unsafe exchange, erase saved choices, or add a standalone error page in this change.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Consent declined | SnapTrade returns `access_denied` with valid state/binding | Return to Connect; say access was not granted and nothing changed; offer a clear retry and back action | Consume the attempt without token exchange; ignore provider detail |
| Callback cannot finish | Missing/expired/replayed correlation or processing failure | Return to Connect with a safe "couldn't finish" explanation and start-again action | Expire the attempt cookie; retain current compensation; hide the cause |
| App access revoked | Refresh or one-refresh/one-retry provider read proves authorization unusable while the Findur session is active | Revoke the current Findur session, expire session cookies, and return to Connect with a plain-language message and fresh hosted SnapTrade authorization action; retain saved choices | Do not show cached accounts or portfolio data, and do not describe this as an individual brokerage problem |
| Brokerage connection disabled | Inventory reports a disabled connection while app authorization remains valid | Explain that the brokerage connection needs attention in SnapTrade and link to its management surface | Do not start the Findur OAuth flow or promise automatic repair |
| Authorization cannot start | Hosted handoff is unavailable, malformed, or fails | Remain on Connect with localized retry/later guidance | Never navigate the browser to API JSON |
| Successful authorization | Initial or returning grant succeeds | Establish/rotate credentials and continue through the existing inclusion resolver | Existing success, replay, and preservation behavior remains unchanged |

</frozen-after-approval>

## Code Map

- `backend/api/openapi.yaml` -- authority for callback, start, and status contracts; regenerate both language bindings.
- `backend/internal/auth/{callback.go,credentials.go}` -- reuse terminal recovery and the existing one-refresh/one-retry `reauthorization-required` transition; add bounded browser/status outcomes only.
- `backend/internal/platform/postgres/oauth_attempts.go` -- expose renewal state and persist safe terminal routes without returning identity or sensitive detail.
- `backend/internal/platform/httpapi/authorization.go` -- give browser form/callback failures clean recovery redirects while retaining categorical JSON and logs for API callers.
- `frontend/src/{App.tsx,auth-status.ts,pages/ConsentPage.tsx,i18n.tsx,styles.css}` -- allowlist recovery hints, permit active-session renewal, and render accessible EN/FR guidance.
- `frontend/src/pages/{PortfolioPage.tsx,PortfolioShowcasePage.tsx}` -- separate revoked-grant renewal from disabled-connection management.
- Existing backend, frontend, and `test/integration/**` auth tests -- extend denial, unauthorized, redirect, active-session, and no-detail coverage.

## Tasks & Acceptance

**Execution:**
- [x] `backend/api/openapi.yaml`, `backend/internal/auth/callback.go`, `backend/internal/platform/postgres/oauth_attempts.go` -- add bounded recovery/status outcomes and regenerate bindings.
- [x] `backend/internal/platform/httpapi/authorization.go` -- route browser failures back to Findur without changing JSON API behavior or callback security.
- [x] `frontend/src/App.tsx`, `frontend/src/pages/{ConsentPage,PortfolioPage,PortfolioShowcasePage}.tsx` -- make active-session renewal reachable and send disabled-connection repair to SnapTrade management.
- [x] `frontend/src/i18n.tsx`, `frontend/src/styles.css`, and existing auth tests -- add concise EN/FR presentation and cover every matrix row.

**Acceptance Criteria:**
- Given a valid Findur session whose SnapTrade grant is proven unusable, when authorization status is resolved, then the current Findur session is revoked, its cookies expire, Connect explains the issue, and its action starts a new hosted SnapTrade authorization while saved choices remain intact.
- Given a failure returns to Findur, when recovery renders, then it contains no raw detail or technical jargon and offers only a safe next action.
- Given the brokerage connection alone is disabled, when recovery is shown, then the action opens SnapTrade connection management and does not mislabel the Findur grant as expired.
- Given authorization succeeds after recovery, when inclusion is resolved, then an existing committed selection returns to Portfolio without replaying initial account choice.

## Implementation Notes

- Product decision (2026-09-23): keep brokerage connection and account discovery on the guarded 24-hour synchronization schedule. Do not add a live SnapTrade inventory request on every login; the UI surfaces a disconnected connection as soon as a scheduled refresh records it.
- Documentation follow-up: explain the possible connection-status delay in the user FAQ. This is tracked in `deferred-work.md` so the rationale is available when the FAQ is written.
- SnapTrade's official documentation confirms that revoking connected-app access invalidates the existing token, while a later OAuth flow can grant access again; it does not document a special `revoked` authorization response on the next login.
- Protected route transitions remount the authorization gate so stale authenticated state cannot render account selection after revocation. Disabled accounts are grouped by stable connection ID and rendered as semantic lists with compact alert headings.

## Spec Change Log

- 2026-09-23: Implemented categorical browser recovery, server-side session revocation, connection-specific repair messaging, deterministic route gating, local scenarios, and EN/FR tests.

## Review Triage Log

- Fixed authorization-status failures to fail closed and reconcile concurrent server-side session invalidation.
- Required the generated `reauthorizationRequired` status field in the frontend and added coverage for the stricter contract.
- Restored exact integration request-count assertions and updated the healthy mixed-account browser fixture expectation.
- Deferred local scenario-script hardening and unrelated account-selection UX changes; the scenario script remains manual-only and is not used by Compose tests.

## Verification

**Commands:**
- `cd frontend && npm run generate:api` -- passed.
- `cd frontend && npm run typecheck` -- passed.
- `cd frontend && npm run lint` -- passed with 0 errors and 4 existing warnings.
- `cd frontend && npm test -- --run` -- passed; 120 tests passed.
- `cd backend && go test ./internal/auth ./internal/platform/httpapi ./internal/platform/oidcfixture ./internal/portfolio` -- passed.
- `cd backend && golangci-lint run` -- passed with `0 issues.` using the pinned local binary on `PATH`.
- `git diff --check` -- passed.
- Focused revoked-route and connection grouping suite (`npm test -- --run src/App.test.tsx src/showcase.test.ts`) -- 92 tests passed.
