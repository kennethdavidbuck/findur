---
title: 'Make the public site’s dating purpose explicit'
type: 'bugfix'
created: '2026-09-22'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A first-time visitor can read Findur’s public site as an investing or portfolio product because its hero uses a constellation metaphor and portfolio-first language without directly naming dating. Customer feedback confirms that this ambiguity has already occurred.

**Approach:** Revise the compact public English and French catalogues so the first visible landing content identifies Findur as a dating app, explains that portfolios offer a different first impression and conversation starter, and reinforces that romantic matches—not wealth or performance—are the point. Preserve the page structure, visual hierarchy, demonstration boundary, and supported consent claims.

</frozen-after-approval>

## Implementation Notes

- Updated the English and French public landing and About catalogues, plus their metadata, to name Findur as a dating app in the first visible content and frame portfolio signals as conversation starters.
- Kept the existing Landing and About component structure, demonstration boundary, mutual-match photo reveal, disclosure language, and explicit anti-wealth-ranking statements unchanged.
- Updated public-site assertions to make the direct dating hero copy a regression-tested contract in both locales.
- Blind review closed direct-entry ambiguity in the public login and About metadata, reduced duplicate hero wording, and made the dating-profile step more natural in both locales.

## Review Triage Log

- `medium / patch` — The `/connect` metadata was brokerage-only, so a visitor arriving from search could still infer an investing product. English and French now identify the private dating demo without changing consent claims.
- `medium / patch` — The visible `/connect` login content was brokerage-only, so a direct visitor could still infer an investing product. Its eyebrow and introduction now identify the private dating demo without changing consent claims.
- `low / patch` — The generic About metadata title did not identify the product category. It now names dating for investors in both locales.
- `low / patch` — Repeating “dating app” in the English and French hero eyebrow and heading weakened the compact hierarchy. The eyebrow now restores the established portfolio-first framing while the H1 remains explicit.
- `low / patch` — The profile-step wording was awkward and less direct than the requested dating language. It now says that people meet through portfolio-led dating profiles, with an idiomatic French counterpart.
