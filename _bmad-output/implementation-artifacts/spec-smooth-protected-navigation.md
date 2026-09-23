---
title: 'Keep protected navigation visually stable'
type: 'bugfix'
created: '2026-09-23'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/spec-handle-snaptrade-authorization-recovery.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Protected navigation currently keys the entire authenticated application by route, so every link destroys the shell, flashes the session gate, and reloads shared preferences and page state. Production latency makes this client-side remount look like a full browser reload.

**Approach:** Keep the authenticated shell and shared state mounted while rechecking authorization for each protected destination, retain the current protected view until the check succeeds, and transition only after the server authorizes it.

</frozen-after-approval>

## Implementation Notes

- Extended the authorization-status hook with a validation key so route rechecks can complete without resetting the last authorized view.
- Kept `ProtectedApp`, its layout, and authenticated preferences mounted across protected routes; the previous route remains visible until the destination check succeeds.
- Stabilized the router callback so retained page effects do not restart merely because the requested URL changed.
- Added regression coverage for shell identity, absence of the full-page session gate, one-time shared preference loading, and delayed destination rendering.
- Guarded fulfilled requests after abort, blocked superseding link activations during validation, and added a progress cursor plus live status without replacing the shell.

## Review Triage Log

- **medium, patched:** A fulfilled request could update state after its effect was disposed; successful completion now checks the disposal guard, with a late-result regression test.
- **medium, deferred:** A transient authorization-status failure still fails closed and routes the user to Connect. This predates the visual regression and changing it requires a distinct security and retry contract.
- **medium, patched:** Retaining the navigation exposed repeated link activation while a destination was pending; links are now temporarily marked unavailable and ignore activation until validation settles.
- **medium, patched:** Slow validation had no feedback; the stable shell now exposes navigation busy state, a progress cursor, and a polite status announcement.
- **medium, patched in scope:** Tests now cover retained shell identity, pending interaction, superseded late fulfillment, and browser navigation during validation. Transient-failure behavior is tracked with the deferred contract above.

## Verification

- `npm test -- --run` -- passed; 130 tests.
- `npm run typecheck` -- passed.
- `npm run lint` -- passed with 0 errors and 4 pre-existing warnings.
- `npm run build` -- passed; Vite emitted the expected local missing `VITE_BUILD_SHA` warning.
- `cd backend && golangci-lint run` -- passed with `0 issues.` using the CI-pinned v2.13.0 binary and sandbox-local caches.
- `git diff --check` -- passed.
