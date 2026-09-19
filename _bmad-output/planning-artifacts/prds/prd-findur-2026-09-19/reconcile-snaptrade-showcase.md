---
title: "Reconciliation: SnapTrade showcase versus candidate-data restrictions"
status: complete
created: 2026-09-19
change_signal: "The PRD restriction on showing holdings and related data may prevent a credible SnapTrade API demonstration."
sources_compared:
  - "briefs/brief-findur-2026-09-19/brief.md"
  - "briefs/brief-findur-2026-09-19/addendum.md"
  - "research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md"
  - "prds/prd-findur-2026-09-19/prd.md"
---

# Conflict and Recommendation Extract

## Superseding change decision

The new product direction supersedes a conservative reading that candidate cards should demonstrate only sparse traits/bands. The PRD should use **rich portfolio-led card visualizations** as a primary SnapTrade showcase, governed by disclosure controls and clear provenance. The recommended pattern is a live, owner-only card preview based on the authenticated owner's connected data plus visually equivalent **synthetic** candidate cards. This does not authorize real cross-user financial disclosure.

## Verdict

The source restriction is **not a global ban on the owner seeing their own live SnapTrade data**. It is a boundary on prospective matches / candidate-card visibility and on cross-user use of real financial data. The current PRD correctly scopes several prohibitions to Candidate Cards, but it has no explicit owner-only private SnapTrade-data showcase requirement. That omission can make the demonstration look as though it only consumes SnapTrade to create opaque derived signals, rather than visibly demonstrating the connected API data.

## Exact-source reconciliation

| Topic | Source position | Current PRD position | Reconciliation |
| --- | --- | --- | --- |
| Owner's live data | Core proof calls for “one real private test portfolio for the owner.” The research explicitly recommends: “Use live SnapTrade data only for your own private, authenticated demonstration account” and “Derive a private portfolio view or compatibility explanation from it.” [Brief: Scope and Roadmap—Core implementation proof; Technical research: Recommendation 3] | PRD names one owner-controlled live portfolio and private derived signals, but does not require a private portfolio view or any visible source-data evidence. [PRD: §1, §4.3 FR-8/9, §6.1] | Add an owner-authenticated **Private Portfolio / SnapTrade Connection Detail** surface. It may visibly demonstrate connected account, position, balance, activity, and/or order data appropriate to the owner's own account, with source coverage and freshness context. |
| Candidate / other-user visibility | Cards “may communicate composition, activity, value bands, shared traits, or meaningful contrasts according to the candidate's disclosure choice”; a candidate “never sees more than the owner chose to disclose.” Success is comparing candidates “without seeing exact private data by default.” [Brief: The Experience 4–5; Success Criteria] | Candidate Cards forbid raw balances, exact holdings, transactions, security names, account identifiers, exact location, and ranking scores. Prototype disclosure levels allow only derived traits/bands. [PRD: UJ-3; FR-6; FR-11; §5] | The source supports rich, portfolio-led visualizations, but restricts private financial detail by default. Preserve no-real-user-data / no-prospective-match raw-data limits; replace the PRD's sparse-card interpretation with disclosure-controlled visualizations built from synthetic candidate data. |
| Research compliance boundary | “Do not reveal raw balances, positions, trades, security names, or financial scores to prospective matches. Keep the live SnapTrade-backed view private to the account holder, and make candidate cards fictional or synthetic.” [Technical research: Production-grade build constraints and future-product boundary] | PRD preserves synthetic candidates and card restrictions, but its lack of an owner-private view leaves the second half of this direction unrepresented. [PRD: FR-9; FR-11; §5; §6.1] | Preserve the distinction explicitly: raw/detail data may be shown only in the authenticated owner's private view; prospective-match/Candidate Card content remains synthetic and disclosure-limited. |
| Demonstration data capabilities | SnapTrade supports current accounts, positions, balances, historical activities, and orders. The research expects their use to demonstrate portfolio, balance, activity, and order data, while deriving allocation breadth, concentration bands, asset-class exposure, and opt-in behaviour patterns. [Technical research: Executive summary; Data and event fit] | FR-8 limits matching inputs to composition, diversification, value bands, and activity recency; this supports matching but does not visibly prove the API's account/position/balance/activity/order surface. [PRD: FR-8] | Use the private owner view to visibly prove the integration. Keep the portfolio-led card itself focused on derived/synthetic signals, rather than turning the dating surface into a brokerage dashboard. |
| Disclosure-model future | The addendum says the detailed design must decide “whether exact holdings, balances, and transactions are ever appropriate to display”; it does not decide they are never appropriate. [Addendum: Tiered Disclosure and Initiation] | FR-6 converts this unresolved question into an MVP assumption: no level permits raw balances, exact holdings, or transactions. [PRD: FR-6] | This is a valid focused-MVP decision for **other-user disclosure**. Do not read it as a restriction on the owner’s private connected-account view. Keep it marked as an MVP assumption/open downstream decision. |

## Recommended PRD decision

Adopt this scope clarification for the demonstration:

1. **Make rich portfolio visualization a required card capability.** Candidate Cards should visibly demonstrate composition, activity, value-band, shared-trait, and/or meaningful-contrast visualizations—the exact kinds of card content the brief permits—rather than reduce the experience to text-like traits. Disclosure Level governs the data granularity and visual elements available. [Brief: The Experience 4; Addendum: Tiered Disclosure and Initiation]
2. **Add a live owner-card preview / showcase.** After OAuth connection, the authenticated owner can preview the same card language from their live SnapTrade-backed data, clearly marked private and freshness-labelled. This may be accompanied by a private portfolio summary/detail to prove selected connected-account, position, balance, activity, and/or order data. [Technical research: Data and event fit; Recommendation 3]
3. **Use synthetic candidate cards for the public demonstration.** Candidate-card visualizations may be rich, but their data must be synthetic and never represented as another real user's connected SnapTrade data. Real live data remains visible only to the authenticated owner. [Brief: Scope and Roadmap—Core implementation proof; Explicit boundaries; Technical research: Production-grade build constraints and future-product boundary]
4. **Keep the showcase read-only, private where live, and non-advisory.** It must not trade, recommend investments, expose credentials, claim real-time freshness, or characterize connected data as complete net worth/financial health. Follow the documented eventually-consistent freshness boundary. [Brief: Explicit boundaries; Technical research: Data and event fit; Recommendation]
5. **Use disclosure, not a blanket visualization ban.** For the MVP, forbid raw account identifiers, exact location, and real cross-user financial data. For synthetic candidate cards, determine which visualized details the focused disclosure tiers allow; do not present illustrative raw values as real verified data. The addendum expressly leaves exact holdings/balances/transactions as a detailed-design decision, but research prohibits exposing those from a real person to prospective matches. [Addendum: Tiered Disclosure and Initiation; Technical research: Production-grade build constraints and future-product boundary]
6. **Add visible provenance.** Label the private preview as the owner's SnapTrade-connected data; label candidate data as synthetic/illustrative where necessary, preventing a reviewer from inferring real cross-user financial sharing. [PRD: FR-9; Technical research: Executive summary]

## Suggested requirement-level changes for later PRD revision (not applied here)

- Add a new functional requirement: “The owner-controlled test user can view a private, SnapTrade-backed portfolio card preview after connection.” Testable consequences: rich visualization uses selected supported live source data; states source coverage/freshness; remains owner-only; handles empty/stale/sync/error/disconnect states.
- Revise `FR-6` and `FR-11` to require rich, disclosure-controlled portfolio visualizations on Candidate Cards. Keep restrictions specific to prospective-match disclosure and real live data, rather than treating visualization itself as prohibited.
- Expand MVP in-scope wording to include “an authenticated owner-only SnapTrade portfolio card preview/detail showcase” and “rich, synthetic, disclosure-controlled candidate-card visualizations.”
- Clarify the non-goal: no raw *real-user* balances, holdings, transactions, security names, account identifiers, or financial scores to prospective matches. Do not weaken `NFR-1` or `NFR-4`: source payloads remain access-controlled and absent from client code, source control, and unredacted logs.
- Add a success proof that a scripted reviewer can connect, observe a private freshness-labelled live-data card preview, and observe rich synthetic candidate-card visualizations whose contents respond to disclosure levels and portfolio signals.

## Residual open decision

The source does not prescribe the exact owner-private fields, visual format, or whether the experience is a summary versus detail view. Choose a small, purposeful set that proves the SnapTrade connection without turning Findur into a portfolio-management product. Any future real-user/cross-user data use still requires written SnapTrade confirmation. [Technical research: Recommendation; Open questions]

## Fidelity audit (current PRD and addendum)

**Verdict: pass with two wording corrections.** The current artifacts preserve the source's core framing after the two requested clarifications: consent and disconnect remain explicit; private matching remains distinct from visible disclosure; the sole live portfolio stays owner-only; Discovery records remain synthetic; cohorts remain invisible; the one-swipe/mutual-photo-reveal flow remains intact; the work remains a test-OAuth demonstration rather than public launch; and no trading, investment advice, or financial-worth score is introduced. The rich-card clarification is contained by provenance, synthetic-data isolation, disclosure controls, and freshness language. The provisional 500-profile population is a stated assumption, not a claim of source support, and does not alter these boundaries. [PRD: §1; UJ-1–5; FR-1–21; FR-22; §5–7; §10; Addendum: §§1–4]

**Exact lines needing correction:**

1. **`prd.md:483`** says the clarification “supersedes the brief's sparse-card interpretation.” The brief was not sparse-card: it already made the portfolio the card hero and allowed composition, activity, value bands, shared traits, and contrasts. Replace with: “The user's clarification supersedes the PRD's earlier derived-traits/bands-only MVP interpretation; it retains the brief's portfolio-led card premise while requiring richer, disclosure-controlled visualization.” [Brief: The Experience 4; Product Promise]
2. **`addendum.md:64`** says “Cards should not collapse SnapTrade-backed data into a decorative badge.” Because Discovery cards are synthetic, “SnapTrade-backed data” can imply prohibited live cross-user data. Replace with: “Cards should not collapse SnapTrade-shaped synthetic portfolio data—or the owner's private SnapTrade-backed preview—into a decorative badge when richer visualization is permitted.” This preserves the owner-only/live versus synthetic-candidate boundary. [Addendum: §1.2; Technical research: Production-grade build constraints and future-product boundary]
