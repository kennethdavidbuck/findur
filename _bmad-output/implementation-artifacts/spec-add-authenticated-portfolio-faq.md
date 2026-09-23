---
title: 'Add an authenticated portfolio FAQ'
type: 'feature'
created: '2026-09-23'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '37bbc0ab09e13bffea6a11c15e4af36066fb067a'
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Regular users see arbitrary-looking freshness labels and several nearly identical timestamps in Portfolio, while the FAQ omits the implemented account-eligibility and retry rules needed to understand the page.

**Approach:** Add a simple bilingual FAQ inside the authenticated shell, linked after Profile, and simplify each Portfolio dataset to neutral Last sync / Next sync timing. Explain account eligibility, inclusion, cadence, independent resources, bounded automatic retry, retained data, and the few actions a user may need to take.

## Boundaries & Constraints

**Always:** Keep the page protected by the existing session gate; place FAQ after Profile in authenticated navigation; support English/French, light/dark themes, phone/desktop layouts, keyboard use, route-heading focus, browser history, and 35% copy expansion; use calm nontechnical language while preserving implemented facts. Explain the server-enforced account rules, five-account limit, explicit inclusion, roughly 24-hour cadence, independent resources, newest-50 activity boundary, automatic 1/2/4/8/16/32-minute retry sequence and provider-directed later retry, retained safe values, reconnect boundary, and what “Check again” and “Try again” actually do. Present Published as Last sync and the ordinary scheduled or retry time as Next sync.

**Never:** Change sync scheduling, freshness classification, retry behavior, APIs, persistence, account inclusion, or public navigation; expose Observed/Retrieved/Published as competing user-facing timestamps; show ordinary current/stale/expired color badges; claim continuous/live updates, guaranteed execution at the displayed Next sync time, complete activity history, or that all datasets share one success/failure state; introduce a visually unrelated card system or FAQ-only interaction convention.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| Authenticated entry | User follows FAQ after Profile or opens `/faq` directly | Protected FAQ route renders, receives current locale/theme, focuses its H1, updates the document title, and marks FAQ current | Existing session recovery handles expired/invalid sessions before protected copy mounts |
| Plain-language guidance | User opens questions about eligible accounts, cadence, timestamps, retries, failures, or actions | Native semantic disclosures expose accurate EN/FR answers and remain operable by keyboard/touch | No data request or live state inference is required |
| Portfolio timing | Dataset has a successful publication, a scheduled retry, or no data yet | Show neutral Last sync and Next sync facts; retry time takes precedence, otherwise Next sync is the ordinary 24-hour expectation | Initial data uses an unavailable placeholder; actual diagnostics retain their existing explanation/action styling |
| Responsive navigation | FAQ appears in the phone bottom bar or desktop rail | Link follows Profile, wraps safely in both languages, and retains the shared current/hover/focus/pressed treatments | Content and navigation remain usable at the 767/768px boundary and zoom/reflow sizes |

</frozen-after-approval>

## Code Map

- `frontend/src/App.tsx` -- extend route parsing, protected-route dispatch, title selection, and existing route-heading focus for `/faq`; do not alter authentication behavior.
- `frontend/src/components/AuthenticatedLayout.tsx` -- extend `ProtectedRoute` and place the localized FAQ destination after Profile using the shared navigation behavior.
- `frontend/src/pages/FaqPage.tsx` -- new reading-width bilingual page using native headings and `details`/`summary`; copy must reflect `PortfolioShowcasePage` and the implemented sync policies.
- `frontend/src/pages/PortfolioShowcasePage.tsx` -- replace freshness badges and competing evidence timestamps with neutral Last sync / Next sync facts while preserving safe row suppression, diagnostics, and recovery actions.
- `frontend/src/i18n.tsx` -- add localized authenticated navigation/title keys while keeping page-specific prose local, matching existing page convention.
- `frontend/src/styles.css` -- compose the established authenticated/Profile geometry, tokens, focus states, and responsive breakpoints; adapt the mobile destination grid without inventing a new visual system.
- `frontend/src/App.test.tsx` and `frontend/src/pages/FaqPage.test.tsx` -- cover protected routing, link order/current state, focus/history/title, disclosure semantics, localized factual copy, and absence of page data fetching.
- `test/integration/browser-session.mjs` -- update authenticated navigation assertions and exercise FAQ at the existing phone/desktop breakpoint checks.
- `docs/account-eligibility.md`, prior merged sync/diagnostic specs, and `backend/internal/{portfolio,platform/postgres}` -- source of truth for eligibility, cadence, and bounded automatic retry; do not change backend behavior or contracts.

## Tasks & Acceptance

**Execution:**
- [x] `frontend/src/pages/FaqPage.tsx`, `frontend/src/i18n.tsx`, `frontend/src/styles.css` -- build the concise EN/FR FAQ from shared visual, localization, semantic-disclosure, responsive, eligibility, and recovery conventions.
- [x] `frontend/src/pages/PortfolioShowcasePage.tsx`, `frontend/src/styles.css` -- show neutral Last sync / Next sync timing and remove ordinary freshness badges/colors without weakening diagnostic or expired-data handling.
- [x] `frontend/src/App.tsx`, `frontend/src/components/AuthenticatedLayout.tsx` -- register `/faq` as a protected destination after Profile with correct title, focus, current state, history, and setup-shell exclusion.
- [x] `frontend/src/App.test.tsx`, `frontend/src/pages/FaqPage.test.tsx`, `test/integration/browser-session.mjs` -- verify the matrix, factual eligibility/retry boundaries, schedule presentation, navigation placement, keyboard use, trailing slash, and breakpoint behavior.

**Acceptance Criteria:**
- Given a signed-in regular user on any protected page, when they choose FAQ after Profile, then a native-looking bilingual help page explains which accounts can be selected, what inclusion means, ordinary waiting, the retry sequence, retained data, and explicit action-required states without requiring financial or system expertise.
- Given English or French and any supported theme/viewport, when the FAQ is opened directly or through browser history, then session protection, document metadata, route focus, navigation state, disclosure semantics, contrast/focus treatments, and reflow match the authenticated app.
- Given a dataset with saved data, when Portfolio renders, then Published appears as Last sync, Next sync uses a retry time when present or the ordinary 24-hour expectation otherwise, and no current/stale/expired freshness badge or semantic color competes with those facts.
- Given the implemented sync policies, when the FAQ describes timing or recovery, then it distinguishes expected timing from guaranteed completion, independent resource updates from whole-account state, status reload from provider synchronization, and retained saved evidence from unavailable data.

## Implementation Notes

Added `/faq` to the protected router and existing authenticated navigation after Profile. The new page inherits locale/theme from the authenticated providers, uses six native disclosure sections, and performs no page-specific data fetch. Stable locale-independent item IDs preserve open disclosures when the language changes; each summary contains an H2 and a visible plus/minus state indicator.

The bilingual FAQ now explains the server-enforced account rules, explicit selection and five-account limit, expected 24-hour schedule, independent resources and newest-50 activity boundary, bounded automatic retry sequence, retained safe values, and the exact Check again, Try again, and reconnect boundaries.

Portfolio datasets now present Published as neutral Last sync timing and use diagnostic retry timing, when present, ahead of the ordinary Published-plus-24-hours Next sync estimate. Observed/Retrieved/Published competition and ordinary freshness badges/colors are removed while diagnostic explanations, recovery actions, initial-sync handling, and expired/unavailable row suppression remain intact.

The responsive navigation has three equal phone destinations in one row and retains the existing desktop rail. Focus, keyboard activation, title, history, `/faq` and `/faq/` entry, locale-state preservation, English/French factual coverage, and narrow expanded-content reflow have focused unit or browser-harness coverage.

The FAQ uses the same 64rem authenticated page container as Portfolio and Profile. Its 45rem reading column is left-aligned within that shared container, matching the Profile form rather than centering the entire page farther to the right.

## Spec Change Log

## Review Triage Log

| ID | Verdict | Route | Evidence |
|---|---|---|---|
| B1 | medium | patch | Cadence copy described per-resource eligibility as work starting; the revised user model must describe the ordinary 24-hour expectation without promising immediate execution. |
| B2 | medium | superseded-and-patch | The user removed ordinary freshness states from the presentation; the FAQ will instead explain Last sync / Next sync plus explicit syncing/failure recovery. |
| B3 | low | patch | The expanded FAQ now names the request-level Try again action alongside Check again because the human explicitly requested complete retry guidance. |
| B4 | medium | patch | Grid styling suppressed the native disclosure marker; add a visible open/closed indicator. |
| B5 | low | patch | Locale-dependent keys remount open disclosures; use stable IDs. |
| B6 | low | patch | Questions need headings inside summaries for heading navigation. |
| B7 | medium | defer | Authenticated-preference loading can briefly mount any protected route; this predates the FAQ and requires an app-shell-wide session-loading contract. |
| B8 | medium | patch | Add keyboard focus and disclosure activation coverage. |
| B9 | medium | patch | Add narrow-phone and expanded-copy reflow coverage. |
| E1 | low | patch | Duplicate of B5; stable IDs preserve disclosure state across locale changes. |
| V1 | medium | patch | Assert exactly three equal phone navigation columns in one row. |
| V2 | low | patch | Cover both `/faq` and `/faq/` direct entry and normalization. |

## Design Notes

FAQ is an authenticated utility destination, not a new primary product area. Its placement after Profile answers the requested discovery path while its route, typography, boundaries, and interactive states remain owned by the existing app shell. Native `details` elements provide progressive disclosure without custom accordion state or scripting.

## Verification

**Commands:**
- `cd frontend && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- expected: all frontend tests, types, lint, and production build pass.
- `COMPOSE_PROJECT_NAME=findur-faq COMPOSE_FILE=compose.yaml:/tmp/findur-auth-faq.compose.yaml POSTGRES_HOST_PORT=55433 BUILD_SHA=<unique-40-hex> ./scripts/compose-test.sh` -- expected: isolated browser/integration contracts pass without using the other stack's ports, images, volumes, containers, or network.
- `cd backend && golangci-lint run` -- expected: mandatory repository lint passes unchanged.
- `git diff --check` -- expected: no whitespace errors.

**Results (2026-09-23):**
- Layout follow-up passed 80 focused FAQ/App tests, frontend lint with the same four pre-existing warnings, production build, browser-script syntax, and the clean isolated Compose suite with an explicit 64rem page / 45rem left-aligned reading-column assertion (`integration contracts passed`, image tag `fa90000000000000000000000000000000000003`).
- Review follow-up focused tests passed: `cd frontend && npm test -- --run src/pages/FaqPage.test.tsx src/App.test.tsx` (2 files, 80 tests). `node --check test/integration/browser-session.mjs` and `git diff --check` also passed.
- Frontend tests passed on the final full run: 8 files and 122 tests. Typecheck and production build passed. ESLint completed with zero errors and four pre-existing warnings.
- The first composed rerun encountered retained synthetic state in the preserved task volume before reaching the changed FAQ journey. After removing only the `findur-faq` containers, network, and synthetic database volume, the clean full browser/integration suite passed with `integration contracts passed` under project `findur-faq`, isolated ports 58080/55433/55173, and unique image tag `fa90000000000000000000000000000000000002`. The containers and network were removed afterward; the clean synthetic volume and images were preserved.
- The mandatory exact `cd backend && golangci-lint run` command was executed but could not start because `golangci-lint` is not installed on the shell `PATH` (`command not found`). The cached v2.13.0 binary then ran with writable task-specific caches and reported `0 issues.`
- `git diff --check` passed.
