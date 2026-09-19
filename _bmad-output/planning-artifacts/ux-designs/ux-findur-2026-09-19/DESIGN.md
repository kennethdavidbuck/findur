---
name: Findur Constellation
description: Dark-first, atmospheric visual system for portfolio-first dating across responsive web, English and French, and light and dark themes.
status: draft
sources:
  - ../../briefs/brief-findur-2026-09-19/brief.md
  - ../../briefs/brief-findur-2026-09-19/addendum.md
  - ../../prds/prd-findur-2026-09-19/prd.md
  - ../../prds/prd-findur-2026-09-19/addendum.md
  - ../../prds/prd-findur-2026-09-19/research-market-landscape.md
  - ../../research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md
updated: 2026-09-19
colors:
  canvas: '#EEF3FF'
  surface: '#FFFFFF'
  surface-subtle: '#E4EBF8'
  surface-emphasis: '#D9E4F6'
  overlay: '#FFFFFF'
  text: '#10172B'
  text-muted: '#52617E'
  text-disabled: '#77839B'
  border: '#A7B6D2'
  border-strong: '#687A9B'
  primary: '#285CC7'
  on-primary: '#FFFFFF'
  primary-subtle: '#DCE7FF'
  secondary: '#06736C'
  on-secondary: '#FFFFFF'
  secondary-subtle: '#D4F3EF'
  accent: '#8050B5'
  on-accent: '#FFFFFF'
  warning: '#7A4800'
  on-warning: '#FFFFFF'
  warning-subtle: '#FFE8B8'
  danger: '#A52B3A'
  on-danger: '#FFFFFF'
  danger-subtle: '#FFE0E4'
  success: '#116B4F'
  on-success: '#FFFFFF'
  success-subtle: '#D4F3E6'
  focus: '#6B4300'
  scrim: '#10172B'
  canvas-dark: '#0B1020'
  surface-dark: '#111A31'
  surface-subtle-dark: '#18233E'
  surface-emphasis-dark: '#202D4A'
  overlay-dark: '#17213A'
  text-dark: '#F4F7FF'
  text-muted-dark: '#AEB9D5'
  text-disabled-dark: '#7784A2'
  border-dark: '#465574'
  border-strong-dark: '#7182A5'
  primary-dark: '#8EB3FF'
  on-primary-dark: '#071021'
  primary-subtle-dark: '#1B315D'
  secondary-dark: '#65EADB'
  on-secondary-dark: '#062B29'
  secondary-subtle-dark: '#123C3B'
  accent-dark: '#D9A5FF'
  on-accent-dark: '#251135'
  warning-dark: '#FFD078'
  on-warning-dark: '#302000'
  warning-subtle-dark: '#4A3612'
  danger-dark: '#FF9CAC'
  on-danger-dark: '#3B0710'
  danger-subtle-dark: '#4B1D29'
  success-dark: '#77E4B7'
  on-success-dark: '#073022'
  success-subtle-dark: '#153D32'
  focus-dark: '#FFD078'
  scrim-dark: '#000000'
typography:
  display:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 56px
    fontWeight: '300'
    lineHeight: '1.04'
    letterSpacing: -0.045em
  display-mobile:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 38px
    fontWeight: '350'
    lineHeight: '1.08'
    letterSpacing: -0.035em
  heading-lg:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.18'
    letterSpacing: -0.025em
  heading-md:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.25'
    letterSpacing: -0.015em
  heading-sm:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 20px
    fontWeight: '650'
    lineHeight: '1.3'
  body-lg:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 18px
    fontWeight: '400'
    lineHeight: '1.55'
  body:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.55'
  body-sm:
    fontFamily: 'Inter, Segoe UI, system-ui, sans-serif'
    fontSize: 14px
    fontWeight: '450'
    lineHeight: '1.5'
  label:
    fontFamily: 'IBM Plex Mono, ui-monospace, monospace'
    fontSize: 12px
    fontWeight: '700'
    lineHeight: '1.35'
    letterSpacing: 0.07em
  caption:
    fontFamily: 'IBM Plex Mono, ui-monospace, monospace'
    fontSize: 11px
    fontWeight: '550'
    lineHeight: '1.45'
    letterSpacing: 0.02em
  data:
    fontFamily: 'IBM Plex Mono, ui-monospace, monospace'
    fontSize: 13px
    fontWeight: '650'
    lineHeight: '1.4'
    letterSpacing: 0em
rounded:
  none: 0px
  xs: 2px
  sm: 4px
  md: 8px
  lg: 12px
  full: 9999px
spacing:
  '1': 4px
  '2': 8px
  '3': 12px
  '4': 16px
  '5': 20px
  '6': 24px
  '8': 32px
  '10': 40px
  '12': 48px
  '16': 64px
  '20': 80px
  margin-mobile: 16px
  margin-tablet: 24px
  margin-desktop: 40px
  content-max: 1440px
  reading-max: 720px
components:
  app-navigation:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text-muted}'
    foreground-dark: '{colors.text-muted-dark}'
    active: '{colors.secondary}'
    active-dark: '{colors.secondary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  swipe-deck:
    background: '{colors.canvas}'
    background-dark: '{colors.canvas-dark}'
    active-outline: '{colors.border-strong}'
    active-outline-dark: '{colors.border-strong-dark}'
    next-outline: '{colors.border}'
    next-outline-dark: '{colors.border-dark}'
  public-header:
    background: '{colors.canvas}'
    background-dark: '{colors.canvas-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  public-footer:
    background: '{colors.canvas}'
    background-dark: '{colors.canvas-dark}'
    foreground: '{colors.text-muted}'
    foreground-dark: '{colors.text-muted-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  candidate-card:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
    radius: '{rounded.xs}'
  candidate-detail:
    background: '{colors.canvas}'
    background-dark: '{colors.canvas-dark}'
    panel: '{colors.surface}'
    panel-dark: '{colors.surface-dark}'
    divider: '{colors.border-strong}'
    divider-dark: '{colors.border-strong-dark}'
  portfolio-node-visualization:
    background: '{colors.surface-subtle}'
    background-dark: '{colors.surface-subtle-dark}'
    node-core: '{colors.secondary}'
    node-core-dark: '{colors.secondary-dark}'
    node-secondary: '{colors.primary}'
    node-secondary-dark: '{colors.primary-dark}'
    node-tertiary: '{colors.warning}'
    node-tertiary-dark: '{colors.warning-dark}'
    connector: '{colors.border-strong}'
    connector-dark: '{colors.border-strong-dark}'
  chart-data-table:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    rule: '{colors.border-strong}'
    rule-dark: '{colors.border-strong-dark}'
  photo-obscure:
    foreground: '{colors.primary}'
    foreground-dark: '{colors.primary-dark}'
    background: '{colors.primary-subtle}'
    background-dark: '{colors.primary-subtle-dark}'
    border: '{colors.primary}'
    border-dark: '{colors.primary-dark}'
  synthetic-data-label:
    foreground: '{colors.secondary}'
    foreground-dark: '{colors.secondary-dark}'
    background: '{colors.secondary-subtle}'
    background-dark: '{colors.secondary-subtle-dark}'
    border: '{colors.secondary}'
    border-dark: '{colors.secondary-dark}'
  freshness-indicator:
    current: '{colors.success}'
    current-dark: '{colors.success-dark}'
    pending: '{colors.primary}'
    pending-dark: '{colors.primary-dark}'
    stale: '{colors.warning}'
    stale-dark: '{colors.warning-dark}'
    unusable: '{colors.danger}'
    unusable-dark: '{colors.danger-dark}'
  account-inclusion-control:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    selected: '{colors.secondary}'
    selected-dark: '{colors.secondary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  disclosure-level-control:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    selected: '{colors.primary}'
    selected-dark: '{colors.primary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  preference-control:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    selected: '{colors.accent}'
    selected-dark: '{colors.accent-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  profile-field:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
    error: '{colors.danger}'
    error-dark: '{colors.danger-dark}'
  primary-action:
    background: '{colors.primary}'
    background-dark: '{colors.primary-dark}'
    foreground: '{colors.on-primary}'
    foreground-dark: '{colors.on-primary-dark}'
    radius: '{rounded.sm}'
  secondary-action:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
    radius: '{rounded.sm}'
  swipe-actions:
    pass: '{colors.text}'
    pass-dark: '{colors.text-dark}'
    interested: '{colors.secondary}'
    interested-dark: '{colors.secondary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  revalidation-notice:
    background: '{colors.primary-subtle}'
    background-dark: '{colors.primary-subtle-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.primary}'
    border-dark: '{colors.primary-dark}'
  recovery-panel:
    background: '{colors.warning-subtle}'
    background-dark: '{colors.warning-subtle-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.warning}'
    border-dark: '{colors.warning-dark}'
  mutual-match-reveal:
    background: '{colors.surface-emphasis}'
    background-dark: '{colors.surface-emphasis-dark}'
    accent: '{colors.accent}'
    accent-dark: '{colors.accent-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
  notification-control:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    selected: '{colors.secondary}'
    selected-dark: '{colors.secondary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  discovery-activity-notice:
    background: '{colors.surface-emphasis}'
    background-dark: '{colors.surface-emphasis-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    accent: '{colors.accent}'
    accent-dark: '{colors.accent-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  profile-preview-frame:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
    live-owner: '{colors.primary}'
    live-owner-dark: '{colors.primary-dark}'
  source-to-experience-trace:
    source: '{colors.primary}'
    source-dark: '{colors.primary-dark}'
    derived: '{colors.accent}'
    derived-dark: '{colors.accent-dark}'
    owner-only: '{colors.primary}'
    owner-only-dark: '{colors.primary-dark}'
    private-matching: '{colors.accent}'
    private-matching-dark: '{colors.accent-dark}'
    visible: '{colors.secondary}'
    visible-dark: '{colors.secondary-dark}'
    connector: '{colors.border-strong}'
    connector-dark: '{colors.border-strong-dark}'
  consent-panel:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  confirmation-dialog:
    background: '{colors.overlay}'
    background-dark: '{colors.overlay-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    danger: '{colors.danger}'
    danger-dark: '{colors.danger-dark}'
  language-and-theme-control:
    background: '{colors.surface}'
    background-dark: '{colors.surface-dark}'
    selected: '{colors.primary}'
    selected-dark: '{colors.primary-dark}'
    border: '{colors.border-strong}'
    border-dark: '{colors.border-strong-dark}'
  safety-notice:
    background: '{colors.warning-subtle}'
    background-dark: '{colors.warning-subtle-dark}'
    foreground: '{colors.text}'
    foreground-dark: '{colors.text-dark}'
    border: '{colors.warning}'
    border-dark: '{colors.warning-dark}'
  loading-skeleton:
    base: '{colors.surface-subtle}'
    base-dark: '{colors.surface-subtle-dark}'
    highlight: '{colors.surface-emphasis}'
    highlight-dark: '{colors.surface-emphasis-dark}'
---

# Findur — Design Spine

## Brand & Style

Findur feels like exploring a constellation of evidence: atmospheric, curious, exact, and recognizably about dating without borrowing luxury or trading-desk signals. Dark is the signature presentation, but light is a first-cut peer rather than a derivative theme. The composition stays clean and the information density restrained; depth is revealed progressively instead of compressed into one dashboard.

The current composition reference is [Constellation](.working/direction-constellation.html). It establishes the portfolio-node hero, dark-first atmosphere, crisp geometry, restrained glow, direct bilingual controls, and purposeful evidence-path motion. These spines win if that artifact and the written contract conflict.

Portfolio information leads; the obscured Photo and personal context remain present but subordinate until Mutual Match. Geometry uses straight rules, clipped or minimally softened corners, and clear alignment. Rounded containers are selective, never the default visual shorthand for friendliness.

Motion explains state: connections pulse to show relationships, a deck edge may shift to imply continuity, and the Mutual Match reveal resolves the previously obscured Photo. Motion never communicates unique information. Under reduced motion, every animation is replaced by the complete final state, persistent text, and the same available actions.

Interaction feedback preserves exact layout geometry. Hover and pressed states never change border width, translate, scale, or reflow an element. On hover-capable pointers, controls may gain a restrained tonal fill or inset line, links gain a clear underline, and portfolio nodes may intensify their existing glow with an inset ring. Pressed state deepens the inset treatment without movement. Focus-visible remains the strongest treatment through the dedicated focus ring; disabled controls do not react to hover. These rules apply consistently in both themes.

### Interaction State Contract (Normative)

This table is the single visual source of truth for interactive states across every route and viewport. Implementation uses shared primitives for these roles; it must not define per-screen hover behavior. The `.working/` previews are illustrations only. If a preview conflicts with this table, this table wins.

| Role | Default | Hover on capable pointers | Focus-visible | Pressed/active | Disabled or passive |
|---|---|---|---|---|---|
| Primary navigation item | Muted text; current item has stronger text plus a persistent geometric marker | Text strengthens; 7% secondary tonal field; 3px inset marker on the navigation edge; no underline | 3px warning focus ring with 3px offset; current marker remains | Inset edge marker deepens to 4px; no movement | Navigation has no disabled state; unavailable destinations use a named gate after navigation rather than a dead item |
| Primary action | Solid primary fill and high-contrast label | Stable 1–2px inset highlight; fill may shift tonally | 3px warning focus ring with 3px offset | 2px inset emphasis; no movement | Muted text plus explicit subdued fill/border; no hover response; prerequisite is explained nearby |
| Secondary or destructive action | Surface/transparent fill with fixed 1px semantic border | Stable 1px inset emphasis and restrained tonal fill | 3px warning focus ring with 3px offset | 2px inset emphasis; no movement | Same disabled treatment as Primary action; never opacity alone |
| Inline or utility link | Semantic link color; underline where needed for recognition | 2px underline with 3–4px offset | 3px warning focus ring with 3px offset | Color/inset emphasis only; no movement | Omit the link or expose an explained disabled control instead |
| Selectable row | Fixed 1px boundary, native control, label, and state text | 4px inset secondary edge plus 7% secondary tonal field | Focus ring encloses the whole row through `:focus-within` | 2px inset emphasis; no movement | Named unavailable reason, muted/hatched field, no pointer cursor or hover response |
| Interactive portfolio node | Fixed 1px outline, existing restrained glow, programmatic label | Existing glow intensifies and a 2px inset ring appears; the same detail is available through focus and tap | 3px warning focus ring; detail opens equivalently | 3px inset ring; no scale or translation | Unsupported nodes are omitted or labelled unavailable; never decorative-but-clickable |
| Passive data surface | Static surface and fixed boundaries | No response | Not placed in the tab order unless it owns an operable control or disclosure | No response | Remains static |

All values above are semantic: secondary, warning, surface, text, and muted resolve through the active light/dark theme tokens. Hover is absent on coarse/touch-only pointers. Focus-visible is never replaced by hover styling.

Findur's installable presentation uses one geometric constellation mark derived from the wordmark—not a portfolio amount, Photo, or candidate data—across the branded Apple touch icon, cross-platform web-app icons, browser favicon, and standalone launch presentation. Light and dark browser chrome use the canvas tokens. Launch and task-switcher surfaces use an opaque brand field with no sensitive content. Architecture/build owns the exact platform asset matrix and metadata implementation.

## Colors

All semantic roles have concrete light and dark values. Components switch token pairs as a unit; content is not duplicated per theme.

- `{colors.canvas}` / `{colors.canvas-dark}` are the page field. Dark is the initial visual signature; light retains the cool atmospheric cast rather than becoming neutral white.
- `{colors.surface}` / `{colors.surface-dark}` hold task content. `{colors.surface-subtle}` and `{colors.surface-emphasis}` create depth without relying on shadows.
- `{colors.text}` and `{colors.text-dark}` carry primary text; muted and disabled pairs are reserved for secondary or unavailable content, never required instructions.
- `{colors.primary}` marks navigation, disclosure inspection, and source evidence. `{colors.secondary}` marks selected interest, selected account coverage, and live relational nodes. `{colors.accent}` is a sparse comparative/Mutual Match accent, not generic decoration.
- Success, warning, and danger pairs always appear with a shape, icon, or explicit label. Red and green never encode performance or human value.
- Focus uses `{colors.focus}` in light and `{colors.focus-dark}` in dark as a 3px outer ring with 2px offset.

Contrast targets are WCAG 2.2 AA: normal text at least 4.5:1, large text at least 3:1, and focus indicators, controls, chart marks, and meaningful graphical objects at least 3:1 against adjacent colors. Primary action text targets at least 4.5:1. Theme QA verifies every load-bearing pairing, including text/canvas, text/surface, muted text/surface, action fill/text, focus/surface, every Freshness State, and chart nodes/connectors. Synthetic, freshness, warning, and error meanings never depend on color alone.

`{colors.border-strong}` against `{colors.surface}` is 4.33:1 and `{colors.border-strong-dark}` against `{colors.surface-dark}` is 4.47:1; these verified pairs define every essential component boundary, control outline, table rule, and chart mark. `{colors.border}` and `{colors.border-dark}` are reserved for decorative separators, atmospheric grids, and the nonessential next-card outline only. `{colors.secondary}` on `{colors.secondary-subtle}` is 4.86:1 and remains suitable for normal-size Synthetic Data Label text.

## Typography

Inter/Segoe UI/system sans carries names, prose, and human context; IBM Plex Mono/system mono carries evidence labels, timestamps, values, and state readouts. The contrast makes portfolio evidence legible without turning the entire product into a terminal.

Use `display` only for a Public Site proposition or a Discovery orientation line; use `display-mobile` below 768px. Candidate names use `heading-sm`; route headings use `heading-lg` or `heading-md`. Body copy never drops below `body-sm`; `caption` is reserved for secondary metadata, not consent or error text. Data aligns by tabular figures where available.

English and French share the same type roles. Layouts allow at least 35% label expansion and wrap without truncating actions, state labels, table headings, or navigation. Do not force uppercase transformations on localized prose. Locale controls date, time, number, percentage, distance, and currency formatting; values and units remain programmatically associated.

## Layout & Spacing

The spacing rhythm is 4px-based, with 16px mobile margins, 24px tablet margins, and 40px desktop margins. Content is capped at `{spacing.content-max}`; long public/legal copy uses `{spacing.reading-max}`.

Below 768px, primary areas use a bottom `App Navigation`; Candidate Detail becomes a full-screen route, the portfolio visualization stacks above its evidence summary, and persistent `Swipe Actions` remain clear of browser chrome, device safe areas, and zoomed text. In standalone display mode, the app shell extends to safe-area edges while navigation and actions pad inward. From 768px to 1023px, content remains primarily single-column with room for a secondary panel. At 1024px and above, Candidate Detail uses a two-column composition: identity/portfolio visualization and permitted data/evidence. Required content and actions stay equivalent across sizes.

Discovery favors one focal Candidate Card rather than a dense grid. Portfolio may use bounded tables after a visual summary; Profile uses a reading-width form. At 200% zoom, desktop can reflow to the tablet composition without loss. At 400% zoom at a 1280px CSS viewport, content resolves to a single readable column with no two-dimensional page scrolling except within intrinsically wide data tables, which also have a non-table summary.

## Elevation & Depth

Depth comes from tonal layers, one-pixel rules, overlap, and restrained node glow. Candidate Card deck depth may show one offset outline behind the active card. Shadows are not a hierarchy system and must not mimic premium/luxury product photography. Overlays use the relevant overlay token above a scrim; background interaction is inert while open.

Atmospheric points or faint grids may appear on large canvas areas at very low contrast. They disappear behind reading surfaces and never reduce text or focus contrast. Light theme uses fewer atmospheric marks than dark theme.

## Shapes

Default surfaces use square corners or `{rounded.xs}`. Form controls and actions may use `{rounded.sm}`; dialogs and bounded preview frames may use `{rounded.md}`. `{rounded.full}` is limited to circular portfolio nodes and never used for generic pills.

Large branded surfaces may use a single clipped 8–22px corner, always with a rectangular layout and focus hit area. Do not clip text, native focus rings, or scroll containers. Photo masks may use a geometric hexagon before Mutual Match; the revealed Photo uses the same reserved footprint so layout does not jump.

## Components

The names below are the canonical component inventory shared with `EXPERIENCE.md`.

| Component | Visual specification |
|---|---|
| `App Navigation` | Mobile: fixed bottom rule, three equal destinations, 48px minimum height; desktop: 224–256px rail. Active state uses secondary color, heavier text, and a geometric marker—not color alone. |
| `Swipe Deck` | One centered active Candidate Card with a strong outline and at most one offset next-card outline using the weak decorative border. Deck position uses text; motion never supplies order or availability. Empty/recovery states occupy the same focal region without a ghost card. |
| `Public Header` | Canvas-level header with wordmark, public routes, entry action, and locale/theme access. One-pixel lower rule; collapses to a labeled menu without hiding legal routes. |
| `Public Footer` | Canvas-level footer with strong top rule, wrapping public/legal/support links, demonstration boundary, and language/theme access where the header is collapsed. No financial or candidate imagery. |
| `Candidate Card` | Focal surface with 1px strong border, one clipped corner, identity strip, dominant visualization, evidence, freshness, and actions. Initial signal hierarchy is fixed: Snapshot shows allocation/asset-class mix from supported positions and cash balances, diversification context, value/activity bands, activity recency, coarse Proximity, source coverage, and Freshness State; Holdings may add supplied instrument names/symbols and kinds, position weights, defensible percentage performance, and activity categories; Full Detail may add supplied quantities, last-known prices, cost basis, exact values by currency, portfolio/account value, defensible performance amounts, and distinctly labelled recent orders or historical activities. Performance is absent or explicitly unavailable when its period, inputs, currency treatment, coverage, or freshness cannot be defended. Source facts and derived signals use explicit labels. The relevance panel names proximity, the selected similar/diversified/complementary preference, one or two derived traits, Disclosure Level, and freshness—never a score, guarantee, cohort, formula, or worth judgment. Photo area remains visually subordinate. |
| `Candidate Detail` | Full canvas on mobile; two-column surface at desktop. Uses rules and tonal panels, not nested rounded cards. Persistent action region visually separated from scrollable detail. |
| `Portfolio Node Visualization` | Labeled circular nodes sized by composition with straight connectors, faint grid, and exact text/legend equivalence. Node color is repeated by label, position, and connector pattern. No red/green performance encoding. |
| `Chart Data Table` | Mono-aligned values, sticky header only when scrolling, 1px strong row rules where boundaries are essential, explicit units and periods. Horizontal overflow is allowed only inside its region; a visible summary precedes it. |
| `Photo Obscure` | Geometric mask with repeated-line fill, explicit “Photo hidden until Mutual Match” text, and no silhouette that implies identity. Revealed state removes the pattern without resizing. |
| `Synthetic Data Label` | Persistent outlined label using text plus diamond marker. It remains adjacent to candidate identity in every responsive state and every theme. |
| `Freshness Indicator` | Text label, timestamp, and distinct icon/shape for connected, syncing, stale, needs reauthorization, failed, or disconnected. Never a lone colored dot. |
| `Account Inclusion Control` | Account rows with provider/account label, selection control, availability state, and summary “Using 2 of 4 connected accounts.” Selected rows use secondary outline plus checked state. “Select all” is a normal labeled control. |
| `Disclosure Level Control` | Three equal options—Snapshot, Holdings, Full Detail—with selected outline/underline and nearby field-impact preview. Initially none is selected or saved; Snapshot carries a neutral “Recommended starting point” annotation without being preselected. Selection changes the preview only. A persistent summary distinguishes “Saved level” from “Previewing — not saved,” followed by explicit Save disclosure level and Cancel actions. Unsaved route exit uses a standard discard/stay confirmation. Never render tiers as prestige badges or ascending status. |
| `Preference Control` | Maximum-distance input and similar/diversified/complementary choices. Selected option uses accent outline and explicit selected label; no ranking score. |
| `Profile Field` | Persistent label, requirement/visibility annotation, input, help/error line, and at least 44px control height. Error uses danger border, icon, and adjacent text. |
| `Primary Action` | Solid primary fill, high-contrast text, 44px minimum height, crisp 4px corner. Disabled state retains readable label and does not use opacity alone. |
| `Secondary Action` | Surface fill, strong 1px outline, 44px minimum height. Hover uses a stable inset line or tonal fill, focus uses the dedicated external ring, and pressed uses a stronger inset line; geometry never changes. |
| `Swipe Actions` | Paired Pass and Interested controls, each at least 48px high on touch surfaces. Interested uses secondary; Pass stays neutral. Both show text and directional icon. |
| `Revalidation Notice` | Compact primary-subtle strip/panel with freshness timestamp, “checking for updates” language, and nonblocking visual activity. Reduced motion uses a static progress/state icon. |
| `Recovery Panel` | Warning-subtle bounded region with named problem, what remains trustworthy, and one primary recovery path plus safe navigation. Never an empty chart placeholder. |
| `Mutual Match Reveal` | Emphasis surface that swaps Photo Obscure for the Photo, uses a restrained accent constellation resolve, retains Synthetic Data Label, and presents Continue Discovery as the sole forward action. No chat affordance. |
| `Notification Control` | Profile setting with separate Mutual Match, Incoming Interest, and connection-action-required choices plus permission/support state. The contextual pre-prompt reuses this language after the first settled Interested decision. It never imitates a browser permission dialog or blocks Discovery. |
| `Discovery Activity Notice` | Discovery child surface for an unacknowledged Incoming Interest or Mutual Match. Uses accent rule/constellation marker, event type, privacy-safe status, and one Open action. It is not a fourth navigation item, chat affordance, or permanent match-history list. |
| `Profile Preview Frame` | Owner-only bounded preview with live-owner label, viewport selector, Disclosure Level/state selectors, and the same card visual language. Phone and desktop selectors are neutral tools, not decorative device chrome. |
| `Source-to-Experience Trace` | Included source data branches into three separately labelled paths: owner-only exact view, private matching input, and candidate-visible disclosure. Each branch carries live-owner or synthetic provenance; only the candidate-visible branch shows Disclosure Level filtering. Missing, blocked, or unsupported stages use a dashed strong outline and explicit text, never inferred values. |
| `Consent Panel` | Reading-width surface with data categories, purpose, private use, possible visibility, limitations, and disconnect consequence before the action. Important statements use body text, not caption. |
| `Confirmation Dialog` | One modal level, clear consequence summary, neutral cancel, danger-styled destructive confirmation, and initial focus on the heading or safe action. No decorative urgency. |
| `Language and Theme Control` | Labeled EN/FR and System/Light/Dark selections with visible current values. Fits wrapping French labels and exposes focus on each option. |
| `Safety Notice` | Warning-subtle surface with 18+, read-only/non-advisory, anti-solicitation, and money-request guidance as context requires. Uses a heading/icon/text combination; never a tiny legal footnote. |
| `Loading Skeleton` | Layout-matched blocks with subtle tonal difference. A short, nonessential shimmer is allowed; reduced motion is static. Labels outside the skeleton state what is loading. |

## Do's and Don'ts

| Do | Don't |
|---|---|
| Let the Portfolio Node Visualization dominate Candidate Card composition. | Lead with the Photo or reduce the portfolio to a badge. |
| Use crisp rules, aligned evidence, selective clipped corners, and breathing room. | Wrap every section in a rounded card or stack decorative shadows. |
| Repeat chart meaning with labels, values, patterns, summaries, and tables. | Depend on hue, node size, animation, or hover alone. |
| Qualify source coverage and Freshness State next to time-sensitive claims. | Say “real time,” “complete portfolio,” or imply a webhook proves change. |
| Keep dark and light semantic roles equivalent and verify both. | Treat light, French, reduced motion, zoom, or installed display as later polish. |
| Use motion to reveal relationships or resolve a state, with static equivalence. | Use ambient motion that distracts from consent, data, or actions. |
| Show Synthetic Data Label and owner/private labels persistently. | Allow responsive layout or detail expansion to drop provenance. |
| Use calm compatibility language and comparison without false precision. | Use wealth/status cues, red/green performance theater, or compatibility scores. |
| Keep account inclusion, OAuth access, and Disclosure Level visually distinct. | Combine separate consent layers into one toggle or badge. |
| Use the constellation mark and opaque brand canvas in icons and launch surfaces. | Put financial values, Photos, candidate identity, or private state into icons, launch images, or preview imagery. |
