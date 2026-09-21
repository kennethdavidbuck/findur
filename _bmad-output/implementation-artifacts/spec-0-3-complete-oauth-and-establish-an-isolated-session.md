---
title: 'Story 0.3: Complete OAuth and Establish an Isolated Session'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '45c0220a55c2b2976d0a9a0ace64236c558d7a2f'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur can begin hosted SnapTrade authorization but cannot safely consume the callback, bind an isolated identity, or establish a private session. Production initiation therefore remains closed and eligible test users cannot complete login.

**Approach:** Complete the authorization-code flow with single-use callback handling, strict OIDC validation, atomic identity/authorization/session persistence, replay-safe recovery, and a bilingual browser result. Open the independent feature gate only when the complete flow is configured and verified.

## Boundaries & Constraints

**Always:** Preserve the existing configured absolute callback URL and fixed `/api/auth/snaptrade/callback` validation unchanged: remote callbacks require HTTPS and loopback HTTP remains available for local testing. Validate binding, state, expiry, callback parameters, RS256 signature, issuer, audience, expiry, plausible issued-at, and nonce before identity/session creation. Exchange and any one-shot compensation revocation must be bounded and outside database transactions. Resume one active user per verified `(snaptrade, subject)` or create a new OAuth-origin user when no identity remains. Encrypt access and refresh tokens independently with versioned AES-256-GCM envelopes and owner/provider/kind/version AAD; store only keyed hashes of session and CSRF secrets. Finalize identity, authorization, session, and terminal attempt state atomically. Keep responses, redirects, UI, and logs categorical and free of provider details and secrets. Prefer pinned, well-maintained, battle-tested packages for OAuth/OIDC, JWT/JWKS, cryptography, UUIDs, and protocol handling; custom code owns only Findur policy and adapter glue.

**Never:** Call user-info, request profile/email scopes, expose or persist plaintext codes, ID/access/refresh tokens, client/session/CSRF secrets, or reconstruct the public callback origin from forwarded headers. Do not hand-roll standards-sensitive protocol, token-validation, cryptographic, or identifier machinery when an appropriate battle-tested package exists. Do not add protected-route middleware, logout, CSRF enforcement, inventory, inclusion, financial data, live provider calls, or reuse the Story 0.1 diagnostic provider client/proxy as OAuth proof. Do not alter the existing callback URL/path policy.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Success | Valid code/state/cookie and verified token | Claim once; create/resume identity; atomically store authorization/session/result; expire attempt cookie and issue opaque session cookie; redirect to safe onboarding | No sensitive browser payload |
| Provider denial or invalid callback | Provider error, missing fields, tampered state, wrong cookie, expired attempt, invalid destination | Perform no token exchange or identity/session write | Expire correlation cookie and show localized restart guidance |
| Invalid token | Bad signature, issuer, audience, time, or nonce | Create no identity, authorization, or session | Record restart-required and show categorical recovery |
| Finalization failure | Tokens issued but local transaction fails | Discard plaintext and perform at most one bounded best-effort revocation | Persist restart-required without reusing the code |
| Replay | Attempt already terminal | Never exchange again; authenticated browser follows recorded safe destination | Unauthenticated browser receives restart guidance |

</frozen-after-approval>

## Code Map

- `backend/internal/auth/authorization.go` -- extend the existing attempt/hash/AEAD service with claim material, exchange/verification/finalization ports, terminal outcomes, session generation, and replay policy.
- `backend/internal/platform/{oidc,postgres,httpapi,config}/` -- adapt discovered token/JWKS/revocation behavior, transactional persistence, callback cookies/redirects, and explicit secret/key-ring policy; retain current callback validation.
- `backend/db/migrations/000003_*.sql` -- add OAuth-origin users, external identities, encrypted authorizations, hashed sessions, and durable terminal callback data without rewriting migration 000002.
- `backend/api/openapi.yaml`, generated Go/TypeScript, generation/CI -- add the callback and safe result contracts reproducibly.
- `backend/cmd/findur/main.go`, `render.yaml` -- compose callback dependencies, declare secret configuration without values, and remove the Story 0.2 production-only forced closure while retaining the explicit gate.
- `frontend/src/{App.tsx,pages/ConsentPage.tsx,i18n.tsx,styles.css,App.test.tsx}` -- enable explicit authorization when available and add accessible EN/FR success/recovery presentation on clean URLs.
- `test/fixtures/wiremock/`, `compose.yaml`, `test/integration/` -- prove a real browser-visible synthetic provider round trip against PostgreSQL; do not substitute `/api/__fixture/proxy/*`.

## Tasks & Acceptance

**Execution:**
- [x] Contract/domain -- define callback/result I/O and implement claim, strict OIDC verification, identity/session policy, replay, and compensation without leaking secrets.
- [x] Storage/config/composition -- add versioned encrypted authorization and hashed-session persistence with one atomic finalization transaction; wire explicit provider/client/key settings and the completed feature gate.
- [x] HTTP/UI -- add no-store callback handling, secure cookie rotation, safe redirects, and accessible bilingual success/restart flows without protected-content flash.
- [x] Tests -- cover every matrix row plus identity create/resume, ciphertext/AAD, rollback, exact configured callback use, browser-reachable provider origin, and zero live credentials/provider access.

**Acceptance Criteria:**
- Given a valid first callback, when finalization succeeds, then the verified subject owns exactly one isolated active user and one opaque session while no provider or browser secret is stored or exposed in plaintext.
- Given callback validation, token validation, or local finalization fails, when recovery completes, then no partial identity/session state survives, the code is never retried, compensation is bounded, and the user receives a safe localized restart path.
- Given a consumed callback is replayed, when the browser returns, then no second exchange or user is created and routing depends only on recorded safe state and current authentication.
- Given the synthetic end-to-end suite passes, when production is configured explicitly, then authorization initiation may open and completes through the real callback rather than the diagnostic proxy.

## Implementation Notes

- Added a separate callback application service and repository port. The service bounds exchange, verification, and one best-effort revocation outside PostgreSQL transactions; finalization atomically commits the active identity, encrypted authorization, hashed session/CSRF material, and durable terminal attempt result.
- OIDC discovery, OAuth exchange, RS256/JWKS verification, issuer/audience/expiry validation, and token revocation use `go-oidc`, `oauth2`, and `go-jose`; Findur adds nonce and issued-at policy. Token envelopes use independently generated AES-256-GCM nonces and owner/provider/kind/version AAD.
- Enabled authorization now requires an explicit confidential-client secret. Token exchange pins OAuth HTTP Basic authentication through `AuthStyleInHeader`, revocation uses the same Basic credentials, and neither operation places the secret in the URL, form body, browser response, or categorical logs.
- Compose uses the preserved full loopback callback URL. The Selenium container shares the frontend network namespace so the real browser can complete the synthetic WireMock-hosted redirect through `127.0.0.1`, PostgreSQL, and the public API route without live credentials.
- Compose accepts shell overrides for live OAuth configuration and exposes PostgreSQL host-only at `127.0.0.1:${POSTGRES_HOST_PORT:-5432}`. Live testing must use `127.0.0.1` consistently for both entry and callback because the correlation cookie is host-only; mixing it with `localhost` correctly fails correlation.
- The no-store `/api/auth/status` contract is the browser's sole authority for the explicit authorization gate and opaque-session authentication. Consent remains disabled while resolving or when closed, forged result query strings cannot claim success, and both successful completion and authenticated replay use the clean `/onboarding/accounts` onboarding shell; legacy `/connect/result` is a replace-only compatibility alias.
- Compose auth settings use shell interpolation with deterministic synthetic defaults. A local live smoke can supply the gate, issuer, client ID/secret, exact loopback callback, and four independent keys in the invoking shell without editing tracked files or creating a tracked `.env`; an unmodified environment continues to run the WireMock flow.

## Spec Change Log

- 2026-09-20: Implemented the approved callback/session story; no frozen intent, boundary, matrix, code-map, or acceptance text changed.

## Review Triage Log

| Finding | Verdict | Evidence and route |
|---|---|---|
| BH-1 | medium | A token response can contain an access token but omit `id_token`; the adapter currently drops the partial token set, preventing compensation. Patch by preserving the issued token for one bounded revoke. |
| BH-2 | medium | Compensation and terminalization inherit request cancellation, so a disconnected browser can abort both after token issuance. Patch cleanup onto a detached, bounded context. |
| BH-3 | medium | `FailCallback` errors are discarded and can leave the attempt `exchanging`. Patch to surface the storage failure while preserving categorical browser recovery and single-use rejection. |
| BH-4 | high | HTTPS issuer validation does not currently reject discovered HTTP token, JWKS, or revocation endpoints, so secrets or tokens could cross cleartext. Patch and test every security-sensitive discovered endpoint before caching. |
| BH-5 | false | `/portfolio` has no satisfied prerequisites in this story, so deliberately recording and following `/onboarding/accounts` is the specified safe-onboarding fallback; the original requested route must not win yet. |
| BH-6 | false | Story 0.3 only establishes and reports a session; idle renewal belongs to Story 0.4's protected-session middleware, which the frozen scope explicitly excludes. |
| BH-7 | false | The acceptance language describes the session produced by one successful callback, not a global one-session limit; Story 0.4 explicitly supports current-session and all-session logout, which requires multiple sessions. |
| BH-8 | low | Only version-1 writes exist and this story has no token-decryption consumer, so rotation cannot yet make a live reader fail. Reject the non-trivial future key-ring configuration change here; extend it with the first token consumer. |
| BH-9 | medium | Terminal attempts became exempt from cleanup and can accumulate; patch bounded terminal-attempt cleanup. Expired-session destruction is part of Story 0.4's session lifecycle rather than this callback story. |
| BH-10 | low | Failing closed to unauthenticated/unavailable on a transient status-store error is safe and leaks no protected data. Reject a new public error state because its complexity outweighs the brief generic restart guidance. |
| BH-11 | medium | Synthetic OIDC routes are composed whenever fixture configuration exists, without an integration-environment invariant. Patch composition so fixture routes cannot be enabled in production. |
| BH-12 | low | Session timestamps can be shortened by the bounded provider round trip. Patch directly by sampling the injected clock again immediately before finalization. |
| BH-13 | false | During the mandated review step, the spec is `in-review` while sprint tracking intentionally remains `in-progress`; final presentation performs the workflow status synchronization. |
| EH-1 | medium | Duplicate of BH-1, independently confirming that a missing ID token after access-token issuance escapes compensation. Same patch. |
| EH-2 | false | Retaining an old refresh token after a fresh grant omits one could revive an invalidated credential; clearing it is safer than silently preserving stale authorization. |
| EH-3 | high | Exchange and verification each receive the full 12-second timeout under a 15-second server write deadline. Patch them to share one bounded callback budget while keeping cleanup independently bounded. |
| EH-4 | medium | A resumed user can become inactive between lookup and atomic finalization, and the transaction does not re-check `active`. Patch a locked active-owner check before authorization/session writes. |
| EH-5 | low | A nil completer leaves `err == nil` and logs callback success despite producing only restart routing. Patch the category to require `result.Success`. |
| EH-6 | medium | A stalled status fetch leaves consent/result permanently resolving. Patch a bounded client timeout with abort cleanup. |
| VG-1 | low | Cookie extraction for authenticated replay is only covered below HTTP. Patch an HTTP test that proves the existing session reaches the completer and is not rotated. |
| VG-2 | low | Access-token AAD is proven, but refresh-token kind/AAD is not. Patch positive refresh decryption plus cross-kind rejection assertions. |
| VG-3 | low | Revocation tests prove Basic authentication but not the token and hint form values. Patch exact form assertions. |
| VG-4 | medium | PostgreSQL `FailCallback` terminal persistence has no real-adapter test. Patch claim/fail/reclaim assertions and prove no partial user state. |
| VG-5 | medium | PostgreSQL session checks only prove a fresh active session. Patch separate idle-expired, absolute-expired, and inactive-owner assertions. |

## Design Notes

The callback handler consumes the exact absolute redirect URI already configured by Story 0.2, while its server route and correlation-cookie path remain the fixed validated path. The session record created here is intentionally dormant infrastructure until Story 0.4 adds authenticated route enforcement and logout.

## Verification

**Commands:**
- `cd backend && go generate ./... && git diff --exit-code -- api internal/generated && go test -race ./... && go vet ./...` -- generated contracts are stable and backend behavior passes.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- UI behavior and production bundle pass.
- `docker compose config && docker compose build && docker compose up --wait && docker compose run --rm integration` -- the synthetic hosted OAuth round trip passes through the public-origin callback with PostgreSQL and WireMock.
- `git diff --check` -- changed files contain no whitespace errors.

**Results:**
- `GOCACHE=/tmp/findur-go-cache go test -race ./...` passed, including the PostgreSQL Testcontainer suite.
- `GOCACHE=/tmp/findur-go-cache go generate ./... && GOCACHE=/tmp/findur-go-cache go vet ./...` passed.
- `npm ci && npm test -- --run && npm run typecheck && npm run lint && npm run build` passed with 20 frontend tests; lint retained two pre-existing Fast Refresh warnings and the local build retained the expected unset build-SHA warning.
- `docker compose config`, `docker compose --profile test config`, `docker compose build`, and `docker compose --profile test run --rm integration` passed. The integration run completed the browser-visible loopback OAuth callback and issued the isolated session.
- Default and shell-overridden `docker compose config --format json` assertions passed for the gate, issuer, client credentials, exact loopback callback, and all local key values without using live secrets.
- `git diff --check` passed.
- A real SnapTrade Test OAuth user completed login locally through `http://127.0.0.1:8080`: the callback logged categorical `succeeded`, `/api/auth/status` resolved authenticated, and PostgreSQL contained exactly one user, external identity, encrypted provider authorization, session, and succeeded attempt. No token, subject, code, state, or credential value was inspected or recorded.

**Matrix coverage:**
- Success and identity create/resume: `TestCallbackSuccessCreatesOpaqueEncryptedSessionMaterial`, `TestCallbackCreatesNewIdentityWhenNoneRemains`, `TestCallbackResumesExistingActiveIdentity`, and `TestOAuthAttemptRepositoryAgainstPostgres`.
- Provider denial and invalid callback correlation: `TestCallbackInputFailuresNeverExchange`, plus PostgreSQL mismatched-binding and expired-attempt assertions in `TestOAuthAttemptRepositoryAgainstPostgres`.
- Invalid signature, issuer, audience, expiry, issued-at, and nonce: `TestCallbackClientRejectsInvalidOIDCClaims` and `TestIssuedTokenValidationFailuresCreateNoStateAndCompensateOnce`.
- Finalization rollback and bounded compensation: `TestCallbackFinalizationRollbackCompensatesOnce` and the transactional rollback assertion in `TestOAuthAttemptRepositoryAgainstPostgres`.
- Authenticated and unauthenticated replay: `TestCallbackReplayNeverExchangesAgain` and `TestCallbackReplayWithoutActiveSessionRequiresRestart`.
- Browser truth, closed gate, and forged success: `TestAuthorizationStatusIsNoStoreAndServerAuthoritative`, `TestAuthorizationStatusUsesOnlyHashedOpaqueSession`, and the frontend tests `rejects a forged success URL and trusts only server session status` / `keeps consent closed with localized guidance when the server gate is closed`.
