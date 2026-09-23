---
title: 'Add scheduled connection and account inventory synchronization'
type: 'feature'
created: '2026-09-22'
status: 'done'
route: 'dispatch'
baseline_commit: '49c1dae4c99064495a4fc4560ce34c8008f6a292'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/spec-schedule-portfolio-account-sync.md'
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A user's first account inventory can remain incomplete after a failed bootstrap, and successful inventories never discover later connection or account changes.

**Approach:** Extend the existing minute worker so each leased 45-second pass first refreshes at most one due user-level `/authorizations` + `/accounts` inventory bundle, including the provider-denominated total value of each account, then drains the existing included-account resource queue. Repair never-synced and retryable legacy states automatically and refresh successful inventory heads every 24 hours without browser activity.

## Boundaries & Constraints

**Always:** Claim a new generation transactionally; call the provider outside transactions through the shared credential source/client, one-second gate, circuit breaker, timeout, singleton lease, and safe structured logging; validate the complete graph before atomically publishing it and upserting stable `(user_id, account_id)` identities. Persist `/accounts.balance.total` as nullable amount and ISO currency on the immutable inventory-account generation without adding another provider call. Treat exact age 24 hours as due. Preserve sync state, balances, positions, activities, and missing-account history by stable identity, but atomically remove currently included accounts that no longer satisfy current inventory eligibility and advance the inclusion lifecycle to fence stale work. Block inventory claims during pending inclusion or unexpired resource claims, and reject inclusion saves during an inventory claim with the existing retryable conflict contract. Retain the last good head on scheduled failure; expose categorized failure only without a good head. Retry after 1, 2, 4, 8, 16, then 32 minutes, honoring a later provider `Retry-After`; success clears failure state, and inactive authorization pauses work.

**Never:** Add public HTTP/frontend APIs, couple refresh to login or page activity, fetch financial resources for unincluded accounts, partially publish a malformed bundle, delete retained versions/membership/history, weaken lifecycle guards, or touch the staged screenshot or untracked `quries.txt`. Record unified inventory/financial-version retention as deferred work rather than adding inventory-only deletion.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Repair | Active user has no head or retry-due failed/legacy inventory | One generation claims and publishes the complete two-call bundle | Bounded retry; categorized failure remains visible without a good head |
| Scheduled refresh | Successful head is exactly 24 hours old | Inventory runs first, once per pass; remaining time drains resource work | Failure retains and serves the good head |
| Concurrent lifecycle | Pending inclusion, active resource claim, or competing worker | Inventory is not claimed; one worker wins when eligible | Inclusion during inventory claim gets retryable conflict; expired claims recover |
| Inventory changed | Accounts are stable, new, missing, or no longer eligible | Stable dependencies and history survive; new identities appear; ineligible included accounts leave active inclusion and the lifecycle advances | Invalid cross-connection references publish nothing |

</frozen-after-approval>

## Code Map

- `backend/db/migrations/000012_*` -- add inventory failure/backoff state, due-query support, and immutable account-total fields without deleting versions.
- `backend/internal/portfolio/inventory.go` -- expose worker claim/finalization ports and shared refresh orchestration while preserving browser bootstrap/retry behavior.
- `backend/internal/platform/postgres/inventory.go` -- select/claim one eligible user, enforce inclusion/resource exclusions, retain last-good heads, back off failures, and atomically publish identities.
- `backend/internal/portfolio/sync.go` -- run at most one inventory refresh before draining account resources under the existing pass lease/deadline.
- `backend/internal/platform/postgres/inclusion.go` -- reject saves against a pending inventory generation using the existing conflict response.
- `backend/internal/platform/postgres/oauth_attempts.go` -- make paused unauthorized inventory immediately eligible after authorization becomes active.
- `backend/cmd/findur/main.go` -- inject the shared inventory service into the existing worker; do not create another client or scheduler.
- `backend/internal/{portfolio,platform/postgres,platform/provider}/*_test.go` -- cover order, due policy, graph publication, failure preservation/backoff, and concurrency.
- `backend/internal/platform/postgres/inventory_test.go` -- seed the 22-connection, 15-account partial unavailable legacy shape and prove unattended repair makes exactly the two inventory calls.
- `_bmad-output/planning-artifacts/epics.md`, prior sync policy notes, and `deferred-work.md` -- supersede bootstrap-once policy and record unified retention follow-up.

## Tasks & Acceptance

**Execution:**
- [x] Add the compatible inventory scheduling migration and repository claim/finalization behavior.
- [x] Integrate one inventory operation into each worker pass before existing account-resource draining.
- [x] Add lifecycle, publication, failure, concurrency, and legacy-repair tests plus the integration fixture.
- [x] Update policy/spec documentation and add the unified retention deferred task.

**Acceptance Criteria:**
- Given the affected active legacy user with 22 connections and partial unavailable inventory, when the worker runs without a browser request, then a complete new generation publishes after exactly `/authorizations` then `/accounts`.
- Given a successful head or dependent account data, when refresh changes the provider graph, then stable account dependencies remain intact and stale provider results cannot publish.
- Given an included account no longer satisfies current inventory eligibility, when a successful generation publishes, then it is removed from active inclusion without deleting its stable identity or financial history.
- Given `/accounts` supplies a total balance, when the generation publishes, then its exact normalized amount and provider currency are retained on that immutable account row; absent totals remain null and malformed totals publish nothing.
- Given due inventory and account-resource work, when a leased pass runs, then no more than one inventory is attempted first and the remaining deadline drains balances, positions, and activities normally.

## Implementation Notes

- Scheduled claims carry the authorization lifecycle generation and finalize under the existing user serialization lock; expired, rotated-authorization, deactivated-owner, blocked-inclusion, and replaced-generation results are discarded.
- The `/accounts` adapter retains the provider-denominated total with decimal and ISO-currency validation. Migration 12 stores it on immutable inventory-account rows; no new provider call or public response field was added.
- Successful publication upserts stable identities, removes only active membership for newly ineligible accounts, fences account work by advancing inclusion lifecycle, and leaves retained financial versions and heads untouched.
- The affected production shape is represented with 22 connections, 15 partial legacy accounts, an unavailable head, active authorization, and no browser request.

## Spec Change Log

- 2026-09-22: At the user's direction, expanded the approved bundle to persist the total account value already returned by `/accounts`, together with its provider currency and without an additional provider request.

## Review Triage Log

- **medium / patch:** Scheduled provider-failure orchestration lacked a service-boundary assertion. Added a `RefreshDue` test proving `FinalizeScheduled` records the categorized retry before returning the provider error.
- **medium / patch:** Inventory backoff coverage stopped after one minute. Added table-driven consecutive failures covering 1, 2, 4, 8, 16, 32, the 32-minute cap, and a later provider `Retry-After`.
- **medium / patch:** Publication tests removed only a missing account. Added present-but-ineligible account, sync, and connection cases that assert membership removal with identity/sync retention.
- **medium / patch:** Fresh-authorization inventory reset was unverified. Added a last-good-head/future-backoff test proving callback reset and immediate claimability.
- **medium / patch:** Account draining after an inventory error was unverified. Added a worker test proving resource claims continue after the scheduled inventory retry is recorded.
- **medium / patch:** A user selected as due could become non-due before its inventory row was locked. Added a post-lock due recheck and serialized browser preparation through the existing user lock.
- **medium / patch:** Inventory finalization and inclusion preparation acquired inventory/inclusion locks in opposite orders. Finalization now locks the existing user serialization row first, matching inclusion and OAuth lifecycle ordering.
- **medium / patch:** Scheduled finalization did not reject an expired lease. The guard now requires `claim_expires_at` to remain strictly later than finalization time, with exact-boundary coverage.
- **high / patch:** Scheduled claims did not fence authorization lifecycle rotation. Claims now carry authorization generation and successful publication requires the same active generation.
- **medium / patch:** Scheduled finalization did not recheck the existing `users.active` invariant. It now locks and validates the owner before publication, with a stale-result test.
- **medium / patch:** The independent review separately identified expired-lease publication. The same strict lease guard and regression case disprove publication at or after expiry.
- **medium / patch:** A last-good unauthorized attempt could retain retry timing across fresh authorization. OAuth finalization now clears inventory backoff for retained-head failures as well as exposed unauthorized failures.
- **low / patch:** Rejected legacy repair could derive a null current status from a missing head row. Recovery now coalesces to `unavailable`.
- **low / patch:** System-generated inclusion history used a client-guessable idempotency key. It now uses a system-prefixed random change UUID.
- **false / reject:** `RefreshDue` ignoring the repository `accepted` flag does not report publication; its boolean means a claim was processed, and safely discarded stale results still complete that operation without mutating the head.
- **maybe-false / defer:** The due query's population-scale plan cannot be graded without representative row counts and `EXPLAIN (ANALYZE, BUFFERS)` evidence; recorded for measurement rather than speculative indexing.
- **medium / patch:** The independent review also identified incomplete retry-policy coverage. The full bounded sequence and provider override are now asserted.
- **medium / patch:** Retained financial history was inferred from foreign-key structure but not directly tested. Publication coverage now seeds and asserts all three financial versions and heads after removal.
- **medium / patch:** Authorization rotation and owner deactivation between claim and finalize lacked coverage. Both transitions now have stale-publication regression cases; the user-row lock serializes concurrent lifecycle transactions.

## Design Notes

Inventory attempt state and served head are distinct: scheduling may advance the claimed generation and record failures while reads continue projecting the last successful immutable head. The inventory and resource claims share the singleton pass but retain separate one-minute recovery leases and lifecycle fences.

## Verification

**Commands:**
- `cd backend && GOCACHE=/tmp/findur-go-cache go test -race ./...` -- all backend and repository race tests pass.
- `cd backend && golangci-lint run` -- mandatory lint passes with the exact command reported.
- `docker compose --profile test run --rm integration` -- unattended legacy repair and existing integration journeys pass.
- `test -z "$(gofmt -l backend)" && git diff --check` -- Go formatting and patch whitespace are clean.
