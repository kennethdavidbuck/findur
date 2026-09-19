# Technology Currency Review

Reviewed 2026-09-19 against:

- `ARCHITECTURE-SPINE.md`
- `DATA-MODEL.md`

Lens: current existence and documented fit of every committed named technology or hosting capability. Only first-party project, vendor, language, or package-maintainer sources were used.

## Verdict

**CHANGES REQUIRED — the general stack is current, but the spine is not yet an executable architecture because two adopted SnapTrade/SDK assumptions conflict with the current published Go SDK.** The Render deployment is feasible in broad terms, but the authentication-critical same-origin rewrite also needs a deployed proof before it is treated as a settled invariant.

## Findings

### 1. Blocking — SnapTrade Go SDK v1.1.0 cannot currently express the documented OAuth request shape

**Location:** `ARCHITECTURE-SPINE.md` AD-5, especially lines 108–111; Deferred “Blocking SDK compatibility gate”

SnapTrade's OAuth documentation says Bearer requests must omit `clientId`, `consumerKey`, `userId`, `userSecret`, `timestamp`, and `Signature`. The current published Go module is v1.1.0. It can attach `Authorization: Bearer`, but its generated account methods still require `userId` and `userSecret` as Go arguments and unconditionally append both query parameters. Passing empty strings still emits `userId=&userSecret=` rather than omitting the fields.

The spine does recognize this in Story 0, which is good risk containment, but AD-5 simultaneously marks use of that SDK as adopted and supplies no supported path if the gate fails. This is therefore a known feasibility blocker, not merely future upgrade coverage. Current official documentation recommends SDKs generally, but the OAuth page does not document a Go-SDK-specific bearer configuration that resolves the generated signatures.

**Required guard:** Before finalizing the architecture, obtain a supported SnapTrade configuration/release that produces a Bearer-only request, or turn the SDK choice into an explicit open blocker with a pre-agreed fallback decision point. Do not let portfolio implementation begin on the assumption that v1.1.0 works.

**Consequence if unchanged:** Story 0 can halt the entire implementation after the architecture has already declared its only provider adapter mandatory.

Sources: [SnapTrade OAuth request requirements](https://docs.snaptrade.com/docs/oauth-apps#7-call-the-snaptrade-api), [published Go SDK v1.1.0 account request source](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/api_account_information.go), [published Go SDK v1.1.0 bearer-header source](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/client.go), [SnapTrade SDK catalogue](https://docs.snaptrade.com/docs/requests).

### 2. High — the “preserve provider precision” invariant is incompatible with the chosen SDK

**Location:** `ARCHITECTURE-SPINE.md` Consistency Conventions / Numeric values; `DATA-MODEL.md` Typed snapshot rows

The decimal persistence stack is valid: PostgreSQL `numeric`, `shopspring/decimal`, and `pgx-shopspring-decimal` work together. The problem occurs before persistence. In SnapTrade Go SDK v1.1.0, position `units`, `price`, and `cost_basis` are decimal strings, but balance `cash`/`buying_power`, account total `amount`, and activity `price`/`units`/`amount` are generated as `float32`. Once JSON has been decoded into `float32`, the application cannot truthfully preserve the provider's original decimal precision.

**Required guard:** Reconcile the invariant with the actual SDK boundary. Either secure corrected SDK model types, or define an explicit accepted precision policy for float-backed fields (including conversion and display/derivation rounding) and stop claiming provider-precision preservation for them. Contract fixtures must include values that expose float32 loss.

**Consequence if unchanged:** Stored decimals can look exact while actually encoding binary float artifacts or already-truncated provider values, undermining matching fixtures and financial displays.

Sources: [SnapTrade Go SDK v1.1.0 balance model](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/model_balance.go), [account-total model](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/model_account_balance_total.go), [activity model](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/model_account_universal_activity.go), [positions response contract](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_getAllAccountPositions), [pgx decimal integration](https://github.com/jackc/pgx-shopspring-decimal).

### 3. High — Render external rewrite support is documented, but the cookie-bearing same-origin topology is not fully proven

**Location:** `ARCHITECTURE-SPINE.md` AD-2 and AD-3

Render's canonical rewrite documentation allows a full public URL as a static-site rewrite destination, and Render's Redwood deployment guide explicitly uses a static-site rewrite to an API service URL. That supports the broad topology. However, the documentation does not specify the authentication-critical behavior this design relies on: forwarding non-GET methods and bodies, preserving multiple `Set-Cookie` headers, which host the backend observes, and preserving the browser-visible origin through OAuth callback and session creation.

There is also contradictory current first-party guidance: Render's current hybrid SPA tutorial says static sites “never proxy traffic” and prescribes direct browser-to-API calls plus CORS. Because AD-2 is intended specifically to avoid cross-origin cookies, the ambiguity matters.

**Required guard:** Add a blocking deployed Render spike that proves `GET`, mutating JSON, OAuth callback query strings, `Set-Cookie`, subsequent host-only cookie replay, `Cache-Control`, and error responses through `/api/*`. State the fallback now: either serve the Vite build from the Go service, or use separate origins with credentialed CORS and a deliberately compatible cookie policy.

**Consequence if unchanged:** The app can reach deployment with authentication/session behavior that cannot work through the selected edge route.

Sources: [Render static-site redirects and rewrites](https://render.com/docs/redirects-rewrites), [Render Redwood API rewrite example](https://render.com/docs/deploy-redwood), [Render hybrid SPA guidance](https://render.com/tutorials/web-service-vs-static-site/the-hybrid-pattern), [Render Blueprint routes](https://render.com/docs/blueprint-spec#static-sites).

### 4. Medium — `openapi-fetch` is in maintenance/deprecation mode

**Location:** `ARCHITECTURE-SPINE.md` AD-12 and Deferred exact versions

`openapi-typescript` 7.x remains current, and `openapi-fetch` 0.17.0 exists and works with OpenAPI 3.1 types. However, its own maintainers announced that `openapi-fetch` and the other non-core clients are moving to maintenance mode with no future feature work; the maintainer file describes those packages as deprecated so effort can return to `openapi-typescript`.

For this small evaluation build, a pinned maintenance-mode dependency may still be a rational choice, but the spine currently presents it as an unqualified current client choice.

**Required guard:** Either explicitly accept the maintenance-mode dependency for this bounded build and pin it, or use generated `openapi-typescript` types with a small Findur-owned `fetch` adapter. Do not defer awareness of the lifecycle state to Story 0.

**Consequence if unchanged:** The frontend starts on a client the maintainers have already moved away from, without a recorded ownership boundary for future fixes.

Sources: [openapi-ts 2026 roadmap](https://github.com/openapi-ts/openapi-typescript/discussions/2559), [maintainer lifecycle notes](https://github.com/openapi-ts/openapi-typescript/blob/main/MAINTAINERS.md), [current releases](https://github.com/openapi-ts/openapi-typescript/releases).

### 5. Medium — exact Go pinning conflicts with Render's native Go runtime

**Location:** `ARCHITECTURE-SPINE.md` Deferred exact versions and Render structural seed

`oapi-codegen` v2.8.x is a current, suitable choice for OpenAPI 3.1, `std-http-server`, and `StrictServerInterface`. Its corresponding OpenAPI-3.1-capable `nethttp-middleware` line requires Go 1.25. Render supports Go natively, but its official runtime documentation says native Go always follows latest stable and cannot be pinned to an exact version; Docker is required for a guaranteed runtime.

**Required guard:** Choose one enforceable rule: use Render native Go with a repository-declared minimum/toolchain and accept Render's automatic updates, or use a pinned Docker base image. Also pin `nethttp-middleware` at an OpenAPI-3.1-capable release alongside `oapi-codegen`, not only the generator.

**Consequence if unchanged:** “Pin supported Go” cannot be implemented as written, and a future Render runtime update can change the deployed toolchain independently of the repository.

Sources: [Render language-version policy](https://render.com/docs/language-support), [Render Docker reproducibility guidance](https://render.com/docs/docker#docker-or-native-runtime), [`oapi-codegen` v2.8 OpenAPI 3.1 release](https://github.com/oapi-codegen/oapi-codegen/discussions/2478), [`nethttp-middleware` OpenAPI 3.1 release](https://github.com/oapi-codegen/nethttp-middleware/releases/tag/v1.2.0), [`oapi-codegen` strict-server documentation](https://github.com/oapi-codegen/oapi-codegen#strict-server).

### 6. Medium — activity dates are nullable upstream but required by the data model

**Location:** `DATA-MODEL.md` Typed snapshot rows / `activity_rows`

The data model requires an activity `date`, but SnapTrade documents `trade_date` as nullable, and the SDK represents it with `NullableTime`. `settlement_date` is not a safe semantic substitute for the recorded activity time.

**Required guard:** Define whether undated activities are discarded with reduced coverage or stored with a nullable recorded date and excluded from date-ordered displays. Test the chosen behavior with a provider fixture.

**Consequence if unchanged:** A valid SnapTrade response can fail normalization or silently acquire an invented date.

Sources: [SnapTrade activities contract](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_getAccountActivities), [Go SDK activity model](https://raw.githubusercontent.com/passiv/snaptrade-sdks/sdks/go/v1.1.0/sdks/go/model_account_universal_activity.go).

## Verified Current and Fit

The following committed choices were verified against current first-party sources and have no technology-currency objection:

- Render supports free static sites, free web services, a single 1 GB free PostgreSQL database expiring after 30 days, and a paid 512 MB web-service plan at approximately $7/month. Free PostgreSQL has no backups or managed pooling, consistent with this non-production evaluation boundary. Sources: [free-tier limits](https://render.com/docs/free), [pricing](https://render.com/pricing).
- Root `render.yaml` Blueprints support web/static services, databases, static routes, health checks, `sync: false` secrets, and `autoDeployTrigger: checksPass`. `preDeployCommand` is paid-compute-only, matching the spine's “when supported” qualifier. Sources: [Blueprint specification](https://render.com/docs/blueprint-spec), [deploy lifecycle](https://render.com/docs/deploys).
- Hobby Render logging and metrics exist with seven-day retention; JSON `log/slog` output is directly suitable. Sources: [Render logs](https://render.com/docs/logging), [Render metrics](https://render.com/docs/service-metrics), [Go `slog`](https://go.dev/blog/slog).
- SnapTrade's Test OAuth app five-user limit; `openid read`; PKCE S256; exact redirect matching; OIDC discovery; RS256/JWKS verification; ten-hour access token; rotating refresh tokens; no refreshed ID token; Basic client authentication; revocation; Sandbox testing; and no `userinfo` endpoint all match the current OAuth guide. Source: [SnapTrade OAuth Apps](https://docs.snaptrade.com/docs/oauth-apps).
- The allowlisted positions, balances, accounts/connections, and paginated activities endpoints and selected fields exist. The ten-minute polling recommendation and rate-limit headers are current. Sources: [positions](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_getAllAccountPositions), [balances](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_getUserAccountBalance), [connection accounts](https://docs.snaptrade.com/reference/Connections/Connections_listBrokerageAuthorizationAccounts), [activities](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_getAccountActivities), [real-time versus Daily](https://docs.snaptrade.com/docs/realtime-data), [rate limits](https://docs.snaptrade.com/docs/ratelimiting).
- `golang.org/x/oauth2` supports PKCE helpers, and `coreos/go-oidc/v3` performs discovery and caches/refetches remote JWKS keys; nonce comparison remains correctly assigned to the application. Sources: [`x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2), [`go-oidc/v3`](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc).
- `pgx/v5` supports `pgxpool`, named arguments, and `StrictNamedArgs` (v5.6.0+); `pgx-shopspring-decimal` supports registering `shopspring/decimal` on each pooled connection. Sources: [`pgx/v5`](https://pkg.go.dev/github.com/jackc/pgx/v5), [decimal integration](https://github.com/jackc/pgx-shopspring-decimal).
- `golang-migrate/migrate/v4`, `golang.org/x/time/rate`, and `sony/gobreaker/v2` are current and fit their stated roles. Note that golang-migrate's PostgreSQL drivers already coordinate migrations with a lock, so any additional outer advisory lock should be justified rather than assumed necessary. Sources: [`golang-migrate`](https://github.com/golang-migrate/migrate), [`x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate), [`gobreaker/v2`](https://github.com/sony/gobreaker).
- WireMock has a maintained official Docker image and supports JSON stub mappings plus request matching for headers, methods, paths, query parameters, and bodies, fitting the proposed curated contract fixtures. Source: [WireMock Docker](https://wiremock.org/docs/standalone/docker/), [request matching](https://wiremock.org/docs/request-matching/).
- React Aria Components remains a current accessibility-first React component foundation. Source: [React Aria getting started](https://react-spectrum.adobe.com/react-aria/getting-started.html).

## Gate Recommendation

Do not finalize the spine until findings 1 and 2 are resolved. Treat finding 3 as a deployment gate that must pass before OAuth/session stories rely on the Render topology. Findings 4–6 can be corrected without changing the overall architecture.

## Resolution Check — 2026-09-19

Re-checked the updated `ARCHITECTURE-SPINE.md` and `DATA-MODEL.md` against the six findings above.

1. **PASS — SnapTrade OAuth request shape.** AD-5 now identifies the published v1.1.0 incompatibility, makes bearer-only WireMock capture a blocking Story 0 gate, and defines a narrow, reproducible upstream-spec overlay/generated-client fallback rather than assuming the SDK works.
2. **PASS — SDK `float32` precision.** The numeric convention and data model now explicitly accept and mark the upstream `float32` precision boundary, define canonical conversion, require loss-revealing fixtures, and no longer claim recovery of original JSON precision.
3. **PASS — Render same-origin rewrite.** AD-2 now makes deployed verification of methods, bodies, query strings, cookies, cache headers, and errors a Story 0 gate and predeclares the same-origin Go-served-assets fallback.
4. **PASS — `openapi-fetch` lifecycle.** AD-12 rejects maintenance-mode `openapi-fetch` and assigns HTTP behavior to a small Findur-owned `fetch` adapter typed from pinned `openapi-typescript` output.
5. **PASS — Go/OpenAPI/Render compatibility.** AD-12 pins both OpenAPI Go components; the version baseline requires Go 1.25 or newer and explicitly records that Render native Go follows current stable and is not exactly pinnable.
6. **PASS — nullable SnapTrade activity dates.** `activity_rows` now permits a nullable date, preserves source time precision, forbids invented midnight values, and excludes undated rows from chronological/recency use with reduced coverage.

**Updated verdict: PASS.** No technology-currency blocker from the six original findings remains. The Story 0 SDK request-shape and Render routing checks are acceptance gates, not unresolved architecture decisions.
