---
title: Findur PRD Addendum
status: final
created: 2026-09-19
updated: 2026-09-19
---

# Findur PRD Addendum

This addendum preserves source-backed depth that should inform UX and architecture without turning the PRD into an implementation design. It is subordinate to [prd.md](prd.md); conflicts must be resolved in favor of the PRD or recorded as a new decision. The finalized UX spines control visual and interaction detail, but they do not independently override product scope, disclosure, safety, or acceptance.

**Current boundary:** live SnapTrade data belongs only to the protected owner and may appear in the private Portfolio Showcase and Profile Preview. Discovery uses synthetic data. Guest Demo and public multi-user use remain out of scope.

## 1. Product Model Details

### 1.1 Portfolio-Governed Exclusivity

Findur's exclusivity is not a universal wealth threshold. A Usable Portfolio grants entry to the differentiated experience. Discovery and initiation rights then vary through standardized disclosure and compatibility rules. The product should avoid named status tiers, public cohort membership, or anything that resembles a wealth leaderboard.

### 1.2 Private Matching Versus Visible Disclosure

The system may use the consented data from Included Accounts privately for matching, subject to purpose limitation and minimization. OAuth Connection, Account Inclusion, private use, owner-only display, and candidate disclosure are separate permissions. A viewer receives only the source fields, Portfolio Visualizations, and Portfolio-Derived Signals allowed by the candidate's saved Disclosure Level. That level may permit selected exact details—such as security names, weights, values, defensible performance context, activities, orders, or transactions—rather than limiting every card to abstract bands.

For this milestone, the only live data belongs to the authenticated owner and may be shown in exact form only in the private Portfolio Showcase and the owner's card preview. Discovery uses visually equivalent synthetic data. This preserves the ability to demonstrate rich card experiences without representing real cross-user financial disclosure as approved.

The product contract is:

| Disclosure Level | Permitted portfolio detail |
| --- | --- |
| Snapshot | Allocation, asset classes, diversification context, value and activity bands, and freshness; no security names or exact monetary values. |
| Holdings | Snapshot plus security names, position weights, percentage performance, and activity categories; no quantities, exact monetary values, or transaction detail. |
| Full Detail | Holdings plus exact quantities and values, portfolio or account value, performance amounts, and recent activity, order, or transaction detail. |

Regular Discovery includes only candidates whose level is equal to or lower than the viewer's level. A higher-disclosure user may initiate toward a lower-disclosure user; the recipient may evaluate the initiator at the initiator's level without increasing their own Disclosure Level or exposing more of their own data. UX may refine labels only through an explicit product change; it may not alter the field contract or initiation asymmetry silently.

No level is selected or saved by default. Snapshot is recommended but not preselected. Until Save succeeds, a selection changes only the preview, which remains visibly marked as unsaved. Cancel restores the saved level, and attempting to leave the route with unsaved changes requires a discard-or-stay decision. Privacy-decreasing saves fail closed: broader representations—whether active, prefetched, cached, offline, or restored from history—are suppressed immediately and remain suppressed while the system retries any failed purge or persistence operation.

Performance may appear only when its period, coverage, currency treatment, inputs, and Freshness State are defensible. Unsupported performance is shown as unavailable, never inferred. Optional tax lots remain owner-only and outside first-cut Candidate Cards.

The disclosure design must answer:

- How are Snapshot, Holdings, and Full Detail communicated without turning them into status labels?
- What does the user preview before selecting a level?
- How does the experience explain downgrade, disconnect, block, or report effects?
- How does the interface communicate symmetry where it exists and intentional asymmetry where it does not?

### 1.3 Candidate Selection

Candidate selection is expected to combine:

- portfolio composition and diversification characteristics;
- value bands rather than exact values;
- activity recency where supported and appropriately qualified;
- the user's preference for similar, diversified, or complementary portfolios;
- maximum distance and nearer-is-stronger Proximity;
- Disclosure Level compatibility and initiation eligibility;
- candidate activity or engagement where synthetic demonstration scenarios need it; and
- potential use of market context, but only when it remains understandable, timely, and non-advisory.

Invisible Cohorts are dynamic implementation concepts, not destinations or identities. They should never appear as user-visible classes such as “elite,” “high value,” or “conservative investor.” The system should offer a human explanation of relevance without disclosing the precise model or making deterministic compatibility claims.

### 1.4 Progressive Identity Reveal

The Photo reveal is the payoff for mutual interest, not the start of a second screening step. The experience should retain one initial swipe decision. After a Mutual Match, the Photo becomes visible and the product acknowledges the match. Adding a second accept/reject decision after the reveal would undermine the portfolio-first premise and should be treated as a product change requiring explicit review.

### 1.5 Timely and Playful Candidate Pools

Event-driven candidate relevance remains a future product direction, not part of the core demonstration commitment. A shared market experience—for example, candidates whose portfolios went through a comparable drawdown—could create timely, playful context if it does not reveal sensitive performance or imply advice.

Before this direction enters scope, product and UX must define cohort entry and exit, source freshness, user controls, explanation, explicit consent, Disclosure Level interaction, and protections against performance exposure or financial shaming. Event-driven cohorts remain internal and non-browseable.

## 2. UX Design Handoff

### 2.1 Personal Profile and Preview

UX should define the smallest credible dating profile rather than treating the Connected Portfolio as the whole person. The assumed first-cut minimum is display name, adult-age confirmation, coarse location, discovery eligibility or preferences, relationship intent, at least one Photo, and a short biography or prompt response. UX may refine labels and validation and may split fields, but product must confirm the field set before accepting the profile story; additional required fields need an explicit product change.

Profile Preview is the owner's truth surface. It should combine personal information and live portfolio presentation and let the owner inspect:

- each Disclosure Level;
- pre-match Candidate Card presentation;
- post-match identity and profile presentation;
- hidden, bucketed, derived, and exact portfolio fields; and
- representative phone and desktop layouts.

Phone and desktop may use different compositions or interactions, but they must preserve information, consent, freshness, and progressive-reveal rules. UX should define responsive behavior, breakpoints, card proportions, visualization fallbacks, and whether desktop preview presents a device frame or the true responsive layout.

### 2.2 Candidate Card Information Hierarchy

The Candidate Card should be a showcase for portfolio visualization while remaining legible as a dating interface. UX should explore a hierarchy such as:

1. a visually dominant portfolio composition, holdings, value/performance, or activity view;
2. concise compatibility explanation;
3. permitted named securities, position details, activity/order details, or value/performance context;
4. coarse Proximity and Freshness State;
5. obscured Photo and limited non-financial identity context; and
6. a clear single swipe action.

This is a hierarchy to test, not a mandated screen layout. The owner's private preview should not collapse live SnapTrade-backed data into a decorative badge, and Discovery cards should give equally rich treatment to clearly synthetic, SnapTrade-shaped portfolio data.

Candidate Detail is a protected, deep-linkable expansion of the card rather than a separate disclosure context. It may use a full-screen phone composition and a two-column desktop composition, but it preserves synthetic provenance, Disclosure Level, Freshness State, and persistent Pass/Interested actions. Back restores the exact safe position in the Swipe Deck. Opening an unavailable or gated deep link takes the user to a named recovery destination without exposing protected content. Incoming Interest is a child surface of Discovery, not an inbox or fourth primary area.

### 2.3 Visualization Strategy

The same visualization system should power both the owner's live card preview and synthetic Candidate Cards so the API proof and dating experience reinforce each other. Candidate Cards are an opportunity to demonstrate multiple SnapTrade data categories through interactions such as:

- allocation or composition charts from holdings or positions;
- expandable named-holding views with weights, quantities, or values where permitted;
- total or banded account value with clear coverage boundaries;
- performance context with time period and source freshness;
- recent activity, order, or transaction timelines; and
- compatibility comparisons that distinguish source facts from derived interpretation.

The UX should select a coherent subset rather than placing every dataset on one card. Card states or lightweight exploration may reveal depth while maintaining a clear swipe decision. Accounts, positions, balances, orders, and activities remain distinct; dataset-specific freshness is visible; currencies are never silently aggregated; and unavailable data is not synthesized.

### 2.4 Public Site, Brand, and Product Narrative

The unauthenticated experience should feel like a credible product, not an OAuth test harness. UX should define a responsive landing page and persistent navigation or footer system covering the pitch, how it works, About, Terms, Privacy, trust and safety, contact or support, and entry into the demonstration.

The public narrative should explain the portfolio-first premise and progressive identity reveal before asking for financial connection. It must distinguish the owner-only demonstration from a public dating launch. Visual identity and copy should feel enjoyable enough for dating and professional enough for connected financial data without relying on luxury-status cues or generic fintech branding.

Legal text requires appropriate review; UX is responsible for readable placement, hierarchy, navigation, and consent moments, not for inventing legal conclusions.

### 2.5 Required Comprehension Tests

Formative review should verify that participants understand:

- OAuth Connection, Account Inclusion, private derivation, and visible Disclosure Level are separate permissions.
- Connected Portfolio data covers only confirmed Included Accounts and may still be incomplete.
- Verified does not mean complete wealth, financial health, responsibility, or identity.
- Private matching inputs and visible disclosures are different.
- The private Portfolio Showcase, owner's card preview, and Discovery cards have different live-versus-synthetic and visibility boundaries.
- Disclosure Level changes access without creating a requirement to reveal more.
- Freshness State qualifies every time-sensitive claim.
- Synthetic Candidates are fictional and do not represent real connected users.
- A Photo reveal follows the original mutual swipe and does not require another decision.

### 2.6 Tone and Anti-Patterns

The experience can use playful market language only when it does not trivialize loss, shame performance, suggest trading action, or expose sensitive data. Avoid:

- “verified wealthy,” “high-value person,” “smart money,” or equivalent labels;
- red/green performance theater that reads as a score of the person;
- luxury imagery that converts compatibility into status ranking;
- false precision in compatibility percentages;
- urgency based on market movement or fear of missing out; and
- copy implying that a Connected Portfolio is complete or continuously current.

### 2.7 Future Chat Boundary

Chat is a stretch goal only. It should not be added as an isolated text box. At minimum, the feature bundle must include controls to unmatch, block, and report; rules for financial solicitation; disclosure suppression after safety actions; moderation routing; and an accountable response model. If the project cannot support that bundle, the matched state should end without messaging.

### 2.8 Navigation, Locale, Theme, and Installation

The authenticated information architecture has three primary areas: Discovery, Portfolio, and Profile. Discovery owns Candidate Detail, Incoming Interest, Mutual Match return paths, and deck recovery. Portfolio owns OAuth lifecycle, Account Inclusion, source status, revalidation, and Portfolio Showcase. Profile owns Personal Profile, Profile Preview, preferences, saved Disclosure Level, language, theme, and notification settings. Public and legal surfaces remain outside the authenticated shell.

Discovery is the default authenticated destination for an eligible owner. Incomplete prerequisites route to a focused recovery surface; OAuth returns confirm outcome before continuing; permitted deep-link intent is restored after gates; and back, refresh, and revalidation preserve only state that remains authorized.

English and French language support and System, Light, and Dark theme support are first-cut parity requirements. UX owns copy fit, locale formatting, theme-safe states and charts, and the 35% label-expansion allowance. The hosted web app is installable and app-like on mobile, but installation does not imply offline access to sensitive portfolio or candidate data. Exact breakpoints, tokens, icon sizes, manifest fields, and platform asset matrices remain UX or architecture detail.

## 3. Architecture Handoff

### 3.1 Authorization and Session Boundary

The technical research favors a SnapTrade OAuth application owned by a Commercial account and serving Personal SnapTrade users. This model uses the authorization code flow with PKCE and keeps confidential-client operations on the server. Architecture should validate exact provider requirements rather than copying this statement as code.

The design should specify:

- exact redirect URI registration and callback validation;
- state, nonce, PKCE, and ID-token validation as applicable;
- encrypted server-side credential and token storage;
- session creation and expiration;
- minimum required scopes;
- disconnect, revocation, deletion, and reauthorization behavior; and
- protection against callback replay and account mix-ups.

Consent has two stages. OAuth continuation authorizes access to the named provider and retrieval of only the minimum metadata needed to list eligible accounts. A separate post-OAuth confirmation authorizes Findur to retrieve, derive from, match with, and preview the selected Included Accounts. No account is included by default, and “connected” must never be treated as “included.” Architecture must verify provider-supported account granularity; where the provider cannot enforce it, Findur still enforces the product boundary independently.

Protected direct entry, authentication gating, OAuth return, refresh, back navigation, and safe fallback behavior must be defined for every applicable route. Intended destinations may be restored only after authentication and prerequisite gates succeed.

The owner must have a SnapTrade Personal account with a supported brokerage already connected before Findur can retrieve portfolio data. OAuth scope selection should be justified field by field: request `openid`, `email`, and `read` only where needed, and request `webhook` only if the architecture actually uses lifecycle notifications. Any feature that could be characterized as portfolio analysis or a signal must document the user action that initiates it and remain subject to written SnapTrade confirmation for cross-user use.

Architecture should model one first-cut entry context: a protected, preconfigured owner session that alone may initiate OAuth and view live data. Guest Demo is deferred. Any future synthetic-only guest context must be sized separately and must not introduce public signup, durable guest accounts, or any route to owner tokens, payloads, caches, or state.

### 3.2 Data and Derivation Boundary

Architecture should make the following layers explicit:

1. authorized source data;
2. normalized portfolio data;
3. private Portfolio-Derived Signals with provenance and Freshness State;
4. candidate-ranking inputs;
5. disclosure-approved source fields, derived fields, and visualization models; and
6. rendered or cached Candidate Card data.

Each transition should have a declared purpose and deletion behavior. Demonstrating useful SnapTrade capabilities is a valid product purpose for the owner-only Portfolio Showcase and card preview, but it does not automatically authorize cross-user disclosure. Account exclusion, Disclosure Level downgrade, or disconnect must immediately suppress broader use and invalidate active UI state, prefetched data, rendered representations, server and client caches, offline output, and navigation-history restoration. If a purge or persistence operation fails, the system retries it without restoring broader state. If adding an account or broadening permission fails during retrieval, persistence, or recalculation, the prior narrower committed set remains authoritative until the broader update succeeds and is confirmed. Any later broadening requires scoped confirmation.

### 3.3 Freshness and Events

Source types have different recency characteristics. Activities may update daily, orders may be more current but cover a limited period, holdings freshness varies by plan and brokerage, and webhooks may signal synchronization lifecycle events rather than provide holdings-change diffs. Architecture should therefore:

- attach source and observation timestamps to derived signals;
- avoid a single misleading global “live” flag;
- define current, syncing, stale, reauthorization-required, failed, and disconnected states;
- make derivation repeatable after refresh; and
- prevent stale cached presentation from surviving a consent or safety change.

Safe stale-while-revalidate behavior may preserve the last trustworthy authorized view with visible dataset-specific freshness while refresh proceeds nondisruptively. Account exclusion, Disclosure Level downgrade, disconnect, or any other permission or safety invalidation overrides that behavior immediately.

### 3.4 Synthetic Data Isolation

The demonstration should make it impossible to confuse Synthetic Candidate fixtures with real connected accounts at the data model level. Suggested safeguards include a required provenance field, separate fixture generation, prohibition of production-looking credentials, and presentation cues appropriate to the review context. Synthetic datasets must contain no copied real financial records.

The Synthetic Population should be generated rather than maintained as hundreds of hand-authored fixtures. The generator should:

- accept a stable seed and configurable population size;
- produce internally coherent profiles, portfolios, disclosure settings, locations, engagement states, and swipe outcomes;
- use scenario tags to guarantee coverage of ranking boundaries, sparse cohorts, disclosure combinations, stale or unsupported data, and reciprocal or non-reciprocal matches;
- vary portfolio structures without encoding claims that demographic traits determine wealth or investing behavior;
- emit stable identifiers so seeded test expectations remain repeatable; and
- validate that generated profiles contain no imported personal or financial records.

Population size is a downstream sizing decision. A provisional starting point of roughly 500 profiles may be useful for implementation planning, but architecture and test design should select the smallest count that provides credible deck depth, scenario coverage, and acceptable performance. It is neither a product requirement nor a fixed architectural ceiling.

### 3.5 Observability

The demonstration needs enough observability to diagnose:

- authorization initiation and callback outcome;
- source connection and synchronization state;
- derivation success or missing inputs;
- Swipe Deck generation and empty results;
- swipe and Mutual Match state transitions; and
- disconnect, deletion, and reauthorization actions.

Events should use opaque test-user identifiers and categorical outcomes. Raw balances, holdings, transactions, tokens, authorization codes, and unnecessary personal information must not appear in logs.

### 3.6 Notifications and Installed-App Badging

Mutual Match, Incoming Interest, and connection-action-required events may drive optional browser notifications and installed-app badges. In-product event state is authoritative; denied permission, unsupported platforms, and delayed, duplicate, or failed delivery must not cause the event to be lost. Events persist until opened, dismissed, or invalidated. Acknowledging an event clears its unread state and badge contribution without deleting the underlying Mutual Match or recovery state. Permission requests are contextual and user-initiated. Categories are separately configurable, and opening a notification passes through authentication and prerequisite gates before restoring a safe destination.

Payloads, previews, badges, titles, URLs, service-worker caches, launch screens, and task-switcher representations contain no candidate identity, portfolio, holding, value, performance, location, or other sensitive detail. Architecture should implement delivery, deduplication, and platform fallback without changing the product-defined contract for persistence, dismissal, invalidation, acknowledgment, or badge clearing and without creating a separate notification inbox.

## 4. Risk Notes

### 4.1 Financial Targeting and Romance Scams

Portfolio visibility may make a dating user more attractive to scammers or enable tailored investment solicitation. Even derived bands can support inference when combined with identity, occupation, location, or screenshots. The public product therefore requires safety design before enabling Discovery, not only after adding messaging.

### 4.2 Screenshot and Inference Risk

Disclosure controls govern what Findur renders; they cannot prevent screenshots, memory, or inference. UX research should test whether combinations of traits, bands, charts, and context inadvertently reveal a narrow range of value or identifiable holdings. Data minimization must evaluate combinations, not fields in isolation.

### 4.3 Sparse-Pool Pressure

Strict Disclosure Level and Proximity rules may create sparse Swipe Decks, especially for privacy-conservative users. The product must not silently relax consent or eligibility to improve inventory. It may explain scarcity, invite the user to voluntarily change a setting, or end the deck honestly.

### 4.4 Commercial and Policy Uncertainty

The current technical research supports the owner-only test demonstration but does not establish permission for cross-user production use. The test OAuth application is limited to five users; production OAuth credentials require authorization from SnapTrade. Neither fact permits expansion into a public or real-candidate pilot. Before launching a public product, obtain written confirmation concerning portfolio-derived compatibility, ranking, disclosure, redistribution, and market-data terms.

Re-check the OAuth model, scopes, data features, plan limits, and broker coverage by 2026-10-01 and before any credential or scope change. Re-check SnapTrade's developer terms and compliance policy before any public prototype or production launch and whenever either policy changes. Also re-check app-store and jurisdictional requirements during launch planning.

## 5. Options Deferred for Later Product Work

The source material intentionally leaves the following choices open:

- any future expansion beyond the fixed first-cut Disclosure Level names and bundles;
- the ranking model and weight allocation;
- whether market-moment signals improve delight without becoming advice or performance exposure;
- a public identity and age-assurance model;
- retention periods for source and derived data;
- production reporting and moderation operations;
- monetization and whether any paid model can avoid selling access to people or their financial data;
- the definition and scope of post-match communication;
- the timing and scope of a signup-free Guest Demo; and
- and SnapTrade's hosted MCP server as a possible future protected-owner evidence-retrieval option, subject to provider and architecture review and Findur-owned disclosure and safety controls.

These are not omissions to fill opportunistically during implementation. Each requires an explicit product decision and, where relevant, UX research, legal review, or commercial approval.
