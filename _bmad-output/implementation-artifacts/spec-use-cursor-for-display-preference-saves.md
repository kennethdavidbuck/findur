---
title: 'Use cursor for display preference saves'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Adding a header spinner during an authenticated display-preference save changes the flex layout and shifts the locale and theme controls.

**Approach:** Remove the inline spinner and show the browser progress cursor on the preference controls while their authenticated save is pending.

</frozen-after-approval>

## Implementation Notes

- Reuse the existing authenticated preference state; do not change persistence or the screen-reader announcements.
- Replaced the inline spinner with a state class that applies the browser progress cursor to the controls and their clickable radio labels, preserving layout width during saves.

## Review Triage Log

- `medium` / defer — concurrent authenticated preference saves can clear the shared saving state when the first request resolves while a later request remains pending. This pre-existing persistence-state race needs a broader request-sequencing fix than this cursor-only correction.
- `low` / rejected — no focused test asserts the pending cursor class. The existing frontend suite passes; adding a controlled asynchronous authenticated-preference harness is disproportionate to this small presentational change.
