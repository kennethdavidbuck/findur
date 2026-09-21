---
title: 'Story 1.1: Select and Confirm Included Accounts'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
baseline_commit: '5fd8555a149eaf65dadf39b3aa2227e1435f4b4e'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Users can inspect masked SnapTrade inventory but cannot choose which accounts Findur may use. The generic eligibility flag also leaves legitimate-looking live accounts marked “Not eligible” without an actionable reason.

**Approach:** Add an owner-scoped, idempotent inclusion workflow: remove and purge immediately, fetch pending additions outside transactions, then atomically publish only complete guarded additions. Add an accessible bilingual selection, confirmation, and recovery UI with specific usability reasons.

**Eligibility decision:** Use probe-backed eligibility. An account with incomplete provider status or category may be selected provisionally, but becomes included only when the confirmed allowlisted fetch proves usable investment data. A failed or inconclusive probe leaves it excluded and publishes nothing.

## Boundaries & Constraints

**Always:** Default to none included; derive ownership from the Actor; distinguish draft, pending, and committed sets; fence changes; publish required data only as complete immutable per-account datasets; use `private, no-store`; preserve bilingual accessible responsive behavior.

**Never:** Treat connected as included; publish partial additions; restore removals; hold transactions across provider calls; persist/log raw payloads or secrets; convert currencies; call forbidden provider endpoints; accept a client owner ID; hand-edit generated code.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| First visit | Masked inventory; no committed set | Named group starts empty; Select All affects usable rows | Disabled rows state why |
| Add | Version, idempotency key, account IDs | Pending until datasets and membership publish atomically | Failure or stale guard publishes nothing; offer scoped retry |
| Remove/mixed | Committed IDs omitted, additions optional | Fence and purge removals before provider work | Failed additions never restore removals |
| Duplicate/stale/foreign | Reused key, old version, or foreign ID | Replay identical result or reject without disclosure/mutation | Safe conflict/reload response |

</frozen-after-approval>

## Code Map

- `backend/internal/{portfolio/inventory.go,platform/provider/inventory.go}` -- reuse scoped prepare/call/finalize and bearer/error patterns; add usability reasons without early financial fetches.
- `backend/internal/{portfolio/inclusion.go,platform/postgres/inclusion.go}`, `backend/db/migrations/000006_*` -- new guarded workflow/schema; do not alter `000005`.
- `backend/provider/{oauth-bearer-overlay.yaml,cmd/materialize/main.go}` -- allowlist balances, all positions, and bounded activities; generated files remain generated.
- `backend/api/openapi.yaml`, `backend/internal/platform/httpapi/authorization.go`, `backend/cmd/findur/main.go` -- define and wire owner-private reads and defended mutations.
- `frontend/src/{inventory.ts,pages/PortfolioPage.tsx,i18n.tsx,styles.css}` -- add draft/committed selection, dialog, announcements, reasons, and recovery.

## Tasks & Acceptance

**Execution:**
- [x] `backend/provider/**`, `backend/internal/platform/{provider,config}/**` -- add three account-data operations, 30-day/500-row activities, minimized normalization, reasons, and bearer tests.
- [x] `backend/db/migrations/000006_*`, `backend/internal/{portfolio,platform/postgres}/**` -- add stable identities, inclusion/change records, typed immutable datasets, fencing, idempotency, atomic publication, and immediate purge.
- [x] `backend/api/openapi.yaml`, `backend/internal/platform/httpapi/**`, `backend/cmd/findur/**` -- add owner-scoped no-store GET/POST, expected version/key, CSRF/origin/fetch defenses, conflicts, wiring, and generated clients.
- [x] `frontend/src/{inventory.ts,pages/PortfolioPage.tsx,i18n.tsx,styles.css}` -- add named checkboxes, usable-only tri-state Select All, draft/committed summary, consequence dialog, cancel focus return, pending/recovery, and EN/FR.
- [x] Backend/React tests and `test/{fixtures/wiremock,integration}` -- prove the matrix, concurrency, isolation, no early financial calls, partial publication, or forbidden endpoints.

**Acceptance Criteria:**
- Given masked inventory, when inclusion opens, then none is included by default and the group distinguishes draft, pending, committed, usable, and specifically disabled rows.
- Given a draft, when confirmation opens or is cancelled, then masked additions/removals and consequences are explicit, coverage is unchanged, and focus returns.
- Given additions, when all required datasets normalize and guards match, then datasets and membership commit atomically; duplicates have one effect.
- Given failed/stale additions or mixed changes, when work settles, then no partial broadening occurs, removals remain purged, and recovery is scoped.
- Given any request, when authorization runs, then only the Actor's accounts can be observed or changed.

## Implementation Notes

- Added a removal-first inclusion state machine with owner locks, expected-version fencing, concurrent idempotency replay, inventory/lifecycle guards, and atomic per-account dataset publication. Provider work occurs only after the prepare transaction commits.
- Added exact-string decimal normalization and typed PostgreSQL `numeric` rows for balances, positions, and activities. Activities are limited to the most recent 30 days and 500 rows; unsupported currency-denominated data fails closed without publishing an addition.
- Added specific inventory usability reasons and provisional selection for incomplete provider status/category. The browser maintains separate draft and committed coverage and exposes scoped failure recovery without restoring removals.
- Corrective audit: confirmed removals now advance the inclusion lifecycle before purge, and reauthorization advances the same fence. PostgreSQL tests prove both an older mixed-change finalizer and a pre-reauthorization finalizer cannot recreate membership or dataset heads.
- Corrective audit: balances, positions, and activities now publish through separate immutable version/head tables with dataset-specific observation, retrieval, and publication metadata. Their three heads and membership move in one transaction; a forced second-account row failure leaves every version, head, and membership unpublished and can be retried safely.
- Corrective audit: missing provider-supplied position `as_of` proof fails closed, initial-sync-pending accounts are unselectable, and provisional status/category accounts remain probeable only after complete required endpoint responses.
- Corrective audit: usable-only tri-state Select All preserves non-usable coverage, failed/pending changes reconstruct only their scoped additions while preserving removals, and a new idempotency key at the current version safely supersedes durable pending work.
- Corrective audit: the confirmation experience now uses React Aria modal/dialog primitives with safe initial focus, Escape dismissal, focus restoration, EN/FR provider-access versus inclusion boundaries, category/coverage summaries, private-purpose limits, and complete purge consequences. Removal confirmation uses the labelled Constellation danger action and light/dark danger tokens.
- Corrective audit: inclusion prepare/finalize now lock the active owner row before inclusion and inventory state, while reauthorization updates inclusion before inventory after its user/authorization lock. The shared user → inclusion → inventory → dataset order removes the previous inversion; a race-detector PostgreSQL test overlaps reauthorization with both prepare and finalize and proves completion without deadlock or post-reauthorization stale publication.
- Corrective audit: account rows now present inclusion readiness from `selectable` plus the specific reason rather than the contradictory generic eligibility flag. EN/FR committed and draft summaries and confirmation use localized `N of M connected accounts` coverage while retaining masked names.
- Corrective audit: removal confirmation now names immediate irreversible deletion of affected source/normalized financial data, derived outputs/signals/matching/ranking, caches and rendered/prefetched state, disclosure-controlled previews, and Discovery outputs; it states that failed additions do not restore removed data without implying provider authorization revocation.
- Review patch: migration `000006` now backfills stable account identities and conservative ready/provisional/unavailable reasons for existing normalized inventory, including the approved provisional status/category path. Partial account normalization makes the whole affected connection consistently unavailable/unselectable, and provider normalization rejects text lengths or decimal exponents that PostgreSQL cannot safely persist. Non-empty typed-row tests read exact balance, position, and activity values back through all three dataset heads.

## Spec Change Log

## Review Triage Log

| ID | Verdict | Evidence and route |
|---|---|---|
| B1 | false | Inventory omission or a disabled connection is not itself a confirmed permission revocation; architecture explicitly retains stale/unavailable cached state for disabled connections. Automatic purge here would contradict that rule. |
| B2 | maybe-false | The UI would show a raw opaque ID and no removal row if committed coverage were absent from current inventory, but reachability depends on reauthorization/provider lifecycle behavior not established by this diff. Defer pending an account-disappearance lifecycle decision. |
| B3 | high | Each account receives a new 10-second provider budget while the HTTP write timeout is 15 seconds. Explicitly deferred by the human for delivery speed; multi-account request budgeting is accepted follow-up risk. |
| B4 | maybe-false | Timestamped successful empty datasets pass, but the pinned contract does not establish whether that response conclusively proves an investment-capable account. Defer until the provider-proof rule defines whether a non-empty position is required. |
| B5 | medium | A lost successful POST response causes the browser to generate a new key and submit the stale version, but reload recovers authoritative state. Deferred under the human-directed narrow review pass. |
| B6 | medium | Provider `RetryAt` is discarded, so a rate-limited inclusion can be retried immediately. Defer to Story 1.3's centralized rate-limit/backoff work rather than expand this story's persistence/API surface. |
| B7 | false | The request itself uses provider-supported inclusive date bounds, and the pinned activity schema expressly allows nullable `trade_date`; rejecting such rows would discard contract-valid activities. |
| B8 | false | Story 1.1 requires provider-supplied observation proof, which a nonzero `as_of` provides. Fresh/stale age classification belongs to Stories 1.2/1.3 and no inclusion-age threshold is specified here. |
| B9 | medium | `SafeLabel` allows 120 runes while position kind and activity type columns allow 60/80, making contract-valid long values fail in PostgreSQL and leave pending work. Patch normalization to reject over-limit rows categorically before publication. |
| B10 | medium | A provider-controlled very large account number can make the pre-existing `MustCompile` path panic. Defer as a pre-existing inventory hardening defect, recorded outside this story. |
| B11 | medium | Decimal syntax accepts exponents beyond PostgreSQL numeric capacity, producing a database error instead of `unusable_data`. Patch normalization with a conservative PostgreSQL-safe exponent bound. |
| B12 | high | Existing rows receive `selectable=false` and no stable identity backfill, so every upgraded ready inventory is unusable immediately. Patch the unshipped migration to backfill readiness/reasons and account identities. |
| B13 | low | The head foreign key does not encode matching owner/account, but every current writer supplies all three values from the same guarded loop and no bad cross-link is reachable. Reject this defense-in-depth schema expansion. |
| B14 | low | Unique no-op keys can add rows, but the same authenticated owner can grow history through real toggles too; retention is the broader issue and no ordinary UI emits no-op POSTs. Reject the non-local special case. |
| V1 | medium | Pre-verified gap: repository success fixtures use empty rows, so typed persisted values could be dropped or swapped unnoticed. Patch with non-empty rows queried through all three heads. |
| V2 | medium | Pre-verified gap: service tests cover only generic provider unavailability, leaving authorization/rate/malformed mappings unproved. Deferred under the human-directed narrow review pass. |
| V3 | medium | Pre-verified gap: failed-state recovery is tested only on initial GET, not immediately after POST. Deferred under the human-directed narrow review pass. |
| VO1 | false | The OpenAPI request-validator middleware runs before the generated handler and enforces item `minLength: 1`; an empty ID is rejected before `canonicalAccountIDs` can remove anything. |
| VO2 | maybe-false | Same underlying proof question as B4: the diff cannot establish whether a timestamped empty positions response is conclusive investment proof. Defer pending that provider/intent decision. |
| VO3 | low | `holdings_unavailable` sets the sync state but the generic unavailable branch wins, hiding the specific recovery reason. Deferred under the human-directed narrow review pass. |
| E1 | high | Same verified migration defect as B12: an existing ready head is not made selectable and identities are absent. Patch the migration backfill. |
| E2 | medium | After one malformed account marks a connection unavailable, later valid rows can still be appended as selectable and inclusion validation checks only the row. Patch by making the final connection result consistently unavailable/unselectable. |
| E3 | false | Same claim as B1: provider omission is not established as a permission decrease and automatic deletion would conflict with retained stale/unavailable state. |
| E4 | low | A historical key replay after later successful changes can pair its old version with current coverage, but the browser never reuses old keys and immediate duplicate replay is correct. Reject this rare archival-response defect; durable change retention/result snapshots need a broader policy. |
| E5 | high | Same verified aggregate-timeout defect as B3. Explicitly deferred by the human for delivery speed. |
| E6 | medium | Same verified length mismatch as B9. Patch provider normalization before persistence. |
| E7 | medium | Same verified numeric-exponent mismatch as B11. Patch provider normalization before persistence. |
| E8 | low | More than 500 simultaneously selectable brokerage accounts is possible in theory but implausible in ordinary use; resolving global inventory and request limits requires a wider scale policy. Reject for this story. |
| E9 | maybe-false | Same conditional orphan-account path as B2. Defer until provider disappearance and retained masked-identity behavior are decided together. |
| E10 | medium | A failed addition that later becomes unselectable remains checked and disabled, so it cannot be removed from the retry draft. Deferred under the human-directed narrow review pass. |
| E11 | medium | Same lost-response idempotency defect as B5. Deferred under the human-directed narrow review pass. |
| E12 | low | SnapTrade documents both date bounds as inclusive, so `end - 30 days` spans 31 dates. Deferred under the human-directed narrow review pass. |
| E13 | maybe-false | Same provider-proof ambiguity as B4/VO2. Defer until conclusive proof for empty investment datasets is specified. |
| E14 | maybe-false | Disabled connections intentionally skip a provider account call, but planning also calls for retained masked unavailable rows. Defer until the retained-inventory behavior is implemented with the disabled-connection lifecycle rather than guessing in this story. |

## Design Notes

Inventory, inclusion, and portfolio lifecycle generations are separate guards. Pending additions are change records. Store values as PostgreSQL `numeric` with ISO currency in typed rows; provider responses remain transient.

## Verification

**Commands:**
- `cd backend && go generate ./... && test -z "$(gofmt -l .)" && go vet ./... && go test -race ./... && go build ./cmd/findur ./cmd/migrate` -- backend passes with no generation drift.
- `npm --prefix frontend ci && npm --prefix frontend run generate:api && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- frontend passes.
- `docker compose config --quiet && ./scripts/compose-test.sh` -- migration, provider contract, and browser journey pass.
- `git diff --check` -- no whitespace errors.

**Results (2026-09-20):**
- `cd backend && GOCACHE=/tmp/findur-go-cache go generate ./... && test -z "$(gofmt -l .)" && GOCACHE=/tmp/findur-go-cache go vet ./... && GOCACHE=/tmp/findur-go-cache go test -race ./... && GOCACHE=/tmp/findur-go-cache go build ./cmd/findur ./cmd/migrate` -- passed, including Docker-backed PostgreSQL concurrency, three scoped dataset heads, atomic rollback/recovery, removal and reauthorization lifecycle fencing, durable-pending supersession, owner isolation, provider proof, and HTTP tests.
- `cd backend && GOCACHE=/tmp/findur-go-cache go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.0 run` -- passed with `0 issues` (mandatory pinned GolangCI-Lint floor).
- `cd frontend && npm ci && npm run generate:api && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- passed: 46 tests, typecheck/build successful, lint 0 errors (2 pre-existing Fast Refresh warnings); the local build emitted the existing undefined `VITE_BUILD_SHA` warning.
- `docker compose config --quiet` -- passed.
- `COMPOSE_PROJECT_NAME=findur-story11 POSTGRES_HOST_PORT=55432 ./scripts/compose-test.sh` -- passed with `integration contracts passed` against an isolated fresh PostgreSQL volume; migration, allowlisted provider contract, idempotent/stale/foreign/removal inclusion behavior, and browser journey completed. The initial default-project run exposed only a stale local volume that had already recorded the earlier uncommitted `000006` shape; that existing volume was preserved and the isolated test-only stack/volume was removed after the passing run.
- `git diff --check` and focused secret-pattern scan over `backend`, `frontend`, and `test` -- passed.
- Focused follow-up: `cd backend && GOCACHE=/tmp/findur-go-cache go test -race ./internal/platform/postgres -run TestInclusionAndReauthorizationShareOwnerLockOrder -count=1` -- passed; concurrent reauthorization with inclusion prepare/finalize completed without deadlock, and stale publication was fenced.
- Focused follow-up full backend and mandatory pinned GolangCI commands above -- passed; GolangCI reported `0 issues`.
- Focused follow-up frontend command above -- passed: 46 tests, including provisional readiness without “Not eligible,” unavailable-account readiness, localized committed/draft/confirmation `N of M` coverage, and the complete irreversible purge consequence set; typecheck/build passed and lint remained at 0 errors with the same 2 warnings.
- `COMPOSE_PROJECT_NAME=findur-story11-lockfix POSTGRES_HOST_PORT=55433 ./scripts/compose-test.sh` -- passed with `integration contracts passed` against a fresh isolated PostgreSQL volume; the isolated test stack and volume were removed after verification.
- Final post-review backend verification: `cd backend && GOCACHE=/tmp/findur-go-cache go generate ./... && test -z "$(gofmt -l .)" && GOCACHE=/tmp/findur-go-cache go vet ./...`, escalated Docker-backed `GOCACHE=/tmp/findur-go-cache go test -race ./...`, and `GOCACHE=/tmp/findur-go-cache go build ./cmd/findur ./cmd/migrate` -- passed. The first sandboxed race invocation failed only because Docker socket access was denied; the required Docker-enabled rerun passed, including the migration backfill and non-empty typed-row tests.
- Final mandatory lint: `cd backend && GOCACHE=/tmp/findur-go-cache GOLANGCI_LINT_CACHE=/tmp/findur-golangci-cache /tmp/findur-go-bin/golangci-lint run` (pinned v2.13.0 binary) -- passed with `0 issues`. The equivalent `go run ...@v2.13.0` attempt was blocked by sandbox DNS before execution, so the already-installed pinned binary was used.
- Final frontend verification: `cd frontend && npm ci && npm run generate:api && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- passed: 46 tests, typecheck/build successful, lint 0 errors with the same 2 pre-existing Fast Refresh warnings and existing undefined `VITE_BUILD_SHA` local-build warning.
- Final fresh-volume journey: `COMPOSE_PROJECT_NAME=findur-story11-final POSTGRES_HOST_PORT=55434 ./scripts/compose-test.sh` -- passed with `integration contracts passed`; `docker compose config --quiet`, fixture JSON parsing, generated/upstream drift checks, `git diff --check`, and the focused secret-pattern review passed. Matches from the secret scan were only field names and explicit synthetic test tokens such as `access-token` and `synthetic-access-token`.

**I/O & Edge-Case Matrix audit (executed):**
- First visit, usable-only checked/mixed/unchecked Select All, disabled reasons, and preservation of disabled committed coverage: `App.test.tsx` — `keeps inclusion draft separate...` and `applies tri-state Select All only to usable rows...` (included in the 46-test frontend run).
- Successful atomic addition, three dataset heads/metadata, forced partial-row rollback and recovery, missing provider observation proof, failed and pending scoped UI recovery: `inclusion_test.go` — `TestInclusionRepositoryPublishesThreeDatasetsAtomicallyAndCanRecover`, `TestInclusionRepositorySupersedesDurablePendingChangeWithoutBroadening`; `account_data_test.go` — `TestAccountDataRejectsMissingProviderObservationProof`; `App.test.tsx` — `reconstructs failed additions...`, `reloads or safely supersedes...` (all executed in the commands above).
- Immediate removal/purge, mixed failure, removal lifecycle fence, and reauthorization fence: `inclusion_test.go` — `TestInclusionRepositoryIsIdempotentOwnerScopedAndRemovalFirst`, `TestReauthorizationFencesPendingInclusionFinalizer` (executed by the Docker-backed race suite).
- Duplicate replay with one claim/effect, stale version, foreign account safe rejection, and owner isolation: `TestInclusionRepositoryIsIdempotentOwnerScopedAndRemovalFirst` plus the passing isolated Compose integration journey.
- Reauthorization versus inclusion lock ordering: `TestInclusionAndReauthorizationShareOwnerLockOrder` overlaps reauthorization with prepare and finalize under a bounded context and verifies no deadlock and no stale post-reauthorization publication (executed directly and within the full race suite).
- Provisional/unavailable readiness, committed/draft/confirmation denominated coverage, and full removal consequences: `App.test.tsx` — `renders connection state before minimized accounts...`, `keeps inclusion draft separate...`, `reconstructs failed additions...`, and `uses a labelled danger action...` (all executed in the 46-test frontend run).
