---
title: 'Refactor OAuth Callback for Submission Quality'
type: 'refactor'
created: '2026-09-20'
status: 'done'
route: 'oneshot'
review_loop_iteration: 1
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/spec-0-3-complete-oauth-and-establish-an-isolated-session.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The verified Story 0.3 callback works, but long procedural functions, repeated protocol and application literals, generated pointer handling, and small overlaps with OAuth/OIDC library helpers make the submission harder to review and maintain. CI also contains a stale hard-coded migration assertion that fails after migration 000003.

**Approach:** Preserve OAuth behavior and security boundaries while extracting cohesive stages and boundary helpers, centralizing constants with their owning packages, using existing OAuth/OIDC helpers, and making the integration schema assertion follow the repository's current migration state.

</frozen-after-approval>

## Implementation Notes

- Centralized provider, route, scope, cookie, callback-state, environment, logging-category, token-kind, and protocol constants in the packages that own them.
- Split OAuth initiation, callback exchange/verification/session establishment, HTTP middleware and input normalization, PostgreSQL finalization, configuration loading, process composition, and the synthetic OIDC provider into focused functions without changing external behavior.
- Replaced manual nonce and HTTP-client context glue with `go-oidc` helpers and retained `oauth2` helpers for PKCE generation, challenge construction, and code exchange.
- Reworked the PostgreSQL adapter integration test into six named scenarios backed by a shared isolated fixture and explicit row-count helpers.
- Added pinned GolangCI-Lint v2 configuration and the official GitHub Action with `gofmt`, `goimports`, static analysis, constant detection, and focused quality linters.
- Fixed CI schema verification to derive the expected version from the latest checked-in up migration instead of hard-coding migration 2.

## Acceptance Criteria

- [x] OAuth/OIDC protocol mechanics use maintained package helpers where their behavior matches the application's error and testability requirements.
- [x] Authorization, callback, HTTP, configuration, persistence, fixture, and composition functions are split into cohesive units with owned constants replacing repeated categorical literals.
- [x] PostgreSQL integration scenarios independently prove claim, replay, rollback, expiration, cleanup, and cancellation behavior.
- [x] GolangCI-Lint runs locally and in CI through an immutable official action release.
- [x] Generated code, race-enabled tests, vet, and the deploy-shaped browser integration remain green.

## Verification

- `gofmt -w cmd internal`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.0 run` — 0 issues
- `go test -race ./...` — passed, including Testcontainers PostgreSQL coverage
- `go vet ./...` — passed
- `go generate ./...` and generated API diff check — passed
- `docker compose config --quiet`, build, health-gated startup, and browser integration — `integration contracts passed`
- `git diff --check` — passed

## Review Triage Log

1. **Accepted and fixed:** `oauth2.GenerateVerifier` can panic on entropy failure and bypassed the injected reader. Verifier bytes now use the service's fallible entropy source while `oauth2.S256ChallengeOption` remains responsible for standards-sensitive PKCE challenge construction.
2. **Accepted and fixed:** generation-failure coverage now tests state, nonce, verifier, browser binding, and verifier-envelope entropy failures independently.
3. **Accepted and fixed:** the rollback test now uses a distinct identity subject, proving the duplicate-session constraint is the failure source.
4. **Accepted and fixed:** rollback verification includes `oauth_attempts` row count and proves the failed attempt remains non-terminal in `exchanging` state.
5. **Accepted and fixed:** cleanup coverage now seeds retained controls and checks exact removed and retained hashes across pending, exchanging, and terminal states.
6. **Accepted and fixed:** the synthetic OIDC fixture now has direct JWKS, Basic authentication, signed nonce, token validation, revocation, malformed-form, method, and route tests.
7. **Partially accepted and fixed:** focused tests cover disabled authorization composition, fixture integration gating/creation, and healthcheck dispatch. An "invalid fixture configuration" branch is not added because `config.Load` validates those values before composition; the fixture constructor's remaining errors are cryptographic operations rather than configuration branches.
8. **Accepted and fixed:** callback test constants are now used consistently instead of duplicating their literals.
9. **Accepted and fixed:** `golangci/golangci-lint-action` is pinned to the immutable SHA for v9.3.0; the linter binary remains pinned to v2.13.0.
10. **Accepted and fixed:** this artifact now records acceptance criteria, exact verification, review iteration, and completion status.

No review finding was deferred.
