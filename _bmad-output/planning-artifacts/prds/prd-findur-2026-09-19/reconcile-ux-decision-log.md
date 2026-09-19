---
title: UX Decision Log Reconciliation Against Final PRD
status: analysis
created: 2026-09-19
updated: 2026-09-19
---

# UX Decision Log Reconciliation

## Scope and precedence

This extraction compares:

- the UX run memory log at `ux-designs/ux-findur-2026-09-19/.memlog.md` (`updated: 2026-09-19T16:12`);
- the PRD run memory log (`updated: 2026-09-19T14:33`);
- the final `prd.md`; and
- the final `addendum.md`.

The UX decision log therefore postdates the final PRD. This is an audit artifact only: it does not apply any change to the PRD, addendum, or either memory log.

The classification test used here is:

- **PRD change** — changes first-cut scope, a user-visible capability, consent/privacy contract, acceptance boundary, or cross-cutting requirement.
- **Addendum change** — preserves a deeper UX/architecture behavior or future technical option without changing the product requirement itself.
- **Already represented** — the final PRD/addendum already carries the product meaning; the UX log only restates or operationalizes it.
- **Intentionally UX-only** — visual direction, component behavior, facilitation, or design-system detail belongs in `DESIGN.md`/`EXPERIENCE.md`, not the PRD set.

## Verdict

Reconciliation is required before the PRD can again be treated as the complete first-cut product authority. The UX log contains **9 PRD-level changes** and **9 addendum-level refinements**. It also contains **7 decisions already represented upstream** and **8 intentionally UX-only decisions**. Two early enhancement decisions—localization and themes—were later overridden in the same UX log; this report treats each progression as one final decision while preserving both source entries.

The most consequential deltas are account-level inclusion and staged consent, first-cut English/French and light/dark support, installable app-like mobile-web behavior, browser notification/badging scope, explicit disclosure saving and gating, and deferral of the Guest Demo.

## A. PRD changes

| UX source | Decision and reconciliation | Existing PRD relationship | Recommended PRD target |
| --- | --- | --- | --- |
| UX log line 18 | Set **WCAG 2.2 AA** as the accessibility target for load-bearing journeys and require equivalent non-visual/non-gesture operation. | NFR-9 through NFR-11 require keyboard operation, labels, non-color indicators, and accessible alternatives, but name no conformance target. This is an additive, testable quality requirement. | Strengthen §8.3, especially NFR-9/NFR-11; preserve proportional milestone depth without weakening the named target. |
| UX log lines 20 and 22 | The initial “design for easy localization” enhancement was overridden: **complete English and French support is required in the first cut**, including language selection, persistence, formatting, and text-expansion tolerance. | The PRD has no localization requirement, language selector, or locale acceptance boundary. The line-22 override is the final authority; line 20 must not be interpreted as current enhancement-only scope. | Add first-cut localization to §6.1 and a testable cross-cutting requirement in §8; include language settings among required surfaces where appropriate. |
| UX log lines 21 and 23 | The initial dual-theme enhancement was overridden: **light and dark themes are required in the first cut**, with system initialization, explicit selection, persistence, accessible contrast, and theme-safe charts/states. | The PRD does not require theming. Line 23 makes this first-cut scope rather than design polish. | Add dual-theme support to §6.1 and §8.3; keep token implementation detail in UX/architecture. |
| UX log line 26 | Add a post-OAuth **Findur account-inclusion step**: default none selected, require at least one for Discovery, support Select all and later edits, disclose selected/connected coverage, and keep inclusion distinct from OAuth access and Disclosure Level. Exclusions trigger recalculation and removal from caches/previews/stale results. | The PRD treats the Connected Portfolio as the consented unit and does not define Findur-level account inclusion. The addendum’s comprehension test says data “may cover only selected accounts,” but no selection capability or default exists. This materially refines consent, eligibility, source coverage, and deletion behavior. | Update the glossary, UJ-1/UJ-2, FR-1 through FR-4, FR-8, FR-19 through FR-21, §6.1, and relevant privacy success criteria. Preserve provider-level account-granular authorization as an architecture verification question rather than an assumed capability. |
| UX log line 33 | Require an **installable, app-like mobile web experience** in the first cut, including home-screen installation behavior and privacy-safe launch/share/offline treatment. | §9.2 requires only a responsive hosted experience; §6.2 excludes native distribution. A PWA remains web, so it does not violate the native-app non-goal, but it expands the platform acceptance boundary beyond responsiveness. | Amend §6.1 and §9.2 to require installable web behavior and add privacy/security outcomes to §8.1. Put exact manifests, icon matrices, safe-area mechanics, and platform assets in the addendum/architecture. |
| UX log line 37 | Make consent explicitly **two-stage**: pre-OAuth continuation authorizes named provider access/categories and only the minimum metadata needed for account selection; post-OAuth account-selection confirmation authorizes Findur use and starts retrieval/derivation. “Connected” never means “included.” | FR-1 requires explicit consent before derivation, but does not define the two grants or prohibit portfolio derivation between OAuth completion and account inclusion. This is a substantive refinement of FR-1/FR-2 and the Connected Portfolio model. | Update FR-1, FR-2, FR-8, glossary definitions, and UJ-1. Keep exact request sequencing and provider constraints in the addendum/architecture. |
| UX log line 38 | **Defer Guest Demo implementation from the first cut** while preserving an isolated synthetic-only extension seam and allowing the Public Site to reserve a future entry without advertising current availability. | FR-27 and §6 describe Guest Demo as a conditional first-cut stretch pending architecture sizing. The UX decision resolves that branch to “not in the first cut,” so leaving FR-27 unchanged would preserve stale optional scope. The isolation contract itself is already represented. | Move FR-27 out of first-cut/conditional scope into §6.3 or another future section; update UJ-6, metrics, open question 10, and addendum references accordingly. |
| UX log line 46 | Add first-cut **optional browser notifications and installed-app badging** for Mutual Match, separately configurable Incoming Interest, and connection-action-required events. Permission is contextual and user initiated; payloads, previews, badges, and URLs contain no candidate or financial detail; in-product event state remains authoritative. | No notification, push, badging, permission, or event-delivery requirement exists. This adds a user-facing capability and privacy/reliability obligations, even though enablement is optional for the user. | Add a functional requirement under Discovery/Match or lifecycle notifications, update §6.1, and add privacy/reliability acceptance criteria. Keep standards-based delivery mechanics in the addendum/architecture. |
| UX log line 47 | Disclosure Level **never autosaves**; starts with no saved tier; Snapshot is recommended but not preselected; Discovery is gated until explicit save succeeds. Preview drafts are visibly unsaved, cancel restores the saved state, dirty route exit confirms discard/stay, and upgrades/downgrades take effect only after explicit save and fail-closed processing. | FR-6 says the user can choose a level but does not define initial state, explicit commit, or draft semantics. FR-23 currently gates Discovery only on required profile fields and a Usable Portfolio, so it omits the newly required saved Disclosure Level. | Update FR-6, FR-23’s Discovery gate, UJ-2/UJ-7, and relevant acceptance/success criteria. Keep exact labels and dirty-form interaction design in the addendum/UX specification. |

## B. Addendum changes

| UX source | Decision and reconciliation | Why addendum, not PRD |
| --- | --- | --- |
| UX log line 16 | Define deep-link/direct-entry, authentication gating, OAuth return, refresh, back navigation, and safe fallback behavior for every applicable route. | The PRD already requires coherent navigation, authenticated OAuth return, and responsive flows. The complete navigation-state contract is downstream UX/architecture depth. Add it to the UX and authorization/session handoffs. |
| UX log line 17 | Use safe **stale-while-revalidate** behavior: show the last trustworthy view with visible freshness, revalidate nondisruptively, and transition honestly when thresholds/source state require it. | FR-3, FR-13, NFR-8/NFR-13, and addendum §3.3 already require honest freshness and recoverability. This selects a deeper experience behavior while leaving thresholds/mechanism to architecture. The addendum must state that consent, account exclusion, disclosure downgrade, disconnect, and other invalidations override stale-view reuse. |
| UX log line 24 | For an eligible configured owner, make Discovery the default authenticated landing; route incomplete/unusable states to focused recovery; preserve safe Discovery state during revalidation; confirm OAuth return before continuing; preserve permitted deep-link intent; keep Portfolio prominent but not mandatory home. | This refines §9.1 navigation and lifecycle behavior without adding a new capability. Add to UX Design Handoff and Authorization/Session Boundary. |
| UX log line 25 | Fix the authenticated IA at three primary areas—**Discovery, Portfolio, Profile**—with the stated child ownership; preserve the hierarchy across phone/desktop and keep public/legal surfaces outside the authenticated shell. | §9.1 lists required surfaces but explicitly delegates final navigation and decomposition to UX. This is the resulting handoff decision. |
| UX log line 31 | Use progressive portfolio depth: Candidate Card to a dedicated deep-linkable detail route; full-screen mobile and two-column desktop compositions; persistent Pass/Interested actions; exact deck-state restoration on Back; disclosure-limited sections; preview-only tier switching. | FR-11 requires rich cards and the addendum already permits card states/lightweight exploration. The dedicated route, responsive composition, and restoration model are detailed UX behavior rather than a new product purpose. Add to Candidate Card Information Hierarchy/Visualization Strategy and session-state handoff. |
| UX log line 34 | Realize FR-6 initiation asymmetry through an **Incoming Interest** child surface under Discovery, using a seeded initiator’s level; Pass returns to Discovery and Interested creates the existing Mutual Match; add neither a fourth primary area nor an inbox/chat. | FR-6 already requires the lower-disclosure recipient to evaluate incoming interest at the initiator’s level. This decision makes that existing requirement navigable and explicitly avoids new scope. Add to the disclosure and IA handoffs. |
| UX log line 36 | Make privacy-decreasing transitions fail closed: suppress broader use/display immediately, invalidate active/prefetched/history/offline outputs, retry failed purge/persistence without restoring broader state, and require scoped confirmation before any later broadening. | FR-4, FR-6, and addendum §3.2 already require invalidation for disconnect/downgrade. This strengthens failure semantics and extends them to account inclusion. Record the complete transition contract in Data and Derivation Boundary; reference the PRD-level account-inclusion change. |
| UX log line 42 | Refine the data-display contract: keep accounts, positions, balances, orders, and activities distinct; show dataset-specific freshness; never aggregate currencies silently; show performance only with defensible period/coverage/currency/freshness; keep optional tax lots owner-only and outside first-cut Candidate Cards; treat unsupported data as unavailable, not inferred. | FR-8, FR-19, FR-21, and NFR-8 already require source honesty and unavailable states, but not FX/performance/tax-lot rules. These are important data/visualization constraints best preserved in Visualization Strategy, Freshness and Events, and Data and Derivation Boundary. |
| UX log line 45 | Preserve SnapTrade’s hosted MCP server only as a **future protected-owner AI evidence-retrieval option**, not a first-cut change or assumed integration replacement, subject to provider/architecture review and Findur-owned review/disclosure/safety constraints. | This is explicitly future and architectural; it does not alter first-cut UX. Add it as a deferred architecture option with the listed verification gates, not as an MVP requirement. |

## C. Already represented

| UX source | Decision | Existing authority / disposition |
| --- | --- | --- |
| UX log line 8 | Identity must bridge dating and money. | PRD §1 and §9.3, plus addendum §2.4, already require a portfolio-first dating identity that is enjoyable yet trustworthy. No PRD-set change. |
| UX log line 9 | Tinder-like core interaction with visible profiles and expressive cards. | FR-11 through FR-15 already establish Candidate Cards, a Swipe Deck, one-stage left/right decisions, match, and reveal. “Tinder-like” remains non-normative inspiration. |
| UX log line 10 | Mobile web is primary; desktop remains supported. | NFR-10 and the PRD run decision at its log line 16 already require phone/desktop functional parity, while §9.2 establishes web delivery. Mobile-first prioritization is UX execution, not a requirement change. |
| UX log line 15 | `EXPERIENCE.md` must explicitly specify accessibility behavior. | NFR-9 through NFR-11 already make accessibility required. The instruction to spell it out in the UX artifact is operational. The separate WCAG 2.2 AA target at UX line 18 is new and is classified above as a PRD change. |
| UX log line 19 | UX defines behavior/outcomes; architecture selects implementation tools/mechanisms unless a UI-system choice is needed. | PRD §0, §9.1, and the existing addendum role already separate product/UX outcomes from technical mechanisms. |
| UX log line 35 | Candidate Card signal tiers, source/derived distinction, and bounded relevance explanation. | FR-6, FR-7, FR-8, FR-11, addendum §§1.2–1.3 and §§2.2–2.3 already carry the same field tiers and forbid scores, formulas, cohorts, guarantees, and worth judgments. The “one or two named traits” composition is an appropriate UX copy constraint and does not require a PRD-set edit. |
| UX log line 39 | Use only A-4 minimum Personal Profile fields in the first cut unless product later decides otherwise. | FR-23 and A-4 already define exactly that provisional minimum and require explicit UX/product reconsideration for additions. The UX assumption resolves the open question to “no additions” without changing the fields. |

## D. Intentionally UX-only

| UX source | Decision | Disposition |
| --- | --- | --- |
| UX log line 6 | The experience should feel simple at its heart. | Preserve as an experience principle in `DESIGN.md`/`EXPERIENCE.md`. It is too qualitative to become an independent PRD acceptance requirement. |
| UX log line 7 | The product should feel fun and alive through purposeful CSS-animated flourishes. | Preserve in visual/motion guidance, including reduced-motion treatment. CSS is an implementation choice and should not enter the PRD/addendum as a product requirement. |
| UX log line 12 | Use coaching mode and visual artifacts during UX facilitation. | Process-only assumption; no product artifact change. |
| UX log line 28 | Constellation is the selected visual-direction anchor; other explorations are non-authoritative. | Visual authority belongs in `DESIGN.md`; no PRD/addendum change. |
| UX log line 29 | Preserve Constellation’s atmospheric, dark-first, restrained-density, crisp-geometry character with limited rounded borders. | Visual-system direction belongs in `DESIGN.md`. “Dark-first” is compatible with, but does not replace, the first-cut light/dark requirement. |
| UX log line 32 | Let specialists make routine evidence-based UX defaults and ask the user only about material product/privacy/scope/brand decisions. | Facilitation/process decision only. |
| UX log line 43 | Hover and pressed states preserve geometry; use tonal/inset/underline/glow feedback; focus remains stronger; disabled controls do not react. | Normative component interaction detail for `DESIGN.md`, not PRD/addendum. |
| UX log line 44 | Shared component roles own interaction states; navigation uses one phone/desktop pattern; passive Portfolio panels must not resemble controls; written design spines outrank previews. | Design-system governance and affordance consistency belong in `DESIGN.md`/`EXPERIENCE.md`. |

## Conflict and dependency notes

1. **Guest Demo is the only direct first-cut scope reversal.** The PRD leaves it eligible as a conditional stretch; the later UX decision defers it. The PRD should not continue to imply architecture may opt it into the first cut unless that UX decision is itself reversed.
2. **Disclosure gating is incomplete in the current PRD.** FR-23 says Discovery opens once required profile fields and a Usable Portfolio exist. The later decision adds a successfully saved Disclosure Level and, through account inclusion, at least one confirmed included account.
3. **“Connected Portfolio” now hides multiple grants.** OAuth connection, minimum account metadata access, Findur account inclusion, portfolio derivation/use, and visible Disclosure Level are distinct. The PRD glossary and consent journey currently compress them too far.
4. **Stale-while-revalidate is subordinate to privacy invalidation.** A last trustworthy view may be shown only while it remains authorized. Account exclusion, disclosure downgrade, disconnect, and other privacy-decreasing events invalidate it immediately.
5. **Installable web is not native distribution.** The later PWA requirement can coexist with the native-app non-goal, but §9.2 should say so explicitly to avoid implementation ambiguity.
6. **Notification delivery is never the system of record.** Match/interest/action-required state must remain recoverable in Discovery even if permission is denied or delivery fails; notification payloads must remain generic.

## Recommended reconciliation order

1. Recast consent and portfolio terminology around OAuth connection, account inclusion, derivation/use, and disclosure.
2. Update first-cut scope for localization, themes, installable web, notifications, disclosure gating, accessibility target, and Guest Demo deferral.
3. Apply the detailed navigation, stale-data, privacy-transition, data-display, and future MCP handoffs to the addendum.
4. Re-run requirement-ID, journey, scope, success-metric, assumptions/open-question, and addendum-consistency checks after edits.
