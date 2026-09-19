# Product-Fidelity Recheck — Prior High Findings

**Scope:** Current `prd.md` and `addendum.md`, limited to the four findings rated High in `review-product-fidelity.md`.
**Result:** All four are **resolved**. No prior High finding remains open.

| Prior finding | Status | Current evidence |
| --- | --- | --- |
| **H-1 — Public owner-access boundary** | **Resolved** | FR-2 now limits live authorization and owner data to the preconfigured authenticated owner, expressly denying OAuth, owner routes, and owner data to unauthenticated and Guest Demo sessions (`prd.md` §4.1, FR-2, consequences 5). UJ-6 distinguishes protected owner entry from an optional signup-free synthetic Guest Demo (`§2.4`). FR-27 requires the Guest Demo to use only synthetic data and forbids OAuth, Portfolio Showcase access, and owner-data retrieval (`§4.9`). Addendum §3.1 requires strictly separated owner and guest contexts. |
| **H-2 — Disclosure discovery/initiation contract** | **Resolved** | The PRD now names Snapshot, Holdings, and Full Detail, with exact field contracts (`prd.md` §4.4, FR-6). It expressly requires regular Discovery to include only candidates at the viewer's level or lower, and preserves the higher-to-lower inbound-initiation exception without increasing the recipient's disclosure (FR-6). Addendum §1.2 repeats the same matrix and states that labels may not alter the field contract or initiation asymmetry silently. |
| **H-3 — Persistent synthetic provenance** | **Resolved** | FR-9 now requires every Synthetic Candidate Card, detail view, and Mutual Match state to carry a persistent, legible synthetic-data indicator that survives expansion, responsive layout, and navigation (`prd.md` §4.5). NFR-5 requires live and synthetic data to remain distinguishable in storage, processing, and presentation. Addendum §3.4 also specifies a provenance field and separate fixture generation. |
| **H-4 — Active-view/cache invalidation after disconnect or downgrade** | **Resolved** | Disconnect now immediately deletes local source payloads, derived signals, and rendered/cached portfolio views, with a stated audit-event exception only (`prd.md` §4.1, FR-4). Downgrade invalidates active views, rendered/prefetched data, and caches, including back navigation (FR-6). Addendum §3.2 expands this to server/client caches and navigation-history restoration. |

## Counts

| Status | Count |
| --- | ---: |
| Resolved | 4 |
| Reduced | 0 |
| Open | 0 |
| Still High | 0 |
