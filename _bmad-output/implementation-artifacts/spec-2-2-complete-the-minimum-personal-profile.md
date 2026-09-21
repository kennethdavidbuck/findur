---
title: 'Story 2.2: Complete the Minimum Personal Profile'
type: 'feature'
created: '2026-09-21'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '0ab133067bf65be12311d2158b4d2d59793d0067'
context:
  - '_bmad-output/implementation-artifacts/epic-2-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The authenticated Profile destination is a placeholder, so the owner cannot add the minimum human context required for a credible adult dating profile or later Discovery readiness.

**Approach:** Make `/profile` a reusable create/edit form backed by an owner-private profile API and atomic persistence. Keep account selection ending at the Portfolio Showcase; a later integration can add a normal `/profile` link without onboarding-step state or changes to the Showcase in this story.

## Boundaries & Constraints

**Always:** Treat the finalized UX experience and design spines as authoritative; include display name, adult attestation, coarse location, relationship intent, short biography, bundled avatar, locale, and theme; make first save atomic and edits version-guarded; retain drafts through validation, network failure, and conflict; expose persistent required/private/visibility annotations, help, associated errors, an error summary, deterministic focus, 44px targets, EN/FR parity, light/dark parity, and responsive reflow; keep the desktop rail and mobile bottom App Navigation available while long Profile content scrolls without obscuring content or focus at zoom or safe-area sizes; scope all reads and writes to the authenticated Actor with `private, no-store`.

**Never:** Modify the Portfolio Showcase or its post-account-save route; persist onboarding progress; accept date of birth, precise/device/IP-derived location, arbitrary avatar URLs, uploads, binaries, or partial first profiles; hand-edit generated API artifacts; use wealth, worth, trading-urgency, luxury, or deterministic compatibility language.

**Product decisions:** Location selection is limited to a compact, migration-seeded set of representative Canadian cities with localized city/province labels; catalogue coordinates remain server-side and exist only for later approximate-distance calculations. Relationship intent is one of `long-term`, `open-to-long-term`, or `figuring-it-out`. The bundled avatar set uses inclusive illustrated human portraits. Display name, relationship intent, and biography are pre-match fields; the avatar is post-match; location, adult attestation, locale, and theme remain private.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| Empty | Authenticated owner has no profile | `/profile` renders a blank localized form with current browser locale/theme | Protected content never flashes before the session resolves |
| Create | Every required value is valid with expected version `0` | One complete profile and adult-attestation timestamp are created atomically | Invalid input persists nothing and focuses a linked error summary |
| Edit | Existing profile and matching version | The complete replacement saves and advances its version | Stale version returns conflict; preserve the draft and offer reload/reapply guidance |
| Catalogue boundary | Submitted location, intent, avatar, locale, or theme is outside its allowlist | Reject the complete mutation | Associate the safe error without echoing unsafe input into options or assets |
| Failure | Load/save is unavailable or session expires | Preserve non-sensitive draft state for retry, or return to secure connection | Never present an unsuccessful draft as saved |

</frozen-after-approval>

## Code Map

- `backend/db/migrations/000008_personal_profile.{up,down}.sql` -- add the seeded location catalogue and zero-or-one complete owner profile with constraints and version.
- `backend/internal/profile/profile.go` -- new domain types, allowlists, validation, create/edit service, and conflict semantics; depend only on authenticated `auth.Actor`.
- `backend/internal/platform/postgres/profile.go` -- actor-scoped reads and atomic insert/version-guarded replacement; return localized location keys, never centroids.
- `backend/api/openapi.yaml` -- add owner-private GET and PUT `/api/profile`, explicit version/conflict/validation schemas, CSRF header, and no-store responses; regenerate Go/TypeScript.
- `backend/internal/platform/httpapi/{authorization.go,httpapi.go}` and `backend/cmd/findur/main.go` -- compose the profile service into existing session/origin/CSRF defenses and safe route categorization.
- `frontend/src/profile.ts` and `frontend/src/pages/ProfilePage.tsx` -- typed fetch adapter plus reading-width React Aria form with create/edit, summary-linked field errors, preserved drafts, save status, and conflict recovery.
- `frontend/src/{App.tsx,i18n.tsx,theme.tsx,styles.css}` -- replace only the `/profile` placeholder, synchronize saved locale/theme with existing providers, and implement Constellation field/action/focus/reflow rules; do not alter `PortfolioPage.tsx`.
- Backend repository/domain/HTTP tests and `frontend/src/App.test.tsx` -- cover owner isolation, atomicity, allowlists, conflict, CSRF, direct-route focus, localization, responsive semantics, and draft preservation.

## Tasks & Acceptance

**Execution:**
- [x] Migration and `internal/profile` files -- establish constrained catalogue/profile persistence and tested create/edit rules.
- [x] OpenAPI, HTTP adapter, and composition files -- expose authenticated no-store GET/PUT and regenerate artifacts.
- [x] Profile frontend, catalogues, theme integration, and styles -- deliver the complete accessible `/profile` experience without Showcase coupling.
- [x] Domain, PostgreSQL, HTTP, and React tests -- prove every matrix row and privacy/accessibility boundary.

**Acceptance Criteria:**
- Given an authenticated owner opens `/profile`, when session and profile loading settle, then the route heading is focused and the complete localized form exposes each field's requirement and visibility without protected-content flash.
- Given a valid first save or version-matched edit, when persistence succeeds, then the complete Actor-scoped profile saves atomically, its version advances, locale/theme stay synchronized, and success is announced without moving focus.
- Given invalid, unavailable, expired-session, or conflicting input, when save fails, then no partial or false-saved state appears and entered values remain recoverable with deterministic field or recovery focus.
- Given unsupported location or avatar input, when the API validates it, then coordinates and arbitrary media never enter the contract or storage.
- Given EN/FR, System/Light/Dark, phone/desktop, forced colors, keyboard, screen reader, touch, or zoom/reflow use, when the form is completed, then equivalent meaning and actions remain available with no clipping or geometry-shifting interaction state.
- Given Profile content exceeds the viewport, when the owner scrolls on phone or desktop, then the appropriate App Navigation remains available and does not cover form content, actions, or visible focus.

## Implementation Notes

- Added a migration-seeded six-city Canadian catalogue whose centroids remain database-only, plus a constrained one-row-per-owner profile with atomic create and version-guarded replacement.
- Added owner-private OpenAPI GET/PUT operations and generated Go/TypeScript bindings; PUT reuses the established session, same-origin, Fetch Metadata, and CSRF defenses.
- Replaced the `/profile` placeholder with an EN/FR form whose validation summary links to fields, recoverable failures preserve the draft, conflicts provide explicit reload/reapply guidance, and successful locale/theme values synchronize with the shared providers.
- Kept authenticated navigation persistent around long content through the shared shell: a scrollable fixed desktop rail and safe-area-aware mobile bottom navigation with reserved content space.

## Spec Change Log

- 2026-09-21: Implemented the minimum personal profile vertical slice and persistent authenticated navigation behavior.

## Review Triage Log

| ID | Verdict / route | Evidence |
| --- | --- | --- |
| E1 | medium / patch | Editing while PUT is pending is reachable because only Save is disabled; the later success state can describe a newer unsaved draft as saved. |
| E2 | medium / patch | The service trims name/biography but success only copies version/preferences, leaving visibly different text beside “Profile saved.” |
| E3 | medium / patch | `navigator.languages.some(fr)` ignores preference order, so `en-CA, fr-CA` incorrectly selects French. |
| E4 | medium / patch | Browser `maxLength` and `.length` count UTF-16 units while the domain counts Unicode code points, rejecting valid astral input. |
| E5 | medium / patch | NUL and display-empty control/format input pass domain validation and can fail PostgreSQL or create an unusable visible identity. |
| E6 | low / rejected | Version precision would require more than nine quadrillion edits; adding a cross-layer maximum for that unreachable normal state is disproportionate. |
| E7 | low / patch | Profile middleware error paths emit `no-store`, not the profile contract’s exact `private, no-store` directive. |
| E8 | low / rejected | The migration creates and seeds the catalogue atomically; an empty successful catalogue requires out-of-band database corruption and a new guard is not warranted here. |
| V1 | medium / patch | Pre-verified: compose/Selenium never visits `/profile`, so production wiring can disappear while every existing command passes. |
| V2 | medium / patch | Pre-verified: profile service tests use a zero Actor and discard repository owner arguments, so Actor identity forwarding is untested. |
| V3 | medium / patch | Pre-verified: the HTTP happy path asserts only status/version and never verifies request/response field mapping. |
| V4 | medium / patch | Pre-verified: the repository edit test changes only display name, leaving complete-replacement mapping unverified. |
| V5 | medium / patch | Pre-verified: no test covers an empty-storage browser whose primary supported locale is French. |
| B1 | false / rejected | No no-default rule exists for minimum-profile relationship intent or bundled portrait; the cited no-default decisions apply to account inclusion and Disclosure Level. |
| B2 | medium / patch | OpenAPI request middleware handles enum/length/const failures before the domain and returns generic `invalid_request`, bypassing linked field errors. |
| B3 | medium / patch | The documented profile PUT 400 body is field-bearing, while schema-invalid requests currently return the generic Error shape. |
| B4 | medium / patch | Confirmed with E2: successful canonical server values are not reapplied to the visible draft. |
| B5 | medium / patch | Confirmed with E4: frontend and backend length semantics diverge for supplementary Unicode characters. |
| B6 | medium / patch | Confirmed with E5: visually blank control/format-only values and PostgreSQL-hostile NUL are accepted by the domain. |
| B7 | false / rejected | The route-exit discard/stay requirement cited by the reviewer is explicitly attached to the future Disclosure Level draft; this story enumerates validation/network/conflict retention. |
| B8 | low / rejected | Shared locale/theme controls intentionally persist explicit browser preferences immediately; a failed profile save does not claim those browser-only choices were saved to the profile. |
| B9 | low / rejected | The six-city catalogue is fixed by this story and protected by migration/domain tests; replacing the bounded domain allowlist with repository-dependent validation is disproportionate future-proofing. |
| B10 | medium / patch | Every server field error currently says “required,” which is false for invalid length, unsafe content, or unsupported catalogue values. |
| B11 | medium / patch | Confirmed with E3/V5: browser locale negotiation ignores supported-language preference order. |
| B12 | maybe-false / rejected | The shell reserves bottom-nav space; no failing focus-scroll reproduction was supplied, so a second scroll-container rule is not justified. |
| B13 | low / patch | Confirmed with E7: profile 415/500/503 paths do not consistently carry `private, no-store`. |
| B14 | medium / patch | Confirmed with V1: the verification claim names a profile browser journey that the current Selenium harness does not execute. |
| B15 | false / rejected | The build workflow deliberately moves the spec to `in-review` before its final sprint-status synchronization step. |

## Design Notes

Use `/profile` itself for the first-cut form. Choose a short biography rather than adding a prompt-picker. Treat completeness as derived from the persisted row; do not store a workflow step. Keep avatar keys stable and asset-owned so later pre-/post-match projection can omit or reveal them server-side.

## Verification

**Commands:**
- `cd backend && go generate ./... && test -z "$(gofmt -l .)" && go vet ./... && go test -race ./... && go build ./cmd/findur ./cmd/migrate` -- backend and generated contract pass without drift.
- `cd backend && golangci-lint run` -- mandatory repository lint passes.
- `cd frontend && npm ci && npm run generate:api && npm test -- --run && npm run typecheck && npm run lint && npm run build` -- frontend contract, behavior, accessibility assertions, types, lint, and production bundle pass.
- `docker compose config --quiet && ./scripts/compose-test.sh` -- migrated PostgreSQL and authenticated profile journey pass in the integration stack.
- `git diff --check` -- no whitespace errors.

**Results:** All commands passed. Frontend lint completed with only the two pre-existing Fast Refresh warnings in `i18n.tsx` and `theme.tsx`; the Docker browser contract also verified the exact 767px/768px App Navigation boundary.
