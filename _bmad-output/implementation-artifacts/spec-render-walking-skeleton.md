---
title: 'Render walking skeleton'
type: 'feature'
created: '2026-09-19'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '3610316cfeeb8a6542e201675434598982cf94e2'
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur has approved planning artifacts but no executable application or deployment. Continuing feature specifications before validating the chosen Render topology would compound untested assumptions.

**Approach:** Build and deploy the smallest useful vertical skeleton: a Go API with PostgreSQL-backed readiness and graceful shutdown, a React/Vite status shell, migrations, CI, and a Render Blueprint for the static site, web service, and 30-day free database.

## Boundaries & Constraints

**Always:** Work directly on `main` using Conventional Commits. Keep application code clearly separated under top-level `backend/` and `frontend/` directories, with only shared delivery/configuration files at repository root. Use Go 1.25+, `net/http`, JSON `slog`, `pgx/v5`, `golang-migrate/v4`, React/Vite/TypeScript, and locked dependencies. Expose transport-only `GET /api/healthz` and `GET /api/readyz`; readiness depends on PostgreSQL while liveness does not. On `SIGTERM`/`SIGINT`, fail readiness, stop new work, drain HTTP for at most 25 seconds, then close PostgreSQL. Define free Render resources in `virginia`, use the database private connection string, and gate deploys with GitHub checks. Keep all tracked material public and credential-free.

**Never:** Implement OAuth, SnapTrade, product-domain tables, synthetic users, discovery, matching, OpenAPI product endpoints, Redis, JWTs, a service worker, credentialed cross-origin CORS, or production controls. Do not claim the larger Story 0 proxy/cookie or SnapTrade gates are complete. Docker Compose and WireMock are the next tranche, not a prerequisite for this first deploy.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|----------------------------|----------------|
| Liveness | Process is serving | `/api/healthz` returns `200` JSON without touching PostgreSQL | Never disclose internal errors |
| Ready | PostgreSQL responds within two seconds | `/api/readyz` returns `200` JSON | Safe categorical response |
| Not ready | PostgreSQL is absent, slow, or shutdown has begun | `/api/readyz` returns `503`; liveness remains independent while process runs | Log only safe category and latency |
| Frontend | API health succeeds or fails | Accessible status region reports connected or temporarily unavailable | No stack/provider details |
| Repeat migration | Schema is already current | Migration command exits successfully | Treat `migrate.ErrNoChange` as success |
| Termination | Root context is cancelled | Readiness fails, HTTP drains, then pool closes | Deadline expiry produces non-zero exit |

</frozen-after-approval>

## Code Map

- `AGENTS.md` -- public-repository security and Conventional Commit policy; do not modify.
- `_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md` -- controlling AD-1, AD-2, AD-13 through AD-15, and AD-22 constraints; do not modify.
- `backend/cmd/findur/main.go` -- new process composition, signals, HTTP server, and PostgreSQL lifecycle.
- `backend/cmd/migrate/main.go`, `backend/db/migrations/` -- new idempotent startup migration path; begin with a no-op baseline rather than premature domain tables.
- `backend/internal/platform/config`, `backend/internal/platform/httpapi`, `backend/internal/platform/lifecycle`, `backend/internal/platform/migrations` -- new focused, testable platform units.
- `frontend/` -- new locked React/Vite status shell; no product screens or protected browser persistence.
- `.github/workflows/ci.yml` -- new authoritative Go/frontend/migration checks.
- `render.yaml` -- new Blueprint with `findur-api-kdb`, `findur-web-kdb`, and `findur-db`; static route targets the API service hostname and precedes the SPA fallback.
- `README.md`, `scripts/smoke-deployment.sh` -- local commands and single-origin deployed smoke check.

## Tasks & Acceptance

**Execution:**
- [x] `backend/go.mod`, `backend/internal/platform/config`, `backend/internal/platform/httpapi` -- add pinned Go foundation, safe configuration, liveness/readiness handlers, request metadata, and tests.
- [x] `backend/internal/platform/lifecycle`, `backend/cmd/findur/main.go` -- compose pgxpool, explicit server timeouts, signal cancellation, readiness transition, ordered graceful drain, and tests.
- [x] `backend/db/migrations`, `backend/internal/platform/migrations`, `backend/cmd/migrate/main.go`, `scripts/render-start.sh` -- add the locked, idempotent migration-before-start path and migration tests.
- [x] `frontend/` -- add the locked Vite/React/TypeScript shell, same-origin health client, accessible status behavior, tests, lint, and production build.
- [x] `.github/workflows/ci.yml`, `.gitignore` -- run Go format/vet/test/build, frontend lint/type/test/build, and migrations twice against a PostgreSQL service.
- [x] `render.yaml` -- define the two services and free database with `checksPass`, private database wiring, health check, `/api/*` rewrite, and SPA fallback.
- [x] `README.md`, `scripts/smoke-deployment.sh` -- document operation and verify the public shell plus rewritten health/readiness without printing sensitive headers.
- [x] Repository -- run the complete local suite, scan tracked changes for sensitive material, and prepare the staged baseline diff for BMAD review. Commit, push, and live Render proof follow the required review gate.

**Acceptance Criteria:**
- Given a clean checkout with supported Go/Node and PostgreSQL, when documented verification runs, then all tests, vet/lint/type checks, builds, and two consecutive migrations pass.
- Given PostgreSQL is available, when the Go process starts, then migrations complete before readiness succeeds and both health endpoints return safe JSON.
- Given PostgreSQL becomes unavailable, when readiness is requested, then it returns `503` without changing liveness or exposing connection details.
- Given the process receives a termination signal, when shutdown begins, then readiness fails before new work is admitted, HTTP drains within 25 seconds, and the database closes afterward.
- Given GitHub checks pass and the Blueprint is deployed, when the static-site origin is smoke-tested, then `/`, `/api/healthz`, and `/api/readyz` all succeed through that one origin.
- Given GitHub checks fail or are absent, when Render evaluates `checksPass`, then it does not automatically deploy the failing revision.

## Implementation Notes

- Implemented the user-requested top-level `backend/` and `frontend/` separation; root holds only shared CI, Render, smoke, tooling, and documentation files.
- Pinned Go, Node, Go modules, and npm packages; dependency lockfiles contain only their standard public registries.
- Independently passed Go race tests/vet/build, frontend tests/typecheck/lint/build, shell syntax, whitespace, and sensitive-value scans.
- Applied the baseline migration twice against disposable PostgreSQL 18.6, then verified the final Render start script migrates before starting the API.
- Live local probes returned health/readiness `200`; after PostgreSQL stopped, health remained `200` while readiness returned safe `503`; `SIGINT` exited cleanly.
- Removed a duplicate in-process migration call during audit so deployment uses the separate one-shot migration executable before `exec` of the server.
- Actual GitHub checks, Render hostname allocation/rewrite, and public smoke verification remain external evidence after review, commit, and push.

## Spec Change Log

## Review Triage Log

| Layer | Finding | Verdict | Route and evidence |
|---|---|---|---|
| verification-gap | Readiness log secrecy lacks a regression assertion | medium | patch — the failure test discards logs, so a future raw-error field could expose the injected credential and host without failing tests. |
| verification-gap | Retry interaction lacks a behavioral test | medium | patch — the button renders, but no test presses it or proves that a second request can recover the UI. |
| verification-gap | Process-level graceful-shutdown composition is unverified | medium | patch — helper ordering is tested and a manual process smoke passed, but CI does not exercise the real startup/signal boundary. |
| verification-gap | Deployment smoke can pass with a missing frontend bundle | medium | patch — it checks only the empty root element and API responses, not the referenced JavaScript asset. |
| edge-case-hunter | Cancellation during the initial database ping is reported as database failure | low | patch — the ping inherits the root context and the error branch does not distinguish intentional cancellation. |
| edge-case-hunter | `PORT` accepts zero or a service name | medium | patch — `net.LookupPort` accepts those forms even though Render supplies a numeric listening port. |
| edge-case-hunter | Readiness can return success after draining begins | medium | patch — `Check` reads `accepting` before the ping and does not re-check it after a successful ping. |
| edge-case-hunter | A client disconnect can leave a truncated status body logged with the sent success status | low | rejected — HTTP status is already committed before encoding can fail, and adding response plumbing for this rare health-response disconnect is disproportionate. |
| edge-case-hunter | Client cancellation is categorized as database unavailability | low | patch — all errors other than `errNotAccepting` currently map to `database_unavailable`; a direct categorical mapping is sufficient. |
| edge-case-hunter | Migration close failures are discarded | medium | patch — the deferred `migrator.Close` explicitly ignores both close errors, allowing a false-success exit. |
| edge-case-hunter | A health request that never settles leaves the UI checking forever | medium | patch — `fetchHealth` has no timeout or abort signal. |
| edge-case-hunter | Smoke-test URL accepts paths, credentials, queries, and fragments | low | patch — only the `https` prefix is validated, so malformed bases can produce misleading requests. |
| blind-hunter | `pgxpool.Close` can exceed the shutdown deadline | false | rejected — the current handlers only ping with a two-second context, HTTP drain precedes pool close, and this change has no unbounded acquired-connection path. |
| blind-hunter | Direct binary startup can report ready before migrations | false | rejected — the deployed process contract is the migration-first start script; direct binary invocation is not the deployment entry point. |
| blind-hunter | `http.Server.ErrorLog` remains unstructured | medium | patch — standard-library server diagnostics bypass the JSON logger used for the rest of the service. |
| blind-hunter | Logging writer does not expose `Unwrap` | low | patch — response-controller behavior can be preserved with a direct `Unwrap` method and no new application surface. |
| blind-hunter | Logging writer records the last repeated `WriteHeader`, not the first | low | rejected — no current handler writes headers twice, and adding recorder state for an invalid future handler is unnecessary for this skeleton. |
| blind-hunter | `PORT` accepts service names and zero | medium | patch — independently confirms the same reachable configuration defect reported by the edge-case layer. |
| blind-hunter | Frontend health fetch lacks a timeout | medium | patch — independently confirms that a stalled request prevents the recovery control from appearing. |
| blind-hunter | Frontend tests omit retry, non-2xx, invalid JSON, and stalled requests | medium | patch — retry is user-visible and timeout/error branches need focused behavioral coverage. |
| blind-hunter | Deployment smoke does not prove the frontend bundle loads | medium | patch — independently confirms that the root-element grep can pass for a blank application. |
| blind-hunter | CI does not exercise migration-first startup, live probes, and signal shutdown | medium | patch — those behaviors were manually verified but are absent from the authoritative GitHub gate. |
| blind-hunter | CI omits the race detector | low | patch — the readiness/lifecycle code is concurrent and changing the existing Go test command to `-race` is direct. |
| blind-hunter | Go module and CI toolchain versions differ | false | rejected — the `go` directive defines the language/module baseline while CI intentionally verifies it on a newer pinned Go 1.25+ toolchain. |
| blind-hunter | The declared npm package-manager version is not explicitly activated | low | rejected — the lockfile is authoritative here; the plausible npm-v11 variation is unlikely to alter this install and enforcing it adds bootstrap work. |
| blind-hunter | GitHub Actions use mutable major tags | low | rejected — jobs have read-only repository permission and pinning external action commits is disproportionate to this credential-free walking skeleton. |
| blind-hunter | CI jobs have no execution timeout | low | patch — a direct job timeout prevents a stuck probe, migration, or test from consuming the platform default duration. |
| blind-hunter | The reserved Render API hostname might be unavailable | maybe-false | defer — only Blueprint creation can establish whether Render allocates `findur-api-kdb.onrender.com`; failure would require the documented same-origin fallback. |

## Design Notes

Render cannot interpolate another service's public hostname into a static rewrite. The Blueprint therefore reserves the explicit globally unique API name `findur-api-kdb` and targets `https://findur-api-kdb.onrender.com/api/*`. If Render rejects that name or the deployed rewrite does not preserve the required semantics, record the evidence and use the architecture's fixed same-origin fallback: serve the Vite build from Go rather than introducing credentialed CORS.

## Verification

**Commands:**
- `cd backend && go test ./... && go vet ./...` -- all backend tests and static analysis pass.
- `cd backend && go build ./cmd/findur ./cmd/migrate` -- both executables compile.
- `npm --prefix frontend ci && npm --prefix frontend test -- --run && npm --prefix frontend run typecheck && npm --prefix frontend run lint && npm --prefix frontend run build` -- locked frontend verifies and builds.
- `cd backend && go run ./cmd/migrate && go run ./cmd/migrate` with CI `DATABASE_URL` -- first and no-change migrations both succeed.
- `./scripts/smoke-deployment.sh https://<actual-static-host>.onrender.com` -- shell and rewritten API checks pass after the user creates the Blueprint.
