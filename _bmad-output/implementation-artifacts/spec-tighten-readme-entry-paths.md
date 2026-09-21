---
title: 'Tighten README entry paths'
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

**Problem:** The README still makes the HTML design explorations harder to view than necessary, lacks a visual brand cue at the top, and lets an optional native-development path distract from the accurate one-command Compose quick start.

**Approach:** Reuse the existing constellation mark as a restrained README header image, link the BMad Method mention to its official documentation, add one concise local-browser command plus labeled URLs for the selected direction and five key-screen HTML designs, and make `./scripts/compose-up.sh` the sole documented startup path while retaining only the most useful development and test follow-ups.

</frozen-after-approval>

## Implementation Notes

- Reused `frontend/public/icon.svg`, the existing constellation mark named by the design spine, as the restrained README header image; no new logo or generated artwork was introduced.
- Linked the BMad Method mention to its official documentation.
- Added one local `python3 -m http.server` command and direct browser URLs for the selected composition and five self-contained key-screen HTML explorations.
- Removed the optional native/database startup section so `./scripts/compose-up.sh` is the sole Quick Start command and is explicitly described as starting PostgreSQL with the rest of the stack.
- Limited the change to `README.md` and this workflow record; application code and runtime behaviour are unchanged.

## Review Triage Log

- `medium` — The design preview server bound to all interfaces by default; patched with `--bind 127.0.0.1` so the optional viewer remains loopback-only.
- `low` — The design-viewing path did not name its Python 3 prerequisite; patched beside the command without adding Python to the application Quick Start requirements.
- `medium` — Removing native startup also removed the Go and Node.js/npm prerequisite needed by the verification commands; patched beside Verification with the repository's pinned `.tool-versions` source.
- `low` — The design-viewing heading was unnecessarily nested and easy to miss; promoted to a level-three section within the planning narrative.
