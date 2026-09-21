# Epic 1 Context: Explore and Control Live Portfolio Data

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Give each authenticated viewer explicit control over which connected SnapTrade accounts Findur may use, then provide a private, honest view of the supported account, balance, position, and recent-activity data. The epic establishes dataset-specific coverage and freshness, bounded activity-driven refresh, traceability from source facts to later product use, recoverable provider states, and a disconnect path that stops use and destroys the complete user aggregate even when provider revocation cannot be confirmed.

## Stories

- Story 1.1: Select and Confirm Included Accounts
- Story 1.2: Inspect the Private Portfolio Showcase
- Story 1.3: Refresh Portfolio Data Without Provider Call Storms
- Story 1.4: Trace Source Data to Product Use
- Story 1.5: Disconnect SnapTrade and Leave Findur

## Requirements & Constraints

- OAuth Connection, Account Inclusion, private derivation, and visible Disclosure Level are distinct permissions. No account is included by default; before confirmation, retain only the minimum masked inventory needed for selection. Retrieval and use of financial data begin only for explicitly confirmed eligible accounts.
- The private Portfolio Showcase must present connections/accounts, balances or values, positions or holdings, and a bounded recent-activities window as separate datasets. Show source, included-account coverage, currency, observation/retrieval and publication context, and dataset-specific freshness. Empty, unavailable, unsupported, syncing, stale, failed, disabled, and reauthorization-required states must remain explicit; missing values are never treated as zero or fabricated.
- Financial access is read-only and purpose-limited. Do not retrieve or infer trading, orders, tax lots, quotes, reference enrichment, provider performance, or implicit currency conversion. Preserve decimal precision provenance and keep different currencies separate.
- Owner-private responses may serialize only the authenticated Actor's committed Included Accounts and must use `Cache-Control: private, no-store`. Live portfolios remain isolated from other real viewers and from generated candidates. Raw provider payloads, credentials, tokens, financial values, and unnecessary personal data must never be persisted in logs or diagnostics.
- Freshness policy is per account and dataset: holdings are current through 15 minutes in real-time mode or 36 hours in Daily mode, stale-usable through 72 hours; activities are current through two calendar days and stale-usable through seven. Missing initial data, revoked authorization, or a disabled connection is unusable.
- Provider failure must preserve the prior trustworthy authorized snapshot when still usable and expose a safe recovery state. Permission decreases, account removal, and disconnect override stale preservation immediately and fail closed.
- Disconnect immediately blocks new work, revokes all Findur sessions, attempts bounded provider revocation, and deletes the full user aggregate from application-controlled active storage, including identity mapping, authorization, profile/configuration, portfolio data, derived data, interactions, anchors, and caches. Local destruction completes even if provider revocation fails; later authorization creates a new user.
- Core flows must meet WCAG 2.2 AA, support keyboard, screen reader, touch, reduced motion, EN/FR, System/Light/Dark, phone/desktop, 200% reflow, and 400% single-dimension reflow except bounded wide-data regions with an equivalent summary.

## Technical Decisions

- Keep SnapTrade behind the server-side `PortfolioProvider` adapter. Browser code calls only the authored OpenAPI 3.1 `/api` contract; generated Go and TypeScript artifacts are not hand-edited. HTTP adapters delegate to application use cases, and protected responses expose purpose-specific owner-private schemas.
- Normalize transient provider responses into immutable typed dataset versions with scoped atomic heads: portfolio-scoped connections/accounts and per-account balances, positions, and activities. Publication moves a head only with a complete normalized row set; fetch or normalization failure leaves the prior head unchanged. Retain noncurrent versions only for authorized anchors or a short cleanup grace.
- Account Inclusion is an idempotent prepare/call/finalize bulk operation. Removals commit and purge immediately; additions remain pending until all required datasets publish successfully. Mixed-change failure never restores removals or broadens the prior committed set. Provider calls run outside database transactions.
- Fence every out-of-transaction provider result with authorization and portfolio lifecycle generation, inclusion version, and required account membership. Removal or disconnect advances the generation before deletion so stale finalizers cannot recreate data, credentials, signals, or anchors.
- Refresh is triggered only by authenticated activity that needs freshness—not rendering, probes, smoke tests, synthetic traffic, or heartbeats. Serve persisted snapshots without refetching while current, collapse concurrent demand onto one guarded operation, honor provider rate-limit headers, use bounded retry and circuit breaking, and never invoke billable manual refresh automatically.
- Rotating-token refresh and disconnect use mutually exclusive generation-guarded leases. Ambiguous refresh fails closed to reauthorization; a safe provider `401` permits one refresh and one request retry. Distinguish brokerage-connection repair from OAuth reauthorization.
- PostgreSQL is the state and coordination layer. Product modules own their tables and coordinate cross-module lifecycle operations through explicit transaction-scoped ports and the shared lock order; network calls never occur inside transactions.

## UX & Interaction Patterns

- Account Inclusion is the post-OAuth portfolio choice for the user's persistent dating profile. It uses one friendly checkbox list, omits closed accounts, shows temporary unavailability only when actionable, provides an accessible tri-state Select All, and distinguishes selected from saved counts in ordinary language. Its compact review contains changed names, resulting count, “change anytime” reassurance, and neutral Back/Save actions—no destructive styling or technical consequence wall. A committed first save announces success and opens the Portfolio Showcase; later edits open from and return to that Showcase. Saved choices persist across ordinary login and reauthorization.
- Portfolio sections pair a text-labelled Freshness Indicator and timestamps with visual summaries and accessible table/equivalent views. Meaning cannot depend on color, hover, motion, node size, or layout; wide tables stay within labelled overflow regions and have a non-table summary.
- Stale-while-revalidate keeps the last permitted trustworthy view and the current task stable. Background changes are announced; material changes are acknowledged before rearrangement. With no trustworthy view, use a layout-matched skeleton with an external loading label. Unusable states transition focus to a Recovery Panel with one clear action.
- Source-to-Experience Trace separates owner-only exact views, private matching inputs, and candidate-visible disclosure. Provider, normalized, derived, and projected stages carry provenance, coverage, freshness, and transformation context; blocked or unsupported paths use explicit text and a strong dashed treatment.
- Disconnect uses a single confirmation dialog with a neutral cancel, danger-styled confirmation, a complete consequence summary, and a clear distinction between guaranteed local deletion, best-effort provider revocation, and backup policy. Clear protected client state before showing a safe public leaving state.

## Cross-Story Dependencies

Story 1.1 establishes the committed inclusion set and initial snapshots consumed by the Showcase, refresh, and trace stories. Its first-time path is `OAuth success → Choose accounts → Portfolio Showcase`; it is not the permanent Portfolio view. Story 1.2 establishes the normalized dataset and freshness presentation used by Stories 1.3 and 1.4, provides **Edit included accounts**, and returns there after a successful edit. Story 1.3 maintains trustworthy versions without weakening inclusion. Story 1.5 supersedes every other operation: its lifecycle fence and deletion boundary must defeat concurrent inclusion, refresh, normalization, derivation, and publication work. Epic 1 depends on Epic 0's authenticated session, masked account inventory, and proven provider adapter; Epic 2 depends on Epic 1's committed snapshots and freshness model.
