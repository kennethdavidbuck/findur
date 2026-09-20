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
