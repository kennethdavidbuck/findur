---
title: 'Polish the public landing-to-login experience'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The public Findur experience overuses internal, institutional language such as “Owner access” and makes the SnapTrade login action difficult to reach behind a long consent page. This weakens the original simple, fun dating-product vision and creates unnecessary friction in the landing-to-login journey.

**Approach:** Polish the public landing, header, About status copy, and SnapTrade entry so the product is immediately understandable, the protected action is plainly labelled “Log in,” and the SnapTrade action is visible without a long scroll. Preserve the existing Constellation visual direction and every trust boundary while using compact hierarchy and progressive disclosure; responsive behavior, 320px-to-desktop reflow, 200–400% zoom, EN/FR expansion, System/Light/Dark themes, keyboard focus, reduced motion, authorization gating, accessible status text, and 44–48px touch targets are acceptance requirements. Do not change backend/authentication behavior, authenticated product screens, or the owner-only demonstration boundary.

</frozen-after-approval>

## Implementation Notes

- Reframed the public copy around the dating experience while retaining the portfolio-first premise, progressive photo reveal, private-demo boundary, and English/French parity.
- Replaced the institutional owner-entry treatment with direct Log in actions in the header and hero; preserved the protected `/connect` route and authorization behavior.
- Reordered the SnapTrade page so the essential password, automatic-inclusion, and publication reassurance plus its action are visible on initial phone view; kept fuller use, sharing, read-only, and disconnect details immediately below.
- Reused the existing theme tokens, Constellation hero, action primitives, and responsive breakpoints; fixed the plain login link's keyboard focus selector and retained 48px primary mobile actions.
- Updated public-flow tests and visually inspected landing/login renders at 320×800, 360×800, and 1440×1000. Frontend tests, typecheck, lint, and build pass; lint retains two pre-existing Fast Refresh warnings. The mandated `cd backend && golangci-lint run` command could not start because `golangci-lint` is not installed in the environment.
- Review restored the architecture-approved isolated test-viewer model and the exact staged consent, credential, included-account, downstream-use, and complete-disconnect boundaries without changing the streamlined page hierarchy.
- Follow-up visual feedback enlarged and tightened the Perspective constellation node so its label stays contained, and gave About a distinct explanatory role covering portfolio signals, user boundaries, and mutual photo reveal.

## Review Triage Log

| Finding | Verdict | Evidence and disposition |
|---|---|---|
| BH-1 | medium | The architecture supersedes the single-owner model with up to five isolated test viewers. Patched all runtime copy; deferred correction of the locked frozen spec wording. |
| BH-2 | medium | “Password stays with SnapTrade” implied provider handling not established by this OAuth flow. Patched EN/FR to state the supported guarantee that Findur never receives or stores brokerage credentials. |
| BH-3 | medium | The compact consent copy omitted connection status from the initial masked inventory. Patched EN/FR and assertions. |
| BH-4 | medium | “Nothing” weakened the required no-default account-inclusion boundary. Patched EN/FR to name accounts explicitly. |
| BH-5 | medium | The shortened copy omitted derivation, previews, Discovery, and the explicit inclusion trigger. Restored them below the immediately visible action and added assertions. |
| BH-6 | medium | The copy overstated separate consent choices after inclusion. Patched it to state that account confirmation starts private analysis/previews while match-visible disclosure remains separate. |
| BH-7 | medium | Disconnect copy described only portfolio data rather than the complete Findur aggregate. Patched EN/FR to describe leaving Findur and complete active-storage deletion. |
| BH-8 | low | English metadata retained the vague “introductions” framing. Patched the title and description to identify portfolio-first dating and mutual photo reveal. |
| BH-9 | medium | Tests asserted fragments without proving the staged-consent distinctions. Expanded EN/FR assertions for credentials, connection status, inclusion, downstream gates, disclosure separation, and disconnect. |
