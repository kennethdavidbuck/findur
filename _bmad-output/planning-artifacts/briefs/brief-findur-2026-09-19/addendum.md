---
title: "Findur Product Brief Addendum"
status: complete
created: 2026-09-19
updated: 2026-09-19
---

# Findur Product Brief Addendum

## Governing Trust Constraint

SnapTrade serves two central product purposes: it supplies the portfolio data that drives candidate selection, and it establishes trust in the financial signals users choose to share.

Findur should make clear that a signal is derived from consented connected-account data as of a stated freshness point. It must not imply that connected data represents complete wealth, every financial account, or overall financial stability. Trust also depends on user-controlled disclosure, secure credential handling, visible disconnect and revocation behavior, and safety controls around post-match interaction.

This addendum is a decision-and-question handoff. Declarative statements record intended product decisions. Terms such as *may*, *should*, *proposed*, and *remain to be defined* mark questions for downstream PRD, UX, or architecture work. The product brief remains the source of high-level scope.

## Candidate Selection Model

Findur uses a hybrid candidate-selection model in which connecting a usable portfolio admits the user; portfolio-derived and product-behavior signals determine the internal cohorts from which the swipe deck is assembled and ranked.

These cohorts are not named destinations, visible tiers, or explicit memberships. As with mainstream swipe-based dating products, the user experiences a continuous card deck without knowing the precise combination of signals that caused a particular candidate to appear.

### User preference inputs

- Find people with similar investing behavior or portfolio characteristics.
- Find people with diversified portfolios.
- Find people whose investing style or portfolio complements the user's own.

These preferences tune candidate retrieval and ranking; they do not open a visible pool browser.

### System-derived eligibility

Potential pool inputs include:

- geographic proximity;
- portfolio contents and composition;
- portfolio value or value bands;
- recent trading activity;
- activity within Findur; and
- combinations of these inputs rather than a single universal rank.

The eligible cohort and card ordering can change as these attributes change. Proximity uses a hybrid model: the user controls a maximum distance, and nearer candidates rank more strongly within that boundary. The product should expose only coarse distance and avoid revealing precise user location.

### Timely and playful pools

Event-driven internal cohorts can turn shared market experiences into candidate-selection signals. One example is temporarily increasing the likelihood of showing users whose portfolios recently experienced a comparable drawdown of approximately 20%.

These concepts require later definition in the PRD: eligibility and ranking rules, data source and freshness, cohort entry and exit, user preference controls, explanation, consent, disclosure granularity, and protection against exposing exact performance or holdings.

## Progressive Identity Reveal

Findur leads candidate cards with portfolio visualization and financial comparison rather than an immediately visible profile photo. There is one swipe stage: when both users swipe right, Findur reveals their photos and opens chat. It does not ask either person to make a second explicit appearance-based swipe, avoiding added complexity and a conspicuous post-reveal rejection experience. Private identity-verification and anti-catfishing safeguards remain to be defined.

## Tiered Disclosure and Initiation

Users choose a standardized financial-information visibility tier. A user may initially see candidates who have disclosed no more than the user has disclosed. A more-disclosing user may initiate a right swipe toward a less-disclosing user; the less-disclosing user can then receive and evaluate that inbound interest without increasing their own disclosure.

Findur may privately use the full portfolio data from consented connected accounts for candidate selection regardless of display tier. Tiers govern what other users can see and who may initiate. The amount and kind of portfolio information shared contribute to exclusivity and help connect people with similar attitudes toward financial openness.

This makes voluntary transparency an access mechanism without excluding lower-disclosure users from Findur. Every displayed field must have a clear product purpose. The first release may implement only a limited version of tiered disclosure to keep scope focused; that limitation does not change the longer-term model.

The detailed design must address:

- standardized tier contents, so users cannot game a nominal tier with low-value fields;
- whether exact holdings, balances, and transactions are ever appropriate to display;
- downgrade and revocation behavior after another user has already seen information;
- screenshot, inference, harassment, and targeting risks;
- sparse decks or unequal agency for users who choose the safest tier; and
- clear explanation of the rule without revealing internal candidate cohorts or ranking logic.

## Roadmap Framing

The roadmap preserves the complete product direction while implementation proceeds through coherent milestones and may stop at an earlier one. These groups are a planning frame, not an implementation commitment.

### Core proof

- Connect and manage a SnapTrade OAuth portfolio.
- Derive a private portfolio-informed matching profile.
- Apply proximity, preferences, portfolio attributes, activity, and disclosure level to candidate selection.
- Present a portfolio-led swipe deck with progressive photo reveal.
- Complete mutual right swipes, reveal photos, and show the match state.

### High-value stretch

- Working one-to-one text chat after a mutual match.
- Unmatch and block controls appropriate to an implemented chat surface.

### Longer-term product direction

- Richer standardized disclosure tiers and initiation rights.
- Broader dynamic and event-driven candidate signals.
- More complete trust, safety, reporting, moderation, retention, and account-lifecycle controls.
