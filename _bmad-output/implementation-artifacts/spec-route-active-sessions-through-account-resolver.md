---
title: 'Route Active Sessions Through Account Resolver'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The active-session consent bypass goes directly to the persisted Portfolio Showcase. Sessions without committed included accounts therefore receive an empty response, and the bypass never invokes the existing account-recovery/onboarding flow.

**Approach:** Send an authenticated consent bypass to the protected account resolver. It must retain the direct portfolio destination for owners with committed inclusion, while loading account onboarding for owners whose portfolio inclusion is incomplete.

</frozen-after-approval>

## Implementation Notes

- `OnboardingAccountSelection` already resolves inclusion and replaces its route with `/portfolio` when an owner has committed accounts; reuse it instead of adding a new account-state endpoint or a provider call from Showcase.
- The Portfolio Showcase intentionally reads only persisted inclusion and has no provider dependency. An empty `/api/portfolio/showcase` response is therefore expected before account inclusion, not a WireMock failure.
- Updated both active-session paths—preflight navigation and a direct `/connect` session check—to enter `/onboarding/accounts`; its existing inclusion resolver preserves `/portfolio` for completed owners.
- Updated active-session tests to supply committed inclusion and a non-empty Showcase, proving the resolver reaches the persisted Portfolio only after it establishes the owner is complete.
- Added an active-session regression with no committed inclusion; it proves the chooser is rendered and requests the saved inventory rather than an empty Showcase.
- A pending inclusion change without committed accounts now remains in account selection, avoiding an empty Showcase before its persisted account data is available. Tests cover both the Login-link preflight and a direct `/connect` active-session entry.

## Review Triage Log

1. **medium / patched:** The resolver previously treated a pending addition as sufficient for Showcase even when no account was committed, reproducing the empty-state problem. Only committed inclusion now completes the resolver.
2. **medium / patched:** The completed-owner test could have passed with a direct Showcase route. It now proves the inclusion request occurs before the Showcase request.
3. **medium / patched:** The account-selection regression covered only the Login link. A direct active `/connect` visit now has equivalent coverage.
