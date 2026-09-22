---
title: 'Redirect Active Sessions from Staged Consent'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A user with a valid Findur session can reach `/connect` and is offered staged SnapTrade consent again, despite already being signed in.

**Approach:** After the existing server-authoritative status check confirms an active session on `/connect`, replace that route with `/portfolio`; leave unauthenticated consent and all other public routes unchanged.

</frozen-after-approval>

## Implementation Notes

- `ConsentPage` already obtains the server-authoritative session state through `useAuthorizationStatus`; reuse it rather than inspecting cookies or adding an API endpoint.
- Pass a dedicated authenticated-session callback from `PublicApp` so the existing public navigation type stays narrow. The callback replaces `/connect` with `/profile`.
- Retain the existing resolving state and protected-route guard. A session expiring between checks remains safely handled by `ProtectedApp`.
- Add a focused app-level regression test that begins at `/connect`, confirms redirection only after an authenticated status response, and proves no authorization POST occurs.
- Added the redirect effect to `ConsentPage`, wired the replace navigation in `PublicApp`, and covered the behavior in `App.test.tsx` plus the composed real-Chromium session journey.
- The composed browser assertion verifies the destination, the absent authorization action, and the authenticated shell's fixed navigation styling.
- `go generate ./...` changes the embedded OpenAPI payload with local Go 1.27.1 even without source-contract changes; the generated file was restored and the intended Go/CI alignment work is explicitly out of scope.
- Review hardening guards the redirect poll before the authenticated navigation exists, verifies the consent action is gone, and proves Back returns to the prior protected route rather than staged consent.
- The user changed the destination from Profile to Portfolio before final commit; both authenticated redirect paths and their tests now use `/portfolio`.

## Review Triage Log

1. **medium / patched:** The browser polling script read computed style from a missing authenticated navigation node during the redirect window. It now guards that node until Profile is mounted.
2. **low / patched:** The unit test's fetch-call assertion could not prove a native authorization form had not submitted. It now asserts the authorization form is absent after redirect; the composed browser test covers the same visible boundary.
3. **medium / patched:** The original regression did not protect the replacement-history contract. The real-Chromium journey now uses Back and verifies the preceding authenticated onboarding route is restored.
