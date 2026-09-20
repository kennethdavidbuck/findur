---
title: 'Configure GitHub Actions for pull requests'
type: 'chore'
created: '2026-09-19'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The CI workflow's pull-request trigger is implicit and does not cover GitHub merge-queue validation, so PR checks are not fully configured for the repository's `main` integration path.

**Approach:** Make the `main` pull-request trigger explicit, run the same checks for merge groups, and prevent superseded PR runs from consuming CI capacity while retaining push validation on `main`.

</frozen-after-approval>

## Implementation Notes

- Updated `.github/workflows/ci.yml` to scope pull-request runs to the `main` base branch and to run the same stable `backend` and `frontend` checks for merge groups.
- Added workflow concurrency keyed by PR number when available and otherwise by ref, so a new commit cancels only an obsolete run for the same PR or ref.
- Retained the existing read-only token permissions and the `push` trigger for post-merge validation on `main`; no secrets or elevated permissions are needed for fork pull requests.
- Verified the workflow parses as YAML and passes `git diff --check`. The trigger and concurrency expressions use documented GitHub Actions contexts; superseded-run cancellation is limited to `pull_request` events.

## Review Triage Log

- `medium` -- unconditional cancellation could discard post-merge or merge-queue validation; patched by limiting cancellation to `pull_request` events.
- `false` -- limiting pull requests to `main` matches the repository's sole integration branch and existing `push` policy; no maintenance or release branches exist.
- `low` -- the implementation record initially lacked verification evidence; patched with the completed static checks.
- `low` -- status remained `in-progress` before finalization; patched to `done`.
