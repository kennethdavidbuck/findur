# Spine Pair Review — findur

## Overall verdict

**Broken as a downstream contract, despite strong structure and unusually thorough state/accessibility work.** The pair resolves its sources, tokens, canonical sections, and explicit component inventory cleanly, but it leaves the source-required incoming-interest asymmetry without a surface or flow, asserts rather than realizes FR-22, and defers two UX-owned Candidate Card decisions that story development needs. The remaining defects are bounded contrast, parity, component-inventory, conditional-state, and inheritance issues rather than wholesale design-system failures.

## 1. Flow coverage — broken

Checked all seven source UJ names and all 27 FR IDs/titles against the PRD and the Key Flows section. UJ-1 through UJ-7 are named verbatim and each has a named protagonist, numbered steps, a climax, and a failure path; every FR name appears verbatim in the coverage table.

### Findings

- **critical** FR-6's required incoming-interest asymmetry has no entry surface, Key Flow, or interaction contract. The spine itself calls this a phase blocker, so a story author cannot implement the lower-disclosure recipient evaluating a higher-disclosure initiator without inventing product behavior. (`EXPERIENCE.md` lines 228–236 and 366; source `prd.md` lines 227–230). *Fix:* commit a named surface and protagonist flow for the inbound-interest case, or record an explicit product change removing that FR-6 consequence from this milestone.
- **high** FR-22 “Generate a varied Synthetic Population” is listed as covered by Flows 3, 9, 10, and 11, but those flows only consume individual synthetic candidates; none covers reproducible generation, population variation, seeded boundary scenarios, or regeneration. (`EXPERIENCE.md` lines 232, 261–271, and 327–358; source `prd.md` lines 278–287). *Fix:* add a named reviewer/operator flow that demonstrates population generation and the required scenario coverage, or remove the unsupported flow mapping and provide an explicit downstream acceptance contract elsewhere.
- **medium** The trace table maps FR-20 only to Flows 1 and 8, although its phone/desktop and pre-/post-match preview behavior is actually in Flow 7; Flow 8 is account inclusion and does not cover owner-card preview. (`EXPERIENCE.md` lines 230, 305–325). *Fix:* map FR-20 to Flow 7 (and Flow 1 only for its disclosure-level slice) and remove Flow 8 from that mapping.

## 2. Token completeness — adequate

Extracted every frontmatter token and every `{path.to.token}` reference from both spines. All references resolve; all colors are six-digit hex strings; every light color has an exact `-dark` peer; typography, radius, spacing, and component token types conform to the working DESIGN.md spec.

### Findings

- **high** `Synthetic Data Label` uses `{colors.secondary}` on `{colors.secondary-subtle}` in light mode, which computes to about **4.30:1**, below the spine's stated 4.5:1 target for normal text. The selected visual direction renders this label as small text, so the large-text exception does not apply. (`DESIGN.md` lines 27–29, 216–222, 380, and 423). *Fix:* adjust one of the light secondary tokens until this component pairing reaches at least 4.5:1, then recheck every component using the pair.
- **medium** `candidate-card.border` uses `{colors.border-strong}` in light mode but `candidate-card.border-dark` uses the weaker `{colors.border-dark}`; the body calls for a “1px strong border.” This breaks theme parity, and the dark pair is only about 2.31:1 against `{colors.surface-dark}`. (`DESIGN.md` lines 176–183 and 418). *Fix:* point the dark mapping to `{colors.border-strong-dark}` or revise the body and verify that the resulting boundary still meets the intended graphical contrast.

## 3. Component coverage — thin

Compared the 26 frontmatter component objects with the 26 DESIGN.md visual rows and 26 EXPERIENCE.md behavioral rows. Those three explicit inventories have exact one-to-one parity and substantive rules.

### Findings

- **high** `Candidate Card` is present in both inventories, but its load-bearing content remains uncommitted: the concrete Portfolio-Derived Signals and bounded “why this person appears” language are still Product + UX phase blockers. Architecture and stories cannot determine the card's data model or acceptance content from the current contract. (`DESIGN.md` lines 418 and 447–455; `EXPERIENCE.md` lines 88, 184–192, and 365). *Fix:* name the approved signal set and explanation contract, including source/derived distinctions and disclosure eligibility, before treating the component as story-ready.
- **medium** `Swipe Deck` is a source-defined, repeatedly used UI unit with routing, restoration, gesture, and state behavior, but it has no row in either canonical component table and no frontmatter component tokens. (`EXPERIENCE.md` lines 47, 60, 134–139, and 261–271; component tables at `DESIGN.md` lines 414–441 and `EXPERIENCE.md` lines 84–111). *Fix:* add `Swipe Deck` to both component tables and the component token map, or explicitly classify it as a Discovery-surface composition and consolidate its visual/behavioral rules there.
- **low** The public footer is required to keep legal/support routes available but is not part of either component inventory. (`EXPERIENCE.md` line 41). *Fix:* add a `Public Footer` visual/behavioral pair or define it as an explicit region of `Public Header`/Public Site rather than leaving it unnamed.

## 4. State coverage — adequate

Walked every IA surface and child route against the State Patterns table and global focus/offline/error rules. Public information, owner entry/session, consent/OAuth, account selection, Portfolio, Discovery, Candidate Detail, Profile/preferences, Profile Preview, Mutual Match, and installed launch all have meaningful cold/load, error, invalidation, offline, or recovery coverage where applicable.

### Findings

- **medium** `Guest Demo` appears as a conditional IA surface, but there is no conditional state row for entry, cold load, reset/expiry, failure, owner-isolation denial, or return to Public Site if the branch ships. The only flow treatment covers the branch when it is deferred. (`EXPERIENCE.md` lines 39, 51, 115–127, and 295–303). *Fix:* add a conditional Guest Demo state contract gated by the architecture decision, or remove the route from IA once the branch is formally deferred.

## 5. Visual reference coverage — strong

`mockups/` and `wireframes/` are absent and `imports/` is empty, so there are no required files to orphan. The selected `.working/direction-constellation.html` is linked inline from Brand & Style, its contribution is named, and the pair states once that the spines win on conflict; the memlog explicitly marks the other three `.working/` directions as unselected.

### Findings

- None.

## 6. Bloat & overspecification — adequate

Most length is earned by the product's sensitive-data, responsive, localization, and accessibility contract. The visual spine uses the spec's token-plus-rationale pattern, and the behavioral spine generally keeps implementation mechanism with architecture.

### Findings

- **low** The same purge/invalidation rule is restated across routing, global state, account inclusion, disclosure, freshness, responsive/privacy, and several flows, with slightly different object lists. That repetition increases drift risk for story extraction. (`EXPERIENCE.md` lines 61, 129, 166–180, 192, 202, 214, 259, 293, 323, and 347). *Fix:* define one consequence matrix by trigger (disconnect, exclusion, downgrade, safety invalidation) and have flows reference it.

## 7. Inheritance discipline — adequate

Both frontmatter source arrays are identical and all six relative paths resolve. All UJ and FR names are verbatim; every EXPERIENCE.md token reference resolves to DESIGN.md; and component names are identical across the explicit component sections.

### Findings

- **medium** The precedence rule says only that the final PRD wins over “subordinate addenda,” but the source set also contains research whose earlier recommendations conflict with final PRD permission for rich synthetic Full Detail cards. A downstream extractor reading all listed sources is not told unambiguously that the final PRD's recorded reconciliation wins those research-era constraints. (`EXPERIENCE.md` line 22; frontmatter lines 4–10; source `prd.md` lines 224–231 and 580–583; `source-extract-research-context.md` lines 9–12). *Fix:* state that the final PRD and its explicit reconciliations govern all earlier brief/addendum/research conflicts, while research remains advisory where the PRD is silent.
- **medium** Two source glossary anchors are not inherited cleanly: the source-defined `Portfolio Showcase` is replaced by the route label `Portfolio` without an explicit mapping, and `Usable Portfolio` is rephrased as “usable portfolio authorization.” This weakens FR-to-surface extraction even though the intended meaning is inferable. (`EXPERIENCE.md` lines 20 and 45–49; source `prd.md` lines 77 and 85). *Fix:* state `Portfolio (contains the Portfolio Showcase)` and use `Usable Portfolio` exactly when referring to the glossary condition.

## 8. Shape fit — strong

DESIGN.md contains every canonical section in the locked order. EXPERIENCE.md contains all eight defaults, both triggered sections (Responsive & Platform; Inspiration & Anti-patterns), and product-specific sections that earn their place through first-cut localization/theme, deep-link recovery, consent/account inclusion, disclosure, and freshness requirements.

### Findings

- None.

## Mechanical notes

- Frontmatter parses in both files; both remain `status: draft`, consistent with the unresolved critical/high findings but not ready for final downstream handoff.
- All six source paths resolve from both spines, and both source lists are identical.
- All 47 distinct `{path.to.token}` references resolve. Every color has a six-digit hex value and a light/dark peer.
- The 26 component names have exact DESIGN.md/EXPERIENCE.md table parity; their kebab-case frontmatter keys map consistently.
- The only relative Markdown link in the pair, `.working/direction-constellation.html`, resolves.
- No Mermaid blocks are present, so there is no Mermaid syntax to validate.
- Finding counts: **critical 1 · high 3 · medium 6 · low 2**.
