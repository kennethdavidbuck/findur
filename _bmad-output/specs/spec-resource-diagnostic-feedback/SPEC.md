---
id: SPEC-resource-diagnostic-feedback
companions:
  - failure-modes.md
  - ../spec-browser-safe-error-contract/SPEC.md
  - ../spec-browser-safe-error-contract/failure-contract.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Resource diagnostic feedback

## Why

Account setup and Portfolio Showcase can return valid but degraded connection, account, inclusion, or dataset state while telling users only that something is unavailable. Users need to understand what succeeded, what is affected, whether saved data remains, and whether to wait, retry, or reconnect without exposing provider internals.

## Capabilities

- **CAP-1**
  - **intent:** Account setup summarizes complete, partial, and failed discovery truthfully.
  - **success:** Mixed outcomes identify how many connections were checked or affected, whether eligible accounts are ready, whether saved choices remain, and when the inventory was updated without claiming total failure after partial success.
- **CAP-2**
  - **intent:** Users can understand and act on each degraded connection.
  - **success:** Active-empty, unsupported-only, disabled, authorization-required, temporarily unavailable, syncing, and unknown connection conditions have distinct safe explanations and only valid contextual actions.
- **CAP-3**
  - **intent:** Users can understand the freshness and recovery state of every included-account dataset.
  - **success:** Balances, positions, and activities render current, stale-usable, expired, syncing, or unavailable feedback even when their row arrays are empty.
- **CAP-4**
  - **intent:** Users receive reason-specific feedback when account inclusion cannot complete.
  - **success:** Every existing inclusion failure reason produces distinct English and French explanation and recovery behavior.
- **CAP-5**
  - **intent:** Background synchronization retains the safe facts required to explain degraded resources after reload.
  - **success:** Applicable reason, retry timing, and last-success metadata survive scheduled work and are returned owner-scoped without discarding policy-permitted stale data.
- **CAP-6**
  - **intent:** Resource feedback remains accessible, localized, and resistant to regression.
  - **success:** Contract, component, and browser tests cover the matrix in `failure-modes.md`, including announcements, timestamps, empty datasets, and action behavior in English and French.

## Constraints

- Request-level failures follow `SPEC-browser-safe-error-contract`; valid partial or degraded snapshots remain HTTP `200` responses.
- Reasons are bounded browser-safe categories and never contain raw SnapTrade responses, provider descriptions, credentials, unmasked financial data, or operator-only diagnostics.
- Reconnect appears only for authorization or connection-repair conditions; transient failures use retry or truthful wait guidance.
- Saved choices and stale-usable evidence remain visible wherever current product policy permits them, with fresh access clearly distinguished from retained data.
- `backend/api/openapi.yaml` is authoritative; durable additions require reversible migrations and owner-scoped persistence.
- English and French provide equivalent meaning, actions, timestamps, and accessibility semantics.
- Verification for every implementation change includes the repository-mandated `cd backend && golangci-lint run`.

## Non-goals

- Changing account eligibility policy or adding providers.
- Redesigning synchronization scheduling or retry algorithms.
- Exposing raw provider error messages.
- Automatically repairing brokerage authorization.
- Redesigning the full Portfolio Showcase.
- Creating or submitting customer-support tickets.

## Success signal

Given the client-reported mixed account state, the page says that some connections were checked but no eligible accounts are ready, explains each institution’s safe condition, shows the last check, and offers scoped retry or reconnect actions; degraded Showcase resources remain visible and explainable after reload.

## Assumptions

- The provider adapter can distinguish zero returned accounts from returned-but-unsupported accounts; if implementation evidence disproves this, both initially map to `no_accounts_returned` rather than inventing a cause.
