---
title: 'Story 1.2: Inspect the Private Portfolio Showcase'
type: 'feature'
created: '2026-09-20'
status: 'draft'
route: 'dispatch'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** After confirming account inclusion, an owner cannot inspect the financial facts Findur retained or tell their source, coverage, currency, age, and limitations. Missing or stale facts could look complete or current.

**Approach:** Add the normal owner-private Portfolio Showcase represented by `key-portfolio-showcase.html`. It reads committed persisted heads for the owner's included accounts and presents accounts, balances, positions, and bounded activities as separate sourced datasets with timestamps, freshness, honest states, and accessible summaries/tables. A successful initial Story 1.1 save navigates here; later edits return here.

## Boundaries & Constraints

**Always:** Scope reads from the Actor; expose only committed included accounts; return `private, no-store`; preserve exact decimal strings and separate ISO currencies; identify observed/retrieved/published context; calculate per-account/dataset freshness; make empty, missing, stale, expired, disabled, syncing, failed, and reauthorization states explicit; support EN/FR and accessible responsive use; provide **Continue setting up your profile** for an incomplete first-time user and one clear **Edit included accounts** path back to the saved profile choice.

**Never:** Fetch provider data on render; expose excluded/foreign accounts, secrets, raw payloads, unmasked IDs, or owner IDs; invent zeroes/totals; convert currency; infer performance; add orders, trading, tax lots, quotes, enrichment, or refresh; hand-edit generated files; duplicate the account chooser alongside Showcase content; reintroduce Story 1.1's intimidating technical copy.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Current | Committed included account with three heads | Separate views show safe source/account labels, coverage, currency, values, and timestamps | No provider call; persisted values only |
| Mixed/null | Multiple currencies or absent optional facts | Keep currencies separate; label unavailable fields | Never coerce null or aggregate currencies |
| Empty | Complete head with no rows | Explicit successful empty state with context | Do not call it failed/unsupported |
| Aged | Age crosses a boundary | Text/timestamp distinguish current, stale-usable, expired | Hide unusable facts; guide without auto-refresh |
| Lifecycle | Disabled, syncing, failed, missing head, or auth required | Explicit state; retain only permitted trustworthy facts | Fail closed; reuse reauthorization recovery |
| Isolation | Foreign/excluded/history rows exist | No account or financial facts leak | Actor-scoped committed heads only |
| Continue | First-time owner reaches Showcase with profile setup incomplete | Offer Continue setting up your profile | Continue to the first unmet Profile prerequisite |
| Edit | Owner activates Edit included accounts | Open the saved chooser as a distinct Portfolio child route/state (prefer `/portfolio/accounts`), not simultaneous Showcase content | Successful edit returns to `/portfolio` |

</frozen-after-approval>

## Code Map

- `backend/internal/portfolio/inclusion.go` -- reuse typed facts; keep Showcase as a separate read model.
- `backend/internal/platform/postgres/{inclusion,inventory}.go`, `backend/db/migrations/000006_account_inclusion.up.sql` -- read committed membership, atomic heads, safe labels, and sync mode without schema/publication changes.
- `backend/api/openapi.yaml`, `backend/internal/platform/httpapi/authorization.go`, `backend/cmd/findur/main.go` -- add, defend, serialize, wire, and generate one persisted Showcase GET.
- `frontend/src/{inventory.ts,pages/PortfolioPage.tsx,i18n.tsx,styles.css}` -- validated fetch plus EN/FR accessible datasets, freshness, tables, overflow, recovery, and reflow.
- Backend/React/integration tests -- reuse existing isolation, precision, HTTP, inclusion, and synthetic journey fixtures.

## Tasks & Acceptance

**Execution:**
- [ ] `backend/internal/portfolio/showcase.go`, `backend/internal/platform/postgres/showcase.go` -- project committed heads, safe metadata, exact nullable values, activity bounds, and freshness without provider access.
- [ ] API/HTTP/wiring files above -- add authenticated no-store contract and regenerate API artifacts.
- [ ] Frontend files above -- render separate localized datasets with text freshness, timestamps, summaries, labelled overflow tables, and recovery.
- [ ] Routing/account-choice integration -- preserve `/onboarding/accounts → /portfolio` after the initial committed save (with `/connect/result` as a replace-only compatibility alias); add the normal-shell Edit included accounts child route/state with saved choices preselected and return to `/portfolio`; provide the first-time Continue setting up your profile action.
- [ ] Backend, React, and integration tests -- prove the matrix, precision, boundaries, isolation/removal, zero read-time provider calls, EN/FR, and accessibility.

**Acceptance Criteria:**
- Given committed included accounts, when Portfolio opens, then each distinct dataset states source, coverage, currency, timestamps, publication context, and freshness.
- Given realtime or delayed/Daily holdings, when age is evaluated, then current means ≤15 minutes or ≤36 hours, stale-usable ends at 72 hours; activities are current through two calendar days and stale-usable through seven.
- Given any non-current state, when rendered, then text names it and a safe action without fabricated facts.
- Given a Showcase request, then only the Actor's committed heads return with `private, no-store` and zero provider calls.
- Given initial account setup completes, then the user arrives at the Showcase; given later editing is requested, then the chooser opens separately with saved choices preselected and returns here after a successful save.
- Given the owner has not completed later profile prerequisites, then the Showcase offers one friendly continuation to the first unmet Profile step without hiding the normal owner Portfolio.
- Given accessible/responsive use, then summaries and labelled bounded overflow keep datasets understandable without color, hover, motion, or layout alone.

## Implementation Notes

## Spec Change Log

## Review Triage Log

## Design Notes

Treat `delayed` as Daily. Prefer `observed_at`; otherwise label/classify `retrieved_at`. `published_at` is context only. Unknown sync mode or missing heads are unavailable. State the 30-day/500-row activity bound. Read all scoped heads consistently; broken heads never permit history exposure or refetch.

## Verification

**Commands:**
- `cd backend && go generate ./... && test -z "$(gofmt -l .)" && go vet ./... && go test -race ./... && go build ./cmd/findur ./cmd/migrate` -- backend passes without generation drift.
- `cd backend && golangci-lint run` -- mandatory lint passes.
- `cd frontend && npm ci && npm run generate:api && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- frontend passes.
- `docker compose config --quiet && ./scripts/compose-test.sh` -- fresh-stack Showcase journey passes.
- `git diff --check` -- no whitespace errors.
