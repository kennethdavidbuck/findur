# Product, Privacy, Consent, and Source-Fidelity Review — Findur PRD

**Review target:** `prd.md` and its subordinate `addendum.md`
**Sources reconciled:** completed product brief and addendum; SnapTrade feasibility research; all four `reconcile-*.md` extracts; PRD and brief memory logs.
**Verdict:** **Conditional pass.** The PRD preserves the core product direction: one owner-controlled live portfolio, synthetic Discovery candidates, rich portfolio-first visualization, disclosure-controlled detail, invisible cohorts, one initial swipe and mutual-photo reveal, a non-advisory stance, and no public multi-user launch. The issues below are requirements-strength gaps. If left unresolved, they could make a publicly hosted demo look like an authorized public/cross-user product or retain data after consent withdrawal.

## Fidelity confirmations

| Source decision | PRD evidence | Review result |
| --- | --- | --- |
| One live owner; no live cross-user financial data | `prd.md` §1, FR-2, FR-9, §5; addendum §1.2 | Preserved. |
| Rich API/card proof without a generic badge | FR-19–21; FR-11; addendum §§2.2–2.3 | Preserved. Live detail is owner-only; Discovery is synthetic. |
| Disclosure-controlled exactness | FR-6, FR-11, FR-20; addendum §1.2 | Preserved in principle; the required field bundles remain an implementation gate. |
| Invisible, continuously evolving cohorts | FR-10; §6.3; addendum §§1.3, 1.5 | Preserved, with market-event behavior appropriately deferred. |
| Single-decision progressive identity reveal | FR-12, FR-15; addendum §1.4 | Preserved. |
| No advice and real-user/cross-user restrictions | FR-1, FR-21, §5, §10; addendum §§3.1–3.2, 4.4 | Preserved, subject to the user-initiation and provenance gaps below. |
| Public demo rather than product launch; safety/legal direction | FR-18, FR-25–26, §6.2, §10; addendum §§2.4, 2.7, 4 | Preserved in intent; external-entry and approval gates need strengthening. |

## Findings

### High

1. **H-1 — A public visitor is offered a sign-in/demonstration path even though only the owner may authorize.**
   **Evidence:** UJ-6 says Alex may choose the “sign-in or demonstration path” (`prd.md` §2.4, lines 58–60); FR-25 requires an entry action (lines 372–381). FR-2 permits OAuth only for the owner-controlled test user (lines 102–112), and the research limits the test app to five users (addendum §4.4).
   **Risk:** A hosted public page can accidentally invite unapproved visitors to OAuth, consume test-app capacity, or imply public availability.
   **Concrete fix:** Add an FR acceptance criterion that the OAuth entry is restricted to the named owner/test identity (or a controlled reviewer allowlist); every other visitor gets a clearly non-authorizing demo/review path and an explicit “not accepting users” message.

2. **H-2 — The directional disclosure-and-initiation rule is no longer testable.**
   **Evidence:** The source says a user initially sees candidates who disclosed no more than the user, while a more-disclosing user may initiate toward a less-disclosing user (`brief/addendum.md` §Tiered Disclosure and Initiation, lines 55–61). FR-6 instead prohibits receiving “more financial detail” and permits higher-to-lower initiation (`prd.md` lines 209–220). FR-10 only names generic “Disclosure Level” eligibility (lines 256–265).
   **Risk:** An implementer can display a redacted high-disclosure candidate to a low-disclosure user and still satisfy FR-6, losing the source's differentiated discovery-access model.
   **Concrete fix:** Specify the eligibility relation separately from display redaction: define which disclosure levels a user may receive in their initial deck, the higher-to-lower inbound-interest exception, and expected behavior for every level pair.

3. **H-3 — Synthetic provenance may be only contextual, not a reliable user-facing boundary.**
   **Evidence:** FR-9 requires synthetic data to be “visibly or contextually” identified only where a *reviewer* could otherwise mistake it for real data (`prd.md` lines 245–255). SM-C4 recognizes that realism confusion is a failure (lines 466–471).
   **Risk:** Rich, exact-looking holdings, values, activity, and Photos can be mistaken by a visitor for actual connected people and for authorized cross-user financial disclosure—the precise boundary the research prohibits.
   **Concrete fix:** Require a persistent, accessible synthetic-demo provenance treatment on Discovery and matched states; prohibit `Verified` or connected-account claims on candidates; include provenance as a required data-model field and test it in the privacy-boundary metric.

4. **H-4 — Withdrawal does not explicitly invalidate information already rendered or cached.**
   **Evidence:** FR-4 stops *new* discovery use and explains deletion/retention (`prd.md` lines 123–131); FR-6 suppresses information only in “subsequent views” (lines 209–220). The addendum appropriately asks about rendered/cache behavior (line 28) and tells architecture to prevent stale cached presentation after a consent or safety change (lines 166–174), but that protection is not a PRD-level acceptance criterion.
   **Risk:** After disconnect, downgrade, block, or report, an active screen, client cache, or precomputed deck can continue showing no-longer-authorized portfolio detail.
   **Concrete fix:** Add a testable lifecycle requirement: invalidate/re-fetch affected card, preview, and derived-signal views immediately on consent or disclosure reduction; prevent reuse from client/server caches; state the remaining retention status.

### Medium

5. **M-1 — The SnapTrade user-initiation condition is an open question, not an acceptance condition.**
   **Evidence:** Research says that analysis or signals must be initiated by the licensed end user (`research.md` lines 46–56). FR-1 records explicit consent (lines 91–101), but the exact qualifying action is still an open question (`prd.md` §11.9), while the addendum only instructs architecture to document it (line 151).
   **Risk:** Automatic derivation and subsequent refreshes may be implemented without a defensible, auditable initiation record or scope, despite the no-advice wording.
   **Concrete fix:** Name the exact owner action that authorizes first derivation, any refresh/rederivation, the purpose it covers, and where the event is recorded. Make written provider confirmation a release gate before any different or cross-user flow.

6. **M-2 — Terms, Privacy, and safety pages can remain drafts with no approval gate before external sharing.**
   **Evidence:** FR-26 permits draft legal content when labeled (lines 383–392); ownership of review and approval is left open (`prd.md` §11.11). The source says legal text needs appropriate review and the technical research conditions any public use on legal/privacy review.
   **Risk:** A publicly hosted financial-data consent flow can be shared externally with unfinished privacy, retention, safety, or contact representations.
   **Concrete fix:** Define a demonstration-release checklist and accountable approver for externally reachable OAuth/consent surfaces. At minimum, require reviewed data-use, owner-only/synthetic boundary, retention/deletion explanation, support contact, and incident/escalation path before sharing the URL beyond controlled reviewers.

7. **M-3 — Personal-profile PII lacks an explicit lifecycle and logging boundary.**
   **Evidence:** FR-23 requires display name, coarse location, Photo, biography, relationship intent, and age confirmation (`prd.md` lines 173–182). NFR-1–5 set strong financial-data controls (lines 479–485) but do not expressly cover the new profile fields; the source requires every required field to earn a discovery or trust purpose (addendum line 61).
   **Risk:** Photos, location, and free text can be retained, served, or logged more broadly than necessary even if portfolio payloads are protected.
   **Concrete fix:** Extend the data inventory/minimization and logging requirements to Personal Profile fields: purpose, audience/state visibility, storage location, deletion on account/demo reset, and exclusion from diagnostics. State that adult-age confirmation does not require collecting a full birth date unless justified.

8. **M-4 — Disclosure-level definitions are deferred but no ready-for-build definition of done prevents arbitrary exactness.**
   **Evidence:** FR-6 permits exact security names, quantities, values, performance, activity, and transactions depending on the level (`prd.md` lines 209–220), while the level bundles are still an open question to resolve before FR-6/20 implementation (§11.1).
   **Risk:** A designer can use eye-catching exact synthetic details without a coherent per-level purpose, preview, downgrade behavior, or inference-risk review; the demonstration then teaches an unsafe future product rule.
   **Concrete fix:** Add an explicit pre-implementation exit criterion: a reviewed level matrix covering every field/visualization, source vs derived status, exact/bucketed/hidden treatment, pre/post-match state, rationale, downgrade invalidation, and synthetic provenance.

### Low

9. **L-1 — Retention remains an acknowledged decision but lacks a bounded demo default.**
   **Evidence:** FR-4 offers immediate deletion, queued deletion, or stated retention (lines 123–131); the exact period remains open (§11.5).
   **Risk:** Implementations can choose materially different retention behavior, weakening the consent explanation and the disconnect demonstration.
   **Concrete fix:** Set a temporary demonstration default (for example, delete source and derived records on disconnect unless a documented technical/legal exception applies), then label the production policy as deferred.

10. **L-2 — The primary end-to-end metric omits recently added profile and public-site requirements.**
    **Evidence:** SM-1 claims to validate FR-1 through FR-21 (`prd.md` lines 449–454), but the required owner profile and preview are FR-23/24 and the public trust surfaces are FR-25/26.
    **Risk:** The headline acceptance run can pass while privacy notices, owner-profile disclosures, or responsive public routes are missing.
    **Concrete fix:** Expand SM-1's scripted path or state which separate acceptance suite validates FR-23–26 and their consent/privacy boundaries.

11. **L-3 — “Continuous” market evolution is preserved only as direction, with no explicit transition criterion.**
    **Evidence:** The brief envisages a living market that evolves with portfolio, engagement, and market changes; the PRD correctly defers evolving and event-driven behavior to §6.3 and addendum §1.5.
    **Risk:** A future team may treat static seeded decks as the permanent model, or add dynamic cohorts without the source's consent/freshness/explanation safeguards.
    **Concrete fix:** Keep the current deferment, but add a future-story trigger that requires cohort lifecycle, user controls, consent, freshness, disclosure interaction, and anti-shaming review before any evolving/event-driven behavior is enabled.

## Severity summary

| Severity | Count |
| --- | ---: |
| Critical | 0 |
| High | 4 |
| Medium | 4 |
| Low | 3 |

## Review conclusion

The current artifacts retain prior decisions rather than reverting to a generic financial-verification dating app: rich visualization is now reconciled with owner-only live data and synthetic Discovery, exactness is disclosure-controlled, and public launch remains gated. Resolve H-1 through H-4 before using the document as an implementation baseline for a hosted demo; they are the shortest path to preserving the intended owner-only, consent-led, non-cross-user boundary in executable requirements.
