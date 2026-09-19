# UX Responsive Recheck

## Result

The three prior high-severity findings are **resolved** in the current PRD and addendum. The prior medium finding on profile completion and eligibility is **reduced**, not fully resolved.

| Prior finding | Status | Current evidence |
| --- | --- | --- |
| Persistent synthetic provenance | **Resolved** | [PRD FR-9](prd.md#L252) requires every synthetic Candidate Card, detail, and Mutual Match state to carry a persistent, legible marker that survives expansion, responsive layout, and navigation (line 261). [FR-24](prd.md#L188) separately requires the owner preview to identify live data and synthetic preview states to identify synthetic data (line 197). |
| Disclosure-level policy matrix | **Resolved** | [PRD FR-6](prd.md#L214) now defines exactly three levels—Snapshot, Holdings, and Full Detail—with their permitted/withheld fields, discovery compatibility, and initiation asymmetry (lines 216–227). [Addendum §1.2](addendum.md#L18) repeats the product contract as a matrix and prohibits UX from altering field policy or initiation asymmetry silently (lines 32–40). |
| Downgrade/disconnect active-render and cache invalidation | **Resolved** | [PRD FR-6](prd.md#L214) now requires an immediate invalidation of active views, rendered and prefetched data, caches, and back-navigation restoration on a downgrade (line 225). [Addendum §3.2](addendum.md#L165) extends the same contract to disconnect and server/client caches (line 176). |
| Profile completion and eligibility | **Reduced** | [PRD FR-23](prd.md#L176) now gates Discovery on valid required profile fields and a Usable Portfolio, identifies incomplete requirements, and returns the owner to the relevant flow (line 186). The documents still do not state draft/save behavior, failed photo processing, or the consequence when an edit invalidates an already completed profile; these are now bounded UX-state gaps rather than an absent eligibility contract. |

## Counts

| Status | Count |
| --- | ---: |
| Resolved | 3 |
| Reduced | 1 |
| Open | 0 |
| Still high | 0 |
