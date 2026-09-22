---
title: 'Fix CI Pending Inventory Observer Test'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The pending-inventory observer regression test can observe its action before the initial inventory effect issues a request, producing a CI-only timing failure.

**Approach:** Synchronize the test on the initial inventory request it asserts, then retain its explicit second-request and rendered-result checks.

</frozen-after-approval>

## Implementation Notes

- `PortfolioPage` renders the pending action before its `useEffect` starts the shared initial inventory request, so finding the action is not a valid synchronization point.
- Update only the focused `App.test.tsx` assertion; application behavior and request implementation remain unchanged.
- Replaced the immediate initial-call assertion with `waitFor`, preserving the test's proof that exactly one shared initial request is made before the observer's explicit check starts a second request.
- The focused test title now names its actual contract: a pending observer can explicitly retrieve persisted status; the no-polling implementation remains outside this timing-race regression.
- Verification: `cd frontend && npm test -- --run src/App.test.tsx` and `npm test -- --run` passed (62 focused / 80 total); `npm run typecheck` and `npm run build` passed; `npm run lint` passed with four existing warnings. `cd backend && golangci-lint run` and `golangci-lint run ./...` both failed before analysis with the local toolchain error `no go files to analyze`.

## Review Triage Log

1. **low / patched:** The test title claimed it proved no polling, while its assertions cover the explicit status check. Renamed it to its actual regression contract, avoiding a misleading coverage claim without adding brittle timer control.
2. **low / patched:** The work artifact did not yet state completion or verification. It now records the completed checks and the backend lint toolchain blocker.
