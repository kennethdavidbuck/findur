# UX, Responsive, and Product-Completeness Review

## Verdict

**Needs targeted refinement before UX design.** The PRD gives UX an unusually strong product frame—portfolio-first discovery, the owner-only/live versus candidate/synthetic boundary, a complete owner preview, no chat, and responsive parity. However, several trust-critical and interaction-critical rules remain principles rather than screen-state acceptance criteria. Resolving the findings below will let the UX phase make design choices without having to make product-policy choices.

## Findings

### High — Synthetic provenance is not a persistent, testable discovery boundary

**Location:** [PRD FR-9](prd.md#L245) consequences, [PRD FR-25–26](prd.md#L368), [addendum §2.4–2.5](addendum.md#L99).

**Issue:** Discovery cards can show exact-looking holdings, values, performance, orders, and transactions, while FR-9 only requires them to be identified “visibly or contextually” where a reviewer *could otherwise* be confused. That leaves UX to decide whether the distinction is a one-time landing disclaimer, an intermittent label, or a persistent card/deck cue. The comprehension list states the desired understanding but neither PRD nor metric makes the placement or successful recovery of that distinction testable.

**Why it matters:** A visitor can reasonably mistake a polished synthetic card for a real person’s connected financial disclosure, undermining the central demonstration boundary and its trust narrative.

**Fix:** Add an explicit provenance pattern and acceptance criterion: a persistent, plain-language synthetic-candidate marker in Discovery (with a first-entry explanation), distinct owner-only/live markers in Portfolio Showcase and Profile Preview, and an accessible equivalent. Test that participants can identify the provenance of a card after navigating away from the landing page or dismissing onboarding.

### High — Disclosure is the product’s central control, but its minimum policy matrix is deferred entirely to UX

**Location:** [PRD FR-6](prd.md#L209), especially lines 215–220; [PRD open question 1](prd.md#L540); [addendum §1.2](addendum.md#L14) and [§2.1](addendum.md#L59).

**Issue:** The PRD specifies categories that *may* be hidden, bucketed, derived, or exact, but supplies no prototype Disclosure Level names, count, baseline bundles, or decision rules. It also makes lower-disclosure reciprocity and higher-to-lower initiation policy requirements. These are product/privacy rules that directly determine card content, comparison, preview states, and comprehension testing—not merely labels or layouts.

**Why it matters:** UX cannot produce a faithful complete owner preview or a credible responsive Candidate Card without inventing disclosure policy. Different reasonable designs would create materially different privacy promises and invalidate the proposed metrics.

**Fix:** Before the UX phase, product should provide a small provisional disclosure matrix: each level, every data category, pre-match versus post-match treatment, reciprocity/initiation rule, and the no-data fallback. UX can then name, explain, compose, and validate the levels; a later decision can revise the matrix without leaving the current prototype untestable.

### High — A disclosure downgrade or disconnect has no required treatment for information already rendered

**Location:** [PRD FR-4](prd.md#L113), [FR-6](prd.md#L215), and [FR-13](prd.md#L304); [addendum §1.2](addendum.md#L23) and [§3.3](addendum.md#L166).

**Issue:** FR-6 only suppresses no-longer-permitted information in “subsequent views.” The addendum correctly asks what happens to rendered or cached information after downgrade, disconnect, block, or report, but leaves it unanswered. Architecture says stale cached presentation must not survive a consent or safety change, yet the PRD gives neither the user-visible behavior nor a test for the open card, preview, browser history/back navigation, or match surface.

**Why it matters:** A user can lower disclosure or withdraw consent while prohibited details remain visible in an active or restored UI, making the privacy control appear ineffective and creating an implementation gap between client cache and server state.

**Fix:** Define the transition contract: invalidate and replace affected current views immediately, prevent restoration from cache/history, state which details are removed versus retained for the owner, and provide a clear post-action confirmation. Add scenarios for downgrade and disconnect while a rich card/preview is open and during an in-progress refresh.

### Medium — Mobile has parity language but no interaction contract for a rich, explorable swipe card

**Location:** [PRD UJ-3](prd.md#L52), [FR-11–12](prd.md#L282), and [NFR-9–11](prd.md#L494); [addendum §2.1–2.3](addendum.md#L59).

**Issue:** The card must be portfolio-first, allow rich/possibly expandable charts and timelines, support a single swipe decision, and work on phone and desktop. The addendum asks UX to choose card proportions, breakpoints, visualization fallbacks, and “lightweight exploration,” but no requirement establishes how scrolling, chart gestures, expansion, swipe actions, undo/confirmation (if any), keyboard controls, and the decision point coexist on a small viewport.

**Why it matters:** A phone implementation can technically include every field while making the decision action unreachable, accidentally swiping during exploration, or giving desktop materially safer/clearer control than mobile.

**Fix:** Add responsive acceptance scenarios rather than a prescribed layout: on a representative phone viewport, users can inspect every permitted visualization and its freshness/provenance, then intentionally take exactly one left/right action without gesture collision; keyboard and explicit button equivalents reach the same outcome. Require documented fallbacks for dense charts and activity timelines.

### Medium — The minimum profile omits completion, correction, and ineligibility states needed to reach Discovery safely

**Location:** [PRD UJ-7](prd.md#L60), [FR-23](prd.md#L173), [FR-13](prd.md#L304), and [MVP scope](prd.md#L411); [addendum §2.1](addendum.md#L59).

**Issue:** FR-23 lists profile categories and says UX determines validation, but does not specify behavior when required fields are absent, a photo fails to upload/process, location or preference is unavailable, a user saves a draft, or edited information makes the profile ineligible. The key journey assumes connection first, profile completion next, then preferences/Discovery, without defining those gates or recovery paths.

**Why it matters:** UX may either block too late (after sensitive connection work), allow a card with misleadingly blank human context, or build incompatible flows for owner preview and synthetic schema validation.

**Fix:** Establish a concise profile-state model: draft, incomplete, complete/eligible, and blocked/error; identify the minimum gate for preview and for Discovery; and require field-level recovery plus a non-destructive save/return path. Keep exact input design with UX, but make the journey and resulting card state testable.

### Medium — Freshness and unavailable-data rules do not fully transfer from the live showcase to mixed-source candidate cards

**Location:** [PRD FR-3](prd.md#L124), [FR-11](prd.md#L282), [FR-19](prd.md#L136), and [FR-22](prd.md#L267); [addendum §2.3](addendum.md#L86) and [§3.3](addendum.md#L166).

**Issue:** The Portfolio Showcase explicitly distinguishes loaded, unavailable, empty, stale, failed, and unsupported per dataset. Candidate Cards merely show “relevant Freshness State,” despite rich cards potentially combining holdings, performance context, activity, and orders whose source ages and coverage differ. The addendum recognizes that these sources have different recency characteristics, but does not define the card’s aggregation/disclosure rule.

**Why it matters:** A card can imply that all visualized information is equally current or complete; a responsive layout may also hide the only freshness/coverage qualifier when a chart is collapsed.

**Fix:** Define a small candidate-card state matrix for each rendered visualization: source/coverage cue, timestamp or qualified freshness wording, and treatment for unavailable/empty/stale/unsupported data. Require the cue and fallback to survive phone/desktop composition changes and screen-reader rendering; cover mixed-freshness fixture scenarios in formative and scripted review.

## Severity Summary

| Severity | Count |
| --- | ---: |
| Critical | 0 |
| High | 3 |
| Medium | 3 |
| Low | 0 |

## Scope Notes

The public landing/pitch, persistent legal/trust routes, owner-only live showcase, minimum personal profile, complete preview, one-swipe/photo-reveal rule, chat exclusion, broad failure/empty states, and phone/desktop parity are all present. The findings above address where those commitments still need a concrete UX-state or acceptance contract. The PRD is appropriately non-prescriptive about final visual layout and brand expression; it is too vague only where a downstream UX choice would itself establish privacy policy, provenance, or a required interaction/recovery behavior.
