---
title: 'Normalize Owner Flow Typography and Spacing'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 1
context:
  - '_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md'
  - '_bmad-output/implementation-artifacts/spec-fix-account-inclusion-ux.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Owner access and authenticated application screens inherit the oversized public display heading and expansive generic page padding, making the connected flow feel discontinuous from the corrected account-selection step.

**Approach:** Apply the authored route-heading scale and a compact 4px-based spacing rhythm consistently across Owner access, account onboarding, and authenticated Discovery/Portfolio/Profile surfaces. Preserve the intentionally large public Landing and About marketing heroes.

</frozen-after-approval>

## Implementation Notes

- `frontend/src/styles.css` currently gives every `h1` a 38–76px display scale and `.section-pad` 72–136px vertical padding; `/connect` and `.private-placeholder` inherit both without a route-level override.
- `DESIGN.md` reserves display type for public proposition/orientation content and defines application route headings at 32px, with 16/24/40px responsive page margins and a 4px spacing rhythm.
- Scope the correction to `.consent-page`/`.consent-panel`, `.authenticated-content`/`.private-placeholder`, and `.portfolio-inventory`; do not reduce Landing/About hero presentation or alter page copy, routing, account behavior, navigation, or backend code.
- Verify computed heading sizes and vertical spacing at `/connect`, `/onboarding/accounts`, and a normal authenticated placeholder, plus confirm public hero sizing remains intentionally larger.
- Implemented the authored 32px/600 route-heading role on Owner access, account onboarding, and shared authenticated placeholders; the public global display treatment remains unchanged.
- Reduced Owner access and authenticated page padding to a 32–64px top rhythm, capped authenticated bottom padding at 80px, and aligned the desktop onboarding rail/content with a 32px nested offset.
- Review consolidated the route-heading declaration across all three surfaces, raised mobile setup margins to the authored 16px, applied 24/40px authenticated margins at wider breakpoints, and corrected the mobile consent cascade so generic section spacing cannot override it.
- Human visual review removed the remaining narrow heading caps, kept the progress rail's accessible name without visually adding “Connection setup,” and restored 64px of desktop setup clearance so header controls cannot cover the OAuth-return text around 1032px.

## Verification

- `cd frontend && npm test -- --run` — 58 tests passed.
- `cd frontend && npm run typecheck` — passed.
- `cd frontend && npm run lint` — 0 errors; 2 pre-existing Fast Refresh warnings.
- `cd frontend && npm run build` — passed.
- `cd backend && golangci-lint run` — `0 issues.` using the pinned v2.13.0 binary.
- `bash scripts/compose-test.sh` — `integration contracts passed`, including the 1032px non-overlap and visually hidden progress-label assertions.
- `git diff --check` — passed.

## Review Triage Log

| Finding | Verdict | Evidence / route |
|---|---|---|
| T1 — generic mobile `.section-pad` overrides the compact consent-page spacing | valid — patch | Increased selector specificity for `.consent-page.section-pad`. |
| T2 — authenticated pages keep 16px side margins at every width | valid — patch | Applied the authored responsive `--page-margin` token to authenticated content. |
| T3 — 34px desktop setup offsets break the 4px rhythm | valid — patch | Replaced 2.125rem offsets with 2rem and preserved grid alignment. |
| T4 — the first pass lacked concrete rendered-layout evidence | valid — verification | Added a real-browser geometry assertion at the reported 1032px width and visually inspected `/connect` at 1280px. |
| T5 — route-heading declarations can drift across three separate selectors | valid — patch | Consolidated the application heading role into one grouped declaration. |
| H1 — consent and placeholder headings remain cramped by a 20-character cap | valid — patch | Removed both route-level width caps; consent now uses the complete reading column. |
| H2 — visible “Connection setup” text was added beneath the wordmark despite not appearing in the UX design | valid — patch | Kept the heading only as a visually hidden accessible label for the rail. |
| H3 — EN/FR/theme controls cover OAuth-return text beginning around 1032px | valid — patch | Restored 64px desktop setup clearance and added a browser regression assertion that header bottom does not cross eyebrow top. |

No review finding was deferred.
