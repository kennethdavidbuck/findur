---
title: 'Restore eligibility for null account categories'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The bulk account-discovery fix excludes every real account whose SnapTrade `account_category` is null, leaving the account chooser empty for the affected user. This is a confirmed regression from the prior provisional-category behavior.

**Approach:** Restore null or missing account categories as provisionally selectable when every other existing eligibility condition passes, while continuing to exclude explicit `DEPOSIT` and `LOC` categories. Apply the same policy to fresh provider normalization, cached inventory projection, and inclusion admission; update the policy documentation and regression coverage without changing the two-request discovery design or other eligibility rules.

</frozen-after-approval>

## Implementation Notes

- Restored null/missing category as selectable-but-provisional in fresh provider normalization; unknown non-null enum values remain excluded.
- Applied the same predicate to cached inventory projection and inclusion admission without rewriting immutable inventory rows.
- Updated focused provider/PostgreSQL regressions, the deterministic large-inventory oracle, and eligibility documentation.
- Verification passed: `GOCACHE=/tmp/findur-go-cache go test -race ./...`, `GOCACHE=/tmp/findur-go-cache GOLANGCI_LINT_CACHE=/tmp/findur-golangci-cache PATH=/tmp/findur-golangci/golangci-lint-2.13.0-linux-amd64:$PATH golangci-lint run` (0 issues), and Node syntax checks for both changed integration scripts.
