---
title: 'Bound account selection and schedule portfolio synchronization'
type: 'feature'
created: '2026-09-21'
status: 'ready-for-review'
route: 'dispatch'
baseline_commit: '829746185ad64d04d387b960901dd25b9d540d38'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md'
  - '_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Account inclusion currently deletes datasets on exclusion, refetches them on re-inclusion, permits thousands of selections, and has no durable refresh loop. Activities are replaced as bounded snapshots, so later synchronization would discard previously discovered transactions.

**Approach:** Limit inclusion to five accounts with an explicit bilingual `N of 5 selected` explanation and reached-limit announcement; fetch account inventory only on its first-ever bootstrap (ordinary later logins reuse the persisted inventory); save selection membership without provider calls so selected accounts appear immediately; show first-time accounts as syncing while the worker prepares their data; retain data and initialization history when excluded; make re-inclusion membership-only; and run a database-leased worker every minute that refreshes included accounts whose complete dataset bundle is at least one day old and prioritizes accounts waiting for their first successful synchronization. Serialize provider dataset calls with at least one second between calls and bound each worker pass to 45 seconds so overdue work drains without bursts or overlapping unbounded runs. Checkpoint balances, positions, and activities independently so a partial failure retries only the incomplete provider resource. Publish balances and positions as replacement snapshots, append/deduplicate the latest 50 provider activities, and return only the newest 50 activities. These decisions supersede Story 1.1's exclusion purge and Story 1.3's activity-triggered-only refresh requirements.

## Boundaries & Constraints

**Always:** Derive ownership from the authenticated actor; validate at most five unique target accounts on both API and domain paths; state the five-account limit before interaction, display `N of 5 selected`, prevent a sixth selection, announce the reached limit accessibly in English and French, and keep deselection available; persist valid membership immediately and render selected account identities while initial datasets sync; preserve excluded data without displaying or routinely refreshing it; retain last-good heads on scheduled failure; claim work durably so multiple processes cannot duplicate a refresh; perform network calls outside transactions; pace provider dataset requests at least one second apart; cap a worker pass at 45 seconds; guard publication against user, authorization, inventory, current selectability, and inclusion lifecycle changes; prioritize unfinished first synchronization; durably checkpoint each successfully published resource and retry only incomplete resources; append activities by stable provider ID; hard-delete through existing disconnect/user cascade behavior.

**Never:** Refetch account inventory merely because an existing user logs in again; refresh portfolio data on re-inclusion; refresh excluded accounts merely because cached data is stale; delete data on exclusion; merge old balances or positions into the current snapshot; expose more than 50 activities; paginate activities beyond the requested first 50; allow scheduler failures to remove membership or overwrite last-good data; log credentials or financial payloads.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First inclusion | Selectable account with no complete heads | Save membership immediately with zero provider calls; navigate to Portfolio, render the account with unavailable datasets, and show a bilingual syncing notice while the worker runs | Successful resources remain checkpointed; the notice remains until every newly selected account completes all three resources, and retryable failures follow bounded backoff |
| Re-inclusion | Previously synchronized excluded account | Restore membership with zero provider calls, regardless of age | Worker refreshes it later when due |
| Exclusion | Included account removed | Membership disappears; datasets and first-sync history remain | In-flight guarded work cannot restore membership |
| Scheduled due work | Included currently selectable account with any missing head or a complete bundle at least 24 hours old | Leased resource calls publish replacement balances/positions and appended activities, checkpointing each success | Preserve last-good heads; retry only the incomplete resource with bounded backoff |
| Selection overflow | Five selected, or more than five unique target IDs sent directly | UI explains and announces the reached limit while permitting deselection; API performs no mutation or provider calls | Return bounded invalid-selection response and never silently disable without an explanation |
| Returning login | Existing owner with a published inventory signs in again | Rotate authorization/session material and return the stored inventory without a provider inventory call | Fence stale in-flight work without clearing the inventory head |
| Backlog / slow provider | More due work exists than safely fits in one minute | Calls remain serialized and at least one second apart; the pass stops by 45 seconds and later ticks drain the backlog | Leases expire safely and no overlapping pass duplicates an account |

</frozen-after-approval>

## Code Map

- `backend/db/migrations/000009_*` -- evolve inclusion-coupled dataset uniqueness; add owner/account refresh claims, attempts, completion, retry state, and append-only normalized activities while migrating current activity heads safely.
- `backend/internal/portfolio/{inclusion.go,sync.go}` -- save bounded membership without provider calls, distinguish first inclusion from re-inclusion, model due/claim/finalize work, and keep provider errors/backoff bounded.
- `backend/internal/platform/postgres/{inclusion.go,sync.go,showcase.go}` -- preserve rows on exclusion, reuse complete heads, lease due work with guarded finalization, publish repeat snapshots, upsert activities, and read newest 50.
- `backend/internal/platform/provider/account_data.go` -- request offset zero/limit 50 activities without the old 30-day window; retain strict typed normalization.
- `backend/cmd/findur/main.go` -- start one cancellation-aware minute scheduler per process; database claims provide cross-process exclusivity.
- `backend/internal/platform/postgres/oauth_attempts.go` -- preserve a published inventory across ordinary returning logins while still fencing stale provider work.
- `backend/api/openapi.yaml`, generated clients, and HTTP tests -- lower inclusion `maxItems` to five and preserve safe validation responses.
- `frontend/src/pages/{PortfolioPage.tsx,PortfolioShowcasePage.tsx}`, `i18n.tsx`, and tests -- enforce and explain five choices, navigate after a pending save, render selected accounts immediately, and retain accessible EN/FR syncing feedback.
- `test/integration/**` -- adapt the large chooser to five selections and assert bounded provider calls, retention, first/re-inclusion, refresh, and latest-50 activity behavior.

## Tasks & Acceptance

**Execution:**
- [x] Add migration and PostgreSQL repository primitives for retained datasets, repeat refresh versions, claims/backoff, guarded publication, and accumulated activities.
- [x] Refactor inclusion/provider/domain logic for fast membership saves, worker-owned initial sync, re-inclusion reuse, five-account validation, and latest-50 activities.
- [x] Add and wire the cancellation-aware minute worker with deterministic unit tests and multi-process-safe repository tests.
- [x] Update Showcase, API/generated code, frontend selection UX, policy documentation, and integration fixtures/tests.

**Acceptance Criteria:**
- Given a never-included account, when it is first included, then membership commits without provider calls, the account appears immediately with syncing feedback, and the worker checkpoints each resource until the initial bundle completes.
- Given a previously synchronized account, when it is excluded and re-included, then its data survives and re-inclusion makes zero provider calls.
- Given an included selectable dataset bundle at least 24 hours old, when workers tick concurrently, then exactly one claim refreshes it and stale results cannot publish.
- Given repeated activity refreshes, when IDs overlap, then unique activities accumulate and Showcase returns the newest 50 only.
- Given five selected accounts, when a sixth is attempted through UI or API, then it is rejected before mutation or provider work.
- Given a user with a published account inventory, when they log in again, then the persisted inventory remains current and no inventory provider request is triggered.

## Implementation Notes

- The composition root creates one purpose-limited SnapTrade client and injects
  it into inventory bootstrap and the worker-owned initial/scheduled refresh path.
  All provider requests therefore share HTTP response limits, error
  categorization, a process-level admission gate, and circuit state.
- Provider calls are serialized at least one second apart. HTTP 429 responses
  open the circuit until the valid `Retry-After` value (delta seconds or HTTP
  date), falling back conservatively when absent or invalid. Repeated transport
  or 5xx failures open a shorter circuit; ordinary 4xx responses do not.
- Each worker process runs an immediate pass and then ticks once per minute. A
  process never overlaps its own passes, and a singleton PostgreSQL pass lease
  permits only one process to drain work globally. Missed ticks coalesce rather
  than queue. A pass admits no more work after 45 seconds. Work is processed one
  account at a time so a backlog drains over later ticks instead of producing a
  burst.
- A PostgreSQL `FOR UPDATE SKIP LOCKED` claim writes a unique claim ID and
  one-minute lease before the transaction commits. Provider network calls occur
  outside the transaction. Publication requires the same live claim plus
  unchanged owner, membership, inclusion version/lifecycle, inventory
  generation, and current selectability. Late or stale results are discarded.
- A scheduled failure preserves membership and all last-good heads, releases the
  lease, and stores a durable retry time using bounded exponential delays of 1,
  2, 4, 8, 16, then 32 minutes. Claim/database failure stops only the current
  pass; the next minute tick tries again. Expired leases recover crashed work.
- Each balance, position, and activity publication commits with its own durable
  checkpoint. After a partial bundle failure, the next claim resumes at the
  incomplete resource rather than repeating provider calls already known to
  have succeeded. The account completion time advances only after all three
  resource checkpoints have advanced for the current refresh cycle.
- Structured logs cover worker/pass start and finish, claim acquisition,
  success, guarded discard, scheduled retry, claim/finalization error, circuit
  rejection category, and shutdown/drain outcome. Logs may contain claim IDs,
  short one-way account references, categorical outcomes, counts, and elapsed
  time; they never contain credentials, labels, balances, positions, activities,
  or provider payloads.
- Graceful shutdown first stops admission and fails readiness, drains HTTP,
  allows an in-flight worker finalization to finish within the shared shutdown
  deadline, waits for the worker goroutine, and only then closes PostgreSQL. A
  deadline overrun is logged and the database is closed deterministically.
- Account inventory is fetched only for the first successful bootstrap. A
  returning OAuth login rotates authorization/session material and increments
  the inclusion lifecycle to fence stale work, but preserves the inventory head
  and versions. Explicit retry/reconnect is the only inventory refresh path.

## Spec Change Log

- 2026-09-22: Human renegotiation replaced inline first-inclusion calls with a
  fast membership save. Selected accounts now appear immediately with syncing
  feedback; the existing minute worker owns initial dataset preparation.

## Review Triage Log

## Design Notes

Selection membership and dataset readiness are separate states. The save transaction makes valid membership visible immediately; a pending inclusion change drives the Portfolio syncing notice until every newly selected account has completed balances, positions, and activities. Refresh age is the completed account refresh cycle: if any head is missing, initial work is due; otherwise the next cycle starts after 24 hours. Each successful resource publication and checkpoint is atomic; the account cycle completes only when all three resource checkpoints have advanced. Retryable initial failures remain worker-visible; scheduled failures retain prior heads and resume at the incomplete resource. Stable activity IDs provide idempotent append semantics, while response ordering uses trade date descending with deterministic ID tie-breaking.

## Verification

**Commands:**
- `cd backend && GOCACHE=/tmp/findur-go-cache go test -race ./...` -- passed; all unit, concurrency, migration, repository, worker, and provider tests pass.
- `cd backend && PATH=/tmp/findur-tools:$PATH GOCACHE=/tmp/findur-go-cache GOLANGCI_LINT_CACHE=/tmp/findur-golangci-cache golangci-lint run` -- passed with `0 issues.` using pinned golangci-lint v2.13.0.
- `cd frontend && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- passed; 74 tests, typecheck, lint with two pre-existing Fast Refresh warnings and no errors, and production build.
- `docker compose -p findur-e2e -f compose.yaml -f /tmp/findur-e2e-override.yaml --profile test run --rm integration` -- passed against a fresh isolated PostgreSQL volume and alternate host ports; all browser and integration contracts passed.
- `node --check test/integration/browser-oauth.mjs && node --check test/integration/browser-session.mjs && node --check test/integration/browser-large-inventory.mjs && git diff --check 829746185ad64d04d387b960901dd25b9d540d38 && test -z "$(gofmt -l backend)"` -- passed; integration script syntax, branch diff, and Go formatting are valid.
