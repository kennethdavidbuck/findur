<!-- bmad:context -->
<!-- Verified 2026-09-19 against 4033c33. Managed by bmad-project-context. -->

## findur

Findur is an early-stage application. Planning artifacts live in `_bmad-output/planning-artifacts/`; seeded external context lives in `docs/context/`.

## Policy

- Treat every tracked file as public. Never commit credentials, OAuth tokens, API keys, user secrets, `.env` files, real user financial data, or unredacted logs.
- Before staging externally sourced or generated material, scan it for sensitive values and remove anything unnecessary.
- Use Conventional Commit messages for every commit, such as `docs: add research report` or `feat(auth): add OAuth callback`.

<!-- /bmad:context -->

## Verification

- Run `cd backend && golangci-lint run` as a mandatory part of verification for every code change. Do not omit it, and report the exact command and result in the handoff.

## Code style

- Format multiline SQL with one selected column, assignment, join condition, or predicate per line where practical. Indent clauses and nested conditions so the query can be reviewed without mentally splitting long lines.
- Use table-driven Go tests for related behavior and failure scenarios. Give each case a descriptive name and assert its expected outcome.
- Use named constants for repeated domain states and policy values such as retry limits and timing; avoid unexplained literals in implementation code.

## Pull requests

- Use `.github/pull_request_template.md` for every pull request. Complete every section, remove placeholder text, report only verification that actually ran, and write `Not applicable` where a section does not apply.
