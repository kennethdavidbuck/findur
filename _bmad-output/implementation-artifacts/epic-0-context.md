# Epic 0 Context: Connect Securely to SnapTrade

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Move an eligible SnapTrade Test OAuth user from the existing bilingual public experience into a secure, isolated Findur session, with clear staged consent and a truthful view of provider connection state and the minimum masked account inventory needed for later account inclusion. Begin by proving the deployed one-origin frontend/API topology and bearer-only provider adapter, then deliver hosted OAuth, protected sessions, logout, callback recovery, and bounded post-authorization inventory retrieval without rebuilding the existing public foundation.

## Stories

- Story 0.1: Deploy and Verify the Exact Build
- Story 0.2: Begin Hosted SnapTrade Authorization Safely
- Story 0.3: Complete OAuth and Establish an Isolated Session
- Story 0.4: Protect the Authenticated Experience and End Sessions
- Story 0.5: Prove Authenticated SnapTrade Access with Masked Inventory

## Requirements & Constraints

- The public experience must explain the portfolio-first proposition, progressive identity reveal, adult-only and test-user boundaries, read-only/non-advisory use, anti-solicitation guidance, and that this is a demonstration rather than a public dating launch. Preserve Home/About and existing entry, locale, theme, responsive, and accessibility behavior; do not advertise public signup, real-user cross-disclosure, or Guest Demo. The missing standalone legal/trust/support suite remains deferred unless a provider requires a specific minimum notice.
- Consent is staged. Before OAuth, explain the named provider, requested categories, private-use and possible-disclosure boundaries, limitations, disconnect consequences, and that authorization initially permits only connection status plus minimum masked account metadata. OAuth connection must not imply account inclusion; no balances, positions, activities, derivation, previews, or Discovery begin until a later explicit inclusion confirmation, with no account selected by default.
- Hosted SnapTrade authorization must use only `openid read`, keep brokerage credentials and provider tokens outside the browser, and return safe authenticated success or categorical recovery states. Only an allowlisted internal destination may survive the handoff.
- Each admitted SnapTrade subject is an isolated viewer; an active subject resumes its user, while authorization after complete deletion creates a new user. There is no Findur password or public candidate account.
- Protected owner routes derive identity from the server-side session and reject client-supplied actor IDs. Unsafe session-authenticated mutations require session-bound CSRF protection and same-origin checks. Support current-session and all-session logout, checked-in idle/absolute expiry, and no protected payloads in browser persistence, URLs, metadata, service-worker caches, or public previews.
- Post-authorization bootstrap performs one bounded inventory operation. Persist only stable opaque connection/account identifiers, a safe brokerage label, account category/type, masked display label, eligibility/availability, and connection/sync state. Do not persist or log raw provider responses, full account identifiers, balances, positions, activities, orders, tokens, secrets, or exact financial values. Missing, unsupported, rate-limited, unauthorized, disabled, or failed states must be explicit and recoverable, never fabricated or presented as current.
- CI and integration verification use PostgreSQL and WireMock with curated fixtures and no live credentials or real financial data. Health, readiness, status, and deployment smoke must make zero SnapTrade calls. Logs expose only non-sensitive categorical outcomes and actionable lifecycle states.

## Technical Decisions

- Use a Go 1.25+ modular monolith with inward-pointing ports and adapters, a React/Vite SPA, PostgreSQL through `pgx/v5`, authored OpenAPI 3.1 with reproducibly generated strict Go and TypeScript artifacts, versioned migrations, and pinned tool/runtime dependencies. Domain and application packages remain independent of HTTP, PostgreSQL, Render, and provider implementation types.
- Render hosts the Vite Static Site and Go Web Service behind one public origin. Route `/api/*` to Go before the SPA fallback. Story 0.1 must prove unsafe JSON, callback query strings, multiple `Set-Cookie` headers, host-only cookie replay, cache headers, and error semantics. If that topology fails, serve immutable SPA assets from Go; do not introduce permissive credentialed CORS.
- GitHub Actions is authoritative: checks precede deploy, main deploys the backend's tested full-SHA `linux/amd64` image by immutable digest and the frontend from the same full commit, superseded deploys are cancelled, and bounded public-origin smoke verifies exact frontend/API build identity. The regular Compose stack contains pinned PostgreSQL, WireMock, backend, and a production-bundle frontend server with watchable development mode; it uses no Docker-in-Docker, host Docker socket, or Testcontainers.
- `/api/healthz` is lightweight liveness and reports the full build SHA without dependencies. `/api/readyz` reports the SHA and checks admission plus migrated PostgreSQL, returns 503 safely when unavailable, and never calls SnapTrade. Graceful shutdown fails readiness first, cancels and drains bounded work, then closes PostgreSQL; durable leases and idempotency guards preserve correctness across hard restarts.
- OAuth uses discovery-derived endpoints, authorization code with PKCE S256, high-entropy state and nonce, an encrypted expiring server-side verifier, and a separate one-time Secure/HttpOnly/host-only/SameSite=Lax correlation cookie. Atomically claim the attempt, exchange outside transactions, validate RS256 issuer/audience/time/nonce claims, and persist identity, independently AES-256-GCM-encrypted access/refresh tokens, session, and terminal result atomically. Sessions are opaque cookies; PostgreSQL stores only keyed session and CSRF hashes.
- All SnapTrade access passes through the Findur-owned `PortfolioProvider`. Prove that each allowlisted request sends only the server-held bearer token and no Commercial authentication fields. If the official SDK cannot satisfy this, use only the approved pinned upstream OpenAPI overlay and reproducible narrow generated client. Provider network work always occurs outside database transactions.

## UX & Interaction Patterns

Preserve English/French and System/Light/Dark parity, WCAG 2.2 AA keyboard and screen-reader operation, visible focus, reduced-motion equivalence, 44px touch targets, responsive reflow, and French label expansion. Use a reading-width Consent Panel with important statements in body text. OAuth outcomes focus their result heading; failures name the categorical problem, disclose no provider details, and offer one safe retry or back action. Protected content must not flash while session gates resolve. Authenticated navigation exposes exactly Discovery, Portfolio, and Profile, using bottom navigation below 768px and an accessible desktop rail above it. Connection confirmation and later account inclusion remain visually and verbally distinct.

## Cross-Story Dependencies

Story 0.1 is a mandatory deployment and request-shape gate for all later stories. Story 0.2 establishes the single-use authorization attempt consumed by Story 0.3. Story 0.3 creates the isolated session protected and terminated by Story 0.4 and the active authorization used by Story 0.5. Story 0.5's masked inventory is the input to Epic 1 account inclusion; it must not pre-empt that later consent step.
