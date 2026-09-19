# Findur UX Source Extract: Product Context

Source-only extraction for UX discovery. This document records what the four supplied product sources say, distinguishes requirements from future direction, and does not resolve open product decisions.

## Source Set and Precedence

- `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md` is the completed high-level Product Brief. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#product-brief-findur`)
- `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md` is a decision-and-question handoff; declarative statements are intended decisions, while *may*, *should*, *proposed*, and *remain to be defined* mark downstream questions. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#governing-trust-constraint`)
- `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md` is the final PRD for the first publicly hosted demonstration and carries globally stable Functional Requirement IDs. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#0-document-purpose`)
- `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md` preserves depth for UX and architecture but is explicitly subordinate to the PRD; conflicts resolve in favor of the PRD or require a new decision. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#findur-prd-addendum`)

## Product Intent and Promise

- Findur is a portfolio-first dating experience for everyday investors who want financial and investing alignment to materially shape romantic discovery. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#executive-summary`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#1-vision`)
- It reverses conventional photo-led dating: a consented Connected Portfolio is the first candidate surface, the portfolio visualization is the visual hero, and the Photo stays obscured until Mutual Match. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#the-problem`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#1-vision`)
- Portfolio data must change the market itself—not act as a decorative badge—by affecting candidate eligibility, ordering, comparison, disclosure rights, and relevance. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#product-promise`)
- The product uses connected financial data as evidence for compatibility, not as a measure of human worth. It does not certify wealth, stability, responsibility, identity, or compatibility. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#1-vision`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#5-non-goals`)
- The milestone proves that the concept can become a coherent, trustworthy experience and visibly demonstrates the breadth of the SnapTrade integration; it is not a public dating launch or market-demand proof. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#1-vision`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#7-success-metrics`)
- SnapTrade has two central product roles: matching intelligence and trust grounded in consented connected-account data, coverage, and freshness. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#product-promise`; `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#governing-trust-constraint`)
- The product must preserve human dating context: the Connected Portfolio shapes discovery but does not replace the Personal Profile. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#43-personal-profile-and-complete-preview`)

## Target Users, Audience, and Stakes

### Primary user

- The primary user is an everyday investor who values evidence and financial openness, wants investing compatibility to inform dating, will connect a portfolio for a more relevant experience, and expects control over what other people see. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#21-primary-user`)
- Core jobs are to assess investing compatibility through evidence, select similar/diversified/complementary preferences, control disclosure without blocking private matching, understand freshness and verification limits, let attraction emerge before appearance, present as a whole person, preview the exact profile presentation, and withdraw consent with understood consequences. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#22-jobs-to-be-done`)

### Milestone audience and access

- One protected, preconfigured owner may authenticate, initiate SnapTrade test OAuth, use live portfolio data, and view private owner routes. Discovery candidates are generated and synthetic. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#0-document-purpose`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#42-connected-portfolio-showcase-and-card-preview`)
- Unauthenticated public visitors can inspect the branded Public Site and trust/legal information before choosing protected owner entry or, only if architecture approves it, an optional Guest Demo. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-26-provide-expected-informational-and-legal-surfaces`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-27-offer-a-signup-free-synthetic-trial-conditional-stretch`)
- The optional Guest Demo is signup-free, ephemeral, synthetic-only, isolated from the owner, and cannot initiate OAuth, access the Connected Portfolio, persist a real-user account, or reach owner routes/data. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-27-offer-a-signup-free-synthetic-trial-conditional-stretch`)
- Non-users for the milestone include the general public seeking a working multi-user dating service, anyone under 18, users seeking investment advice/trading/wealth certification/exact records, and real candidate users other than the protected owner. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#23-non-users-for-this-milestone`)

### Stakes and regulatory posture

- This is a consumer dating concept handling sensitive connected financial data, with risks of romance scams, tailored investment solicitation, financial targeting, screenshots, inference, harassment, and performance shaming. Safety design is needed before Discovery, not only after messaging. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#41-financial-targeting-and-romance-scams`; `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#tiered-disclosure-and-initiation`)
- “Verified” is deliberately narrow: information derived from accounts the user chose to connect, as of a stated freshness point. It does not mean complete net worth, every account, financial health, financial responsibility, or identity. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary`; `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#product-promise`)
- Findur is read-only and non-advisory: no trades, investment recommendations, brokerage actions, portfolio-management tools, real-time market processing, or universal financial-worth score. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#5-non-goals`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#81-privacy-and-security`)
- The hosted demonstration must state its adult-only 18+ boundary, use unambiguously adult Synthetic Candidates, and surface anti-solicitation, anti-money-request, anti-targeting, consent, and verification guidance before or during Discovery—not only in legal text. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#48-demonstration-safety-and-trust`)
- Public multi-user use is gated on commercial permission, privacy/legal review, production security review, identity and adult-age controls, reporting/blocking/moderation/anti-scam operations, support and data-quality operations, validated disclosure/accessibility/comprehension, and app-store review if applicable. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#10-future-public-launch-gates`)
- Terms, Privacy, and trust-and-safety content require appropriate/qualified review; UX owns readable placement, hierarchy, navigation, and consent moments rather than legal conclusions. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#24-public-site-brand-and-product-narrative`)

## Exact Product Terminology

Use source-defined capitalization and meanings unless a later explicit product change says otherwise.

| Term | UX-relevant meaning | Source |
| --- | --- | --- |
| Candidate Card | Discovery card for one Synthetic Candidate; only level-permitted information; Photo obscured until Mutual Match. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Connected Portfolio | User-authorized accounts/data read through SnapTrade; not complete wealth or financial health. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Disclosure Level | One of Snapshot, Holdings, or Full Detail; controls visible source fields/signals and initiation eligibility, not private consented calculations. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Discovery | Surface for evaluating Candidate Cards and swiping. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Freshness State | Source/recency context: connected, syncing, stale, needs reauthorization, failed, or disconnected. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Guest Demo | Conditional, signup-free, ephemeral, synthetic-only trial with no OAuth, Connected Portfolio, or persisted real-user account. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Invisible Cohort | System-managed, non-browseable eligibility group used to assemble a Swipe Deck; names, rules, and membership stay hidden. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Mutual Match | State created by positive decisions on both sides; triggers Photo reveal. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Personal Profile | Non-financial identity/dating information required for discovery, Candidate Cards, and matched state. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Photo | Candidate identity imagery, obscured before Mutual Match and visible afterward. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Portfolio Showcase | Owner-only surface for live SnapTrade account, balance/value, holding/position, activity/order, connection, and freshness data. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Portfolio Visualization | Chart/composition/holding/trend/activity/comparison or other source/derived portfolio representation; Candidate Card detail obeys Disclosure Level. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Portfolio-Derived Signal | Trait, band, comparison, or activity indicator derived for private matching or permitted display; not advice or a worth score. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Profile Preview | Owner-only whole-profile representation across Disclosure Levels and pre-/post-match identity states. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Proximity | Approximate distance for eligibility/ranking; exact location is never shown. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Public Site | Unauthenticated shell for brand, proposition, explanation, legal/trust/contact paths, and demo entry. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Swipe Deck | Ordered Candidate Cards assembled from eligible Invisible Cohorts. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Synthetic Candidate | Fictional candidate containing no real person's financial data. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Synthetic Population | Reproducibly generated candidate collection exercising eligibility, ranking, disclosure, visualization, swipe, match, and edge states. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Usable Portfolio | Connected Portfolio with sufficient accessible data and acceptable Freshness State. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |
| Verified | Limited provenance/freshness claim; never completeness, wealth, stability, responsibility, or identity certification. | `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary` |

## Required Platform and Form Factors

- The milestone is a publicly hosted, responsive web experience. Native iOS and Android apps are not required. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#92-platform`)
- Every core flow and Public Site surface must work on contemporary phone and desktop viewports with parity for required information, disclosure rules, and actions even when layout differs. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#83-accessibility-and-responsive-use`)
- Profile Preview must offer representative phone and desktop presentations; the source leaves open whether desktop preview uses a device frame or the true responsive layout. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-20-preview-the-owners-candidate-card`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#21-personal-profile-and-preview`)

## Named Capabilities and Product Rules

### Consent-led connection and lifecycle

- Before connection, explain categories accessed, private use, possible visible disclosure, limitations, and disconnect; distinguish private matching from user-visible disclosure; require an explicit consent action before derivation. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-1-explain-consent-before-connection`)
- The protected owner authorizes through SnapTrade test OAuth and returns to authenticated success or safe failure. Denial, invalid callback, expired flow, and provider error must be recoverable. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-2-authorize-through-snaptrade-test-oauth`)
- Show source coverage, freshness/refresh context, and qualified—not “real time”—language. Unusable freshness prevents new portfolio-derived discovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-3-represent-connection-and-freshness`)
- Disconnect immediately removes connected status, stops new portfolio-derived use, deletes local source payloads/signals/rendered or cached views, explains local/provider boundaries, and requires fresh authorization to reconnect. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-4-disconnect-and-withdraw-consent`)

### Portfolio Showcase and source transformation

- The private Portfolio Showcase must represent accounts, balances/account values, holdings/positions, activities/orders, connection, and freshness when available. Each section distinguishes loaded, unavailable, empty, stale, failed, and unsupported. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-19-present-available-live-portfolio-data`)
- It must make the transformation from SnapTrade source data to Portfolio-Derived Signals and Candidate Card visualizations understandable, never fabricating missing values or offering trades/advice. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-21-demonstrate-source-to-experience-transformation`)

### Personal Profile and preview

- Initial required categories are an assumption: display name, adult-age confirmation, coarse location, discovery eligibility/preferences, relationship intent, at least one Photo, and a short biography or prompt response. Exact labels, formats, optional fields, and validation belong to UX. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
- UX must distinguish required, optional, private, pre-match visible, and post-match visible data, and explain which states an edit affects. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
- Discovery remains unavailable until required profile fields are valid and a Usable Portfolio is connected; each missing requirement must be identified with a path back to the relevant profile or connection flow. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
- Profile Preview combines current Personal Profile and live portfolio presentation across all prototype Disclosure Levels and pre-/post-match states; it clearly identifies private owner data and synthetic states. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-24-preview-the-complete-profile`)

### Preferences, disclosure, and initiation

- The user sets maximum distance and selects similar, diversified, or complementary investing compatibility. Maximum distance is a hard boundary and nearer candidates rank more strongly within it. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-5-set-discovery-preferences`)
- **Snapshot** permits allocation, asset classes, diversification context, value/activity bands, and Freshness State; no security names or exact monetary values. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)
- **Holdings** adds security names, position weights, percentage performance, and activity categories; no quantities, exact monetary values, or transaction detail. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)
- **Full Detail** adds exact quantities/values, portfolio/account value, performance amounts, and recent activity/order/transaction detail. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)
- Regular Discovery shows only candidates at an equal or lower Disclosure Level and only their permitted detail. A higher-disclosure user may initiate toward a lower-disclosure user; the lower-disclosure recipient may evaluate the initiator at the initiator's level without exposing more of their own information. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)
- Downgrading immediately invalidates no-longer-permitted active views, prefetched/rendered data, caches, and back-navigation restoration. Account identifiers, credentials, tokens, and exact location are never disclosable. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#32-data-and-derivation-boundary`)
- Explain that preferences, Disclosure Level, Proximity, and portfolio compatibility affect Discovery, but do not expose cohort names, membership, boundaries, ranking scores, or exact formulas. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-7-explain-preference-effects-without-exposing-internal-logic`)

### Discovery and Candidate Cards

- Assemble and order the Swipe Deck from portfolio signals, compatibility preference, Disclosure Level, Proximity, and candidate activity. Changing a material input must visibly change eligibility/order in seeded scenarios. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-10-assemble-an-eligible-swipe-deck`)
- Sparse eligibility must produce an honest low-inventory state, never a silent relaxation of privacy or eligibility rules. It may explain scarcity and invite a voluntary settings change. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-10-assemble-an-eligible-swipe-deck`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#43-sparse-pool-pressure`)
- Candidate Cards must give portfolio content greater visual prominence than the obscured Photo, use charts/visual forms rather than a badge or text-only summary, provide comparison context, coarse Proximity, and Freshness State, and distinguish facts from derived comparisons when confusion is plausible. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-11-present-a-portfolio-first-candidate-card`)
- Potential visualization components include allocation/composition charts, expandable holding views, account value or bands with coverage boundaries, performance with period/freshness, activity/order/transaction timelines, and source-versus-derived comparisons. Use a coherent subset rather than every dataset at once. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#23-visualization-strategy`)
- Every synthetic Candidate Card, candidate detail, and Mutual Match state carries a persistent, legible synthetic-data indicator through expansion, responsive layout, and navigation. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-9-isolate-live-and-synthetic-data`)
- One left/right swipe on the combined card is persisted for the session and advances the deck; repeated input cannot duplicate the decision. Photo reveal never creates a second accept/reject gate. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-12-make-one-swipe-decision`)

### Mutual Match

- A positive owner swipe plus the Synthetic Candidate's preconfigured positive decision creates exactly one Mutual Match. Non-reciprocal right swipes and all left swipes do not reveal the Photo. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-14-create-a-mutual-match`)
- The matched state immediately reveals the Photo as payoff for the original swipe, acknowledges the match, and provides a path back to Discovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-15-reveal-identity-after-mutual-interest`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#14-progressive-identity-reveal`)
- Chat is not required for this milestone. If later included, it cannot be an isolated text box; it needs unmatch, block, report, solicitation rules, post-safety-action disclosure suppression, moderation routing, and an accountable response model. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#27-future-chat-boundary`)

## Required and Implied Surfaces

The following surfaces are explicit or directly implied by a required action/state. Final navigation and decomposition are left to UX. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#91-required-product-surfaces`)

### Public and entry surfaces

- Branded responsive landing/Public Site with product pitch, portfolio-first/progressive-reveal explanation, demonstration boundary, and clear demo entry. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-25-present-the-findur-proposition-and-brand`)
- Stable unauthenticated About, Terms, Privacy, trust and safety, and contact/support routes reachable through persistent navigation or footer paths. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-26-provide-expected-informational-and-legal-surfaces`)
- Protected-owner authentication/entry and session boundary (required by access rules; exact authentication UI is not specified). (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-2-authorize-through-snaptrade-test-oauth`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#31-authorization-and-session-boundary`)
- Optional Guest Demo entry and synthetic trial, conditional on architecture sizing. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-27-offer-a-signup-free-synthetic-trial-conditional-stretch`)

### Owner setup and control surfaces

- Pre-connection consent and data-purpose explanation. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-1-explain-consent-before-connection`)
- SnapTrade handoff/callback return, including success, denial, invalid callback, expired flow, and provider-error outcomes. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-2-authorize-through-snaptrade-test-oauth`)
- Connection/freshness management with syncing, stale, reauthorization-required, failed, disconnected, and reconnect states. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-3-represent-connection-and-freshness`)
- Disconnect/revocation confirmation explaining immediate local effects and any provider-side boundary. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-4-disconnect-and-withdraw-consent`)
- Personal Profile create/edit flow, including Photo input, adult confirmation, coarse location, intent, biography/prompt, validation, and incomplete-requirement recovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
- Compatibility preferences and Disclosure Level selection/preview. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#44-compatibility-preferences-and-disclosure`)
- Private Portfolio Showcase with dataset subsections and source-to-experience explanation. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#42-connected-portfolio-showcase-and-card-preview`)
- Profile Preview with Disclosure Level switching, pre-/post-match switching, live-owner labeling, and phone/desktop presentations. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-24-preview-the-complete-profile`)

### Discovery and outcome surfaces

- Discovery/Swipe Deck with rich Candidate Cards, one-stage swipe actions, and persistent synthetic provenance. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#46-portfolio-first-discovery`)
- Candidate Card depth/expanded states may be used for rich data, but the exact interaction is not selected. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#23-visualization-strategy`)
- Discovery unavailable/empty states for disconnected, syncing, stale, reauthorization-needed, failed, and no eligible candidates, each with recovery where possible. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-13-handle-discovery-edge-states`)
- Mutual Match/photo-reveal state with return to Discovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#47-mutual-match-and-progressive-identity-reveal`)
- An incoming-interest evaluation surface is implied by the required higher-to-lower Disclosure Level initiation asymmetry, but it is not named or included in the key milestone journeys. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)

## Key Journeys

1. **UJ-1 — Maya connects and explores her live portfolio:** enter hosted demo; understand private versus visible data; authorize through test OAuth; inspect accounts, balances/values, holdings/positions, activities/orders, and freshness; preview the live card at different Disclosure Levels; recover from connection/refresh failure. Climax: Maya understands both integration breadth and the boundary between source data, derived compatibility, and public disclosure. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
2. **UJ-2 — Maya sets compatibility and disclosure boundaries:** with a Usable Portfolio, set maximum distance, select similar/diversified/complementary behavior, and choose a Disclosure Level with effect explanations but no cohort/ranking exposure. Climax: Maya can predict the privacy effect before Discovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
3. **UJ-3 — Maya explores a deep, visual Swipe Deck:** evaluate a large generated Synthetic Population through rich level-limited portfolio visualizations, coarse Proximity, freshness, and obscured Photo; never see exact location, identifiers, internal ranking, or a worth score; swipe once on the whole card. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
4. **UJ-4 — Maya reaches Mutual Match and Photo reveal:** swipe right on a reciprocating Synthetic Candidate; see immediate Photo reveal and matched state without a second decision; return to Discovery. Chat is not required. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
5. **UJ-5 — Maya withdraws access:** disconnect; immediately lose current-connected presentation and portfolio-derived Discovery use; understand deletion/retention effects; reconnect only through fresh authorization. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
6. **UJ-6 — Alex understands Findur before signing in:** use branded Public Site to understand premise, connection/disclosure, obscured Photos, and demonstration status; inspect About, trust and safety, Privacy, Terms, and contact; then choose owner entry or conditional Guest Demo. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
7. **UJ-7 — Maya completes and previews her whole profile:** add minimum dating/personal context, Photo, short self-description, and preferences; preview the combined live profile at all Disclosure Levels and pre-/post-match states before anything could be shared. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)

## Trust, Privacy, and Safety Constraints

- Treat private matching use, owner-only exact display, and candidate-visible disclosure as separate permissions. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#12-private-matching-versus-visible-disclosure`)
- Minimize collected and client-delivered financial data to fields with a defined showcase, preview, matching, disclosure, or lifecycle purpose; exact live values remain owner-only in this milestone. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#81-privacy-and-security`)
- Never expose credentials, tokens, account identifiers, exact location, raw live payloads in logs, or unnecessary PII. Use least privilege and read-only access. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#81-privacy-and-security`)
- Live owner and synthetic data must remain distinguishable in storage, processing, and presentation; no generated financial record may be copied from a real person. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-9-isolate-live-and-synthetic-data`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-22-generate-a-varied-synthetic-population`)
- Missing/unsupported data is unknown, not zero or a negative trait; provider failure must lead to an honest recoverable state, never fabricated currency or false freshness. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-8-derive-a-private-matching-profile`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#82-reliability-and-observability`)
- Disclosure controls cannot prevent screenshots, memory, or inference; research must test combinations of fields/charts/context, not only individual fields. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#42-screenshot-and-inference-risk`)
- Do not improve inventory or engagement by weakening disclosure, surfacing hidden data, or coercing higher disclosure. Do not hide stale/failure states or let realism cause synthetic-data confusion. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#73-counter-metrics-and-guardrails`)

## States, Errors, and Feedback

### Portfolio and provider states

- Freshness State vocabulary: connected, syncing, stale, needs reauthorization/reauthorization-required, failed, disconnected; loaded, unavailable, empty, and unsupported apply at dataset-section level. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#3-glossary`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-19-present-available-live-portfolio-data`)
- Authorization outcomes include success, denial, invalid callback, expired flow, and provider error. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-2-authorize-through-snaptrade-test-oauth`)
- Long-running provider work must expose progress or a pending Freshness State instead of blocking without feedback. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#84-performance`)
- An unusable Freshness State blocks new derived output; stale/failed/disconnected states cannot be presented as current. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-3-represent-connection-and-freshness`)

### Profile, disclosure, and navigation states

- Profile inputs need valid/invalid, required/optional, private/pre-match/post-match, edited, and incomplete-gate states. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
- Disclosure preview must make hidden, bucketed, derived, and exact fields legible at each level. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#21-personal-profile-and-preview`)
- Downgrade/disconnect/safety changes must remove stale detail from active views, cache, prefetch, rendered representations, and navigation-history restoration. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#32-data-and-derivation-boundary`)

### Discovery and match states

- Candidate/deck variants include compatible preference modes, Disclosure Levels, Proximity boundaries, portfolio patterns, current/stale/unsupported source states, empty activity, sparse eligibility, and reciprocal/non-reciprocal swipes. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-22-generate-a-varied-synthetic-population`)
- Discovery-unavailable states require an appropriate next action where recovery is possible and must not silently substitute unqualified candidates. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-13-handle-discovery-edge-states`)
- Swipe state is session-persisted and idempotent; match creation is one-time/idempotent. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-12-make-one-swipe-decision`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-14-create-a-mutual-match`)

## Accessibility and Responsive Requirements

- Core flows must be keyboard-operable with meaningful labels, visible focus states, error text, and non-color indicators. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#83-accessibility-and-responsive-use`)
- Photo obscuring, Portfolio Visualizations, and swipe controls need accessible alternatives conveying equivalent state and actions. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#83-accessibility-and-responsive-use`)
- Required information, disclosure rules, and actions must remain equivalent on phone and desktop even if composition and interaction differ. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#21-personal-profile-and-preview`)
- Persistent synthetic indicators must survive card expansion, responsive layout, and navigation. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-9-isolate-live-and-synthetic-data`)
- Cached/synthetic Discovery should become interactive within three seconds under ordinary broadband, excluding OAuth/provider sync; this is an explicit provisional assumption. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#84-performance`)

## Brand, Voice, and Visual Direction Hints

- The experience should feel enjoyable enough to be recognizably about dating and professional enough to deserve trust with financial data. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#93-aesthetic-and-tone`)
- Product language should be direct, calm, non-judgmental, and explicit about uncertainty. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#93-aesthetic-and-tone`)
- The Public Site should feel like a credible product, not an OAuth test harness, and should avoid both generic fintech branding and luxury-status cues. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#24-public-site-brand-and-product-narrative`)
- Portfolio Visualizations—not a badge, text-only summary, or Photo—must dominate Candidate Cards. The same visualization system should support live owner previews and synthetic Discovery cards. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#22-candidate-card-information-hierarchy`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#23-visualization-strategy`)
- The Photo reveal is the emotional payoff after mutual interest; do not add a second appearance-based rejection decision. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#progressive-identity-reveal`)

## Explicit Inspirations and Anti-Patterns

### Inspirations / analogies explicitly present

- Mainstream swipe-based dating products inspire the invisible continuous card deck: users do not browse named pools or see the exact signals that selected a candidate. No particular dating brand is named. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#candidate-selection-model`)
- The recognizable dating interaction is a continuous Swipe Deck with a single combined-card decision and a mutual-interest reveal. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#46-portfolio-first-discovery`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#14-progressive-identity-reveal`)

### Anti-patterns

- No generic verification badge standing in for material portfolio influence; no photo-first card with financial detail as decoration. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#the-problem`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-11-present-a-portfolio-first-candidate-card`)
- No named status tiers, visible cohorts, wealth leaderboard, universal wealth threshold, or public financial-status hierarchy. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#11-portfolio-governed-exclusivity`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#5-non-goals`)
- Avoid “verified wealthy,” “high-value person,” “smart money,” or equivalents; luxury imagery; false-precision compatibility percentages; market-movement urgency/FOMO; and claims that a portfolio is complete or continuously current. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#26-tone-and-anti-patterns`)
- Avoid red/green performance theater that reads as scoring the person; avoid playful copy that trivializes loss, shames performance, suggests trading, or exposes sensitive data. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#26-tone-and-anti-patterns`)
- Do not reduce people to portfolio value, expose exact ranking logic, make deterministic compatibility promises, or imply advice/financial worth. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#93-aesthetic-and-tone`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#13-candidate-selection`)
- Do not add chat without the associated safety system; if the bundle is unsupported, end the experience at matched state. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#27-future-chat-boundary`)

## Required Comprehension Outcomes

Formative review must test that participants understand selected-account coverage; the narrow meaning of Verified; private matching versus visible disclosure; live-owner versus synthetic boundaries; disclosure changes access without forcing more disclosure; freshness qualification; Synthetic Candidates' fictional status; and that Photo reveal follows the original mutual swipe with no second decision. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#25-required-comprehension-tests`)

## Open Questions, Assumptions, and UX Handoffs

### Explicit UX-owned or UX-involved open questions

- Which Portfolio-Derived Signals communicate similar, diversified, and complementary compatibility without implying advice or worth? Owner: product + UX; resolve before Candidate Card approval. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- What lightweight explanation communicates why a Synthetic Candidate appears without exposing ranking or Invisible Cohort logic? Owner: UX; resolve before Discovery copy/interaction approval. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- Which additional Personal Profile fields, if any, are required beyond the assumed minimum? Owner: UX + product; resolve before profile setup approval. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- If chat enters scope, what minimum block/report/moderation/unmatch/anti-solicitation bundle ships with it? Owner: product + safety + UX; resolve before chat becomes an epic/story. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- How should Snapshot, Holdings, and Full Detail be communicated without status labeling; what is previewed before choice; how are downgrade/disconnect/block/report effects explained; and how are symmetry/asymmetry explained? (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#12-private-matching-versus-visible-disclosure`)
- UX must choose responsive behavior, breakpoints, card proportions, visualization fallbacks, and device-frame versus true-responsive desktop preview. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#21-personal-profile-and-preview`)
- UX must select a coherent visualization subset/card-depth model rather than display every dataset on one card. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#23-visualization-strategy`)
- Brief-stage open decisions included exact tier contents/names, Candidate Card visual language, explanation depth, playful-event/performance boundaries, identity/anti-catfishing/reporting/moderation, and chat fit. The PRD resolves some but not all, as detailed under conflicts below. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#open-decisions-for-the-prd-and-ux`)

### Explicit assumptions affecting UX

- `[ASSUMPTION]` Minimum Personal Profile categories are display name, adult confirmation, coarse location, discovery eligibility/preferences, relationship intent, at least one Photo, and short biography/prompt. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#12-assumptions-index`)
- `[ASSUMPTION]` At least five representative formative reviewers can be recruited. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#12-assumptions-index`)
- `[ASSUMPTION]` Cached/synthetic Discovery interactivity target is three seconds under ordinary broadband. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#12-assumptions-index`)
- `[ASSUMPTION]` Scripted review runs are the milestone's primary acceptance method. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#12-assumptions-index`)

### `[NOTE FOR UX]` markers

- No literal `[NOTE FOR UX]` marker appears in the four supplied sources. UX-directed handoffs and owned questions are captured above from the PRD and addenda.

### Other unresolved dependencies with UX impact

- Architecture must set precise source-age thresholds for current/stale/unusable Freshness States. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- Product must decide the evidence/recruitment plan for formative review and ownership of legal/trust content approval. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- Architecture must decide whether Guest Demo is feasible without material scope expansion before epics/stories. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
- Identity verification, adult-age assurance, anti-catfishing, retention periods, production reporting/moderation, monetization, and post-match communication scope remain future decisions. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#5-options-deferred-for-later-product-work`)

## Source Conflicts and Scope Reconciliation

1. **Disclosure-level names and bundles are simultaneously fixed and called deferred.** The final PRD fixes three levels—Snapshot, Holdings, Full Detail—with exact field contracts; the subordinate PRD addendum says exact names and bundles remain deferred. Apply the PRD contract unless product records a new decision. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#5-options-deferred-for-later-product-work`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#findur-prd-addendum`)
2. **Brief-stage tier openness was resolved by the PRD.** The Product Brief listed exact tier contents/names as open and allowed a focused/limited first release; the PRD now makes all three prototype levels and their field behavior testable requirements. UX may refine labels only through explicit product change, not silently. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#open-decisions-for-the-prd-and-ux`; `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#tiered-disclosure-and-initiation`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#12-private-matching-versus-visible-disclosure`)
3. **Open-entry product vision versus owner-only milestone.** The brief says anyone with a Usable Portfolio can join and describes open-entry, portfolio-governed exclusivity. The PRD's current boundary permits only one protected owner, uses synthetic Discovery, and excludes public signup/real candidates. This is a product-direction versus milestone-scope distinction that all public copy must make explicit. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#executive-summary`; `_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#disclosure-exclusivity-and-trust`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#0-document-purpose`)
4. **Chat moved from high-value stretch to safely deferred/non-required.** The brief positions working one-to-one chat as a high-value stretch. The PRD says chat is not required, lists it outside MVP, and prohibits public messaging under the demonstration safety model. The PRD addendum permits it only with the full safety bundle. Design the matched state to end without messaging unless scope is explicitly changed. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#high-value-stretch`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#62-out-of-scope-for-mvp`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-18-preserve-a-future-safety-boundary`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#27-future-chat-boundary`)
5. **The PRD expands the core experience beyond the brief's proof list.** The final requirements add a branded Public Site, minimum Personal Profile, complete Profile Preview, detailed private Portfolio Showcase, stable legal/trust/contact surfaces, and persistent synthetic labeling. Treat these as required despite their lighter or absent treatment in the brief. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md#core-implementation-proof`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#61-in-scope`)
6. **Rich exact-looking portfolio detail is allowed only inside strict boundaries.** The brief warns against exact private data by default and left exact holdings/balances/transactions open. The PRD permits them under Full Detail for synthetic Discovery and in owner-only private live surfaces, while forbidding real cross-user disclosure in this milestone. This is a refined boundary, not permission to expose owner data publicly. (`_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/addendum.md#tiered-disclosure-and-initiation`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`)

## Surface-Closure Gaps for UX to Resolve

These are source-backed needs without a fully specified landing surface, journey, or interaction contract. They should be probed or explicitly marked as assumptions rather than invented.

1. **Incoming-interest asymmetry has no milestone journey or named surface.** FR-6 requires a lower-disclosure recipient to evaluate interest from a higher-disclosure initiator at the initiator's level, but the seven UJs cover only owner-initiated swiping and immediate synthetic match outcomes. Clarify whether the demonstration must show an inbound-interest state and, if so, how it is entered. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`)
2. **Protected-owner authentication is required but not designed.** Public entry, owner-only OAuth, and protected owner routes imply an authentication/session surface, yet the sources do not define the sign-in mechanism, recovery, session expiry UX, or transition from Public Site to owner context. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-2-authorize-through-snaptrade-test-oauth`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#31-authorization-and-session-boundary`)
3. **Overall IA/order is open.** Required surfaces are enumerated, but the sequence and navigation among connection, Personal Profile, preferences, Showcase, Profile Preview, Discovery, settings, and disconnect are not fixed. UJ-7 says “after connecting,” while the requirements only define gating before Discovery. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#24-key-user-journeys`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#91-required-product-surfaces`)
4. **Profile field interactions are underspecified.** Photo acquisition/upload, adult confirmation form, coarse-location method, relationship-intent options, prompt model, optional fields, validation, and privacy labels are explicitly delegated to UX. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-23-complete-a-minimum-personal-profile`)
5. **Disclosure change semantics lack user-facing interaction detail.** The invalidation behavior is exact, but the sources do not decide whether downgrades require confirmation, how consequences are previewed, what happens mid-card, or how the user understands irreversible prior viewing/screenshot risk. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-6-choose-a-disclosure-level`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#42-screenshot-and-inference-risk`)
6. **Candidate relevance explanation is unresolved.** Discovery must offer a human explanation without exposing rankings/cohorts, but the exact copy, placement, and depth are a UX-owned open question. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-7-explain-preference-effects-without-exposing-internal-logic`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
7. **Candidate Card depth model is open.** The sources support rich data and possible expansion/lightweight exploration but do not choose between a single card, tabs, progressive disclosure, expansion, or a candidate-detail surface. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#23-visualization-strategy`)
8. **Recovery actions are required but not mapped.** Each disconnected/syncing/stale/reauthorization/failed/empty state needs an appropriate next action, but the exact destination and retry/refresh/reconnect behavior are not specified. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-13-handle-discovery-edge-states`)
9. **Sparse-deck recovery needs a consent-safe path.** The product may explain scarcity and invite a voluntary setting change, but UX must decide which settings can be offered, how impact is previewed, and how to avoid coercive disclosure pressure. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#43-sparse-pool-pressure`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#73-counter-metrics-and-guardrails`)
10. **Contact/support versus reporting is unclear for the demo.** Public contact must accept privacy/safety/product concerns, while production reporting/moderation is deferred. Clarify how concerns are raised and labeled without implying an operational reporting system that does not exist. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-26-provide-expected-informational-and-legal-surfaces`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#5-options-deferred-for-later-product-work`)
11. **Matched-state persistence/history is unspecified.** The matched state needs a return path to Discovery, but no matches list, revisit behavior, or session-reset behavior is required. Do not invent one without a decision. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-15-reveal-identity-after-mutual-interest`)
12. **Guest Demo creates a conditional branch in the IA.** Architecture must first decide whether it ships; UX should not close public entry/navigation around an assumed guest branch before that decision. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#fr-27-offer-a-signup-free-synthetic-trial-conditional-stretch`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#11-open-questions`)
13. **Accessibility alternatives are requirements without selected patterns.** Equivalent alternatives are required for charts, Photo obscuring, and swipe actions, but the sources do not define chart summaries/data tables, obscured-image semantics, keyboard bindings, announcements, or reduced-motion behavior. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#83-accessibility-and-responsive-use`)
14. **Visual identity remains deliberately open.** The desired trust/dating balance and anti-patterns are clear, but no color, typography, shape, motion, imagery, iconography, or component-system decision is present in these sources. (`_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md#93-aesthetic-and-tone`; `_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md#24-public-site-brand-and-product-narrative`)
