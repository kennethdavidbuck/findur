---
title: 'Accept SnapTrade connection and account webhooks'
type: 'feature'
created: '2026-09-23'
status: 'done'
route: 'dispatch'
baseline_commit: 'd0ac1c831af03015c2a7c1198f42fb4d963bc760'
review_loop_iteration: 0
context:
  - '_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
  - '_bmad-output/implementation-artifacts/spec-add-scheduled-connection-and-account-inventory-synchronization.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Findur learns connection and account lifecycle changes only through its daily inventory refresh, even though SnapTrade sends signed OAuth webhooks containing the event type and affected connection/account IDs.

**Approach:** Accept SnapTrade's `oauth_v1` webhooks, authenticate and normalize them at a thin HTTP boundary, and default to safe event-type logging without database mutation. An explicit processing feature flag applies supported lifecycle events as small idempotent updates to the current PostgreSQL inventory through a reusable application service. Make no SnapTrade API request; retain the scheduled refresh only as independent reconciliation, and defer inbox/outbox and message delivery.

## Boundaries & Constraints

**Always:** Verify the documented HMAC signature with the Commercial consumer key; require the configured OAuth client ID and schema version; default processing off so authenticated events are safely acknowledged and logged by event type only. When processing is explicitly enabled, resolve the webhook `userId` through the SnapTrade external identity; lock the owner and current inventory state before applying idempotent updates; update only the current head; perform inventory mutation, inclusion cleanup, and lifecycle fencing atomically. Record new connection/account IDs as safe, non-selectable provisional rows when the webhook lacks display metadata. Keep the transport independent so a later message consumer can call the same application service.

**Never:** Call SnapTrade from webhook processing; trust a webhook subject as a Findur user ID; store or log the consumer key, signature, raw payload, or financial data; delete stable account identities or retained financial history; make provisional accounts selectable; cancel or overwrite an active inventory claim; add an inbox, queue, outbox, frontend workflow, out-of-order delivery ledger, or holdings/transaction webhook processing in this slice.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Log-only default | Valid signed webhook while processing is not explicitly enabled | Log only the safe event type and acknowledge without database mutation | Return success after authentication and envelope validation |
| Explicit lifecycle | New signed `BROKEN`, `FIXED`, `DELETED`, or `ACCOUNT_REMOVED` event | Update/delete the identified current-head rows; remove newly ineligible inclusion and fence stale sync work | Duplicate delivery remains harmless through idempotent SQL |
| New identity | Signed `CONNECTION_ADDED` or `NEW_ACCOUNT_AVAILABLE` | Upsert a safe provisional connection/account identified by the supplied IDs; account remains non-selectable until reconciliation hydrates it | Missing required IDs is invalid; no placeholder guesses beyond bounded generic values |
| Irrelevant event | Valid signed but unsupported event type | Record/acknowledge without changing inventory | Return success to avoid futile retries |
| Untrusted request | Missing/bad signature, malformed JSON, wrong schema/client | No database mutation | Return bounded `400` or `401` response |
| Persistence failure | Valid supported event but transaction fails | No partial inventory or inclusion change | Return `503` so delivery may retry |
| Inventory busy | Valid supported event while a worker owns the inventory claim | Do not mutate or cancel the claim | Return `503` so delivery may retry |
| Browser projection | Signed event applied with processing enabled while an authenticated owner has inventory | The next normal inventory load/reload renders the changed connection/account state | No push channel, polling loop, or new frontend state layer |

</frozen-after-approval>

## Code Map

- `backend/internal/portfolio/webhook.go` -- new transport-neutral event model/service and supported lifecycle mapping; no provider dependency.
- `backend/internal/platform/postgres/inventory_lifecycle.go` -- resolve external subject, lock the current head, apply idempotent row updates, and reuse inclusion cleanup semantics without a schema change.
- `backend/internal/platform/httpapi/webhook.go` -- public bounded JSON handler, canonical HMAC verification, safe responses/logging, and call into the application service.
- `backend/internal/platform/httpapi/httpapi.go` -- register and categorize the optional webhook route outside browser session/CSRF middleware.
- `backend/internal/platform/config/config.go`, `backend/cmd/findur/main.go`, `render.yaml`, `compose.yaml` -- optional consumer-key configuration, log-only-by-default processing flag, and dependency composition.
- `backend/internal/auth/authorization.go`, `backend/internal/platform/oidcfixture/fixture.go` -- request `webhook` scope only when webhook reception is configured; existing grants require reauthorization.
- `backend/internal/{portfolio,platform/postgres,platform/httpapi,platform/config,auth}/*_test.go` -- table-driven event, authentication, idempotency, locking, rollback, and scope coverage.
- `test/integration/{integration.mjs,browser-session.mjs}` and `compose.yaml` -- send canonically signed webhook-shaped requests through the real stack and prove a browser reload renders the database transition without provider traffic.

## Tasks & Acceptance

**Execution:**
- [x] Add the transport-neutral event application service.
- [x] Implement guarded current-head PostgreSQL updates and inclusion/lifecycle cleanup.
- [x] Add the authenticated bounded HTTP receiver and optional configuration/composition.
- [x] Default to log-only receipt and require an explicit feature flag for database processing.
- [x] Request the OAuth webhook scope when enabled and update fixtures.
- [x] Add table-driven domain, repository, handler, configuration, and authorization tests.
- [x] Add lean compose-stack coverage for signed events, zero provider callbacks, and resulting browser inventory state.

**Acceptance Criteria:**
- Given processing is left at its default and a valid signed webhook is received, then the event type is safely logged and acknowledged without a database mutation.
- Given processing is explicitly enabled, an active Findur owner, and a valid supported signed webhook, when it is received, then the identified lifecycle change is visible from the current persisted inventory after one atomic transaction and no outbound SnapTrade request occurs.
- Given the same delivery more than once, when it is received, then each request succeeds and the persisted result is unchanged after the first application.
- Given a removal or disabling event for an included account, when it commits, then active inclusion is removed, sync claims are fenced, and retained identities and financial history remain.
- Given a future HTTP replacement by a message consumer, when it submits the same normalized event to the application service, then identical database behavior is available without HTTP dependencies.
- Given the compose stack and an authenticated browser, when a signed connection/account event is posted and the inventory view is normally reloaded, then the UI reflects the transition and WireMock records no webhook-triggered SnapTrade request.

## Implementation Notes

- Added `POST /api/webhooks/snaptrade` with bounded JSON decoding and canonical HMAC-SHA256 verification using the documented `Signature` header.
- Added `SNAPTRADE_WEBHOOK_PROCESSING_ENABLED`, default `false`; Render remains log-only while the compose integration stack explicitly enables processing.
- Kept HTTP, application service, and entity-oriented inventory lifecycle persistence separate so a later durable consumer can reuse the same service.
- Current-head mutations serialize per owner and preserve active inventory claims. Connection deletion relies on the existing current-inventory foreign-key cascade; connection broken only changes lifecycle state. Stable account identities and retained financial history are preserved.
- New identities use safe provisional rows. Provisional accounts use the existing `provisional_category` reason so the normal accounts UI renders them disabled until scheduled reconciliation hydrates them.

## Spec Change Log

- 2026-09-23: User explicitly changed the approved behavior to log-only by default, with database processing enabled only by an environment flag.

## Review Triage Log

| Finding | Verdict | Evidence and route |
|---|---|---|
| Existing OAuth grants are not automatically forced through the new scope consent | medium | SnapTrade confirms existing grants need reauthorization, but the approved intent explicitly excludes a new frontend workflow; document/manual reauthorization remains the bounded rollout path, so reject as out of scope. |
| Event timestamps are parsed but not freshness-bounded | medium | SnapTrade's official verification example rejects payloads older than 300 seconds; patch the envelope check with bounded past/future skew. |
| Log-only mode skips event-specific identifier validation | medium | The handler returns before service validation, so malformed supported events receive `204`; patch by sharing transport-neutral event validation before the mode branch. |
| Go's default JSON encoding is not fully compatible with the documented canonicalization | medium | Default HTML escaping and non-ASCII handling differ from documented compact, sorted Python JSON; patch canonical encoding and add known vectors. |
| Webhooks should require an active local provider authorization | false | The approved identity boundary requires an active Findur owner, not active local token lifecycle; accepting signed removal/broken events during local reauthorization is desirable and makes no provider call. |
| Missing inventory state/head is acknowledged as a no-op | medium | An addition can arrive before first publication and be discarded; patch active known owners to return retryable busy until a head exists. |
| Lifecycle mutations do not reconcile the current aggregate inventory state | medium | Empty/add/broken/fixed transitions can leave contradictory `current_status`; patch a bounded current-head state recomputation without rewriting historical versions. |
| `CONNECTION_FIXED` does not restore child account usability | false | The webhook lacks account metadata needed to prove usability; conservative child state until scheduled reconciliation is required by the approved no-provider-call design. |
| `CONNECTION_BROKEN` leaves child rows marked usable internally | medium | Read filtering hides those rows and the persisted child semantics contradict the disabled connection; patch children to non-selectable `connection_disabled` state atomically. |
| Rediscovered account under a different connection keeps its old association | medium | Both current row and stable identity use `DO NOTHING`; patch conflict handling to move the association while retaining hydrated metadata. |
| Processing enabled without a consumer key silently disables the endpoint | medium | The explicit processing request is contradictory and currently starts successfully; patch configuration validation. |
| Webhook request logs use the `unmatched` route category | low | `routeCategory` omits the new public path; patch the direct mapping. |
| Edge review: processing enabled without a consumer key | medium | Duplicate of the verified configuration contradiction; grouped into the same patch. |
| Edge review: known owner without a current inventory head is acknowledged | medium | Duplicate of the verified lost-addition path; grouped into the same retryable patch. |
| Expired inventory claim can be followed by a later finalizer | low | Scheduled finalization already rejects an expired lease and interactive finalization publishes reconciliation as a new generation; rejecting every expired claim could strand updates, so reject and codify the intended active-versus-expired behavior. |
| Edge review: account connection changes are ignored | medium | Duplicate of the verified association defect; grouped into the same patch. |
| Edge review: aggregate status is not recomputed | medium | Duplicate of the verified current-state contradiction; grouped into the same patch. |
| HTTP tests do not assert the normalized event mapping | medium | Pre-verified gap allows field swaps or omissions to pass; patch table-driven mapping and real validation coverage. |
| Production composition does not verify log-only default propagation | medium | Pre-verified gap allows `buildWebhook` to hard-code processing on; patch a composed log-only test using a nil persistence pool as a mutation tripwire. |
| Successful `CONNECTION_ADDED` persistence is unverified | medium | Pre-verified gap covers only routing and busy rejection; patch success, safe-field, and duplicate assertions. |
| Expired claims permitting mutation are unverified | low | Pre-verified gap around intended lease semantics; patch a focused expired-claim success case. |
| Inactive-owner lifecycle rejection is unverified | medium | Pre-verified gap allows removal of the `owner.active` guard; patch a no-mutation case. |

## Design Notes

Webhook payloads identify additions but do not contain Findur's required brokerage label, account type/category, masked label, sync status, or balance. Provisional rows therefore use existing bounded generic labels and unavailable/unknown policy states; the scheduled full inventory refresh may later replace them with provider metadata. Updating current-head rows is an explicit first-cut exception to the normal immutable inventory publication path. Row locks and idempotent state-setting operations provide bounded safety; durable receipt/order handling can be added with a future message-consumer/outbox design.

## Verification

**Commands:**
- `cd backend && GOCACHE=/tmp/findur-go-cache go test -race ./...` -- all backend and repository race tests pass.
- `cd backend && golangci-lint run` -- mandatory lint passes.
- `docker compose -p findur-webhook-verify --profile test run --rm --build integration` -- isolated clean-stack run passed; signed duplicate webhooks updated persisted state and the real browser projection without provider callbacks.
- Post-review compose rerun reached an unrelated existing profile focus assertion before the webhook scenario; backend race, focused webhook repository, handler, configuration, and composition tests all passed after the review fixes.
- `test -z "$(gofmt -l backend)" && git diff --check` -- formatting and patch whitespace are clean.
