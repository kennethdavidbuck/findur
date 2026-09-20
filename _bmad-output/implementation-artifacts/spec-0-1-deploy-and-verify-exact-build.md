---
title: 'Story 0.1: Deploy and Verify the Exact Build'
type: 'feature'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '81716e2c4ff9d9f525a20e50b6b3cac0c0cd19bd'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur cannot yet prove that the public frontend and API are the exact commit CI tested, and its local workflow lacks the production-like PostgreSQL/WireMock topology needed for later integration work.

**Approach:** Complete Story 0.1 with exact-SHA container/deploy verification, an unlinked same-origin status diagnostic, and a lean Compose stack for build, production-bundle testing, fixtures, and watch development.

## Boundaries & Constraints

**Always:** Preserve existing checks, public EN/FR/theme/accessibility behavior, PostgreSQL-backed readiness, dependency-free liveness, migration-first startup, graceful drain, and Render `/api/healthz`. Build the backend through multiple stages into a pinned distroless runtime containing only the API, migrator, migrations, and required runtime material. Compose provides pinned PostgreSQL, WireMock, backend, and production frontend proxy/server; its watch mode rebuilds/restarts Go and refreshes Vite. Embed the same validated full Git SHA independently in backend and frontend. Keep fixtures synthetic and tracked material credential-free.

**Never:** Use Docker-in-Docker, the host Docker socket, Testcontainers, live SnapTrade credentials/calls, frontend assets in the backend image, permissive credentialed CORS, or claim Render deploys the byte-identical tested static bundle. Do not add product-domain seed data; “seeded” means migrations plus curated WireMock/integration fixtures until domain stories introduce schemas.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Build identity | Valid full SHA at build time | Health, readiness, frontend metadata, and `/__status` report it | Reject missing/malformed production identity |
| Readiness | Accepting and migrated PostgreSQL reachable | `200 {status:"ready", buildSha}` | Safe `503 unavailable` with SHA; no internal detail |
| Status route | Browser loads `/__status` | One `/api/readyz` call and explicit match result | Unavailable, malformed, stale, or mismatch is failure |
| Deploy | Main checks pass | Publish `linux/amd64` image by SHA/digest, invoke protected hooks, bounded smoke | Newer run cancels older; stale deploy remains pending |
| Proxy contract | Synthetic requests traverse frontend `/api/*` | Body/query/cookies/cache/error semantics survive | Gate fails without CORS workaround |

</frozen-after-approval>

## Code Map

- `.github/workflows/ci.yml` -- preserve backend/frontend gates; add image, fixture, deploy, concurrency, and smoke jobs restricted to main pushes.
- `render.yaml` -- retain static rewrite order and liveness probe; switch backend to exact published image and disable independent auto-deploy where supported.
- `backend/cmd/findur/main.go`, `backend/internal/platform/httpapi/` -- inject build identity while preserving liveness/readiness and drain contracts.
- `backend/cmd/migrate`, `backend/db/migrations`, `scripts/render-start.sh` -- preserve migration-first execution inside the runtime image.
- `frontend/src/App.tsx`, `frontend/vite.config.ts` -- preserve public routes; add unlinked `/__status`, build metadata, and configurable same-origin development proxy.
- `Dockerfile`, frontend container/proxy config, `.dockerignore`, `compose.yaml` -- new multi-stage/distroless backend plus lean production-like stack and watch configuration.
- `test/integration/`, `test/fixtures/wiremock/` -- new synthetic topology/provider request fixtures and browser/API verification; no live provider access.
- `scripts/smoke-deployment.sh`, `README.md` -- bounded expected-SHA smoke and documented build/up/watch/test commands.

## Tasks & Acceptance

**Execution:**
- [x] Container files -- build pinned multi-stage backend and frontend images; use distroless for the backend runtime and retain migrations/health behavior.
- [x] `compose.yaml`, fixtures, integration harness -- run healthy PostgreSQL, WireMock, backend, and production SPA proxy; add watch development and direct service-DNS tests without privileged Docker features.
- [x] Backend/frontend identity and status files -- expose full SHA in both probes and frontend metadata; implement tested `/__status` success and categorical failure states.
- [x] Proxy/provider contract tests -- prove unsafe JSON, callback query, multiple cookies, host-only replay, caching, errors, and bearer-only allowlisted fixture requests; add reproducible contract-generation drift gates only for artifacts introduced here.
- [x] CI/Render/deploy scripts -- retain all current checks, publish exact `linux/amd64` SHA/digest, invoke secret hooks without disclosure, cancel superseded deploys, poll both SHAs, and execute browser plus shell smoke only on main.
- [x] Documentation and security scan -- document simple build/up/watch/integration workflows and verify no secrets or real financial data enter tracked fixtures/logs.

**Acceptance Criteria:**
- Given a PR or non-main check run, when all checks finish, then no deployment occurs and existing build/test/lint/typecheck/migration behavior remains.
- Given the default Compose stack, when services become healthy, then the production bundle reaches the API through SPA fallback and `/api/*`, PostgreSQL is migrated, WireMock fixtures are ready, and no privileged Docker mechanism exists.
- Given a main deploy, when smoke completes, then backend digest and Render frontend derive from the same full SHA and public browser/API evidence matches it; stale, malformed, unhealthy, or mismatched artifacts fail within bounded time.
- Given any health/status/smoke/fixture test, when outbound calls and logs are inspected, then no SnapTrade call, secret, personal data, or financial value appears.

## Implementation Notes

- Added one multi-stage root Dockerfile: pinned Go/Node builders, a non-root distroless backend containing the API, migration command/files, and embedded SHA, plus the production Nginx frontend bundle.
- Centralized backend environment reads in `internal/platform/config`; runtime components, including `localHealthcheck`, receive typed configuration or injected dependencies.
- Compose is the sole infrastructure integration target. It builds explicit full-SHA local image tags, runs PostgreSQL, WireMock, backend, frontend, and profile-scoped browser/integration services, and supports backend rebuild plus frontend source sync through `develop.watch`.
- CI builds the Compose backend image once, tests that exact local image, and on `main` only pushes it to public Docker Hub under the full Git SHA before deploying its registry digest. No mutable release tag or second backend build is used.
- Render image-service creation requires the documented one-time `bootstrap-once` reference. Normal releases never move that tag or edit `render.yaml`; the Docker Hub namespace secret must match the Blueprint's public repository.
- Added in-process proxy/provider regression tests and reusable real-browser verification for local Compose and the deployed public origin.
- Docker execution is unavailable to the agent sandbox because access to `/var/run/docker.sock` is denied. Human verification on 2026-09-20 built both application images, brought PostgreSQL 18, WireMock, backend, and frontend to healthy state, and completed the disposable integration/browser runner with `integration contracts passed`.

## Spec Change Log

## Review Triage Log

| Layer | Finding | Verdict and evidence | Route |
|---|---|---|---|
| verification-gap | Migration execution can be removed without failing CI | **medium** — readiness only pings PostgreSQL and the baseline schema is not consumed, so the current lifecycle gate does not prove `schema_migrations` reached version 1 cleanly. | patch |
| verification-gap | Production SHA validation lacks a `run`-level test | **medium** — helper and configuration tests do not prove `run` connects production mode to `ValidateProduction`; the production image always supplies a valid SHA. | patch |
| blind-hunter | WireMock healthcheck only prints a version | **medium** — the command can succeed before the fixture HTTP listener is ready, allowing health-gated dependants to start early. | patch |
| blind-hunter | Provider client uses `http.DefaultClient` | **medium** — it has no overall timeout, so a stalled fixture/provider exchange can retain a diagnostics request indefinitely. | patch |
| blind-hunter | Status-page readiness fetch has no timeout | **medium** — an open request leaves the page permanently in `checking` rather than the required unavailable state. | patch |
| blind-hunter | Status-page labels bypass EN/FR messages | **medium** — `/__status` is rendered inside the existing locale provider but displays English after the user has selected French. | patch |
| blind-hunter | The `stale` result is currently reachable only via HTTP 409 | **low** — deployed readiness currently emits 200/503, but this does not create a false success: actual cross-revision responses become `mismatch`, and any future 409 still fails closed. The proposed semantic expansion would add unnecessary state rules. | reject |
| blind-hunter | Integration fetches lack per-request timeouts | **medium** — a stalled request can consume the full job timeout and makes the intended bounded gate needlessly slow to diagnose. | patch |
| blind-hunter | Module-asset integration check accepts an SPA fallback | **medium** — it reads the body without checking status or content type, and fallback HTML includes the expected SHA. | patch |
| blind-hunter | Dockerfile bases use version tags rather than digests | **false** — the story's exact-build guarantee applies to the already-built Compose artifact that CI pushes by SHA/digest; it does not promise that rebuilding a commit later reproduces identical base layers, and each base uses an explicit versioned tag. | reject |
| blind-hunter | Compose third-party services use version tags rather than digests | **false** — the integration gate tests one resolved Compose run and publishes only its already-tested backend artifact; explicit service versions satisfy the topology constraint without claiming future rebuild reproducibility. | reject |
| blind-hunter | Frontend watch misses non-source build inputs | **medium** — `frontend-dev` syncs only `src` and rebuilds only for the lockfile, so edits to `index.html`, `package.json`, or Vite configuration can remain stale. | patch |
| blind-hunter | Smoke polling can reuse stale response files | **medium** — curl failures leave the prior iteration's files intact, allowing frontend and API evidence from different iterations to appear converged. | patch |
| blind-hunter | Lifecycle gate does not detect forced container termination | **medium** — `docker compose stop` itself succeeds when Docker escalates to SIGKILL, so graceful shutdown can regress without failing the gate. | patch |
| edge-case-hunter | Status-page connection can remain open | **medium** — confirmed duplicate symptom of the missing fetch timeout; an abort deadline is required to reach unavailable. | patch |
| edge-case-hunter | Migration execution ignores startup cancellation | **medium** — `migrations.Up` is synchronous and does not observe `rootCtx`, so shutdown requested during a blocked migration cannot promptly end startup. | patch |
| edge-case-hunter | WireMock health claim does not probe the listener | **medium** — confirmed duplicate symptom of the version-only healthcheck; HTTP readiness must determine Compose health. | patch |

## Design Notes

The reference Compose patterns retained are health-gated dependencies, an explicit backend build target, and Compose `develop.watch` rebuilds. Its Docker-socket test runner, echo-only service, and separate lint/test/contract image services are intentionally omitted; Findur uses one disposable integration runner. The default Compose path tests deploy-shaped artifacts, while watch mode remains a developer workflow rather than a competing production topology. If the Render static rewrite cannot preserve the proxy contract, use the approved Go-served-SPA fallback and record the evidence.

## Verification

**Commands:**
- `docker compose config && docker compose build && docker compose up --wait` -- all four core services build and become healthy.
- `docker compose run --rm integration` -- production bundle, proxy fidelity, fixture, identity, readiness, and browser checks pass over service DNS.
- `cd backend && go test -race ./... && go vet ./...` -- backend behavior and architecture checks pass.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- locked frontend checks pass.
- `./scripts/smoke-deployment.sh https://HOST EXPECTED_FULL_SHA` -- bounded deployed exact-build verification passes.

**Observed results:**
- Human local Docker run: application images built, all default services healthy, and `./scripts/compose-test.sh` reported `integration contracts passed`.
- Agent checks: backend race suite and vet passed; frontend 15 tests, typecheck, lint (two existing Fast Refresh warnings), and exact-SHA production build passed; shell/Node/YAML/JSON/Compose validation and `git diff --check` passed.
