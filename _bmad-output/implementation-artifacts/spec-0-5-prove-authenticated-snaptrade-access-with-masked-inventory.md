---
title: 'Story 0.5: Prove Authenticated SnapTrade Access with Masked Inventory'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
review_loop_iteration: 1
baseline_commit: 'c7b4e69345b485985d59e37cc41479e3b98a6174'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** An authenticated user reaches only a Portfolio placeholder, so Findur has not proved that its OAuth bearer can retrieve the user's connection lifecycle and minimum masked account inventory without crossing the later account-inclusion boundary.

**Approach:** Add one owner-scoped, bounded inventory bootstrap that lists connections before their accounts through a purpose-limited `PortfolioProvider`, publishes only normalized masked inventory, and renders persisted categorical results with exact safe recovery actions. Generate the provider client reproducibly from SnapTrade's exact-revision upstream OpenAPI specification plus a minimal checked-in OAuth-bearer overlay limited to these allowlisted operations.

## Boundaries & Constraints

**Always:** Derive the owner only from `auth.Actor`; preserve the existing token-envelope contract; pin the upstream provider specification by full revision, keep the overlay minimal and reviewable, and fail CI on generated-client drift; run provider calls outside transactions; claim bootstrap/retry work once and atomically publish a generation-checked normalized version. Persist only opaque connection/account IDs, safe brokerage label, category/type, masked display label, eligibility/availability, sync mode/state, safe status/retry metadata, and timestamps required for later selection. Return protected payloads as `private, no-store`; show connection state before accounts; preserve EN/FR, themes, responsive shell, heading focus, and no-default-inclusion copy.

**Never:** Accept a client actor ID; handwrite or fork a general SnapTrade client; persist/log raw provider bodies, tokens, full account numbers, exact financial values, or unused fields. Do not call global account, balance, position, activity, order, trading, manual-refresh, webhook, or reference-data operations. Do not add account selection/inclusion, token refresh, disconnect, background polling, browser persistence, live CI credentials, Commercial auth fields, permissive CORS, or alter OAuth scope/callback/session/logout/health/readiness contracts.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| First bootstrap | Active session/authorization; no inventory head | Claim once; decrypt bearer; list connections then accounts; publish masked version; return ready/empty | Concurrent reads observe pending/persisted state and do not duplicate calls |
| Persisted revisit | Existing terminal inventory state | Serve the normalized snapshot without a provider call | Keep the last trustworthy head and categorical status |
| Incomplete provider data | Missing/unsupported/malformed fields | Mark affected connection/account unavailable or ineligible without inference | Normalization failure exposes no provider detail and offers safe retry |
| Provider lifecycle/failure | Disabled, 401, 429, timeout, transport/5xx | Preserve trustworthy data; report disabled/unauthorized/rate-limited/unavailable state | Offer repair, reauthorize, or bounded retry as appropriate; honor safe retry timing |
| Explicit retry | Authenticated same-origin CSRF-protected action | Claim one new generation and repeat the bounded operation | Reject invalid session/defenses; stale completion cannot replace a newer head |

</frozen-after-approval>

## Code Map

- `backend/internal/portfolio/` -- add provider/repository ports, minimized values, categorical errors, and prepare-call-finalize inventory orchestration; keep provider and transport types out.
- `backend/internal/auth/callback.go` -- extract reusable access-token envelope encryption/decryption without changing existing AAD or ciphertext compatibility.
- `backend/internal/platform/provider/` -- retain the Story 0.1 diagnostic client; add the pinned narrow bearer-only connections/accounts adapter, bounded decoding, minimization, and request-shape tests.
- `backend/internal/platform/provider/` -- reject duplicate provider IDs, derive effective connection freshness from both freshness fields, mask short numbers and account-number text inside labels, and return trustworthy normalized partial connections together with a categorical error. A 429 without usable timing receives a conservative persisted retry floor; parse applicable account reset timing after `Retry-After`.
- `backend/db/migrations/000005_*.sql`, `backend/internal/platform/postgres/inventory.go` -- append portfolio inventory lifecycle/version/head/typed-row storage, owner isolation, single claims, and atomic publication; never rewrite migrations 000001-000004.
- `backend/internal/platform/postgres/oauth_attempts.go` -- successful replacement of a provider authorization invalidates that owner's prior inventory state in the same transaction so reauthorization bootstraps with the new bearer rather than looping on an old terminal state.
- `backend/api/openapi.yaml`, generated Go/TypeScript artifacts, `backend/internal/platform/httpapi/inventory.go` -- define authenticated inventory GET and defended retry POST with categorical, minimized responses and no client owner field.
- `backend/api/openapi.yaml` -- document cookie authentication and the required retry CSRF header exactly as enforced at runtime.
- `backend/internal/platform/config/config.go`, `backend/cmd/findur/main.go` -- validate provider base/policy and compose token, portfolio, repository, provider, and HTTP dependencies without affecting diagnostics.
- `frontend/src/{App.tsx,i18n.tsx,styles.css}`, `frontend/src/pages/PortfolioPage.tsx`, inventory client/tests -- replace only the Portfolio placeholder with the bilingual persisted result/recovery surface; preserve shell and public flows. Share one in-flight initial GET across React development remounts, offer an explicit status check for a concurrent `pending` observer without polling, render availability separately from eligibility/category/sync state, and keep the authenticated header sticky in the desktop rail layout.
- `test/fixtures/wiremock/`, `test/integration/` -- execute synthetic success, empty, disabled, 401, 429, malformed, and transient fixtures through WireMock and prove ordering, minimization, persistence/recovery, and no render-driven repeat call. Unit coverage must also hold inventory composition open for existing sessions when authorization initiation is gated closed and exercise every frontend recovery branch.

## Tasks & Acceptance

**Execution:**
- [x] `backend/internal/portfolio/`, token/provider adapters -- implement bounded owner-scoped orchestration, compatible token decryption, mapping, failure classification, and matrix unit tests.
- [x] `backend/db/migrations/000005_*.sql`, `backend/internal/platform/postgres/inventory.go` -- persist isolated immutable normalized inventory and guarded claims/heads; test concurrency, rollback, stale generations, isolation, and zero inclusion.
- [x] `backend/api/openapi.yaml`, generated clients, `backend/internal/platform/httpapi/`, composition/config -- expose protected GET/retry contracts with session/CSRF/origin defenses, safe caching/logging, and generated drift tests.
- [x] `frontend/src/`, WireMock/integration files -- render accessible EN/FR success and recovery states, exact boundary copy/actions, and prove every fixture plus persisted revisit behavior.

**Acceptance Criteria:**
- Given an active authorization without inventory, when Portfolio first loads, then exactly one bounded bearer-only connection/account operation runs outside transactions and publishes only the allowed masked fields.
- Given any later render, revisit, locale/theme change, or shell navigation, when no explicit retry is required, then persisted state is served without another SnapTrade request and meaningful focus is preserved.
- Given success, empty, disabled, unauthorized, rate-limited, malformed, or transient fixtures, when automated checks run, then owner isolation, request shape, normalization, no-default-inclusion, categorical recovery, and absence of sensitive/financial data are proven without live credentials.

## Implementation Notes

- The narrow provider contract is a deterministic projection of the two approved operations and their transitive response schemas from pinned upstream SnapTrade OpenAPI revision `9a7c008c1b0c7ea9f7944938556d04f66b381082`. The checked-in overlay contains only operation selection, Commercial-query removal, and OAuth bearer security; its generated package lives at `backend/internal/generated/providerapi/`, separate from the Findur API package.
- `SNAPTRADE_API_BASE_URL` is typed with the other SnapTrade OAuth/provider settings and remains independent from the integration-only diagnostic fixture URL in config, Compose, and documentation. Production uses the validated `https://api.snaptrade.com` default; Story 0.5 intentionally leaves the Render Blueprint unchanged so reconciliation cannot reapply its provisioning image.
- A provider `429` persists safe retry timing from `Retry-After` (delta or HTTP date) or `X-RateLimit-Reset`; explicit retries before that instant serve the persisted category without claiming a generation or making a provider call. No circuit breaker was added because the approved scope requires bounded explicit retry rather than background or automatic retry policy.
- Pending claims carry a durable 30-second lease. Concurrent requests inside the lease observe pending without provider work; after expiry, either GET or explicit retry can generation-safely reclaim once, and a completion from the crashed generation cannot publish.
- Generic provider `403`/`404` responses remain ambiguous and map to unavailable. Only a successful connection object with `disabled=true` establishes disabled state. An inventory-endpoint `401` is separately typed in the browser and replaces the route with authentication recovery rather than presenting a provider retry.
- A terminal version may contain only the normalized rows actually established before a later provider failure. Disabled connections and safe partial connection lifecycle data are publishable; a failure with no new trustworthy rows retains the prior head. Reauthorization deletes the old inventory state transactionally so the new bearer starts a fresh generation.
- Account labels never reveal a complete short account number or a known full number embedded in the provider display name. Duplicate IDs are malformed. Effective connection freshness is delayed when either upstream freshness component is delayed.
- A missing or invalid 429 delay persists a conservative 60-second retry floor. `Retry-After` remains authoritative; applicable account-level reset timing is a fallback. The browser provides no automatic retry or polling.
- The initial browser request is shared while in flight so React Strict Mode cannot strand the active mount on a competing pending response. A separate concurrent observer can explicitly check persisted status without causing provider work.

## Spec Change Log

- Review loop 1: the independent review and live SnapTrade run found stale terminal inventory surviving reauthorization, disabled/partial normalized rows being discarded, incomplete masking/freshness/rate-limit normalization, an inaccurate HTTP security contract, incomplete categorical UI/integration coverage, and a reproducible React Strict Mode pending-response race. The code map and design notes now require transactional invalidation on reauthorization, safe terminal/partial publication, exact minimization and throttling guards, faithful contract metadata, one shared in-flight GET plus explicit pending status check, full categorical presentation, and execution of every fixture. Known-bad states to avoid are reconnect loops, full short-number disclosure, repeated headerless 429 calls, permanently pending first renders, stale active rows under disabled state, and green tests that never exercise required recovery paths. KEEP: the checksum-pinned upstream projection and minimal OAuth overlay; isolated generated provider package; owner derived only from `auth.Actor`; AES-GCM envelope compatibility; prepare/call/finalize transactions with generation and lease guards; bearer-only allowlisted requests; `private, no-store`; accounts nested under connections; no default inclusion or financial values; EN/FR and theme parity; validated SnapTrade provider-base configuration kept separate from fixtures; and the unchanged Render Blueprint.

## Review Triage Log

| Finding | Verdict | Evidence | Route |
|---|---|---|---|
| BH-01 authorization renewal retains terminal inventory | high | `storeAuthorization` replaces the bearer but does not invalidate `portfolio_inventory_state`; ordinary GET cannot reclaim a terminal state, so reauthorization returns to the same reconnect-only state. | bad_spec |
| BH-02 disabled connections are discarded | medium | `Finalize` publishes rows only for ready/empty, while the service supplies disabled connections under `StateDisabled`; the institution and repair context are lost. | bad_spec |
| BH-03 short account numbers are fully exposed | high | `maskedSuffix` returns every alphanumeric rune when length is four or less, making the displayed suffix the complete number. | patch |
| BH-04 provider account name can contain the full number | high | The schema permits user/brokerage-authored names and `SafeLabel` only trims/bounds them; a name containing the known raw number is persisted verbatim. | patch |
| BH-05 effective freshness ignores institution delay | medium | The pinned schema says either delayed component makes the connection effectively delayed, but normalization reads only `snaptrade`. | patch |
| BH-06 headerless 429 is immediately retryable | medium | Invalid/absent timing produces nil `RetryAt`, which `Prepare` treats as eligible immediately; applicable account reset timing is not parsed. | bad_spec |
| BH-07 later account failure discards normalized connections | medium | `Load` returns nil on any account-call error and failure finalization supplies no connections, losing already established lifecycle data on an initial load. | bad_spec |
| BH-08 malformed account publishes ready without recovery | medium | A malformed account marks its connection unavailable but returns success, so the top-level state is ready and exposes no retry path. | bad_spec |
| BH-09 UI omits eligibility and sync categories | medium | The page reads connection status and account category/type but omits connection availability/eligibility/sync mode and account eligibility/sync state; the live run showed the resulting ambiguity. | bad_spec |
| BH-10 disabled repair and reauthorization may require distinct flows | maybe-false | The current action is shared, but whether SnapTrade authorization can also repair a disabled connection is not established by the repository or pinned contract. Provider UX evidence would settle it. | defer |
| BH-11 retry button remains enabled before `retryAt` | low | The click is misleading, but the repository prevents another provider claim and safely returns the persisted state; a live timer would add complexity for an uncommon harmless action. | reject |
| BH-12 pending observers can remain stuck | high | A concurrent GET returns pending once and the page has no later observation path; the live Strict Mode run reproduced this while the claimant successfully published ready. | patch |
| BH-13 retained rows falsely appear current | false | The page does not display `updatedAt`, and the top-level unavailable/malformed category remains visible above retained rows; the cited timestamp cannot currently mislabel the rendered rows as fresh. | reject |
| BH-14 OpenAPI omits cookie auth and optionalizes CSRF | medium | No cookie security scheme/operation security is declared and `X-CSRF-Token` lacks `required: true`, unlike runtime enforcement. | patch |
| BH-15 non-success fixtures never execute through WireMock | medium | Only success mappings are registered and journal assertions cover only the two success requests; required categorical integration paths are absent. | bad_spec |
| BH-16 result boundary copy omits later-disabled surfaces | low | The consent page names them, but the result boundary repeats only balances, positions, and activity; expanding the existing sentence is a direct bilingual copy correction. | patch |
| EH-01 disabled retry can retain stale active rows | medium | Because disabled is non-publishing, a prior head remains attached to the new disabled state and can render stale active connections. | bad_spec |
| EH-02 duplicate provider IDs are not rejected early | medium | Arrays have no uniqueness guard; duplicate IDs reach unique database constraints and leave the generation pending until lease recovery. | patch |
| EH-03 short-number masking claim fails | high | Carried BH-03 evidence: the full value is used whenever the cleaned account number has at most four characters. | patch |
| EH-04 safe retry claim fails without timing headers | medium | Carried BH-06 evidence: nil timing permits an immediate explicit retry instead of a conservative retry floor. | bad_spec |
| EH-05 overlay validation does not lock exact operations/removals | medium | Validation enforces only two GETs and an allowlist of removable parameter names; it does not require the approved paths or both removals. | patch |
| EH-06 pending observer lacks a completion path | high | Carried BH-12 evidence, confirmed by the live first render and subsequent successful manual refresh. | patch |
| VG-01 gate-closed inventory composition lacks regression proof | medium | Pre-verified: no test calls `buildInventory` with initiation disabled, so a regression returning nil would pass existing config/default-Compose tests and lock out valid sessions. | patch |
| VG-02 frontend categorical recovery branches lack tests | medium | Pre-verified: disabled, categorical unauthorized, rate-limited, and malformed 200 responses are absent from frontend tests. | patch |
| VG-03 disabled normalized rows are lost | medium | `StateDisabled` carries connections from the service but PostgreSQL neither versions nor inserts them; current layer tests stop short of this combination. | bad_spec |
| LIVE-01 development double mount races to pending | high | Logs showed the quick pending response winning the active Strict Mode mount while the original request published ready; refresh immediately rendered the stored accounts. | patch |
| LIVE-02 authenticated header scrolls away with real inventory length | low | The header has no positioning rule while the rail is fixed; the live eight-account page exposed the mismatch. | patch |
| LIVE-03 available-but-unknown accounts are visually ambiguous | medium | Live categorical storage contained four available and four unavailable accounts, all ineligible with unknown category; the UI only explains the unavailable four. | bad_spec |
| FBH-01 reauthorization can reuse a generation | high | Deleting `portfolio_inventory_state` lets the next claim restart at generation 1, so an old generation-1 completion can pass the new claim's guard. | patch |
| FBH-02 partial rate limits lose retry timing | medium | Partial-row publication unconditionally writes `retry_at=NULL`, even when the categorical provider error supplied a safe retry deadline. | patch |
| FBH-03 an account-fetch failure drops later connection metadata | medium | The adapter returns on the first failed account request although the already-fetched connection list may contain later trustworthy lifecycle rows. | patch |
| FBH-04 malformed or duplicate connections drop later valid rows | medium | Normalization returns immediately for an unusable or duplicate connection rather than preserving independent later rows while reporting malformed data. | patch |
| FBH-05 separator variants can bypass account-number redaction | high | Exact raw and contiguous-alphanumeric replacement does not match the same account number rendered with different separators. | patch |
| FBH-06 unknown account status can be treated as available | medium | JSON decoding accepts an unknown string enum and normalization checks only the known unavailable/archived values without calling the generated enum validator. | patch |
| FBH-07 account limits are not operation-wide | false | The operation is still deterministically bounded by 100 connections, 500 accounts per connection, sequential requests, and per-response byte limits; the approved intent sets no smaller aggregate ceiling. | reject |
| FBH-08 the shared initial request can cross a logout/login boundary | high | The module-global promise has no session key and is not reset by successful logout, so a later mounted session can join an old request. | patch |
| FBH-09 mixed active and disabled connections lack an exact repair action | maybe-false | The missing per-row action is visible, but the repository and pinned contract do not establish whether hosted authorization repairs one disabled connection or a distinct provider flow is required. | defer |
| FBH-10 retry 403 is presented as a retryable outage | medium | The client collapses the defended endpoint's explicit 403 into the same generic error as a transient backend failure, causing the same ineffective action to be offered again. | patch |
| FBH-11 boundary copy omits derivation, previews, and Discovery | false | Those surfaces are not required by Story 0.5's approved result-boundary copy; the rendered copy names every financial/provider operation explicitly excluded by this story. | reject |
| FBH-12 retry timestamps ignore the selected locale | low | `toLocaleString()` receives no locale even though the active EN/FR locale is available; passing it is a direct presentation correction. | patch |
| FBH-13 inventory OpenAPI omits runtime 503 | medium | Both typed inventory operations can reach the strict handler's 503 response path, but neither operation documents that categorical response. | patch |
| ECH-01 reauthorization ABA defeats the generation guard | high | Independently confirms FBH-01: deleting state makes an old and new claim share generation 1. | patch |
| ECH-02 finalization after reauthorization can return a zero snapshot | medium | When deletion leaves no state, an old completion returns the zero-valued preparation snapshot; preserving a monotonic invalidated state removes this same ABA-rooted path. | patch |
| ECH-03 partial publication clears safe retry timing | medium | Independently confirms FBH-02 at the PostgreSQL publication boundary. | patch |
| ECH-04 the shared request is not owner-scoped | high | Independently confirms FBH-08 across a completed logout and later authenticated mount. | patch |
| ECH-05 formatted account numbers bypass masking | high | Independently confirms FBH-05 for semantically identical account numbers with different separators. | patch |
| ECH-06 JSON null lists become trustworthy empty inventory | medium | Decoding JSON `null` into a slice succeeds with a nil slice, and the adapter currently treats it as an empty provider list. | patch |
| ECH-07 huge numeric Retry-After can overflow duration | medium | A numeric value can fit the platform integer while overflowing `time.Duration` multiplication, producing unsafe retry timing. | patch |
| ECH-08 generation can overflow at MaxInt64 | low | Reaching the condition requires more than nine quintillion claims for one owner; adding production complexity for this unreachable lifecycle is not justified. | reject |
| VG-01 partial categorical rows lack PostgreSQL publication coverage | medium | Provider and service tests stop before persistence, while repository tests do not finalize malformed or unavailable states with new trustworthy rows; a publication regression would pass. | patch |

## Design Notes

The first authenticated inventory GET is the bootstrap trigger: a short prepare claim precedes the provider call, and a short guarded finalize publishes it. Failures persist a safe terminal category without moving a prior trustworthy head. Only the explicit defended retry or recovery of an expired durable claim may create another generation; ordinary component renders never do.

## Verification

**Commands:**
- `cd backend && go generate ./... && git diff --exit-code -- api internal/generated && go test -race ./... && go vet ./...` -- generated contracts are stable and backend/unit/PostgreSQL behavior passes.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- UI states, accessibility behavior, generated types, and production bundle pass.
- `docker compose config && docker compose build && docker compose up --wait && docker compose run --rm integration` -- synthetic OAuth-to-inventory flow and all WireMock request/count/privacy assertions pass.
- `git diff --check` -- changed files contain no whitespace errors.
