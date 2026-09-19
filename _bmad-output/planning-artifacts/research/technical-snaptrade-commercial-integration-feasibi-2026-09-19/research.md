---
title: 'Technical research: SnapTrade Commercial integration feasibility for Findur'
type: 'technical'
topic: 'SnapTrade Commercial integration feasibility for Findur'
decision: 'Can Findur use SnapTrade OAuth to demonstrate a portfolio-informed matching experience as a non-production SnapTrade job-application project?'
source: 'native web research; first-party SnapTrade sources'
status: complete
preset: 'standard'
validation: 'normal'
claims_verified: 0
claims_unverified: 4
created: '2026-09-19'
updated: '2026-09-19'
---

# Technical research: SnapTrade Commercial integration feasibility for Findur

**Decision this research serves:** Can Findur use SnapTrade OAuth to demonstrate a portfolio-informed matching experience as a non-production SnapTrade job-application project?

## Executive summary

**Yes for an evaluation-only prototype.** Findur's intended “Commercial OAuth login” can work if it means a **SnapTrade OAuth app registered under your Commercial account**: a user signs in with a SnapTrade Personal account, grants `read` access, and can also grant OpenID Connect sign-in and webhook scopes. This is distinct from SnapTrade's Commercial API-key model, where Findur creates a separate SnapTrade user and owns the brokerage-connection lifecycle. [1][2]

For Findur's proposed experience, the OAuth model is the closer fit: users can authenticate with SnapTrade, authorize existing brokerage data, and avoid Findur holding a Commercial `userSecret` or managing their brokerage connection. It provides the portfolio, balance, activity, and order data from which Findur could derive private compatibility features. [1][3]

Because this is a private application project that will not launch, the cross-user data question is a **demonstration boundary**, not a launch gate. Use your own connected account only for live data, and use synthetic candidate profiles or mock-derived signals for the matching pool. Develop the system with production-grade security, reliability, testing, and observability nevertheless. If the concept ever moves beyond that scope, SnapTrade's published terms and compliance policy make cross-user use of portfolio-derived attributes unresolved until SnapTrade gives written confirmation for the exact data flow. [5][6]

## Authentication and onboarding

SnapTrade's current OAuth documentation says an OAuth app needs a Commercial SnapTrade account, while the people authorizing the app have SnapTrade Personal accounts. OAuth supports `openid`, `email`, `read`, and `webhook` scopes; `openid` enables SnapTrade sign-in, `read` supplies account data, and `webhook` enables ongoing notifications. OAuth uses an authorization-code flow with PKCE, a confidential server-side client, and production HTTPS redirect URIs. [1]

This corrects the earlier framing: “Commercial OAuth” is a reasonable shorthand for this product setup, but technically it is **a Commercial-account-owned OAuth app serving Personal SnapTrade users**, rather than the Commercial API-key workflow. In the latter, Findur would create a stable `userId`, store a `userSecret`, generate a Connection Portal link, and retrieve account data using those credentials. [2]

**Recommended product flow:** Findur account/session → “Continue with SnapTrade” → SnapTrade OIDC/OAuth consent with the minimum `openid email read webhook` scopes → encrypted server-side token storage → create Findur's own session → derive private, freshness-labelled matching features. The user must already have (or create) a Personal SnapTrade account and connect a brokerage in SnapTrade. [1]

## Data and event fit

The documented account-data surface supports current accounts, positions, balances, historical activities, and orders. That is enough to compute features such as allocation breadth, concentration bands, asset-class exposure, and opt-in behaviour patterns. This is a product inference from the data surface—not a guarantee that a particular matching model will be useful. [3]

Freshness prevents a “live trades” promise. Activities update daily and may lack time-of-day; orders provide a shorter lookback with richer timing, and recent orders are the more current endpoint. Holdings freshness depends on plan and brokerage. Webhooks report connection and sync lifecycle, but a holdings-updated event indicates a sync attempt/completion rather than a portfolio-change diff. Store source and freshness metadata alongside every derived feature, and design the first release as eventually consistent. [3][4]

## Production-grade build constraints and future-product boundary

For a private demonstration, SnapTrade's test OAuth app limit of five users is sufficient; use it unless SnapTrade specifically authorizes production credentials. The application should still be engineered to production standards: a secure backend, exact redirect-URI handling, confidential treatment of the client secret and tokens, defensive callback validation, encryption at rest, revocation/deletion handling, and observable failure paths. [1]

SnapTrade's published compliance policy requires clear explanation of data use and sharing. It says market data may be used only for the licensed end user and that trade suggestions, signals, or portfolio analysis must be initiated by that user. Its developer terms also prohibit selling, repackaging, or redistributing end-user data and market data without the stated permissions. [5][6]

For the prototype, do not reveal raw balances, positions, trades, security names, or financial scores to prospective matches. Keep the live SnapTrade-backed view private to the account holder, and make candidate cards fictional or synthetic. Public sources do not answer whether a consented, coarse compatibility result—such as “investment styles align”—may influence one user's candidate pool using another user's financial data; that question matters only if the project becomes a real product.

## Recommendation

1. **Use SnapTrade OAuth/OIDC, not the Commercial API-key user-registration flow**, for the stated “login with SnapTrade” experience. It directly fits users who already manage brokerage connections in SnapTrade. [1]
2. **Build production-grade OAuth boundaries.** Keep all client secrets and tokens server-side; validate PKCE, state, nonce, redirect URI, and ID tokens; encrypt persisted credentials; make revocation, deletion, and connection errors visible and testable. [1]
3. **Use live SnapTrade data only for your own private, authenticated demonstration account.** Derive a private portfolio view or compatibility explanation from it; use synthetic data for every candidate profile and match pool. [3][4]
4. **Keep matching read-focused and freshness-aware.** Do not claim real-time event processing or expose holdings, amounts, tickers, trades, or financial scores on candidate cards. [3][4]
5. **Treat written SnapTrade confirmation as a future-product requirement.** Obtain it before using one person's financial data—even in a derived form—to affect what another real user sees. [5][6]

## Open questions

- If the project becomes public, does SnapTrade approve use of consented, derived portfolio attributes to rank or introduce a user to another user?
- If the project becomes public, does a compatibility score count as a disclosure, market-data use, portfolio analysis, or a signal under the compliance policy?
- Which broker connected to your own SnapTrade Personal account best demonstrates the intended portfolio and activity data?
- What deletion/revocation behaviour should the prototype demonstrate if access is withdrawn?

## Source appendix

| Ref | Claim/finding | Publisher | Published | Accessed | Confidence |
| --- | --- | --- | --- | --- | --- |
| [1] | OAuth/OIDC model, scopes, confidential client, production requirements | [SnapTrade — Build an OAuth App](https://docs.snaptrade.com/docs/oauth-apps) | Not stated | 2026-09-19 | Medium |
| [2] | Commercial API-key model, user registration, connections, and secrets | [SnapTrade — Personal vs Commercial](https://docs.snaptrade.com/docs/personal-vs-commercial) | Not stated | 2026-09-19 | Medium |
| [3] | Accounts, positions, balances, activities, orders, and their limits | [SnapTrade — Account Data](https://docs.snaptrade.com/docs/account-data) | Not stated | 2026-09-19 | Medium |
| [4] | Sync/freshness and webhook semantics | [SnapTrade — Syncing and Data Freshness](https://docs.snaptrade.com/docs/syncing), [Webhooks](https://docs.snaptrade.com/docs/webhooks) | Not stated | 2026-09-19 | Medium |
| [5] | Compliance policy on data use, sharing, signals, and analysis | [SnapTrade — Application Compliance Policy](https://snaptrade.com/compliance-policy) | 2024-09-19 | 2026-09-19 | Medium |
| [6] | Contractual data-use restrictions and security obligations | [SnapTrade — Developer Terms of Use](https://snaptrade.com/developer-terms-of-use) | 2026-05-22 | 2026-09-19 | Medium |

## Staleness map

Re-check the OAuth model, scopes, data features, plans, and broker coverage by **2026-10-01**. Re-check the developer terms and compliance policy before any public prototype or production launch, and whenever SnapTrade changes either document.
