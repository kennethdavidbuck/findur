# PRD Quality Review Recheck — Findur

## Resolved — Define the public-entry and owner-access boundary

FR-2 now limits OAuth and live data to a preconfigured authenticated owner and expressly denies both to unauthenticated visitors and Guest Demo sessions; conditional FR-27 confines any Guest Demo to isolated, ephemeral synthetic data and names the architecture-sizing decision owner and trigger (§4.1, FR-2; §4.9, FR-27; §11.10).

## Resolved — Make the prototype Disclosure Level contract decidable before implementation

FR-6 now defines Snapshot, Holdings, and Full Detail field-by-field, including the visibility rule and immediate downgrade invalidation, while addendum §1.2 repeats the same contract and prevents UX from changing it silently (§4.4, FR-6; addendum §1.2).

## Resolved — Resolve or explicitly constrain disconnect retention

FR-4 now requires immediate deletion of locally stored source payloads, derived signals, rendered/cached views, and permits only categorical non-financial security or audit events to remain (§4.1, FR-4).
