# Epic 2 Context: Transform SnapTrade Data into a Dating Profile

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Turn the authenticated owner's explicitly included, usable portfolio into a small explainable matching profile, combine it with the minimum adult dating profile, let the owner choose discovery preferences and a separate candidate-visible Disclosure Level, and provide an owner-private preview before Discovery can become available. The epic keeps source data, private derivation, and visible projection distinct and never places the live owner into another viewer's deck.

## Stories

- Story 2.1: Derive a Private Portfolio Matching Profile
- Story 2.2: Complete the Minimum Personal Profile
- Story 2.3: Set Discovery Distance and Compatibility Preferences
- Story 2.4: Preview and Save a Disclosure Level
- Story 2.5: Preview the Complete Live-Owner Profile and Readiness

## Requirements & Constraints

- A Personal Profile contains bounded display name, adult attestation, one application-owned coarse location, allowlisted relationship intent, bounded biography or prompt response, one bundled allowlisted avatar, locale, and theme. Adult attestation stores a timestamp, never a birth date. The first save is all-or-nothing; edits use optimistic concurrency.
- Coarse locations come only from a migration-seeded catalogue shared by live and generated profiles. Browsers receive a stable key and localized label, never centroids. Do not request device location, infer from IP, or accept coordinates, street addresses, or postal codes.
- Every first-cut Photo is a bundled placeholder avatar. Do not accept uploads, remote URLs, binary media, or create storage or image-processing infrastructure.
- English and French, System/Light/Dark, phone/desktop, keyboard, screen reader, touch, forced colors, reduced motion, 200% zoom, and applicable 400% reflow are first-cut peers. Layouts allow at least 35% French expansion. Locale-aware presentation must not duplicate stored profile content.
- Profile fields have persistent labels, required/private/pre-match/post-match visibility annotations, help and associated errors, and at least 44px targets. Failed validation preserves values, summarizes errors with links, and focuses deterministically. Save success is announced without stealing focus; conflicts preserve the draft and offer safe reload/reapply guidance.
- Discovery readiness is server-derived from a complete profile, saved preferences, explicitly saved disclosure, non-empty committed account inclusion, and usable included positions/signals. Authentication or client claims cannot mark readiness.
- Copy is direct, calm, and non-judgmental. Preserve explicit 18+ demonstration positioning and avoid wealth, luxury, trading urgency, financial shaming, deterministic compatibility, and personal-worth claims.

## Technical Decisions

- Keep profile use cases and rules in `internal/profile` within the Go modular monolith. HTTP and PostgreSQL remain adapters. Every owner-private input receives the authenticated immutable Actor; clients never provide or override the owner ID.
- The React SPA consumes only the authored OpenAPI 3.1 contract. Regenerate Go server/models and TypeScript types; never hand-edit generated artifacts. Owner-private responses use `Cache-Control: private, no-store`, and unsafe session-authenticated requests require the established CSRF, same-origin, and Fetch Metadata protections.
- Persist zero or one profile per OAuth user. Generated users use the same profile shape but cannot authenticate. Profile persistence belongs to the profile module and participates in the shared application UnitOfWork pattern where a mutation spans repositories.
- Persist authenticated locale as `en|fr` and theme as `system|light|dark`; copy current pre-login browser choices into the first profile save. Frontend catalogues own localized messages and `Intl` formatting.
- Use React Aria Components as the accessibility-first foundation where applicable. Shared primitives own focus, hover, pressed, disabled, field, action, and navigation behavior; route code must not redefine interaction states.

## UX & Interaction Patterns

- Profile is one of exactly three authenticated primary areas: Discovery, Portfolio, and Profile. `/profile` is the Personal Profile surface for this story and later grows to contain preferences, Disclosure Level, preview, language/theme, and settings. Do not introduce stored onboarding-step or progress state.
- Initial account selection still ends at the Portfolio Showcase. A later ordinary link can take an incomplete owner from that Showcase to `/profile`; this story must not modify the Showcase. Direct entry to `/profile` resolves session before protected content renders and focuses the route heading.
- Use a reading-width form on the Constellation visual system: crisp strong boundaries, restrained tonal layers, selective small corners, persistent field metadata, and one clear primary save action. Do not wrap every section in decorative cards.
- Profile states include empty/incomplete, editing, saving, saved, validation failure, network/save failure, and optimistic-concurrency conflict. Preserve the current task and entered values across recoverable failures.
- Interaction geometry never shifts on hover or press. Focus-visible uses the dedicated high-contrast ring. Disabled actions do not rely on opacity alone and explain prerequisites when needed.

## Cross-Story Dependencies

Story 2.1 supplies private portfolio signals but is not required to create or edit Personal Profile fields. Story 2.2 can therefore be implemented independently at `/profile`, while Discovery remains gated. Stories 2.3 and 2.4 add preferences and disclosure, and Story 2.5 combines those with the profile and usable portfolio into the complete preview and server-derived readiness result. Epic 2 consumes Epic 1's committed snapshots and freshness model; no Epic 2 story may weaken account inclusion or source-data boundaries.
