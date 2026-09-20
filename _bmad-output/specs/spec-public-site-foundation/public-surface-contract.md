# Public Surface Contract

## Route catalogue

| Route | Required content | Primary actions |
|---|---|---|
| `/` | Findur proposition; portfolio-first discovery; progressive identity reveal; concise explanation of how the concept works; 18+ and evaluation-demonstration boundary | Read About; unavailable owner entry |
| `/about` | Product purpose; why portfolio context precedes identity; consent and disclosure posture; current synthetic-candidate and non-public-launch boundary | Return home; unavailable owner entry |

Both routes are public, static, directly addressable, refresh-safe under the existing SPA fallback, and assigned localized document titles and descriptions. Intentional preview metadata remains generic and contains no candidate, identity, or financial detail.

## Shared public shell

The Public Header contains the Findur wordmark/home link, About navigation, language control, theme control, and owner-entry action. On narrow viewports, a labeled React Aria menu may collapse navigation without hiding either public route or either preference control. The current route is identified without relying on color alone.

The Public Footer repeats Home and About, states the evaluation-demonstration boundary, and exposes language/theme controls when the compact header makes them less immediate. It does not expose links for Privacy, Terms, Trust and safety, Contact/support, or Guest Demo until those surfaces are implemented.

Owner entry is a readable unavailable control with adjacent or associated explanatory text. It is not a dead link, does not imply public signup, and makes no request.

Every route transition updates the document title and moves focus to the route heading. Language and theme changes preserve the route and usable focus.

## Reusable foundation

Implement one semantic token layer for the adopted canvas, surface, text, muted text, border, primary, secondary, accent, focus, and state roles in both light and dark modes. Add typography, spacing, content-width, reading-width, corner, safe-area, and responsive tokens required by these pages. Component styles consume semantic tokens rather than embedding page-specific color values.

The focused primitive inventory is:

- public layout and main-content landmark;
- constrained content container and reading-width variant;
- section composition and shared heading/body/eyebrow styles;
- primary action, secondary action, and inline navigation link treatments;
- Public Header and Public Footer;
- language and System/Light/Dark theme controls; and
- compact public navigation menu where required by viewport width.

React Aria Components owns the behavior of interactive controls for which it provides a suitable primitive. Native landmarks, headings, paragraphs, and anchors remain semantic HTML. Shared controls implement the adopted interaction-state contract once; pages do not redefine hover, focus, pressed, or unavailable behavior.

## Localization and theme behavior

English and French catalogues contain every visible string, accessible name, title, and metadata value owned by this slice. Missing keys fail visibly during development or tests rather than silently falling back to mixed-language UI. French copy is coherent French and may reflow; controls do not truncate it.

Public locale and theme preferences may persist in browser storage because they are non-sensitive. With no stored locale, use English. With no stored theme, use System. Apply the resolved theme early enough to avoid a conspicuous incorrect-theme flash. While System is selected, respond to `prefers-color-scheme` changes; explicit Light or Dark selection overrides the system value.

## Verification matrix

| Area | Required evidence |
|---|---|
| Routes | Direct render and navigation for `/` and `/about`; refresh compatibility is preserved by the deployed SPA fallback. |
| Localization | Every route and shared control switches fully between EN and FR, preserves route, and restores the stored choice. |
| Theme | System/Light/Dark each resolve correctly; explicit choice restores; System reacts to media-query changes. |
| Accessibility | Landmarks and heading hierarchy are meaningful; current navigation is exposed; compact menu and selectors operate by keyboard; focus is visible; unavailable owner entry is announced; touch targets and contrast meet the adopted contract. |
| Responsive/reflow | Content remains usable on contemporary phone and desktop widths and under the adopted 200% and 400% zoom/reflow checks; French labels do not clip or hide actions. |
| Motion | Any atmospheric or state motion is nonessential and has a complete reduced-motion presentation. |
| Privacy/network | Loading and operating either route emits no application API request and stores only locale/theme preferences. |
| Quality gates | Frontend lint, typecheck, automated tests, and production build pass. |
