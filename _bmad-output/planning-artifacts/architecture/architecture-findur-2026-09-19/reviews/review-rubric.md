# Architecture Reviewer Gate — Rubric Walker

**Artifacts reviewed:** `ARCHITECTURE-SPINE.md`, `DATA-MODEL.md`

**Rubric:** BMad Architecture good-spine checklist
**Verdict:** **Conditional pass — structurally strong, but not ready to finalize until three high-risk seams are made explicit.**

The spine is unusually complete for a first-cut evaluation build. Its paradigm and dependency direction are clear; the API, provider, persistence, transaction, privacy, deployment, CI, observability, accessibility, localization, synthetic-data, and operational boundaries are all represented. The deterministic lint pass reports no mechanical findings. Most ADs state real divergence points and pair them with testable constraints. The remaining issues are concentrated in concurrency at permission boundaries, callback recovery, and security/lifecycle decisions that are named but not actually fixed.

## High findings

### H1 — Permission withdrawal has no explicit stale-work fence

**Evidence:** `ARCHITECTURE-SPINE.md:96-100`, `:127-130`, `:136`, and `:245`; `DATA-MODEL.md:94-99`, `:166-175`, and `:276-278`.

AD-7 promises that Account Inclusion removal or disconnect immediately suppresses and deletes financial state, while AD-20 correctly requires provider work outside transactions. Those two rules create an unavoidable race: a refresh, inventory fetch, token refresh, or account-addition call can be in flight while the permission-decreasing operation commits. The generic instruction to publish with a “version-checked transaction” does not say which consent/lifecycle version every publisher must capture and validate. Separate implementers could therefore guard inclusion changes but not background refresh, or guard snapshot publication but permit a rotating token refresh to finalize after disconnect began.

This is a real divergence and privacy boundary, not implementation detail. The spine should require a monotonic authorization/inclusion generation (or equivalent lifecycle fence) captured before every provider call and rechecked in the publish transaction. Disconnect must first transition to a terminal/disconnecting state and bump the fence in a short transaction; that state must reject new leases/calls, and every in-flight finalizer must fail closed. Account removal must similarly prevent all stale publishers from recreating deleted account-scoped state. Add integration tests that pause each external call, commit disconnect/removal, then prove the delayed finalizer cannot publish credentials, inventory, snapshots, signals, or anchors.

**Disposition:** Autofix in AD-4, AD-7, AD-20, and the data-model lifecycle rules before finalization.

### H2 — The OAuth callback is single-use, but its recoverable terminal outcome is not modeled

**Evidence:** `ARCHITECTURE-SPINE.md:81`, `:83-88`, and `:241-246`; `DATA-MODEL.md:79-87`. The UX source also requires callback replay/refresh to resolve idempotently to a recorded result or safe recovery state.

The attempt is consumed before the code exchange, which is the correct anti-replay posture, but the model contains only expiry and `consumed_at`. It does not define a durable callback state/result or what happens when token exchange succeeds but identity/authorization/session finalization fails. The authorization code cannot safely be replayed, yet a valid provider grant may now exist. Implementers could show a generic error, retry the consumed code, retain uncommitted tokens in memory, or create duplicate users/sessions on callback refresh.

Require a small callback state machine with one durable terminal outcome and safe recovery route. The code exchange remains outside transactions. A guarded finalize transaction should atomically bind `(provider, sub)`, persist encrypted authorization state, and create the session/result. If a response may have minted tokens but local finalization cannot commit, discard plaintext immediately, attempt bounded revocation outside a transaction when the token is known, and require fresh authorization. Replayed callbacks return the recorded safe result and never exchange the code again.

**Disposition:** Autofix in AD-3/AD-4/AD-20 and `oauth_attempts` (or a separate callback-result record).

### H3 — “CSRF protection” is a requirement, not an architecture decision

**Evidence:** `ARCHITECTURE-SPINE.md:88`, `:167-175`, and `:261-267`.

AD-3 fixes a cookie-authenticated same-origin SPA and then says only “and CSRF protection.” That leaves incompatible choices at the HTTP/frontend seam: synchronizer token, double-submit cookie, Origin checking, Fetch Metadata enforcement, or reliance on `SameSite`. Some choices alter generated OpenAPI headers and browser-client behavior; others do not meet the same threat model. The current Rule is not independently enforceable and does not fully prevent the stated session-security divergence.

Bind one concrete policy for every unsafe method. A suitable first-cut policy is a server-issued session-bound CSRF token carried in a required custom header, plus strict same-origin `Origin`/`Referer` validation and rejection of cross-site Fetch Metadata; `SameSite=Lax` remains defense in depth rather than the sole control. Explicitly exempt only OAuth callback and safe methods, rotate/invalidate with the session, and cover the policy in HTTP integration tests and OpenAPI where applicable.

**Disposition:** Discuss only if the team prefers another standard pattern; otherwise autofix AD-3 and AD-12.

## Medium findings

### M1 — The profile cardinality conflicts with the required onboarding lifecycle

**Evidence:** `DATA-MODEL.md:21-25`, `:122-132`; `ARCHITECTURE-SPINE.md:83`, `:296-300`.

The ERD says every user has exactly one profile and portfolio, while the prose describes a complete profile as containing all required fields. A newly authenticated OAuth user necessarily exists before completing profile, Account Inclusion, preferences, and first Disclosure Level save. The model does not decide whether the profile row is absent, nullable as a draft, or initially populated with invented defaults. This can lead to incompatible schemas and route-gate logic, including accidental default consent or discovery eligibility.

Fix the lifecycle: create identity/authorization/session first; represent incomplete onboarding without fabricated values; define whether `profiles` is optional until first save or is a versioned draft with nullable fields; and make Discovery readiness a derived server-side predicate over profile completeness, a committed non-empty inclusion set with usable positions, saved preferences, and explicitly saved disclosure. The ERD cardinality and constraints must match that rule.

**Disposition:** Autofix the data model and add the readiness predicate to an AD or consistency convention.

### M2 — Several adopted rules deliberately leave implementation choices open

**Evidence:** `ARCHITECTURE-SPINE.md:84`, `:138`, `:182`, `:202`, and `:309-310`.

Phrases such as “a maintained OIDC verifier such as,” “Prefer” the rate/circuit-breaker packages, “when supported,” and “an accessibility-first foundation such as React Aria Components” do not bind independently built units. The Deferred section then postpones all exact stack versions to Story 0. The reviewer rubric specifically requires named technology to be verified-current and rejects adopted rules that allow incompatible choices without a revisit owner.

The currently named version families are plausible: official releases show `oapi-codegen` v2.8.0, `openapi-typescript` 7.13.0, and `openapi-fetch` 0.17.0. Render documents that `preDeployCommand` is available for paid web services but not free web services, which supports the existing startup-lock fallback. However, the spine should either bind the selected packages and verified compatible version/toolchain set now, or remove package preferences from the invariant and make Story 0 a named blocking open question with a single owner and acceptance artifact. In particular, `oapi-codegen` v2.8 requires a supported modern Go toolchain and its generated code also requires the corresponding runtime package.

Primary references checked:

- https://github.com/oapi-codegen/oapi-codegen/releases
- https://github.com/openapi-ts/openapi-typescript/releases
- https://render.com/docs/deploys#pre-deploy-command
- https://github.com/passiv/snaptrade-sdks

**Disposition:** Autofix the wording; pin exact versions together in Story 0 if the repository does not yet exist, but do not call the spine final until that blocking gate has an explicit outcome.

### M3 — Important “bounded” durations are not centrally owned or acceptance-testable

**Evidence:** `ARCHITECTURE-SPINE.md:88`, `:96-100`, `:117`, `:128`, `:136-139`, `:164`, and `:246`; `DATA-MODEL.md:81-87`, `:107-109`, and `:269-278`.

OAuth-attempt expiry, session idle/absolute expiry, refresh-lease deadline, wait/retry bound, recent-activity window, old-snapshot grace, current-card anchor lease, and transaction retry limit are all described qualitatively. Different stories can select incompatible values, and several values directly control security, privacy retention, or provider load. The freshness thresholds are centralized and testable; these values need the same treatment.

The spine need not carry every number, but it must name one typed configuration owner and require checked-in evaluation defaults with validation bounds and acceptance tests. Security/retention defaults should not be mutable client inputs. Changes to retention, OAuth/session expiry, and provider retry budgets should be reviewed as architecture/config changes rather than scattered constants.

**Disposition:** Autofix through one configuration convention and a short defaults table, or explicitly defer each family with owner and revisit condition.

## Low findings

### L1 — The capability map does not fully reconcile the cited PRD/UX surface set

The map covers the data-heavy core but omits the public shell, trust/safety/legal/contact surfaces, onboarding/readiness gates, locale/theme/installability, and the explicit supersession of notifications. Most of these need no new AD, but omitting them from the only capability map weakens the claim that every source capability landed. Add rows mapping them to `web`, `internal/profile`/auth gating, AD-16/AD-21, or the upstream reconciliation section. This is especially useful because FR-28 and the one-owner assumption were intentionally superseded rather than accidentally missed.

### L2 — A few data-model enforcement statements overclaim database enforcement

`DATA-MODEL.md:185` says partial unique indexes enforce the “next sequence”; they enforce uniqueness, not next/gap-free allocation. `:251` correctly admits the opposite-origin swipe invariant is service/test enforced, while `:283-285` generally labels the section Database Enforcement. Reword the sequence claim to uniqueness/monotonic allocation under a locked head or sequence source, and distinguish database constraints from service invariants consistently.

## Checklist result

| Checklist dimension | Result | Notes |
| --- | --- | --- |
| Real divergence points fixed | Conditional | Most are fixed; permission-decrease fencing, callback recovery, CSRF, and onboarding state remain. |
| `Binds` / `Prevents` / `Rule` present | Pass | Lint reports zero findings; all 21 ADs carry the required structure. |
| Rules enforce their stated prevention | Conditional | H1-H3 and M2 are the substantive exceptions. |
| Nothing dangerous deferred | Conditional | Exact stack compatibility is a blocking Story 0 gate; okay only while status remains draft. Qualitative security/retention bounds need ownership. |
| Named technology verified-current | Conditional | Core version families were verified against primary release/docs pages; the complete compatible pin set is still deferred. |
| Brownfield ratification | Not applicable | This is an early-stage planning repository, not an implemented application. |
| Source-spec capability coverage | Conditional | Core FRs/NFRs are covered or explicitly superseded; public/trust/onboarding coverage should be visible in the map, and callback recovery has a direct UX mismatch. |
| Parent spine inheritance | Not applicable | No parent architecture spine is declared. |
| Structural dimensions swept | Pass with fixes | Paradigm, module boundaries, state mutation, data ownership, API, security, deployment, CI, operations, and deferred production controls are all represented. |

## Gate recommendation

Resolve H1-H3 and M1 before changing `status: draft` to `final`. M2 may remain a Story 0 gate only if the architecture remains draft until the selected stack is pinned and compatibility checks pass. M3 and both low findings are safe to apply during final polish. No redesign of the modular-monolith, OpenAPI, PostgreSQL, Render, SnapTrade-adapter, or synthetic-population direction is indicated.

## Resolution Check

Focused recheck against the updated `ARCHITECTURE-SPINE.md` and `DATA-MODEL.md`:

| Finding | Result | Resolution evidence |
| --- | --- | --- |
| H1 — stale-work fence | **PASS** | Authorization and portfolio lifecycle generations now fence every provider call/finalizer; disconnect becomes `disconnecting`, blocks new claims, and wins against in-flight refresh. Account removal advances the same fence, and delayed-publication integration tests are required. |
| H2 — OAuth callback recovery | **PASS** | The callback is now a durable `pending -> exchanging -> succeeded|restart-required` state machine. Replay never re-exchanges the code; identity, encrypted authorization, session, and terminal result finalize atomically, with bounded revocation and fresh authorization after an ambiguous failure. |
| H3 — CSRF decision | **PASS** | Unsafe session-authenticated methods now require a session-bound synchronizer token in `X-CSRF-Token`, same-origin validation, and Fetch Metadata enforcement; exemptions, rotation, OpenAPI declaration, and defense-in-depth role of `SameSite` are explicit. |
| M1 — onboarding/profile lifecycle | **PASS** | OAuth profiles are explicitly optional during onboarding and created only by a complete atomic first save. Discovery readiness is a server-derived predicate over profile, preferences, disclosure, inclusion, and usable signals; the ERD cardinality now agrees. |
| M2 — nonbinding technology choices | **PASS** | OIDC, rate limiting, circuit breaking, React Aria, OpenAPI generation, and request validation are now direct selections rather than “prefer/such as” examples. Current version baselines and checked-in Story 0 pinning artifacts are named; the SnapTrade incompatibility has a deterministic test gate and bounded generated-client fallback. |
| M3 — bounded policy ownership | **PASS** | `internal/platform/config` is now the sole typed owner of all identified timeout, expiry, lease, retention, retry, and drain limits, with checked-in defaults, validation bounds, acceptance tests, and architecture-review requirements. |

The two low findings are also resolved: the capability map now includes the public shell, onboarding/readiness, locale/theme/installability, and notification exclusion; dataset sequence wording now distinguishes uniqueness from serialized monotonic allocation and separates service invariants from database constraints.

**Remaining finalization blockers: none from this review.** The deterministic spine lint still reports zero findings. Story 0's Render topology proof, SnapTrade bearer-shape compatibility gate, and checked-in dependency pins remain explicit implementation prerequisites, not unresolved architecture divergence.
