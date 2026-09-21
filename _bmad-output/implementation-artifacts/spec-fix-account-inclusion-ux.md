---
title: 'Correct Story 1.1 Account Selection UX and Flow'
type: 'bugfix'
created: '2026-09-20'
status: 'done'
route: 'dispatch'
baseline_commit: '278582cbbb3641c04992c9fb6eede8022f22139d'
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/spec-1-1-select-and-confirm-included-accounts.md'
  - '_bmad-output/implementation-artifacts/epic-1-context.md'
  - '_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/.working/key-account-selection.html'
  - '_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/.working/key-portfolio-showcase.html'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 1.1's deployed Portfolio page does not read as the authored **Choose accounts** step after OAuth. It repeats the same accounts in an inventory section and a selection section, exposes internal committed/draft/retry machinery as permanent product content, uses a busy cursor for a merely disabled action, provides weak success feedback, uses alarming implementation-level removal copy, and diverges materially from the approved composition, density, hierarchy, and sizing.

**Approach:** Restore Story 1.1 as the focused post-OAuth **Choose what Findur may use.** step for configuring the portfolio portion of the user's persistent dating profile. Retain the authored visual language and single-list composition while replacing its bank/compliance framing with concise, warm dating-profile onboarding language. Render one compact integrated selection list, show selected versus saved coverage only as needed, keep internal alternate/recovery examples out of the normal successful screen, provide unmistakable save feedback, and use one short, friendly review modal before saving. The saved profile choice persists across every later login and can later be reopened deliberately from Story 1.2's Portfolio Showcase through **Edit included accounts**; the selector is not the permanent Portfolio dashboard and is never a per-session choice.

**Boundaries:**
- Always: preserve the existing inclusion API and backend save/retrieval behavior; preserve owner isolation, idempotency, concurrency, accessibility, and EN/FR; present one clear, friendly post-OAuth dating-profile setup task; keep the review modal compact and reassuring.
- Never: render the inventory and selector as duplicate live sections; render the design board's alternate/reference column as simultaneous product UI; imply a disabled action is busy; expose terms such as masked inventory, connection inventory, committed coverage, inclusion readiness, confirmation scope, permitted private purposes, sync mode, purge, recalculation, or internal lifecycle state in user-facing copy; use destructive styling or a wall of text in the modal; claim the selector is the normal completed Portfolio view.
- Not in this correction: changing account persistence/retrieval/deletion semantics; implementing Story 1.2's Showcase; redesigning the entire authenticated shell; changing the broader onboarding requirements.

</frozen-after-approval>

## Acceptance Criteria

1. Ready inventory renders exactly one account list. Each row contains its selection control, masked identity, relevant availability context, and concise availability state. The checkbox is the only row-level draft indicator.
2. The screen is recognizably a dating-profile onboarding step through the authored **Choose what Findur may use.** heading (FR: **Choisissez ce que Findur peut utiliser.**), a short explanation that the choice can be changed later, a lightweight progress context, and the authored visual composition. It does not look like a bank consent workflow or a generic permanent Portfolio dashboard.
3. Selected and saved coverage are distinct when the draft differs but use ordinary language and remain concise; neither summary repeats all selected account names.
4. A disabled Review action uses a normal unavailable cursor. A busy state appears only while the save request is actually in flight.
5. Successful save produces an unambiguous localized **Account choices saved** status, updates saved coverage, and presents the setup continuation without technical processing prose.
6. Failed or pending additions are represented accurately with no redundant pair of recovery buttons and without presenting a failed draft account as saved.
7. Deselection is presented as an ordinary reversible profile edit. The review contains no deletion, purge, retrieval, irreversibility, internal-system consequence, or danger-styled language.
8. The surface follows the authored UX's main-column composition, density, typography hierarchy, row sizing, and responsive behavior. The artifact's right-side alternate/reference states are not rendered as simultaneous sections.
9. Frontend tests cover the single list, disabled state, draft/saved coverage, success, removal confirmation, pending, failed, EN/FR-sensitive copy, and absence of duplicate/internal presentation.
10. The saved selection is durable configuration: reauthorization, temporary unavailability, staleness, refresh, and ordinary later login preserve it. Newly discovered accounts begin unselected. Only an explicit selection edit or full disconnect/deletion removes a saved choice.
11. Copy and interaction framing make clear that the user is configuring the portfolio portion of a persistent Findur dating profile, not choosing temporary data for the current browser session.
12. The review modal contains only a friendly title, resulting selected count, changed account names grouped as added/removed when applicable, a short reminder that the choice can be changed anytime, and Back/Save actions. One concise inline/live-region message reports progress, success, or failure. There is no provider-access essay, data-category inventory, private-purpose list, lifecycle explanation, or technical wall of text.
13. Accounts whose usability reason is `account_closed` are omitted from the profile chooser and its counts. They are not presented as disabled profile options. Temporary availability problems remain distinct and are surfaced only when the user needs an actionable explanation, especially for a previously saved choice.
14. Initial post-OAuth setup uses the authored onboarding navigation context: on desktop the setup progress rail replaces the ordinary Discovery/Portfolio/Profile left rail and shows the complete friendly journey **Connect → Choose accounts → Portfolio showcase → Preferences & privacy → Preview & discover**, with Choose accounts current. On mobile the compact local step treatment is **Connect → Choose accounts → Portfolio**, and the bottom app navigation may remain available with normal prerequisite gates. Later **Edit included accounts** reuses the account control inside the ordinary app shell and does not replay onboarding progress.
15. The active initial step is grounded in durable state, not a decorative counter: successful OAuth/inventory plus absent committed account choices produces **Choose accounts**; a successful committed save navigates to `/portfolio`. Pending or failed saves remain on the setup step with concise feedback.

## Implementation Tasks

- [x] Rework `PortfolioPage.tsx` ready state into one integrated selection panel and keep raw inventory cards only for non-ready/recovery states where needed.
- [x] Align `styles.css` with the authored account-selection main column without broad shared-shell regressions.
- [x] Replace implementation-level `i18n.tsx` copy with concise, warm profile-setup language and clear success/status text.
- [x] Reduce the account-choice confirmation modal in `PortfolioPage.tsx` to the compact friendly review defined above.
- [x] Update `App.test.tsx` to prove every acceptance state and remove assertions that require duplicated content or implementation-heavy modal presentation.
- [x] Verify all frontend checks plus the repository-mandated backend lint command; inspect the final diff against this spec.

## Implementation Notes

- Story 1.1 setup is isolated to `/onboarding/accounts`; legacy `/connect/result` immediately replace-routes there for compatibility. The ordinary `/portfolio` destination remains the Story 1.2 handoff.
- A committed initial save announces success and navigates to `/portfolio`; pending and failed saves remain in setup.
- The desktop setup rail shows the full five-step onboarding journey while the compact local 1–2–3 treatment remains available for narrow layouts.
- The existing inclusion API and backend behavior are unchanged.

## Spec Change Log

- 2026-09-21: Human review clarified that account choice is persistent dating-profile setup, not a permanent Portfolio dashboard or per-login choice; replaced technical/banking copy with friendly product language.
- 2026-09-21: Human review retained a compact modal but removed destructive styling, technical consequence prose, and the original wall of text.
- 2026-09-21: Human review required closed accounts to be omitted, initial save to advance to the Showcase, later editing to return there, and the full onboarding progress to be explicit.

## Review Triage Log

| Finding | Verdict | Evidence / route |
|---|---|---|
| B1 — `/onboarding/accounts` replays setup after reauthorization even when a committed selection exists | valid — patch | `PortfolioPage` loads committed inclusion but never completes an already-finished initial step; route from durable inclusion state. |
| B2 — `/portfolio` is not yet the full Showcase | false positive — Story 1.2 | This correction explicitly hands the destination to Story 1.2 and uses an honest interim continuation, without claiming the Showcase is implemented here. |
| B3 — later `/portfolio/accounts` editing is not implemented | false positive — Story 1.2 | AC14 defines the later integration contract; implementing the Showcase child route remains outside this correction. |
| B4 — successful save feedback lasts only 600 ms | valid — patch | The local timer unmounts the only status message; carry a durable completion notice onto the destination instead. |
| B5 — a user can edit the draft during the completion timer and lose that change | valid — patch | The 600 ms timer leaves controls interactive; remove the timer and complete atomically. |
| B6 — browser Back can reopen `/onboarding/accounts` after save | valid — patch | Completion currently uses `pushState`; replace the setup history entry. |
| B7 — conflict/stale saves leave an unrecoverable stale version | valid — patch | Save catch only sets a generic error; reload durable inclusion before offering another review. |
| B8 — empty inventory has no recovery action | valid — patch | `empty` maps to no recovery control; provide a clear reconnect/manage-connection path. |
| B9 — all filtered accounts produce an empty fieldset and unusable select-all | valid — patch | Ready state can contain only closed accounts; render an explicit no-available-accounts state and no selection controls. |
| B10 — a saved temporarily unavailable row is visually disabled while its checkbox remains enabled | valid — patch | Row class uses `account.selectable`, but input also permits committed accounts; style from the actual disabled condition. |
| B11 — nested live regions can double-announce save state | valid — patch | Feedback wrapper has `aria-live` and children also have status/alert roles; retain one announcement surface. |
| B12 — “Connection complete” appears during disabled/unauthorized recovery | valid — patch | Setup eyebrow/progress are unconditional; recovery must not claim the connection step is complete. |
| B13 — counts say “connected accounts” after closed/hidden accounts are removed | valid — patch | Denominator is the visible chooser list, not all connected accounts; label it as available accounts. |
| B14 — raw provider `account.type` leaks into the localized UI | valid — patch | The row prints an unlocalized provider value; show the localized category only. |
| B15 — the setup wordmark links around the prerequisite to `/portfolio` | valid — patch | `AuthenticatedLayout` always links the brand to Portfolio; make it non-navigating while setup owns the route. |
| B16 — routing, save-reconciliation, zero-state, and accessibility edges lack regression coverage | valid — patch | Existing tests cover the happy path but not the reviewed edge cases; add focused tests. |
| V1 — durable committed-state redirect is untested | valid — patch | Same gap as B1; test that a committed initial choice replaces `/onboarding/accounts`. |
| V2 — POST responses with pending/failed status are insufficiently covered | valid — patch | Add explicit assertions that these statuses remain on setup with accurate feedback. |
| V3 — request failure after Save is untested | valid — patch | Add a reconciliation test for an ambiguous failed response and a true unchanged failure. |
| V4 — setup suppression of ordinary app navigation is not asserted | valid — patch | Add a direct navigation-absence assertion for initial setup. |
| E1 — a lost response after a committed server save can be shown as failure and retried | valid — patch | Catch path does not reconcile with GET; reload and treat a matching committed selection as success. |
| E2 — completion timer races with further edits | valid — patch | Duplicate of B5, independently confirmed; remove the timer. |
| E3 — every account can be hidden while inventory is `ready` | valid — patch | Duplicate of B9, independently confirmed; supply a useful zero state. |
| E4 — duplicate masked labels collide as modal list keys | valid — patch | `ChangeList` keys by display label; key changes by immutable account id. |
| E5 — reauthorization replays a completed chooser | valid — patch | Duplicate of B1, independently confirmed; redirect from durable committed state. |
| E6 — unsaved temporarily unavailable accounts are silently omitted | valid — patch | Filtering hides every unsaved temporary account; retain actionable temporary rows while continuing to omit closed accounts. |
| R1 — completed users with non-ready inventory replay onboarding | medium | Inclusion is loaded only after `inventory.state === 'ready'`, so durable committed choices cannot redirect during recovery states. Route: patch by loading inclusion independently. |
| R2 — direct `/portfolio` can describe unsaved choices as saved | false | The frozen scope assigns the normal Portfolio destination and prerequisite integration to Story 1.2; this correction owns the OAuth-return setup route, whose navigation is suppressed until completion. Existing B2 already records the interim Portfolio placeholder as the Story 1.2 handoff. |
| R3 — Select All remains operable when no row is selectable | medium | The fieldset is disabled only while saving and the aggregate checkbox has no `selectable.length === 0` guard, so it can be activated without changing state. Route: patch. |
| R4 — counts call every visible row an “available account” | low | Unsupported or temporarily unavailable rows are included in the denominator, so the label overstates selectable availability. Route: patch the ordinary-language denominator label. |
| R5 — pending additions appear cancellable even though an unchecked draft cannot supersede them | low | The pending message remains visible and no save occurs when the draft equals committed state; changing this safely would require a new cancel/check-status interaction not settled by this correction. Rejected as a rare edge whose fix adds lifecycle complexity. |
| R6 — the just-saved announcement reappears after leaving and returning to Portfolio | low | `portfolioSaved` survives route changes inside `ProtectedApp`, and the current test explicitly demonstrates the stale remount. Route: patch by consuming the transient notice after navigation away. |
| R7 — stale transient failure feedback survives a new draft | low | `saveFailed` is cleared only on another save, not when the user changes choices. Route: patch by clearing transient feedback on draft edits; durable change feedback remains authoritative. |
| R8 — the focused chooser heading suppresses its visible focus cue | medium | `.portfolio-inventory h1:focus-visible` overrides the global accessible outline even though route entry programmatically focuses that heading. Route: patch by removing the override. |
| R9 — the desktop setup grid clips between 56rem and 68.4rem | false | The container width is `min(100%, 64.1875rem)` and its second track is `minmax(0, 47.5rem)`, so that track shrinks inside the padded content box instead of forcing the claimed fixed width. |
| R10 — duplicate masked labels are ambiguous in the review modal | medium | Chooser rows include brokerage context but `ChangeList` emits only `maskedLabel`; two providers can yield identical masked labels at confirmation. Route: patch by retaining compact brokerage context. |
| R11 — unsaved temporary failures are shown without an actionable reason to keep them visible | medium | The initial chooser retains every non-closed temporary row even when it is neither selectable, saved, nor part of a pending/failed draft. Route: patch by showing temporary rows only when selected/saved, while preserving unsupported-category explanations. |
| R12 — the OIDC fixture does not bind the verifier to the S256 challenge | medium | The token validator accepted any non-empty verifier before this correction; the new authorization endpoint exposes that existing fixture limitation but did not cause the validator design. Route: defer as test-harness hardening outside the account-choice UX correction. |
| R13 — fixture authorization codes are reusable | medium | The fixture already accepted any non-empty code without issued-code state; single-use authorization-code enforcement is a pre-existing test-harness gap. Route: defer with R12. |
| R14 — Select All tri-state behavior lost regression coverage | medium | Verification-gap evidence found no remaining test that activates or inspects the aggregate control after its filtering logic changed. Route: patch with focused checked/mixed/unchecked coverage. |
| R15 — review Back/Escape and focus restoration lost regression coverage | medium | Verification-gap evidence shows current tests only save or inspect the dialog; none dismisses it or proves focus return and draft preservation. Route: patch. |
| R16 — non-ready inventory blocks durable-state redirect | medium | Same root cause as R1, independently reproduced by the edge-case review. Route: patch with R1. |
| R17 — a lost POST response with durable pending state is announced as failure | medium | Reconciliation unconditionally sets `saveFailed` whenever committed IDs do not yet match, masking a recovered `pending` or `failed` durable change. Route: patch by preferring the durable status. |
| R18 — an all-unselectable chooser exposes a no-op Select All | medium | Same root cause as R3, independently reproduced by the edge-case review. Route: patch with R3. |
| R19 — a zero-row chooser hides a durable pending or failed outcome | medium | Inclusion feedback is nested only inside the `accounts.length > 0` branch, so filtered/closed-account states suppress the persisted operation result. Route: patch by rendering the single feedback region for both branches. |
| R20 — durable inclusion loads only after ready inventory | false | The current inclusion effect has no ready-state guard and runs independently; the actual regression is the separate initial-inventory effect returning while inventory is null. |
| R21 — direct `/portfolio` describes unsaved choices as saved | false | Carried R2: the current Portfolio placeholder is the explicit Story 1.2 handoff, while this correction owns the OAuth-return setup route. |
| R22 — Select All is enabled when no row is selectable | false | The aggregate checkbox now includes `selectable.length === 0` in its disabled condition, with focused test coverage. |
| R23 — chooser counts label unselectable rows as available | false | The localized denominator now says “accounts shown,” matching the visible-row calculation. |
| R24 — pending additions appear cancellable despite remaining in flight | low | Carried R5: the edge is real, but cancellation/supersession semantics are not settled and a safe fix would add lifecycle complexity. Rejected. |
| R25 — the saved announcement reappears after later navigation | false | `ProtectedApp` now clears the transient notice whenever the requested route leaves `/portfolio`, and the return-path test covers it. |
| R26 — transient save failure survives draft edits | false | Account and Select All edits now clear `saveFailed`; durable pending/failed status intentionally remains authoritative. |
| R27 — focused chooser heading lacks a visible focus cue | false | The chooser-specific outline suppression was removed, so the global `h1:focus-visible` rule applies. |
| R28 — the setup grid clips at intermediate desktop widths | false | Carried R9: the `minmax(0, 47.5rem)` content track shrinks inside the bounded grid rather than forcing the claimed width. |
| R29 — duplicate masked labels are ambiguous in the review | false | Review entries now include both the masked label and brokerage label and key by immutable account id. |
| R30 — unsaved temporary failures are shown without actionable context | false | Temporary unavailable rows are now hidden unless saved or present in the recovery draft; unsupported-category rows retain their explanation. |
| R31 — fixture PKCE verifier is not bound to its challenge | medium | Carried R12: this is a pre-existing integration-fixture limitation already deferred outside the account-choice UX correction. |
| R32 — fixture authorization codes are reusable | medium | Carried R13: this is the same pre-existing fixture-hardening work already deferred. |
| R33 — Select All tri-state lacks regression coverage | false | Current tests activate Select All through mixed, checked, and unchecked states while preserving a committed unavailable row. |
| R34 — review Back/Escape and focus restoration lack coverage | false | Current happy-path coverage dismisses with both Back and Escape, verifies the draft remains, and asserts focus returns to Review. |
| R35 — initial null inventory suppresses the first inventory request | medium | The initial-load effect returns before calling `getPortfolioInventory`, leaving the page pending until Check again/Retry performs a request. Route: patch by moving the null guard to the inclusion-loading effect. |
| R36 — an inventory update can overwrite an actively edited draft | false | Inventory changes are user-triggered only in non-ready recovery states where chooser editing is unavailable; reaching ready performs the one necessary inclusion load before editing begins. |
| R37 — a hidden closed/missing failed addition is silently resubmitted | medium | `recoveryDraft` retains failed additions while chooser filtering can omit them, and Save currently serializes the entire draft. Route: patch by preserving hidden committed IDs but excluding hidden additions from the desired set. |
| R38 — fixture permits PKCE mismatch and code replay | medium | Carried R31/R32: the combined finding is pre-existing fixture hardening and remains deferred. |
| R39 — visual-capture readiness timeouts fail as null dereferences | low | The polling loops do not assert success before querying required elements, obscuring the actual readiness failure. Route: patch with explicit timeout errors. |

## Story 1.2 Handoff

After the initial save, Story 1.2 supplies the normal private Portfolio Showcase represented by `key-portfolio-showcase.html`. That Showcase owns balances, positions, activity, freshness, a **Continue setting up your profile** action for an incomplete first-time user, and one **Edit included accounts** link. The edit link opens the same selector as a distinct Portfolio child route/state (prefer `/portfolio/accounts`) inside the ordinary app shell with saved choices preselected; a successful edit returns to `/portfolio` rather than stacking both screens together. The initial `/onboarding/accounts` path uses onboarding progress and also returns to `/portfolio` after a committed save.

## Verification

- `cd frontend && npm test -- --run`
- `cd frontend && npm run typecheck`
- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd backend && golangci-lint run`
- `git diff --check`
