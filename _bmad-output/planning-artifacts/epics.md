---
stepsCompleted:
  - step-01-validate-prerequisites
  - step-02-design-epics
  - step-03-create-stories
inputDocuments:
  - _bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md
  - _bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/addendum.md
  - _bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md
  - _bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/DATA-MODEL.md
  - _bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md
  - _bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md
---

# findur - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for findur, decomposing the requirements from the PRD, UX Design, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: A user can review the SnapTrade data categories requested, the minimum masked account metadata needed for account selection, private-use and visible-disclosure boundaries, product limitations, and disconnect consequences before authorization.

FR2: A SnapTrade Personal subject admitted by the Test OAuth app can authorize Findur through SnapTrade OAuth, return to a safe authenticated success or failure state, and retrieve the minimum masked connection/account inventory needed for later inclusion. The architecture broadens the PRD's single-owner wording to at most five mutually isolated authenticated viewers; there is no separate Findur password, public candidate signup, or real-user cross-disclosure. A subject resumes an existing active identity, but authorizing after a prior full deletion creates a new Findur user.

FR3: Findur can represent committed Included Account coverage, dataset-specific source and observation context, and Freshness State wherever portfolio-derived claims are material, without presenting stale, failed, disconnected, or reauthorization-required data as current.

FR4: A user can disconnect SnapTrade and leave Findur. The operation immediately blocks new work, revokes every Findur session, attempts provider revocation, and deletes the entire user plus all provider identity, personal profile, preferences, disclosure, financial source/normalized/derived data, interaction state, cached/prefetched/rendered/restoration state, and in-flight publication authority from application-controlled active storage. Local destruction completes even if provider revocation fails; a later authorization creates a new user.

FR5: A user can set a maximum distance and select similar, diversified, or complementary investing compatibility. In the first-cut architecture, similar means smallest asset-kind allocation difference, diversified means broader instrument-kind coverage then lower largest-position concentration, and complementary means lower asset-kind allocation overlap.

FR6: A user can inspect and explicitly save Snapshot, Holdings, or Full Detail; distinguish previewed from saved state; understand the permitted fields and inference risk; cancel or discard drafts safely; and have privacy-decreasing changes suppress broader output immediately and fail closed.

FR7: Findur can explain that preferences, saved Disclosure Level, approximate Proximity, and portfolio compatibility affect Discovery without exposing Invisible Cohorts, internal scores, formulas, boundaries, or deterministic/worth-based claims.

FR8: Findur can derive a private, versioned matching profile from Included Account data in a Usable Portfolio, with source, coverage, and Freshness metadata. First-cut derived matching inputs are limited to asset-kind allocation, asset-kind breadth, and largest-position concentration; missing or unsupported inputs remain unknown.

FR9: Findur can combine each authenticated viewer's isolated live matching profile with generated candidates while making live and synthetic provenance persistent and unambiguous in storage, processing, and presentation. Real viewers never appear in another viewer's Discovery.

FR10: Findur can assemble and deterministically order an eligible Swipe Deck using hard disclosure/distance/readiness rules, supported compatibility signals, approximate proximity, prior decisions, candidate state, and stable tie-breakers; sparse inventory is reported honestly without relaxing privacy boundaries.

FR11: A user can view a portfolio-first Candidate Card with disclosure-controlled visualizations, bounded relevance context, approximate Proximity, source coverage, Freshness State, obscured placeholder avatar, and persistent synthetic provenance, and can open a protected deep-linkable Candidate Detail without changing those boundaries.

FR12: A user can make one idempotent Pass or Interested decision on the combined Candidate Card using gesture, keyboard, or explicit controls; the decision advances the deck and Photo reveal never creates a second acceptance gate.

FR13: Findur can explain and provide a safe recovery action when Discovery is unavailable or materially changed because of profile, inclusion, disclosure, freshness, authorization, provider, connectivity, or eligible-inventory state, without substituting unqualified candidates.

FR14: Findur can create a Mutual Match exactly once when an authenticated viewer's Interested decision meets that generated candidate's deterministic reciprocal positive decision; the swipe, reciprocal decision, and match settle atomically.

FR15: Findur can reveal the matched generated candidate's bundled placeholder avatar immediately after Mutual Match, retain synthetic provenance, announce the matched state, and provide a path back to Discovery without chat or a second decision.

FR16: Findur can state and enforce adult-only positioning for the demonstration, including adult attestation for authenticated profiles and unambiguously adult generated candidates.

FR17: Findur can surface concise trust guidance before or during Discovery that prohibits investment solicitation, money requests, financial targeting, and claims that connected or Verified data establishes wealth, financial responsibility, identity, or personal worth.

FR18: Findur can preserve the boundary between demonstration-safe functionality and future public functionality that requires commercial permission, legal/privacy/security review, identity and age controls, blocking/reporting/moderation, anti-scam operations, and production support.

FR19: Each authenticated viewer can inspect their own live Included Account data in a private Portfolio Showcase, including supported account inventory, balances/values, positions/holdings, recent activities, connection state, coverage, currency, and dataset-specific freshness, with explicit empty/unavailable/unsupported/failed states.

FR20: Each authenticated viewer can privately preview how their live portfolio and disclosure projections would appear across every Disclosure Level and representative phone/desktop compositions using the same visualization and projection rules as Discovery, without placing live data in any deck.

FR21: The Portfolio Showcase can trace supported SnapTrade source data into owner-only exact views, private derived/matching inputs, and candidate-visible projections so a reviewer can distinguish source, derivation, disclosure, provenance, missing data, and read-only/non-advisory boundaries.

FR22: Findur can reproducibly generate a shared, varied Synthetic Population from an explicit algorithm version, seed, and size, using stable IDs and required scenario tags to exercise compatibility, disclosure, proximity, freshness, sparse, unsupported, and reciprocal/non-reciprocal cases without importing real records.

FR23: Each authenticated viewer can atomically create and edit a minimum Personal Profile containing display name, adult attestation, selected coarse location area, relationship intent, short biography/prompt response, an allowlisted bundled avatar, locale, and theme; readiness remains gated until all prerequisites are valid.

FR24: Each authenticated viewer can open a private complete Profile Preview that combines their Personal Profile with live portfolio presentation across Disclosure Levels, pre-/post-match states, and phone/desktop compositions without mutating saved settings or entering any deck.

FR25: An unauthenticated visitor can understand Findur's portfolio-first proposition, progressive identity reveal, intended audience, adult-only and demonstration boundaries, and how an eligible test user enters, through a branded responsive Public Site.

FR26: [PARTIALLY DELIVERED; REMAINDER DEFERRED FOR MVP] Preserve the existing public landing/About navigation and concise demonstration boundary. Do not build the missing standalone Terms, Privacy, Trust and Safety, and Contact/Support suite unless a minimal provider requirement makes a specific notice necessary.

FR27: [NON-GOAL FOR MVP] The first cut does not implement or advertise Guest Demo; the product and architecture preserve only a future strictly synthetic, ephemeral, owner-isolated extension boundary.

FR28: [SUPERSEDED FOR FIRST CUT] The PRD's browser-notification and installed-app-badging requirement is excluded by architecture AD-17. Mutual Match and seeded Incoming Interest occur and persist only within active authenticated Discovery; no Web Push, subscription, badge, notification inbox, or background-delivery subsystem is implemented.

### NonFunctional Requirements

NFR1: Secrets, OAuth credentials, authorization codes, ID/access/refresh tokens, encryption keys, live portfolio payloads, and `.env` files must never be committed, exposed client-side, or included in logs.

NFR2: Financial collection and client delivery must be purpose-limited to committed Included Account fields needed for the Portfolio Showcase, preview, matching, disclosure, or lifecycle; launch, sharing, task-switcher, icon, metadata, and other ambient surfaces contain no candidate or financial detail.

NFR3: Data access defaults to least privilege and read-only behavior; the first cut provides no trading capability.

NFR4: Logs and diagnostics must expose actionable categorical lifecycle failures without raw holdings, balances, transactions, codes, tokens, exact financial values, or unnecessary personal information.

NFR5: Generated candidate data and every authenticated viewer's live portfolio must remain distinguishable and isolated in storage, processing, authorization, caching, and presentation.

NFR6: The hosted demonstration must expose actionable diagnostic states for OAuth initiation/callback, token lifecycle, provider sync, normalization/derivation, Account Inclusion, deck generation, swipe/match settlement, and disconnect.

NFR7: OAuth callbacks, inclusion changes, refresh/publication, disclosure changes, generated ticks, swipes, matches, and disconnect must be idempotent or version/lease guarded wherever duplication, replay, or concurrency could corrupt state.

NFR8: Provider unavailability, incomplete source data, missing currency support, or unsupported fields must degrade to an honest recoverable state; Findur never fabricates values, converts currency implicitly, or claims false freshness/precision.

NFR9: Load-bearing journeys conform to WCAG 2.2 AA, including keyboard operation, semantic labels, visible focus, associated instructions/errors, 4.5:1 normal-text contrast, 3:1 large-text/control/meaningful-graphic contrast, and meaning that never relies on color, motion, hover, node size, or layout alone.

NFR10: Every core and public flow retains functional parity on contemporary phone and desktop sizes; content reflows at 200% and at 400% in one dimension except bounded wide-data regions with non-table summaries; touch targets are at least 44 CSS px and primary swipe actions at least 48 CSS px.

NFR11: Photo obscuring/reveal, portfolio visuals, swipe controls, hover detail, motion, and charts provide equivalent keyboard, screen-reader, touch, text/table, and reduced-motion experiences; passive surfaces are not placed in the tab order.

NFR12: Cached or generated Discovery should become interactive within three seconds on ordinary broadband, excluding provider authorization/synchronization time.

NFR13: Long-running provider work exposes progress or pending Freshness State without an unexplained blocking interface, while safe stale-while-revalidate preserves only currently authorized trustworthy views.

NFR14: English and French, locale-aware values, and System/Light/Dark themes have functional and accessibility parity; explicit choices persist, and layouts tolerate at least 35% label expansion without lost meaning or actions.

NFR15: Installed/standalone web presentation preserves safe areas, session/route gates, responsive parity, privacy-safe launch/task-switcher treatment, and generic metadata; installability does not provide offline access to protected portfolio or candidate data.

### Additional Requirements

- AR1 — Mandatory sequencing gate: Epic 0 Story 0.1 (the architecture's Story 0) must validate and extend the existing Render walking skeleton and prove the deployed one-origin topology before authentication or portfolio stories depend on it: containerized Go 1.25+ modular monolith, React/Vite Render Static Site, PostgreSQL, root `render.yaml`, Static Site `/api/*` rewrite to Go, SPA fallback, health/readiness, unsafe JSON forwarding, callback query preservation, multiple `Set-Cookie` forwarding, host-only cookie replay, cache headers, and error forwarding. CI controls deployment: after backend/frontend checks pass, main-branch CI publishes and deploys the backend's exact full-SHA image/digest, deploys the static frontend from the exact full Git SHA, cancels superseded runs, and performs bounded post-deploy polling and smoke verification. If the topology proof fails, serve built SPA assets from Go; do not adopt permissive credentialed CORS.
- AR2 — The same Story 0 must pin runtime/tool/library versions and prove the current official SnapTrade Go SDK/configuration emits bearer-only requests for every allowlisted OAuth operation in WireMock. If it cannot, use the pre-authorized pinned upstream OpenAPI plus narrow checked-in overlay and reproducible generated client; no handwritten client, broad fork, or assumed SDK compatibility is allowed.
- AR3 — Organize the backend as one Go modular monolith with ports and adapters. Product/domain packages own rules and use cases; HTTP, PostgreSQL, SnapTrade, time, and randomness are edge adapters; dependencies point inward and no CQRS, services, queues, or event sourcing are introduced for the first cut.
- AR4 — The React SPA calls only Findur's authored OpenAPI 3.1 `/api` contract and never SnapTrade directly. Generate strict Go server types/glue and TypeScript browser types with pinned tools; generated artifacts are not hand-edited and CI fails on generation drift.
- AR5 — Model up to five independent OAuth-origin viewers admitted by the SnapTrade Test OAuth app. A verified `(provider, sub)` creates or resumes one active isolated Findur user; there is no hard-coded owner, separate username/password, public signup, or real-viewer candidate visibility. Full deletion removes the provider-subject mapping, so later authorization creates a new Findur user.
- AR6 — Implement SnapTrade-hosted authorization code flow with PKCE S256, high-entropy state/nonce, a separate expiring one-time pre-login correlation cookie, allowlisted internal return routes, discovered OIDC metadata/JWKS/endpoints, exact redirect URI validation, strict ID-token validation, atomic attempt consumption, replay-safe terminal states, and safe recovery for ambiguous finalization.
- AR7 — Issue opaque host-only `Secure`, `HttpOnly`, `SameSite=Lax` Findur session cookies; store only keyed token hashes; support multiple sessions, logout/logout-all, bounded idle and absolute expiry; and protect every unsafe session request with a session-bound synchronizer token, same-origin Origin/Referer validation, and Fetch Metadata checks.
- AR8 — Encrypt access and refresh tokens independently with AES-256-GCM, unique nonces, owner/provider/kind/version AAD, and a versioned key ring. Serialize rotating refresh and disconnect through generation-guarded operation leases; ambiguous refresh fails closed to reauthorization and disconnect intent wins.
- AR9 — Keep SnapTrade behind one `PortfolioProvider` adapter and purpose-limit retrieval to connection/account inventory, per-account balances, all positions, and a bounded recent activities window. Exclude trading, tax lots, orders, provider returns/performance, manual refresh, webhooks, symbols/reference data, quotes, fundamentals, sectors, and external enrichment unless architecture changes.
- AR10 — Before Account Inclusion confirmation, retrieve and retain only minimal masked inventory. Provider responses remain transient; normalize only purpose-required fields, never persist/log raw bodies, preserve decimal precision provenance, attach currency, and never silently aggregate or convert currencies.
- AR11 — Publish connections/accounts and per-account balances/positions/activities as immutable typed dataset versions with scoped atomic heads and dataset-specific observation, retrieval, publication, schema, and Freshness metadata. Failed fetch/normalization leaves the prior trustworthy head unchanged.
- AR12 — Fence every out-of-transaction provider result with authorization/portfolio lifecycle generation, inclusion version, and required membership. Account removal or disconnect advances the generation before deletion; stale finalizers discard transient results and cannot recreate data, credentials, signals, or anchors.
- AR13 — Make Account Inclusion an idempotent prepare/call/finalize bulk change. Removals commit and purge immediately; additions remain pending while retrieval occurs outside transactions and become effective only after complete guarded publication. Failed additions preserve the prior narrower committed set and never restore removals.
- AR14 — Drive refresh from authenticated activity only, never page rendering, synthetic views, health probes, deployment smoke, or heartbeats; persist and serve normalized snapshots without refetching; collapse concurrent refresh demand onto one guarded operation; bound provider calls, rate limits, retries, and circuit breaking; honor `Retry-After` and provider reset headers; never trigger billable manual refresh automatically; and distinguish disabled brokerage connection repair from expired OAuth reauthorization.
- AR15 — Centralize and acceptance-test first-cut freshness policy: holdings current at no more than 15 minutes in real-time mode or 36 hours in Daily mode, stale-usable through 72 hours; activities current through two calendar days and stale-usable through seven; missing initial data, revoked grants, or disabled connections are unusable.
- AR16 — Use one shared schema and versioned snapshot/signal builder for OAuth and generated users, with immutable `origin`. Generated users have no OAuth identity, credentials, or session; every real viewer sees only generated candidates; reciprocal outcomes are deterministic and viewer-specific.
- AR17 — Generate the shared population idempotently from algorithm version, seed, and size with stable IDs and scenario tags; validate the disclosure/proximity/freshness/portfolio/match boundary matrix before activation; accept no OAuth rows or external exports as generator input.
- AR18 — Advance generated presence and coherent portfolio events in one bounded deterministic singleton PostgreSQL tick using `pg_try_advisory_xact_lock`; preserve cross-field invariants, maintain minimum online inventory, inject clock/seed, and coalesce missed intervals rather than replaying them.
- AR19 — Recalculate candidate selection rather than persisting a full deck. Persist only swipes, matches, and a short-lived current-candidate/navigation anchor that pins the projected signal/disclosure snapshot; preserve safe active reading through immaterial refresh and invalidate immediately on permission decrease or lost eligibility.
- AR20 — Use a fixed migrated catalogue of coarse city/region areas; collect no browser/device/IP/address/postal-code location. Calculate inclusive maximum-distance eligibility server-side with tested Haversine centroids; expose only localized area labels and configured approximate distance bands, never coordinates or exact distance.
- AR21 — Derive only asset-kind allocation, breadth, and largest-position concentration for ranking. Similar uses smallest allocation difference; diversified uses broader kind coverage then lower concentration; complementary uses lower allocation overlap; approximate proximity is secondary and stable candidate ID is the final tie-breaker. No internal score is exposed.
- AR22 — Compute position value only with units, price, and currency. Compare positions only within an account total's currency, report excluded-row/account coverage, equally weight usable account vectors, and keep cash outside the ranking vector. Missing/stale data remains unknown.
- AR23 — Server-side disclosure projection must use separate owner-private, pre-match candidate, post-match candidate, and preview response schemas. Higher-tier fields and unrevealed avatar keys are never serialized and hidden only with CSS; protected responses are `private, no-store`.
- AR24 — PostgreSQL is the shared state and coordination layer using `pgx/v5`, strict named authored-query parameters, versioned `golang-migrate` up/down migrations, native locks, restrictive foreign keys, constraints, and explicit privacy lifecycle transactions; Redis and sticky sessions are not introduced.
- AR25 — Application use cases own explicit UnitOfWork transaction scopes and pass transaction-bound repositories. No provider/network/filesystem/sleep work occurs inside a database transaction; use prepare/call/finalize with guarded publication and retry only complete idempotent local use cases for recognized serialization/deadlock failures.
- AR26 — Derive an immutable authenticated Actor from the server-side session; clients never provide the acting user ID. Scope every private read/write by actor and resource identity, and integration-test cross-user access attempts for all private resources and lifecycle mutations.
- AR27 — Enforce sole module write ownership and the common lock order: auth/user lifecycle, portfolio/inclusion, dataset/signal heads, disclosure/anchors, then swipes/matches. Cover disconnect, disclosure, inclusion, swipe settlement, and simulation publication with concurrency tests.
- AR28 — Persist directed viewer/generated swipes and resulting match atomically. Incoming Interest is a one-time in-session generated scenario under Discovery. Exclude Web Push, subscriptions, notification outboxes, app badges, persistent notification/activity feeds, chat, and match history.
- AR29 — Use React Aria Components as the accessibility-first foundation; Candidate Card and Profile Preview share disclosure and visualization components. Profiles select bundled allowlisted placeholder avatars; do not create uploads, remote image URLs, object storage, or media-processing endpoints.
- AR30 — Ship a generic Web App Manifest and cache only hashed public assets; ship no service worker or offline authenticated experience. Never store protected payloads in local/session storage, IndexedDB, or Cache Storage, and clear in-memory query/navigation state on logout, session failure, disconnect, inclusion removal, or disclosure downgrade.
- AR31 — Use frontend English/French catalogues and `Intl` formatting; persist authenticated locale and `system|light|dark` theme in the profile while pre-login choices may use non-sensitive browser storage and are copied after login.
- AR32 — CI is authoritative and must cover OpenAPI/generation drift, Go formatting/vetting/tests, frontend lint/type/test/build, migrations against Compose PostgreSQL, and provider/OAuth/rate-limit/failure integration against Compose WireMock without live credentials. On main only, after checks pass, CI publishes the tested backend image, triggers its protected Render deploy hook with the exact full-SHA tag or digest, triggers the static frontend hook for the exact `GITHUB_SHA`, waits until the public frontend and API independently expose that full SHA, then runs deployment smoke. Applicable Render automatic deploys are disabled to prevent duplicate or racing deployments.
- AR33 — Use Render-native health, metrics, deploy events, and safe `slog` JSON output. Embed Render's full Git commit SHA independently in the static frontend and API; expose it through non-sensitive frontend build metadata and API health/readiness responses. Provide an unlinked public `/__status` route that executes a same-origin fetch and compares frontend/API SHAs without exposing configuration or provider state. Read configuration once at startup, fail closed on missing secrets, centralize bounded policy values, and keep secrets in Render `sync: false` settings only.
- AR34 — The Go process owns graceful lifecycle: readiness follows validated config/migrations/dependencies, turns false before drain, stops background claims, propagates cancellation and operation deadlines, drains HTTP/workers before closing PostgreSQL, and remains correct after hard kill through durable leases and idempotency.
- AR35 — Use PostgreSQL `numeric` and `shopspring/decimal`; OpenAPI represents decimal values as strings and money as amount plus ISO-4217 currency. Preserve string precision; mark and test unavoidable SDK `float32` precision loss; never use universal integer cents or persist binary-float artifacts.
- AR36 — Retain noncurrent normalized versions only while pinned or within the bounded cleanup grace; never build indefinite financial history. Inclusion removal overrides grace for the affected account, disclosure downgrade removes higher-detail projections without deleting still-consented private source data, and disconnect overrides every grace and destroys the complete user aggregate.
- AR37 — Integration tests must prove atomic publication, replay/idempotency, operation-lease serialization, lifecycle fencing, rollback on errors/panics, absence of open transactions across provider calls, simulation singleton behavior, optimistic concurrency, actor isolation, privacy deletion, and safe graceful-restart recovery.
- AR38 — Disconnect means leaving Findur and supersedes architecture/data-model provisions that retain external identity, nonfinancial profile, or categorical user-linked records. Use deletion as the default: block new work, fence in-flight finalizers, revoke all sessions, attempt provider revocation outside transactions, and explicitly or safely cascade-delete every user-owned row and client/server representation. Retain no tombstone or relationship shell unless implementation demonstrates an absolute integrity need that cannot be met by absence checks; the MVP has no real-to-real history requiring retention.
- AR39 — Automated deployment smoke never calls SnapTrade. It receives the expected full Git SHA, verifies the public shell and same-origin module asset, requires frontend build metadata plus `/api/healthz` and `/api/readyz` to report that SHA, and validates the `/__status` frontend-to-backend result. Readiness checks local admission, configuration/migrations, and PostgreSQL only. Live SnapTrade proof occurs through normal user-owned hosted OAuth, followed automatically by one bounded purpose-limited inventory retrieval; CI uses WireMock and never receives personal or provider credentials.
- AR40 — Containerize the Go backend and deploy it to Render as a prebuilt `linux/amd64` image published by CI under the full Git SHA and immutable digest; the frontend remains a Render Static Site built from the same exact Git SHA with pinned Node, lockfile, and build command. A regular Docker Compose stack—not Docker-in-Docker, a Docker socket mount, or Testcontainers—provides PostgreSQL, WireMock, backend, and a frontend production-build server for local/integration testing, with a watchable development mode. CI tests the production backend image and container-served frontend build, triggers the backend image deploy by exact tag/digest and the static frontend deploy by exact commit, then requires deployed SHA and browser smoke agreement. Byte-identical frontend artifact deployment is not claimed; reproducible same-source build plus deployed verification is the accepted MVP boundary.

### UX Design Requirements

UX-DR1: Implement the complete semantic Findur Constellation token system from `DESIGN.md`: paired light/dark colors, Inter/system human typography, IBM Plex Mono/system evidence typography, 4px spacing scale, restrained radii, focus colors, component role tokens, and theme-equivalent meaning.

UX-DR2: Implement shared interaction primitives from the normative Interaction State Contract for navigation, primary/secondary/destructive actions, links, selectable rows, interactive portfolio nodes, and passive data surfaces; hover/pressed states cannot move, scale, reflow, or change border width, and focus-visible remains strongest.

UX-DR3: Build the authenticated responsive App Navigation with exactly Discovery, Portfolio, and Profile: bottom navigation under 768px and a 224–256px desktop rail, programmatic current-area state, at least 48px mobile height, and safe area-local restoration.

UX-DR4: Preserve and selectively refine the existing Public Header and Public Footer, including entry, existing Home/About navigation, locale/theme controls, responsive behavior, keyboard order, and French label fit. Missing standalone legal/trust/support destinations are deferred with FR26 unless a minimal provider notice is required.

UX-DR5: Build a one-card Swipe Deck with explicit textual position, at most one decorative next-card outline, stable empty/recovery substitution, safe next-card preparation, exact current-card restoration, and no cohort exposure or silent eligibility widening.

UX-DR6: Build a shared Candidate Card component whose dominant content is a Portfolio Node Visualization, with evidence/relevance summary, approximate proximity, coverage/freshness, Disclosure Level, persistent Synthetic Data Label, subordinate Photo Obscure, and explicit Swipe Actions.

UX-DR7: Build protected Candidate Detail as a full-screen phone route and two-column desktop route with disclosure-permitted progressive sections, persistent provenance/freshness/photo state and Pass/Interested actions, focus on entry, and exact safe Back restoration.

UX-DR8: Build Portfolio Node Visualization with a programmatic title/summary and synchronized accessible Chart Data Table or equivalent. Labels, values, units, periods, source/derived distinctions, patterns, and coverage/freshness must provide complete nonvisual equivalence without hover, color, size, position, or motion dependence.

UX-DR9: Build Photo Obscure as an explicit geometric hidden state that exposes no hidden image description or cached preview; Mutual Match swaps in the same-size bundled placeholder avatar without layout jump and with concise accessible naming or declared decorative equivalence.

UX-DR10: Keep Synthetic Data Label visually and programmatically adjacent to identity on every generated Candidate Card, Detail, Incoming Interest, Preview, and Mutual Match state; responsive reflow, expansion, and navigation cannot remove it.

UX-DR11: Build Freshness Indicator for connected, syncing, stale, needs reauthorization, failed, and disconnected using text, timestamp, and shape/icon—not color alone—with access to dataset-specific details and no unsupported “live” claim.

UX-DR12: Build Account Inclusion Control as a named checkbox group with none selected by default, usable and disabled-with-reason rows, tri-state Select All, distinct draft and committed coverage, one scoped confirmation summary, and stable announcements for apply, purge, recalculation, success, and failure.

UX-DR13: Build Disclosure Level Control as a named three-option single-select with no default, Snapshot marked neutrally recommended, concrete hidden/bucketed/derived/exact examples, persistent saved-versus-preview summary, explicit Save/Cancel, discard-or-stay route protection, and screenshot/memory/inference warnings strongest at Full Detail.

UX-DR14: Build Preference Control for maximum distance and similar/diversified/complementary mode with explicit selected state and effects explanation, but no score, cohort, formula, or deterministic claim.

UX-DR15: Build Profile Field primitives with persistent label, required/optional and private/pre-match/post-match annotations, at least 44px height, preserved values, associated help/error text, and error treatment using text, icon/structure, and semantic color.

UX-DR16: Build Primary Action, Secondary Action, and Swipe Actions primitives with stable geometry, at least 44px targets (48px for swipe actions), keyboard/touch support, idempotent pending state, readable non-opacity-only disabled state, and visible prerequisites.

UX-DR17: Build Revalidation Notice and Loading Skeleton as distinct patterns: safe background revalidation retains the trustworthy task and focus, whereas skeletons are limited to cold loads, match expected layout, never imitate real values, and have external loading labels.

UX-DR18: Build Recovery Panel for every named gate/failure/empty state with the condition, what remains trustworthy, and the exact safe recovery action; recovery never silently changes disclosure, distance, or account selection.

UX-DR19: Build Mutual Match Reveal as a one-time accessible state transition that replaces Photo Obscure, retains provenance, announces once, uses reduced-motion final-state equivalence, and offers Continue Discovery as its sole forward action with no chat affordance.

UX-DR20: Build in-session Incoming Interest as a Discovery child surface at the initiator's Disclosure Level, explicitly preserving the viewer's lower level, with shared Detail/evidence and Pass/Interested behavior; do not create a fourth navigation area, inbox, feed, or history.

UX-DR21: Build Profile Preview Frame as an owner-private truth surface with live-owner label and independent controls for inspected Disclosure Level, pre-/post-match identity state, and representative phone/desktop composition; preview controls do not mutate saved state.

UX-DR22: Build Source-to-Experience Trace with three explicit branches—owner-only exact view, private matching input, candidate-visible disclosure—carrying provenance throughout and showing Disclosure Level filtering only on the visible branch; blocked/missing/unsupported stages use explicit text and strong dashed treatment.

UX-DR23: Build Consent Panel before OAuth with named provider/categories, minimum masked metadata, purpose, private-use/visible-disclosure distinction, limitations/inference risk, and disconnect effect. Continuing authorizes provider access/metadata only, not inclusion, retrieval, derivation, preview, or Discovery.

UX-DR24: Build one-level Confirmation Dialog for disconnect and consequential exclusions/downgrades, with clear effects on source/derived/cache/preview/Discovery outputs, a neutral cancel, semantic destructive confirmation, deterministic focus, and focus return to the invoker.

UX-DR25: Build Language and Theme Control for English/French and System/Light/Dark that applies immediately, persists appropriately, preserves the permitted route/task/focus, updates document language and locale formatting, and supports 35% French label expansion.

UX-DR26: Build context-appropriate Safety Notice covering 18+, read-only/non-advisory use, anti-solicitation, money requests, and financial targeting with a Trust and Safety link; required guidance cannot be reduced to a caption or legal-only placement.

UX-DR27: Implement routing/gating without protected-content flash: validate internal intended destinations, focus route/result/gate headings, show OAuth outcome before Account Inclusion, restore permitted deep-link intent only after session/readiness checks, and use named recovery destinations for invalid or unauthorized routes.

UX-DR28: Implement the permission-change consequence matrix exactly: additions never broaden before guarded success; exclusions, downgrade, disconnect, and safety invalidation suppress affected active/prefetched/rendered/history/offline/revalidation output immediately and retry fail-closed without restoring broader state.

UX-DR29: Implement safe stale-while-revalidate: show the last permitted trustworthy view with timestamp, revalidate without focus/task reset, announce material changes before rearranging the task, and transition to focused recovery when freshness becomes unusable; permission decreases always override preservation.

UX-DR30: Implement responsive behavior at phone `<768px`, tablet `768–1023px`, and desktop `≥1024px`, including mobile full-screen Detail, desktop two-column Detail, single focal cards, safe-area-aware persistent actions, bounded table overflow with summaries, 200% reflow, and applicable 400% single-column reflow.

UX-DR31: Implement the Constellation visual hierarchy: portfolio evidence leads; Photo and identity remain subordinate pre-match; surfaces use crisp rules, tonal depth, restrained glow, clipped/minimally softened corners, and no luxury/status, generic-fintech, dashboard-density, or ubiquitous-rounded-card styling.

UX-DR32: Implement motion only to explain relationships, deck continuity, or reveal state; never convey unique meaning. Reduced motion reaches the same complete final state with identical text/actions, no ambient distraction, and no focus or comprehension loss.

UX-DR33: Meet the detailed accessibility floor across EN/FR, light/dark, phone/desktop, and installed display: native semantics, logical reading/tab order, unclipped focus, route and state announcements, gesture alternatives, forced-color support, stable error association, zoom/reflow, and deterministic focus after invalidation.

UX-DR34: Implement privacy-safe installable presentation using an opaque branded constellation mark, theme-aware chrome, generic titles/metadata/icons, safe-area shell, and no financial/photo/candidate state in launch screens, task switchers, URLs, favicons, social previews, or recent-content surfaces.

UX-DR35: Keep existing public copy and any minimal provider-required notice readable at bounded width and use direct, calm, non-judgmental language that avoids wealth labels, luxury cues, red/green performance theater, trading urgency, financial shaming, false precision, or claims of completeness/currentness; production-grade legal copy is outside the MVP.

UX-DR36: Present Snapshot, Holdings, and Full Detail with the exact field contract and distinct source-versus-derived labeling; keep currencies separate, label order/activity coverage and date precision, show unavailable performance rather than inventing it, and never reduce portfolio or compatibility meaning to a universal score.

### FR Coverage Map

FR1: Story 0.2 - Explain staged consent before starting SnapTrade authorization.

FR2: Stories 0.2–0.5 - Authenticate through SnapTrade, establish isolated access, and retrieve masked inventory.

FR3: Stories 1.1–1.3 - Present committed coverage, source context, and honest dataset freshness.

FR4: Story 1.5 - Leave Findur, revoke sessions, and destroy the complete user aggregate.

FR5: Story 2.3 - Save maximum distance and supported compatibility preference.

FR6: Story 2.4 - Inspect, explicitly save, and safely decrease a Disclosure Level.

FR7: Stories 2.3, 2.4, and 3.3 - Explain preference, disclosure, and relevance without exposing ranking logic.

FR8: Story 2.1 - Derive the approved private matching signals from included SnapTrade data.

FR9: Stories 3.1–3.3 - Isolate real viewers and distinguish generated candidates throughout processing and presentation.

FR10: Story 3.2 - Assemble and deterministically order an eligible Swipe Deck.

FR11: Story 3.3 - Present disclosure-controlled portfolio-first Candidate Cards and protected Candidate Detail.

FR12: Story 3.4 - Persist one idempotent Pass or Interested decision and advance Discovery.

FR13: Story 3.7 - Explain and recover safely from Discovery gates, stale state, and sparse inventory.

FR14: Stories 3.4 and 3.5 - Settle reciprocal generated interest and create one Mutual Match atomically.

FR15: Story 3.6 - Reveal the bundled placeholder avatar after Mutual Match and return to Discovery.

FR16: Stories 0.2, 2.2, and 3.1 - Preserve adult-only positioning, attestation, and generated-candidate constraints.

FR17: Stories 0.2 and 3.2 - Present non-advisory, anti-solicitation, money-request, and financial-worth guidance.

FR18: Stories 0.2, 0.4, 3.5, and 3.6 - Maintain the demonstration boundary and exclude public multi-user and messaging scope.

FR19: Stories 1.1–1.3 - Retrieve and show supported live SnapTrade datasets in the private Portfolio Showcase.

FR20: Stories 2.4 and 2.5 - Preview the live portfolio across Disclosure Levels and representative layouts.

FR21: Stories 1.4, 2.1, and 2.5 - Trace source data into normalized, privately derived, and visible outputs.

FR22: Story 3.1 - Generate a reproducible, varied synthetic population covering the required scenarios.

FR23: Stories 2.2 and 2.5 - Create/edit the minimum Personal Profile and enforce readiness.

FR24: Story 2.5 - Preview the complete live-owner profile across identity and disclosure states.

FR25: Story 0.2 - Preserve the bilingual public proposition and activate its protected owner-entry path.

FR26: Story 0.2 - Preserve landing/About and defer the missing standalone legal/trust/support suite.

FR27: Stories 0.2 and 3.2 - Keep Guest Demo unimplemented and unadvertised.

FR28: Stories 3.5 and 3.6 - Keep interest and match outcomes in-session without push, badges, feeds, or background delivery.

## Epic List

### Epic 0: Connect Securely to SnapTrade

An eligible SnapTrade test user can move from the existing bilingual public experience into a securely authenticated, isolated Findur session, understand the staged consent boundary, and see their provider connection and eligible account inventory. The epic begins by validating and extending the existing Render walking skeleton and proving the approved bearer-only SnapTrade adapter path; it reuses and selectively polishes the current public pages instead of rebuilding them.

**FRs covered:** FR1, FR2, FR16, FR17, FR18, FR25, FR26, FR27

**Implementation notes:** Story 0.1 is the architecture-mandated Story 0 deployment/request-shape gate. It adds a regular Docker Compose development/integration stack, an exact-SHA/digest backend image deployment, exact-commit Render Static Site deployment, frontend/API build identity, automated post-deploy smoke, and the unlinked `/__status` same-origin frontend-to-backend diagnostic; health, readiness, status, and smoke make no SnapTrade calls. Compose uses explicit PostgreSQL and WireMock services without Docker-in-Docker, a host Docker socket, or Testcontainers. The frontend production bundle is tested in a container using the same pinned recipe Render uses, while the accepted Static Site boundary is reproducible rebuild plus deployed browser smoke rather than byte identity. Subsequent Epic 0 stories deliver working hosted SnapTrade OAuth login, isolated server-side sessions and logout, callback recovery, one bounded post-authorization connection/account-inventory retrieval, and wiring of the existing disabled owner-entry affordance. Personal SnapTrade credentials remain only in SnapTrade's hosted UI; CI uses WireMock and contains no live credentials. Preserve the implemented EN/FR, System/Light/Dark, responsive, accessible public foundation. Production-grade public/legal copy and the missing standalone route suite remain outside the MVP.

### Epic 1: Explore and Control Live Portfolio Data

The authenticated user can explicitly choose which connected accounts Findur may use, inspect the supported live SnapTrade account, balance, position, and recent-activity data with honest dataset-specific freshness and coverage, receive bounded activity-driven refresh without provider-call storms, understand how source data feeds the product, recover from provider states, and disconnect SnapTrade to leave Findur with destruction of the complete user aggregate.

**FRs covered:** FR3, FR4, FR19, FR21

**Dependency:** Uses Epic 0's authenticated session, proven provider adapter, and account inventory; it requires no later epic to deliver a complete private SnapTrade Portfolio Showcase, rate/freshness lifecycle, and destruction-by-default consent withdrawal.

### Epic 2: Transform SnapTrade Data into a Dating Profile

The user can turn the included live portfolio into the approved private compatibility signals, complete a personal profile, choose distance and compatibility preferences, explicitly control disclosure, and privately preview the resulting live-owner presentation across disclosure and identity states before entering Discovery.

**FRs covered:** FR5, FR6, FR7, FR8, FR20, FR23, FR24

**Dependency:** Builds only on Epic 1's committed included-account snapshots and freshness model; completion produces a fully explicit, server-derived Discovery-ready state without requiring Discovery itself.

### Epic 3: Demonstrate Portfolio-First Discovery and Matching

The ready user can explore a reproducible population of synthetic candidates through disclosure-controlled portfolio-first cards and detail, understand why candidates appear, make one durable Pass or Interested decision, handle honest Discovery edge states, review in-session generated interest, and reach a one-time Mutual Match with placeholder-avatar reveal.

**FRs covered:** FR9, FR10, FR11, FR12, FR13, FR14, FR15, FR22, FR28

**Dependency:** Uses Epic 2's saved profile, disclosure, preferences, and live-derived viewer signals. It completes the evaluator-facing dating proof without Web Push, app badging, notification feeds, Guest Demo, uploads, chat, match history, or real cross-user financial disclosure.

## Epic 0: Connect Securely to SnapTrade

An eligible SnapTrade test user can move from the existing bilingual public experience into a securely authenticated, isolated Findur session, understand the staged consent boundary, and see their provider connection and eligible account inventory. The epic begins by validating and extending the existing Render walking skeleton and proving the approved bearer-only SnapTrade adapter path; it reuses and selectively polishes the current public pages instead of rebuilding them.

### Story 0.1: Deploy and Verify the Exact Build

As a project owner,
I want CI to deploy and verify the exact tested commit,
So that I know the public frontend and backend are healthy, connected, and running the intended build.

**Acceptance Criteria:**

**Given** backend and frontend checks run for a pull request or non-main branch
**When** those checks complete
**Then** no Render deployment is triggered
**And** the existing build, test, lint, typecheck, migration, and provider-fixture checks retain their current behavior.

**Given** all required backend and frontend checks pass for a push to `main`
**When** the deployment job starts
**Then** it publishes the tested backend image for `linux/amd64` under the full `GITHUB_SHA` and records its immutable digest
**And** it triggers the protected backend Render deploy hook with that exact image reference
**And** it triggers the protected static-frontend Render deploy hook for the exact full `GITHUB_SHA`
**And** automatic deployment is disabled where applicable
**And** deploy-hook URLs are read only from GitHub Actions secrets and never printed.

**Given** a newer `main` commit starts deployment while an older deployment job is still running
**When** GitHub Actions applies deployment concurrency
**Then** the older job is cancelled
**And** only the newest commit remains eligible to complete smoke verification.

**Given** CI builds the backend image and Render builds the static frontend from the selected commit
**When** each artifact is created
**Then** each independently embeds the same expected full Git SHA
**And** the frontend exposes it through non-sensitive build metadata
**And** the API includes it in both `/api/healthz` and `/api/readyz`
**And** the backend image contains the application, migration command/files, and health behavior required at runtime without frontend assets.

**Given** a developer or CI job starts the regular Docker Compose stack
**When** its services become healthy
**Then** it provides pinned PostgreSQL, WireMock, backend, and a server for the frontend production build
**And** backend integration tests address those services directly over the Compose network
**And** no Docker-in-Docker daemon, host Docker socket mount, or Testcontainers dependency is used
**And** watchable development mode supports backend rebuild/restart and frontend refresh without changing the production deployment topology.

**Given** frontend verification runs before deployment
**When** the production frontend bundle is built and served in Compose
**Then** browser/integration tests exercise that bundle through a production-like SPA fallback and `/api/*` proxy to the backend
**And** the build uses the same pinned Node version, lockfile, and command configured for Render
**And** the accepted boundary is a reproducible Render rebuild from the same exact commit rather than a claim that Render deploys the byte-identical tested bundle.

**Given** the Render Blueprint configures the Go web service
**When** Render evaluates service health
**Then** `healthCheckPath` remains `/api/healthz`
**And** the platform health probe remains a lightweight process-liveness check rather than the PostgreSQL-backed readiness check.

**Given** the API process is running
**When** `/api/healthz` is requested
**Then** it returns HTTP 200 with `status: "ok"` and the full build SHA
**And** it does not query PostgreSQL, SnapTrade, or another external service.

**Given** the API is accepting work and PostgreSQL is reachable after required migrations
**When** `/api/readyz` is requested
**Then** it returns HTTP 200 with `status: "ready"` and the full build SHA
**And** it returns HTTP 503 with a non-sensitive unavailable state when admission or PostgreSQL readiness fails
**And** it never calls SnapTrade.

**Given** the Go process receives a shutdown signal
**When** graceful drain begins
**Then** readiness turns false before new background claims stop
**And** request/operation cancellation propagates while HTTP work and workers drain within the configured bound before PostgreSQL closes
**And** durable leases and idempotent operations remain safe after a hard kill or restart.

**Given** a visitor opens the unlinked `/__status` frontend route
**When** the page loads
**Then** it displays the frontend build SHA
**And** it performs one same-origin request to `/api/readyz`
**And** it displays the API build SHA, readiness state, and whether the builds match
**And** an unavailable, malformed, or mismatched response produces an explicit failure state without exposing configuration, database details, provider state, or secrets.

**Given** the deployed-browser status check runs in automation
**When** `/__status` loads for the expected build
**Then** the browser test verifies that frontend JavaScript executed and successfully called the backend through the public `/api/*` rewrite
**And** the reported frontend and API SHAs both equal the expected full SHA.

**Given** CI has triggered the two Render deployments
**When** post-deploy verification begins
**Then** it polls the public origin only within a configured attempt and time limit
**And** an older healthy deployment is treated as pending rather than successful
**And** verification proceeds only after both services expose the expected full SHA.

**Given** both deployed services report the expected SHA
**When** the deployment smoke script runs
**Then** it verifies the public HTML shell, same-origin module asset, non-HTML JavaScript response, API health, API readiness, frontend build metadata, and exact frontend/backend build match
**And** it fails for missing assets, SPA-fallback masquerading as JavaScript, unhealthy state, timeout, missing SHA, stale SHA, or mismatched SHA.

**Given** any health, readiness, build-metadata, status-page, or deployment-smoke operation executes
**When** its external calls are inspected
**Then** it makes zero SnapTrade requests
**And** records no credentials, secrets, financial values, or personal data.

**Given** the implementation baseline is built in CI
**When** architecture and contract checks run
**Then** the pinned Go, Node, React, Vite, OpenAPI generation, PostgreSQL migration, WireMock, and provider-adapter versions are reproducible
**And** authored OpenAPI 3.1 changes regenerate strict Go and TypeScript artifacts without drift
**And** application/domain packages remain independent of HTTP, PostgreSQL, Render, and provider implementation types.

**Given** the approved SnapTrade operations are exercised against WireMock
**When** their outbound requests are captured
**Then** each uses only the server-held OAuth bearer token required by that operation
**And** no Commercial `clientId`, `consumerKey`, `userId`, `userSecret`, `timestamp`, or `Signature` field is emitted
**And** a failing official SDK path activates only the approved pinned-upstream OpenAPI overlay and reproducible narrow-client fallback.

**Given** authentication-critical proxy fixtures exercise the one-origin topology
**When** requests and responses pass through the configured frontend-to-API route
**Then** unsafe JSON bodies, callback query strings, multiple `Set-Cookie` headers, host-only cookie replay, cache headers, and error status/body semantics are preserved
**And** a failed topology gate requires the documented Go-served-SPA fallback rather than permissive credentialed CORS.

### Story 0.2: Begin Hosted SnapTrade Authorization Safely

As an eligible test user,
I want to understand the connection boundary and begin hosted SnapTrade authorization,
So that I can authorize Findur without giving it my brokerage credentials or triggering premature financial-data retrieval.

**Acceptance Criteria:**

**Given** an unauthenticated visitor opens the existing public site
**When** Home, About, the public header/footer, locale/theme controls, or owner entry render
**Then** the current responsive bilingual portfolio-first proposition and progressive identity-reveal explanation are preserved and selectively refined rather than rebuilt
**And** adult-only, eligible-test-user, and demonstration-not-public-launch boundaries are clear
**And** Guest Demo, public signup, and real-user cross-disclosure are not advertised
**And** the missing standalone Terms, Privacy, Trust/Safety, and Contact suite remains deferred unless SnapTrade requires one specific minimum notice.

**Given** SnapTrade application configuration is available
**When** a visitor follows the existing owner-entry path from the public experience
**Then** Findur presents a concise staged-consent screen before beginning authorization
**And** the existing English/French, responsive, theme, and accessibility behavior is preserved rather than replaced.

**Given** the staged-consent screen is shown
**When** the visitor reviews what will happen
**Then** it identifies SnapTrade as the connection provider
**And** states that this step initially retrieves only provider-connection status and the minimum masked account inventory needed for account selection
**And** states that no account is included by default
**And** states that balances, positions, activities, signal derivation, profile previews, and Discovery remain unavailable until the user explicitly includes at least one account.

**Given** the visitor reviews the same consent
**When** the product boundary is described
**Then** it states that Findur never receives the user's brokerage password, cannot trade, does not provide financial advice, and does not treat wealth as personal worth
**And** it states that disconnecting SnapTrade means leaving Findur and deleting the user's data from app-controlled active storage
**And** the essential meaning is available in both English and French.

**Given** the visitor only loads, rerenders, navigates back to, or changes presentation settings on the public or consent experience
**When** no explicit authorization action is submitted
**Then** no SnapTrade authorization attempt or provider request is created.

**Given** the visitor explicitly begins authorization
**When** the backend creates the authorization attempt
**Then** it generates high-entropy state, nonce, and PKCE verifier values
**And** uses the S256 PKCE challenge method
**And** stores only an encrypted, expiring server-side verifier and the minimum correlation metadata
**And** sets a one-time, Secure, HttpOnly, host-only, SameSite=Lax correlation cookie that is not an authenticated application session.

**Given** an intended post-login destination accompanies the request
**When** the authorization attempt is created
**Then** only an allowlisted internal destination is accepted
**And** external, protocol-relative, malformed, or otherwise unapproved destinations are rejected or replaced with the safe default.

**Given** Findur constructs the authorization request
**When** it redirects to SnapTrade's hosted authorization endpoint
**Then** provider endpoints come from the validated OIDC discovery document
**And** only the approved `openid read` scopes are requested
**And** email, profile, trade, webhook, or broader scopes are not requested
**And** the exact configured HTTPS callback URI is used.

**Given** provider discovery or authorization initialization fails
**When** Findur handles the failure
**Then** it presents a safe retry path without raw provider bodies, secrets, or configuration details
**And** it leaves no incomplete reusable authorization attempt
**And** any already-authenticated session remains unaffected.

**Given** an authorization attempt is expired or already claimed
**When** it is submitted again
**Then** it cannot be reused.

**Given** automated tests exercise authorization initiation
**When** the generated attempt and redirect are inspected
**Then** all state, nonce, PKCE, cookie, scope, callback, expiry, and allowlist requirements are verified against local provider fixtures
**And** no live SnapTrade request or real credential is used.

**Given** callback completion is not yet available in the deployed build
**When** that build is released
**Then** the public authorization action remains safely feature-gated
**And** a visitor cannot be redirected into an incomplete authentication flow.

### Story 0.3: Complete OAuth and Establish an Isolated Session

As an eligible test user,
I want Findur to securely complete SnapTrade authorization,
So that I can enter a private session without exposing provider credentials or tokens to my browser.

**Acceptance Criteria:**

**Given** SnapTrade redirects to the registered callback
**When** Findur receives the authorization response
**Then** it validates the correlation cookie, state, attempt expiry, and exact callback context before exchanging the code
**And** atomically claims the authorization attempt so only its first valid callback may proceed.

**Given** the callback contains a provider error, missing parameter, mismatched binding, expired attempt, or unregistered destination
**When** Findur evaluates it
**Then** no authorization code is exchanged
**And** the user receives a safe restart path without raw provider details, secrets, or protected data.

**Given** a valid authorization attempt has been claimed
**When** Findur exchanges the authorization code
**Then** it uses the stored PKCE verifier and confidential-client authentication required by SnapTrade
**And** sends the request outside any database transaction with a bounded timeout
**And** never sends or exposes the client secret, code, or returned tokens to the browser.

**Given** SnapTrade returns tokens
**When** Findur validates the ID token
**Then** it verifies the RS256 signature, discovered issuer, audience, expiry, issued-at plausibility, and original nonce
**And** rejects an invalid or unverifiable token before creating an identity or session
**And** does not call a user-info endpoint or require email/profile data.

**Given** the verified SnapTrade subject belongs to an active Findur user
**When** callback finalization succeeds
**Then** the same isolated user is resumed
**And** no second user is created.

**Given** the verified SnapTrade subject has no remaining Findur identity
**When** callback finalization succeeds
**Then** a new OAuth-origin user is created
**And** a subject previously erased through disconnect is treated as new rather than silently restoring deleted state.

**Given** valid access and refresh tokens are received
**When** the authorization is persisted
**Then** each token is independently encrypted with AES-256-GCM using a current versioned key
**And** associated data binds the envelope to its owner, provider, token kind, and envelope version
**And** no plaintext token, authorization code, ID token, client secret, or session secret is stored or logged.

**Given** identity and authorization validation succeeds
**When** Findur finalizes the callback
**Then** identity binding, encrypted authorization storage, session creation, and the attempt's terminal result commit atomically
**And** no provider network operation occurs inside that transaction.

**Given** callback finalization creates a Findur session
**When** the response returns through the public same-origin deployment
**Then** it expires the pre-login correlation cookie
**And** issues a cryptographically random opaque session cookie marked Secure, HttpOnly, host-only, and SameSite=Lax
**And** PostgreSQL stores only keyed hashes of the session identifier and its session-bound CSRF token.

**Given** token issuance may have succeeded but local finalization fails
**When** Findur handles the ambiguous outcome
**Then** it discards plaintext token material immediately
**And** makes at most one bounded best-effort revocation when a token is known
**And** marks the attempt `restart-required` without retrying the authorization code.

**Given** an already-consumed callback is replayed
**When** Findur handles it
**Then** it never exchanges the code again
**And** an already-authenticated browser follows the recorded safe internal destination
**And** an unauthenticated browser receives a safe authorization restart.

**Given** callback completion succeeds
**When** Findur chooses the post-login destination
**Then** it restores only a previously allowlisted internal route whose prerequisites are satisfied
**And** otherwise routes to the safe authenticated onboarding destination.

**Given** callback and session behavior is tested automatically
**When** the suite runs against WireMock and PostgreSQL
**Then** it covers success, denial, tampered state, cookie mismatch, expiry, replay, invalid ID-token claims, identity resumption/creation, encrypted storage, transaction failure, and revocation compensation
**And** uses no live SnapTrade account or real credential.

**Given** the complete authorization flow passes its automated checks
**When** the build is released
**Then** the authorization feature gate may be enabled
**And** the public owner-entry path completes rather than entering a partial flow.

### Story 0.4: Protect the Authenticated Experience and End Sessions

As an authenticated user,
I want my Findur session and protected routes to remain isolated and controllable,
So that only I can access my private experience and I can end that access safely.

**Acceptance Criteria:**

**Given** a request carries a valid, unexpired opaque session cookie
**When** session middleware authenticates it
**Then** it derives an immutable Actor from the server-side session
**And** owner-scoped use cases ignore and reject any client-supplied acting user identifier
**And** another authenticated user cannot read or mutate the first user's resources by guessing opaque IDs.

**Given** an unauthenticated, expired, or revoked session requests a protected route
**When** routing evaluates access
**Then** no protected content flashes or enters browser storage
**And** the user is sent to a named authentication recovery surface with only an allowlisted intended destination.

**Given** an authenticated user enters the application shell
**When** navigation is rendered
**Then** it exposes exactly Discovery, Portfolio, and Profile
**And** uses bottom navigation below 768px and an accessible desktop rail at or above 768px
**And** an incomplete user is routed to the earliest unmet prerequisite rather than shown misleading content.

**Given** a session-authenticated unsafe request is submitted
**When** the API authorizes it
**Then** it requires the session-bound `X-CSRF-Token`, a valid same-origin Origin or permitted Referer fallback, and acceptable Fetch Metadata
**And** failure produces a non-sensitive rejection without performing the mutation
**And** safe methods and the OAuth callback remain the only exemptions.

**Given** a session remains active
**When** idle or absolute expiry policy is evaluated
**Then** the checked-in 12-hour idle and 7-day absolute limits are enforced server-side
**And** session use never extends the absolute limit.

**Given** a user chooses Log out
**When** the request succeeds
**Then** only the current session is revoked and its cookie expired
**And** protected in-memory query/navigation state is cleared before the public destination renders.

**Given** a user chooses Log out all sessions
**When** the request succeeds
**Then** every current Findur session for that user is revoked
**And** all affected browsers must authenticate again.

**Given** private API responses or authenticated UI state exist
**When** the application caches, logs, installs, or updates document metadata
**Then** protected payloads are never placed in local storage, session storage, IndexedDB, Cache Storage, service-worker caches, URLs, titles, manifests, favicons, or social metadata
**And** protected API responses use `private, no-store`
**And** only hashed public assets are cacheable.

**Given** locale, theme, route, focus, or responsive mode changes
**When** the authenticated shell updates
**Then** English/French and System/Light/Dark behavior remains functionally equivalent
**And** the permitted task and meaningful focus are preserved
**And** controls meet keyboard, touch-target, contrast, zoom, reflow, and reduced-motion requirements.

**Given** authenticated navigation and shared actions are implemented
**When** their visual and interaction states render
**Then** they use the semantic Constellation tokens and React Aria-based navigation, action, link, selectable-row, and passive-surface primitives
**And** hover/pressed states do not move, scale, reflow, or change border width
**And** focus-visible remains the strongest interaction state.

**Given** the web app is installed or shown in standalone presentation
**When** launch, safe-area, icon, chrome, or task-switcher surfaces appear
**Then** they use only the opaque branded constellation mark and generic privacy-safe metadata
**And** installability does not add a service worker, offline protected experience, or recent-content exposure.

### Story 0.5: Prove Authenticated SnapTrade Access with Masked Inventory

As a newly authenticated user,
I want Findur to retrieve my connection status and masked eligible accounts,
So that I can verify the SnapTrade integration before choosing any financial account for use.

**Acceptance Criteria:**

**Given** OAuth callback finalization created an active authorization
**When** the post-authorization bootstrap runs for the first time
**Then** it makes one bounded, purpose-limited inventory operation through the `PortfolioProvider` adapter
**And** lists connection lifecycle state before listing accounts for each connection
**And** runs all provider calls outside database transactions.

**Given** the inventory operation calls SnapTrade
**When** WireMock captures each outbound request
**Then** it carries only the user's server-held OAuth bearer token
**And** excludes Commercial authentication fields and every trading capability
**And** conforms to the pinned request/response fixtures.

**Given** SnapTrade returns connection and account inventory
**When** Findur maps and persists it
**Then** it retains only stable connection/account identifiers, safe brokerage label, account category/type, masked display label, availability, sync mode/state, and fields strictly required for later account selection
**And** raw responses, full account identifiers, balances, positions, activities, orders, and unused fields are neither persisted nor logged
**And** no account is included by default.

**Given** an account inventory field is missing or unsupported
**When** Findur presents the inventory
**Then** it labels the field or account honestly as unavailable or ineligible
**And** never fabricates or infers financial values.

**Given** inventory retrieval is pending, rate-limited, unavailable, unauthorized, or fails normalization
**When** the authenticated result screen renders
**Then** it identifies the categorical state, what remains trustworthy, and an exact safe retry, repair, or reauthorization action
**And** exposes no raw provider body or credential
**And** never represents a failed or disabled connection as current.

**Given** the result screen is rerendered, revisited, or restyled
**When** no explicit retry or freshness policy requires work
**Then** it serves the persisted normalized inventory
**And** does not make another SnapTrade request merely because the page rendered.

**Given** inventory retrieval succeeds
**When** the user reviews the OAuth outcome
**Then** the UI confirms connection separately from account inclusion
**And** explains that balances, holdings, activities, derivation, previews, and Discovery are still unavailable until explicit account confirmation
**And** focuses the outcome heading before offering the next Account Inclusion step.

**Given** automated verification runs in CI
**When** success, empty, disabled, unauthorized, rate-limit, malformed, and transient-failure fixtures execute
**Then** request shape, minimization, normalization, isolation, no-default-inclusion, and safe recovery are proven against WireMock
**And** no live credentials or user financial data are required.

## Epic 1: Explore and Control Live Portfolio Data

The authenticated user can explicitly choose which connected accounts Findur may use, inspect the supported live SnapTrade account, balance, position, and recent-activity data with honest dataset-specific freshness and coverage, receive bounded activity-driven refresh without provider-call storms, understand how source data feeds the product, recover from provider states, and disconnect SnapTrade to leave Findur with destruction of the complete user aggregate.

### Story 1.1: Select and Confirm Included Accounts

As an authenticated user,
I want to explicitly choose which eligible accounts Findur may use,
So that no portfolio data is retrieved or used without my separate, scoped confirmation.

**Acceptance Criteria:**

**Given** masked account inventory is available
**When** Account Inclusion opens
**Then** it renders a named checkbox group with no account included by default
**And** distinguishes the editable draft from committed coverage
**And** shows usable rows and disabled rows with specific reasons
**And** provides an accessible tri-state Select All that affects only usable rows.

**Given** the user changes the draft selection
**When** they review the confirmation summary
**Then** it names added and removed masked accounts
**And** explains that confirmed accounts may be retrieved for the owner-only showcase, private derivation/matching, and disclosure-controlled previews
**And** explains that exclusion purges affected source, derived, cached, preview, and Discovery outputs
**And** Cancel returns focus to the invoker without changing committed coverage.

**Given** the user confirms one or more additions
**When** inclusion processing begins
**Then** additions remain pending and cannot broaden any view or downstream use
**And** balances, positions, and the bounded recent-activities window are retrieved only for those pending accounts
**And** provider calls run outside database transactions.

**Given** every required dataset for an added account is fetched and normalized successfully
**When** the guarded finalizer publishes the change
**Then** immutable dataset versions and the new inclusion membership become committed atomically
**And** the account becomes available to the private Portfolio Showcase
**And** the operation is idempotent under duplicate submission.

**Given** an addition fails, times out, is unsupported, or loses its lifecycle/version guard
**When** the operation settles
**Then** the prior narrower committed inclusion set remains authoritative
**And** transient results are discarded without publishing partial datasets
**And** the user sees a scoped recovery action.

**Given** the user confirms removal of a committed account
**When** the prepare transaction succeeds
**Then** the account is removed from committed inclusion immediately
**And** the lifecycle/inclusion generation advances before deletion
**And** its dataset versions, heads, derived outputs, anchors, and cached or rendered representations are purged without retention grace
**And** any stale in-flight finalizer is unable to recreate them.

**Given** a change contains both removals and additions
**When** an addition later fails
**Then** confirmed removals remain effective and deleted
**And** the failed additions do not broaden the last committed narrower set.

**Given** account inclusion is mutated or read
**When** authorization is evaluated
**Then** the server scopes the operation from the authenticated Actor and resource ownership
**And** cross-user IDs cannot reveal or change another user's inventory or selection.

### Story 1.2: Inspect the Private Portfolio Showcase

As an authenticated user,
I want to inspect the supported data from my Included Accounts,
So that I can understand exactly what Findur received before it powers any dating experience.

**Acceptance Criteria:**

**Given** an Included Account has successfully published data
**When** the Portfolio Showcase loads
**Then** it presents connection state, committed account coverage, balances/values, positions/holdings, and recent activities as distinct datasets
**And** every dataset shows source, observation/retrieval time, publication time, coverage, and Freshness State
**And** it never claims the portfolio is complete net worth, financial health, responsibility, or identity.

**Given** normalized provider data is published
**When** it is stored
**Then** connections/accounts and each per-account dataset use immutable typed versions with scoped atomic heads
**And** a failed fetch or normalization leaves the prior trustworthy head unchanged
**And** raw provider payloads are never persisted.

**Given** money or decimal values are available
**When** the API and UI represent them
**Then** decimal precision is preserved using database numeric values and decimal strings
**And** money always carries its ISO-4217 currency
**And** different currencies are neither silently converted nor aggregated
**And** unavoidable upstream float precision loss is explicitly tested and bounded.

**Given** a balance, position, activity field, currency, account total, or whole dataset is empty, missing, unsupported, stale, syncing, or failed
**When** its section renders
**Then** the state is shown explicitly rather than as zero or fabricated data
**And** excluded-row/account coverage is reported where material
**And** unsupported performance remains unavailable rather than inferred.

**Given** recent activities are displayed
**When** their coverage is described
**Then** the bounded date window, row ceiling, available date precision, and dataset freshness are visible
**And** orders, tax lots, trading, quotes, reference enrichment, and provider performance endpoints remain absent.

**Given** the user changes viewport, zoom, theme, language, input method, or assistive technology
**When** the Portfolio Showcase reflows
**Then** its visualizations have programmatic titles/summaries and synchronized accessible tables or equivalents
**And** wide data has a non-table summary and bounded overflow
**And** meaning does not depend on color, hover, motion, node size, or position.

**Given** one authenticated user requests showcase data
**When** private projections are produced
**Then** only that Actor's committed Included Accounts are serialized
**And** responses are `private, no-store`
**And** another real user's live data can never enter the response.

**Given** a dataset version is no longer current
**When** bounded cleanup evaluates it
**Then** it is retained only while an authorized anchor pins it or during the configured short cleanup grace
**And** Findur never accumulates indefinite financial history
**And** account exclusion or disconnect overrides the grace immediately.

### Story 1.3: Refresh Portfolio Data Without Provider Call Storms

As an authenticated user,
I want portfolio data to refresh thoughtfully and show its true freshness,
So that I can trust what I see without Findur abusing provider limits or making billable requests.

**Acceptance Criteria:**

**Given** the user performs an authenticated activity that materially needs portfolio freshness
**When** dataset policy is evaluated
**Then** holdings are current through 15 minutes in real-time mode or 36 hours in Daily mode and stale-usable through 72 hours
**And** activities are current through two calendar days and stale-usable through seven days
**And** missing initial data, revoked authorization, or a disabled connection is unusable.

**Given** data remains current under its dataset policy
**When** the user navigates or revisits a view
**Then** Findur serves the persisted normalized snapshot
**And** no provider request is made.

**Given** data needs revalidation but remains stale-usable
**When** an eligible authenticated activity triggers refresh
**Then** the last permitted trustworthy view stays visible with its timestamp and stale state
**And** background work does not reset focus or the current task
**And** material changes are announced before rearranging the experience.

**Given** a view has no trustworthy snapshot yet
**When** its first bounded load is pending
**Then** any loading skeleton matches the expected layout, imitates no real value, and has an external accessible loading label
**And** skeletons are not substituted for stale-while-revalidate when a permitted trustworthy view exists.

**Given** page rendering, synthetic-candidate activity, health/readiness/status probes, deployment smoke, polling heartbeats, or public traffic occurs
**When** refresh triggers are inspected
**Then** they make zero SnapTrade requests
**And** no billable provider manual-refresh operation is invoked automatically.

**Given** concurrent requests demand the same refresh
**When** refresh coordination runs
**Then** one generation-guarded operation performs the bounded provider work
**And** other callers reuse current data or wait/retry within a strict bound
**And** no database transaction remains open during network work.

**Given** the provider responds with `429`, `Retry-After`, reset headers, transient failure, or an open circuit
**When** Findur schedules further work
**Then** it honors the strongest valid provider delay, applies bounded retry/circuit policy, and surfaces an honest pending or stale state
**And** it does not spin, fan out, or bypass limits.

**Given** the access token needs refresh
**When** authorization coordination runs
**Then** one version/generation-guarded lease exchanges the rotating refresh token outside a transaction
**And** a compare-and-swap finalizer installs new encrypted envelopes only for the winning current lease
**And** ambiguous delivery or failed guarded publication transitions to reauthorization-required without replaying the refresh token.

**Given** a provider API call returns `401`
**When** recovery is safe
**Then** Findur refreshes once and retries the safe request once
**And** a second failure clears usable authorization and requires reauthorization.

**Given** refresh completes after inclusion, authorization, or user lifecycle has changed
**When** the finalizer checks its authorization generation, inclusion version, and required membership
**Then** stale results are discarded
**And** they cannot recreate deleted datasets, credentials, signals, or navigation anchors.

**Given** freshness becomes unusable or the connection is disabled
**When** the view transitions
**Then** it moves to a focused Recovery Panel that distinguishes brokerage repair from OAuth reauthorization
**And** never presents the old data as current or silently changes account selection.

### Story 1.4: Trace Source Data to Product Use

As an authenticated user or evaluator,
I want to trace how supported SnapTrade data can flow through Findur,
So that I can distinguish owner-only source facts, private matching inputs, and disclosure-controlled visible output.

**Acceptance Criteria:**

**Given** the user opens Source to Experience from the Portfolio Showcase
**When** the trace renders
**Then** it presents three explicit branches: owner-only exact view, private matching input, and candidate-visible disclosure
**And** provider, normalized, derived, and projected stages are visually and programmatically distinct
**And** Disclosure Level filtering appears only on the candidate-visible branch.

**Given** supported account, balance, position, or activity fields are traced
**When** a stage is available
**Then** it names the source dataset, included-account coverage, applicable freshness, transformation version, and destination purpose
**And** marks owner-only exact fields separately from derived values
**And** carries live-owner provenance throughout.

**Given** matching signals or disclosure choices have not yet been configured
**When** their downstream branches are inspected
**Then** the trace shows the stage as not yet configured or blocked
**And** explains the exact prerequisite without fabricating an output or depending on Discovery.

**Given** a field or stage is missing, unsupported, excluded, stale-unusable, or filtered by disclosure
**When** the trace displays that path
**Then** it uses explicit text and a strong dashed blocked treatment
**And** never substitutes zero, synthetic data, or a universal compatibility/wealth score.

**Given** the trace is used with keyboard, touch, screen reader, reduced motion, forced colors, EN/FR, light/dark, phone, or desktop
**When** a branch is explored
**Then** complete equivalent information remains available without hover, color, motion, layout, or node size as the sole carrier of meaning.

### Story 1.5: Disconnect SnapTrade and Leave Findur

As an authenticated user,
I want disconnecting SnapTrade to erase my Findur presence,
So that my identity, financial information, and product activity do not remain in app-controlled active storage.

**Acceptance Criteria:**

**Given** the user invokes Disconnect SnapTrade
**When** the confirmation dialog opens
**Then** it clearly states that disconnect means leaving Findur
**And** summarizes deletion of identity, profile, preferences, disclosure, sessions, provider authorization, portfolio data, derived data, swipes/matches, and cached/restorable state
**And** separates guaranteed local deletion from best-effort provider revocation and infrastructure backup policy
**And** Cancel is neutral and returns focus to the invoker without changing state.

**Given** the user confirms the destructive action
**When** disconnect preparation commits
**Then** the user/authorization lifecycle generation advances and status blocks all new sessions, refreshes, provider calls, and publications
**And** every Findur session is revoked immediately
**And** client in-memory protected state is cleared before a safe public leaving state renders.

**Given** provider revocation can be attempted
**When** disconnect owns the mutually exclusive authorization lease
**Then** it revokes the newest known refresh token through the discovered endpoint outside any database transaction
**And** waits only to the bounded lease/call deadline
**And** never delays local destruction indefinitely.

**Given** provider revocation succeeds, fails, times out, or is ambiguous
**When** local finalization runs
**Then** the complete user aggregate is explicitly or safely cascade-deleted from app-controlled active storage
**And** deletion includes the user, external identity, OAuth attempts/authorization, tokens, sessions, profile, locale/theme, preferences, disclosure, connections/accounts, inclusion, dataset versions/rows/heads, signals, anchors, swipes, matches, and user-scoped caches
**And** no tombstone or relationship shell is retained absent a demonstrated integrity requirement.

**Given** provider, refresh, normalization, derivation, or publication work was already in flight
**When** it attempts to finalize after disconnect preparation
**Then** generation and existence guards force it to discard transient results
**And** it cannot recreate any user-owned row, credential, signal, or cache.

**Given** local deletion partially fails
**When** the operation recovers or is retried
**Then** the user remains blocked and broader state remains suppressed
**And** the idempotent deletion resumes without restoring access or data
**And** safe categorical diagnostics contain no PII, financial values, tokens, or provider bodies.

**Given** the same SnapTrade subject authorizes Findur later
**When** OAuth callback identity binding runs
**Then** the deleted mapping is absent and a new Findur user is created
**And** no prior profile, portfolio, preference, disclosure, interaction, or match state is restored.

**Given** deletion behavior is integration-tested
**When** cross-user, duplicate, concurrent-refresh, revocation-failure, rollback, restart, and stale-finalizer scenarios execute
**Then** the departing user's complete aggregate is absent
**And** unrelated users and shared generated candidates remain intact.

## Epic 2: Transform SnapTrade Data into a Dating Profile

The user can turn the included live portfolio into the approved private compatibility signals, complete a personal profile, choose distance and compatibility preferences, explicitly control disclosure, and privately preview the resulting live-owner presentation across disclosure and identity states before entering Discovery.

### Story 2.1: Derive a Private Portfolio Matching Profile

As an authenticated user,
I want Findur to derive a small, explainable matching profile from my Included Accounts,
So that portfolio compatibility can work without exposing raw financial detail or inventing certainty.

**Acceptance Criteria:**

**Given** the user has current or stale-usable included position datasets
**When** the versioned signal builder runs
**Then** it derives only asset-kind allocation, asset-kind breadth, and largest-position concentration
**And** stores source dataset versions, inclusion version, coverage, freshness, algorithm version, and live-owner provenance with the immutable signal set
**And** no universal compatibility, wealth, responsibility, or financial-health score is created.

**Given** a position has units, price, and currency
**When** its value contributes to a vector
**Then** value is computed with decimal arithmetic
**And** it is compared only against an account total in the same currency
**And** binary-float artifacts and implicit currency conversion are excluded.

**Given** multiple Included Accounts have usable vectors
**When** the user-level signal is assembled
**Then** usable account vectors are equally weighted
**And** cash is excluded from the ranking vector
**And** excluded rows/accounts and their reasons reduce reported coverage rather than becoming zero.

**Given** required source values are missing, unsupported, stale-unusable, or internally inconsistent
**When** derivation evaluates them
**Then** affected inputs remain unknown
**And** unsupported performance or synthetic substitutes are not inferred
**And** a signal set is usable only when its documented minimum coverage is met.

**Given** included source heads change through refresh or inclusion
**When** recalculation completes successfully
**Then** the new signal set and its active head publish atomically under lifecycle and inclusion guards
**And** a failed or stale calculation leaves the prior still-permitted trustworthy head unchanged
**And** a permission decrease removes affected signals immediately.

**Given** Source to Experience is reopened after derivation
**When** the private matching branch renders
**Then** it names the derived traits, source coverage, freshness, and version without exposing exact formulas or raw account values
**And** the candidate-visible branch remains controlled separately by a saved Disclosure Level.

### Story 2.2: Complete the Minimum Personal Profile

As an authenticated user,
I want to create and edit the minimum personal profile needed for the demonstration,
So that my portfolio evidence has a meaningful, adult dating context.

**Acceptance Criteria:**

**Given** the user opens Personal Profile
**When** fields render
**Then** the form includes display name, adult attestation, coarse location area, relationship intent, short biography or prompt response, bundled avatar choice, locale, and theme
**And** every field has a persistent label, required/private/visibility annotation, help and associated error text, and at least a 44px target.

**Given** the user selects a location
**When** available choices are presented
**Then** they come only from the fixed migrated coarse city/region catalogue
**And** Findur never requests browser/device location, infers from IP, accepts coordinates, street address, or postal code, or exposes centroids.

**Given** the user selects an avatar
**When** the profile is validated
**Then** only an allowlisted bundled placeholder key is accepted
**And** arbitrary URLs, uploads, binary media, object storage, and image-processing endpoints do not exist.

**Given** the user attests adult status
**When** the profile is saved
**Then** Findur stores the attestation timestamp rather than a date of birth
**And** the experience preserves explicit 18+ demonstration positioning and safety guidance.

**Given** any required value is missing, invalid, too long, or outside an allowlist
**When** Save is submitted
**Then** no partial first profile is created
**And** errors are associated, summarized, and focused deterministically without discarding valid entered values.

**Given** all required values are valid
**When** Save succeeds
**Then** the first profile is created atomically and later edits use optimistic concurrency
**And** pre-login English/French and System/Light/Dark choices are copied into authenticated settings
**And** a conflicting edit preserves the user's draft and offers safe reload/reapply guidance.

**Given** profile content is rendered
**When** language, theme, phone/desktop layout, zoom, forced colors, or assistive technology changes
**Then** meaning, actions, focus, reading order, 35% French expansion, contrast, and reflow remain equivalent
**And** copy avoids wealth, luxury, financial-shaming, trading-urgency, or worth-based language.

### Story 2.3: Set Discovery Distance and Compatibility Preferences

As an authenticated user,
I want to choose a maximum distance and a portfolio compatibility style,
So that I can shape Discovery using understandable boundaries without being shown a manipulative score.

**Acceptance Criteria:**

**Given** the user opens Discovery Preferences
**When** controls render
**Then** they can enter a bounded positive maximum distance and select exactly one of similar, diversified, or complementary
**And** selected state, help, validation, keyboard/touch operation, and save state are explicit.

**Given** compatibility modes are explained
**When** the user compares them
**Then** similar is described as favoring closer asset-kind allocation
**And** diversified as favoring broader instrument-kind coverage and lower largest-position concentration
**And** complementary as favoring lower asset-kind allocation overlap
**And** no internal score, precise formula, Invisible Cohort, guarantee, or worth judgment is exposed.

**Given** maximum distance is saved
**When** eligibility later evaluates a candidate
**Then** the boundary is inclusive and calculated server-side from coarse area centroids using tested Haversine logic
**And** users see only localized area labels and configured approximate distance bands, never coordinates or exact distance.

**Given** valid preferences are submitted
**When** persistence succeeds
**Then** they replace the prior version atomically using optimistic concurrency
**And** a distance change invalidates the current candidate anchor
**And** a failed or conflicting save leaves the prior committed preferences authoritative.

**Given** the saved preference summary is shown
**When** its effect is explained
**Then** it states that compatibility, maximum distance, saved Disclosure Level, freshness, and candidate availability shape Discovery
**And** it does not imply deterministic matching or disclose ranking internals.

### Story 2.4: Preview and Save a Disclosure Level

As an authenticated user,
I want to inspect and explicitly save my Disclosure Level,
So that I understand exactly what could be visible before Discovery becomes available.

**Acceptance Criteria:**

**Given** the user has never saved a Disclosure Level
**When** the control opens
**Then** Snapshot, Holdings, and Full Detail appear as one named single-select group with no default selection
**And** Snapshot is neutrally marked as the recommended starting point without being preselected
**And** a persistent summary distinguishes no saved level, saved level, and previewing-not-saved state.

**Given** the user inspects Snapshot
**When** its field contract is shown
**Then** it permits allocation/asset-class mix, diversification context, value/activity bands, activity recency, approximate proximity, coverage, and Freshness State
**And** prohibits security names and exact monetary values.

**Given** the user inspects Holdings
**When** its field contract is shown
**Then** it may add supplied instrument names/symbols/kinds, position weights, defensible percentage performance, and activity categories
**And** prohibits quantities, exact monetary values, and transaction detail
**And** unsupported performance is explicitly unavailable.

**Given** the user inspects Full Detail
**When** its field contract is shown
**Then** it may add supplied quantities, last-known prices, cost basis when supported, exact values by currency, account/portfolio value, defensible performance amounts, and recent historical activities
**And** it does not invent unavailable orders, timestamps, performance, or cross-currency totals.

**Given** any level is previewed
**When** the preview changes
**Then** hidden, bucketed, derived, and exact examples update without mutating the saved level
**And** source facts remain distinct from derived values
**And** screenshots, memory, and combined inference limits are stated before save, with the strongest warning adjacent to Full Detail.

**Given** the user explicitly saves an initial level or an upgrade
**When** the scoped confirmation and persistence succeed
**Then** the selected level becomes the only saved policy version
**And** no broader candidate projection or Discovery eligibility exists before that success
**And** failure preserves the prior narrower policy.

**Given** the user confirms a downgrade
**When** the decrease begins
**Then** higher-detail active, prefetched, rendered, cached, history-restored, preview-share, and revalidation output is suppressed immediately
**And** the narrower policy remains authoritative even if persistence must retry
**And** still-consented owner-private source and matching data are not unnecessarily deleted.

**Given** the user cancels or attempts to leave with a changed draft
**When** route protection runs
**Then** Cancel restores the saved preview
**And** dirty exit offers only Discard changes or Stay with deterministic focus
**And** no selection autosaves.

### Story 2.5: Preview the Complete Live-Owner Profile and Readiness

As an authenticated user,
I want to preview my complete dating presentation across privacy and identity states,
So that I can verify the experience before entering Discovery.

**Acceptance Criteria:**

**Given** a saved profile and Usable Portfolio exist
**When** Profile Preview opens
**Then** it combines personal information with the same disclosure projections and portfolio visualization components used by Candidate Cards
**And** it is clearly labelled owner-private and live-owner
**And** live data is never inserted into any Discovery deck.

**Given** the user operates preview controls
**When** they switch Snapshot/Holdings/Full Detail, pre-match/post-match identity, or representative phone/desktop composition
**Then** only the inspected preview changes
**And** saved disclosure, profile, preferences, and eligibility remain unchanged
**And** each state shows the same permitted and prohibited fields as its corresponding product state.

**Given** pre-match identity is previewed
**When** the frame renders
**Then** the photo uses the explicit obscure state without hidden image description or cached preview
**And** portfolio evidence leads while identity remains subordinate.

**Given** post-match identity is previewed
**When** the frame renders
**Then** the same-size bundled placeholder avatar replaces the obscure state without layout jump
**And** no chat, upload, or second acceptance action appears.

**Given** preview data is missing, blocked, unsupported, revalidating, invalidated, or stale-unusable
**When** the relevant state renders
**Then** it names what remains trustworthy and the exact recovery action
**And** never fabricates, silently broadens disclosure, or restores invalidated content.

**Given** readiness is evaluated
**When** the server derives the result
**Then** Discovery is ready only with a complete valid profile, saved preferences, explicitly saved disclosure head, non-empty committed inclusion, and usable included positions/signals
**And** authentication alone or a client claim cannot mark a user ready
**And** every unmet prerequisite links to its focused recovery surface.

**Given** the preview is exercised across supported presentation modes
**When** accessibility and responsive tests run
**Then** phone/desktop, EN/FR, light/dark, 200% and applicable 400% reflow, keyboard, screen reader, touch, forced color, and reduced motion preserve equivalent content and actions
**And** no protected content enters browser persistence or ambient metadata.

## Epic 3: Demonstrate Portfolio-First Discovery and Matching

The ready user can explore a reproducible population of synthetic candidates through disclosure-controlled portfolio-first cards and detail, understand why candidates appear, make one durable Pass or Interested decision, handle honest Discovery edge states, review in-session generated interest, and reach a one-time Mutual Match with placeholder-avatar reveal.

### Story 3.1: Generate a Reproducible Synthetic Population

As a project evaluator,
I want a deep and reproducible population of varied synthetic candidates,
So that the complete Discovery experience can be demonstrated without exposing any real person's data.

**Acceptance Criteria:**

**Given** an explicit algorithm version, seed, and population size
**When** the generation command runs against clean synthetic state
**Then** it creates stable generated user IDs, profiles, coarse locations, portfolio snapshots, signals, disclosure settings, presence, reciprocal outcomes, and scenario tags deterministically
**And** rerunning the same command is idempotent.

**Given** generated portfolio data is built
**When** it enters the domain
**Then** it uses the same typed snapshot and signal schemas/builders as OAuth-origin data
**And** every generated user has immutable `origin: generated`
**And** no generated user can own an OAuth identity, authorization, provider token, or Findur session.

**Given** the generator validates its scenario matrix
**When** population activation is attempted
**Then** it requires credible coverage across all Disclosure Levels, compatibility modes, inside/at/outside distance boundaries, portfolio patterns, sparse/unsupported inputs, freshness states, presence states, and reciprocal/non-reciprocal outcomes
**And** all generated profiles are unambiguously adult
**And** an incomplete matrix fails before activation.

**Given** generator input is inspected
**When** data provenance is verified
**Then** it contains no OAuth-origin row, live portfolio export, provider payload, real user record, or external personal data
**And** synthetic provenance remains explicit in storage and projections.

**Given** generated state advances over time
**When** the bounded deterministic tick runs
**Then** one PostgreSQL transaction acquires `pg_try_advisory_xact_lock` and updates presence plus coherent portfolio events
**And** concurrent tick attempts do no duplicate work
**And** missed intervals are coalesced rather than replayed without bound
**And** clock and randomness are injectable for tests.

**Given** a tick changes generated data
**When** invariants are checked
**Then** cross-field portfolio values remain coherent, required minimum online inventory is preserved, scenario identity remains stable, and affected signals publish atomically
**And** no real viewer's data or authorization state is read or mutated.

### Story 3.2: Assemble an Eligible and Ordered Swipe Deck

As a Discovery-ready user,
I want a credible sequence of eligible synthetic candidates,
So that I can explore portfolio compatibility within my saved privacy and distance boundaries.

**Acceptance Criteria:**

**Given** the authenticated viewer is server-derived as Discovery-ready
**When** candidate eligibility is calculated
**Then** only generated candidates with valid profiles, usable signals, permitted freshness/presence, no prior viewer decision, compatible disclosure, and inclusive maximum-distance eligibility may appear
**And** no OAuth-origin user can ever become a candidate.

**Given** regular Discovery compares disclosure policies
**When** candidate eligibility is filtered
**Then** a candidate's saved Disclosure Level must be equal to or lower than the viewer's saved level
**And** the filter is never relaxed to increase inventory.

**Given** eligible candidates remain
**When** ordering uses the viewer's saved mode
**Then** similar orders by smallest asset-kind allocation difference
**And** diversified orders by broader kind coverage then lower largest-position concentration
**And** complementary orders by lower asset-kind allocation overlap
**And** approximate proximity is secondary and stable candidate ID is the final tie-breaker
**And** no internal score or Invisible Cohort is returned to the client.

**Given** a current candidate is selected
**When** Findur prepares navigation state
**Then** it persists only a short-lived anchor pinning candidate, signal versions, disclosure version, offered time, expiry, and safe navigation position
**And** it does not persist a full deck
**And** recalculation remains deterministic from authoritative state.

**Given** the candidate pool is sparse or empty
**When** Discovery renders
**Then** it states the eligible-inventory condition honestly and offers a safe recovery or retry
**And** never widens distance, disclosure, freshness, consent, or other eligibility silently
**And** exposes no cohort size or hidden population membership.

**Given** cached generated state and viewer signals are available
**When** Discovery opens on ordinary broadband
**Then** the first actionable card or explicit recovery state becomes interactive within three seconds excluding provider work
**And** opening Discovery itself triggers no SnapTrade request.

**Given** the viewer enters Discovery for the first time or from a relevant recovery
**When** trust guidance is presented
**Then** it concisely states 18+, read-only/non-advisory use, no investment solicitation, no money requests, no financial targeting, and no wealth/responsibility/identity inference
**And** preserves the demonstration-only boundary without advertising Guest Demo or public signup.

### Story 3.3: Explore a Portfolio-First Candidate Card and Detail

As a Discovery-ready user,
I want to inspect a candidate's permitted portfolio evidence before deciding,
So that the swipe decision reflects compatibility while preserving privacy and synthetic provenance.

**Acceptance Criteria:**

**Given** Discovery has a current anchored candidate
**When** the Candidate Card renders
**Then** one focal card leads with a Portfolio Node Visualization and includes an evidence/relevance summary, approximate proximity, coverage, freshness, Disclosure Level, Photo Obscure, and explicit Pass/Interested actions
**And** it shows textual deck position with at most one decorative next-card outline
**And** portfolio evidence remains visually dominant over identity.

**Given** any generated candidate surface is shown
**When** identity is presented
**Then** a persistent Synthetic Data label remains visually and programmatically adjacent to identity
**And** responsive reflow, expansion, detail navigation, incoming interest, and match state cannot remove it.

**Given** a candidate projection is generated
**When** the candidate's Disclosure Level is applied
**Then** the server serializes only fields allowed by that level using the dedicated pre-match schema
**And** higher-tier values and unrevealed avatar keys are absent rather than CSS-hidden
**And** values retain source/derived labels, currency, period, coverage, and freshness qualifiers.

**Given** the Photo Obscure is shown pre-match
**When** the card or accessibility tree is inspected
**Then** it exposes no hidden image pixels, URL, cached preview, or hidden-image description
**And** its geometric state is understandable without color or motion.

**Given** the user opens Candidate Detail
**When** the protected route loads
**Then** focus moves to its heading and the same candidate, disclosure version, provenance, freshness, photo state, and pending decision are preserved
**And** phone uses a full-screen composition while desktop uses a two-column composition
**And** only disclosure-permitted progressive sections exist in the response.

**Given** the user follows a valid deep link or activates Back
**When** route/session/readiness/anchor gates pass
**Then** the exact safe current-card position and available decision are restored
**And** an invalid or unauthorized link renders a named recovery destination without protected-content flash.

**Given** charts or portfolio nodes render
**When** a user cannot perceive or operate the visual treatment
**Then** a programmatic title/summary and synchronized accessible data table or equivalent conveys labels, values, units, periods, provenance, coverage, and freshness
**And** keyboard, touch, screen reader, forced colors, zoom/reflow, and reduced motion provide equivalent access.

**Given** the relevance explanation is opened
**When** Findur explains why the person appears
**Then** it may name approximate proximity, selected compatibility mode, one or two derived traits, disclosure, and freshness
**And** never reveals a score, exact formula, deterministic claim, cohort, wealth judgment, exact location, or account identifier.

### Story 3.4: Make One Durable Pass or Interested Decision

As a Discovery user,
I want to Pass or express Interested using my preferred input method,
So that one deliberate decision advances the deck reliably without duplicate effects.

**Acceptance Criteria:**

**Given** a current candidate has no prior decision from the viewer
**When** the user activates Pass or Interested by explicit control, keyboard, touch, or swipe gesture
**Then** every input method invokes the same idempotent decision command
**And** explicit controls remain at least 48px with visible focus and readable pending/disabled states
**And** gesture direction or motion is never the only way to decide.

**Given** a decision command is pending
**When** duplicate clicks, gesture completion, retry, or network replay occurs
**Then** at most one directed viewer-to-generated swipe is persisted
**And** conflicting second decisions are rejected without rewriting history
**And** controls cannot emit parallel mutations.

**Given** the viewer chooses Pass
**When** the transaction commits
**Then** the directed decision and anchor advancement settle atomically
**And** no match is created
**And** the next eligible card or honest empty state becomes current.

**Given** the viewer chooses Interested
**When** the generated candidate's deterministic viewer-specific reciprocal decision is evaluated
**Then** the swipe, reciprocal decision, optional match, and anchor advancement settle in one transaction using the common lock order
**And** a positive reciprocal outcome creates exactly one Mutual Match
**And** a non-reciprocal outcome advances without falsely implying a match.

**Given** a decision transaction fails or the anchored candidate becomes ineligible before commit
**When** the UI receives the result
**Then** it retains or safely replaces the current decision surface according to authoritative state
**And** provides a focused recovery action
**And** never shows an uncommitted match or loses a committed decision.

**Given** decision concurrency and isolation tests run
**When** duplicate, reordered, cross-user, stale-anchor, rollback, panic, and restart cases execute
**Then** one durable outcome remains
**And** one viewer cannot observe or mutate another viewer's swipes or matches.

### Story 3.5: Respond to an In-Session Incoming Interest

As a Discovery user,
I want to evaluate a seeded synthetic candidate who already expressed interest,
So that the demonstration can show disclosure asymmetry and reciprocal choice without notifications or an inbox.

**Acceptance Criteria:**

**Given** the deterministic seeded scenario becomes available during an authenticated Discovery session
**When** Incoming Interest is offered
**Then** it appears only as a Discovery child surface
**And** no fourth navigation area, inbox, feed, history, Web Push subscription, app badge, notification payload, or background-delivery system is created.

**Given** the synthetic initiator has a higher Disclosure Level than the viewer
**When** the incoming card or shared Candidate Detail renders
**Then** the initiator is projected at the initiator's saved level
**And** the viewer's own lower saved level and candidate-visible information remain unchanged
**And** the asymmetry is explained explicitly without pressuring an upgrade.

**Given** Incoming Interest is displayed
**When** identity and evidence render
**Then** the same synthetic provenance, freshness, coverage, Photo Obscure, visualization, accessibility, and server-side disclosure rules as ordinary Candidate Card/Detail apply.

**Given** the viewer chooses Pass
**When** the idempotent decision commits
**Then** the incoming scenario is settled once and the exact safe Discovery state is restored
**And** no match is created.

**Given** the viewer chooses Interested
**When** the idempotent settlement commits
**Then** the preconfigured reciprocal interest and viewer decision create exactly one Mutual Match atomically
**And** the viewer's saved disclosure remains unchanged
**And** duplicate submission cannot create another swipe or match.

**Given** the incoming scenario is unavailable, invalidated, already settled, or no longer eligible
**When** its route is opened
**Then** a focused recovery state returns to safe Discovery
**And** no candidate detail leaks through cached, history-restored, or prefetched state.

### Story 3.6: Reveal a Mutual Match Once

As a user who reaches a Mutual Match,
I want the candidate's placeholder photo to be revealed clearly,
So that the demonstration completes the progressive identity-reveal promise without adding chat or another consent step.

**Acceptance Criteria:**

**Given** an Interested decision transaction creates a Mutual Match
**When** its result is rendered
**Then** the Photo Obscure is replaced immediately by the same-size allowlisted bundled placeholder avatar
**And** no second acceptance decision is requested
**And** layout does not jump.

**Given** the reveal transition occurs
**When** assistive technology and visual presentation announce it
**Then** the matched state is announced once with concise accessible identity naming
**And** synthetic provenance, disclosure, and relevant freshness remain present
**And** reduced-motion mode reaches the identical complete final state without animation.

**Given** the match response is serialized
**When** the post-match projection is inspected
**Then** it uses the dedicated post-match schema
**And** reveals only the placeholder avatar and fields permitted by the candidate's Disclosure Level
**And** does not expose hidden financial fields, account identifiers, exact location, or any real viewer data.

**Given** the Mutual Match result is visible
**When** forward actions are presented
**Then** Continue Discovery is the sole forward action
**And** chat, messaging, contact exchange, unmatch/block/report stubs, match history, notification controls, and another disclosure prompt are absent from this MVP.

**Given** Continue Discovery is activated
**When** the next state is calculated
**Then** the settled candidate is not offered again
**And** the next eligible anchored candidate or honest recovery/empty state renders
**And** the match is not re-announced as a new event.

### Story 3.7: Recover Safely from Discovery Changes and Failures

As a Discovery user,
I want clear recovery when eligibility or trustworthy data changes,
So that Findur never substitutes unsafe candidates or leaves me acting on invalid information.

**Acceptance Criteria:**

**Given** Discovery is blocked by incomplete profile, no Included Accounts, no saved disclosure, unusable signals/freshness, reauthorization, disabled connection, provider failure, offline state, or no eligible candidates
**When** the route evaluates its gate
**Then** a Recovery Panel names the condition, what remains trustworthy, and the exact safe action/destination
**And** heading focus and status announcement occur deterministically
**And** no protected candidate content flashes first.

**Given** safe revalidation is in progress and the current anchored view remains permitted and trustworthy
**When** refreshed data is not materially different
**Then** the current card/detail and focus remain stable with a visible timestamp and progress state
**And** provider work does not block the interface without explanation.

**Given** safe revalidation produces a material but still-permitted change
**When** it would rearrange the current task
**Then** Findur announces that an update is available before replacing or re-anchoring content
**And** the user's committed swipe state is preserved.

**Given** inclusion removal, disclosure downgrade, disconnect, session failure, safety invalidation, or loss of candidate eligibility occurs
**When** the active or historical view is reevaluated
**Then** affected active, prefetched, rendered, cached, offline, Back-restored, and revalidation content is suppressed immediately
**And** focus moves to the appropriate safe replacement
**And** stale finalizers cannot restore the broader state.

**Given** an anchored candidate changes immaterially while remaining eligible
**When** the viewer is actively reading or deciding
**Then** the pinned signal/disclosure projection remains stable until acknowledgement or expiry
**And** permission decreases and lost eligibility always override that stability.

**Given** an error, sparse state, or connectivity failure occurs
**When** recovery options are offered
**Then** Findur never silently changes maximum distance, compatibility mode, Disclosure Level, account selection, freshness policy, or candidate eligibility
**And** no live owner data, raw diagnostics, secrets, exact coordinates, or hidden cohort information is exposed.
