---
title: 'Story 0.4: Protect the Authenticated Experience and End Sessions'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: 'e71f598ce1f75187a9b0282f32fd1b0d0269b85d'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 0.3 issues opaque sessions, but Findur does not yet authenticate protected requests, enforce the checked-in session/CSRF policy, render a private shell, or let a user end the current session safely.

**Approach:** Add a focused server-side session lifecycle and immutable Actor boundary, protect unsafe requests, expose one clear current-session logout, and gate a bilingual responsive authenticated shell without implementing later portfolio or discovery features.

## Boundaries & Constraints

**Always:** Derive Actor only from a hashed active session; enforce 12-hour rolling idle and 7-day non-extending absolute expiry. Unsafe authenticated requests require the session-bound `X-CSRF-Token`, exact same-origin Origin or permitted Referer, and acceptable Fetch Metadata. Return categorical `401`/`403` and `private, no-store`; logout revokes only the authenticated browser session, expires both cookies, and clears all protected browser data before navigating to `/`. Retain only locale and theme preferences. Gate private content before mounting; expose exactly Discovery, Portfolio, and Profile, with bottom navigation below 768px and a rail at or above it. Route this incomplete user to Portfolio. Preserve EN/FR, System/Light/Dark, focus, safe areas, reduced motion, zoom, and installable privacy. Use small single-purpose functions, typed constants, and shared request/session helpers.

**Never:** Accept a client actor ID, store plaintext session/CSRF values, extend absolute lifetime, trust forwarded host headers, or log sensitive request/session data. Do not add logout-all, cross-session management, or revoke another browser's session. Do not rewrite migration 000003, couple sessions to OAuth initiation/discovery, or add JWTs, Redis, permissive CORS, a service worker, protected browser persistence, provider calls, disconnect/deletion, inventory/inclusion, or later product features.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Protected entry | Valid cookie; protected route | Touch idle capped by absolute; inject Actor; render shell/Portfolio gate | Fail closed without private render |
| Invalid session | Missing, revoked, expired, or inactive owner | No private mount/persistence; show `/connect` recovery | Categorical `401`; allowlisted route only |
| Unsafe request | Valid session, CSRF, origin, Fetch Metadata | Mutate for derived Actor | Any failed defense returns `403` before mutation |
| Current logout | Valid protected request | Revoke current session, expire cookies, clear protected browser data, then navigate to `/` | On failure stay authenticated with localized retry |
| Expiry boundary | Session used near seven days | Renew idle to `min(now + 12h, absolute)` | Reject at either expiry; never renew |

</frozen-after-approval>

## Code Map

- `backend/internal/auth/{callback.go,session.go}` -- inject lifetimes; isolate `Actor`, authenticate/touch, CSRF, revocation, cleanup, and status while reusing hashed-secret/replay logic.
- `backend/db/migrations/000004_session_lifecycle.*.sql`, `backend/internal/platform/postgres/{oauth_attempts.go,sessions.go}` -- add lifecycle state, atomic capped touch, current-session revocation, and cleanup; preserve 000003/finalization.
- `backend/api/openapi.yaml`, `backend/internal/platform/httpapi/{authorization.go,session.go,httpapi.go}` -- generate logout contracts and enforce Actor, origin/Referer, Fetch Metadata, CSRF, cookies, and no-store policy.
- `backend/internal/platform/config/config.go`, `backend/cmd/findur/main.go`, `compose.yaml`, `render.yaml` -- validate public origin and compose sessions independently of OAuth initiation.
- `frontend/src/{App.tsx,auth-status.ts,session.ts,i18n.tsx,styles.css,components/AuthenticatedLayout.tsx}` -- add no-flash auth routing, centralized logout, neutral metadata, and the three-link shell; reuse preferences/focus.
- Existing backend/frontend tests and `test/integration/{integration.mjs,browser-session.mjs}` -- cover the matrix, isolation, responsive navigation, and browser privacy.

## Tasks & Acceptance

**Execution:**
- [x] Domain/storage paths -- add policy, Actor authentication, capped renewal, current-session revocation, cleanup, and migration 000004 without changing OAuth grants/000003.
- [x] Contract/HTTP/config paths -- generate the logout operation; enforce all request defenses and keep sessions available when login closes.
- [x] Frontend paths -- add the no-flash Portfolio gate, responsive shell, localized logout/retry, safe metadata, and protected browser-data clearing.
- [x] Code Map tests -- cover every matrix row, concurrency/isolation/spoofing, 767/768 layout, focus recovery, and storage absence.

**Acceptance Criteria:**
- Given any owner-scoped operation, when it is authorized, then only the immutable session Actor can select resources and guessed IDs or supplied actor fields cannot cross user boundaries.
- Given protected routes and shared actions, when used across authentication states, locales, themes, viewport boundaries, refresh/back, and standalone launch, then private content never flashes, the current task/focus remains meaningful, and only the three specified destinations appear.
- Given generated artifacts and the synthetic stack, when verification runs, then isolated sessions prove renewal, current-session logout without affecting another browser, no protected browser data after logout, expiry, CSRF rejection, and no sensitive persistence.

## Implementation Notes

- Added a dedicated immutable Actor/session service and PostgreSQL adapter. Active-session lookup and rolling renewal are one atomic statement capped by the seven-day absolute deadline; revocation addresses only the presented hash.
- Added generated logout contracts and independent session configuration. Origin/Referer and Fetch Metadata fail before session touch; invalid sessions return `401`, while CSRF and request-defense failures return `403` with `private, no-store`.
- Kept failed CSRF checks read-only by separating active-session lookup from the atomic touch, and verified that a concurrent touch cannot revive a revoked session.
- Added a no-flash bilingual authenticated shell and centralized logout. Successful logout clears all browser storage except locale/theme before replacing history with `/`; a failed logout leaves a focused, retryable private shell.

## Spec Change Log

- 2026-09-20: Implemented Story 0.4 without changing the approved intent or migration 000003.
- 2026-09-20: Final review hardened categorical logout errors and confirmed Cache Storage and IndexedDB cleanup; no work was deferred.

## Review Triage Log

| Finding | Verdict | Evidence |
|---|---|---|
| VG1 gate-closed authenticated status lacks a direct test | low | The runtime session branch is independent of the initiation gate, but no test pins that combination; add the focused HTTP test. |
| VG2 populated 000003-to-000004 migration lacks an upgrade test | low | The SQL backfills `last_used_at` from `created_at` before adding `NOT NULL`, so the risk is limited; a version-stepped harness is disproportionate for this submission and is rejected. |
| VG3 seven-day lifetime is not pinned independently | low | Existing tests reuse the production constant; add literal seven-day persistence and cookie assertions. |
| VG4 missing Origin and missing Fetch Metadata lack isolated tests | low | Explicitly invalid values are covered, but omissions are separate reachable defenses; add table cases. |
| VG5 Cache Storage and IndexedDB cleanup are untested | medium | Logout promises all protected browser data is gone, while current tests cover only local/session storage; extend the browser test. |
| EC1 OpenAPI leaves CSRF optional | false | This is intentional so a missing defense reaches the handler and returns the specified categorical `403`; a request without CSRF is not valid merely because a generated type can represent it. |
| EC2 logout error cache contract disagrees with runtime | low | Shared `SafeError` specifies `no-store` while logout emits `private, no-store`; define logout-specific error responses. |
| EC3 invalid/expired logout leaves stale cookies | medium | `ErrUnauthenticated` returns a generated 401 without expired cookies, causing retries with known-invalid browser credentials; expire both cookies and let the client complete local cleanup. |
| EC4 origin casing/default-port comparison rejects legitimate requests | low | Browser-generated Origin values and the configured deployed origin are canonical in normal operation; normalization adds branches for an unlikely configuration-only edge and is rejected. |
| EC5 malformed percent escapes can prevent logout request | low | `decodeURIComponent` can throw before fetch; return an empty token so the server handles it categorically. |
| EC6 empty-string localStorage key survives cleanup | medium | The `key &&` guard demonstrably skips a valid storage key; test `key !== null` instead. |
| EC7 failed or blocked browser-store deletion is treated as success | medium | Rejected/false cache deletion and IndexedDB error/blocked paths currently resolve; cleanup must reject so navigation does not claim data was cleared. |
| EC8 an indefinitely pending browser API can trap logout | low | A timeout would allow navigation before confirmed deletion or introduce cancellation complexity; this speculative platform failure is rejected. |
| EC9 obsolete authorization-result component and copy remain | low | The route now canonicalizes directly to Portfolio, making the component/copy unreachable; delete them as a direct cleanup. |
| EC10 verification claims spoofing and storage absence without coverage | medium | A forged-success URL test already covers spoofing, but Cache Storage/IndexedDB absence is unverified; extend the browser test. |
| EC11 logout can claim no protected data despite cleanup failures | medium | The swallowed deletion failures are real and share EC7's root cause; require confirmed cleanup before navigation. |
| BH1 OpenAPI leaves CSRF optional | false | Same as EC1: optional transport binding preserves the required categorical `403` path and does not make a CSRF-less request valid. |
| BH2 logout SafeError cache contract mismatch | low | Same verified mismatch as EC2; define logout-specific response contracts. |
| BH3 generated 204 Set-Cookie serializer is unsafe | false | Production uses the custom `logoutResponse` visitor, which calls `http.SetCookie` once per cookie; the generated visitor is unreachable. |
| BH4 duplicate CSRF headers bypass categorical/no-store errors | low | Generated parameter binding runs before registered middlewares and uses its default error handler; configure the standard wrapper error handler. |
| BH5 session-store errors become forbidden | low | `AuthorizeUnsafe` non-domain failures are currently collapsed into `403`; propagate them to the existing response-error path. |
| BH6 status-store outages look unauthenticated | low | The UI remains fail-closed and the transient misclassification exposes no private content; a new availability UX exceeds the submission-critical scope and is rejected. |
| BH7 reusable Actor boundary is missing | false | `SessionService.Authenticate` and `AuthorizeUnsafe` are the single Actor-producing boundary; logout correctly discards the Actor because it performs no owner-scoped resource operation. |
| BH8 session cleanup has no production scheduler | low | Rows contain hashed/encrypted material and cleanup scheduling is operational work with no immediate logout/privacy failure; reject for this submission. |
| BH9 private shell does not continuously recheck auth | low | Every protected mount/refresh is gated and no private feature data exists in this story; continuous polling/visibility policy is not justified here and is rejected. |
| BH10 logout 401 incorrectly leaves a retrying private shell | medium | An expired/revoked session is already ended, so the client should clear protected storage and leave the shell rather than claim it remains active. |
| BH11 logout 401 retains stale cookies | medium | Same verified root cause as EC3; expire cookies on unauthenticated logout responses. |
| BH12 frontend accepts any 2xx as logout success | low | The endpoint contract is specifically `204`; require that status before clearing and navigating. |
| BH13 browser-store failures are ignored | medium | Same verified root cause as EC7; inspect deletion outcomes and reject error/blocked deletion. |
| BH14 browser-store cleanup branches lack tests | medium | Same verified gap as VG5; seed and verify Cache Storage and IndexedDB in the browser integration. |
| BH15 public recovery heading is not focused on first mount | low | Logout removes the formerly focused control, but adding universal initial-page autofocus would alter public navigation behavior; no sensitive-data consequence and the nontrivial UX change is rejected. |

## Design Notes

Status stays categorical `200`; protected failures use `401`/`403`. Actor is server-only. Logout revokes only its presented session hash; another browser remains authenticated. The single logout action sits outside navigation; failure keeps a retryable shell. Success clears protected browser data before navigating to `/`. Authenticated result/incomplete entries canonicalize to Portfolio.

## Verification

**Commands:**
- `cd backend && go generate ./... && git diff --exit-code -- api internal/generated && go test -race ./... && go vet ./...` -- generated code and backend checks pass.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- frontend checks pass.
- `docker compose config && docker compose build && docker compose up --wait && docker compose run --rm integration` -- synthetic stack passes.
- `git diff --check` -- changed files contain no whitespace errors.
