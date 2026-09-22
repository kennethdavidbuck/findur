---
title: 'Stabilize Consent Session Transition'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The active-session bypass correctly keeps signed-in users out of staged consent, but an unauthenticated click performs a second status check after navigation and briefly leaves the public layout's main area empty.

**Approach:** Reuse the click's server-authoritative unauthenticated result when entering consent, and give direct `/connect` checks a neutral, stable loading surface. Active sessions must still bypass consent directly to `/portfolio`; unauthenticated users must reach the staged consent page without a collapsing transition.

</frozen-after-approval>

## Implementation Notes

- The defect was introduced by `0a80c4a`: `PublicApp` checks status before navigation, then `ConsentPage` independently checks it again and returns `null` while resolving.
- Keep server status authoritative; do not inspect cookies or persist client-side authentication state.
- Cover the delayed unauthenticated login click so it proves exactly one status request and an immediately rendered consent page after the response; retain active-session redirect coverage.
- Reused the pre-navigation unauthenticated status only for the ensuing consent mount; direct `/connect` still makes its own server-authoritative check and now renders a bounded neutral checking surface instead of an empty main area.
- Keep that neutral surface mounted until an authenticated direct entry is replaced, so the effect-driven redirect cannot introduce a final empty-frame collapse.
- Review hardening makes the click preflight abortable and route-scoped; superseded responses cannot undo a newer navigation, and only a one-time router handoff can seed consent with a completed status.

## Review Triage Log

1. **medium / patched:** A result arriving while the user is already on `/connect` would change the initial-status prop without remounting its hook, leaving the checking state unresolved. The router now keys the one-time consent handoff by request ID.
2. **medium / patched:** A pending Login preflight could complete after a user selected another public page and reverse that newer navigation. Public navigation now aborts the request, and its result checks its cancellation signal.
3. **medium / patched:** Concurrent Login clicks could settle out of order. A monotonically increasing request ID rejects superseded results.
4. **medium / patched:** Browser back/forward could reuse a retained handoff status. The router keeps the handoff in its route state and clears it on `popstate`, so history entry always performs a fresh check.
- Updated `frontend/src/{App.tsx,auth-status.ts,styles.css,pages/ConsentPage.tsx,App.test.tsx}`.
