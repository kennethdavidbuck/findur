---
title: 'Discover and filter usable investment accounts with two requests'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — updated for explicitly renegotiated eligibility rules">

## Intent

**Problem:** Account discovery can hide usable accounts, multiply requests with connection count, and return accounts unsuitable for Findur. Provider-masked account numbers also lose distinguishing suffixes.

**Approach:** Fetch complete OAuth connection/account arrays with at most one request each, then apply the server policy in `docs/account-eligibility.md`: explicit INVESTMENT category, open/null status, completed holdings sync, no holdings-unavailable flag, active connection. Reject malformed/failed bulk responses without publishing partial results. Preserve safe masked suffixes, owner isolation, existing membership/removal behavior, and explicit confirmation. Enforce eligibility for persisted inventory and new selections. Keep the frontend's existing ready/error behavior; only correct empty-state wording for filtered results. Exercise 1,000 varied synthetic accounts through WireMock and the actual chooser, asserting exact resulting IDs, grouping, two discovery calls, and loading/interaction performance. No partial-loading UI, large-confirmation scheduling, database batching, activity pagination, refresh scheduling, migration, deployment, or large-list UI redesign.

</frozen-after-approval>

## Implementation Notes

- Current branch: `fix/bulk-account-discovery`. Follow-up discussion specs are committed separately on local `docs/account-scale-followups` (3b32365); no push or PR is authorized for this close-out.
- User approved the final server rules after reviewing SnapTrade documentation. Null status is acceptable with all other requirements met; unknown category is not. Transactions-sync completion, raw type, deprecated metadata, positive balance, and recency are not additional eligibility requirements.
- Provider overlay/materializer/generated contract use GET /accounts. Both inventory endpoints return complete arrays with no paging contract; unsupported envelopes fail validation. Preserve 100-connection, 5,000-account, and 8 MiB account-list bounds; other provider responses retain their 1 MiB cap.
- Inclusion allows the inventory scale of 5,000 IDs and an independently tested 1 MiB body; other request bodies retain 4 KiB.
- The previously added frontend partial-result rendering, availability checks, and retry/draft reconciliation were removed at the user's direction. Only empty-state EN/FR wording changes remain in production frontend code.
- Successful filtering with no qualifying accounts yields empty; all-disabled connections retain disabled. Provider errors publish no partial payload. Existing persisted snapshots are filtered and legacy selectable flags cannot bypass admission checks for new additions.
- Updated deterministic fixture: 1,000 accounts, 50 connections, 392 returned/eligible/selectable/visible accounts, 608 exclusions. Explicit slot oracle includes null status, investment cash, and incomplete transaction sync; excludes unknown category, missing holdings sync, unavailable holdings, closed/archived/unavailable status, and disabled connections.
- The chooser test stops at reviewing the draft; large-confirmation throughput is a separate follow-up. Last-row WebDriver clicks use immediate test scrolling to avoid racing product smooth scrolling.
- Preserve user-provided files separately; do not include their screenshot in the commit. Fixtures are entirely invented.

## Verification

- Provider acceptance matrix and boundary tests passed; cached-inventory/admission regressions passed. `cd backend && GOCACHE=/tmp/findur-go-cache go test -race ./...` passed after the final cached-projection and boundary-test changes. A targeted sandbox attempt could not access Docker; the full-suite run with Docker access passed.
- `cd backend && golangci-lint run`: passed, 0 issues, with the installed linter/cache paths configured through the environment.
- `cd frontend && npm test -- --run && npm run typecheck && npm run lint && npm run build`: passed, 73 tests. Two existing react-refresh warnings remain; local Vite warns about unset VITE_BUILD_SHA, while the composed build supplies it.
- `node --check test/integration/browser-large-inventory.mjs`, fixture exact-ID cross-check, and `git diff --check`: passed.
- `docker compose build backend frontend`, then final `docker compose build backend`, `docker compose up --no-build --wait --force-recreate wiremock`, `docker compose up --no-build --wait backend`, and `docker compose --profile test run --rm integration`: passed. Active connections with only excluded accounts persist empty and show accurate browser copy; the mixed fixture returns exactly 392 accounts and retains provider-masked suffixes.
- Final composed metrics: 1,000 source accounts / 50 connections, 392 selectable visible accounts, exactly 2 discovery calls; fetch/persist 1,356 ms, navigation/render 265 ms, select-all 27 ms, deselect-all 20 ms, last-row click 106 ms, review 66 ms. Local smoke measurements only; no production latency or large-confirmation guarantee.

## Review Triage Log

- Earlier pending-state, saved-choice retry, recovered-draft consistency, and partial-message findings are superseded: their frontend implementation and added tests were removed with the abandoned partial-result scope.
- Earlier per-connection failure-order and row-quarantine findings are superseded by bulk all-or-none validation; current tests require no partial payload on failure.
- Earlier partial-inventory confirmation coverage is replaced by persisted-filter/admission tests under the final server rules.
- Patched: request item-count and byte-limit regressions cover independent limits; non-inclusion endpoints retain their original bound.
- Patched: stress assertions check exact account IDs/ownership and labels, include a real last-row click and timing, and preserve primary errors during cleanup.
- Deferred by user: large-confirmation success/throughput acceptance is held on the separate follow-up branch; no success claim is made here. Large-list UI redesign remains in the deferred ledger.

### Final policy review

- Medium, patched: cached ready snapshots could project zero accounts while retaining ready status. Repository now projects empty only for that successful state; tests preserve pending/failure/disabled states and avoid refresh or historical rewrites.
- Medium, patched: accepted cached null-status accounts retained old eligibility/reason flags. Matching rows now project eligible/selectable/ready consistently, with a regression retaining historical storage unchanged.
- Medium, deferred: excluded committed accounts cannot be removed through the chooser because they are absent from inventory. API removal remains possible; automatic removal or retaining excluded chooser rows would contradict this narrowly approved filter. This extends the already-deferred absent-inventory membership UI concern; documented explicitly and recorded on the follow-up branch.
- Low, patched: provider byte-limit tests used invalid identities and could fail for another reason. Otherwise valid payloads now pass at exactly 1 MiB / 8 MiB and fail one byte above.
- Low, patched: duplicate connection validation masked the missing-ID case. Independent cases now reach each failure.
- Low, patched: connection-count ceiling lacked boundary assertions. Valid 100 connections pass and 101 fail before the account request.
- Low, patched: null status plus null/missing/unrecognized category lacked direct matrix coverage. All combinations now assert exclusion.
- Low, patched: composed fixture did not verify provider-masked suffix retention. It now includes masked numbers and asserts full rendered account labels.
- Low, patched: filtered-empty behavior lacked composed coverage. An active-connection fixture containing only excluded accounts verifies persisted empty state and browser wording before the mixed fixture.
- Low, patched: policy documentation omitted daily provider caching. It now states that explicit retry reloads SnapTrade cached inventory and does not force brokerage refresh.
