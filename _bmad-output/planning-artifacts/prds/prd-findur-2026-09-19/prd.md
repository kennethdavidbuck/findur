---
title: Findur Product Requirements Document
status: final
created: 2026-09-19
updated: 2026-09-19
---

# PRD: Findur

## 0. Document Purpose

This PRD defines the product requirements for Findur's first publicly hosted demonstration. It is written for product, UX, architecture, engineering, and review teams. It carries forward the decisions in the [product brief](../../briefs/brief-findur-2026-09-19/brief.md), its [addendum](../../briefs/brief-findur-2026-09-19/addendum.md), and the [SnapTrade feasibility research](../../research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md). Features contain globally stable Functional Requirement (FR) IDs; inferred decisions are marked `[ASSUMPTION]` and indexed in §12. Technical mechanisms and deeper UX guidance are preserved in [addendum.md](addendum.md), not prescribed here.

**Current operating boundary:** one protected owner may use live SnapTrade data; Discovery uses generated synthetic profiles; the public product is a responsive demonstration, not a public dating launch.

## 1. Vision

Findur is a portfolio-first dating experience for everyday investors who want financial and investing alignment to meaningfully shape romantic discovery. Instead of leading with profile photos and adding a financial badge, Findur makes a consented connected portfolio the first candidate surface. Photos remain obscured until two people express mutual interest.

The product uses connected financial data as evidence for compatibility, not as a measure of human worth. Private portfolio-derived signals influence candidate eligibility, ordering, comparison, and permitted disclosure. The viewer sees only standardized information the candidate has agreed to share. “Verified” means that information was derived from accounts the user chose to connect at a stated freshness point; it does not mean complete net worth, every account, financial health, or financial responsibility.

This milestone proves that the idea can become a coherent, trustworthy experience and visibly demonstrates the breadth of the SnapTrade integration. It uses one owner-controlled live portfolio in SnapTrade's test OAuth environment, a private live-data showcase and card preview, and a large generated population of visually equivalent Synthetic Candidates. Portfolio visualizations—not a generic verification badge—are the center of the demonstration. It is a publicly hosted demonstration, not a public dating launch.

## 2. Target User and Jobs

### 2.1 Primary User

The primary user is an everyday investor who values evidence and financial openness, wants investing compatibility to inform dating, will connect a portfolio in exchange for a more relevant experience, and expects control over what other people can see.

### 2.2 Jobs To Be Done

- When deciding whom to date, help me assess investing compatibility through meaningful evidence rather than self-description alone.
- Let me express whether I prefer similar, diversified, or complementary investing behavior.
- Let me control how much portfolio-derived information another person can see without preventing the system from finding relevant candidates.
- Give me confidence that connected data is current enough for its stated purpose and is not being represented as my complete financial life.
- Let attraction emerge from compatibility before appearance, while preserving a clear, enjoyable dating flow.
- Let me present enough personal and dating context to be understood as a person, not only as a portfolio.
- Let me preview exactly how my profile and financial disclosure will appear before anyone else could see it.
- Let me withdraw consent, disconnect my portfolio, and understand what happens to derived information afterward.

### 2.3 Non-Users for This Milestone

- Members of the public seeking a functioning multi-user dating service.
- Anyone under 18.
- People seeking investment advice, portfolio analysis, trading, wealth certification, or access to another person's exact financial records.
- Real candidate users other than the owner-controlled test user.

### 2.4 Key User Journeys

- **UJ-1. Maya connects and explores her live portfolio.** Maya is an everyday investor evaluating whether Findur offers a more meaningful dating experience. She enters the hosted demonstration, sees what data Findur will use privately and what another person could see, and authorizes Findur through SnapTrade test OAuth. Findur returns her to a private portfolio showcase where she can inspect the available accounts, balances and values, holdings or positions, activities or orders, and source freshness. She then previews how those live fields become her own Candidate Card at different Disclosure Levels. The value lands when Maya sees both the breadth of the SnapTrade integration and the boundary between connected data, derived compatibility, and public disclosure. If connection or refresh fails, she sees a recoverable state rather than an empty or misleading profile.

- **UJ-2. Maya sets compatibility and disclosure boundaries.** With a Usable Portfolio connected, Maya selects a maximum distance, chooses whether she wants similar, diversified, or complementary investing behavior, and chooses a standardized Disclosure Level. Findur explains that these settings affect who appears and what may be shown without exposing internal cohort or ranking logic. The value lands when Maya can predict the privacy effect of her choices before entering Discovery.

- **UJ-3. Maya explores a deep, visual Swipe Deck.** Maya opens Discovery and receives an ordered Swipe Deck drawn from a large generated Synthetic Population rather than a small fixed set of profiles. Each Candidate Card leads with rich visualizations assembled from synthetic SnapTrade-shaped data and limited by the candidate's Disclosure Level. Depending on that level, a card may show allocation, named holdings, position weights or values, portfolio value or performance context, and recent activity alongside coarse Proximity and freshness context. The candidate's Photo remains obscured. Maya can see the deck respond credibly to compatibility, distance, disclosure, activity, and portfolio differences without seeing an exact location, account identifiers, internal ranking logic, or a financial-worth score. She swipes once based on the combined card.

- **UJ-4. Maya reaches a Mutual Match and sees the Photo reveal.** Maya swipes right on a Synthetic Candidate whose preconfigured decision is also positive. Findur creates a Mutual Match and reveals the candidate's Photo without asking for a second acceptance decision. Maya can review the matched state and return to Discovery. Chat is not required for this milestone.

- **UJ-5. Maya withdraws access.** Maya disconnects her portfolio after trying the experience. Findur immediately prevents stale connected status from being presented as current, suppresses portfolio-derived discovery use, explains the resulting product state, and initiates the applicable deletion or retention behavior. She can reconnect through a fresh authorization flow.

- **UJ-6. Alex understands Findur before signing in.** Alex arrives at the public web app without an account. A branded landing experience explains the portfolio-first premise, how connection and disclosure work, why Photos remain obscured, and that the current product is a demonstration. Alex can inspect About, trust and safety, Privacy, Terms, and contact information before choosing the protected owner entry or, if included after architecture sizing, a signup-free Guest Demo using only ephemeral synthetic data.

- **UJ-7. Maya completes and previews her whole profile.** After connecting her portfolio, Maya provides the minimum personal and dating information needed to make a Candidate Card meaningful, adds a Photo and short self-description, and sets discovery preferences. She opens Profile Preview and switches between pre-match and post-match views and among Disclosure Levels. The value lands when the preview combines her own live portfolio visualizations with her personal profile exactly as the corresponding experience would present them, before anything could be shared.

## 3. Glossary

- **Candidate Card** — A discovery card for one Synthetic Candidate. It presents only information allowed by that candidate's Disclosure Level and keeps the Photo obscured until a Mutual Match.
- **Connected Portfolio** — The accounts and data a user explicitly authorizes Findur to read through SnapTrade. It is not a complete statement of wealth or financial health.
- **Disclosure Level** — One of three standardized settings—Snapshot, Holdings, or Full Detail—that controls which source fields and Portfolio-Derived Signals another user may see and the user's initiation eligibility. It does not limit Findur's consented private matching calculations.
- **Discovery** — The Findur surface on which a user evaluates Candidate Cards and swipes.
- **Freshness State** — The recorded source and recency context for Connected Portfolio data, including connected, syncing, stale, needs reauthorization, failed, or disconnected states.
- **Guest Demo** — A conditional signup-free trial that uses an ephemeral synthetic viewer profile and Synthetic Candidates only. It cannot initiate OAuth, access the Connected Portfolio, or persist a real-user account.
- **Invisible Cohort** — A system-managed, non-browseable eligibility group used to assemble a Swipe Deck. Users do not see cohort names, rules, or membership.
- **Mutual Match** — The state created when both sides have a positive swipe decision, triggering Photo reveal.
- **Photo** — Candidate identity imagery that remains obscured on a Candidate Card and becomes visible after a Mutual Match.
- **Personal Profile** — The user's non-financial identity and dating information, including the minimum information needed for discovery, Candidate Cards, and the matched state.
- **Profile Preview** — An owner-only representation of the complete Personal Profile and Connected Portfolio presentation at each relevant Disclosure Level and pre-match or post-match state.
- **Portfolio Showcase** — The owner-only surface that visibly demonstrates live SnapTrade account, balance/value, holding/position, activity/order, connection, and freshness data.
- **Public Site** — The unauthenticated product shell that presents Findur's brand, product proposition, explanation, legal information, trust posture, contact path, and entry into the demonstration.
- **Portfolio Visualization** — A chart, composition view, holding view, trend, activity treatment, comparison, or other visual representation of source or derived portfolio data. Its detail is governed by a Disclosure Level when shown on a Candidate Card.
- **Portfolio-Derived Signal** — A trait, band, comparison, or activity indicator calculated from a Connected Portfolio for private matching or permitted display. It is not investment advice or a financial-worth score.
- **Proximity** — Approximate distance information used for eligibility and ranking. Exact location is never shown.
- **Swipe Deck** — The ordered set of Candidate Cards assembled for the current user from eligible Invisible Cohorts.
- **Synthetic Candidate** — A fictional candidate profile containing no real person's financial data and used to demonstrate discovery and matching.
- **Synthetic Population** — The reproducibly generated collection of Synthetic Candidates used to exercise eligibility, ranking, disclosure, Portfolio Visualization, swipe, match, and edge-case behavior at meaningful scale.
- **Usable Portfolio** — A Connected Portfolio with sufficient accessible data and an acceptable Freshness State to power the demonstration.
- **Verified** — A limited claim that displayed or derived information came from user-authorized Connected Portfolio data at a stated freshness point. It does not certify completeness, wealth, stability, responsibility, or identity.

## 4. Features and Functional Requirements

FR IDs are globally stable. They remain unchanged when later decisions add requirements, so numeric order within a feature is not significant.

### 4.1 Consent-Led Portfolio Connection

**Description:** The user connects a portfolio through a consent flow that explains purpose, scope, visibility, and limitations before authorization. Findur communicates lifecycle states instead of treating connection as a one-time badge. Realizes UJ-1 and UJ-5.

#### FR-1: Explain consent before connection

The user can review the categories of Connected Portfolio data Findur intends to access, how they will be used privately, what may become visible, and how to disconnect before beginning authorization.

**Consequences (testable):**

- The pre-connection experience distinguishes private matching use from user-visible disclosure.
- It states that connection does not prove complete net worth, financial health, financial responsibility, or identity.
- It states that Findur does not trade or provide investment advice.
- Portfolio derivation begins only after the user takes an explicit action consenting to use the Connected Portfolio for Findur's matching and card-preview experience.

#### FR-2: Authorize through SnapTrade test OAuth

The owner-controlled test user can authorize Findur through SnapTrade's test OAuth flow and return to the demonstration in an authenticated success or failure state.

**Consequences (testable):**

- A valid authorization produces a Connected Portfolio without exposing credentials in the browser or repository.
- A denial, invalid callback, expired flow, or provider error produces a safe recoverable state.
- The experience explains that the owner must first have a SnapTrade Personal account with an eligible brokerage connection before Findur can read portfolio data.
- The demonstration does not support public signup or real candidate connections.
- Live SnapTrade authorization and owner data are available only after authenticating as the preconfigured owner; unauthenticated visitors and Guest Demo sessions cannot start OAuth or reach owner routes or data.

#### FR-3: Represent connection and freshness

Findur can represent the Connected Portfolio's source coverage, last known refresh context, and Freshness State wherever portfolio-derived claims are material.

**Consequences (testable):**

- The UI never labels stale, failed, disconnected, or reauthorization-required data as current.
- When refresh timing is uncertain, Findur uses qualified language rather than “real time.”
- An unusable Freshness State prevents new portfolio-derived discovery output until recovered.

#### FR-4: Disconnect and withdraw consent

The user can disconnect a Connected Portfolio and receive confirmation of the resulting data state. Realizes UJ-5.

**Consequences (testable):**

- Disconnect immediately removes the user's connected status and stops new portfolio-derived discovery use.
- Disconnect immediately deletes locally stored source payloads, Portfolio-Derived Signals, and rendered or cached portfolio views for the demonstration; only categorical security or audit events containing no financial values may remain.
- The user receives confirmation of the local deletion boundary and any provider-side connection state Findur cannot delete itself.
- Reconnection requires a fresh authorization flow.

### 4.2 Connected Portfolio Showcase and Card Preview

**Description:** After connection, the owner can explore the live data Findur receives from SnapTrade and see how it becomes a portfolio-first card. This private surface makes API usage visible without exposing the owner's live financial data to another real person. Realizes UJ-1.

#### FR-19: Present available live portfolio data

The owner-controlled test user can inspect the live SnapTrade data available from the Connected Portfolio in a private Portfolio Showcase.

**Consequences (testable):**

- The Portfolio Showcase represents connected accounts, balances or account values, holdings or positions, activities or orders, and connection status when each dataset is available from the provider.
- Each section distinguishes loaded, unavailable, empty, stale, failed, and unsupported states instead of silently omitting them.
- Exact source values may be shown to the authenticated owner; credentials, tokens, and provider secrets may not.
- Source and Freshness State context accompanies time-sensitive values.

#### FR-20: Preview the owner's Candidate Card

The owner can preview how their live Connected Portfolio would appear on their Candidate Card under each available Disclosure Level.

**Consequences (testable):**

- The preview uses the same Portfolio Visualization components and disclosure rules as cards in Discovery.
- Changing the Disclosure Level immediately changes which fields are hidden, bucketed, derived, or exact.
- The owner can inspect representative phone and desktop presentations without either mode omitting required information or controls.
- The preview is clearly labeled as the owner's private preview and is never placed in another real user's Swipe Deck during this milestone.

#### FR-21: Demonstrate source-to-experience transformation

The Portfolio Showcase can make the relationship between SnapTrade source data, Portfolio-Derived Signals, and Candidate Card visualizations understandable to a reviewer.

**Consequences (testable):**

- A reviewer can identify which visible experiences are backed by accounts, balances or values, holdings or positions, activities or orders, and freshness or connection metadata.
- Missing provider data produces a visible unsupported or unavailable treatment rather than fabricated values.
- The experience remains read-only and makes no trade or investment recommendation.

### 4.3 Personal Profile and Complete Preview

**Description:** Portfolio data shapes discovery but does not replace the human profile. The owner supplies enough personal and dating context to create a meaningful Candidate Card, then previews the combined experience before entering Discovery. Realizes UJ-7.

#### FR-23: Complete a minimum Personal Profile

The owner-controlled test user can create and edit the Personal Profile needed for Candidate Cards, discovery eligibility, and the matched state.

**Consequences (testable):**

- The initial profile captures a display name, adult-age confirmation, coarse location, discovery eligibility or preferences, relationship intent, at least one Photo, and a short biography or prompt response. `[ASSUMPTION: These categories are the minimum viable Personal Profile; UX design will determine exact labels, input formats, optional fields, and validation.]`
- Findur distinguishes required, optional, private, pre-match visible, and post-match visible information.
- The user can correct profile information and understand which experience states the change affects.
- Synthetic Candidates use the same Personal Profile schema as the owner preview.
- Discovery remains unavailable until every required Personal Profile field is valid and a Usable Portfolio is connected; Findur identifies each incomplete requirement and returns the owner to the relevant profile or connection flow.

#### FR-24: Preview the complete profile

The owner can open a Profile Preview that combines the Personal Profile with the live Candidate Card preview.

**Consequences (testable):**

- Profile Preview supports every prototype Disclosure Level and both pre-match and post-match identity states.
- The pre-match view obscures the Photo and shows only information allowed at that stage; the post-match view reveals the Photo and any information designated for the matched state.
- The preview reflects current personal-profile edits and live portfolio presentation without entering the owner into a real Swipe Deck.
- The preview persistently identifies live owner data, while Discovery and synthetic preview states persistently identify synthetic data.
- Representative phone and desktop previews preserve the same disclosure, freshness, and identity-reveal rules while allowing different responsive layouts.

### 4.4 Compatibility Preferences and Disclosure

**Description:** The user shapes candidate relevance through a small set of understandable preferences and a standardized Disclosure Level. Findur explains the effect on experience without revealing ranking formulas or Invisible Cohorts. Realizes UJ-2.

#### FR-5: Set discovery preferences

The user can set a maximum distance and choose similar, diversified, or complementary investing compatibility.

**Consequences (testable):**

- Saved preferences influence Swipe Deck eligibility or ordering.
- Proximity uses the user's maximum as an eligibility boundary and favors nearer eligible candidates within that boundary.
- Changing a preference can produce a different Swipe Deck without requiring reconnection.

#### FR-6: Choose a Disclosure Level

The user can choose Snapshot, Holdings, or Full Detail and preview the source fields and Portfolio-Derived Signals each permits.

**Consequences (testable):**

- **Snapshot** shows allocation, asset classes, diversification context, value and activity bands, and Freshness State; it does not show security names or exact monetary values.
- **Holdings** adds security names, position weights, percentage performance, and activity categories; quantities, exact monetary values, and transaction detail remain hidden or bucketed.
- **Full Detail** adds exact quantities and values, portfolio or account value, performance amounts, and recent activity, order, or transaction detail.
- In regular Discovery, a user sees only candidates whose Disclosure Level is equal to or lower than their own and sees only the detail permitted by the candidate's level.
- A higher-disclosure user may initiate interest toward a lower-disclosure user. The lower-disclosure recipient may evaluate that incoming interest using the initiator's chosen level without increasing their own Disclosure Level; the recipient's own information remains limited to their lower level.
- Changing to a lower Disclosure Level immediately invalidates no-longer-permitted active views, rendered data, prefetched card data, and caches; reopening or navigating back cannot reveal the prior detail.
- Account identifiers, credentials, tokens, and exact location are never disclosable.
- During this milestone, exact live fields are shown only in the owner's private Portfolio Showcase and Profile Preview; Candidate Cards in Discovery use synthetic data.

#### FR-7: Explain preference effects without exposing internal logic

Findur can explain that preferences, Disclosure Level, Proximity, and portfolio compatibility affect Discovery while keeping Invisible Cohort membership and exact ranking logic private.

**Consequences (testable):**

- Candidate relevance language avoids guarantees and financial-worth judgments.
- No screen exposes an Invisible Cohort name, boundary, membership list, or ranking score.

### 4.5 Private Portfolio Profile and Candidate Selection

**Description:** Findur converts the Connected Portfolio into a private matching profile, then combines it with preferences, Disclosure Level, Proximity, and candidate activity to assemble a continuously ordered Swipe Deck from a large generated Synthetic Population. Realizes UJ-3.

#### FR-8: Derive a private matching profile

Findur can derive focused Portfolio-Derived Signals from a Usable Portfolio and attach source and Freshness State metadata to each signal.

**Consequences (testable):**

- Source and derived inputs may represent accounts, balances or values, holdings or positions, composition, diversification, performance context, activities or orders, and activity recency when supported by available source data.
- Missing or unsupported inputs remain unknown; they are not inferred as zero or negative traits.
- No single universal wealth, desirability, stability, or financial-responsibility score is created.

#### FR-9: Isolate live and synthetic data

The demonstration can combine the owner-controlled live matching profile with Synthetic Candidates without representing synthetic values as connected or verified real-user data.

**Consequences (testable):**

- Only the owner-controlled test user supplies live Connected Portfolio data.
- Every candidate financial record and swipe response in Discovery is synthetic, including any exact-looking holding, value, performance, order, activity, or transaction detail.
- Synthetic cards use the same Portfolio Visualization and Disclosure Level rules as the owner's live card preview.
- Every synthetic Candidate Card, candidate detail, and Mutual Match state carries a persistent, legible synthetic-data indicator that cannot disappear through card expansion, responsive layout, or navigation.

#### FR-10: Assemble an eligible Swipe Deck

Findur can assemble and order a Swipe Deck using compatibility preference, portfolio signals, Disclosure Level, Proximity, and available candidate activity.

**Consequences (testable):**

- Changing at least one material preference or portfolio-derived input changes eligibility or order in a seeded demonstration scenario.
- Candidates beyond the maximum distance are excluded.
- Candidate selection operates through Invisible Cohorts; users cannot browse cohorts directly.
- Sparse eligibility produces an honest low-inventory state rather than silently weakening privacy boundaries.

#### FR-22: Generate a varied Synthetic Population

Findur can reproducibly generate a large Synthetic Population whose profiles and portfolios exercise the product's matching, disclosure, visualization, and state logic.

**Consequences (testable):**

- The population is large and varied enough to populate credible Swipe Decks across the supported compatibility preferences, Disclosure Levels, Proximity ranges, portfolio patterns, freshness conditions, and match outcomes.
- The demonstration includes intentional boundary and edge-case scenarios, including sparse eligibility, maximum-distance boundaries, low and high disclosure, empty activity, stale data, unsupported fields, and reciprocal or non-reciprocal swipes.
- The population can be regenerated without hand-authoring individual candidate fixtures.
- No generated profile or portfolio is copied from a real person's financial records or presented as a real connected user.

### 4.6 Portfolio-First Discovery

**Description:** Candidate Cards make rich Portfolio Visualizations the hero while preserving consent boundaries and a recognizable dating interaction. The card conveys enough portfolio substance for one swipe decision and demonstrates the breadth of SnapTrade-shaped data without exposing real cross-user financial data in this milestone. Realizes UJ-3.

#### FR-11: Present a portfolio-first Candidate Card

The user can view Candidate Cards whose primary content is a rich, disclosure-controlled Portfolio Visualization, supported by Portfolio-Derived Signals, comparison context, coarse Proximity, and relevant Freshness State.

**Consequences (testable):**

- Portfolio content has greater visual prominence than the obscured Photo and uses charts or other visual forms rather than reducing the portfolio to a badge or text-only summary.
- Content is limited by the candidate's Disclosure Level; permitted synthetic details may include allocation, security names, position weights or quantities, values, performance context, activities, orders, or transactions.
- The card identifies data freshness and distinguishes source facts from derived comparisons where confusion is plausible.
- The card does not expose account identifiers, credentials, tokens, exact location, or internal ranking scores.
- The card avoids language implying investment advice or a judgment of financial worth.

#### FR-12: Make one swipe decision

The user can swipe left or right once on the combined Candidate Card.

**Consequences (testable):**

- A swipe is persisted for the demonstration session and advances the deck.
- A Photo reveal never triggers a second accept-or-reject gate.
- Repeated input does not create duplicate swipe decisions.

#### FR-13: Handle discovery edge states

Findur can explain when Discovery is unavailable because the portfolio is disconnected, syncing, stale, needs reauthorization, has failed, or yields no eligible Synthetic Candidates.

**Consequences (testable):**

- Each state offers an appropriate next action where recovery is possible.
- Findur does not substitute unqualified candidates merely to keep the deck populated.

### 4.7 Mutual Match and Progressive Identity Reveal

**Description:** Mutual interest reveals the Photo and creates a matched state. The milestone ends at that proof; messaging is deferred. Realizes UJ-4.

#### FR-14: Create a Mutual Match

Findur can create a Mutual Match when the owner-controlled test user's right swipe meets a Synthetic Candidate's preconfigured positive decision.

**Consequences (testable):**

- A right swipe without reciprocal interest does not reveal the Photo.
- A Mutual Match is created only once for the same pair.
- A left swipe never creates a Mutual Match.

#### FR-15: Reveal identity after mutual interest

Findur can reveal the matched Synthetic Candidate's Photo and present a clear matched state immediately after a Mutual Match.

**Consequences (testable):**

- The Photo remains obscured before the Mutual Match and is visible afterward.
- The reveal is a consequence of the original swipe, not a second decision stage.
- The matched state offers a path back to Discovery.

### 4.8 Demonstration Safety and Trust

**Description:** Even without public multi-user interaction, the hosted experience must not normalize unsafe financial disclosure or misleading dating claims. It establishes product rules that future public work cannot silently bypass.

#### FR-16: Enforce adult-only positioning

Findur can state that the experience is for adults and prevent the demonstration from presenting minors as users or candidates.

**Consequences (testable):**

- Entry messaging states an 18+ boundary.
- All Synthetic Candidates are unambiguously adults.

#### FR-17: Communicate prohibited interpretations and conduct

Findur can surface concise trust guidance that prohibits investment solicitation, money requests, financial targeting, and claims that Verified data establishes financial responsibility or personal worth.

**Consequences (testable):**

- Trust language is accessible before or during Discovery, not hidden only in legal text.
- The demonstration never encourages contacting a candidate for investment opportunities.

#### FR-18: Preserve a future safety boundary

The product requirements can distinguish demonstration-safe behavior from functionality that requires production moderation, reporting, blocking, identity, age-assurance, and legal review.

**Consequences (testable):**

- No public multi-user or messaging feature is enabled under the demonstration's safety model.
- Any future public-launch plan must pass the release gates in §10 rather than treating the demonstration as production-ready.

### 4.9 Public Product Shell and Trust Information

**Description:** The hosted web app presents Findur as a credible product before authentication. Public-facing pages establish the proposition, brand, demonstration boundary, and information a visitor expects before connecting financial data. Realizes UJ-6.

#### FR-25: Present the Findur proposition and brand

An unauthenticated visitor can understand what Findur is, who it serves, how portfolio-first discovery works, and how to enter the demonstration.

**Consequences (testable):**

- The Public Site includes a branded landing experience, concise product pitch, explanation of the portfolio-first and progressive-reveal model, and a clear entry action.
- It distinguishes the current demonstration from a public dating launch and avoids implying production availability or unsupported guarantees.
- Core public content and navigation remain usable on phone and desktop.
- Branding and product language follow §9.3 rather than defaulting to generic authentication scaffolding.

#### FR-26: Provide expected informational and legal surfaces

An unauthenticated visitor can access About, Terms, Privacy, trust and safety, and contact or support information from persistent public navigation or footer paths.

**Consequences (testable):**

- Each required surface has a stable route and is reachable without authentication.
- Terms and Privacy explain the demonstration's live-owner/synthetic-candidate boundary and the intended handling of connected financial data; draft legal content is labeled as such until professionally reviewed.
- Trust and safety content states the adult-only, non-advisory, anti-solicitation, consent, disclosure, and reporting expectations appropriate to the current scope.
- Contact or support information gives a visitor a clear way to raise a privacy, safety, or product concern.

#### FR-27: Offer a signup-free synthetic trial (conditional stretch)

If architecture confirms that it does not materially expand session, security, or delivery scope, a public visitor can enter a Guest Demo without creating an account.

**Consequences (testable):**

- The Guest Demo uses a preset or generated synthetic viewer profile and Synthetic Candidates only.
- A guest can exercise the representative profile, preference, Discovery, swipe, Mutual Match, and Photo-reveal experience without accessing live financial data.
- Guest state is isolated from the owner, expires or resets, and does not persist a real-user account or collect information unnecessary for the trial.
- Guest routes cannot initiate SnapTrade OAuth, access the Portfolio Showcase, or retrieve owner data.
- If architecture defers the capability, the Public Site and protected owner demonstration remain complete without it.

## 5. Non-Goals

- Providing investment advice, recommendations, brokerage actions, or portfolio-management tools.
- Certifying a person's net worth, stability, responsibility, compatibility, identity, or personal worth.
- Building a wealth leaderboard, universal wealth threshold, or paywall tied to access to another person's financial data.
- Launching a public multi-user dating service in this milestone.
- Connecting real candidate financial data or using one real person's financial data to affect another real user's experience.
- Exposing live owner portfolio data to another real person during this milestone, or displaying portfolio detail beyond the owner's selected Disclosure Level in a future multi-user product.
- Exposing credentials, tokens, account identifiers, or exact location in the Portfolio Showcase or on Candidate Cards.
- Exposing Invisible Cohorts, exact ranking formulas, or a browseable financial-status hierarchy.
- Trading, real-time market processing, production brokerage coverage, or claims of complete asset visibility.
- Shipping public chat, monetization, native mobile apps, or production moderation operations in this milestone.

## 6. MVP Scope

### 6.1 In Scope

- Responsive, publicly hosted demonstration surface.
- One owner-controlled live portfolio through SnapTrade test OAuth.
- Minimum Personal Profile creation and editing for the owner and equivalent generated fields for Synthetic Candidates.
- Consent, callback, connection, Freshness State, failure, reauthorization, and disconnect flows.
- Private Portfolio Showcase covering available accounts, balances or values, holdings or positions, activities or orders, and provider lifecycle states.
- Complete live owner Profile Preview across prototype Disclosure Levels, pre-match and post-match states, and representative phone and desktop presentations.
- Private Portfolio-Derived Signals with source/freshness metadata.
- Maximum distance, compatibility preference, and focused Disclosure Level controls.
- Generated Synthetic Population at sufficient scale and variety to exercise Invisible Cohort eligibility and ranked Swipe Decks.
- Rich, portfolio-first Candidate Cards with disclosure-controlled visualizations, obscured Photos, and synthetic SnapTrade-shaped data.
- One-stage left/right swiping, Mutual Match, and Photo reveal.
- Trust language, adult-only positioning, and explicit demonstration boundaries.
- Branded, responsive Public Site with product pitch, About, Terms, Privacy, trust and safety, contact or support, and entry into the demonstration.

### 6.2 Out of Scope for MVP

- **One-to-one chat:** high-value stretch goal; requires unmatch, block, report, moderation, and anti-solicitation controls before inclusion.
- **Public signup and real candidates:** deferred until commercial permission, privacy, security, safety, moderation, and legal gates are satisfied.
- **Persistent guest accounts:** the optional Guest Demo is ephemeral and synthetic; it does not create a second account model.
- **Native mobile distribution:** responsive hosted experience is sufficient to prove the concept.
- **Monetization:** no pricing, subscription, advertising, or paid access model is defined.
- **User-defined disclosure policies and unrestricted custom fields:** the demonstration uses a focused, designed set of levels even though those levels may show rich detail.
- **Dynamic market-moment signals:** deferred until core source-data visualizations and consent comprehension are validated.
- **Production analytics and operations:** only enough observability to diagnose the demonstration is required.

### 6.3 Post-MVP Direction, Not Commitment

The following groups preserve the product direction without committing the current milestone to implement them. Planning may deliberately stop after the core demonstration.

- **Richer discovery:** more expressive Disclosure Levels, additional initiation rights, deeper Portfolio Visualizations, and more sophisticated compatibility signals.
- **Evolving candidate market:** Swipe Deck eligibility and ordering that respond over time to portfolio characteristics, activity, engagement, Proximity, preferences, disclosure, and appropriately bounded market context.
- **Timely and playful relevance:** shared market experiences or events that create dating-context relevance without becoming investment advice, performance shaming, or a path around consent.
- **Communication and trust:** post-match chat only with the associated block, report, unmatch, anti-solicitation, moderation, identity, retention, and account-lifecycle controls.
- **Public product readiness:** the commercial, privacy, security, legal, safety, and operational gates in §10.

## 7. Success Metrics

These metrics evaluate the demonstration as a product proof, not as evidence of market demand or production readiness.

### 7.1 Primary Metrics

- **SM-1 — End-to-end proof completion:** 100% of scripted review runs can complete connect, inspect live portfolio data, preview the live owner card, set preferences, enter Discovery, swipe, reach a Mutual Match, reveal a Photo, and disconnect without manual data repair. Validates FR-1 through FR-21. `[ASSUMPTION: Scripted review runs are the primary acceptance method for this milestone.]`
- **SM-2 — Material portfolio influence:** In every seeded comparison test, changing a material portfolio-derived input or compatibility preference changes at least one candidate's eligibility, order, or stated comparison. Validates FR-5, FR-8, and FR-10.
- **SM-3 — SnapTrade demonstration breadth:** Every available core dataset—accounts, balances or values, holdings or positions, and activities or orders—is visibly represented in the Portfolio Showcase or explicitly shown as unavailable, empty, or unsupported; connection and Freshness State are also demonstrated. Validates FR-2, FR-3, and FR-19 through FR-21.
- **SM-4 — Verification comprehension:** At least 80% of a small formative review group can correctly explain that Verified means consented source data at a stated freshness point—not complete wealth or financial health. Validates FR-1, FR-3, FR-11, and FR-17. `[ASSUMPTION: The team can recruit at least five representative reviewers for formative testing.]`

### 7.2 Secondary Metrics

- **SM-5 — Disclosure comprehension:** At least 80% of formative reviewers can predict which portfolio fields are hidden, bucketed, derived, or exact under each prototype Disclosure Level. Validates FR-6, FR-11, and FR-20.
- **SM-6 — Privacy boundary integrity:** Zero test cases expose live owner data outside the private Portfolio Showcase or card preview, expose details beyond the selected Disclosure Level, reveal exact location or account identifiers, or reveal a Photo before Mutual Match. Validates FR-6, FR-9, FR-11, FR-15, and FR-20.
- **SM-7 — Lifecycle clarity:** At least 80% of formative reviewers can identify the current Freshness State and the effect of disconnecting. Validates FR-3, FR-4, FR-13, and FR-19.
- **SM-8 — Portfolio-first perception:** A majority of formative reviewers identify the Portfolio Visualizations—not the obscured Photo—as the primary basis for the initial swipe and can name at least two SnapTrade-backed data categories demonstrated by the experience. Validates FR-11, FR-12, FR-19, and FR-21.
- **SM-9 — Synthetic population depth:** The demonstration supplies credible multi-card Swipe Decks across every supported compatibility mode and Disclosure Level, while also reproducing the planned sparse, boundary, and match-outcome scenarios without real financial records. Validates FR-9, FR-10, and FR-22.
- **SM-10 — Profile-preview fidelity:** For every tested Disclosure Level and identity state, the phone and desktop Profile Preview shows the same permitted information and hides the same prohibited information as the corresponding Candidate Card or matched state. Validates FR-20, FR-23, and FR-24.
- **SM-11 — Public product completeness:** Every required Public Site surface is reachable without authentication on phone and desktop, communicates the demonstration boundary, and offers a clear demonstration entry or contact path. Validates FR-25 and FR-26.
- **SM-12 — Guest isolation, if shipped:** No Guest Demo test can initiate OAuth, retrieve live owner data, mutate owner state, or persist a real-user account. Validates FR-2 and FR-27.

### 7.3 Counter-Metrics and Guardrails

- **SM-C1 — Wealth-status misinterpretation:** Do not improve engagement by turning exact values or holdings into a universal status score. Any reviewer interpretation that Findur certifies wealth or responsibility is a comprehension defect, not a positioning win. Counterbalances SM-8.
- **SM-C2 — Disclosure pressure:** Do not improve deck size or match completion by weakening a user's Disclosure Level, surfacing hidden data, or coercing increased disclosure. Counterbalances SM-1 and SM-5.
- **SM-C3 — Freshness concealment:** Do not improve apparent connection success by hiding stale, failed, or reauthorization-required states. Counterbalances SM-1 and SM-7.
- **SM-C4 — Synthetic realism confusion:** Do not make the demonstration feel more convincing by allowing reviewers to mistake Synthetic Candidates for real people with connected financial accounts. Counterbalances SM-1.

### 7.4 Formative Signals, Not Acceptance Gates

For every formative review session, capture perceived Candidate Card relevance, comfort with the Disclosure Levels, progression toward a Mutual Match, and any safety or trust concern. Report the distribution and qualitative themes rather than treating the small demonstration sample as a market-validation metric.

## 8. Cross-Cutting Non-Functional Requirements

### 8.1 Privacy and Security

- **NFR-1:** Secrets, OAuth credentials, and tokens must never be committed to the repository, exposed to client-side code, or included in logs. Live portfolio payloads must not be committed or logged.
- **NFR-2:** Findur must limit collected and client-delivered financial data to fields with a defined Portfolio Showcase, card-preview, matching, disclosure, or lifecycle purpose; the authenticated owner may receive the live view data required by FR-19 and FR-20.
- **NFR-3:** Data access must default to least privilege and read-only behavior; no trading capability is permitted.
- **NFR-4:** Logs and diagnostics must identify lifecycle failures without recording raw holdings, balances, transactions, access tokens, or unnecessary personal data.
- **NFR-5:** Synthetic Candidate data and the owner-controlled live portfolio must remain distinguishable in storage, processing, and presentation.

### 8.2 Reliability and Observability

- **NFR-6:** The hosted demonstration must expose actionable diagnostic states for authorization, callback, sync, derivation, deck generation, and disconnect failures.
- **NFR-7:** Repeated callbacks, refresh events, swipes, and match creation must be idempotent where duplication could corrupt the demonstration state.
- **NFR-8:** Provider unavailability or incomplete source data must degrade to an honest recoverable state, never fabricated currency or false freshness.

### 8.3 Accessibility and Responsive Use

- **NFR-9:** Core flows must be operable by keyboard and expose meaningful labels, focus states, error text, and non-color indicators.
- **NFR-10:** Every core flow and Public Site surface must remain usable on contemporary phone and desktop viewport sizes, with functional parity for required information, disclosure rules, and actions even when layouts differ.
- **NFR-11:** Photo obscuring, portfolio visuals, and swipe controls must have accessible alternatives that communicate equivalent state and actions.

### 8.4 Performance

- **NFR-12:** Cached or synthetic Discovery state should become interactive within three seconds under ordinary broadband conditions, excluding third-party authorization or provider synchronization time. `[ASSUMPTION: Three seconds is an appropriate demonstration target until architecture establishes a measured budget.]`
- **NFR-13:** Long-running provider work must expose progress or a pending Freshness State rather than block the interface without feedback.

## 9. Information Architecture, Platform, and Tone

### 9.1 Required Product Surfaces

UX must provide coherent navigation across the Public Site, Personal Profile, connection lifecycle, Portfolio Showcase, Profile Preview, preferences and disclosure, Discovery, Mutual Match, and trust or privacy information. Relevant failure, empty, freshness, reauthorization, and disconnect states must remain reachable and understandable. Final navigation and screen decomposition belong to UX design.

### 9.2 Platform

The milestone is a responsive hosted experience. Native iOS and Android applications are not required.

### 9.3 Aesthetic and Tone

Findur should feel enjoyable enough to be recognizably about dating and professional enough to deserve trust with financial data. Product language should be direct, calm, non-judgmental, and explicit about uncertainty. It must avoid luxury-status cues, financial shaming, trading hype, and language that reduces people to portfolio value.

## 10. Future Public-Launch Gates

The following gates are not requirements for operating the owner-only demonstration, but they are mandatory prerequisites before public multi-user use:

1. Written confirmation that the intended SnapTrade plan and terms permit portfolio-derived attributes from one real user to affect another real user's experience.
2. Privacy and legal review for target jurisdictions, including consent, disclosure, deletion, retention, derived data, market-data rights, and consumer-protection claims.
3. Production security review covering authorization, secrets, credential storage, session boundaries, abuse resistance, logging, and incident response.
4. Identity and adult-age controls appropriate to a dating product.
5. Blocking, reporting, financial-disclosure suppression, moderation workflows, anti-scam and anti-solicitation controls, and accountable safety operations.
6. Production broker coverage, refresh behavior, degraded states, support model, and data-quality monitoring.
7. Validated Disclosure Levels, screenshot/inference risks, accessibility, and comprehension through representative user research.
8. App-store policy review if native distribution is pursued.

## 11. Open Questions

These questions do not block the demonstration unless noted; they must be resolved before the related future scope begins.

1. Which Portfolio-Derived Signals best communicate similar, diversified, and complementary compatibility without implying advice or worth? **Owner:** product + UX. **Revisit:** during `bmad-ux`, before Candidate Card design is approved.
2. What precise source-age thresholds move a signal among current, stale, and unusable Freshness States? **Owner:** architecture. **Revisit:** during `bmad-architecture`, before the data lifecycle is approved.
3. Which lightweight explanation best communicates why a Synthetic Candidate appears without exposing ranking or Invisible Cohort logic? **Owner:** UX. **Revisit:** during `bmad-ux`, before Discovery copy and interaction design are approved.
4. What evidence should be collected from formative reviewers, and who will recruit them? **Owner:** product owner. **Revisit:** before scheduling formative review sessions.
5. Before any public launch, what commercial approval and legal characterization are required for cross-user portfolio-derived ranking and disclosure? **Owner:** product owner + qualified counsel/SnapTrade contact. **Revisit:** before any real-candidate or public-launch scope is approved.
6. If chat becomes a milestone stretch, what minimum block, report, moderation, unmatch, and anti-solicitation bundle must ship with it? **Owner:** product + safety + UX. **Revisit:** before chat enters an epic or story.
7. What user action and product language are sufficient to characterize portfolio derivation as user-initiated under the applicable SnapTrade policy and intended use? **Owner:** architecture + product owner. **Revisit:** during architecture before implementing portfolio derivation; seek written provider confirmation if ambiguity remains.
8. Which additional Personal Profile fields, if any, are required for the demonstration beyond the minimum categories in FR-23? **Owner:** UX + product. **Revisit:** during `bmad-ux`, before Profile setup is approved.
9. Who owns review and approval of Terms, Privacy, and trust and safety content before the hosted demonstration is shared externally? **Owner:** product owner. **Revisit:** before external hosting or sharing; obtain qualified review where required.
10. Can the Guest Demo be delivered without materially expanding session, security, or schedule scope? **Owner:** architect. **Revisit:** during architecture sizing and decide before epics and stories.

## 12. Assumptions Index

- **A-1 (§7.1, SM-1):** Scripted review runs are the milestone's primary acceptance method. **Owner:** product. **Revisit:** before QA/test planning.
- **A-2 (§7.1, SM-4):** At least five representative formative reviewers can be recruited. **Owner:** product owner. **Revisit:** before formative review scheduling.
- **A-3 (§8.4, NFR-12):** Three seconds is a reasonable provisional interactivity target for cached or synthetic Discovery state. **Owner:** architecture. **Revisit:** during performance-budget definition.
- **A-4 (§4.3, FR-23):** Display name, adult-age confirmation, coarse location, discovery eligibility or preferences, relationship intent, at least one Photo, and a short biography or prompt response form the minimum viable Personal Profile. **Owner:** UX + product. **Revisit:** during profile-design approval.

## 13. Source and Evidence Notes

- Product decisions come from the completed product brief and addendum linked in §0.
- The user's clarification that cards should demonstrate rich portfolio visualization supersedes the PRD's earlier derived-traits-and-bands-only MVP interpretation; it does not replace the brief's consent and disclosure framing. The reconciliation is recorded in [reconcile-snaptrade-showcase.md](reconcile-snaptrade-showcase.md).
- SnapTrade constraints come from the technical feasibility research linked in §0 and must be refreshed before 2026-10-01 where that report marks claims as time-sensitive.
- Current market and safety grounding is summarized in [research-market-landscape.md](research-market-landscape.md). It supports differentiation and launch-gate requirements but does not prove demand.
