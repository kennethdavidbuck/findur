---
title: 'Polish cash portfolio evidence and sync feedback'
type: 'chore'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Cash-only accounts currently retain an empty Positions card that adds no useful portfolio evidence, and the manual account-sync refresh control does not use the app's existing busy-cursor feedback.

**Approach:** In the private Portfolio Showcase, suppress Positions only after that dataset is safely available and has no holdings, while retaining balances, activities, and non-current dataset feedback. Make a manual “Check again” refresh temporarily non-interactive and use the established progress cursor until the reload completes.

</frozen-after-approval>

## Implementation Notes

- `PortfolioShowcasePage` omits Positions whenever no holdings rows are present, including the initial syncing window; balances, activities, and the broader account syncing notice remain visible.
- The manual Showcase refresh now disables its link and applies the established `cursor: progress` convention until the request settles.
- `/onboarding/accounts` checks durable inclusion first; saved or pending additions replace the chooser with `/portfolio`, while `/portfolio/accounts` remains the explicit editor route.
- Added Vitest coverage for cash-only evidence, refresh feedback, and automatic onboarding routing.
- Verification: frontend tests (77), typecheck, lint (two existing Fast Refresh warnings), production build, and `git diff --check` passed. `cd backend && golangci-lint run` could not run because `golangci-lint` is not installed in the environment.
