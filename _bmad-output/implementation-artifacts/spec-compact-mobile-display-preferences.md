---
title: 'Compact mobile display preferences'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<!-- frozen: intent -->
## Intent

**Problem:** Authenticated header display controls take up a disproportionate amount of room on small screens and announce an unhelpful saved-state message.

**Approach:** Hide authenticated header display controls at the existing mobile breakpoint. Keep public-site controls available. Replace the visible status text with a compact saving spinner whose label is available to assistive technology only while a save is in progress.
<!-- /frozen: intent -->

## Implementation Notes

- Reuse the existing `47.999rem` responsive breakpoint.
- Do not change the authenticated preference persistence model or the profile settings controls.
- Implemented the responsive authenticated-header hide, compact saving indicator, and localized assistive-only save/retry announcements outside the hidden header.
- Blind review found that hiding the header would otherwise remove mobile assistive feedback; the shell-level live region resolves it without visible status text.
