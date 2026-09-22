---
title: 'Fix Showcase Provider-Call Race'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The browser integration journey measures provider reads after an unrelated active-session routing check, allowing the asynchronous sync worker to add a read and falsely attribute it to rendering or reloading the Portfolio Showcase.

**Approach:** Keep the completed-session routing coverage, but run it outside the Showcase reload measurement window so the assertion compares provider calls immediately before and after only the browser reload.

</frozen-after-approval>

## Implementation Notes

- The application’s Showcase read path stays unchanged: it is provider-free and reads persisted data only.
- `test/integration/browser-session.mjs` is the only planned code change; no runtime behavior, schemas, database state, or external side effects change.
- The rendered-table readiness check is insufficient because empty datasets do not render tables. The journey now polls the persisted inclusion state until its initial change is committed before measuring provider calls.
- Moved the completed-session routing assertion after the reload provider-call comparison. That route test can now take as long as it needs without expanding the measured Showcase interval.
- Restored the rendered-table readiness wait after the completed-session route check, so the following responsive-layout assertions never run against the Showcase loading state.

## Review Triage Log

- `patched` — The completed-session route check only waits for a heading, while the responsive assertions require rendered tables. The journey now waits for the five expected tables after that callback.
- `patched` — The initial-sync wait could otherwise fail without useful state. Its assertions now include the final inclusion response.
