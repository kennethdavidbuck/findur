- source_spec: `_bmad-output/implementation-artifacts/spec-render-walking-skeleton.md`
  summary: Confirm that Render allocates `findur-api-kdb.onrender.com` and that the static-site API rewrite reaches it; otherwise serve the Vite build from Go.
  evidence: The Blueprint reserves a globally allocated hostname that cannot be proven before creation. A failed allocation or rewrite would make the current same-origin topology unusable and would settle this deferred finding.
- source_spec: `_bmad-output/implementation-artifacts/spec-public-site-foundation.md`
  summary: Re-derive the canonical public-site spec so language and theme preferences are footer-only.
  evidence: The user explicitly removed the header preference controls during implementation, superseding `public-surface-contract.md`, which still says the Public Header contains them; bmad-spec is the canonical spec's sole writer.
- source_spec: `_bmad-output/implementation-artifacts/spec-public-site-foundation.md`
  summary: Re-derive the canonical public-site spec so the footer does not repeat Home and About navigation.
  evidence: The user explicitly removed redundant footer route links during implementation, superseding `public-surface-contract.md`, which still requires them; bmad-spec is the canonical spec's sole writer.
- source_spec: `_bmad-output/implementation-artifacts/spec-0-2-begin-hosted-snaptrade-authorization-safely.md`
  summary: Define an OIDC discovery metadata cache freshness and refresh policy before initiation is opened in production.
  evidence: The Story 0.2 client intentionally caches validated discovery metadata until process restart, while the provider's endpoint-rotation expectations and an appropriate refresh SLA are not established by the current architecture or provider contract.
- source_spec: `_bmad-output/implementation-artifacts/spec-0-5-prove-authenticated-snaptrade-access-with-masked-inventory.md`
  summary: Define the correct per-connection repair action when a ready inventory contains both active and disabled SnapTrade connections.
  evidence: The current UI identifies the disabled connection but offers recovery only for a top-level disabled state; the repository and pinned API contract do not establish whether hosted reauthorization repairs one disabled connection or whether SnapTrade requires a distinct connection-repair flow.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Define and implement an aggregate provider-call budget for bulk multi-account inclusion.
  evidence: Each account currently receives a fresh 10-second budget while the HTTP write timeout is 15 seconds; the human explicitly accepted deferral of this multi-account risk to keep Story 1.1 moving.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Preserve removable masked UI rows when a committed account is absent from the current inventory head.
  evidence: If that lifecycle state is reachable, the current UI retains the ID in its draft but has no checkbox or masked label; provider disappearance and retained-identity semantics must be settled together.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Define what positive provider evidence makes a provisional unknown-status or unknown-category account investment-usable.
  evidence: Successful timestamped empty datasets currently pass completeness, while the approved intent does not say whether empty positions are conclusive proof or require at least one investment row.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Carry provider retry timing through inclusion failures and suppress premature retries.
  evidence: `ProviderError.RetryAt` is discarded by inclusion failure mapping; Story 1.3 owns the centralized rate-limit and backoff policy needed to resolve this coherently.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Harden pre-existing account-number redaction against extremely large provider-controlled regular expressions.
  evidence: `safeAccountLabel` builds a `MustCompile` expression from an unbounded account number, which could panic on a pathological response even though ordinary provider values are short.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Retain one browser idempotency key across an ambiguous account-inclusion POST outcome.
  evidence: A lost successful response followed by immediate retry uses a new key and stale version; authoritative reload recovers today, and the human-directed narrow pass deferred the stronger retry behavior.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Expand inclusion recovery verification for categorical service failures and immediate failed POST responses.
  evidence: Current tests prove generic provider failure and persisted failed-state reload, but do not cover every service mapping or a failed mixed-change POST in the same render cycle.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Keep a failed draft addition removable if refreshed inventory later marks it unselectable.
  evidence: The checkbox is disabled for an unselectable, uncommitted ID even when recovery reconstructed that ID into the draft, so the owner cannot narrow the retry without a reload or other state change.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Align the inclusive activity query dates with an exact 30-calendar-date definition.
  evidence: SnapTrade documents both endpoints as inclusive, so subtracting 30 days spans 31 dates; this direct off-by-one was deferred in the human-directed narrow review pass.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md`
  summary: Define retained masked account rows and reasons for disabled brokerage connections without calling the provider.
  evidence: Disabled connections intentionally skip account retrieval, while UX planning expects unavailable accounts to remain recognizable; the stable cached-inventory lifecycle must decide how those requirements combine.

- source_spec: `_bmad-output/implementation-artifacts/spec-fix-account-inclusion-ux.md`
  summary: Harden the local OIDC integration fixture with challenge-bound, expiring, single-use authorization codes.
  evidence: The fixture's token endpoint predated this correction and accepts any non-empty code verifier and authorization code; binding issued codes to the S256 challenge would let browser integration detect wrong-verifier and replay regressions without expanding the account-selection UX fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-public-site-polish.md`
  summary: Correct the frozen polish spec's obsolete owner-only demonstration wording.
  evidence: The architecture spine supersedes the single-owner assumption with up to five isolated SnapTrade test viewers; runtime copy was corrected, but the workflow forbids implementation agents from changing frozen human-owned intent.

- source_spec: `_bmad-output/implementation-artifacts/spec-deployment-health-smoke.md`
  summary: Update the governing Epic 0 planning acceptance criteria from exact frontend/API SHA equality to the approved deployment-health contract.
  evidence: The approved follow-up treats differing component revisions as diagnostic information, but `epics.md` still requires exact-SHA convergence; changing planning artifacts is deferred to a planning correction rather than this implementation fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md`
  summary: Preserve a committed account as an explicit lifecycle row when current inventory metadata is absent or the inventory lifecycle is not ready.
  evidence: The Showcase inner-joins the current inventory head and does not consult `current_status`, so an account can disappear or old facts can remain readable during later pending or failed inventory states; the human accepted deferral to close Story 1.2.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md`
  summary: Tighten Showcase API fail-closed validation for expired rows, future or malformed timestamps, currency codes, and dataset-specific row shapes.
  evidence: Review confirmed that expired rows still serialize and that current runtime/schema checks admit malformed evidence shapes; the human explicitly asked not to expand the privacy pass during close-out.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md`
  summary: Complete Showcase evidence presentation for non-USD money, stale summaries/actions, activity price and units, and fully localized French context.
  evidence: Review confirmed these display gaps, including a dollar sign on every currency; the human accepted the current page and requested wrap-up.

- source_spec: `_bmad-output/implementation-artifacts/spec-1-2-inspect-the-private-portfolio-showcase.md`
  summary: Expand Showcase recovery and lifecycle regression coverage.
  evidence: Current tests do not click the Showcase retry action, show populated stale rows, or independently exercise every fail-closed lifecycle predicate; deferred at the human's request to keep moving.
