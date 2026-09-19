# Adversarial Seam Review

## Verdict

**Changes required before handoff.** The chosen structure is coherent, but five cross-unit seams remain under-specified. Independently built units can satisfy their local rules and still violate real-user isolation, resurrect deleted financial state, strand a live rotated provider grant, mutate another module's tables inconsistently, or fail to restore a pinned disclosure view.

## Findings

### 1. Critical — Authenticated actor authority is not a binding application/repository contract

**Evidence:** The spine requires isolated viewers (`ARCHITECTURE-SPINE.md:26`, AD-9), opaque sessions (AD-3), generated strict handlers (AD-12), and consistently classified authorization failures, but never specifies where the authenticated actor is introduced or which identifiers may be trusted. The data model's ownership foreign keys prove that a row belongs to *a* user; they cannot prove it belongs to the current session's user.

**Compliant conflict construction:**

- The HTTP unit authenticates the session and passes a path/body `userId` or `portfolioId` to a service, assuming authorization is a service concern.
- The service accepts that opaque ID and calls a repository, assuming the generated handler already authorized it.
- The repository correctly loads the requested aggregate and observes every transaction and ownership rule.

All three units can look locally correct, yet user A can address user B's owner-private portfolio, inclusion state, profile, or disconnect operation. This is especially plausible because the API convention explicitly permits opaque IDs and the architecture does not ban client-supplied actor IDs.

**Required convergence:** Bind an authenticated `Actor`/principal at the application boundary. Owner-scoped commands and queries derive `user_id` only from the verified server session, never from a client field or route parameter. Application services authorize all addressed subordinate resources against that actor before mutation or projection, and repository operations for OAuth-owned data include the actor/owner scope rather than loading by globally opaque ID alone. Define the deliberate exception for generated candidate reads. Add cross-user integration tests for every owner-private read and mutation, including disconnect and account inclusion.

### 2. High — Out-of-transaction portfolio work has no lifecycle fence against removal or disconnect

**Evidence:** AD-20 mandates prepare/call/finalize and a version-checked final transaction, but it does not name the version that ordinary activity-driven refresh captures and validates. `account_inclusion_changes` guards its own addition flow; `dataset_versions`/`dataset_heads` have publication sequences but no authorization/inclusion lifecycle generation. AD-7 simultaneously requires account removal/disconnect to delete and suppress affected data immediately.

**Compliant conflict construction:**

1. The refresh service verifies account X is included, commits, and begins a slow SnapTrade fetch outside a transaction.
2. The inclusion service removes X (or disconnect deletes the portfolio data) in its short transaction.
3. The refresh service normalizes its already obtained response and publishes a new version/head in its own valid short transaction.

The refresh and removal units each follow AD-7 and AD-20 locally, but together they resurrect financial data after consent withdrawal and can make it visible again.

**Required convergence:** Introduce one monotonic portfolio authorization/inclusion lifecycle generation (or an equivalent explicit tombstone/fence). Every provider-data claim captures it; every finalize transaction must require the grant to remain active, the account to remain included, and the captured generation to remain current. Removal and disconnect advance the generation before deleting/suppressing data. A failed guard discards the transient result and may not create a dataset version/head. Apply the same fence to inventory refresh, account refresh, signal publication, and Account Inclusion addition finalization. Test delayed-finalize races against both removal and disconnect.

### 3. High — Refresh-token rotation and disconnect can leave a live provider grant after local deletion

**Evidence:** AD-4 separately defines a refresh lease and a disconnect that revokes the stored refresh token before deleting local state. It does not define their ordering or compensation when they overlap.

**Compliant conflict construction:**

1. Refresh claims the lease for refresh token R1 and calls SnapTrade outside a transaction.
2. Disconnect reads/revokes R1 and deletes the local authorization/portfolio state; revocation can fail because R1 has already been consumed.
3. SnapTrade returns rotated token R2 to the refresh worker.
4. Refresh's guarded publication correctly fails because the row is disconnected/deleted, so R2 is discarded locally.

Both flows obey their stated compare-and-swap and local deletion rules, but R2 can remain a live untracked provider grant. The blanket “provider success cannot be committed => reauthorization-required” rule also has no row/state to update after disconnect.

**Required convergence:** Make disconnect intent win through a defined authorization lifecycle protocol. At minimum, mark disconnecting and invalidate new refresh claims in a short transaction before external revocation; a refresh result whose install loses to disconnect must make one bounded best-effort revocation of the newly returned refresh token before discarding it. Define bounded handling for an already in-flight lease (wait, takeover deadline, or compensation), and only then finalize local deletion. This coordination must not hold a transaction across either external call.

### 4. High — The modules have no single-writer/table ownership map for cross-module invariants

**Evidence:** The structural seed assigns product areas to `auth`, `portfolio`, `discovery`, `matching`, `profile`, and `simulation`, while the data model creates shared aggregates. The rules require one transaction to cross those areas—for example, disclosure save plus anchor invalidation; inclusion removal plus anchor/signal/data deletion; swipe plus reciprocal swipe/match/anchor transition; and simulation publication across generated state, snapshots, signals, presence, and activities. AD-20 defines how to propagate a transaction, but not which module owns each write or which use case may coordinate foreign repositories.

**Compliant conflict construction:** The profile unit can expose direct anchor invalidation because AD-7 requires it, while discovery also treats anchors as its owned state. Portfolio can delete signals/anchors on removal while discovery cleanup deletes the same rows in another lock order. Simulation can either write portfolio tables directly or call portfolio services that try to open their own outer use case. Every team can honor transaction propagation yet choose incompatible write APIs, lock order, and invariant ownership.

**Required convergence:** Add a terse write-ownership matrix for shared tables/aggregates and name the application use case that owns each cross-module transaction. A table has one write-owning module; cross-module orchestration calls transaction-scoped ports supplied by that owner, not another module's SQL repository. Fix a lock/acquisition order for the few multi-aggregate operations. In particular, assign ownership for `users`, `profiles`/disclosure, dataset/signal heads, anchors, swipes/matches, and generated-state publication.

### 5. Medium — Candidate anchors claim to pin a disclosure version that does not exist immutably

**Evidence:** `candidate_anchors` references “the candidate disclosure version” (`DATA-MODEL.md:239`), but `disclosure_settings` is one mutable current row with a version counter (`DATA-MODEL.md:130-132`). There is no immutable disclosure-version entity, projection snapshot, or foreign key target. AD-11 promises a stable pinned card across material updates.

**Compliant conflict construction:** Discovery stores integer version 4 on an anchor; profile overwrites the single settings row to version 5. One presenter recomputes the anchored card using current settings, while another assumes the old policy can be recovered. Neither has a canonical version-4 record. Privacy decreases are invalidated, but nondecreasing material changes can still make Detail/Back disagree with the originally offered card.

**Required convergence:** Choose one representation and bind it: either store disclosure settings as immutable versions referenced by anchors, or persist a purpose-limited immutable candidate projection/policy snapshot for the anchor. Keep immediate invalidation for permission decreases. Retention cleanup must protect the selected representation until the anchor settles or expires.

## Exit Criteria

The gate can pass this lens when the spine/data companion establish:

1. actor-derived owner scoping from session through service and repository;
2. a lifecycle generation/fence for every external-work finalize;
3. explicit refresh/disconnect race compensation;
4. single-writer ownership and named coordinators for cross-module transactions; and
5. a real immutable target for the anchor's disclosure pin.

## Resolution Check

1. **PASS — Authenticated actor authority.** AD-20 now requires session middleware to derive an immutable `Actor`, bans client override of acting `user_id`, assigns service authorization and owner-scoped repository access, defines the generated-candidate exception, and requires cross-user tests.
2. **PASS — Lifecycle fence for delayed provider work.** AD-7 and the portfolio model now define a monotonic lifecycle generation plus inclusion-version/account-membership guards for inventory, account, signal, and addition finalizers. Removal/disconnect advance the fence before deletion, and stale work cannot publish.
3. **PASS — Refresh/disconnect race.** AD-4 now makes the operation lease mutually exclusive, makes disconnect intent block new claims and win lifecycle guards, bounds waiting for an in-flight refresh, and requires best-effort revocation compensation for a losing rotated token without holding a transaction over network I/O.
4. **PASS — Cross-module write ownership.** The spine now supplies a sole-write-owner matrix, named coordinating use cases, owner-supplied transaction-scoped ports, a common lock order, and concurrency-test coverage.
5. **PASS — Immutable disclosure pin.** The data model now defines immutable `disclosure_versions` with a current head; candidate anchors reference `disclosure_version_id`, and retention protects that version until settlement or expiry.

**Remaining blockers:** None from this adversarial seam review.
