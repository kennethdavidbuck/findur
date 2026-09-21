---
title: 'Fix deployment health smoke'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: 'b861582359e452733b1618d939327d95bfe26a54'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Render independently retains unchanged frontend or backend resources, so their valid build SHAs can differ after a partial deployment. The status page and production smoke currently treat that expected state as a failed deployment.

**Approach:** Define deployment health as a successfully loaded public frontend and a ready same-origin API. Keep both build SHAs visible for diagnosis, but do not require them to match in the status page or deployment smoke checks.

**Decision:** The post-deploy gate verifies the health of the application currently served at the public origin. It exercises `/__status` in a browser; a consistently cached older frontend is not a release-freshness failure when that frontend and its ready same-origin API are healthy.

</frozen-after-approval>

## Implementation Notes

- Reclassified `/__status` from exact-build comparison to public deployment health: a ready same-origin API is healthy, while frontend and API revisions remain separate diagnostics.
- Removed the incoming-SHA equality requirement from the shell and browser deploy smoke checks; each still validates its own full revision metadata, public assets, same-origin routing, and API readiness.
- Updated English/French status copy, unit coverage, CI labels, and the deployment command documentation.
- Verification: frontend test/typecheck/lint/build, shell and Node syntax checks, and `git diff --check` passed. Frontend lint retains two pre-existing Fast Refresh warnings. `cd backend && golangci-lint run` could not run because `golangci-lint` is not installed.
- Paused after new deployment evidence suggested CDN or browser caching. A health-only smoke can pass a consistently cached older frontend, so the release-freshness guarantee needs an explicit human decision.
- Local verification: built and served the frontend with SHA `89abcdef0123456789abcdef0123456789abcdef` while the ready backend reported `0123456789abcdef0123456789abcdef01234567`; both the shell smoke and a real Chromium `/__status` smoke passed.

## Code Map

- `scripts/smoke-deployment.sh` -- public-origin shell smoke; validates the shell and `/__status` routes, frontend-module identity, and API liveness/readiness without comparing frontend and API revisions.
- `test/integration/browser-status.mjs` -- executes `/__status` in a real browser; can verify the rendered frontend revision, same-origin readiness request, and API diagnostic separately.
- `.github/workflows/ci.yml` -- invokes the backend deploy API and frontend deploy hook on every main push, then runs shell and browser deployment-health smoke checks.
- `frontend/src/status.ts`, `frontend/src/pages/StatusPage.tsx` -- classify and present readiness plus revision diagnostics; do not restore a user-facing failure merely because independently served revisions differ.

## Review Triage Log

- **false** — Runtime validation of the frontend SHA is unnecessary: the production Vite build rejects a missing or malformed SHA before a deployable frontend exists.
- **low, rejected** — The UI intentionally treats an unavailable revision diagnostic separately from API readiness; production readiness responses are validated full SHAs by the smoke checks.
- **patch** — Removed locale-dependent English label assertions from the browser smoke and retained structural diagnostic assertions.
- **false** — `/__status` is an SPA rewrite to the same `index.html` and module asset already validated by the shell smoke.
- **patch** — Retried the browser health smoke for bounded post-deploy propagation.
- **false** — A cached legacy frontend can retain its former mismatch behavior, but the approved health-only contract applies once this deployed frontend is served.
- **low, rejected** — The mixed-revision behavior lacks a fixture-only automated test, but a deploy-shaped local Compose run with intentionally different real frontend/backend SHAs passed both shell and browser smoke checks.
