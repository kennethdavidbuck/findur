---
id: SPEC-browser-safe-error-contract
companions:
  - failure-contract.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Browser-safe error contract

## Why

Users encounter generic, inconsistent, and sometimes raw API failures even when Findur knows the failure category and possible recovery. Establish one browser-safe request-failure contract and one exhaustive presentation path so users receive localized, actionable feedback and future changes cannot silently add an unhandled error.

## Capabilities

- **CAP-1**
  - **intent:** Failed HTTP requests expose stable browser-safe facts that clients can interpret without parsing exception text.
  - **success:** Every documented non-2xx application response validates against the standard envelope in `failure-contract.md`.
- **CAP-2**
  - **intent:** The frontend classifies API, transport, timeout, and malformed-response failures consistently.
  - **success:** Each failure class deterministically produces a typed UI failure code without depending on `Error.message`.
- **CAP-3**
  - **intent:** Users receive localized explanations and recovery choices appropriate to each known failure.
  - **success:** Every API and client-only failure code has exhaustive English and French presentation metadata, while unknown codes use a safe generic fallback.
- **CAP-4**
  - **intent:** User-visible failures preserve work and provide accessible, truthful recovery.
  - **success:** Component tests verify announcements, focus, retained input or state, action availability, and successful retry or redirection for each affected flow.
- **CAP-5**
  - **intent:** Repository controls make the error contract mandatory for future implementation work.
  - **success:** Generated types, exhaustive compilation, contract tests, project guidance, and the pull-request checklist fail or flag any newly introduced code without its required handling.

## Constraints

- `backend/api/openapi.yaml` is the source of truth; generated Go and TypeScript files are never hand-edited.
- The API returns bounded codes and facts, never localized prose, raw provider messages, stack traces, secrets, unmasked financial data, or operator-only diagnostics.
- The frontend owns all user-facing wording and must provide equivalent English and French treatment.
- Independently deployed versions must remain safe: unknown codes receive generic fallback presentation.
- Valid responses containing degraded resources remain successful domain responses and are governed by `SPEC-resource-diagnostic-feedback`.
- Existing owner isolation, request defenses, session recovery, and `no-store` requirements remain unchanged.
- Verification for every implementation change includes the repository-mandated `cd backend && golangci-lint run`.

## Non-goals

- Exposing provider-specific raw diagnostics.
- Changing internal logging, metrics, or operational alerting.
- Persisting connection or dataset diagnostic history.
- Creating a customer-support workflow.
- Redesigning the application’s visual system.

## Success signal

A forced authorization-start, profile, inventory, preference, inclusion, Showcase, or logout failure stays inside Findur, explains the safe failure category in the selected language, offers only a valid recovery action, preserves user work, and never exposes raw JSON; adding a new API error code without a presentation and test fails verification.

## Assumptions

- No external support-ticket system is currently integrated; `supportReference` is omitted unless the backend already has a safe correlation value to supply.
