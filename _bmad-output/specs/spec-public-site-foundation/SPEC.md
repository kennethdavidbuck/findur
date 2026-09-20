---
id: SPEC-public-site-foundation
companions:
  - public-surface-contract.md
  - ../../planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md
  - ../../planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md
  - ../../planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Findur Public Site Foundation

## Why

Findur's deployed frontend is still a walking-skeleton status card, so an unauthenticated visitor cannot understand the product or its evaluation boundary. This slice establishes a credible bilingual public presence and the reusable presentation foundation that later public pages can extend, without waiting for authentication or other backend behavior.

## Capabilities

- **CAP-1**
  - **intent:** An unauthenticated visitor can understand Findur from a bilingual landing page.
  - **success:** At `/`, English and French presentations accurately communicate portfolio-first discovery, progressive identity reveal, the adult-only boundary, and that this is an evaluation demonstration rather than a public dating launch.

- **CAP-2**
  - **intent:** An unauthenticated visitor can learn the product purpose and current scope from an About page.
  - **success:** `/about` is directly addressable and explains the product purpose, responsible framing, and current non-public-launch scope in English and French.

- **CAP-3**
  - **intent:** Visitors can move through the public slice using a consistent shared shell.
  - **success:** `/` and `/about` use the same responsive Public Header and Public Footer, with working home, About, language, and theme controls and an honest unavailable state for owner entry.

- **CAP-4**
  - **intent:** Visitors can use the public slice in English or French without losing their place.
  - **success:** Changing language immediately replaces all visible and accessible interface copy, preserves the current route, persists the preference, and remains usable with either catalogue.

- **CAP-5**
  - **intent:** Visitors can view the public slice using their preferred color presentation.
  - **success:** System, Light, and Dark modes apply immediately and persist; System follows subsequent operating-system theme changes.

- **CAP-6**
  - **intent:** Later public surfaces can reuse a coherent visual and interaction foundation.
  - **success:** Landing and About are composed from shared semantic tokens and focused public layout, typography, action/link, navigation, language, and theme primitives with consistent default, hover, focus-visible, pressed, and unavailable states.

- **CAP-7**
  - **intent:** Maintainers can verify the public slice remains functional and accessible as it evolves.
  - **success:** Automated tests cover both routes, navigation, language switching, theme behavior, unavailable owner entry, and key accessible semantics; frontend lint, typecheck, tests, and production build all pass.

## Constraints

- The final Architecture Spine is authoritative. Implement in the existing React 19, Vite, and TypeScript frontend, using React Aria Components for interactive primitives as required by AD-16; use native semantic HTML where it is the accessible primitive.
- The slice is static and backend-independent. It must not implement authentication, call an API, render a protected shell, persist protected data, or simulate a successful owner sign-in.
- Follow the adopted Constellation direction and `DESIGN.md` semantic tokens, interaction geometry, responsive rules, and peer light/dark treatments; do not create a competing style system.
- Meet the applicable adopted WCAG 2.2 AA, keyboard, focus-visible, reduced-motion, touch-target, phone/desktop, and 200%/400% reflow requirements.
- Keep complete English and French catalogues in the frontend. Persist only locale and `system|light|dark` preference in non-sensitive browser storage, and tolerate at least 35% label expansion.
- Preserve AD-21: produce Vite static output with generic non-sensitive manifest and metadata, remain compatible with the configured SPA fallback, and add no service worker.
- Expose only `/` and `/about` in this slice. Omit unfinished public destinations rather than linking to empty, placeholder, or misleading routes.
- Build only the reusable primitives exercised by these pages; a comprehensive generic design-system package is outside this slice.

## Non-goals

- Functional owner authentication, OAuth, session handling, or authenticated routes.
- Privacy, Terms, Trust and safety, Contact/support, Guest Demo, or custom 404 content.
- Discovery, Portfolio, Profile, Candidate, or other protected product interfaces.
- Service-worker caching, offline behavior, analytics, forms, content management, or backend-managed localization.
- Final legal copy or claims that Findur is publicly available as a dating service.

## Success signal

From a fresh browser on phone and desktop, a visitor can open `/`, understand Findur's premise and demonstration boundary, move to `/about`, switch every surface between English and French and among System/Light/Dark modes, and encounter no backend request or dead public link. A developer can add the next public information page by composing the same shell, tokens, catalogue pattern, and focused primitives rather than duplicating page-specific infrastructure.

## Assumptions

- Until backend authentication exists, owner entry is visibly unavailable, performs no navigation or network request, and explains that owner access is coming later.
- With no stored preference, the public slice starts in English and uses System theme.
