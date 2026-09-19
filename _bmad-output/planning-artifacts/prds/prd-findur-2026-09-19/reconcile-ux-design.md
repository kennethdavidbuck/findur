---
title: Findur UX Visual Design Reconciliation
status: complete
created: 2026-09-19
source: ../../ux-designs/ux-findur-2026-09-19/DESIGN.md
against:
  - prd.md
  - addendum.md
---

# Reconciliation: UX Visual Design Against Final PRD

## Verdict

**Needs explicit product reconciliation before the design can be treated as normative.** The design strongly preserves the PRD's portfolio-first hierarchy, progressive Photo reveal, disclosure boundaries, freshness qualification, synthetic-data provenance, safety tone, and no-chat milestone boundary. It also introduces several product behaviors and scope commitments that the final PRD and addendum never adopted. The most important defect is document authority: the design says its spine wins over the written contract, while the final addendum explicitly makes itself subordinate to the PRD. The final PRD remains the product source of truth unless an explicit product change is approved.

This extraction does not modify `prd.md` or `addendum.md`. Recommended mappings preserve existing FR and NFR IDs wherever the design consequence fits an existing requirement.

## 1. Direct Conflict

### C-1 — Design claims authority over the final product contract (Critical)

**Design signal:** `DESIGN.md` §Brand & Style says that the Constellation composition reference and design spines win if the artifact and written contract conflict. The Interaction State Contract separately and appropriately establishes precedence only among visual artifacts.

**Final-source position:** `addendum.md` states that it is subordinate to `prd.md`, and that conflicts must be resolved in favor of the PRD or recorded as a new decision. The PRD also fixes the Disclosure Level field contract, milestone scope, live-versus-synthetic boundary, and public-launch boundary.

**Consequence:** The broad design-precedence sentence could silently change product scope, consent, disclosure, or safety decisions. In particular, the Candidate Card specification already varies the Disclosure Level contents beyond FR-6.

**Required resolution:** Keep the visual precedence rule only for conflicts among previews, tokens, and visual-state specifications. Product behavior, scope, data disclosure, safety, and NFR conflicts must defer to `prd.md`; any intentional change must be recorded as a product decision first. No new FR ID is needed.

## 2. New or Changed Product Decisions

### D-1 — Disclosure selection becomes an explicit preview/save transaction (High)

**Design decision:** The `Disclosure Level Control` starts with no level selected or saved; Snapshot is recommended but not preselected. Selecting a level changes only a preview. A separate Save action commits it, Cancel abandons it, and route exit with unsaved changes triggers a discard/stay confirmation.

**Current coverage:** FR-6 requires the user to choose and preview a level but does not define initial state, recommendation behavior, persistence, or unsaved-change handling. The addendum asks what the user previews before selecting but does not answer it.

**Product consequence:** This is consent-affecting behavior, not merely visual treatment. It changes when disclosure becomes effective and prevents accidental disclosure changes.

**Recommended placement if adopted:** Extend **FR-6** consequences with the no-default state, preview-versus-saved distinction, explicit save/cancel behavior, and unsaved-exit confirmation. Keep outlines, annotations, control geometry, and presentation language in the addendum/design.

### D-2 — Account-level inclusion controls are assumed without a defined permission model (High)

**Design decision:** `Account Inclusion Control` lists provider/accounts, selection and availability states, a “Using 2 of 4 connected accounts” summary, and Select all.

**Current coverage:** The PRD says the user authorizes a Connected Portfolio and that coverage may be incomplete, but it never requires an in-product account-selection control or says what exclusion changes. FR-1 covers consent, FR-8 covers derivation, and FR-19 covers the owner showcase; none defines whether an excluded account is unread, owner-only, omitted from matching, omitted from candidate disclosure, or deleted.

**Product consequence:** The component creates a new account-level consent and data-purpose boundary. Without semantics, owner views, matching inputs, disclosure previews, source-coverage claims, and disconnect/deletion behavior can disagree.

**Recommended placement if adopted:** Extend **FR-1**, **FR-8**, and **FR-19** rather than creating parallel FRs. Define what selection controls, when a change takes effect, how excluded accounts affect every derived/cached view, and how coverage is disclosed. Keep row styling and summary layout in the addendum. If account-level selection is not intended for this milestone, remove the control from the canonical component inventory.

### D-3 — Candidate-card field bundles drift beyond the fixed Disclosure Level contract (High)

**Design decision:** The Candidate Card adds activity recency and source coverage to Snapshot; instrument symbols and kinds to Holdings; and last-known prices, cost basis, and currency-specific exact values to Full Detail. It also requires performance to be absent or explicitly unavailable whenever period, inputs, currency treatment, coverage, or freshness cannot be defended.

**Current coverage:** FR-6 and addendum §1.2 fix the three field bundles. Some design additions may be explanatory context rather than disclosed fields, but last-known price and cost basis are material new disclosure categories. Conversely, FR-6 currently says Holdings “adds” percentage performance and Full Detail “adds” performance amounts without the design's defensibility exception.

**Product consequence:** Silent additions can expand financial-data exposure, while unconditional performance can create misleading output when source inputs are inadequate.

**Recommended placement:**

- Preserve the existing **FR-6** bundles unless product explicitly approves added field categories.
- Add the defensibility/explicit-unavailable condition to **FR-6**, and reinforce it in **FR-8**, **FR-11**, and **NFR-8**. This is a product truthfulness criterion, not merely visual guidance.
- Treat Proximity, source coverage, and activity recency as card context governed by FR-11 unless product intends them to be part of a Disclosure Level bundle.
- Keep exact hierarchy, visual placement, labels, and chart selection in the addendum/design.

### D-4 — Notification preferences and asynchronous activity are new milestone scope (High)

**Design decision:** `Notification Control` introduces separate Mutual Match, Incoming Interest, and connection-action-required preferences plus permission/support state and a contextual pre-prompt. `Discovery Activity Notice` introduces unacknowledged Incoming Interest or Mutual Match events with an Open action.

**Current coverage:** The PRD covers a Mutual Match produced immediately when the owner's Interested swipe meets a Synthetic Candidate's preconfigured positive decision. It has no notification permission model, delivery channel, asynchronous Incoming Interest journey, unread/acknowledgement state, or connection-action notification requirement. These behaviors are absent from MVP scope and success metrics.

**Product consequence:** This is more than a component. It adds event production, permission state, acknowledgement, privacy-safe content, lifecycle behavior, and possibly browser/OS notification integration. “Incoming Interest” also changes the current owner-led discovery journey unless it is only a seeded in-app demonstration event.

**Required resolution:** Either remove/defer both components, or approve a new functional requirement and add it to MVP scope with explicit in-app versus external delivery, synthetic-event origin, permission handling, acknowledgement, persistence, expiry, privacy, and test consequences. Existing **FR-14** can continue to own match creation, but it should not be stretched to silently authorize a notification subsystem.

### D-5 — English/French and System/Light/Dark parity are unapproved scope commitments (High)

**Design decision:** The document describes English and French, direct bilingual controls, localized date/time/number/percentage/distance/currency formatting, 35% label expansion, and System/Light/Dark selections. It says light, French, and theme parity are first-cut requirements rather than later polish.

**Current coverage:** The PRD requires a responsive hosted experience but names no supported languages, locale behavior, theme selection, or multi-theme acceptance. The addendum also does not adopt them.

**Product consequence:** Full localization and multi-theme parity materially expand copy, legal-content, formatting, QA, accessibility, and acceptance scope. French legal and trust content may also require qualified translation/review rather than UI-only localization.

**Recommended placement if adopted:** Add the support commitment to MVP scope and §9.2; extend **NFR-9** and **NFR-10** with functional parity across both languages and all supported themes, locale-aware values/units, wrapping without lost actions, and theme/locale control accessibility. Keep the 35% design allowance, exact typography, tokens, and theme palette in the addendum/design. If not adopted, label these as post-MVP design readiness rather than normative milestone requirements.

### D-6 — Installable/standalone web-app presentation is new platform scope (Medium)

**Design decision:** The design requires Apple touch and cross-platform web-app icons, favicon, standalone launch/task-switcher presentation, browser-chrome theming, and safe-area behavior in standalone display mode.

**Current coverage:** The PRD requires a responsive hosted experience and explicitly excludes native mobile distribution. It does not require installability, standalone display, a web-app manifest, or installed-mode QA.

**Product consequence:** Installable presentation is compatible with the hosted-web boundary, but it is an additional supported mode. The design does not require offline operation, so installability must not be interpreted as an offline-capable PWA requirement.

**Recommended placement if adopted:** Add installable/standalone web presentation to §6.1 and §9.2 and extend **NFR-10** for safe-area and information/action parity. Extend **NFR-2** only as needed to prohibit sensitive identity/financial content in launch, task-switcher, and icon assets. Exact asset sizes, manifest metadata, and browser-specific matrices belong only in the addendum/architecture handoff.

## 3. Missing Acceptance Criteria, NFRs, or Scope Definitions

### G-1 — Accessibility requirements are too broad to accept the design's normative claims (High)

**Design-derived acceptance needs:**

- WCAG 2.2 AA contrast: 4.5:1 normal text, 3:1 large text, and 3:1 meaningful non-text graphics, controls, and focus indicators.
- Meaning is not conveyed by color, motion, hover, node size, or layout alone.
- Every hover-only detail is available by keyboard focus and tap; passive surfaces are not added to the tab order.
- Reduced-motion presentation reaches the complete final state with the same text and actions.
- Controls preserve readable disabled states; required instructions and errors are not demoted to caption styling.
- Touch controls meet the design's stated minimums: 44px generally and 48px for primary swipe actions.
- At 200% zoom the layout reflows without loss; at 400% zoom it resolves to one dimension except an intrinsically wide data region, which also has a non-table summary.

**Current coverage:** NFR-9 through NFR-11 require keyboard operation, labels, focus, error text, non-color indicators, responsive parity, and accessible alternatives, but they omit measurable conformance, zoom/reflow, reduced motion, target sizes, and equivalent access to interactive visualization details.

**Recommended placement:** Strengthen existing **NFR-9**, **NFR-10**, and **NFR-11**; no new IDs are necessary. Keep exact focus-ring thickness, inset values, breakpoint numbers, border treatments, and hover fills in the addendum/design.

### G-2 — Candidate Detail is named but not functionally defined (Medium)

**Design signal:** Candidate Detail is a full-screen route on mobile, a two-column surface on desktop, and owns a persistent swipe-action region. The PRD mentions “candidate detail” only in FR-9's synthetic-label consequence; it is absent from §9.1's required surfaces and has no entry, exit, disclosure, or action acceptance criteria.

**Product consequence:** It is unclear whether Candidate Detail is required scope, optional card expansion, or a visual exploration. Its existence affects navigation, swipe persistence, disclosure enforcement, synthetic provenance, and responsive parity.

**Recommended placement:** If required, add Candidate Detail to §9.1 and extend **FR-11** and **FR-12** so opening/closing it does not change the candidate's Disclosure Level, synthetic status, Freshness State, or available swipe decision. Keep full-screen/two-column composition and persistent-action placement in the addendum/design. If it is not required, remove the unexplained reference from FR-9 and mark the component exploratory.

### G-3 — Three-destination App Navigation is a visual constraint without an information-architecture decision (Medium)

**Design signal:** The canonical App Navigation has exactly three equal mobile destinations and a desktop rail, but the destinations are not named. The PRD requires coherent access across many product surfaces and delegates final navigation/decomposition to UX.

**Product consequence:** Implementers cannot know which surfaces are top-level, which are nested, or how prerequisite-gated destinations behave. The Interaction State Contract additionally requires unavailable destinations to open a named gate rather than appear disabled.

**Recommended placement:** Record the route/destination map and gating behavior in the UX addendum or the companion experience specification. Promote it to §9.1 only if the exact three-destination structure is a fixed product requirement. FR-23's existing prerequisite consequence should remain the source for why Discovery is unavailable.

### G-4 — Destructive confirmation behavior is not tied to a named action (Low)

**Design signal:** A generic `Confirmation Dialog` defines destructive confirmation and safe initial focus. FR-4 requires post-disconnect confirmation of resulting data state but does not explicitly require pre-disconnect confirmation.

**Product consequence:** “Confirmation” can mean before or after deletion. Disconnect immediately deletes local source payloads and derived views, so the intended user checkpoint should be unambiguous.

**Recommended placement:** If a pre-disconnect checkpoint is intended, add it to **FR-4** while retaining the existing post-action data-state confirmation. Dialog styling and focus placement stay in the addendum/design.

## 4. Details That Belong Only in the Addendum/Design

The following are valuable handoff constraints but do not belong in the PRD unless product later makes them outcome-level commitments:

- Exact color tokens, contrast-pair calculations, typography families/sizes/weights, spacing tokens, radii, corner clipping, elevation, glow, grid, and shadow rules.
- Exact hover, focus, pressed, inset, underline, border-width, and marker styling after the outcome-level accessibility criteria are captured in NFR-9 through NFR-11.
- Exact 768px/1024px breakpoints, 224–256px navigation-rail width, mobile/desktop component composition, reading widths, and table-header behavior.
- Portfolio-node geometry, connector patterns, mask shape, deck-edge styling, chart/table visual treatment, and the particular constellation motion language.
- The precise visual layout of `Source-to-Experience Trace`, `Consent Panel`, `Recovery Panel`, `Revalidation Notice`, `Profile Preview Frame`, and `Mutual Match Reveal`; their required outcomes are already covered by FR-1, FR-3, FR-13, FR-15, FR-20, FR-21, FR-24, and NFR-8.
- Exact platform icon/manifest/metadata matrices, provided installable mode is first approved as product scope.
- Visual anti-patterns such as avoiding generic pills, rounded-card stacks, prestige badges, red/green performance theater, and luxury/trading-desk styling; these appropriately elaborate PRD §9.3 and addendum §2.6.

## 5. Confirmed Alignment — No PRD Change Needed

The design correctly carries forward these final decisions:

- Portfolio visualization remains more prominent than the obscured Photo.
- Photo reveal occurs only after the original reciprocal swipe and does not add a second acceptance step.
- The matched state provides Continue Discovery and no chat affordance.
- Synthetic provenance remains persistent across cards, detail, responsive layouts, and Mutual Match.
- Freshness and unavailable/unsupported states are explicit; the experience avoids “real time” claims and fabricated values.
- Source facts, private matching inputs, owner-only exact views, and candidate-visible disclosure remain distinct.
- Compatibility avoids scores, guarantees, visible cohorts, formulas, wealth judgments, and advice.
- Public/legal/trust routes, adult-only language, anti-solicitation guidance, and the demonstration boundary remain visible.
- Responsive variants preserve required information, consent, disclosure, freshness, and actions.

## 6. Recommended Reconciliation Order

1. Correct the design's authority statement so `prd.md` remains controlling.
2. Decide whether account-level inclusion, disclosure preview/save semantics, and notification/activity behavior are milestone requirements.
3. Decide whether English/French, multi-theme parity, and installable standalone presentation are in MVP scope or design-ready future scope.
4. Reconcile the Candidate Card field bundles, especially cost basis/last-known price and the performance-defensibility guardrail, against FR-6.
5. Strengthen NFR-9 through NFR-11 with measurable accessibility, zoom/reflow, reduced-motion, and interaction-equivalence criteria.
6. Define Candidate Detail and the three-destination navigation model in the UX handoff, promoting only fixed product outcomes into the PRD.
