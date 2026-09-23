---
title: 'Implement resource diagnostic feedback'
type: 'feature'
created: '2026-09-23'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '0bdc180a1742ed163774c17ec6b90c9a42a38f49'
context:
  - '_bmad-output/specs/spec-resource-diagnostic-feedback/SPEC.md'
  - '_bmad-output/specs/spec-resource-diagnostic-feedback/failure-modes.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Valid inventory and portfolio responses collapse provider preparation, connection repair, and Findur account-data fetch failures into generic unavailable or syncing messages. The cause, retry timing, and last success disappear after reload, making retained data and recovery actions misleading.

**Approach:** Implement the canonical resource diagnostic contract across provider normalization, durable synchronization state, OpenAPI, and the Portfolio Showcase status area. Keep the pre-existing account selector and its per-account availability reasons unchanged; connection diagnostics remain contract data rather than new onboarding presentation. Keep SnapTrade-side connection/preparation facts distinct from Findur-side resource failures, preserve usable saved data, remove the repetitive per-account “Currencies stay separate” annotation while retaining the no-conversion policy, and make the richer scrubbed inventory the primary local mixed-state demo with paired populated data plus deterministic degraded-state scenarios.

## Boundaries & Constraints

**Always:** Use bounded reasons and actions; persist only safe categorical metadata; keep provider and Findur success timestamps semantically distinct; retain owner isolation, lifecycle guards, stale values, EN/FR parity, and per-resource presentation; clear a diagnostic only when the affected resource succeeds. Keep the pre-existing account selector UI and existing per-account availability reasons only: do not add a global connection summary, Last checked block, per-connection status cards, or new onboarding Check again/Retry/Reconnect controls. The selector may show muted, disabled `unsupported_category` and `connection_disabled` accounts for context, but server-side inclusion admission must revalidate the current selectable predicate; selectable accounts sort before disabled accounts, with a locale-aware masked-label order and stable ID tie-breaker inside each group. Keep exactly one Portfolio-level “Check again” link only while at least one dataset is in initial `sync_pending`; it refreshes status in place without moving focus or scroll. The expected initial-sync banner uses the established blue informational treatment and is not repeated as an alert in every dataset section. Dataset diagnostics are passive; reconnect remains only for genuine authorization or connection repair. Successful zero-row datasets use friendly resource-specific empty copy rather than transport-shaped “complete dataset” language. Use existing semantic colors sparingly and accessibly: danger only for errors/action-required states, warning only for unexpectedly delayed or degraded states, success only for confirmed health, informational blue for the single expected initial-sync banner, and neutral styling for ordinary information; never rely on color without copy, icon, or state text.

**Never:** Store or expose raw provider errors; change account eligibility, scheduling, retry policy, or inclusion membership; infer recovery timing; add webhooks, provider-specific prose, support workflows, or the separate browser-safe non-2xx contract.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| Provider discovery | Empty, unsupported-only, disabled, or holdings-pending connection | Connection diagnostic explains the safe condition and valid action | Unknown combinations use `unknown`; no raw detail |
| Findur fetch failure | Resource receives authorization, rate-limit, provider, or malformed/unusable failure | Matching dataset retains permitted rows and exposes reason, retry time, and last Findur success after reload | Other datasets retain their own state; retry remains scheduled |
| Recovery | Previously failing resource publishes successfully | New head and checkpoint publish atomically; obsolete diagnostic clears | Stale workers cannot write or clear diagnostics |
| Presentation | Current, stale, expired, syncing, or unavailable dataset, including zero rows | Existing dataset section shows truthful localized copy and timestamp; one top banner offers “Check again” only during any initial dataset sync; dataset sections have no retry/check buttons | Inclusion pending no longer relabels unrelated failures; status refresh preserves focus and scroll; reconnect appears only for authorization/repair |
| Local mock default | Boot the normal Compose stack with WireMock | Scrubbed 22-connection/29-account inventory is returned as a deliberate mixed-state demo: healthy realtime/delayed accounts, populated and genuinely empty holdings, named selectable positions/balances/activities failure scenarios, initial sync, disabled/repair-required, unsupported-only, and no-account connections are visible where product policy permits; every selectable account has paired balances, positions, and activities responses; synthetic account labels plainly and truthfully name the scenario they model | No account remains pending merely because its mock endpoint is missing; choosing a named failure account triggers that exact mapped failure; unselectable repair/unsupported accounts retain the existing passive availability reason; scenario labels, suffixes, and values remain synthetic and non-recognizable |
| Diagnostic E2E | Test activates a named WireMock override | One provider or Findur-sync condition is reproduced deterministically and reset without changing the default dataset | Scenarios remain isolated and do not rely on request ordering |
| Production backfill | Migration is applied to an existing successful inventory | Migration marks the inventory for one guarded worker refresh; it performs no provider call itself and clears the marker only after publication | Existing retry timing and immutable evidence timestamps are not rewritten; account datasets are not forced merely to create empty diagnostics |

</frozen-after-approval>

## Code Map

- `backend/internal/platform/provider/{inventory.go,account_data.go}` and `backend/internal/portfolio/{inventory.go,inclusion.go,sync.go}` -- reuse normalized provider states, bounded `ProviderError`, resource claims, and retry timing; do not retain provider bodies.
- `backend/db/migrations/000009_portfolio_sync.up.sql`, new `000013_*` migration, and `backend/internal/platform/postgres/{sync.go,showcase.go,inventory.go}` -- extend owner-scoped sync/inventory state, guarded failure persistence, success clearing, and diagnostic projection.
- `backend/api/openapi.yaml`, generated Go/TypeScript bindings, and `backend/internal/platform/httpapi/authorization.go` -- define and serialize optional `ResourceDiagnostic` on connections and dataset contexts.
- `frontend/src/{inventory.ts,showcase.ts,pages/PortfolioShowcasePage.tsx,styles.css}` -- validate diagnostics and render Portfolio resource reason, timestamps, and the bounded top-level contextual action without changing the account selector.
- `test/fixtures/wiremock` and `test/integration` -- adopt the 22-connection/29-account scrubbed inventory from `docs/mock-data-set` as the default mixed-state demo; give connections/accounts clear synthetic scenario-oriented names, pair selectable accounts with varied populated/empty balances, positions, and activities, preserve provider-side lifecycle variants, and use temporary WireMock admin overrides for isolated Findur failures and recovery.

## Tasks & Acceptance

**Execution:**
- [x] Domain, migration, and PostgreSQL repositories -- persist bounded resource failure reason/resource/time, reuse retry and per-resource success timestamps, project safe diagnostics, clear them atomically on matching success, and mark existing inventories for one guarded post-deploy refresh without issuing external work inside the migration.
- [x] Provider inventory and OpenAPI/HTTP mapping -- classify connection discovery conditions, expose the shared diagnostic contract, and regenerate bindings.
- [x] Frontend Portfolio Showcase -- replace inferred generic causes with exhaustive localized passive diagnostic presentation while keeping stale or empty evidence visible; use friendly balance/position/activity-specific empty copy; retain one non-jumping top “Check again” only for initial dataset sync, reconnect only for authorization/repair, and remove the redundant balance-summary “Currencies stay separate” line without adding aggregation or conversion. Do not add connection diagnostics, summaries, timestamps, or controls to onboarding/account selection.
- [x] Fixtures and tests -- cover reason/action matrices, persistence/recovery/guards/isolation, generated contracts, EN/FR accessibility, mixed connections, zero-row resources, the rich default scrubbed stack, truthful scenario-oriented account names, varied populated account data, named default resource failures, passive unsupported/repair reasons, and deterministic WireMock failure/recovery installation/reset.

**Acceptance Criteria:**
- Given a successful degraded response, when either sync layer has a known safe condition, then the affected connection or dataset explains the reason and valid recovery without exposing provider detail.
- Given a Findur resource failure followed by reload and later success, when Portfolio is viewed, then retained data and diagnostic timestamps survive reload and the diagnostic clears only after that resource publishes.
- Given mixed or empty connection results, when account setup renders, then the pre-existing selector and per-account availability reasons remain unchanged, with no new global summary, Last checked block, connection cards, or diagnostic controls.
- Given any supported diagnostic in English or French, when rendered with zero or stale rows, then text, accessibility semantics, and actions remain truthful and independently testable.
- Given ordinary initial sync, progress, or successful empty data, when rendered, then it uses neutral or blue informational treatment and status semantics rather than an alert role, alarming icon, warning yellow, or error red; yellow is reserved for unexpected delay/degraded-but-usable state and red for blocking/action-required failure.
- Given one or more datasets in initial sync, when Portfolio renders or “Check again” is used, then exactly one top-level check link appears and refreshes without focus/scroll movement; given any later transient resource failure, then no per-dataset retry/check control is shown because the worker retries automatically.
- Given a normal local mock boot, when the user explores selection and Portfolio, then the scrubbed accounts demonstrate varied connection/account lifecycle states and populated/empty financial datasets without accidental pending states; given an isolated diagnostic browser scenario, each overridden Findur failure produces and later clears the expected diagnostic.
- Given an existing production inventory when the migration deploys, then the normal worker refreshes its provider-derived diagnostic state once without falsifying retry metadata or forcing account-data refreshes.

## Implementation Notes

Implemented the shared bounded diagnostic contract through provider normalization, durable inventory/account resource state, OpenAPI bindings, HTTP mapping, and localized Showcase presentation. Migration `000013` marks only existing successful inventory heads for one normal guarded worker refresh; it does not issue provider work or rewrite publication, retry, or account-dataset evidence.

The existing account selector remains the only onboarding presentation. Its two approved passive account classes remain muted and disabled, selectable rows sort first, and PostgreSQL inclusion admission independently rechecks the current server-side selectable predicate.

The default local fixture now provides 22 connections and 29 synthetic scenario-labelled accounts with paired datasets, truthful named resource failures, passive repair/unsupported cases, and deterministic temporary failure/recovery overrides. Healthy position freshness uses WireMock retrieval-time rendering so it cannot age into an expired default state.

## Spec Change Log

## Review Triage Log

| ID | Verdict | Route | Evidence |
|---|---|---|---|
| B1 | false | reject | `ShowcaseRepository` deliberately fails closed when sync mode is unknown, and `ClassifyFreshness` explicitly classifies that mode as unavailable; the claimed safe retained presentation is therefore not established and the suppression predates this change. |
| B2 | medium | patch | The new broad unusable branch maps both `unavailable` and `unknown` account sync states to `sync_pending`; those states can occur and would falsely describe an outage as ordinary initial preparation. |
| B3 | false | reject | The approved public contract attaches diagnostics to `InventoryConnection` and `DatasetContext`, not `PortfolioInventory`; a request-level inventory state continues to expose its bounded state and retry time without inventing a new top-level API surface. |
| B4 | medium | patch | `connectionDiagnostic` selects the first matching condition, so mixed pending/unavailable/unsupported-only rows receive a specific cause even though the approved matrix requires mixed ambiguous combinations to use `unknown`. |
| B5 | false | reject | The canonical failure matrix explicitly permits `provider_unavailable` to recommend `retry` or `wait`; Portfolio remains passive and renders no retry control, while the worker still performs the scheduled retry. |
| B6 | medium | patch | Both new browser guards accept independently valid but incompatible reason/action pairs; a malformed `provider_unavailable` + `reconnect` payload would render the global reconnect action contrary to the contract. |
| B7 | low | patch | The new diagnostic timestamps are accepted as arbitrary strings and then formatted as dates, so malformed successful payloads can render `Invalid Date`; a direct date-time guard is a small correction. |
| B8 | medium | patch | Per-dataset diagnostics are required to be passive, but the implementation assigns `status` or `alert` live-region roles to every affected dataset, producing repeated announcements and alarms. |
| B9 | medium | patch | The new semantic freshness classes are overridden by the later, more-specific dataset freshness rule, so the reviewed blue/warning/error palette does not reach the intended compact state labels. |
| B10 | medium | patch | When an in-place check fails, the focused control is removed; after retry succeeds, `focusedHeading` remains set and no deterministic focus target is restored. |
| B11 | false | reject | The named initial-sync account intentionally uses complete paired endpoints: its scenario is the real interval between inclusion and the first worker publication, after which successful completion is expected. |
| B12 | false | reject | Permanent first-fetch failure accounts model no-data failures; retained-data recovery is intentionally exercised through the temporary override mechanism and repository behavior rather than requiring every default failure account to have prior history. |
| B13 | medium | patch | The temporary override integration only probes WireMock directly; it does not prove that the worker persists, reloads, presents, and later clears the diagnostic. |
| E1 | medium | patch | This independently confirms B4: mixed unselectable reason classes are reachable and the current precedence emits a misleading specific connection cause. |
| E2 | false | reject | The backend does not produce `freshness=unavailable` with retained rows: unknown-mode accounts take the fail-closed empty path, while retained known-mode datasets classify current, stale, or expired. |
| E3 | medium | patch | This independently confirms B6: semantically incompatible diagnostic pairs pass runtime validation and can expose the wrong recovery control. |
| E4 | low | patch | This independently confirms B7: diagnostic timestamps are not validated before formatting and malformed values can render untruthful evidence. |
| E5 | false | reject | This repeats B3, but the frozen API scope intentionally has no top-level inventory diagnostic; connection diagnostics and request state are the approved public surfaces. |
| E6 | medium | patch | This independently confirms B13: direct WireMock assertions do not verify the claimed Findur failure/reload/recovery flow. |
| V1 | medium | patch | Pre-verified gap: the override is removed before any Findur worker or browser request observes it, so end-to-end classification, persistence, UI, and recovery can all regress undetected. |
| V2 | medium | patch | Pre-verified gap: the scoped-clear assertion starts with the other resource diagnostic already null, so clearing every diagnostic would still pass. |
| V3 | medium | patch | Pre-verified gap: the repository test manually sets the backfill marker and no migration test verifies the `000013` update predicate or preservation of retry/publication evidence. |
| V4 | medium | patch | Pre-verified gap: helper and parser tests do not prove provider connection diagnostics survive persistence, reload, and HTTP serialization. |
| V5 | medium | patch | Pre-verified gap: no UI test combines retained rows, a degraded diagnostic, `lastSuccessfulAt`, and `retryAt`, so the defining degraded presentation can regress unnoticed. |
| V6 | low | patch | Pre-verified gap: the selector test covers only distinct English labels, leaving the French collator branch and stable ID tie-breaker unobserved. |
| V7 | medium | patch | Pre-verified gap: no contract test enumerates all selectable default accounts across balances, positions, and activities, so a missing mapping can leave a demo account accidentally pending. |

## Design Notes

`lastSuccessfulAt` initially means Findur successfully fetched and published the resource. SnapTrade brokerage-sync timestamps are not placed into that field. `unusable_data` maps to safe dataset reason `unknown`; inclusion keeps its existing `unusable_data` value.

Human clarification (2026-09-23): connection diagnostics remain available in the shared API contract but are not presented as new onboarding/account-selection UI. Only the scrubbed mock account names are changed there so each selectable scenario is easy to identify.

Human clarification (2026-09-23): diagnostic color is a secondary cue. Empty information is neutral, the single expected initial-sync banner uses established informational blue without repeated dataset alerts, unexpectedly delayed/degraded states use the established warning palette, and action-required/errors use the established danger palette.

Human clarification (2026-09-23): the existing selector may expose the limited greyed-out unsupported-category and connection-disabled accounts with their passive reasons. All selectable accounts appear first and both availability partitions sort consistently by localized masked label with a stable account-ID tie-breaker. Inclusion save continues to enforce the current server-side selectable predicate, so a crafted request cannot newly include a displayed disabled account.

## Verification

**Commands:**
- `cd frontend && npm run generate:api && npm run typecheck && npm run lint && npm test -- --run && npm run build`
- `cd backend && go generate ./... && go test ./...`
- `cd backend && golangci-lint run`
- `git diff --check`

**Results (2026-09-23):**
- Frontend generation, typecheck, lint, 114 tests, and production build passed. Lint reported four pre-existing warnings and no errors.
- Backend generation and the full race-enabled test suite passed, including Docker-backed PostgreSQL migration/repository tests.
- The mandatory exact `cd backend && golangci-lint run` invocation could not start because `golangci-lint` is not installed on the default shell `PATH` (`command not found`). The same repository lint ran successfully with the locally cached `golangci-lint` v2.13.0 executable and reported `0 issues`.
- The full clean isolated Compose integration/browser suite passed with `integration contracts passed`, including the rich default, initial sync, selector stress path, and deterministic failure/reload/recovery.
- Integration script syntax checks and `git diff --check` passed.
