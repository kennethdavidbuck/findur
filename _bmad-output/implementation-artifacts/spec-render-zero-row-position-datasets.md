---
title: 'Render zero-row position datasets'
type: 'chore'
created: '2026-09-23'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '_bmad-output/specs/spec-resource-diagnostic-feedback/SPEC.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The Portfolio Showcase hides the entire Positions dataset whenever it contains zero rows, leaving no place to distinguish a genuinely empty account from future syncing or error diagnostics.

**Approach:** Always render the Positions dataset and let the existing dataset presentation show its empty, unavailable, expired, or populated state. Add a concise, styled selection-screen message stating that only investment accounts Findur can currently use are available to select. This does not change eligibility rules or add provider metadata, persistence, or new diagnostic handling.

</frozen-after-approval>

## Implementation Notes

- Removed the position-row-count rendering gate in `PortfolioShowcasePage`; the existing `Evidence` component now presents zero-row position datasets consistently with balances and activities.
- Updated portfolio tests to expect the explicit successful-empty message for positions. No provider, persistence, API, or diagnostic behavior changed.
- Added localized selection-policy copy and a lightweight visual treatment above the existing selection limit; the underlying account filtering and selectability policy remain unchanged.
- Verification passed for the focused 65-test frontend suite, TypeScript typecheck, production build, and `git diff --check`. Frontend lint completed with four pre-existing warnings and no errors. The mandatory `cd backend && golangci-lint run` command was attempted but could not run because `golangci-lint` is not installed.

## Review Triage Log

- `medium` — The first selection-policy sentence overstated the current policy because provisionally categorized accounts can be selectable; patched with general support-boundary and readiness copy in English and French.
- `low` — The original empty-position test relied on a global duplicate-message count; patched by scoping the assertion to the Positions dataset section.
- `medium` — Zero-row unavailable and syncing Positions states lacked section-specific regression assertions; patched with scoped checks.
- `low` — French zero-row Positions rendering lacked direct coverage; patched with a localized dataset assertion.
- `low` — Verification evidence was absent while implementation was in progress; recorded the actual command results before completion.
