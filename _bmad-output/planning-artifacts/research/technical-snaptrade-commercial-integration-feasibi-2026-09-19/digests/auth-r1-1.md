# SnapTrade Commercial integration feasibility — authentication digest

## Decision

**Yes.** Findur can register a SnapTrade OAuth app under its **Commercial account**. A user then signs in with a SnapTrade Personal account, authorizes Findur to access existing brokerage connections, and can use OpenID Connect for sign-in. This is distinct from the Commercial API-key model, where Findur creates a SnapTrade user and owns that user's brokerage-connection lifecycle. [C1][C6]

## Claims

### C1 — Commercial is the documented fit for an app that owns its users' connection lifecycle

- **Claim:** SnapTrade directs a product that “creates and manages brokerage connections for its own users” to Commercial; Commercial is for a company building an app for its own end users, and its connections open for an app-managed SnapTrade user.
- **Exact URL:** https://docs.snaptrade.com/docs/build-with-ai and https://docs.snaptrade.com/docs/personal-vs-commercial
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High
- **Class:** Primary product documentation

### C2 — Commercial-account OAuth and Commercial API-key integration are distinct models

- **Claim:** A SnapTrade OAuth app is registered under a Commercial account, while its end users sign in with existing SnapTrade Personal accounts and grant bearer-token access. The separate Commercial API-key model uses signed requests, `userId`, and `userSecret` values. “Commercial OAuth” is therefore a reasonable shorthand for a Commercial-account-owned OAuth app, but it should not be implemented as the Commercial API-key user-registration flow. The Connection Portal may itself handle broker OAuth redirects, which is distinct from Findur's OAuth authorization with SnapTrade.
- **Exact URL:** https://docs.snaptrade.com/docs/personal-vs-commercial; https://docs.snaptrade.com/docs/authentication-methods; https://docs.snaptrade.com/docs/connections
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High (terminology conclusion is an inference directly from the model comparison)
- **Class:** Primary product documentation; documented-inference

### C3 — Commercial API-key flow, for comparison

- **Claim:** The documented Commercial flow is: (1) Findur creates a Commercial account/key and keeps its `clientId` and `consumerKey` on a secure backend; (2) for each Findur end user, Findur calls `POST https://api.snaptrade.com/snapTrade/registerUser` using a stable, immutable non-email `userId`; (3) SnapTrade returns a generated `userSecret`, which Findur must securely store and which is returned only at registration; (4) Findur calls `POST https://api.snaptrade.com/snapTrade/login` under Commercial authentication with that `userId` and `userSecret`; (5) it opens the returned, user-specific Connection Portal URL for the account owner; (6) after completion, Findur calls user-scoped data endpoints with `userId` and `userSecret` to obtain accounts, positions, balances, orders, and activities.
- **Exact URL:** https://docs.snaptrade.com/docs/getting-started; https://docs.snaptrade.com/docs/authentication-methods; https://docs.snaptrade.com/reference/Authentication/Authentication_registerSnapTradeUser; https://docs.snaptrade.com/reference/Authentication/Authentication_loginSnapTradeUser
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High
- **Class:** Primary API reference and product documentation

### C4 — Connection Portal is an app-managed brokerage-connection UI, not Findur credential collection

- **Claim:** The Connection Portal is SnapTrade's UI for linking brokerage accounts and handles brokerage selection, broker OAuth redirects, username/password authentication, and MFA. The generated login URL expires after five minutes. Findur should generate it server-side on user intent, return the URL to its client, and open it; custom redirect/status handling is supported after the connection attempt.
- **Exact URL:** https://docs.snaptrade.com/docs/connections; https://docs.snaptrade.com/reference/Authentication/Authentication_loginSnapTradeUser; https://docs.snaptrade.com/docs/implement-connection-portal
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High
- **Class:** Primary product documentation and API reference

### C5 — Commercial secrets must remain server-side; user secrets are authorization material

- **Claim:** SnapTrade says to keep the Commercial `consumerKey` and every user's `userSecret` on a secure backend and not expose them in browser or mobile clients. Commercial HTTP requests are signed with the `consumerKey`; user-scoped calls include `userId` and `userSecret`. The Commercial `consumerKey` represents the integration, while the per-user secret is an additional data-access credential.
- **Exact URL:** https://docs.snaptrade.com/docs/authentication-methods; https://docs.snaptrade.com/docs/personal-vs-commercial; https://docs.snaptrade.com/reference/Authentication/Authentication_registerSnapTradeUser
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High
- **Class:** Primary product documentation and API reference

### C6 — SnapTrade OAuth is the fit when Findur wants SnapTrade-managed connections

- **Claim:** OAuth lets an app registered under a Commercial SnapTrade account request access to brokerage accounts a user already manages in SnapTrade Personal. The app receives OAuth tokens instead of registering a second SnapTrade user, storing `userSecret`, signing requests, or embedding the Connection Portal just to access those existing accounts. OAuth requires a confidential server-side client; authorization-code flow uses PKCE, and the client secret/tokens must remain confidential. This model fits Findur when SnapTrade manages the brokerage-connection lifecycle.
- **Exact URL:** https://docs.snaptrade.com/docs/oauth-apps
- **Publisher:** SnapTrade
- **Publication date:** Not stated
- **Accessed:** 2026-09-19
- **Confidence:** High
- **Class:** Primary OAuth documentation

## Implementation constraints and gaps

- **Production prerequisite:** Official docs say a production Commercial key requires Dashboard approval/billing steps; the detailed OAuth guide says Production OAuth also requires KYC. The reviewed pages do not establish the exact current commercial contract, pricing, approval timeline, jurisdictional availability, or whether a portfolio-informed matching use case is permitted. Confirm these with SnapTrade before product commitment. [C3]
- **Broker coverage is conditional:** broker availability and required approvals/BYO keys vary. Some broker connections can later disable when tokens expire and require the end user to reconnect through the portal. [C4]
- **Data/privacy/compliance gap:** These sources establish the technical authorization path, not Findur's legal basis, consent language, data minimization/retention policy, investment-advice status, or financial-regulatory obligations. Those require legal/privacy review and, if applicable, SnapTrade contract review.
- **Recommended architecture inference:** authenticate the person to Findur through the SnapTrade OAuth/OIDC flow; perform token exchange and data access in Findur's backend; associate encrypted OAuth tokens with the Findur account; and offer revocation plus connection-status handling. This is an implementation inference supported by the OAuth security and lifecycle requirements, not a complete security design. [C6]

## Sources considered

| Source | Exact URL | Publisher | Date | Relevance |
| --- | --- | --- | --- | --- |
| Getting Started with SnapTrade | https://docs.snaptrade.com/docs/getting-started | SnapTrade | Not stated | Commercial quickstart, production key, registration, portal, data retrieval |
| Authentication Methods | https://docs.snaptrade.com/docs/authentication-methods | SnapTrade | Not stated | Commercial vs Personal API-key model; request credentials and server-only rule |
| SnapTrade Personal vs Commercial | https://docs.snaptrade.com/docs/personal-vs-commercial | SnapTrade | Not stated | Ownership, authorization, user-registration, and portal comparison |
| Register user endpoint | https://docs.snaptrade.com/reference/Authentication/Authentication_registerSnapTradeUser | SnapTrade | Not stated | Registration endpoint and immutable user-ID / generated-secret rules |
| Generate Connection Portal URL endpoint | https://docs.snaptrade.com/reference/Authentication/Authentication_loginSnapTradeUser | SnapTrade | Not stated | Login URL endpoint, expiry, redirect, and connection permissions |
| Connection Portal | https://docs.snaptrade.com/docs/implement-connection-portal | SnapTrade | Not stated | Client opening pattern, status/redirect behavior, browser guidance |
| Connections | https://docs.snaptrade.com/docs/connections | SnapTrade | Not stated | Portal's broker-credential/MFA/OAuth role and reconnection behavior |
| Build an OAuth App | https://docs.snaptrade.com/docs/oauth-apps | SnapTrade | Not stated | OAuth's distinct Personal-user model and confidential-client constraints |
| Broker Access Guide | https://docs.snaptrade.com/docs/broker-access-guide | SnapTrade | Not stated | Broker-specific production availability/approval constraints |

All sources above were accessed on **2026-09-19**. No project materials, third-party articles, or training-memory claims were used as evidence.
