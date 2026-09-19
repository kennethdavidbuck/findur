<!-- bmad:context -->
<!-- Verified 2026-09-19 against 4033c33. Managed by bmad-project-context. -->

## findur

Findur is an early-stage application. Planning artifacts live in `_bmad-output/planning-artifacts/`; seeded external context lives in `docs/context/`.

## Policy

- Treat every tracked file as public. Never commit credentials, OAuth tokens, API keys, user secrets, `.env` files, real user financial data, or unredacted logs.
- Before staging externally sourced or generated material, scan it for sensitive values and remove anything unnecessary.
- Use Conventional Commit messages for every commit, such as `docs: add research report` or `feat(auth): add OAuth callback`.

<!-- /bmad:context -->
