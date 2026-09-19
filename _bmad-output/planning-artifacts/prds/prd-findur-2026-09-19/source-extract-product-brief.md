---
title: "Findur PRD source extraction: product brief"
status: complete
created: 2026-09-19
purpose: "Provenance-preserving extraction for PRD creation; not a PRD rewrite."
sources:
  - "briefs/brief-findur-2026-09-19/brief.md"
  - "briefs/brief-findur-2026-09-19/addendum.md"
  - "research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md (only where it qualifies a brief decision)"
---

# Findur — Source Extraction for PRD Creation

## Vision, problem, and product promise

- Findur is a portfolio-first dating experience for everyday investors who want financial and investing alignment to materially inform romantic discovery. It reverses conventional photo-led discovery: a connected portfolio is the first candidate surface, while photos are progressively revealed after mutual interest. [Brief: Executive Summary; Vision]
- The problem is both evidentiary and experiential: conventional dating provides little trustworthy evidence of financial habits, stability, openness, or investing behaviour, and photos dominate the initial filter. A prompt or verification badge is insufficient because portfolio data must change the discovery market itself. [Brief: The Problem]
- The promise is trusted, portfolio-informed matching: consented data shapes candidate eligibility, ordering, comparison, disclosure rights, and timely relevance. "Verified" means derived from accounts the user chose to connect at a stated freshness point—not complete net worth, every account, or an assessment of financial health. [Brief: Product Promise]
- Findur is not investment advice, a trading product, or a certification of financial responsibility/personal worth. [Brief: Who This Serves; Scope and Roadmap—Explicit boundaries]

## Target users, stakes, and form factor

- Primary user: an everyday investor who values evidence and financial openness, wants investing alignment in dating, will connect a real portfolio for a more relevant/exclusive experience, and expects control over what others can see. [Brief: Who This Serves]
- The immediate artifact is a publicly hosted, responsive demonstration—not a public launch—using SnapTrade's test OAuth environment, one real private owner portfolio, and synthetic candidate profiles/signals. No production credentials or real candidate financial data. [Brief: Scope and Roadmap—Core implementation proof; Explicit boundaries]
- Product stakes are high around financial privacy, consent, safety, and trust even in the demonstration. The technical research says test OAuth's five-user limit is adequate for this private prototype; any public/cross-user use of portfolio-derived attributes remains unresolved pending written SnapTrade confirmation. [Technical research: Executive summary; Production-grade build constraints and future-product boundary]
- Form factor is a responsive hosted experience; source does not prescribe web versus mobile-first interaction beyond that. [Brief: Scope and Roadmap—Core implementation proof]

## Core user journey / use cases

1. **Connect:** User authorizes Findur through SnapTrade OAuth and connects a usable portfolio; they can later disconnect it with clear consent. [Brief: The Experience; Success Criteria]
2. **Set intent and boundaries:** User selects maximum distance, chooses similar/diversified/complementary investing preferences, and selects financial-information disclosure level. [Brief: The Experience]
3. **Discover:** System derives private portfolio-informed signals and assembles a continuously ranked, invisible-cohort swipe deck using portfolio attributes/activity, proximity, preferences, disclosure, engagement, and possibly market context. Exact location stays private. [Brief: The Experience; Product Promise; Addendum: Candidate Selection Model]
4. **Evaluate and swipe:** Portfolio visualization is the card hero; cards can communicate permitted composition, activity, value bands, shared traits, or contrasts while photos are obscured. The engine may use full consented data privately, but a viewer receives only the owner's selected disclosure. User makes one initial swipe decision. [Brief: The Experience; Addendum: Progressive Identity Reveal; Tiered Disclosure and Initiation]
5. **Match:** Mutual right swipes reveal photos and the match state. Chat is a high-value stretch goal, not core proof; unmatch/block accompany chat if implemented. [Brief: The Experience; Scope and Roadmap; Addendum: Roadmap Framing]

## Differentiators and product rules

- **Portfolio-led, not badge-led:** portfolio evidence changes discovery and initiation, rather than decorating a conventional profile. [Brief: Product Promise]
- **Progressive identity reveal:** one swipe stage only; do not introduce a second photo-based acceptance/rejection step after reveal. [Addendum: Progressive Identity Reveal]
- **Open entry, portfolio-governed exclusivity:** a usable connected portfolio admits a user; differentiated discovery and initiation rights emerge after signup rather than from a universal wealth threshold. [Brief: Disclosure, Exclusivity, and Trust; Addendum: Candidate Selection Model]
- **Tiered disclosure / asymmetric initiation:** users initially see candidates who disclose no more than they do; a more-disclosing user may initiate toward a less-disclosing user, who can evaluate interest without increasing disclosure. Tiers govern peer visibility and initiation, not the system's private matching use of consented data. [Brief: Disclosure, Exclusivity, and Trust; Addendum: Tiered Disclosure and Initiation]
- **Continuous invisible cohorts:** cohorts are neither named user tiers nor browseable destinations; their changing eligibility/ranking logic should not be exposed. [Brief: The Experience; Addendum: Candidate Selection Model]
- Tone: enjoyable enough to feel like dating, professional enough to earn trust. [Brief: Product Promise]

## MVP / capability scope

### Core implementation proof

- SnapTrade test-OAuth connect, callback, token lifecycle, data freshness, disconnect, and failure states. [Brief: Scope and Roadmap]
- Private portfolio-informed matching profile; candidate selection based on a focused combination of portfolio inputs, activity, user preferences, disclosure level, and hybrid proximity (user maximum distance plus nearer-is-stronger ranking). [Brief: Scope and Roadmap; Addendum: Candidate Selection Model]
- Focused initial disclosure control; invisible cohort-based deck; portfolio-led cards; one-stage swiping; mutual match and photo reveal. [Brief: Scope and Roadmap]
- Candidate pool must be synthetic; only owner-authenticated live SnapTrade data may be used. Do not expose raw balances, holdings, trades, security names, or financial scores on candidate cards. [Brief: Scope and Roadmap—Explicit boundaries; Technical research: Production-grade build constraints and future-product boundary]

### Stretch and longer direction

- Stretch: post-match one-to-one text chat, then appropriate unmatch/block controls; additional disclosure tiers and dynamic signals. [Brief: Scope and Roadmap]
- Longer direction: richer visualizations; sophisticated similarity/diversification/complementarity/activity/value/market-moment signals; mature reporting, moderation, anti-scam, retention, identity, and account-lifecycle controls; analytics/observability for relevance and trust. [Brief: Scope and Roadmap]

## Success metrics and counter-metrics

- First-experience success: users can connect/disconnect with clear consent; understand verification boundaries; receive a deck materially influenced by portfolio data, preferences, disclosure, and proximity; evaluate useful portfolio-led cards without exact private data by default; perceive disclosure as meaningful access; and complete a mutual-match/photo-reveal flow. [Brief: Success Criteria]
- Usability/product signals: connection completion, verification-boundary comprehension, perceived card relevance, disclosure-tier comfort, progression to mutual matches, and safety/trust concerns. [Brief: Success Criteria]
- Counter-metrics / guardrails implied by the source: misunderstanding "verified" as complete wealth or stability; unwanted exact-data/location exposure; disclosure pressure or unequal agency/sparse decks for safest-tier users; screenshot/inference/harassment/financial-targeting risk; bullying, misuse, catfishing, or post-match safety harms. [Brief: The Problem; Disclosure, Exclusivity, and Trust; Addendum: Tiered Disclosure and Initiation; Progressive Identity Reveal]

## Business model / market posture

- No pricing, revenue model, or acquisition model is specified. The stated exclusivity mechanism is differentiated discovery and initiation rights earned through voluntary, standardized financial disclosure—not a wealth paywall or universal wealth threshold. [Brief: Disclosure, Exclusivity, and Trust; Addendum: Tiered Disclosure and Initiation]
- The current commercial posture is evaluation/job-application prototype only. A real product requires written SnapTrade confirmation before one person's financial data, even in derived form, affects what another real user sees. [Technical research: Executive summary; Recommendation]

## Constraints, risks, and non-functional implications

- Minimize unnecessary PII; distinguish private matching inputs from user-visible data; ensure every displayed field has a product purpose; provide disclosure, disconnect, revocation, blocking, retention, and misuse-prevention controls. [Brief: Disclosure, Exclusivity, and Trust]
- SnapTrade integration should use a Commercial-account-owned OAuth app serving Personal SnapTrade users, with OAuth/OIDC authorization-code + PKCE and server-side confidential-client handling. This is preferred over the Commercial API-key user-registration model for "Continue with SnapTrade." [Technical research: Authentication and onboarding; Recommendation]
- Handle exact redirect URIs, PKCE/state/nonce/ID-token validation, encrypted server-side credential storage, deletion/revocation, connection errors, and observable failure paths. [Technical research: Authentication and onboarding; Production-grade build constraints and future-product boundary]
- Data is eventually consistent: activities may update daily; orders offer a more current but limited lookback; holdings freshness depends on plan/brokerage; webhooks indicate sync lifecycle rather than a holdings-change diff. Derived features need source/freshness metadata and freshness-labelled UI claims. [Technical research: Data and event fit]
- Do not claim real-time processing; use matching data read-focused; do not trade, recommend investments, or imply unconnected assets are visible. [Brief: Scope and Roadmap—Explicit boundaries; Technical research: Recommendation]
- Published policy/terms create a future compliance boundary for sharing, redistributing, or using portfolio/market data cross-user. Re-check OAuth/data conditions by 2026-10-01 and terms/policy before public release. [Technical research: Production-grade build constraints and future-product boundary; Staleness map]

## Assumptions and open questions for the PRD

- Define which derived portfolio signals/activity sources are used, their freshness rules, and the source/broker coverage constraints. [Brief: Open Decisions; Technical research: Data and event fit]
- Define standardized tier names/contents; whether any exact holdings, balances, or transactions may ever be shown; downgrade/revocation after viewing; and protections from gaming, screenshots, inference, harassment, and targeting. [Brief: Open Decisions; Addendum: Tiered Disclosure and Initiation]
- Define eligible-cohort/ranking rules, entry/exit, user preference controls, explanation/consent, and how to communicate relevance without revealing cohort logic. [Brief: Open Decisions; Addendum: Candidate Selection Model]
- Define the portfolio-card visual language and boundary between playful market-event relevance (for example comparable drawdown) and sensitive performance disclosure. [Brief: Open Decisions; Addendum: Timely and playful pools]
- Define identity verification, anti-catfishing, reporting, moderation, retention, account lifecycle, and whether live chat fits the milestone. [Brief: Open Decisions; Scope and Roadmap]
- If public launch is contemplated, resolve SnapTrade approval for cross-user derived compatibility/ranking use and whether it is a disclosure, market-data use, portfolio analysis, or signal. [Technical research: Open questions]

## Downstream technical / UX handoff (from the addendum)

### Architecture / technical design

- Specify OAuth/OIDC scope minimization, token/key lifecycle and encryption, callback validation, backend/session boundary, sync/webhook processing, freshness metadata, revocation/deletion, error handling, observability, and synthetic-data isolation. [Technical research: Authentication and onboarding; Data and event fit; Recommendation]
- Specify portfolio-derived feature pipeline and private-versus-visible data boundary. Candidate-selection data may include composition, value bands, activity, proximity, Findur activity, and combinations, but the exact model remains intentionally non-user-visible. [Addendum: Candidate Selection Model; Tiered Disclosure and Initiation]

### UX / safety design

- Design standardized disclosure tiers, initial-view and initiation rules, clear verification/freshness explanations, card information hierarchy, coarse-distance treatment, progressive photo reveal, and chat/unmatch/block if chat is included. [Addendum: Tiered Disclosure and Initiation; Progressive Identity Reveal; Brief: Product Promise]
- Test comprehension that connected data is partial and consented—not a wealth/stability verdict—and that higher disclosure creates access without coercion or exposure of internal ranking logic. [Brief: The Problem; Success Criteria; Addendum: Tiered Disclosure and Initiation]
