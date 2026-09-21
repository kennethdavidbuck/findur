---
title: 'Polish Authenticated UI Consistency'
type: 'chore'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md'
  - '_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md'
  - '_bmad-output/implementation-artifacts/spec-2-2-complete-the-minimum-personal-profile.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The authenticated Profile experience feels visually separate from Portfolio because its page width, route-heading hierarchy, and vertical rhythm use a different presentation grammar. Smaller shell inconsistencies, including uneven state layouts and navigation geometry, weaken the final level of polish.

**Approach:** Apply a restrained frontend-only consistency pass using Portfolio and the finalized Constellation design spine as references. Align authenticated page frames, headings, spacing, bounded surfaces, responsive states, and shell styling while preserving Profile's reading-width form and every existing behavior, accessibility contract, locale, and theme.

</frozen-after-approval>

## Implementation Notes

- Standardized the authenticated page frame at 64rem while retaining the Profile form's 45rem reading width.
- Brought Profile and account-selection route headings onto Portfolio's compact hierarchy, tightened Profile section rhythm, and gave the visibility guide a complete bounded surface.
- Corrected the two-item mobile navigation grid and defined the display, mono, and disabled-text tokens already referenced by authenticated styles.
- Kept loading/error headings stable with loaded routes, constrained validation feedback to the form column, and preserved readable pending-action labels after independent review.
- Verified desktop and 390px Profile/Portfolio captures in dark French UI: both desktop route frames measured 1024px with 40px headings, the Profile form remained 720px, mobile navigation resolved to two equal columns, and neither viewport had page-level horizontal overflow. The user approved the rendered result.
- Verification passed: 63 focused Vitest cases; frontend typecheck, lint (two pre-existing Fast Refresh warnings, zero errors), and production build; `docker compose config --quiet`; a fresh isolated `./scripts/compose-test.sh` browser journey; `cd backend && golangci-lint run` with 0 issues; and `git diff --check`.

## Review Triage Log

| Finding | Verdict and evidence |
|---|---|
| Portfolio state headings differed from the loaded route heading. | low / patched — state and loaded headings now share the same size, weight, spacing, and line height. |
| Profile validation summary could span beyond the reading-width form. | low / patched — the summary now uses the same 45rem maximum as the form. |
| Pending-action labels used the lower-contrast disabled token. | medium / patched — progress-bearing and disabled authenticated actions now use the readable muted-text token. |
| The spec context did not list the account-inclusion implementation artifact. | false — the context list is optional and already names the authoritative design spine plus the two page implementations; no account-selection behavior or requirements changed. |
| The diff added no screenshot regression suite or new breakpoint tests. | low / rejected — existing real-browser coverage exercises the 767/768px transition and authenticated flows; a fresh clean-database journey, direct captures, metric checks, and human visual approval covered this CSS-only pass. Adding screenshot infrastructure is disproportionate here. |
| Authenticated styles still hard-coded mono/display font stacks after adding tokens. | low / patched — authenticated Profile and Portfolio rules now consume the shared font tokens. |
