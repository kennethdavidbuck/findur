---
title: 'Surface the depth of Findur planning evidence'
type: 'chore'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '{project-root}/AGENTS.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The README links the main BMad documents but still summarizes each discipline too broadly, leaving the concrete depth of research, product reconciliation, UX exploration, accessibility, responsiveness, architecture review, and story-level delivery work hard to see.

**Approach:** Add a compact, curated guide to the strongest evidence within each planning and delivery area, including UX visual explorations and key-screen mockups, accessibility and responsive contracts, source-backed research, PRD review/reconciliation, architecture invariants and adversarial resolution, and implementation tracking. Keep the README readable and let linked artifacts carry the detail.

</frozen-after-approval>

## Implementation Notes

- Added a curated “What is captured inside” guide beneath the BMad artifact map rather than expanding the top-level table into an exhaustive index.
- Surfaced research claims and citation verification, PRD reconciliation and resolved review findings, UX HTML mockups and accessibility/responsive commitments, architecture adversarial resolution, and the delivery/deferred-work tracking chain.
- Kept story completion out of the prose; `sprint-status.yaml` remains the live status source.
- Limited the change to `README.md` and this workflow record; application code and runtime behaviour are unchanged.

## Review Triage Log

- `medium` — The accessibility/trust review predates remediation and its negative verdict could be mistaken for current UX status; patched by labeling it historical, explaining its role, and stating that it does not describe the finalized spines.
- `medium` — Linking the complete HTML directory did not distinguish the selected design from alternatives and exploratory screens; patched by linking Constellation directly, naming the remaining artifact roles, and making written-spine precedence explicit.
- `medium` — The README claimed a visible reconciliation trail without linking the reconciliation records themselves; patched with direct representative links for the brief, SnapTrade research, and UX experience alongside the fidelity recheck.
