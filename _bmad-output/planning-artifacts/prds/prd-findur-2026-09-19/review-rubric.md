# PRD Quality Review — Findur

## Overall verdict

This is a strategically coherent, unusually candid PRD for an owner-only, publicly hosted product demonstration. Its portfolio-first thesis, synthetic-data boundary, explicit non-goals, and measurable proof criteria make the intended milestone clear. It is not fully ready to green-light implementation because the public visitor's permitted path and the actual Disclosure Level contract—both central to the privacy promise—remain undecided; retention behavior is likewise an unresolved requirement masquerading as a selectable state.

## Decision-readiness — adequate

The milestone decision is unusually explicit: §1 calls it "a publicly hosted demonstration, not a public dating launch," §2.3 excludes "Real candidate users," and §10 makes the commercial, legal, safety, and operations gates for a public product unmistakable. Trade-offs are surfaced rather than softened: live data is restricted to the owner while Discovery uses a Synthetic Population (§§1, 4.5), and the addendum states that current research "does not establish permission for cross-user production use" (addendum §4.4).

The remaining decision problem is the path between a public visitor and the owner-only experience. §4.9 promises a public visitor "a clear entry action," while FR-2 limits authorization to "the owner-controlled test user" and says the demonstration does not support public signup. The document does not decide what that entry action permits or how the boundary is enforced.

### Findings

- **high** Define the public-entry and owner-access boundary (§§1, 2.3, 4.9; FR-2) — "publicly hosted demonstration," a visitor can "enter the demonstration," and "owner-controlled test user" are all true individually, but they do not specify whether an anonymous visitor reaches a landing-only experience, a gated viewer path, or an owner-authenticated session. That ambiguity affects the scope, security model, and test cases for a public deployment. *Fix:* Decide the permitted audience and entry states, state the access-control rule that prevents non-owner OAuth or live-data access, and add verifiable consequences to FR-2/FR-25.

## Substance over theater — strong

The vision is specific to the product: "a consented connected portfolio [is] the first candidate surface" and Photos remain obscured until mutual interest (§1). The differentiation earns its space through concrete implications in FR-1 through FR-21, not generic novelty language. The addendum also rejects status-tier framing directly (addendum §1.1), so the trust posture has operational consequences rather than serving as brand furniture.

NFRs are mostly product-specific: NFR-5 isolates live and synthetic data, NFR-8 prohibits fabricated currency or false freshness, and NFR-11 covers alternatives for obscured Photos and portfolio visuals. The text avoids a standalone persona section; Maya and Alex serve named, decision-bearing journeys.

## Strategic coherence — strong

The PRD has a clear thesis: financial/investing alignment should shape initial dating discovery without converting a person into a financial-worth score (§1). Its MVP logic follows that thesis: a real owner portfolio proves the integration, synthetic cards safely prove multi-candidate discovery, and Mutual Match proves the progressive identity reveal (§§4.2–4.7, §6.1).

Success Metrics test the thesis rather than merely activity: SM-2 requires a material input to alter eligibility, ordering, or comparison; SM-8 tests portfolio-first perception; SM-6 and SM-C1 through SM-C4 explicitly prevent apparent success from eroding privacy or truthfulness (§7). The PRD also correctly labels formative signals as not market-demand evidence (§7.4).

## Done-ness clarity — thin

Most FRs have concrete consequences: FR-3 prohibits presenting stale data as current, FR-10 tests maximum-distance exclusion and honest sparse results, and FR-15 makes Photo reveal testable. The target performance bound in NFR-12 and the scripted proof flow in SM-1 add useful acceptance anchors.

However, the primary disclosure feature is only testable after a future product decision. FR-6 says each Disclosure Level defines whether major categories are "hidden, bucketed, derived, or shown exactly," while the same requirement defers the "final names and field bundles" to UX and §11 question 1 asks which exact bundles should exist. This is the safety-critical product contract for the card and preview, not merely a visual-label decision.

### Findings

- **high** Make the prototype Disclosure Level contract decidable before implementation (§4.4, FR-6; §11.1; addendum §1.2) — FR-6 requires every level to govern allocation, holdings, values, performance, and activity, but the PRD leaves "which fields are hidden, bucketed, derived, or exact" open. As a result, FR-6, FR-11, FR-20, SM-5, and the no-overexposure part of SM-6 cannot be independently accepted. *Fix:* Add a product-owned, versioned disclosure matrix for the demonstration (including downgrade/cached-content behavior); UX may refine names and presentation afterward.

## Scope honesty — thin

The document does substantial honest scoping work. §5 and §6.2 explicitly exclude public candidates, chat, monetization, native apps, production operations, and dynamic market signals; §10 separates launch prerequisites from demonstration requirements. All five inline `[ASSUMPTION]` tags round-trip to §12, and §11 presents unresolved issues rather than implying they have been settled.

Question 5 nevertheless conflicts with a promised current behavior: FR-4 says the user is told whether derived data is "deleted immediately, queued for deletion, or retained for a stated reason and period," while §11.5 leaves the retention period open. Unlike a later public-launch decision, that choice determines what the owner is told in the demonstration and how disconnect can be accepted.

### Findings

- **medium** Resolve or explicitly constrain disconnect retention (§4.1, FR-4; §11.5; addendum §§3.2, 3.3) — "What retention period, if any" remains open even though FR-4 requires a stated post-disconnect outcome and the addendum requires a deletion behavior for each data transition. The question has no owner, decision date, or phase-blocker label. *Fix:* Choose the demonstration's source and derived-data deletion/retention rule, or remove the affected capability from the demo until that rule is approved; mark the remaining production policy as deferred with an owner and trigger.

## Downstream usability — strong

This is a chain-top PRD and is structured for extraction: the glossary defines the domain language, seven UJs have named protagonists, FRs have stable unique IDs, and SMs cite the FRs they validate. The addendum puts implementation guidance—such as OAuth/session details, synthetic-generator characteristics, and observability—where architecture and UX can consume it without confusing it for a product requirement.

The card, preview, freshness, synthetic-data, and match boundaries are consistently named across §§2–8 and the addendum. The order of FR headings is non-monotonic (FR-19–24 precede FR-5), but the ID set is complete and the cross-references checked here resolve.

## Shape fit — strong

The PRD fits a UX-heavy, multi-stakeholder dating concept while calibrating its actual milestone down to a one-owner demonstration. Named UJs carry enough context to guide UX, and the public visitor journey is separated from the connected owner and Synthetic Candidates (§2.4). The substantial safety, consent, freshness, and disclosure treatment matches the sensitive financial-data concern instead of forcing generic consumer-product structure.

The use of synthetic candidates is not a shape mismatch: it is the stated mechanism that lets the demonstration exercise a realistic Swipe Deck without pretending to be a public dating service (§§1, 4.5, 6.1).

## Mechanical notes

- Glossary usage is consistent for Connected Portfolio, Disclosure Level, Freshness State, Portfolio-Derived Signal, Synthetic Candidate, and Verified.
- FR IDs are unique and cover FR-1 through FR-26; headings are intentionally non-monotonic by feature grouping, which is a minor reading cost but not an ID-continuity defect.
- All five inline `[ASSUMPTION]` tags have matching entries A-1 through A-5 in §12, and the index contains no unmatched assumption.
- UJ-1 through UJ-7 each name a protagonist (Maya or Alex); their terminology and referenced product states resolve through the glossary.
