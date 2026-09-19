---
title: "Product Brief: Findur"
status: complete
created: 2026-09-19
updated: 2026-09-19
---

# Product Brief: Findur

## Executive Summary

Findur is a portfolio-first dating experience for everyday investors who consider financial and investing alignment important in a relationship. Conventional dating products lead with appearance and leave financial compatibility difficult to assess until much later. Findur reverses that order: a user connects a portfolio through SnapTrade, and consented portfolio data becomes the foundation for candidate selection, exclusivity, and trust.

Anyone with a usable connected portfolio can join. Portfolio attributes, proximity, preferences, activity, and disclosure shape an invisible candidate market and a continuous swipe deck. Cards lead with portfolio visualization rather than a clear photo. Mutual right swipes reveal photos and open the match, letting users first express interest through financial and investing signals without a second appearance-based rejection step.

## The Problem

Financial habits, stability, openness, and investing behavior can matter deeply to long-term compatibility, yet normal dating channels provide little trustworthy evidence about them. People either avoid the subject, rely on self-description, or discover significant misalignment after investing time in a relationship.

Existing photo-led discovery also makes appearance the dominant first filter. For investors who want financial alignment to carry real weight, changing a profile prompt or adding a verification badge does not go far enough. The underlying discovery market must respond to portfolio data.

At the same time, a connected investment portfolio is not a complete financial identity. It does not prove total wealth, income, debt, every account, or overall financial stability. Findur must create useful trust without overstating what the data establishes.

## Who This Serves

The primary user is an everyday investor who:

- wants romantic discovery to account for financial and investing alignment;
- is willing to connect a real portfolio in exchange for a more relevant and exclusive experience;
- values evidence and financial openness more than an unverified profile claim; and
- still expects control over what another person can see.

Findur is not intended to provide investment advice or certify that someone is financially responsible. Its job is to create a trusted dating context from the portfolio evidence a user voluntarily provides.

## Product Promise

> Findur helps investors discover romantic partners through trusted, portfolio-informed matching, where consented investment data materially shapes who appears, how compatibility is understood, and who can initiate.

SnapTrade serves two central roles:

1. **Matching intelligence:** connected portfolio data drives internal candidate eligibility and ranking rather than sitting behind a decorative badge.
2. **Trust foundation:** financial signals can be tied to consented connected accounts, their coverage, and their freshness instead of relying entirely on self-reporting.

“Verified” therefore means derived from the accounts a user chose to connect, as of a stated freshness point. It never means complete net worth or a comprehensive assessment of financial health.

Findur does not merely add wealth verification to a conventional dating profile. Portfolio data changes the market itself: candidate eligibility, ordering, comparison, disclosure rights, and timely relevance can all depend on it. The portfolio replaces the photo as the first discovery surface, while progressive identity reveal preserves human connection after mutual interest. The tone remains fun enough to feel like dating and professional enough to earn trust.

## Disclosure, Exclusivity, and Trust

Findur is open-entry but portfolio-governed. Exclusivity begins after signup, through differentiated discovery and initiation rights rather than a universal wealth threshold.

Users choose from standardized disclosure tiers. A user may initially discover candidates who have disclosed no more than they have. A more-disclosing user may initiate toward a less-disclosing user; the recipient can evaluate and respond without being forced to disclose more. This turns willingness to share meaningful financial context into a source of access and a signal of like-minded openness.

Every displayed field must have a product purpose. Findur should minimize unnecessary personally identifiable information (PII), distinguish private matching inputs from user-visible information, and provide clear controls for disclosure, disconnect, and revocation. Blocking, misuse prevention, retention boundaries, and protection against bullying or financial targeting are product requirements—not cleanup work.

## The Experience

1. **Connect:** The user authorizes Findur through SnapTrade OAuth and connects a usable portfolio.
2. **Set intent and boundaries:** The user chooses a maximum distance, indicates whether they prefer similar, diversified, or complementary investing styles, and selects a financial-information disclosure level.
3. **Discover:** Findur privately derives portfolio-informed signals and builds a continuously ranked candidate deck. Proximity acts as both a user-controlled boundary and a ranking weight. Exact location remains private.
4. **Evaluate:** Each card makes the portfolio the visual hero. It may communicate composition, activity, value bands, shared traits, or meaningful contrasts according to the candidate's disclosure choice. Photos remain obscured.
5. **Swipe:** Users make one initial decision. The matching engine can use the full consented connected portfolio privately, but a candidate never sees more than the owner chose to disclose.
6. **Match:** Mutual right swipes reveal photos and open the match. Working one-to-one chat is a high-value implementation stretch goal.

Candidate cohorts remain invisible and evolve with portfolio characteristics, activity, proximity, engagement, preferences, disclosure, and market events. Shared experiences can add playful relevance without exposing exact performance or ranking logic.

## Success Criteria

The first experience succeeds when a user can:

- connect and later disconnect a SnapTrade portfolio with clear consent;
- understand what connected data does and does not verify;
- receive a candidate deck materially influenced by portfolio data, preferences, disclosure, and proximity;
- compare candidates through useful portfolio-led cards without seeing exact private data by default;
- feel that greater voluntary disclosure creates meaningful access rather than arbitrary complexity; and
- reach a mutual match and photo reveal through a single, coherent swipe flow.

For later usability testing, important signals include connection completion, comprehension of verification boundaries, perceived relevance of candidate cards, comfort with disclosure tiers, progression to mutual matches, and safety or trust concerns raised during the journey.

## Scope and Roadmap

### Core implementation proof

- A publicly hosted, responsive demonstration using only SnapTrade's test OAuth environment.
- Secure connect, callback, token lifecycle, freshness, disconnect, and failure states.
- One real private test portfolio for the owner, with synthetic candidate profiles and portfolio signals.
- Portfolio-derived matching inputs, proximity based on the user's selected maximum distance and ranking weight, a focused version of disclosure control, and an invisible cohort-based swipe deck.
- Portfolio-led candidate cards, one-stage swiping, mutual match, and photo reveal.

### High-value stretch

- Working one-to-one text chat after a match.
- Unmatch and block controls appropriate to that chat surface.
- Additional disclosure tiers and more dynamic candidate signals.

### Longer-term product direction

- Richer portfolio visualizations and justified voluntary disclosure options.
- More sophisticated similarity, diversification, complementary-style, activity, value, and market-moment signals.
- Mature reporting, moderation, anti-scam, retention, identity, and account-lifecycle controls.
- Observability and product analytics that reveal whether portfolio-led discovery improves match relevance and trust.

### Explicit boundaries

- Findur will not launch to real public users or use SnapTrade production credentials.
- It will not execute trades, recommend investments, or treat portfolio composition as a universal score of personal worth.
- It will not claim access to unconnected assets or present connected holdings as complete net worth.
- It will not use real candidate financial data in the demonstration pool.

## Vision

Findur imagines a dating market where a connected portfolio becomes a living compatibility surface. As users, portfolios, and market conditions change, discovery can adapt through financial characteristics, behavior, disclosure, proximity, preferences, and timely context—while remaining trustworthy and under user control.

## Open Decisions for the PRD and UX

- Which derived portfolio signals and activity sources are used, and how freshness affects them.
- The exact contents and names of disclosure tiers.
- The visual language of a portfolio-led candidate card.
- How much explanation accompanies a recommendation without exposing internal cohort logic.
- The boundaries between playful market-event relevance and sensitive performance disclosure.
- Identity verification, anti-catfishing, reporting, and moderation behavior.
- Whether live chat fits within the implementation milestone.
