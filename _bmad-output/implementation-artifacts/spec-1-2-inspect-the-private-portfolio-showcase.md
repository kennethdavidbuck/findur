---
title: 'Story 1.2: Inspect the Private Portfolio Showcase'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
baseline_commit: '0ab133067bf65be12311d2158b4d2d59793d0067'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** After confirming account inclusion, an owner cannot inspect the financial facts Findur retained or tell their source, coverage, currency, age, and limitations. Missing or stale facts could look complete or current.

**Approach:** Add the normal owner-private Portfolio Showcase represented by `key-portfolio-showcase.html`. It reads committed persisted heads for the owner's included accounts and presents accounts, balances, positions, and bounded activities as separate sourced datasets with timestamps, freshness, honest states, and accessible summaries/tables. A successful initial Story 1.1 save navigates here; later edits return here.

## Boundaries & Constraints

**Always:** Scope reads from the Actor; expose only committed included accounts; return `private, no-store`; preserve exact decimal strings and separate ISO currencies; identify observed/retrieved/published context; calculate per-account/dataset freshness; make empty, missing, stale, expired, disabled, syncing, failed, and reauthorization states explicit; support EN/FR and accessible responsive use; make the account-choice confirmation plainly say that saving opens the Portfolio Showcase; provide one clear **Edit included accounts** path back to the saved account choice.

**Never:** Fetch provider data on render; expose excluded/foreign accounts, secrets, raw payloads, unmasked IDs, or owner IDs; invent zeroes/totals; convert currency; infer performance; add orders, trading, tax lots, quotes, enrichment, or refresh; hand-edit generated files; duplicate the account chooser alongside Showcase content; reintroduce Story 1.1's intimidating technical copy; add or alter the separate Profile flow or its connection to Portfolio.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Current | Committed included account with three heads | Separate views show safe source/account labels, coverage, currency, values, and timestamps | No provider call; persisted values only |
| Mixed/null | Multiple currencies or absent optional facts | Keep currencies separate; label unavailable fields | Never coerce null or aggregate currencies |
| Empty | Complete head with no rows | Explicit successful empty state with context | Do not call it failed/unsupported |
| Aged | Age crosses a boundary | Text/timestamp distinguish current, stale-usable, expired | Hide unusable facts; guide without auto-refresh |
| Lifecycle | Disabled, syncing, failed, missing head, or auth required | Explicit state; retain only permitted trustworthy facts | Fail closed; reuse reauthorization recovery |
| Isolation | Foreign/excluded/history rows exist | No account or financial facts leak | Actor-scoped committed heads only |
| Confirm | Owner reviews selected accounts before an initial committed save | Friendly copy and primary action make clear that saving opens the Portfolio Showcase | Failed/pending save stays on the chooser with existing concise recovery |
| Edit | Owner activates Edit included accounts | Open the saved chooser as a distinct Portfolio child route/state (prefer `/portfolio/accounts`), not simultaneous Showcase content | Successful edit returns to `/portfolio` |

</frozen-after-approval>

## Code Map

- `backend/internal/portfolio/inclusion.go` and `backend/internal/portfolio/inventory.go` -- reuse the Actor boundary and exact nullable decimal value types; keep the Showcase a new read-only service with no provider or token dependency.
- `backend/internal/platform/postgres/inclusion.go` (`loadCommittedAccountIDs`, `publishAccountData`) and `000006_account_inclusion.up.sql` -- read only owner-scoped committed membership and each current dataset head/version/rows; never join historical versions. Reuse safe masked labels, current inventory connection state, and sync mode; no schema or publication change is needed.
- `backend/api/openapi.yaml`, `backend/api/generate.go`, `backend/internal/platform/httpapi/authorization.go`, `backend/cmd/findur/main.go` -- add and generate one authenticated `GET /api/portfolio/showcase`, serialized from the new read model with `Cache-Control: private, no-store`; do not hand-edit generated clients.
- `frontend/src/App.tsx` -- replace the `/portfolio` placeholder, change first successful inclusion save to a replace navigation to `/portfolio`, retain `/connect/result` only as the account-choice compatibility alias, and preserve modifier-safe navigation/title/focus handling.
- `frontend/src/showcase.ts` (new), `frontend/src/pages/PortfolioShowcasePage.tsx` (new), `frontend/src/inventory.ts` -- use the existing no-store, same-origin fetch and runtime-validation pattern; keep account selection in `PortfolioPage.tsx` rather than duplicating it.
- `frontend/src/{i18n.tsx,styles.css,App.test.tsx}` -- add mirrored EN/FR copy, the normal authenticated shell, text-equivalent summaries plus labelled table overflow, focus/recovery behavior, and route/API tests.
- `backend/internal/{portfolio,platform/postgres,platform/httpapi}/*_test.go`, `frontend/src/App.test.tsx`, `test/fixtures/wiremock` and integration coverage -- extend owner isolation, head-only precision, no-provider-read, HTTP privacy, freshness, empty/unavailable, routing, locale, and accessible responsive evidence without exposing IDs or raw payloads.

## Tasks & Acceptance

**Execution:**
- [x] `backend/internal/portfolio/showcase.go`, `backend/internal/platform/postgres/showcase.go` -- define and project the Actor-scoped committed-head read model: safe labels/source/coverage/sync context, exact nullable decimal strings, bounded activity context, timestamps, and per-account/dataset freshness. Missing/broken heads and unknown sync mode fail closed as unavailable; no service dependency may reach the provider.
- [x] `backend/api/openapi.yaml`, `backend/internal/platform/httpapi/authorization.go`, `backend/cmd/findur/main.go` -- publish and wire the authenticated no-store Showcase GET, using the established session defenses and safe 401/error serialization; regenerate Go and TypeScript artifacts.
- [x] `frontend/src/{showcase.ts,pages/PortfolioShowcasePage.tsx,App.tsx,inventory.ts,i18n.tsx,styles.css}` -- fetch and render distinct accounts, balances, positions, and activities datasets in the normal shell. Use source/coverage/publication context, text-plus-shape freshness and timestamps, friendly safe recovery, visible non-table summaries before labelled bounded-overflow tables, locale-aware EN/FR text, exact values/currencies without totals or conversion, and reflow-safe mono evidence styling. Make initial account-save confirmation clearly name the Portfolio Showcase destination.
- [x] `frontend/src/{App.tsx,pages/PortfolioPage.tsx}` -- make the initial committed save replace to `/portfolio`; retain `/portfolio/accounts` as the separately rendered saved-choice editor and return it to Showcase after save. Do not render chooser and Showcase together or restore the intimidating technical copy.
- [x] Backend, frontend, and integration tests -- prove the matrix, exact precision/nulls, actor/excluded/removal/history isolation, zero Showcase provider calls, no-store behavior, freshness thresholds, EN/FR, route/focus behavior, summaries/overflow, and accessible recovery.

**Acceptance Criteria:**
- Given committed included accounts, when Portfolio opens, then each distinct dataset states source, coverage, currency, timestamps, publication context, and freshness.
- Given realtime or delayed/Daily holdings, when age is evaluated, then current means ≤15 minutes or ≤36 hours, stale-usable ends at 72 hours; activities are current through two calendar days and stale-usable through seven.
- Given any non-current state, when rendered, then text names it and a safe action without fabricated facts.
- Given a Showcase request, then only the Actor's committed heads return with `private, no-store` and zero provider calls.
- Given initial account setup completes, then the user arrives at the Showcase; given later editing is requested, then the chooser opens separately with saved choices preselected and returns here after a successful save.
- Given accessible/responsive use, then summaries and labelled bounded overflow keep datasets understandable without color, hover, motion, or layout alone.

## Implementation Notes

- Added an owner-private, persisted-head Showcase API and responsive EN/FR evidence view; no provider dependency exists on the read path.
- Kept account summaries in normal document flow and confined horizontal scrolling to labelled evidence tables at narrow widths.
- Added real PostgreSQL isolation/head/null/precision coverage plus Docker-stack Selenium coverage for initial save, responsive layouts, locale, edit-return behavior, no-store, and zero provider reads on render/reload.
- Limited the initial account-selection progress rail to Connect, Choose accounts, and Portfolio Showcase; the saved-choice editor remains a separate route without that onboarding rail.
- Stopped the progress connector at the final Portfolio step and limited authenticated MVP navigation to Portfolio and Profile.
- Made the supplied synthetic Individual, IRA, and Cash Account payloads the default local WireMock inventory and exercised all three accounts end to end in the Docker/Selenium journey.

## Spec Change Log

## Review Triage Log

The human requested a narrow close-out and accepted the current behavior on 2026-09-21. Findings below were verified and recorded but not expanded into this story.

| Layer | Finding | Verdict and evidence | Route |
|---|---|---|---|
| verification-gap | Lifecycle suppression lacks separate repository cases for every predicate. | medium — pending is covered, but disabled/unavailable/unknown variants are not. | deferred by human direction |
| verification-gap | No populated stale-usable UI case. | medium — the stale fixture is empty, so row visibility is not regression-tested. | deferred by human direction |
| verification-gap | Showcase retry is not exercised. | medium — the 503 state is tested, but the retry click and second request are not. | deferred by human direction |
| verification-gap | A committed account missing from the current inventory head disappears. | medium — inner joins remove the account instead of exposing an unavailable row. | deferred by human direction |
| blind-hunter | Current-inventory inner joins can make a committed account vanish. | medium — confirmed by the account query's inner joins. | deferred by human direction |
| blind-hunter | Inventory `current_status` is not consulted. | high — a retained prior head can remain readable during a later failed or pending inventory state. | accepted by human for this close-out |
| blind-hunter | Distinct lifecycle failures collapse to `unavailable`. | medium — the response cannot distinguish wait, edit, and reconnect guidance. | deferred by human direction |
| blind-hunter | Expired rows remain in the API response. | high — React hides them, but repository datasets still serialize the rows. | accepted by human for this close-out |
| blind-hunter | Future timestamps classify as current. | medium — negative age is clamped to zero without a skew bound. | deferred by human direction |
| blind-hunter | Stale balances appear in the summary without a nearby stale label. | medium — freshness is shown only in the detailed ledger. | deferred by human direction |
| blind-hunter | Stale-usable datasets lack a dataset-level action. | medium — actions are rendered only for expired and unavailable datasets. | deferred by human direction |
| blind-hunter | Activity price and units are omitted from the table. | medium — both values are retained in the read model but absent from the activity columns. | deferred by human direction |
| blind-hunter | Some French presentation remains English or transport-shaped. | medium — the Showcase eyebrow, raw sync modes, and backend coverage copy are not localized. | deferred by human direction |
| blind-hunter | Date-time strings are not runtime-validated for parseability. | medium — arbitrary strings pass and can render as `Invalid Date`. | deferred by human direction |
| blind-hunter | ISO currency codes are not schema-constrained. | low — currency remains separate, but uppercase three-letter form is not enforced at this boundary. | rejected for narrow close-out |
| blind-hunter | Generic dataset validation permits mixed row collections. | medium — a malformed balances dataset can carry positions and still appear empty. | deferred by human direction |
| blind-hunter | A late 401 can redirect after Showcase unmount. | medium — the session-expired callback is outside the `alive` guard. | deferred by human direction |
| blind-hunter | Showcase 503 uses the generic `no-store` error contract. | low — it lacks the endpoint's declared `private` directive, while still prohibiting storage. | rejected for narrow close-out |
| edge-case-hunter | Missing current inventory metadata silently removes a committed account. | medium — independently confirms the inner-join lifecycle gap. | deferred by human direction |
| edge-case-hunter | Activity price and units are not rendered. | medium — independently confirms retained evidence is omitted from the table. | deferred by human direction |
| edge-case-hunter | Every currency amount is decorated with a dollar sign. | high — non-USD values are visibly mislabelled despite retaining their ISO code. | accepted by human for this close-out |
| edge-case-hunter | Stale-usable datasets have no recovery action. | medium — independently confirms the non-current action gap. | deferred by human direction |

## Design Notes

Treat `delayed` as Daily. Prefer `observed_at`; otherwise label/classify `retrieved_at`. `published_at` is context only. Unknown sync mode or missing heads are unavailable. State the 30-day/500-row activity bound. Read all scoped heads consistently; broken heads never permit history exposure or refetch. The presentation remains a calm evidence view: no performance/wealth framing, color-only freshness, hover-only facts, layout-shifting interaction, or generic placeholder treatment.

## Verification

**Commands:**
- `cd backend && go generate ./... && test -z "$(gofmt -l .)" && go vet ./... && go test -race ./... && go build ./cmd/findur ./cmd/migrate` -- backend passes without generation drift.
- `cd backend && golangci-lint run` -- mandatory lint passes.
- `cd frontend && npm ci && npm run generate:api && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- frontend passes.
- `docker compose config --quiet && ./scripts/compose-test.sh` -- fresh-stack Showcase journey passes.
- `git diff --check` -- no whitespace errors.
