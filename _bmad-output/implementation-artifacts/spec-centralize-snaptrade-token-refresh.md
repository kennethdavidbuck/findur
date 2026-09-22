---
title: 'Centralize SnapTrade access-token refresh'
type: 'feature'
created: '2026-09-22'
status: 'done'
route: 'dispatch'
baseline_commit: 'fcfc170ca5db486c2d5e12cf97b9dd135115484a'
review_loop_iteration: 0
context:
  - '_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Portfolio inventory and the scheduled account-sync worker decrypt and use the stored SnapTrade access token directly. They never evaluate expiry or use the encrypted refresh token, so an ordinary token expiry causes continuing provider failures and leaves data stale.

**Approach:** Establish one server-side, per-user credential source that every SnapTrade portfolio operation uses. It provides a valid access token, coordinates rotating refreshes, persists replacements safely, and refreshes then retries a safe provider read once after a `401`.

## Boundaries & Constraints

**Always:** Keep access and refresh tokens server-only, encrypted under the existing token cipher, and absent from logs. Treat an access token as refresh-due 15 minutes before its recorded expiry, using this one shared policy for every caller. Claim a per-authorization refresh lease in a short PostgreSQL transaction, call the OAuth token endpoint outside every transaction, then atomically install both rotated envelopes and expiry only if the lease is still current. Other callers reuse the usable token or wait within a bounded request context; they must never independently spend the same rotating refresh token. A token-endpoint result that may have reached SnapTrade but cannot be safely persisted must transition the authorization to reauthorization-required. On a provider `401`, force at most one refresh and retry only the same safe read once; a second unauthorized result disables usable authorization and directs the existing recovery paths to reauthorization. When authorization becomes unusable, the sync worker stops claiming that user's account resources until a fresh OAuth grant restores access. Temporary provider errors continue through bounded per-resource backoff. Log refresh claims, contention, success, `401` recovery, and authorization failures as structured categorical events without secrets or financial data. Normal Findur session validation remains independent from provider-token refresh.

**Never:** Expose token material to HTTP clients, hold a database transaction across OAuth or SnapTrade network calls, repeat an ambiguous refresh exchange, refresh on page rendering/probes/heartbeats alone, or alter the existing authorization-code callback identity verification.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Valid credential | Access token remains outside the refresh window | Any inventory or sync caller obtains it from the shared source and sends its safe provider request | No token-endpoint request |
| Due credential | Token expires within 15 minutes (or is already expired) and has a decryptable refresh token | Exactly one lease holder refreshes and atomically installs the replacement pair; concurrent callers obtain the installed access token | Invalid grant, missing refresh token, ambiguous delivery, or failed guarded install marks reauthorization required |
| Provider rejection | Safe SnapTrade read returns `401` | Force one coordinated refresh and retry the exact read once | A second `401` marks authorization unusable; no additional retry |
| Coordination race | Another worker/process holds a nonexpired refresh lease | Caller waits and re-reads authorization within its timeout | Lease expiry safely permits recovery; callers do not send another refresh concurrently |
| Authorization failure | A refresh failure requires reauthorization while several account resources are due | The worker stops claiming further resources for that user; a fresh OAuth grant restores eligibility | Existing pending work remains available to resume without needless provider calls before reauthorization |

</frozen-after-approval>

## Code Map

- `backend/internal/auth/token_envelope.go` -- reuse the owner/provider/kind/version-bound AES-GCM envelopes for both replacement tokens; add no plaintext persistence.
- `backend/internal/auth/callback.go` and `backend/internal/platform/postgres/oauth_attempts.go` -- callback currently stores access, optional refresh, and expiry; preserve this fresh-grant path while adding authorization lifecycle/credential state needed by refresh.
- `backend/internal/platform/oidc/discovery.go` -- reuse validated discovery, confidential-client HTTP Basic authentication, and the discovered token endpoint; add a refresh-token grant that does not expect an ID token.
- `backend/db/migrations/000011_*` -- add the additive authorization row version, lifecycle/status, and refresh-lease fields required for cross-process coordination and fail-closed recovery.
- `backend/internal/platform/postgres` -- implement short read/claim/finalize credential repository operations with compare-and-swap guards; no transaction spans a provider call.
- `backend/internal/portfolio/inventory.go` and `sync.go` -- remove direct `DecryptAccess` ownership and depend on the common credential-backed safe-read seam.
- `backend/internal/platform/provider/{inventory.go,account_data.go}` -- preserve strict response normalization and unauthorized classification while allowing one authorized safe-read retry through the shared source.
- `backend/cmd/findur/main.go` -- construct one credential source and inject it into inventory bootstrap and scheduled sync.
- `backend/internal/{auth,portfolio,platform/oidc,platform/postgres,platform/provider}/*_test.go` and `test/integration/**` -- cover expiry, rotation, contention, fail-closed outcomes, and one-401 retry behavior, including WireMock request and token-call counts.
- `_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md` -- record the 15-minute refresh window, shared caller boundary, and worker authorization pause.

## Tasks & Acceptance

**Execution:**
- [x] Add the additive authorization-state migration and a PostgreSQL credential repository that reads usable credentials, claims/releases guarded refresh leases, atomically installs rotated envelopes, and records reauthorization-required outcomes.
- [x] Add a discovered OAuth refresh-grant client and a domain credential source that applies expiry policy, bounded contention handling, and fail-closed ambiguous-outcome behavior.
- [x] Route both inventory bootstrap and each scheduled account-resource request through one credential-backed safe-read wrapper; remove direct persisted-access-token use from portfolio callers.
- [x] Gate scheduled claims on usable authorization, and make a fresh OAuth grant resume pending account work.
- [x] Add structured, secret-free operational logs for refresh coordination and recovery.
- [x] Update AD-4 and related architecture notes for the implemented refresh policy and worker behavior.
- [x] Wire the common source in the composition root and add focused unit/repository tests plus WireMock integration scenarios for every matrix scenario, using an expired test token to trigger proactive refresh.

**Acceptance Criteria:**
- Given any portfolio caller needs a SnapTrade credential, when it requests one, then it uses the same per-user credential source and never reads/decrypts the access-token column itself.
- Given concurrent processes find a credential expiring within 15 minutes, when they need a provider read, then one refresh exchange occurs before expiry, no database transaction remains open during it, and all successful callers use the resulting access token.
- Given a normal Findur session resumes while a provider token remains valid, when session state is checked, then no refresh is requested solely because of that login.
- Given a safe SnapTrade request returns `401`, when recovery runs, then exactly one refresh and one retry occur; another `401` transitions the authorization to reauthorization-required.
- Given a refresh exchange has ambiguous delivery or cannot safely finalize, when recovery is evaluated, then the refresh token is not replayed and provider access is withheld pending fresh authorization.
- Given authorization requires reauthorization, when the sync worker scans due work, then it skips that user's resources until a new OAuth grant is stored.

## Implementation Notes

The refresh lease and credential version are stored with the encrypted authorization row. The token exchange runs after the lease transaction commits; a guarded update installs both rotated envelopes. A stale or expired lease cannot replay the old refresh token. Discovery failures known to precede the exchange receive bounded retry and guarded lease release; ambiguous refresh outcomes and unreadable credentials require a fresh OAuth grant. The callback restores active status, advances the lifecycle, and makes pending sync work eligible again.

## Spec Change Log

## Review Triage Log

| Finding | Verdict and evidence | Route |
|---|---|---|
| Blind 1: short token lifetime can refresh on every read | Low: a token whose entire lifetime is under the 15-minute policy window can rotate on successive reads. The provider's documented lifetime is much longer; changing the policy needs expiry-history state and is disproportionate here. | Reject |
| Blind 2: discovery failure forces reauthorization | Medium: discovery runs before the token request, so reauthorization was unnecessary. Bounded pre-send retry now releases the guarded lease and returns a temporary failure. | Patch |
| Blind 3: retry token-endpoint 500/503 | False: the endpoint may have rotated the token before sending an error; replay would violate the rotating-token safety rule. | Reject |
| Blind 4: contention timeout reported as reauthorization | Medium: a live lease can still succeed when a waiter times out. The waiter now returns a transient timeout. | Patch |
| Blind 5: failed install permits old-token use | Medium: an active row could retain the lease after a database error. Reads now wait on any outstanding lease; its expiry transitions the row to reauthorization-required. | Patch |
| Blind 6: access path ignores live lease | Medium: the old access token was returned during a refresh lease. The access path now observes the lease before supplying any token. | Patch |
| Blind 7: ForceRefresh masks repository errors | Low: a transient database read was reported as reauthorization. The original database error now propagates. | Patch |
| Blind 8: expired refresh result installed | Medium: a nonzero past expiry was accepted. The source now requires an expiry later than the current time. | Patch |
| Blind 9: granted scope not stored | Low: scope persistence was already an unimplemented architecture rule before this change, and no current portfolio caller branches on scope. | Defer |
| Blind 10: expired lease does not advance generation | Medium: the expiration path did not fence the authorization lifecycle like other reauthorization paths. It now increments generation. | Patch |
| Blind 11: missing edge tests | Medium: the actionable missing paths were pre-send failure, contention timeout, and expired refresh response. Focused tests now exercise each. | Patch |
| Edge 1: concurrent 401 may rotate short-lived token twice | Medium: a contender previously treated a newly rotated but short-lived token as due. Forced recovery now reuses a newer unexpired token. | Patch |
| Edge 2: contention timeout implies reauthorization | Medium: the same live-lease timeout as Blind 4 now returns a transient timeout. | Patch |
| Edge 3: past-expiry response accepted | Medium: the same expiry validation as Blind 8 now rejects it before persistence. | Patch |
| Edge 4: database failure during reauthorization transition | Medium: a database outage can prevent the durable transition after a second 401. Reads fail while the database is unavailable; the transition error is logged, but a later process could read the active row after recovery. This requires a separate durable recovery design. | Defer |
| Verification gap: PostgreSQL reauthorization transition untested | Medium: in-memory tests could not verify the SQL guard. The sync repository test now calls the real transition and verifies that claims pause until a fresh callback. | Patch |

## Design Notes

The credential source is a domain boundary, not an HTTP middleware concern: sync jobs, inventory loads, and future server-side SnapTrade use cases invoke the same safe-read operation. The source owns expiry policy and token rotation; the provider adapter stays responsible for request shapes, rate limiting, transport handling, and normalized `401` classification. The new migration is additive because the existing `provider_authorizations` table already owns the encrypted token pair.

## Verification

**Commands:**
- `cd backend && go test ./...` -- passed; all packages, including the PostgreSQL repository tests, passed.
- `cd backend && golangci-lint run` -- passed with `0 issues.`
- `git diff --check` -- passed with no whitespace errors.
- `env COMPOSE_PROJECT_NAME=findur-refresh-complete POSTGRES_HOST_PORT=5437 ./scripts/compose-test.sh` -- passed before review fixes; browser contracts and WireMock credential scenarios passed.
- `env COMPOSE_PROJECT_NAME=findur-refresh-reviewed POSTGRES_HOST_PORT=5438 ./scripts/compose-test.sh` -- browser contracts passed; the Go WireMock runner initially failed because its new discovery mapping ranked below the static fixture mapping.
- `docker compose -p findur-refresh-reviewed --profile test run --rm integration-go` -- passed after correcting the mapping priority; all WireMock credential scenarios passed, including one token exchange for six concurrent callers and no exchange during discovery outage.
- `env COMPOSE_PROJECT_NAME=findur-refresh-rebased POSTGRES_HOST_PORT=5439 ./scripts/compose-test.sh` -- passed on the feature branch rebased onto `origin/main`; browser contracts and all WireMock credential scenarios passed.
- `cd backend && go test ./internal/auth -run TestCredentialSource -count=1` -- passed after the final log assertion.
