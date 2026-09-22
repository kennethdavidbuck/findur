---
title: 'Unify authenticated display preferences'
type: 'bugfix'
created: '2026-09-22'
status: 'in-review'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: 'b483f1203f2b9aeb13108b98edf186e18fdd99a8'
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The authenticated header changes theme and locale only in browser storage. A later Profile load applies the stale profile values, visibly undoing the user’s choice; the current full-profile API cannot safely save a preference for an incomplete profile.

**Approach:** Give each authenticated owner a small, independently persisted display-preference record. Hydrate one shared authenticated preference state from it, let header and Profile controls autosave through that state, and retain browser storage only as a best-effort anonymous/startup cache.

## Boundaries & Constraints

**Always:** Preserve the current route and immediately apply a deliberate locale/theme selection. Public preferences remain browser-only; use a shared storage adapter whose in-memory fallback supports reads/writes when `localStorage` is missing, blocked, or throws. The authenticated server record wins over that cache and refreshes it when possible. When no server record exists, seed it once from the current in-memory preference (which may originate from cached or browser/system defaults). Authenticate and protect reads/writes with the same session, same-origin, CSRF, and `private, no-store` guarantees as profile endpoints. Keep the complete personal-profile form and its validation/version lifecycle intact; its display controls must use the common autosave state rather than participate in the form submission.

**Never:** Do not store account preferences or protected API payloads in browser storage; do not require profile completion to save display preferences; do not let stale cache overwrite a saved authenticated choice; do not remove authenticated header controls; do not create a second divergent preference state.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First authenticated use | No server preference; cache/default is `fr`/`dark` | Current choice remains visible and is persisted as the initial owner preference | Storage unavailability does not prevent the authenticated request or default choice |
| Public site without browser storage | `localStorage` is unavailable, denied, or throws | Public locale/theme changes survive in the in-memory storage adapter for the active page session | Browser/system defaults seed the adapter; no error is surfaced or storage is required |
| Returning authenticated use | Cache is `en`/`light`; server is `fr`/`dark` | Server choice becomes shared UI state and refreshes cache if available | Cache is never sent to overwrite server |
| Header or Profile change | Saved profile is complete or incomplete | UI applies instantly, then the common preference service persists the selected value(s) | A failed request reports recoverable status and does not silently claim it saved |
| Session/defense failure | Autosave returns 401/403 | Recover through the existing protected-session path | Do not retain an unsafely assumed saved state |
| Concurrent update | Preference save sees a newer value | Resolve using the latest server response/reload rather than losing a selection silently | Tell the user a retry/reload is needed if automatic resolution cannot preserve it |

</frozen-after-approval>

## Code Map

- `backend/api/openapi.yaml` -- Existing generated contract contains only full profile GET/PUT. Add the protected independent preference contract and no-store response headers, then regenerate the Go and TypeScript clients.
- `backend/db/migrations/000008_personal_profile.up.sql` -- Stores required profile locale/theme and cannot represent an incomplete owner. Add a forward migration and rollback for owner display preferences without changing the existing profile schema/records.
- `backend/internal/profile/profile.go` and `backend/internal/platform/postgres/profile.go` -- Existing domain/repository validation and complete-profile persistence patterns; extend or introduce a narrowly scoped preference service/repository with owner-scoped validation and conflict semantics.
- `backend/internal/platform/httpapi/authorization.go` and `backend/internal/platform/httpapi/httpapi.go` -- Wire, authorize, validate, and no-store the new API beside the protected profile routes.
- `backend/cmd/findur/main.go` -- Compose the preference repository/service into the HTTP handler.
- `frontend/src/profile.ts` -- Reuse request, CSRF, session, defense, and no-store behavior for typed preference calls.
- `frontend/src/i18n.tsx`, `frontend/src/theme.tsx`, `frontend/index.html`, and a small shared storage adapter -- Current guarded cache/default and immediate DOM application; replace direct browser-storage dependence with a safe in-memory fallback for public and authenticated use while allowing an authenticated source to apply state without accidentally triggering divergent persistence.
- `frontend/src/App.tsx` -- Knows authenticated routing and provider composition; install one authenticated preference hydration/autosave boundary here or in a dedicated provider.
- `frontend/src/components/PublicLayout.tsx` and `frontend/src/components/AuthenticatedLayout.tsx` -- `PreferenceControls` currently directly mutates local state; route authenticated usage through the shared autosaving controller while leaving public behavior cache-only.
- `frontend/src/pages/ProfilePage.tsx` -- Stop profile-load/form-save ownership of shared display settings; render the same controller-backed controls and retain the rest of the form draft/version behavior.
- `frontend/src/App.test.tsx`, `frontend/src/pages/ProfilePage.test.tsx`, backend profile/postgres/http API tests -- Extend coverage for hydration, autosave, security, server/cache precedence, failures, and unavailable browser storage.

## Tasks & Acceptance

**Execution:**

- [ ] `backend/api/openapi.yaml`, generated clients, migrations, `internal/profile`, `internal/platform/postgres`, `internal/platform/httpapi`, and `cmd/findur/main.go` -- implement an owner-scoped, validated, protected display-preference read/create/update path independent of personal profile.
- [ ] `frontend/src/profile.ts` and a focused authenticated-preferences provider/module -- add typed API access, one-time authenticated hydration, initial seeding, optimistic shared updates, cache refresh, and an accessible recoverable save state.
- [ ] `frontend/src/App.tsx`, `i18n.tsx`, `theme.tsx`, `index.html`, a safe storage adapter, `PublicLayout.tsx`, `AuthenticatedLayout.tsx`, and `ProfilePage.tsx` -- distinguish public cache-only controls from authenticated autosaving controls; make cache reads/writes transparently fall back to page-lifetime memory; make both authenticated entry points render the same state; and remove display fields from full-profile submission ownership.
- [ ] Relevant frontend and backend tests -- cover every matrix case, API authorization/CSRF/no-store behavior, and safe localStorage failure handling.

**Acceptance Criteria:**

- Given an authenticated owner selects French or Dark in the header, when they navigate or reload, then the selection remains applied from their owner preference rather than reverting to stale local storage or profile data.
- Given a saved owner preference conflicts with a browser cache, when an authenticated shell hydrates, then the saved owner preference wins and the cache is refreshed only when browser storage is usable.
- Given an authenticated owner has no complete profile, when they choose a display preference, then it saves successfully without creating or validating a personal profile.
- Given an authenticated owner changes a setting from Profile, when they change it, then it autosaves through the same state and a later profile form save cannot overwrite it with a stale draft.
- Given browser storage is disabled, when anonymous or authenticated preferences load/change, then the UI remains usable, system/browser defaults still apply, and authenticated choices persist server-side.
- Given public-site browser storage is disabled or absent, when a visitor changes locale or theme, then the choice is retained by the page-lifetime fallback object without errors or a dependency on browser storage.

## Implementation Notes

## Spec Change Log

## Review Triage Log

## Design Notes

Use a single visible preference selection immediately. A compact saving/saved or retry message must not move focus, interrupt the current task, or make users wonder whether the selected screen appearance is temporary.

## Verification

**Commands:**

- `cd frontend && npm test -- --run` -- expected: relevant component/provider tests pass.
- `cd frontend && npm run build` -- expected: generated client and production TypeScript build succeed.
- `cd backend && go test ./...` -- expected: domain, persistence, HTTP, and composition tests pass.
- `cd backend && golangci-lint run` -- expected: mandatory lint passes with no findings.
