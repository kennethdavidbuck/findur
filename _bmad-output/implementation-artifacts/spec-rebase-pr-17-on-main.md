---
title: 'Rebase PR #17 onto main'
type: 'chore'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/spec-2-2-complete-the-minimum-personal-profile.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** PR #17 no longer applies cleanly because the portfolio showcase work merged into `main` after the profile branch diverged.

**Approach:** Rebase the PR branch onto `main` with `git pull --rebase origin main`, compose both features at each conflict, regenerate API bindings from the combined OpenAPI contract, verify the integrated branch, and update the PR with `git push --force-with-lease`.

</frozen-after-approval>

## Implementation Notes

- Ran `git pull --rebase origin main` on `feat/2-2-minimum-personal-profile` and replayed the original three PR commits onto `ef0441c`.
- Preserved `main`'s portfolio showcase services, routes, fixtures, styles, and browser journey while composing the profile service, API, page, styles, and journey alongside them.
- Merged the authored OpenAPI contract and regenerated both Go and TypeScript bindings rather than hand-merging generated files.
- Updated the shared navigation test with endpoint-specific showcase and profile fixtures and a live heading lookup after the async profile route replaces its loading view.
- Verified the combined backend, frontend, generated artifacts, mandatory lint, production build, and Docker Compose integration contracts.
- Hardened the shared navigation test so unexpected API requests receive 404 instead of silently reusing the showcase fixture.

## Review Triage Log

- low / patch -- The navigation test accepted unknown URLs as showcase requests; it now recognizes the status, showcase, and profile endpoints explicitly and returns 404 otherwise.
- false / rejected -- A separate Verification section is intentionally absent because the Build workflow requires one-shot specs to contain only frontmatter, frozen Intent, and Implementation Notes; exact commands and outcomes are reported in the handoff.
