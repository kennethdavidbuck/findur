---
title: 'Backend readability cleanup'
type: 'refactor'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

Clean up long, dense handwritten backend functions, starting with portfolio inclusion and showcase and applying the same conservative style to nearby hotspots. Extract cohesive responsibilities, improve names and logical spacing, and expand hard-to-read composite literals. Keep function parameter lists on one line and retain the established Go style. Preserve behavior, public interfaces, SQL semantics, transaction and lock order, provider-call order, error handling, and serialization. Make incremental edits and retain completed work; do not discard the cleanup wholesale.

</frozen-after-approval>

## Implementation Notes

- Investigation: the main orchestration hotspots are `backend/internal/platform/postgres/inclusion.go` and `showcase.go`. Adjacent dense code is in provider inventory normalization and HTTP portfolio response mapping; the portfolio domain services need only modest spacing and separation of freshness policy. Generated code, contracts, migrations, auth lifecycle, and unrelated files remain outside the edit scope.
- No unresolved intent choices or irreversible actions. This is a bounded internal refactor with no new public API. Existing repository conventions take priority over broad architecture migrations (named SQL arguments and application transaction ownership are separate work).
- Plan: extract removal and dataset publication helpers; separate showcase account loading from projection; clarify provider account normalization and HTTP mapping literals; apply matching light cleanup to domain inclusion/showcase. Preserve query ordering, same-transaction execution, nil/empty values, and failure precedence.
- Verification: focused portfolio/provider/HTTP/PostgreSQL tests, full backend race suite, `cd backend && golangci-lint run`, formatting and diff checks. PostgreSQL tests use isolated Testcontainers fixtures. Baseline test invocation initially hit the sandbox's read-only Go cache and was retried with the required permission.
- Implemented incremental cleanup in six Go files: PostgreSQL inclusion preparation/removal/publication helpers; showcase account loading and usability predicate; provider account metadata and usability classification helpers; expanded HTTP response mappings; domain inclusion spacing and activity freshness helper. Existing single-line parameter lists, transaction scopes, statement order, status precedence, and nil/empty response handling are preserved.
- Verification passed: baseline `cd backend && go test ./internal/portfolio ./internal/platform/postgres ./internal/platform/provider ./internal/platform/httpapi`; post-edit focused `cd backend && go test ./internal/portfolio ./internal/platform/postgres`; full `cd backend && go test -race ./...`, including PostgreSQL integration tests.
- Mandatory `cd backend && golangci-lint run`: passed, `0 issues.` The first attempt found no binary on PATH; reran successfully with the existing `/tmp/findur-golangci/golangci-lint-2.13.0-linux-amd64` directory added to PATH and cache access permitted. No dependency or tool-version files changed.
- Review refinements: documented preparation lock preconditions and existing account-usability precedence; renamed the account initializer; reformatted publication and showcase SQL clauses; extended existing provider characterization tests with closed/unavailable holdings and missing-status/unsupported-category overlaps. This adds one test file to the six implementation files.
- Final verification after refinements: `cd backend && go test -race ./...` passed, including PostgreSQL integration and the new provider cases; `cd backend && golangci-lint run` passed with `0 issues.` using the same PATH setup. `gofmt -l` returned no files, and `git diff --check` passed. Independent follow-up review found no concrete correctness issues; no work was deferred.

## Review Triage Log

- Low, patched: the extracted preparation helper depended on caller-held owner/inclusion/inventory locks and prior version/target validation. Its sole caller already satisfies these; an explicit precondition comment now preserves that contract for future edits.
- Low, rejected: grouping three version/generation arguments into a new private state type would add a one-use abstraction without correcting a demonstrated mismatch. The sole call passes `version, inventoryGeneration, lifecycleGeneration` in declaration order, matching existing repository conventions and the requested conservative style.
- Low, patched: `accountMetadata` also supplied initial availability and sync defaults. Renamed it `initialAccount` so its initialization role is clear.
- Low, patched documentation: unavailable holdings already set availability false before usability classification, so the account-unavailable reason wins over sync-unavailable. Preserved this existing behavior and documented its precedence rather than changing the product outcome.
- Medium, patched: existing provider tests covered isolated provisional conditions but not overlapping classification conditions. Added three characterization cases through the provider `Load` boundary, asserting reasons and selection/eligibility outcomes.
- Low, patched: extracted dataset publishers retained dense SQL lines. Split clauses across lines in the publishers and showcase head readers, preserving SQL tokens, argument order, and single-line function signatures.
