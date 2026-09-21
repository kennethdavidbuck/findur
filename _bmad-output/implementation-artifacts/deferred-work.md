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
