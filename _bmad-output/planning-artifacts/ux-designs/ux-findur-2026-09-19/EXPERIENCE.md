---
name: Findur
status: draft
sources:
  - ../../briefs/brief-findur-2026-09-19/brief.md
  - ../../briefs/brief-findur-2026-09-19/addendum.md
  - ../../prds/prd-findur-2026-09-19/prd.md
  - ../../prds/prd-findur-2026-09-19/addendum.md
  - ../../prds/prd-findur-2026-09-19/research-market-landscape.md
  - ../../research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md
updated: 2026-09-19
---

# Findur — Experience Spine

## Foundation

Responsive, installable mobile web with desktop parity. Phone is the primary form factor; the same web product supports desktop and standalone display when saved to a phone home screen. No implementation framework or UI system is selected. `DESIGN.md` is the visual identity reference; this spine owns behavior. English/French and light/dark are first-cut requirements.

The protected, preconfigured owner is the only live-data user. Discovery candidates are synthetic. For a ready owner, Discovery is the default authenticated destination. Incomplete Personal Profile, no included account, no Usable Portfolio, no explicitly saved Disclosure Level, or an unusable Freshness State produces a focused gate with a direct recovery path. During safe syncing or revalidation, the last trustworthy Discovery view remains available with freshness context.

Product scope and definitions are inherited from the frontmatter sources rather than repeated here. The final PRD plus its explicit reconciliations govern every conflict with earlier briefs, addenda, or research; research is advisory where the final PRD is silent.

The installed experience uses a persistent app shell, device safe areas, branded launch presentation, standalone navigation, and theme-aware browser chrome. A cold start, refresh, home-screen launch, or deep link restores the requested safe route after session and gate checks; it never flashes protected content. Exact manifest, browser metadata, platform asset matrix, snapshot protection, and cache mechanisms belong to architecture/build.

## Information Architecture

### Public shell

| Surface | Route role | Purpose |
|---|---|---|
| Public Site | Public default | Brand, portfolio-first premise, progressive identity reveal, 18+ and demonstration boundary, entry action |
| About | Stable public route | Product purpose and current scope |
| Trust and safety | Stable public route | Adult-only, read-only/non-advisory, anti-solicitation, money-request, targeting, and concern guidance |
| Privacy | Stable public route | Connected-data and live-owner/synthetic boundary; draft status until reviewed |
| Terms | Stable public route | Demonstration terms; draft status until reviewed |
| Contact or support | Stable public route | Privacy, safety, and product concern path without implying production reporting operations |
| Protected owner entry | Entry route | Authenticate the preconfigured owner; preserve a permitted intended destination |
| Guest Demo | Deferred extension seam | Not advertised or implemented in the first cut. Preserve an isolated, ephemeral, synthetic-only branch contract for later activation without redesigning Discovery. |

`Public Header` and `Public Footer` keep About, Trust and safety, Privacy, Terms, and Contact or support reachable without authentication. Intentional sharing metadata describes the public proposition only. Candidate, Portfolio, Profile, callback, error, and authenticated URLs use neutral non-sensitive titles/descriptions and must not generate social previews containing financial or identity content.

### Authenticated shell

| Primary area / route | Purpose | Entry and child routes |
|---|---|---|
| Discovery | Default for a ready owner; owns the Swipe Deck, preferences access, safe deck recovery, Incoming Interest, and Mutual Match outcome | Candidate Detail is a dedicated progressive, deep-linkable child route; Incoming Interest is a seeded single-interest child surface; Mutual Match is transient and returns to Discovery |
| Portfolio (contains the Portfolio Showcase) | Private Connected Portfolio, accounts, balances/values, holdings/positions, activities/orders, source coverage, freshness, connection controls, account inclusion, and source-to-experience explanation | Connect consent, OAuth return, account selection/settings, freshness details, reauthorization, disconnect |
| Profile | Personal Profile, Disclosure Level, complete Profile Preview, notification preferences, language/theme, and account settings | Edit profile, preferences/disclosure/notifications, phone/desktop and pre-/post-match preview |

`App Navigation` contains exactly Discovery, Portfolio, and Profile. Mobile uses a bottom app-shell treatment; desktop uses a left rail. Public/legal surfaces remain outside the authenticated shell. Guest Demo implementation and dedicated mocks are deferred from the first cut; its isolation contract remains an architecture extension seam. Chat and a matches-history surface are excluded.

### Routing and gates

1. Resolve route and session without rendering protected contents.
2. If unauthenticated and the route is protected, send the visitor to protected owner entry and retain only a validated internal intended destination.
3. After authentication, route to the intended destination when permitted; otherwise use the most specific gate: required Personal Profile, portfolio connection, account inclusion, or reauthorization/recovery.
4. OAuth return first shows an explicit connection result. Success continues to required account selection; it does not silently drop into Discovery.
5. A ready owner entering the shell without an intended destination lands in Discovery.
6. Candidate detail opens on its stable route. Mobile is full-screen; desktop is two-column. Back restores the exact Swipe Deck position, scroll, current candidate, prior decisions, and uncommitted view state that remains permitted.
7. Any permission-decreasing transition follows the permission-change consequence matrix and takes precedence over history restoration.
8. Incoming Interest is reached only from a seeded Discovery scenario. It presents the initiator at the initiator's Disclosure Level while leaving the owner's lower Disclosure Level unchanged; Pass returns to Discovery and Interested uses the existing Mutual Match result.

Every full route transition updates the document title and moves focus to the route/result heading. OAuth return focuses the outcome heading; a conditional gate that replaces content in place focuses its gate heading. Same-route cancel/retry returns focus to its invoker and outcomes are announced once, not duplicated through a live region.

## Voice and Tone

Brand posture lives in `DESIGN.md`. Microcopy is direct, calm, non-judgmental, and explicit about uncertainty.

| Use | Avoid |
|---|---|
| “Portfolio-informed compatibility” | “Financial score” or deterministic compatibility |
| “Using 2 of 4 connected accounts” | “All your finances” or wording that implies completeness |
| “Last refreshed 18 Sep, 21:40” | “Live” or “real time” |
| “Photo hidden until Mutual Match” | Playful concealment that makes identity feel like a prize |
| “Why this person appears” with bounded factors | Cohort names, rank, formula, or false-precision percentage |
| “This information came from accounts you selected.” | “Verified wealthy,” “financially healthy,” or identity certification |
| “We’re checking for updates. You can keep reviewing the last trustworthy view.” | Blocking spinner with no explanation |
| “No eligible Synthetic Candidates fit these settings.” | Silent widening of distance or disclosure |

Errors name what happened, what remains safe/trustworthy, and the next action. French copy is authored and reviewed as French—not mechanically shortened to fit an English layout. “Snapshot,” “Holdings,” and “Full Detail” remain the source-defined Disclosure Level names until an explicit product change supplies approved French product labels; surrounding explanatory copy is localized.

## Component Patterns

Visual specifications live in `DESIGN.md.Components`. These names are the canonical inventory.

| Component | Use | Behavioral rules |
|---|---|---|
| `App Navigation` | Authenticated shell | Exactly Discovery, Portfolio, Profile. Announces current area; preserves area-local state when safe. Hidden behind neither gestures nor hover. |
| `Swipe Deck` | Discovery | Owns one active Candidate Card, order/position, decision idempotency, exact safe restoration, next-card preparation, and empty/recovery substitution. It never exposes Invisible Cohorts and never silently relaxes eligibility. |
| `Public Header` | Public shell | Exposes public routes, entry, language, and theme. A compact menu preserves all destinations, keyboard order, and dismissal behavior. |
| `Public Footer` | Public shell | Repeats stable About, Trust and safety, Privacy, Terms, and Contact or support routes plus demonstration status. Links wrap in EN/FR and remain available without authentication. |
| `Candidate Card` | Discovery and preview | One active card. Opens Candidate Detail without committing a decision. Shows only candidate-permitted data and retains Synthetic Data Label, freshness, and explicit actions. |
| `Candidate Detail` | Progressive portfolio depth | Dedicated route. Sections appear only when Disclosure Level permits. Persistent Pass/Interested actions; Back performs exact safe deck restoration. |
| `Portfolio Node Visualization` | Candidate Card, Candidate Detail, Portfolio, Profile Preview | Has programmatic title/summary and an adjacent Chart Data Table or equivalent values. Selection/focus synchronizes node and corresponding text; no essential hover-only content. |
| `Chart Data Table` | Accessible and exact data alternative | Caption names coverage, period, source, and freshness. Supports keyboard scrolling where needed; summary stays outside overflow. Hidden columns never reveal disallowed detail. |
| `Photo Obscure` | Pre-match identity | Announces “Photo hidden until Mutual Match”; does not expose image alt text, URL, or cached preview. Mutual Match replaces it with the real image named concisely from the display name, for example “Photo of Alex”; alt text never infers appearance traits. |
| `Synthetic Data Label` | Every synthetic Candidate Card, Candidate Detail, preview state, and Mutual Match | Persistent in visual and accessible identity; never removed by expansion, responsive reflow, or navigation. |
| `Freshness Indicator` | Every material time-sensitive claim | Announces state and last trustworthy timestamp. Opens freshness details. Never claims a webhook proves holdings changed. |
| `Account Inclusion Control` | Post-OAuth and Portfolio settings | A labelled checkbox group. Individual usable accounts are checkboxes; Select all exposes checked/mixed/unchecked. Unavailable accounts are disabled with an associated reason. Draft coverage and committed “Using N of M connected accounts” are distinct. One scoped Apply/Confirm summarizes additions/removals, categories, private purpose, and coverage before retrieval/derivation broadens. Saving, recalculating, success, and failure are announced once without moving focus. |
| `Disclosure Level Control` | Profile settings and Profile Preview | One labelled native radio group or equivalent single-select semantics for Snapshot, Holdings, Full Detail. Initially no level is selected or saved; Snapshot is marked “Recommended starting point” but is not preselected. Selecting a tier changes only the inspected preview. A persistent summary distinguishes saved level from “Previewing — not saved,” and explicit Save disclosure level/Cancel actions commit or discard the draft. Route exit with a changed draft asks Discard changes or Stay. Each option associates a concise hidden/bucketed/derived/exact description. A concise screenshot/memory/inference limit appears before save, strongest beside Full Detail. |
| `Preference Control` | Discovery preferences and Profile | Controls maximum distance and similar/diversified/complementary preference. Changes rebuild eligible ordering; never exposes ranking or Invisible Cohorts. |
| `Profile Field` | Personal Profile | Labels required/optional and private/pre-match/post-match visibility. Error is adjacent, associated, and retained until corrected; save preserves entered values. |
| `Primary Action` | One main action per region | Executes on click/tap/Enter/Space once; pending state prevents duplicate submission without erasing label. |
| `Secondary Action` | Alternate or navigational action | Never visually competes with the main action; retains equivalent keyboard and touch operation. |
| `Swipe Actions` | Discovery and Candidate Detail | Pass and Interested are always explicit alternatives to swipe and keyboard arrows. One decision per candidate; duplicate input is ignored and announced. |
| `Revalidation Notice` | Safe stale-while-revalidate behavior | Keeps the last trustworthy view usable, states visible freshness, runs checking in the background, and reports material changes without moving focus or replacing the active task. |
| `Recovery Panel` | Conditional gate or failed/empty state | Names the condition and links to the exact recovery surface. No recovery action silently changes disclosure, distance, or selected accounts. |
| `Mutual Match Reveal` | Reciprocal right-swipe result | Reveals Photo immediately as the consequence of the original decision, announces Mutual Match once, retains synthetic provenance, and offers Continue Discovery. No chat or second appearance gate. |
| `Notification Control` | Contextual opt-in and Profile settings | After the first settled Interested decision, offers a value-specific pre-prompt: “Want to know if it becomes mutual?” Enable is the direct user gesture that may invoke browser permission; Not now continues without loss. Settings separately control Mutual Match, Incoming Interest, and connection-action-required categories and expose unsupported/denied/install-required states without nagging. |
| `Discovery Activity Notice` | Discovery child surface | Persists each unacknowledged Incoming Interest or Mutual Match until opened or dismissed in-app. Push and badge are optional delivery signals; this in-app event is authoritative. Opening performs authentication, eligibility, permission, and safety checks before showing the permitted reveal. It does not become a fourth navigation area, chat affordance, or permanent history. |
| `Profile Preview Frame` | Owner-only complete preview | Switches Disclosure Level, pre-/post-match state, and representative phone/desktop composition without changing saved settings. Labels live owner/private versus synthetic states. |
| `Source-to-Experience Trace` | Portfolio explanation | Included source data branches to owner-only exact view, private matching input, and candidate-visible disclosure. Every branch carries live-owner/synthetic provenance; Disclosure Level appears only on candidate-visible disclosure. Unknown, blocked, and unsupported remain explicit. |
| `Consent Panel` | Before OAuth | Presents provider, named access/categories, retrieval of minimum masked account metadata for selection, private-use/visible-disclosure distinction, screenshot/inference limits, and disconnect. Continue authorizes provider access and metadata retrieval only; it does not authorize portfolio retrieval, derivation, or Discovery use. |
| `Confirmation Dialog` | Disconnect and consequential downgrade/exclusion | One modal only. States affected source, derived, cache, preview, and Discovery outputs; cancel is always available; returns focus to invoker. |
| `Language and Theme Control` | Public and Profile settings | Language: English/French. Theme: System/Light/Dark. Applies immediately, persists, and preserves route/task/focus where feasible. |
| `Safety Notice` | Public Site, consent, and Discovery context | Presents 18+, read-only/non-advisory, anti-solicitation, money-request, and financial-targeting guidance where relevant; links to Trust and safety. |
| `Loading Skeleton` | Cold route/data load only | Matches expected layout, has an external loading label, does not imitate real values, and yields without focus jump. Revalidation uses the last trustworthy view instead. |

## State Patterns

| Surface | Required states and treatment |
|---|---|
| Public Site and public information | Cold load uses stable text-first structure; missing route offers public navigation; network failure keeps already loaded public content where safe. Draft legal status remains visible. No authenticated state appears in previews or metadata. |
| Guest Demo (deferred extension) | When later activated: entry, cold load, ready synthetic trial, reset/expiry, recoverable generation failure, owner-route/isolation denial, and return to Public Site. Reset removes ephemeral guest state; failure never redirects to owner entry/OAuth or exposes owner data. In the first cut no guest route or action is advertised. |
| Protected owner entry/session | Signed out, authenticating, invalid/expired session, authenticated. Preserve validated internal intent; reject external/open redirects. Session expiry keeps destination only, clears sensitive rendered state, and returns after successful authentication. |
| Connect consent and OAuth return | Prerequisite not met, ready, provider handoff, success, denial, invalid callback, expired flow, provider error. Every non-success state has safe retry/back; callback details and tokens never render. |
| Account selection | Initial none-selected, usable/unavailable accounts, draft unchecked/mixed/checked, committed coverage, applying, exclusion purge, recalculating, success, addition failure, exclusion retry. Discovery gate stays closed at zero. Confirmed exclusions suspend broader use immediately; failed additions retain the prior narrower committed set. |
| Portfolio | Cold load, loaded, section empty, unavailable, unsupported, syncing, stale, failed, needs reauthorization, disconnected, offline. Each dataset section states its own condition; missing is never zero. Last trustworthy values remain only while architecture classifies them usable. |
| Discovery | Cold load, ready, safe revalidation, material update, sparse/no eligible candidates, profile incomplete, no included accounts, no saved Disclosure Level, syncing without trustworthy view, unusable/stale, needs reauthorization, failed, disconnected, offline. Recovery never inserts unqualified candidates or relaxes consent. |
| Candidate detail | Loading, ready by Disclosure Level, revalidating, permission/detail invalidated, candidate unavailable, offline with permitted trustworthy state. Invalidation follows the consequence matrix, replaces sensitive detail with a Recovery Panel, and provides safe Back. |
| Incoming Interest | Seeded available, ready at initiator's Disclosure Level, Pass, Interested, Mutual Match, invalidated/unavailable, duplicate decision ignored. Owner Disclosure Level never changes; Back/Pass returns to the exact safe Discovery state. |
| Notifications and unread activity | Not asked, contextual pre-prompt, browser prompt, granted, denied, unsupported, installation required, category enabled/disabled, push delivered/dismissed/opened, badge set/cleared, pending in-app event, invalidated event. Denial never blocks the product or triggers repeated prompting. Push loss never loses the authoritative in-app event. |
| Profile and preferences | Empty/incomplete, editing, disclosure none-saved, previewing-unsaved, discard/stay confirmation, saving, saved, save failure. Preserve entries and identify each gate. `[ASSUMPTION]` Minimum fields follow PRD A-4: display name, adult-age confirmation, coarse location, discovery eligibility/preferences, relationship intent, at least one Photo, and short biography or prompt response. |
| Profile Preview | Loading, ready, no saved Disclosure Level, previewing Snapshot/Holdings/Full Detail without mutation, live-owner private, synthetic comparison, invalidated disclosure/account data, unsupported dataset. Preview controls never mutate the saved setting until explicit Save disclosure level succeeds. |
| Mutual Match | Reciprocal one-time reveal, non-reciprocal continuation, duplicate event ignored, creation failure with retry-safe state. No history/list is implied after leaving. |
| Installed/standalone launch | First launch, returning session, expired session, deep-linked launch, offline launch. Show opaque branded launch field, then public or gated shell; never show a stale task-switcher or launch snapshot of financial data. Unsafe/private data is not available from offline cache. |

Focus states follow `DESIGN.md` focus tokens. Errors remain next to the affected control and are summarized at the top of a submitted form with links to fields. Success messages are concise and do not steal focus. A permission/freshness/safety invalidation that removes focused content moves focus deterministically to the replacement Recovery Panel or status heading, announces why the content disappeared, and exposes the next safe action; when a route closes, focus moves to the destination heading rather than a removed card. Offline behavior is honest: public/static shell content may remain available; authenticated portfolio and candidate data is shown only if the last trustworthy view is explicitly safe for that state, otherwise it is suppressed.

## Interaction Primitives

- Tap/click activates labeled controls. Touch targets are at least 44×44 CSS px; primary mobile actions target 48px height.
- Pointer hover is a quiet desktop enhancement, never a prerequisite: buttons and selectable surfaces use stable tonal/inset emphasis, links underline, and portfolio nodes repeat the same detail available by focus and tap. Hover and pressed feedback never translates, scales, changes border width, or causes layout shift; disabled controls do not react.
- `DESIGN.md` → Interaction State Contract is normative across the product. Shared component primitives own navigation, action, link, selectable-row, and portfolio-node states; route-level code may compose those primitives but must not redefine their interaction behavior. Passive data surfaces do not receive pointer, pressed, or keyboard-focus affordances.
- Horizontal swipe on the active Candidate Card maps left to Pass and right to Interested only after a clear threshold; a canceled gesture snaps back. Explicit Swipe Actions always remain visible.
- Left/Right arrows operate the focused Candidate Card only when they do not conflict with a focused control, text field, table, or screen-reader browsing mode. Enter opens Candidate Detail; Escape closes the topmost dialog/menu or returns from detail when safe.
- Tab order follows reading order. Roving focus may be used within radio-like option groups; the selected value and group label are announced.
- Browser and in-app Back are equivalent. They do not replay OAuth, duplicate a swipe, restore invalid detail, or escape the authenticated/public boundary incorrectly.
- Pull-to-refresh is not required. If exposed by the browser, it must not duplicate actions or erase deck state.
- No drag-only control, hover-only disclosure, stacked modal, autoplay carousel, infinite hidden navigation, or long-press-only action.

## Accessibility Floor

WCAG 2.2 AA is the minimum for the complete English/French, light/dark, phone/desktop, and installed-web surface.

- Landmarks, headings, lists, tables, fields, status messages, dialogs, and navigation use native semantics first. Candidate identity, synthetic provenance, Photo state, Disclosure Level, freshness, and account coverage are programmatically named.
- All core actions are keyboard operable. Focus is visible, never clipped or obscured by sticky regions. Every full route, OAuth outcome, or direct-entry result focuses its updated heading; every in-place gate/invalidation focuses its replacement heading. Opening Candidate Detail moves focus to its heading; safe Back returns focus to the originating Candidate Card; confirmation cancel returns focus to its invoker.
- Screen readers receive concise route titles and changes to save, revalidation, invalidation, swipe, and Mutual Match states. Polite announcements cover background changes; errors and match creation are assertive only when immediate attention is required.
- Portfolio Node Visualization has a text summary and Chart Data Table/equivalent. Chart relationships, labels, periods, units, source coverage, and freshness do not depend on position, size, color, animation, or pointer hover.
- Photo Obscure conveys the state without exposing the hidden image description. After Mutual Match it is replaced by an image with concise alt text based on the display name, for example “Photo of Alex,” or is decorative only when adjacent identity text is explicitly declared equivalent; appearance traits are never inferred. Swipe Actions provide non-gesture parity. Reduced motion displays final node/deck/reveal states immediately and preserves comprehension and focus.
- Account inclusion is a named checkbox group with a tri-state Select all, unavailable reasons, and separate draft/committed coverage; save and recalculation announcements are concise and do not move focus. Disclosure Level is a named single-select group whose inspected preview and saved value are announced distinctly.
- At 200% zoom, controls and content reflow without loss. At 400% zoom where WCAG reflow applies, use a single column; wide data tables may scroll in their own labeled region and retain a readable summary.
- Errors use text plus icon/structure, identify the field or state, preserve input, and provide recovery. Disabled controls explain the prerequisite when the reason is not apparent.
- EN/FR labels, accessible names, instructions, errors, date/number/currency output, and chart summaries are complete and language-correct. The document language changes with the selected locale; mixed fixed terms are marked appropriately where needed.
- Both themes meet contrast targets defined in `DESIGN.md`; operating-system forced colors and high-contrast behavior retain borders, labels, selection, and focus.

## Localization & Theme

English and French are first-cut peers. A first visit selects language from browser preference when supported, otherwise English; explicit choice persists for the current owner/browser and overrides later browser changes. Language switching keeps the current permitted route and task, updates document language, copy, accessible names, and locale formatting, and does not duplicate stored content records.

System is the initial theme mode and follows the operating-system preference until the person chooses Light or Dark. Explicit theme choice persists. Browser/standalone chrome tracks `{colors.canvas}` or `{colors.canvas-dark}` without exposing state through title/icon changes. Portfolio visualization semantics are token-driven and retain label/pattern/table equivalence across themes.

Use locale-aware dates, times, decimal/group separators, percentages, distances, and currencies. Never concatenate translated fragments around values. Layouts tolerate French expansion and zoom; truncation cannot hide an account, Disclosure Level, action consequence, Freshness State, or error.

## Deep Links & Navigation Recovery

- Stable public links open directly without session checks.
- Protected links resolve session first and never render protected content underneath a gate. A valid internal intended route survives sign-in, OAuth handoff, refresh, browser restart, and installed launch; query strings cannot carry sensitive values.
- Candidate Detail is deep-linkable within the protected demonstration. If the referenced synthetic candidate remains eligible and permitted, open it and seed a safe deck return context. If not, explain that it is unavailable and return to Discovery; do not substitute another candidate.
- OAuth callback URLs are single-use transition points, not shareable content. Refresh/replay resolves idempotently to the recorded result or a safe recovery state.
- Back from Candidate Detail restores exact deck position/state. If disclosure, account inclusion, connection, or freshness invalidated that state, restore only the permitted shell and explain the change.
- Public sharing metadata is intentional and generic. Authenticated and financial routes exclude sensitive social-preview data, sensitive URL paths/parameters, and indexable/private snapshots.

## Data Consent & Account Inclusion

OAuth access, Findur-level account inclusion, and Disclosure Level are three separate layers and must never be collapsed in copy or control hierarchy.

Consent is staged:

1. **Pre-OAuth Continue** authorizes access through the named provider for the stated categories and retrieval of only the minimum masked account metadata needed for selection. No financial payload retrieval, portfolio derivation, ranking, preview, or Discovery use begins at this stage.
2. **Post-OAuth Apply/Confirm** authorizes Findur use only for the selected accounts and then begins permitted financial source retrieval and derivation. Connected never means included.

After OAuth, `Account Inclusion Control` is a labelled checkbox group with **none selected by default**. Each usable account has a checkbox; unavailable accounts remain visible but unselectable with a reason. **Select all** reports checked, mixed, or unchecked across usable accounts. Discovery requires **at least one** committed included account. Draft coverage is labelled separately from committed coverage such as **“Using 2 of 4 connected accounts.”**

Edits remain a draft until one scoped Apply/Confirm. The summary names newly included and excluded accounts, data categories, private purpose, committed coverage change, and the difference between provider access and Findur inclusion. Additions begin retrieval/derivation only after confirmation. Inclusion remains editable in Portfolio settings.

Findur may retain a minimal masked connection inventory so connected accounts can be shown and reselected: provider name, masked account label, provider connection status, selection eligibility/reason, and an opaque account reference. This inventory contains no balance/value, holding/position, activity/order/transaction, performance, or Portfolio-Derived Signal. Disconnect removes it under the PRD local deletion boundary; only permitted categorical security/audit events without financial values may remain.

Architecture must verify whether SnapTrade supports provider-granular account authorization. Findur-level inclusion remains required regardless and must not claim provider revocation it cannot perform.

### Permission-change consequence matrix

This matrix is the single contract for purge, suppression, retry, cache/history/offline invalidation, and failure behavior. Other sections and flows reference it.

| Confirmed trigger | Immediate effect | Retained state | Completion/failure behavior |
|---|---|---|---|
| Account addition/re-inclusion | Do not broaden use until scoped Apply/Confirm succeeds; then retrieve the added-account data and recalculate source-normalized/derived signals, ranking inputs, caches, Candidate Card output, Profile Preview, and revalidation output for the new committed set. | Prior narrower committed inclusion and outputs until the recalculation succeeds. | If retrieval/apply fails, remain on the narrower set; announce failure once and never render the uncommitted addition. |
| Account exclusion | Suspend excluded accounts from all use/display immediately; invalidate their source/normalized data, signals, ranking inputs, active/prefetched/rendered views, server/client/offline caches, Profile Preview, revalidation responses, and Back restoration; recalculate every output from the remaining committed set. Gate Discovery while purge/recalculation is incomplete if remaining state cannot be safely separated. | Minimal masked connection inventory only; unaffected included accounts when safely separable. | Purge/recalculation retries fail closed. Never restore the broader set. State “access removed” separately from “recalculation pending/failed.” |
| Disclosure Level upgrade | Do not broaden candidate-visible output until a separate scoped confirmation and policy persistence succeed. | Prior lower saved Disclosure Level and outputs. | Failure retains the lower level; no higher-detail preview is treated as saved or shareable. |
| Disclosure Level downgrade | Suppress all higher-level candidate-visible output immediately; invalidate broader active/prefetched/rendered views, caches, Profile Preview share state, revalidation responses, offline state, and Back restoration. | Connected source and private matching data that remain permitted; the confirmed narrower policy. | Persistence retries in the background fail closed. Never restore the broader level without a later explicit upgrade. |
| Disconnect | Remove connected status; stop all new portfolio-derived use; remove local source/normalized payloads, signals, ranking inputs, rendered/prefetched views, caches, previews, revalidation outputs, and masked connection inventory. | Only allowed categorical security/audit events without financial values. | Local removal remains effective if provider-side revocation fails; explain the provider boundary and require fresh authorization to reconnect. |
| Safety/permission invalidation | Suppress the affected viewer's financial disclosure and invalidate active/prefetched/history/offline/revalidation copies immediately. | Only data still permitted for unaffected contexts. | Failure remains closed; focus the replacement status heading and expose the safe next action. |

## Data Visualization & Disclosure

Candidate Card stays scannable: a dominant Portfolio Node Visualization, concise bounded relevance explanation, coarse Proximity, Freshness Indicator, Photo Obscure, Synthetic Data Label, and Swipe Actions. Candidate Detail carries progressive depth. Initial signal and relevance contracts are committed as follows:

- Snapshot uses allocation/asset-class mix derived from supported current positions and cash balances, diversification context, value/activity bands, activity recency, coarse Proximity, source coverage, and Freshness State; no security names or exact monetary values.
- Holdings may add supplied instrument names/symbols and kinds, position weights, defensible percentage performance, and activity categories; no quantities, exact monetary values, or transaction detail. Performance requires a named period, supported inputs, explicit currency treatment, source coverage, and freshness; otherwise it is labelled unavailable or unsupported.
- Full Detail may add supplied quantities, last-known prices, cost basis, exact values by currency, portfolio/account value, defensible performance amounts, and recent orders or historical activities. Orders and activities remain distinctly labelled because their coverage and cadence differ; a date-only activity never receives an invented time.
- Source facts and Portfolio-Derived Signals are labelled distinctly.
- “Why this person appears” uses proximity, the selected similar/diversified/complementary preference, one or two named derived traits, Disclosure Level, and freshness. It never states a score, guarantee, Invisible Cohort, formula, or worth judgment.

Ordinary Discovery cannot override a candidate's Disclosure Level. Exact live owner values appear only in Portfolio and owner-only Profile Preview; Discovery remains synthetic. Source facts and Portfolio-Derived Signals are visually and semantically distinguished. Account kind, institution, instrument kind, currency, dataset-specific period, units, selected-account coverage, and freshness travel with the claims they qualify. Multiple currencies remain separate unless architecture defines and discloses a defensible FX conversion policy. Unknown, empty, unavailable, failed, and unsupported are different states. Optional tax-lot detail is owner-only and outside the first-cut candidate-card contract. No portfolio or compatibility meaning is reduced to a universal score.

Before saving any Disclosure Level, the preview states that screenshots, memory, and combined-field inference cannot be revoked; Full Detail places this warning immediately beside its exact-value/activity summary and confirmation. The downgrade confirmation repeats the limit. Upgrade/downgrade failure behavior follows the permission-change consequence matrix.

Incoming Interest uses the same Candidate Card/Candidate Detail contract at the initiator's Disclosure Level. The lower-disclosure owner can Pass or express Interested without increasing or exposing more of the owner's own Disclosure Level.

`Source-to-Experience Trace` starts with included source data and branches into: (1) owner-only exact view, (2) private matching input, and (3) candidate-visible disclosure. Live-owner versus synthetic provenance travels through each branch. Disclosure Level filtering and visible-field effects appear only on branch 3; blocked/unsupported inputs remain explicit rather than flowing forward.

## Freshness/Revalidation

Safe stale-while-revalidate is an experience behavior, not an implementation prescription:

1. Show the last trustworthy permitted view immediately, with its visible Freshness State, source coverage, and last successful timestamp.
2. Revalidate in the background without blocking, resetting the deck, moving focus, or replacing the current task with a skeleton.
3. Apply immaterial updates nondisruptively. If a material change affects eligibility, ordering, visible values, or explanation, preserve the current reading position and show a notice that lets the user acknowledge the updated state before it rearranges their task.
4. Architecture owns thresholds for current, stale, and unusable. When the source crosses an unusable boundary or requires reauthorization, suppress new derived output and transition to Recovery Panel; do not keep presenting it as current.
5. Any permission-decreasing trigger overrides stale preservation and follows the permission-change consequence matrix; a last trustworthy view can contain only currently permitted data.

| Experience Freshness State | `Freshness Indicator` token | Required label/meaning |
|---|---|---|
| connected | `{components.freshness-indicator.current}` / `{components.freshness-indicator.current-dark}` | Connected; timestamp identifies last trustworthy refresh. |
| syncing | `{components.freshness-indicator.pending}` / `{components.freshness-indicator.pending-dark}` | Syncing/checking; not proof that holdings changed. |
| stale | `{components.freshness-indicator.stale}` / `{components.freshness-indicator.stale-dark}` | Stale; last trustworthy timestamp and architecture-defined usability remain explicit. |
| needs reauthorization | `{components.freshness-indicator.unusable}` / `{components.freshness-indicator.unusable-dark}` | Needs reauthorization; block new derived output and offer recovery. |
| failed | `{components.freshness-indicator.unusable}` / `{components.freshness-indicator.unusable-dark}` | Failed; name affected source/dataset and recovery. |
| disconnected | `{components.freshness-indicator.unusable}` / `{components.freshness-indicator.unusable-dark}` | Disconnected; no connected/current presentation or derived Discovery use. |

## Responsive & Platform

| Context | Behavior |
|---|---|
| Phone `<768px` | Mobile-first single column; App Navigation at bottom; Candidate Detail full-screen; Swipe Actions remain visible above safe-area/browser chrome; tables scroll only within their labeled region. |
| Tablet `768–1023px` | Single-column primary task with optional adjacent summary; navigation may use bottom bar or compact rail according to available width, with the same three areas. |
| Desktop `≥1024px` | Left App Navigation; Candidate Detail uses a wider two-column composition; Candidate Card remains one focal item, not a grid. |
| Standalone home-screen display | Opaque branded launch, safe-area padding, persistent app shell, no dependency on browser Back chrome, and the same deep-link/session recovery. Theme-aware status/browser chrome stays legible. |
| Browser tab/shared link | Meaningful public title and generic public preview metadata; authenticated titles remain neutral. No sensitive financial/identity data in URL, favicon badges, preview image, or metadata. |

Installability is an enhancement to access, not a separate product or offline financial experience. Public/static shell assets may cache safely; sensitive financial payloads, Candidate Detail, Profile Preview, Photos, and authenticated HTML snapshots must not leak through service-worker caches, launch/task-switcher snapshots, social previews, or OS-level recent-content surfaces. Architecture/build chooses the exact web-app manifest, Apple/cross-platform icons, browser metadata, caching controls, and platform-specific snapshot protections.

## Inspiration & Anti-patterns

- Adopt the recognizable continuous Swipe Deck and single combined-card decision from mainstream swipe dating, without exposing Invisible Cohorts or exact ranking logic.
- Preserve the Constellation direction's atmosphere, clean composition, sparse node map, and crisp evidence geometry. It is a composition reference, not a substitute for this contract.
- Reject photo-first cards, generic verification badges, wealth/status tiers, credit thresholds, public leaderboards, pay-to-see financial data, luxury imagery, red/green performance theater, trading urgency, false precision, and coercive disclosure.
- Reject chat in this milestone. Mutual Match ends with Photo reveal and return to Discovery.
- Reject invented matches history, silent deck widening, or hidden account/disclosure coupling.

## Key Flows

### Source requirement coverage

| Source requirement names | Key Flow coverage |
|---|---|
| FR-1 Explain consent before connection; FR-2 Authorize through SnapTrade test OAuth; FR-3 Represent connection and freshness; FR-19 Present available live portfolio data; FR-21 Demonstrate source-to-experience transformation | Flow 1; Flow 8 |
| FR-20 Preview the owner's Candidate Card | Flow 1 (Disclosure Level slice); Flow 7 (complete phone/desktop and pre-/post-match preview) |
| FR-5 Set discovery preferences; FR-6 Choose a Disclosure Level; FR-7 Explain preference effects without exposing internal logic | Flow 2; Flow 12 |
| FR-8 Derive a private matching profile; FR-9 Isolate live and synthetic data; FR-10 Assemble an eligible Swipe Deck; FR-11 Present a portfolio-first Candidate Card; FR-12 Make one swipe decision; FR-13 Handle discovery edge states | Flow 3; Flow 9; Flow 10; Flow 11; Flow 12 |
| FR-22 Generate a varied Synthetic Population | Flow 13 (non-public formative-review procedure) |
| FR-14 Create a Mutual Match; FR-15 Reveal identity after mutual interest | Flow 4; Flow 11 |
| FR-4 Disconnect and withdraw consent | Flow 5 |
| FR-16 Enforce adult-only positioning; FR-17 Communicate prohibited interpretations and conduct; FR-18 Preserve a future safety boundary; FR-25 Present the Findur proposition and brand; FR-26 Provide expected informational and legal surfaces; FR-27 Offer a signup-free synthetic trial (conditional stretch) | Flow 6 |
| FR-23 Complete a minimum Personal Profile; FR-24 Preview the complete profile | Flow 7 |

### Flow 1 — UJ-1. Maya connects and explores her live portfolio.

1. Maya enters the protected owner experience and reads Consent Panel: named provider/categories, minimum masked account metadata retrieval, private-use/visible-disclosure distinction, screenshot/inference limitations, and disconnect.
2. She explicitly continues to SnapTrade test OAuth, authorizing provider access and metadata retrieval only; no derivation or Discovery use begins. Findur preserves her intended Portfolio destination.
3. OAuth returns success and focuses the result heading. Account Inclusion Control appears with no accounts selected.
4. She makes a draft selection, reads the scoped Apply/Confirm summary, and confirms Findur use for those accounts. Only then does source retrieval/derivation begin; the committed summary reports “Using 2 of 4 connected accounts.”
5. Portfolio shows accounts, balances/values, holdings/positions, activities/orders, coverage, and Freshness Indicator with distinct empty/unavailable/unsupported states.
6. She follows Source-to-Experience Trace and opens Profile Preview Frame at Snapshot, Holdings, and Full Detail.
7. **Climax:** Maya can point to what came from selected accounts, what was derived privately, and what each Disclosure Level could reveal—without mistaking the view for complete wealth or advice.

Failure: OAuth denial/invalid callback/expiry/provider error returns a safe result with retry/back. Refresh failure preserves only an architecture-classified trustworthy view with timestamp; unusable or reauthorization-required state blocks new Discovery and routes to recovery.

### Flow 2 — UJ-2. Maya sets compatibility and disclosure boundaries.

1. Maya opens Profile preferences after she has a Usable Portfolio.
2. Preference Control sets a maximum distance and similar, diversified, or complementary compatibility.
3. Disclosure Level Control begins with no saved selection; Snapshot is visibly recommended but not preselected. Maya inspects Snapshot, Holdings, and Full Detail through concrete hidden/bucketed/derived/exact examples while the persistent summary says either “No saved level” or identifies the prior saved level and marks the inspected tier “Previewing — not saved.”
4. Profile Preview Frame shows the inspected level without saving it. Before save, Maya reads the screenshot/memory/inference limit; Full Detail places the stronger warning beside its exact-value/activity summary.
5. Maya activates Save disclosure level to commit the inspected tier and hears the concise saved level. Cancel restores the saved preview. Leaving with a changed draft asks Discard changes or Stay; no selection ever autosaves. She sees that preference, Disclosure Level, Proximity, and portfolio compatibility shape Discovery without exposing scores or cohorts.
6. **Climax:** Before Discovery, Maya can predict what another person could see and knows this is separate from selected-account coverage and private matching use.

Failure: an unconfirmed edit, preview switch, route exit, refresh, or Back action changes neither saved level nor visibility. The initial no-saved-level state keeps Discovery gated. A confirmed downgrade follows the permission-change consequence matrix and remains narrower through persistence failure; a failed upgrade retains the lower saved level.

### Flow 3 — UJ-3. Maya explores a deep, visual Swipe Deck.

1. Ready-owner launch lands Maya in Discovery at the exact session deck position or a new eligible deck.
2. Candidate Card presents a dominant Portfolio Node Visualization, bounded relevance explanation, coarse Proximity, Freshness Indicator, Photo Obscure, and Synthetic Data Label.
3. Maya uses summary/table equivalence to inspect the same chart meaning without relying on color.
4. She opens Candidate Detail; mobile uses full-screen and desktop two columns, with only disclosure-permitted sections.
5. She returns and the exact Candidate Card/deck position is restored.
6. She uses Interested once; the decision persists and the next card advances.
7. After this first settled Interested decision, Notification Control may ask “Want to know if it becomes mutual?” Enable triggers the browser permission request from that direct action; Not now continues without penalty or repeated prompting.
8. **Climax:** Maya makes one combined decision from meaningful synthetic portfolio evidence and human context, without seeing Photo, exact location, rank, or a worth score.

Failure: no eligible candidates produces an honest sparse state and voluntary preferences path; it never widens distance/disclosure. If detail becomes invalid, sensitive content disappears and safe Back returns to Discovery.

### Flow 4 — UJ-4. Maya reaches a Mutual Match and sees the Photo reveal.

1. Maya selects Interested on a reciprocating Synthetic Candidate; the input becomes idempotently pending and repeated activation does not duplicate the swipe.
2. If Findur is foregrounded, Discovery Activity Notice announces the Mutual Match. If it is not foregrounded and Maya enabled the category, a privacy-safe push says only “You have a new mutual match. Open Findur to reveal it.” An installed app may also show the count of unacknowledged events as a badge.
3. Push activation navigates to a non-sensitive Discovery activity route. Findur authenticates Maya and rechecks eligibility, block/report state, Disclosure Level, and event validity before fetching or rendering candidate content; notification payload, URL, preview, and badge contain no name, Photo, portfolio, institution, value, or compatibility detail.
4. If the push is dismissed, unavailable, denied, or unsupported, the authoritative unacknowledged event remains in Discovery. It is not a permanent matches list and disappears after acknowledgement or invalidation.
5. Mutual Match Reveal replaces Photo Obscure with the image named from the display name, for example “Photo of Alex,” and announces the result once. Synthetic Data Label remains visible and copy explains the reveal follows the original mutual decision.
6. Maya chooses Continue Discovery; the event is acknowledged, its badge contribution clears, and no chat invitation appears.
7. **Climax:** The Photo reveal remains immediate emotional payoff whether entered in-app or through a safe notification deep link, with no second appearance-based decision.

Failure: non-reciprocal interest advances normally without Photo reveal. Match-creation failure keeps a retry-safe state and never shows the Photo early. Invalidated, blocked, or already-acknowledged notification events open a privacy-safe status in Discovery and expose no candidate data. Push delivery failure never removes the pending in-app event.

### Flow 5 — UJ-5. Maya withdraws access.

1. Maya opens connection settings in Portfolio.
2. She invokes disconnect and reads Confirmation Dialog summarizing the disconnect row of the permission-change consequence matrix and provider-side limits.
3. She confirms once.
4. Active sensitive views close; App Navigation remains, Portfolio becomes disconnected, and focus moves to the Discovery gate heading.
5. The interface confirms the local deletion boundary and offers fresh authorization.
6. **Climax:** Maya can see that Findur stopped using and presenting the connection, and understands what Findur cannot change at the provider.

Failure: if provider-side revocation cannot complete, local deletion/use stop still completes and the unresolved provider boundary is stated with a safe next action; no stale portfolio content returns.

### Flow 6 — UJ-6. Alex understands Findur before signing in.

1. Alex lands on the Public Site in browser or installed shell.
2. He learns the portfolio-first premise, progressive Photo reveal, consent/disclosure distinction, 18+ boundary, and owner-only demonstration scope.
3. Public Header and Public Footer take him to About, Trust and safety, Privacy, Terms, and Contact or support without authentication.
4. He chooses protected owner entry. The deferred Guest Demo is not advertised in the first cut.
5. **Climax:** Alex can accurately describe Findur and its demonstration boundary before any financial connection request.

Failure: a Guest Demo link when deferred returns to the complete Public Site with an honest unavailable message; it cannot fall into owner authentication or OAuth.

### Flow 7 — UJ-7. Maya completes and previews her whole profile.

1. After connecting, Maya opens Profile and sees each required field plus its private/pre-match/post-match visibility.
2. She supplies the minimum Personal Profile, including adult confirmation, coarse location, relationship intent, Photo, and short self-description.
3. Validation keeps her values and points to incomplete fields; Discovery remains gated until valid.
4. She sets preferences and Disclosure Level, then opens Profile Preview Frame.
5. She switches Snapshot/Holdings/Full Detail, pre-/post-match, and representative phone/desktop compositions without changing saved settings.
6. **Climax:** Maya sees the whole live-owner presentation exactly within each disclosure and identity state before anything could be shared.

Failure: unsupported/missing live data is labeled rather than fabricated. A save failure preserves edits and keeps the previous valid preview until retry.

### Flow 8 — Account selection after OAuth (Maya, choosing the boundary of Findur use)

1. Maya returns from OAuth to connection success; no account is selected by default.
2. Account Inclusion Control presents a labelled checkbox group. She hears usable and unavailable accounts with reasons; Select all reports unchecked/mixed/checked.
3. She selects two accounts in the draft set. Draft “2 of 4 selected” remains distinct from committed “Using 0 of 4 connected accounts.”
4. Apply/Confirm summarizes the accounts, categories, private purpose, provider-access boundary, and coverage change. Maya confirms; only then does retrieval/derivation begin.
5. Concise status announcements report applying, recalculating, and “Using 2 of 4 connected accounts” without moving focus. Discovery unlocks only when at least one included account forms a Usable Portfolio.
6. Later in Portfolio settings, she excludes one account, confirms the scoped change, and the permission-change consequence matrix applies immediately.
7. **Climax:** Committed coverage and every permitted output reflect only included accounts; the excluded account cannot return through refresh, Back, offline state, or revalidation.

Failure: zero selected keeps Discovery gated. A failed addition retains the prior narrower committed set. A confirmed exclusion suspends the account immediately and keeps broader Discovery gated while purge/recalculation retries; it never rolls back to the broader set.

### Flow 9 — Direct/deep-link gating and recovery (Maya, reopening an installed candidate link)

1. Maya launches an installed deep link to Candidate Detail.
2. The opaque brand launch surface appears while session and route permission resolve; no candidate or financial snapshot is exposed.
3. If signed out, protected owner entry retains the validated internal destination.
4. After sign-in, gates check Personal Profile, included accounts, Usable Portfolio, freshness, and candidate eligibility; the active gate heading receives focus.
5. If permitted, Candidate Detail opens and establishes a safe Discovery return context.
6. **Climax:** Maya reaches the intended candidate without losing security boundaries, and Back returns to the exact safe deck state.

Failure: expired session, ineligible candidate, invalid external redirect, or invalidated detail goes to the most specific Recovery Panel and then Discovery/Public Site; sensitive parameters and previews are never echoed.

### Flow 10 — Progressive candidate detail (Maya, comparing evidence before deciding)

1. Maya opens Candidate Detail from the active Candidate Card.
2. Focus moves to the candidate heading; Synthetic Data Label, Disclosure Level, freshness, and Photo state remain present.
3. She reads the Portfolio Node Visualization summary, then navigates permitted holdings/activity sections and Chart Data Table.
4. Swipe Actions remain available without covering content at zoom or mobile safe areas.
5. Maya selects Pass or Interested once, or presses Back without deciding.
6. **Climax:** She gets enough depth to decide while the candidate's disclosure boundary remains exact across mobile full-screen and desktop two-column layouts.

Failure: revalidation removes permission/detail under the consequence matrix; content is replaced immediately, focus moves to the replacement heading, the reason is announced once, and safe Back remains.

### Flow 11 — Accessible non-visual discovery (Noor, keyboard and screen-reader user)

1. Noor enters Discovery and hears the route, deck position, candidate identity, Synthetic Data Label, Photo state, Disclosure Level, freshness, and concise portfolio summary.
2. She navigates directly from Portfolio Node Visualization to Chart Data Table/equivalent and hears labels, values, units, period, source coverage, and derived/source distinctions.
3. She opens Candidate Detail with Enter; focus moves to its heading and sections follow reading order.
4. She activates Interested through Swipe Actions rather than a gesture; duplicate input is ignored and announced.
5. On Mutual Match, the state and the named Photo reveal are announced once; reduced motion shows the final visual state immediately.
6. **Climax:** Noor completes the same one-stage decision and match outcome with equivalent evidence and no swipe, hover, color, size, or motion dependency.

Failure: an error is announced with its affected control and recovery; focus remains stable. Back returns focus to the originating card unless state invalidation requires the Discovery gate.

### Flow 12 — Incoming Interest asymmetry (Maya, reviewing interest without raising her disclosure)

1. In a seeded Discovery scenario, Maya—saved at Snapshot—opens the Incoming Interest child surface; no fourth navigation destination or inbox is created.
2. The Synthetic Candidate who already expressed interest appears at the initiator's Holdings or Full Detail Disclosure Level with persistent provenance, freshness, and an explanation of the asymmetric rule.
3. The interface explicitly confirms that Maya's own Snapshot level and candidate-visible information remain unchanged.
4. Maya reviews the permitted initiator evidence in Candidate Detail with the same chart/table and non-gesture alternatives as the Swipe Deck.
5. Pass records one decision and returns to the exact Discovery state. Interested records one decision and creates the existing Mutual Match/Photo reveal because the initiator's positive decision already exists.
6. **Climax:** Maya can evaluate the initiator's higher disclosure and respond without increasing or exposing more of her own Disclosure Level.

Failure: if the seeded interest becomes unavailable, stale beyond usability, or disclosure-invalidated, content fails closed under the consequence matrix; focus moves to the replacement heading and the safe action returns to Discovery. No matches history or chat is created.

### Flow 13 — Synthetic Population acceptance (Priya, formative-review facilitator)

This is a non-public review procedure, not an owner or Guest Demo surface.

1. Priya records the generator version, stable seed, and configured population size, then generates the Synthetic Population twice from the same inputs.
2. She verifies stable identifiers and equivalent outputs without hand-authored candidate fixtures, then regenerates with a different seed to confirm controlled variation.
3. She runs the acceptance matrix across similar, diversified, and complementary modes; Snapshot, Holdings, and Full Detail; inside/at/outside maximum-distance boundaries; varied portfolio patterns; and current, syncing, stale, needs-reauthorization, failed, and disconnected conditions.
4. She verifies guaranteed sparse eligibility, empty activity, unsupported fields, missing/unknown inputs, boundary candidates, and both reciprocal and non-reciprocal outcomes.
5. She confirms every generated candidate/profile/portfolio is labelled synthetic, internally coherent, exercises the shared visualization/disclosure rules, and contains no imported real personal or financial record.
6. She records deck-depth/scenario evidence and reruns generation from the recorded seed as another facilitator would.
7. **Climax:** Priya accepts FR-22 only when the population is reproducible, varied across every required mode/state, regenerable without hand-authored fixtures, and demonstrably free of real records.

Failure: any nondeterminism for a fixed seed, uncovered matrix cell, copied/real record, incoherent fixture, or hand-authored dependency fails the review. The population is not released to the demonstration until regenerated and the full procedure passes.

## Phase Blockers and Open Questions

- **Architecture:** define current/stale/unusable thresholds and material-change rules before freshness implementation can be accepted.
- **Architecture/provider:** verify whether SnapTrade supports provider-granular account authorization; UI must accurately distinguish provider authorization from Findur account inclusion.
- **Future architecture/provider:** preserve SnapTrade's hosted MCP connector as an owner-only AI evidence-retrieval option. Before use, verify permitted Findur/server-agent usage, OAuth and revocation boundaries, selected-account enforcement, AI-provider data handling and retention, freshness/provenance propagation, and whether written SnapTrade approval is required. It does not alter the first-cut UX or replace the Commercial/OAuth production assumption.
- **Product:** confirm whether Personal Profile needs fields beyond PRD A-4 before profile approval.
- **Product/qualified reviewers:** assign approval of Terms, Privacy, and trust/safety copy before external hosting.
- **Architecture/build:** define the installable web asset/metadata matrix, standalone/session restoration, safe-area strategy, offline-cache exclusions, and platform snapshot/privacy controls before mobile-web release acceptance.
- **Architecture/build:** define standards-based Web Push, browser capability detection, explicit permission/subscription lifecycle, minimal encrypted payload, authenticated notification-click routing, badge reconciliation, delivery retries, logout/revocation cleanup, and privacy-safe telemetry. On iOS/iPadOS, handle Home Screen installation requirements without blocking core use or repeatedly prompting.
