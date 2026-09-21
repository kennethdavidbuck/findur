---
title: 'Keep Account Selection Accessible During Onboarding'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/spec-fix-account-inclusion-ux.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The recent account-selection corrections infer that a committed account choice means the owner should automatically skip `/onboarding/accounts`, making the corrected screen inaccessible before the wider onboarding journey is complete.

**Approach:** Stop auto-redirecting committed choices away from `/onboarding/accounts`; load them preselected, send initial saves to an interim `/onboarding/portfolio` Showcase route, and add an expedient `/portfolio/accounts` edit route plus a simple localized link from the current Portfolio placeholder. Do not change Story 1.2 in this fix.

</frozen-after-approval>

## Implementation Notes

- Removed the committed-selection redirect from `PortfolioPage`; existing inclusion state now populates the chooser normally.
- Added `/portfolio/accounts` in the ordinary authenticated shell, reusing the chooser without its onboarding progress rail.
- Initial onboarding saves replace the account route with `/onboarding/portfolio`; later edit saves return to `/portfolio`.
- Added a localized Portfolio link to the edit route while preserving modified-click browser behavior and clearing transient save feedback on entry.
- Review is disabled only when no visible account is selected; saving an unchanged committed selection advances without issuing a redundant API write.
- The interim onboarding Portfolio route suppresses ordinary app navigation, and late edit-save responses cannot navigate after the chooser unmounts.
- Restored 80px of rail-top clearance so the first visible onboarding step begins below the absolutely positioned Findur wordmark.
- Per human direction, omit new automated tests; still run the repository-mandated backend lint command and a fast frontend typecheck.

## Review Triage Log

- **Medium — patched:** unchanged committed choices left returning onboarding users without a forward action; Review now remains available for a non-empty saved selection, and unchanged Save advances without resubmitting it.
- **Medium — patched:** a late edit-save response could replace a route visited while the request was pending; completion navigation now requires the chooser to remain mounted.
- **Medium — patched:** `/onboarding/portfolio` exposed ordinary app navigation despite remaining inside onboarding; it now uses the navigation-suppressed onboarding shell, with full Showcase progress presentation left to Story 1.2.
