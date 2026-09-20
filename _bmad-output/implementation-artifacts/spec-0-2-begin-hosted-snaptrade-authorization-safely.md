---
title: 'Story 0.2: Begin Hosted SnapTrade Authorization Safely'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '7baac9281a2887612453563202b9f2e1c126813d'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Eligible test users cannot review Findur's connection boundary or safely initiate hosted SnapTrade authorization. OAuth must not collect brokerage credentials, imply account inclusion, retrieve financial data, or expose an incomplete callback flow.

**Approach:** Add bilingual consent and an explicit POST backed by a short-lived encrypted attempt and validated OIDC discovery. Keep public initiation independently gated off until Story 0.3 delivers the callback.

## Boundaries & Constraints

**Always:** Preserve Home/About, EN/FR, theme, responsive, focus, and accessibility behavior. Explain SnapTrade, minimum masked inventory, no default inclusion, delayed financial-data use, no password access/trading/advice/wealth judgment, and deletion on disconnect. Generate independent high-entropy state, nonce, verifier, and browser binding; use PKCE S256, exact `openid read`, validated discovery, exact callback, ten-minute expiry, keyed hashes, AES-256-GCM, and a Secure/HttpOnly/host-only/SameSite=Lax cookie. Production requires the configured HTTPS callback; non-production may use only an explicitly configured provider-supported HTTP loopback callback so Story 0.3 can complete local login. Allow only exact internal return routes. Keep provider I/O outside transactions and failures categorical and no-store.

**Never:** Initiate from load, rerender, navigation, presentation changes, or GET. Store/log plaintext OAuth material or provider bodies. Add callback exchange, sessions, inventory/inclusion, financial datasets, derivation, previews, Discovery, signup, Guest Demo, or deferred legal routes. Reuse the Story 0.1 diagnostic client, infer enablement from secrets, or call live SnapTrade in tests.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Consent only | Visit `/connect`, back, locale/theme change | Render localized boundary; create no attempt/request | Preserve route/task/focus |
| Gate closed | Explicit begin while callback feature is unavailable | No discovery, cookie, redirect, or database write | Show localized unavailable guidance |
| Begin | Gate enabled; valid config and return route | Persist one attempt, set cookie, redirect to discovered endpoint | Code/S256/`openid read`/exact callback only |
| Invalid return | External, protocol-relative, malformed, encoded, or unlisted | Use safe default | Never reflect rejected target |
| Init failure | Discovery/config/storage/URL failure | Leave no reusable attempt; preserve sessions | Safe retry without private detail |
| Reuse | Expired or atomically claimed attempt | Refuse a second claim | Return restart-required semantics without provider exchange |

</frozen-after-approval>

## Code Map

- `frontend/src/{App.tsx,components/PublicLayout.tsx,pages/ConsentPage.tsx,i18n.tsx,styles.css,App.test.tsx}` -- add `/connect`, gated owner entry, complete EN/FR content, existing focus/theme/reflow patterns, and no-implicit-request tests.
- `backend/api/openapi.yaml`, generated Go/TypeScript, generation scripts, `.github/workflows/ci.yml` -- introduce the minimal authored contract and drift gate.
- `backend/internal/auth/` -- own creation/claim rules and clock, randomness, crypto, repository, and discovery ports.
- `backend/internal/platform/{postgres,oidc,config,httpapi}/`, `backend/db/migrations/000002_oauth_attempts.*.sql` -- hashed/encrypted persistence, cached discovery, typed policy/gate, safe HTTP behavior, and atomic non-reuse.
- `backend/cmd/findur/main.go` -- assemble auth dependencies without changing lifecycle/probes.
- `test/fixtures/wiremock/`, `test/integration/`, `compose.yaml` -- synthetic discovery and PostgreSQL initiation evidence; preserve Story 0.1 contracts.

## Tasks & Acceptance

**Execution:**
- [x] Contract/tooling -- generate the initiation API reproducibly and fail CI on drift.
- [x] Auth/storage/adapters/config -- create, expire, and atomically claim attempts without plaintext or transaction-spanning network work.
- [x] HTTP/UI -- add safe POST, redirects/cookie/errors/allowlist/gate, and accessible bilingual consent without implicit requests.
- [x] Tests -- cover the matrix, redirect/cookie contract, concurrent non-reuse, cleanup, public regressions, and zero live-provider access.

**Acceptance Criteria:**
- Given the public experience in either locale/theme/layout, when it renders, then its proposition and adult/test/demo boundaries remain intact without signup, Guest Demo, or real-viewer disclosure.
- Given consent, when reviewed, then authorization, later inclusion/use/disclosure, limitations, and disconnect consequences are distinct and complete in EN/FR.
- Given initiation tests, when storage, discovery, redirect, cookies, logs, and failures are inspected, then the security/expiry rules hold and no live credential, financial data, broad scope, or reusable failure exists.
- Given Story 0.3 callback completion is absent, when the deployed public build is used, then initiation remains closed even if provider configuration is present.

## Implementation Notes

- Added the authored OpenAPI 3.1 initiation contract with pinned `oapi-codegen` strict Go bindings, pinned request validation, generated browser types, and backend/frontend CI drift gates.
- Added an inward-facing authorization service with injected clock/randomness/discovery/repository ports. `golang.org/x/oauth2` owns authorization URL and PKCE S256 construction; `github.com/coreos/go-oidc/v3/oidc` owns discovery and issuer validation.
- Added keyed state/nonce/browser-binding hashes, AES-256-GCM verifier envelopes with attempt-specific AAD, a ten-minute PostgreSQL attempt record, atomic first-claim SQL, and cleanup of expired/claimed records. Hashing and encryption keys must be distinct.
- Added explicit typed configuration. Initiation defaults closed and is forced closed in production for this story; local synthetic integration must opt in. HTTP callbacks are restricted to loopback outside HTTPS.
- Added the generated strict POST adapter with runtime OpenAPI request validation and bounded bodies, categorical no-store failures, domain-owned safe return-route normalization, and the Secure/HttpOnly/host-only/SameSite=Lax one-time cookie.
- Propagated request cancellation through discovery and PostgreSQL, bounded authorization work below the HTTP write timeout, removed network I/O from discovery-cache synchronization, and added request-correlated safe failure stages.
- Added the `/connect` consent route in complete English and French, preserving public navigation, focus, theme, responsive behavior, and a disabled authorization action until callback completion.
- Added WireMock discovery and deploy-shaped PostgreSQL initiation evidence without live provider access. Added a Testcontainers-backed repository component test for real PostgreSQL claim concurrency, expiry, cleanup, and cancellation. Tightened the Compose PostgreSQL healthcheck to avoid admitting the backend during first-time database initialization.

## Spec Change Log

- 2026-09-20: Implemented Story 0.2 without changing the frozen intent, constraints, matrix, or acceptance criteria.

## Review Triage Log

- Deferred beyond Story 0.2: remove Story 0.1 fixture diagnostics from the production-compiled artifact by using a separate integration build target or binary; runtime gating alone does not meet the project's intended production-artifact policy.
- Story 0.3 end-to-end OAuth coverage must expose the synthetic provider on a browser-reachable origin distinct from Findur and traverse the real Findur initiation and callback APIs. The existing `/api/__fixture/proxy/*` diagnostic must not substitute for the hosted-provider origin.

| Finding | Verdict | Evidence and route |
|---|---|---|
| BH-1 | medium | `oauth2.GenerateVerifier` can panic and bypasses the injected randomness port; patch generation and its failure test. |
| BH-2 | medium | Cleanup errors are discarded while new encrypted attempts continue to be created; patch as a categorical storage failure. |
| BH-3 | false | The ten-minute boundary governs claimability, which SQL enforces; the story does not require physical deletion without subsequent traffic and cleanup remains opportunistic. |
| BH-4 | medium | The configurable callback path can disagree with the fixed callback-cookie path; patch configuration validation to the implemented route. |
| BH-5 | medium | Non-production configuration currently permits cleartext discovery on arbitrary remote hosts; patch HTTP issuers to explicit integration mode or loopback. |
| BH-6 | false | This integration deliberately narrows the issuer to SnapTrade's exact root issuer; a trailing slash or tenant path is a different issuer identifier. |
| BH-7 | false | OIDC metadata may omit supported scopes, response types, or PKCE capabilities; absence is not proof the required request is unsupported, and the provider safely rejects unsupported requests. |
| BH-8 | medium | Discovery has a time bound but no byte bound; patch a strict metadata response limit. |
| BH-9 | low | Unsupported media types are emitted as 400 despite the authored 415 contract; patch HTTP classification and contract test. |
| BH-10 | low | ServeMux's automatic 405 bypasses the endpoint's no-store categorical response; patch an explicit path fallback. |
| BH-11 | false | Production initiation is forcibly closed in Story 0.2; origin and abuse controls belong to the story that opens initiation. |
| BH-12 | low | The owner-access explanation has an ID but the adjacent link no longer references it; patch the accessible description. |
| BH-13 | low | The disabled primary action does not directly reference its unavailable explanation; patch its accessible description and test. |
| BH-14 | false | `in-review` is the required review-step state; sprint status is synchronized only after review completes. |
| EH-1 | low | Duplicate of BH-9: the unsupported-media response violates the declared 415 contract; patch once and retain this separate triage row. |
| EH-2 | low | A fixed 600-second cookie can outlive the database attempt after initialization delay; patch Max-Age from remaining lifetime. |
| EH-3 | maybe-false | The story requires cached metadata but defines no refresh SLA and provider rotation behavior is unverified; defer a cache-freshness decision. |
| EH-4 | low | Duplicate of BH-12: the owner-access link lost its accessible description; patch once and retain this separate triage row. |
| VG-1 | low | Tests prove unsafe-return defaulting but not preservation of `/portfolio`; add direct service evidence. |
| VG-2 | low | The HTTPS-issuer to HTTP-authorization-endpoint rejection is untested; add direct discovery evidence. |

## Design Notes

The gate is product state, not inferred configuration. Local fixtures may enable initiation; production remains closed. `/connect` may render while submit is unavailable.

The official Go SDK v1.1.0 does not meet the OAuth request-shape gate. Its `SetOAuthClientCredentials` helper uses the client-credentials grant rather than authorization-code plus PKCE, while generated account methods still require and emit `userId`/`userSecret` even when the transport carries a bearer token. Use the architecture-approved pinned upstream OpenAPI with a narrow overlay/generated client; use the OAuth app client ID/secret only for authorization URL identity and backend token/refresh/revocation client authentication.

Authorization attempt persistence contains no plaintext OAuth correlation or verifier material. Discovery completes before the bounded storage operations, and URL validation occurs before insertion, so initialization failures cannot leave an incomplete reusable attempt.

## Verification

**Commands:**
- `cd backend && go generate ./... && git diff --exit-code -- api internal/generated && go test -race ./... && go vet ./...` -- contract and backend pass; the ordinary Go suite requires Docker for its isolated PostgreSQL Testcontainer.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- frontend passes.
- `docker compose config && docker compose build && docker compose up --wait && docker compose run --rm integration` -- synthetic integration passes.
- `git diff --check` -- changed files contain no whitespace errors.

**Results:**
- Go generation was checksum-stable; all backend packages passed under the race detector, including the Testcontainers-backed PostgreSQL repository test, and `go vet` passed.
- TypeScript generation was checksum-stable; 18 frontend tests passed, followed by typecheck, lint (two pre-existing Fast Refresh warnings, zero errors), and production build.
- Compose config/build/start and the black-box integration runner passed. The schema reported `2:false`; repository-level PostgreSQL tests proved atomic single-use claiming, wrong-binding and expiry rejection, cleanup, and context cancellation without exposing database internals to the browser/API suite.
- The final sensitive-value scan and `git diff --check` passed. No live provider request, real credential, or financial value was used.
