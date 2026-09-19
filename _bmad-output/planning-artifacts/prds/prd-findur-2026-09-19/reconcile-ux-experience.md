---
title: Findur UX Experience Reconciliation
status: final
created: 2026-09-19
updated: 2026-09-19
source: ../../ux-designs/ux-findur-2026-09-19/EXPERIENCE.md
targets:
  - prd.md
  - addendum.md
---

# Reconciliation: UX Experience Spine Against the Final PRD

## Scope and verdict

This extraction compares the current UX behavior source, [`EXPERIENCE.md`](../../ux-designs/ux-findur-2026-09-19/EXPERIENCE.md), with the final [`prd.md`](prd.md) and [`addendum.md`](addendum.md). It records only product-requirement consequences: changed decisions, direct conflicts, missing or underspecified journey/state/gating/accessibility/responsive requirements, and qualitative intent at risk of being lost. It does not propose edits to either target document and excludes implementation choices unless they create a user-visible or acceptance consequence.

**Verdict: reconciliation required before architecture, epics, or acceptance planning.** The experience remains aligned with the PRD's central product model—one protected live-data owner, synthetic Discovery, portfolio-first cards, standardized disclosure, one-stage mutual interest, and no chat—but it makes several material decisions the final PRD does not authorize or test. The most consequential are a two-stage account-inclusion consent model, explicit unsaved/saved Disclosure Level gating, first-cut installability/English-French/light-dark scope, notification and unread-activity behavior, and an unconditional UX deferral of the PRD's conditional Guest Demo. It also turns broad accessibility and responsive statements into a much stronger acceptance floor that is currently absent from the final requirements.

## Reconciliation register

| ID | Severity | Classification | Requirement consequence | Existing requirement disposition |
|---|---|---|---|---|
| UX-R1 | Critical | New consent and gating decision | OAuth access, Findur account inclusion, and visible disclosure become three separate permission layers. OAuth retrieves only minimum masked account metadata; no portfolio retrieval or derivation starts until a post-OAuth, none-selected-by-default account set is explicitly confirmed. Zero included accounts gates Discovery. | Extend FR-1, FR-2, FR-3, FR-4, FR-19, FR-21, and FR-23; do not renumber. |
| UX-R2 | High | New functional scope | Notification preferences, contextual Web Push permission, privacy-safe push, app badges, and authoritative in-app Discovery activity are required by the experience but absent from the PRD and MVP scope. | Requires an explicit accept/defer decision. If accepted, add a new FR rather than overloading FR-14/FR-15; add privacy/reliability NFR consequences. |
| UX-R3 | High | Changed milestone scope | The experience makes installable/standalone mobile web, English and French parity, and System/Light/Dark themes first-cut requirements. The PRD requires only a responsive hosted experience and is silent on locale and theme. | Add cross-cutting NFRs if accepted; otherwise remove these claims from first-cut UX acceptance. Preserve NFR-9 through NFR-11. |
| UX-R4 | High | Direct scope conflict | PRD FR-27 and Open Question 10 retain Guest Demo as a conditional stretch pending architecture sizing. The experience says it is not advertised or implemented in the first cut and keeps only an extension seam. | Resolve FR-27 explicitly: retain conditional delivery or mark deferred. Do not leave acceptance dependent on contradictory scope statements. |
| UX-R5 | High | New disclosure and gating decision | Disclosure begins with no saved level, Snapshot is recommended but not preselected, previewing never saves, Save/Cancel is explicit, changed drafts require Discard/Stay, and Discovery remains gated until a level is saved. | Extend FR-6 and FR-13; keep FR IDs. |
| UX-R6 | High | Underspecified accessibility acceptance | WCAG 2.2 AA across EN/FR, light/dark, phone/desktop, and installed use; exact focus recovery; chart/table equivalence; reduced-motion parity; zoom/reflow; language-correct accessible output; and forced-colors behavior are not testable from PRD NFR-9 through NFR-11 as written. | Expand NFR-9 through NFR-11 rather than creating parallel accessibility requirements. |
| UX-R7 | High | New route/journey behavior | Candidate Detail becomes a dedicated progressive, deep-linkable route with disclosure-controlled sections, persistent Pass/Interested actions, protected route resolution, and exact safe restoration of deck position and undecided state. The PRD does not require this surface or restoration contract. | Extend FR-11 through FR-13 and §9.1; keep FR IDs. |
| UX-R8 | High | New freshness behavior | During safe sync/revalidation, the last trustworthy permitted Discovery view remains usable. Immaterial updates apply nondisruptively; material changes are announced and acknowledged before reordering the active task. Permission decreases override stale preservation. | Extend FR-3, FR-13, NFR-8, and NFR-13; keep IDs. Architecture still owns freshness thresholds. |
| UX-R9 | Medium | Missing demo journey/state coverage | FR-6 contains Incoming Interest asymmetry, but the PRD has no journey or acceptance path that proves it. The experience adds a seeded Discovery child surface at the initiator's level, with Pass or Interested and no increase to the owner's level. | Add FR-6 consequences and scripted acceptance coverage; no new FR is necessary unless Incoming Interest becomes broader than the seeded demonstration. |
| UX-R10 | Medium | Navigation and gate resolution | The authenticated shell is exactly Discovery, Portfolio, and Profile; ready owners default to Discovery; the most specific unmet prerequisite gate replaces the requested route; OAuth success must show an outcome and proceed to account selection. | Extend §9.1 plus FR-2, FR-13, and FR-23. The exact bottom-bar/left-rail composition can remain in UX. |
| UX-R11 | Medium | Privacy and recovery acceptance | Protected routes must resolve session and permission before rendering, retain only validated internal destinations, reject open redirects, clear sensitive rendered state on expiry, keep sensitive values out of URLs/social previews, and prevent protected content flashes. | Extend NFR-1, NFR-2, NFR-5, NFR-7, and §9.1; keep IDs. |
| UX-R12 | Medium | Permission-decrease contract | Account exclusion, disclosure downgrade, disconnect, and safety invalidation immediately suppress broader data across active, prefetched, rendered, history-restored, revalidation, offline, client-cache, and server-cache states; failures remain closed. Additions/upgrades retain the prior narrower committed state until success. | Consolidate as consequences of FR-4 and FR-6 and as privacy/reliability acceptance under NFR-2/NFR-8. Account inclusion adds consequences to FR-1/FR-19. |
| UX-R13 | Medium | Responsive and installed-use acceptance | Phone, tablet, desktop, standalone, and browser-tab contexts have explicit behavioral parity, safe-area and sticky-action rules, one focal card, full-screen versus two-column Candidate Detail, and no dependency on browser Back chrome. | Expand NFR-10 and §9.2 if installability is accepted. Breakpoints and rail/bar layout may remain UX-owned. |
| UX-R14 | Medium | Data-integrity and disclosure precision | Performance needs a named period, supported inputs, currency treatment, source coverage, and freshness; currencies stay separate absent an explicit FX policy; orders and activities keep distinct coverage/cadence; unknown, empty, unavailable, failed, and unsupported are distinct; date-only activity never gains an invented time. | Extend FR-3, FR-8, FR-11, FR-19, FR-20, FR-21, and NFR-8 without renumbering. |
| UX-R15 | Medium | Safety/comprehension decision | Before saving disclosure, the user sees that screenshots, memory, and combined-field inference cannot be revoked, with the strongest warning beside Full Detail; downgrade repeats the limit. | Extend FR-1 and FR-6 and add to the formative comprehension plan. This operationalizes addendum §4.2 rather than changing it. |
| UX-R16 | Low | Synthetic acceptance specificity | Fixed seed and generator version must reproduce stable identifiers and equivalent output, and the acceptance matrix explicitly covers every compatibility mode, disclosure level, distance boundary, portfolio pattern, freshness condition, sparse/unsupported state, and reciprocal outcome. | Add consequences to FR-22 and SM-9; keep FR-22. |

## Detailed findings

### 1. Consent, account inclusion, and readiness gates

`EXPERIENCE.md` → **Data Consent & Account Inclusion**, **Permission-change consequence matrix**, Flow 1, and Flow 8 establish a product model stronger than the PRD:

- Pre-OAuth Continue authorizes provider access and retrieval of minimum masked account metadata only.
- Post-OAuth account inclusion is a separate consent event. No usable account is selected by default.
- Account selection is a labelled multi-select with unavailable accounts and reasons, tri-state Select all, draft versus committed coverage, and a scoped confirmation of additions, exclusions, categories, purpose, and coverage.
- Financial retrieval, derivation, ranking, Profile Preview broadening, and Discovery use begin only after the included account set is confirmed.
- Discovery requires at least one included account that yields a Usable Portfolio.
- A failed addition retains the prior narrower set. A confirmed exclusion suppresses broader use immediately and can gate Discovery during purge/recalculation.
- A minimal masked connection inventory may remain while connected, but disconnect removes it under FR-4.

The PRD distinguishes private matching from visible disclosure and mentions selected accounts in comprehension material, but it never requires this account-inclusion control or post-OAuth gate. FR-1's “explicit action” is compatible with the UX decision but does not identify whether that action is pre-OAuth consent or post-OAuth account confirmation. FR-2's statement that valid authorization produces a Connected Portfolio is likewise ambiguous when authorization initially yields only masked metadata.

**Required resolution:** adopt the staged model as consequences under the existing FRs, or simplify the experience. Architecture must also verify whether provider authorization itself can be account-granular; Findur-level inclusion must not claim provider-side revocation it cannot perform.

### 2. Disclosure selection becomes an explicit commit workflow

`EXPERIENCE.md` → **Disclosure Level Control**, **State Patterns**, Flow 2, and the permission-change matrix add all of the following to FR-6:

- No level is initially selected or saved.
- Snapshot may be recommended, but cannot be preselected.
- Inspecting Snapshot, Holdings, or Full Detail changes preview only.
- Saved level and “Previewing — not saved” remain visibly distinct.
- Save and Cancel are explicit; route exit with a changed draft asks Discard changes or Stay.
- Discovery is gated until a Disclosure Level is successfully saved.
- Upgrade cannot broaden output until confirmation and persistence succeed; failure retains the lower level.
- Downgrade suppresses broader output immediately and remains fail-closed through persistence failure.

This does not conflict with FR-6's field contract; it closes an important ambiguity. It materially changes onboarding, readiness, persistence, failure handling, and acceptance and therefore cannot remain only in a component description.

### 3. Notifications and Discovery activity are an unapproved feature

`EXPERIENCE.md` → **Notification Control**, **Discovery Activity Notice**, **State Patterns**, Flow 3, Flow 4, and the final phase blocker introduce a coherent but substantial notification feature:

- A contextual pre-prompt appears after the first settled Interested decision.
- The direct Enable action may invoke browser permission; Not now does not penalize or repeatedly nag.
- The user has separate categories for Mutual Match, Incoming Interest, and connection action required.
- Privacy-safe push carries no candidate or financial detail and opens a non-sensitive route that rechecks session, eligibility, permission, safety, and event validity before rendering.
- An unacknowledged in-app event is authoritative even when push is denied, lost, or dismissed; an installed experience may add a badge.
- Activity is not a fourth navigation area, chat feature, or permanent match history.

None of this appears in the PRD's feature list, MVP scope, NFRs, or success metrics. It adds permission lifecycle, subscription cleanup, event persistence, authenticated click routing, badge reconciliation, capability detection, platform-specific install constraints, delivery failure, and telemetry concerns. This should be either explicitly deferred from the experience or accepted as a new FR with supporting NFRs and architecture sizing.

### 4. Guest Demo status conflicts

PRD FR-27 and Open Question 10 preserve Guest Demo as a conditional stretch: architecture sizes it and it may ship if it does not materially expand scope. `EXPERIENCE.md` → **Public shell**, **Authenticated shell**, **State Patterns**, and Flow 6 state a firmer decision: no Guest Demo route or action is advertised or implemented in the first cut; only an isolation extension seam remains.

The latter may be the better scope choice, but it is a product decision rather than a UX implementation detail. Acceptance cannot simultaneously treat FR-27 as conditionally deliverable and the experience as definitively deferred. Keep FR-27 and record the sizing outcome, or revise downstream UX to cover the conditional feature.

### 5. Installability, localization, and theme expand the milestone

`EXPERIENCE.md` → **Foundation**, **Localization & Theme**, **Responsive & Platform**, and **Accessibility Floor** make these first-cut requirements:

- installable responsive mobile web with standalone home-screen behavior;
- an opaque branded launch, persistent shell, safe areas, standalone navigation, and theme-aware browser chrome;
- English and French as peers, including accessible names, errors, chart summaries, and locale-aware dates, numbers, distances, percentages, and currencies;
- System/Light/Dark theme modes with persisted override and forced-colors/high-contrast integrity.

The PRD promises responsive hosted phone/desktop use and excludes native mobile distribution. An installable web experience is compatible with that boundary but is not implied by it. The PRD says nothing about bilingual delivery or multiple themes. These choices create real content, QA, accessibility, asset, session-restoration, cache, and release work. They require explicit milestone acceptance or explicit deferral.

### 6. Candidate Detail and exact recovery add a required product surface

`EXPERIENCE.md` → **Authenticated shell**, **Routing and gates**, **Candidate Detail**, **Deep Links & Navigation Recovery**, Flow 9, and Flow 10 specify a dedicated Candidate Detail route that the PRD does not list:

- It is progressive and shows only disclosure-permitted sections.
- It is a protected deep link, full-screen on phone and two-column on desktop.
- Pass and Interested remain available; opening detail does not commit a decision.
- Back restores exact deck position, scroll, candidate, prior decisions, and safe undecided view state.
- If permission, eligibility, connection, account inclusion, or freshness changed, invalidation wins over restoration and a focused Recovery Panel explains why.
- Direct entry first resolves session and gates without flashing protected content, then establishes a safe Discovery return context.

This feature is a natural elaboration of FR-11 through FR-13, but it affects journey length, routing, persistence, deep-link safety, responsive acceptance, and effort. It should be captured in those existing FRs and the required-surface list if it remains first cut.

### 7. Incoming Interest needs demonstrable coverage

PRD FR-6 already requires the key asymmetric rule: a lower-disclosure recipient may evaluate a higher-disclosure initiator without increasing the recipient's own level. `EXPERIENCE.md` → **Routing and gates**, **Incoming Interest** state, and Flow 12 is the first place that rule becomes exercisable:

- a seeded single-interest child surface is reached from Discovery;
- the initiator appears at the initiator's level;
- the owner receives an explicit assurance that their own lower disclosure remains unchanged;
- Pass returns to the exact safe deck state; Interested uses the existing Mutual Match because the initiator's positive decision already exists;
- invalidation fails closed and creates neither chat nor match history.

This does not need a new FR if it remains a seeded demonstration path, but FR-6 needs acceptance consequences and SM-1 or the scripted review plan needs a run that proves the asymmetry.

### 8. Safe stale-while-revalidate changes active-task behavior

`EXPERIENCE.md` → **Freshness/Revalidation** distinguishes safe revalidation from cold loading:

1. Show the last trustworthy, currently permitted view with source coverage and timestamp.
2. Revalidate in the background without a skeleton, focus movement, or deck reset.
3. Apply immaterial updates nondisruptively.
4. When a material update affects eligibility, order, values, or explanation, preserve reading position and require acknowledgement before rearranging the task.
5. At an architecture-defined unusable boundary, suppress new derived output and move to recovery.
6. Permission decreases always override stale preservation.

FR-3, FR-13, NFR-8, and NFR-13 support honest freshness and pending states but do not require continuity of the last trustworthy task or define material-change handling. The experience's version should be accepted explicitly because it controls what the owner may continue to act on while data is changing.

### 9. Accessibility floor is much stronger than the PRD

PRD NFR-9 through NFR-11 cover keyboard operation, labels, focus, error text, non-color indicators, responsive parity, and alternatives for obscured photos, visualizations, and swiping. `EXPERIENCE.md` → **Accessibility Floor**, **Interaction Primitives**, and Flow 11 add testable requirements currently at risk of being treated as optional design detail:

- WCAG 2.2 AA is the minimum across the complete EN/FR, light/dark, phone/desktop, and installed surface.
- Route changes, OAuth outcomes, direct entry, gates, invalidations, Candidate Detail, Back, and dialog cancellation have deterministic focus destinations.
- Background changes are announced politely; urgent errors and match creation are assertive only when needed; outcomes are announced once.
- Every portfolio visualization has a programmatic title/summary and an adjacent table or equivalent exact-data view, with periods, units, coverage, source/derived distinction, and freshness.
- Hidden Photo metadata, image descriptions, URLs, and cached previews cannot leak before Mutual Match.
- Non-gesture actions are always available; reduced motion reaches the same final state without loss of comprehension.
- Content reflows at 200%; at 400% it uses a single column where WCAG reflow applies, with only labelled data regions scrolling independently.
- Errors preserve entered values, identify the affected field/state, and provide recovery; disabled controls explain non-obvious prerequisites.
- Forced colors and high contrast retain selection, borders, labels, and focus.

These are acceptance consequences, not merely presentation preferences. They should refine the existing NFR IDs rather than fragment into a second accessibility contract.

### 10. Responsive contexts need a product-level parity decision

`EXPERIENCE.md` → **Responsive & Platform** defines phone, tablet, desktop, standalone, and normal browser contexts. The exact breakpoint values and bottom-bar/left-rail choice can remain UX-owned, but several consequences belong in NFR-10 and the platform statement:

- the same three authenticated areas remain available in every layout;
- Candidate Detail is full-screen on phone and may be two-column on desktop, while the active card remains one focal item rather than a grid;
- actions remain reachable above safe areas and browser chrome;
- wide tables scroll only inside a labelled region with a readable summary;
- standalone use cannot depend on browser Back chrome;
- route, session, gate, and task restoration behave the same after browser restart or home-screen launch;
- no sensitive state appears in task-switcher snapshots, authenticated HTML caches, social previews, or recent-content surfaces.

If installability is deferred, its standalone-specific acceptance should move with it; phone/desktop responsive parity remains required by PRD NFR-10.

### 11. Privacy-safe routing and metadata are missing acceptance detail

Across **Foundation**, **Routing and gates**, **State Patterns**, **Deep Links & Navigation Recovery**, and **Responsive & Platform**, the experience requires:

- no protected content underneath a session or prerequisite gate;
- validated internal intended destinations only, with no external/open redirect;
- sensitive rendered state cleared on session expiry while a safe destination may be retained;
- single-use, idempotent OAuth callback results;
- no financial, identity, or candidate details in URL parameters, public metadata, social previews, titles, badges, or authenticated snapshots;
- generic public sharing metadata and neutral titles for protected routes;
- public/static content may be cached, but protected financial/profile/candidate content cannot leak through offline or service-worker behavior.

These consequences refine the PRD's privacy, least-privilege, data-isolation, idempotency, and responsive requirements. They should remain attached to NFR-1, NFR-2, NFR-5, and NFR-7 rather than becoming an untraceable UX-only privacy checklist.

### 12. Data disclosure carries additional honesty constraints

`EXPERIENCE.md` → **Data Visualization & Disclosure** keeps the PRD's Disclosure Level field contract but adds needed rules for source integrity:

- performance requires supported inputs, a named period, explicit currency treatment, source coverage, and freshness, otherwise it is unavailable or unsupported;
- multiple currencies remain separate unless an explicit, disclosed FX policy is approved;
- orders and activities remain distinct because their cadence and coverage differ;
- date-only source activity never receives an invented time;
- source facts and Portfolio-Derived Signals are semantically distinguishable;
- account/instrument kind, currency, units, dataset period, account coverage, and freshness travel with the qualified claim;
- unknown, empty, unavailable, failed, and unsupported are separate states, and missing is never represented as zero.

These points are consistent with the PRD and addendum but more testable. They should be captured as consequences under existing data, portfolio, visualization, and reliability requirements.

## Qualitative UX intent at risk of silent loss

The final PRD and addendum retain much of the product tone, but the following experience-level intent is easy to lose when work is reduced to FR completion:

- **Uncertainty is a visible behavior, not just cautious copy.** Every error names what happened, what remains trustworthy, and the next safe action. Safe existing work remains available while background checks run.
- **Privacy controls must feel like boundaries, not status tiers.** Snapshot cannot become the “low” tier, Full Detail cannot become prestige, and a hidden Photo cannot be treated playfully as a prize.
- **Connected never means complete or included.** Coverage language such as “Using 2 of 4 connected accounts” is preferred to “all your finances,” and account inclusion remains distinct from provider access and candidate-visible disclosure.
- **Portfolio depth remains legible evidence.** The visual direction favors clean composition, sparse evidence geometry, and a single focal card; it rejects a generic verification badge, dashboard density, red/green performance theatre, and false-precision compatibility scores.
- **Dating energy must coexist with financial restraint.** The interface can feel enjoyable and recognizably romantic without luxury imagery, trading urgency, performance shaming, solicitation cues, or fintech-test-harness stiffness.
- **Localization must preserve meaning.** French is authored and reviewed as French, layouts allow expansion, and truncation cannot hide coverage, disclosure, freshness, consequences, actions, or errors.
- **Recovery is non-coercive.** Empty or sparse Discovery may invite a voluntary preference change, but never silently widens distance, disclosure, included accounts, or eligibility.

## Existing decisions adequately preserved; no requirement change needed

The experience is already consistent with the final PRD/addendum on these core points:

- one protected owner is the only live-data user, and every Discovery candidate is synthetic;
- exact live values remain in the owner's private Portfolio Showcase and Profile Preview;
- Snapshot, Holdings, and Full Detail preserve the FR-6 field contract and initiation asymmetry;
- portfolio visualizations lead Candidate Cards, with source/derived and freshness context and no worth score;
- one Pass/Interested decision produces a Mutual Match and immediate Photo reveal without a second appearance gate;
- no chat, public multi-user launch, browseable cohort, ranking formula, exact location, or production moderation claim enters the milestone;
- disconnection deletes local source/derived/display state, stops new portfolio-derived use, and requires fresh authorization;
- the generated Synthetic Population remains reproducible, varied, isolated from real records, and capable of boundary scenarios;
- public About, Terms, Privacy, trust and safety, contact/support, demonstration boundary, and adult-only positioning remain reachable without authentication.

## Product decisions required before downstream work

1. **Consent model:** confirm the two-stage metadata-then-account-inclusion model, including none-selected default and Discovery gating.
2. **Milestone scope:** accept or defer installability, English/French parity, System/Light/Dark themes, and the full notification/activity feature independently rather than as one bundled UX assumption.
3. **Guest Demo:** resolve the contradiction between conditional PRD scope and unconditional UX deferral.
4. **Disclosure persistence:** confirm no default level, explicit save/cancel, and fail-closed upgrade/downgrade behavior.
5. **Required surface:** confirm Candidate Detail as a first-cut protected route with deep-link and exact safe-return behavior.
6. **Acceptance floor:** adopt the detailed WCAG 2.2 AA and responsive parity consequences under existing NFR IDs.
7. **Architecture handoff:** resolve freshness materiality/thresholds, provider-versus-Findur account granularity, protected-route restoration, cache/snapshot exclusions, and—only if notifications are accepted—the complete Web Push and badge lifecycle.
