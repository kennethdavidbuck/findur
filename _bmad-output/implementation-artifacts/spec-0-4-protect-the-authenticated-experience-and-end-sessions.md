---
title: 'Story 0.4: Protect the Authenticated Experience and End Sessions'
type: 'feature'
created: '2026-09-20'
status: 'draft'
route: 'dispatch'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 0.3 issues opaque sessions, but Findur does not yet authenticate protected requests, enforce the checked-in session/CSRF policy, render a private shell, or let a user end current or all sessions safely.

**Approach:** Add a focused server-side session lifecycle and immutable Actor boundary, protect unsafe requests, expose explicit current/all logout operations, and gate a bilingual responsive authenticated shell without implementing later portfolio or discovery features.

## Boundaries & Constraints

**Always:** Derive Actor only from a hashed active session; enforce 12-hour rolling idle and 7-day non-extending absolute expiry. Unsafe authenticated requests require the session-bound `X-CSRF-Token`, exact same-origin Origin or permitted Referer, and acceptable Fetch Metadata. Return categorical `401`/`403` and `private, no-store`; revoke only the requested scope and expire both cookies. Gate private content before mounting; expose exactly Discovery, Portfolio, and Profile, with bottom navigation below 768px and a rail at or above it. Route this incomplete user to Portfolio. Preserve EN/FR, System/Light/Dark, focus, safe areas, reduced motion, zoom, and installable privacy. Use small single-purpose functions, typed constants, and shared request/session helpers.

**Never:** Accept a client actor ID, store plaintext session/CSRF values, extend absolute lifetime, trust forwarded host headers, or log sensitive request/session data. Do not rewrite migration 000003, couple sessions to OAuth initiation/discovery, or add JWTs, Redis, permissive CORS, a service worker, protected browser persistence, provider calls, disconnect/deletion, inventory/inclusion, or later product features.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Protected entry | Valid cookie; protected route | Touch idle capped by absolute; inject Actor; render shell/Portfolio gate | Fail closed without private render |
| Invalid session | Missing, revoked, expired, or inactive owner | No private mount/persistence; show `/connect` recovery | Categorical `401`; allowlisted route only |
| Unsafe request | Valid session, CSRF, origin, Fetch Metadata | Mutate for derived Actor | Any failed defense returns `403` before mutation |
| Current logout | Valid protected request | Revoke current session, expire cookies, clear memory before `/` | Stay safely authenticated with localized retry |
| All logout | Actor has multiple sessions | Revoke all; other browsers fail their next check | Idempotent; retain provider authorization |
| Expiry boundary | Session used near seven days | Renew idle to `min(now + 12h, absolute)` | Reject at either expiry; never renew |

</frozen-after-approval>

## Code Map

- `backend/internal/auth/{callback.go,session.go}` -- inject corrected lifetimes; isolate Actor/authentication/CSRF/renewal/logout from callback code.
- `backend/internal/platform/postgres/{oauth_attempts.go,sessions.go}`, `backend/db/migrations/000004_*.sql` -- add last-seen/revoked state, atomic touch, scoped revocation, cleanup, and isolation.
- `backend/internal/platform/httpapi/{authorization.go,session.go,httpapi.go}` -- add protected context/middleware, origin defenses, private caching, logout, cookie expiry, and safe logs.
- `backend/internal/platform/config/config.go`, `backend/cmd/findur/main.go` -- own typed policy and compose sessions independently of the login gate.
- `backend/api/openapi.yaml` and generated Go/TypeScript -- add `/api/auth/logout`, `/api/auth/logout-all`, CSRF headers, and categorical errors; never hand-edit outputs.
- `frontend/src/{App.tsx,auth-status.ts,session.ts,i18n.tsx,styles.css,components/AuthenticatedLayout.tsx}` -- no-flash routing, centralized unsafe fetch/logout, shell/navigation, recovery, metadata, and state clearing.
- `test/integration/` plus existing Go/frontend tests -- cover session isolation, CSRF, logout scopes, responsive navigation, and browser privacy.

## Tasks & Acceptance

**Execution:**
- [ ] Session domain/storage -- add policy-injected creation, focused Actor/touch/revoke/cleanup operations, and migration 000004 without altering OAuth behavior.
- [ ] Contract/HTTP/composition -- generate logout contracts and enforce reusable auth, CSRF/origin/metadata, cache, cookie, and error policy even when new login is closed.
- [ ] Authenticated UI -- add the no-flash gate, three-area React Aria shell, honest Portfolio prerequisite/recovery, centralized logout, and responsive accessible parity.
- [ ] Tests -- cover the matrix, multi-user/session concurrency, actor spoofing, 767/768 layout, state clearing, and browser privacy.

**Acceptance Criteria:**
- Given any owner-scoped operation, when it is authorized, then only the immutable session Actor can select resources and guessed IDs or supplied actor fields cannot cross user boundaries.
- Given protected routes and shared actions, when used across authentication states, locales, themes, viewport boundaries, refresh/back, and standalone launch, then private content never flashes, the current task/focus remains meaningful, and only the three specified destinations appear.
- Given generated artifacts and the synthetic stack, when verification runs, then isolated sessions prove renewal, both logout scopes, expiry, CSRF rejection, and no sensitive persistence.

## Implementation Notes

## Spec Change Log

## Review Triage Log

## Design Notes

`/api/auth/status` remains categorical browser truth; Actor stays server-only. Current logout revokes by session hash and all-session logout by Actor. The shell contains only honest prerequisite/recovery content; later stories own product data. Prefer cohesive files/helpers over larger callback/transport functions; name repeated policy/protocol literals.

## Verification

**Commands:**
- `cd backend && go generate ./... && git diff --exit-code -- api internal/generated && go test -race ./... && go vet ./...` -- generated contracts are stable; domain, PostgreSQL, middleware, and composition tests pass.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- protected routing, accessibility behavior, types, lint, and production bundle pass.
- `docker compose config && docker compose build && docker compose up --wait && docker compose run --rm integration` -- PostgreSQL/WireMock/browser session lifecycle passes without live credentials.
- `git diff --check` -- changed files contain no whitespace errors.
