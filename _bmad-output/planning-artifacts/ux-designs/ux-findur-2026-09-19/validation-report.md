# Validation Report — Findur

- **DESIGN.md:** `DESIGN.md`
- **EXPERIENCE.md:** `EXPERIENCE.md`
- **Run at:** 2026-09-19T15:57:19-03:00

## Overall verdict

The draft pair has strong structure, complete token resolution, strong visual-reference discipline, and unusually thorough state and accessibility coverage. It is not yet safe as a downstream implementation contract because the source-required incoming-interest asymmetry has no committed surface or flow, FR-22 is mapped without an acceptance flow, Candidate Card signals remain unresolved, and several privacy-decreasing transitions can be interpreted as fail-open.

The combined accessibility and trust lens found no critical defects and judged accessibility close to strong. The required fixes are targeted: make privacy reductions immediate and fail-closed, define deterministic focus and consent transitions, correct two contrast mappings, and commit the missing load-bearing content and flows. This does not require expanding into speculative production moderation or identity systems.

## Category verdicts

- Flow coverage — **broken**
- Token completeness — **adequate**
- Component coverage — **thin**
- State coverage — **adequate**
- Visual reference coverage — **strong**
- Bloat & overspecification — **adequate**
- Inheritance discipline — **adequate**
- Shape fit — **strong**
- Accessibility — **adequate, close to strong**
- Trust, consent, privacy, and safety — **adequate, unsafe at privacy-decreasing transitions**

## Findings by severity

### Critical (1)

**[Flow coverage] — FR-6 incoming-interest asymmetry has no surface or Key Flow** (`EXPERIENCE.md` source coverage and Phase Blockers)

The lower-disclosure recipient must be able to evaluate a higher-disclosure initiator without increasing their own disclosure, but the contract leaves the entry surface and interaction undefined.

Fix: commit a named surface and named-protagonist flow, or explicitly change the milestone requirement.

### High (8)

**[Flow coverage] — FR-22 Synthetic Population generation is asserted but not demonstrated** (`EXPERIENCE.md` source coverage and Flows 3, 9–11)

The mapped flows consume candidates but do not cover reproducible generation, varied seeded scenarios, boundary cases, or regeneration.

Fix: add a reviewer/operator acceptance flow or a dedicated downstream acceptance contract.

**[Token completeness] — Synthetic Data Label misses normal-text contrast target** (`DESIGN.md` colors and `synthetic-data-label`)

Light-mode `{colors.secondary}` on `{colors.secondary-subtle}` is approximately 4.30:1, below the stated 4.5:1 target for the small label.

Fix: adjust the light token pairing and recheck every component using it.

**[Component coverage] — Candidate Card content contract is unresolved** (`DESIGN.md` Candidate Card; `EXPERIENCE.md` Data Visualization & Disclosure and Phase Blockers)

The approved Portfolio-Derived Signals and bounded “why this person appears” language are not committed, so stories cannot determine the data model or acceptance content.

Fix: define the initial signal set, source/derived distinction, disclosure eligibility, and explanation grammar.

**[Accessibility] — Account selection lacks complete accessible selection and progress semantics** (`EXPERIENCE.md` Account Inclusion Control and account-selection flow)

The contract does not define checkbox-group semantics, Select all mixed state, unavailable-account reasons, draft versus committed coverage, or concise save/recalculation announcements.

Fix: specify the labelled group, tri-state Select all, unavailable reasons, draft/committed counts, and non-disruptive progress announcements.

**[Accessibility] — Privacy invalidation can remove focused content without a deterministic destination** (`EXPERIENCE.md` routing, state, accessibility, and invalidation rules)

Downgrade, exclusion, disconnect, freshness, or safety invalidation may remove an active route or focused element while focus restoration remains vague.

Fix: focus the replacement Recovery Panel/status heading, announce why content disappeared, and expose the next safe action; cancel returns focus to the invoker.

**[Trust] — Pre-OAuth and post-OAuth actions form an ambiguous two-stage consent sequence** (`EXPERIENCE.md` Consent Panel, Data Consent & Account Inclusion, Flows 1 and 8)

The current wording could authorize blanket matching use or permit derivation before account inclusion is confirmed.

Fix: define pre-OAuth permission as provider access plus minimum account metadata retrieval only; post-OAuth confirmation authorizes Findur use for selected accounts and begins source retrieval/derivation.

**[Trust] — Account-exclusion failure can fail open** (`EXPERIENCE.md` account inclusion states and Flow 8)

One rule requires immediate purge, while the failure path retains the prior broader committed set.

Fix: suspend confirmed exclusions immediately, invalidate all outputs, gate Discovery while purge/recalculation retries, and never silently restore the broader set. Failed additions may retain the narrower prior set.

**[Trust] — Disclosure-downgrade failure can fail open** (`EXPERIENCE.md` disclosure rules, revalidation, and Flow 2)

The failure language could leave higher disclosure active until a retry succeeds.

Fix: confirmed downgrades suppress higher-level output immediately and retry persistence in the background; only a separately confirmed upgrade may broaden visibility.

### Medium (14)

**[Flow coverage] — FR-20 trace mapping points to the wrong flows.** Map phone/desktop and pre-/post-match Profile Preview to Flow 7; remove Flow 8 and retain Flow 1 only for its disclosure slice.

**[Token completeness] — Candidate Card dark border uses the weak border token.** Map it to `{colors.border-strong-dark}` or revise the stated strong-boundary rule and verify graphical contrast.

**[Component coverage] — Swipe Deck lacks a canonical component contract.** Add it to both component inventories and tokens, or explicitly define it as a Discovery composition with consolidated rules.

**[State coverage] — Conditional Guest Demo lacks conditional states.** If retained in IA, define entry, cold load, reset/expiry, failure, isolation denial, and return-to-Public-Site behavior; otherwise remove it after formal deferral.

**[Inheritance discipline] — Source precedence does not cover research conflicts.** State that the final PRD and its explicit reconciliations govern all earlier brief, addendum, and research conflicts; research is advisory where the PRD is silent.

**[Inheritance discipline] — Two glossary mappings drift.** Name `Portfolio (contains the Portfolio Showcase)` and use `Usable Portfolio` exactly for the glossary condition.

**[Accessibility] — Disclosure selection lacks saved-versus-preview semantics.** Use a labelled single-select group, separate inspected preview from saved value, describe each scope, and announce only concise preview and saved status.

**[Accessibility] — Revealed Photo lacks an explicit accessible equivalent.** Replace the obscured object with a real image named concisely, such as “Photo of Alex,” or declare adjacent identity text equivalent; never infer appearance traits.

**[Accessibility] — Weak border tokens conflict with the 3:1 non-text contrast promise.** Reserve weak borders for nonessential separators and use strong tokens for required boundaries, controls, and chart marks; document verified pairings.

**[Accessibility] — OAuth results and route gates lack initial-focus rules.** On full route transitions focus the updated result/route heading; on in-place replacement focus the gate heading; avoid duplicate live announcements.

**[Trust] — Source-to-Experience Trace collapses permission branches.** Branch included source data into owner-only exact view, private matching input, and candidate-visible disclosure, with provenance and blocked/unsupported stages.

**[Trust] — Exclusion purge conflicts with re-selectable connected-account inventory.** Retain only minimal masked connection metadata needed for selection status; purge values, holdings, activity, and derived signals.

**[Trust] — Screenshot and inference limitations appear too late.** Put concise limits in Disclosure Level preview before saving, with stronger proximity for Full Detail; retain the downgrade warning as a reminder.

**[Trust] — Inclusion expansion lacks explicit scope confirmation.** Treat edits as a draft set and require one scoped Apply/Confirm summary before newly included accounts begin retrieval and derivation.

### Low (2)

**[Component coverage] — Public Footer is unnamed.** Add a paired visual/behavioral component or explicitly make it a defined region of the Public Site contract.

**[Bloat] — Invalidation rules are repeated with drifting object lists.** Create one consequence matrix for disconnect, exclusion, downgrade, and safety invalidation; have flows reference it.

## Reviewer files

- `review-rubric.md`
- `review-accessibility-trust.md`

## Mechanical notes

- Both frontmatters parse and remain `status: draft`.
- All six source paths and the selected Constellation artifact resolve.
- All token references resolve and all light colors have dark peers.
- The explicit 26-component inventories have one-to-one parity.
- Freshness vocabulary needs an explicit state-to-token mapping: experience states are connected, syncing, stale, needs reauthorization, failed, and disconnected; visual tokens are current, pending, stale, and unusable.
- The Constellation artifact is illustrative only; its prototype-sized text, decorative low-contrast lines, and span-based controls are not component semantics.
