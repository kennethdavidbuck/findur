---
title: 'Add README hero and stage highlights'
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

**Problem:** The small README icon does not capture the updated landing page's full-width constellation atmosphere, and the BMad stage map still sends readers into dense artifacts before conveying the most important decisions and features from each stage.

**Approach:** Rebase onto the updated main branch, replace the icon with a full-width Findur constellation hero aligned with the revised public page, and rewrite every stage summary to surface its concrete product, research, UX, architecture, and delivery highlights while retaining links to deeper evidence.

</frozen-after-approval>

## Implementation Notes

- Rebased the documentation branch onto updated `origin/main` at `b861582359e452733b1618d939327d95bfe26a54` before adapting the README to the current landing-page direction.
- Generated a 2172×724 Findur hero with OpenAI's built-in image generator from a fresh constellation prompt, then used an edit pass to add the exact `FINDUR` wordmark. No third-party image was used. The repository copy is a visually checked, 72 KB WebP at `docs/assets/findur-readme-hero.webp`; the original generated PNG remains in the generator's output history.
- Replaced generic stage descriptions with compact, evidence-backed highlights covering the product premise, provider research, requirements, UX, architecture, and epic structure.
- Kept current story progress out of the overview; the README points to the living sprint status and implementation artifacts instead.
- No application code or runtime configuration changed.

## Verification

- `file docs/assets/findur-readme-hero.webp` — passed; 2172×724 WebP.
- Visual inspection of the repository WebP — passed; wordmark and four-node constellation composition remained intact after compression.
- Local README link/image validator — passed; all 26 repository-relative targets resolve.
- `docker compose config --quiet` — passed.
- `sh -n scripts/compose-up.sh` — passed.
- `git diff --check` — passed.

## Review Triage Log

- `medium` / patched — The 1.5 MB PNG was excessive for a README; converted it to a visually checked 72 KB WebP.
- `low` / rejected — The panoramic details become smaller on narrow screens, but the centered wordmark remains legible and the following H1 carries the semantic product name; a separate responsive asset would add complexity without losing information.
- `low` / patched — The original alt text described only four nodes despite the decorative detail; the banner now has empty alt text because the adjacent H1 supplies its meaning without duplication.
- `low` / patched — Asset provenance was incomplete; the implementation notes now record the generation/edit mode, source basis, retained original, and repository derivative.
- `low` / patched — Stage cells were longer than necessary; each was condensed while retaining the decisions the user asked to surface.
- `medium` / patched — The research row could imply that order retrieval was adopted; it now describes provider capabilities and explicitly notes the narrower adopted architecture.
- `medium` / patched — “Hosted OAuth” blurred distinct provider experiences; the row now names SnapTrade's OAuth/OIDC flow.
- `medium` / patched — “One durable swipe” overstated persistence; it now says “one persisted in-session swipe decision.”
- `medium` / patched — “No public launch” conflicted with the hosted demonstration; it now excludes a real-user dating-service launch.
- `false` — The UX row described WCAG 2.2 AA as an accessibility floor, not proven conformance; it was tightened to “specified” language for additional clarity.
- `medium` / patched — “The design adds” sounded like implementation status; the architecture row now says what the architecture specifies.
- `medium` / patched — The delivery row omitted Epic 0's deployment and integration foundation; it now includes exact-revision, same-origin, synthetic-integration foundations.
- `low` / patched — The implementation record lacked check results; this Verification section records only the checks actually run.
