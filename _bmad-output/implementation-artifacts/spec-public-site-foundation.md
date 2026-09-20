---
title: 'Build the bilingual public site foundation'
type: 'feature'
created: '2026-09-19'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-public-site-foundation/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-public-site-foundation/public-surface-contract.md'
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur's frontend is still a health-check card, so an unauthenticated visitor cannot understand the product, its demonstration boundary, or its visual identity.

**Approach:** Replace the status shell with responsive Landing and About routes composed from reusable public-layout primitives, complete English/French catalogues, persistent System/Light/Dark preferences, React Aria interactive controls, and the architecture's Constellation design and privacy constraints.

</frozen-after-approval>

## Implementation Notes

- Replaced the health-check shell with static `/` and `/about` presentation; the browser makes no application API request.
- Kept routing dependency-free with semantic anchors, History API navigation, `popstate` restoration, localized metadata, and route-heading focus.
- Added typed English/French catalogues and validated `findur-locale` persistence; added System/Light/Dark handling through `findur-theme`, live `prefers-color-scheme`, and early theme initialization in `index.html`.
- Used React Aria `RadioGroup`/`Radio` for preference selection and `Button` for the unavailable owner action; ordinary navigation and document structure remain semantic HTML.
- Built the reusable public shell, focused action/link/layout styles, exact semantic theme tokens, Constellation hero, responsive reflow, reduced-motion behavior, and generic manifest metadata.
- Following live user feedback during implementation, language and theme controls appear only in the shared footer; the header remains focused on brand, public routes, and owner entry.
- Following visual review feedback, public-route hover uses the same underline language as the active route—2px on hover and 3px when current—with no competing tonal block or inset marker.
- Following footer review feedback, repeated Home/About links were removed; the single header navigation owns routes, while the footer stays focused on the brand boundary and persistent display preferences.
- Replaced health-check tests with public-site coverage for static/no-fetch behavior, direct and client routing, unknown-path normalization, focus/metadata, bilingual route persistence, keyboard selection, explicit theme restoration, and live System-theme changes.
- Browser inspection covered the Landing and About pages in light/dark and English/French presentations with no console or runtime errors; the footer-only preference controls keep the header visually focused.
- Blind review fixes made the owner-entry reason visible, canonicalized unknown paths, differentiated navigation landmarks, added visible route-heading focus, synchronized pre-paint locale/theme/browser chrome, aligned semantic tokens, replaced the old favicon with a constellation mark, and expanded route/preference/keyboard tests.
- Pre-commit public-file scanning covers machine-specific home paths and local usernames; the scan also corrected an older absolute path in the deferred-work ledger.

## Review Triage Log

- `medium / defer` — The earlier canonical public-surface contract still places preferences in the header; the user's later footer-only direction is implemented and recorded here, while re-derivation through bmad-spec is recorded in `deferred-work.md` because that skill is the canonical spec's sole writer.
- `medium / patch` — The unavailable owner-entry reason was visually hidden and its disabled button was not keyboard-focusable; the reason is now persistent visible text associated with the control.
- `medium / patch` — Unknown paths rendered Home without changing the URL; route initialization and `popstate` now replace unknown paths with `/`, with regression coverage.
- `low / patch` — Header and footer navigation shared one accessible name; each landmark now has a localized distinct name.
- `low / patch` — The focused route heading suppressed its outline; `:focus-visible` now uses the required warning-color ring.
- `low / patch` — Pre-paint dark resolution left browser chrome light; early initialization now updates `color-scheme` and `theme-color` before React loads.
- `low / patch` — The storage-error branch forced light mode; it now respects the system preference and records System as the preference.
- `low / patch` — Stored French could render before the document language changed; early initialization now sets `html.lang` before React loads.
- `low / patch` — Several semantic token values and token layers diverged from `DESIGN.md`; exact subtle/warning values plus spacing, corner, margin, and safe-area tokens are now present.
- `medium / patch` — The old yellow letter favicon conflicted with Constellation; it is replaced by a non-sensitive geometric constellation mark in the adopted palette.
- `medium / patch` — Verification under-covered bilingual routes and preference behavior; tests now cover French About and landmark copy, route preservation, keyboard locale selection, stored Light override, and unknown-path normalization alongside the existing browser/theme checks.
