---
title: SnapTrade feasibility research reconciliation
source: ../../research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md
target: prd.md and addendum.md
created: 2026-09-19
---

# SnapTrade feasibility research reconciliation

This extract records material gaps from the SnapTrade feasibility research only. It does not modify the PRD, addendum, or memlog.

## Material gaps

1. **OAuth onboarding preconditions and scope boundary are incomplete.** The addendum correctly identifies a Commercial-account-owned OAuth app serving Personal SnapTrade users and a confidential authorization-code/PKCE flow. It does not capture the remaining feasibility conditions: the owner/test user needs a SnapTrade Personal account with a brokerage connected in SnapTrade before Findur can read portfolio data, and the OAuth consent should request only the documented scopes needed for the selected experience (`openid`, `email`, `read`, and `webhook` only if lifecycle notifications are actually used). Record the user-facing precondition in the PRD and the exact scope justification in the addendum/architecture handoff.

2. **The test-versus-production credential boundary needs an explicit approval gate.** The PRD confines the milestone to test OAuth and one owner-controlled user, but does not preserve the researched test-app limit of five users or say that production OAuth credentials require SnapTrade authorization. Add this as a feasibility/release boundary in the addendum or future-launch gates; it prevents the hosted demonstration from being treated as permission to switch credentials or expand participant access.

3. **Policy-controlled analysis needs a user-initiation boundary.** SnapTrade's published compliance policy says trade suggestions, signals, and portfolio analysis must be initiated by the licensed end user. The PRD prohibits advice but permits automatic derivation of portfolio signals for discovery. Before relying on any feature that might be characterized as analysis or a signal, define the user action that initiates it for the owner-only demonstration and make written SnapTrade confirmation a gate for any different cross-user flow. This is distinct from the existing broad no-advice language.

4. **Source revalidation triggers are underspecified.** The PRD references a 2026-10-01 refresh generally, but does not retain the research's event-based triggers: re-check OAuth model, scopes, data features, plan limits, and broker coverage by that date; re-check developer terms and the compliance policy before any public prototype or production launch and whenever either policy changes. Addendum-level operational guidance should assign these checks before implementation or credential/scope changes.

## No material gaps found in the following researched areas

- Data availability and qualified handling of accounts, positions, balances, activities, and orders.
- Event limitations: webhooks describe connection/sync lifecycle and are not treated as holdings-change diffs.
- Per-source freshness, eventual consistency, stale/failed/reauthorization states, and prevention of stale presentation after consent changes.
- Server-side credential handling, encryption, callback validation, revocation/deletion, sensitive-log exclusion, and owner-live versus synthetic-candidate isolation.
