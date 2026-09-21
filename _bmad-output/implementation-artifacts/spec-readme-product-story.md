---
title: 'Reframe the README around the Findur product story'
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

**Problem:** The README accurately documents infrastructure and deployment, but it does not yet give readers a concise account of what Findur is, how its product and technical decisions were developed, or how the implemented slice relates to the broader plan.

**Approach:** Reorganize the README around an inviting, playful product overview, the product boundary, BMad planning-to-delivery story, and a streamlined local quick start. Link the primary planning and delivery artifacts, identify Codex with GPT-5.6 Sol at medium and high reasoning effort as the principal AI tooling, preserve technically important verification and deployment guidance, avoid a story-level progress snapshot, and do not mention the repository's private submission context.

</frozen-after-approval>

## Implementation Notes

- Reframed `README.md` around the product concept, trust boundary, BMad artifact chain, and local evaluation path while preserving concise verification and deployment guidance.
- Kept progress intentionally dynamic: the README explains the roles of epics, story specs, and `sprint-status.yaml` without naming the currently completed stories.
- Shifted the product voice toward a playful dating experience—“Meet the portfolio before the profile photo”—while keeping financial-data claims and consent boundaries precise.
- Added a repository-as-knowledge-base section explaining how a reader or coding agent can trace decisions across the versioned BMad artifacts and implementation evidence.
- Tightened native prerequisites and PostgreSQL startup, pointed provider configuration to its canonical validation source, and restored the one-time Render bootstrap boundary and current free-database lifetime.
- Used the official model identifier `gpt-5.6-sol` and described medium/high reasoning effort; clarified that AI tooling is not part of the application runtime.
- Limited the implementation footprint to `README.md` and this workflow spec; no application code, configuration, API, or runtime behaviour changed.

## Review Triage Log

- `medium` — The planned profile/discovery journey and “delivered” wording could imply unfinished flows were available; patched by labeling the journey as planned, changing the process wording to “being planned and built,” and retaining the artifact-based status pointers instead of a story snapshot.
- `medium` — Removing the one-time `bootstrap-once` image procedure made a fresh Render Blueprint deployment incomplete; patched by restoring the bounded bootstrap publication steps and immutable normal-release rule.
- `medium` — The native PostgreSQL example did not identify a concrete database matching its credentials; patched with `docker compose up --wait postgres` and the Compose service's synthetic local connection string.
- `low` — Native prerequisites omitted the pinned Go and Node.js/npm toolchains; patched by separating container and native requirements and linking `.tool-versions`.
- `medium` — The live SnapTrade configuration boundary was too vague to be actionable; patched with the required configuration categories and a link to the canonical environment names and validation rules.
- `low` — The playful opening lacked a concrete example of portfolio-informed compatibility; patched with similar, complementary, and diversification-oriented discovery examples while preserving the non-wealth boundary.
- `low` — “Temporary” weakened the operational warning for free Render Postgres; patched with the current official 30-day limit, lack of backups, maintained Render documentation, and demonstration-window guidance.
