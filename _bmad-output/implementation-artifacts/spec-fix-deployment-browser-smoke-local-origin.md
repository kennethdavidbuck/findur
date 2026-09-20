---
title: 'Fix deployment browser smoke local-origin coupling'
type: 'bugfix'
created: '2026-09-20'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-0-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 0.3 added a loopback-only OAuth journey to the browser helper shared by Compose integration and production deployment smoke. The production Selenium container cannot reach a frontend at its own `127.0.0.1:8080`, causing the post-deployment browser smoke to fail after the exact build has deployed successfully.

**Approach:** Make the local OAuth journey explicitly opt-in for the Compose integration caller while retaining public-origin exact-build verification for the standalone deployment caller.

</frozen-after-approval>

## Implementation Notes

- Split public exact-build and local OAuth browser verification into explicit helpers, with only bounded WebDriver session management shared between them.
- Kept the Compose integration journey enabled with its configured loopback origin while the direct deployment-smoke entry point remains status-only; no production endpoint, provider behavior, or deployment trigger changed.
- Restricted the OAuth helper to an exact HTTP `127.0.0.1` origin so it cannot be accidentally aimed at a deployed environment.
- Verified Node syntax and whitespace, ran the status-only browser entry point successfully against the deploy-shaped stack, and ran the full Compose integration journey successfully after the split and again after review hardening.

## Review Triage Log

| Finding | Verdict | Evidence and route |
|---|---|---|
| BH-1 | medium | The new OAuth helper accepted any origin and could be misdirected at production. Patched it to accept only an exact HTTP `127.0.0.1` origin and verified rejection of a public HTTPS origin. |
| BH-2 | false | The browser shares the frontend network namespace and intentionally addresses its fixed internal port 8080; host-side Compose port overrides do not change that browser-visible origin. |
| BH-3 | low | A session-deletion failure could mask an earlier assertion, but this behavior predates the fix, requires a rare double failure, and adding multi-error cleanup handling is not justified for this regression. |
| BH-4 | false | The workflow waits for Selenium readiness before running the helper, and WebDriver protocol responses are JSON; the proposed proxy/plain-text response was not shown reachable in this topology. |
| BH-5 | low | PR CI does not reproduce the standalone container topology, but the production entry point now imports only status verification and was exercised directly without the OAuth helper. A navigation-recording harness would add disproportionate machinery. |
| BH-6 | false | This is a one-shot spec, whose workflow format intentionally retains only frontmatter, frozen intent, and implementation notes; verification evidence is recorded above. |
