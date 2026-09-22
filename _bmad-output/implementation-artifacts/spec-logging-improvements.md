---
title: 'Logging Improvements: Correlate Authenticated HTTP Requests Safely'
type: 'feature'
created: '2026-09-22'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '0ab78620926896682ea2c3b639e2b9a452c27ef4'
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Operators can correlate an HTTP request with its opaque request ID but cannot identify the authenticated Findur owner who made it. They do not have direct database access, so support and operational investigations need privacy-preserving identity and request context in the structured logs themselves. Every identifier must state its provenance: Findur IDs and SnapTrade resource IDs are different values.

**Approach:** Enrich request-scoped structured logs with a server-derived `findur_user_id` after session validation, set it once in request metadata, and make it available to every `slog` event emitted during that request. Preserve the existing concise request-completion event, add already-available SnapTrade resource UUIDs at relevant operation boundaries, and add centralized, categorical error events for unexpected failures.

## Boundaries & Constraints

**Always:** Use uppercase underscore (`THIS_CASE`) structured keys consistently for every record changed by this task. Generate and return a request ID for every HTTP request; log a completion event with `REQUEST_ID`, method, categorized route, status, and latency. Add `FINDUR_USER_ID` only after `SessionService.Authenticate` or `AuthorizeUnsafe` returns the immutable server-derived Actor; do not set it for unauthenticated, rejected-before-authentication, or failed-session requests. Set that request-metadata value once, and automatically enrich the completion event and every HTTP `slog` record using its context. At the generated-handler error boundary, emit one `ERROR` event for an unexpected response/serialization/handler failure with a stable categorical failure field and the request context; preserve client validation and authorization rejections as normal response/completion events rather than noisy errors. When already present at an operation boundary, log an explicitly named SnapTrade UUID: `SNAPTRADE_ACCOUNT_ID` for a scheduled account claim, and bounded UUID-validated `SNAPTRADE_ACCOUNT_IDS` for an authenticated account-inclusion request. Include a `RESOURCE` category where the operation already knows it. `SNAPTRADE_USER_ID` and `SNAPTRADE_CONNECTION_ID` are omitted unless a future operation already supplies those exact provider values; this code must never label the Findur user UUID or OAuth/OIDC subject as a SnapTrade identifier. Never make a provider/database lookup solely to enrich logs. Replace raw worker error values with stable categorical failure fields. Preserve JSON logs to stdout and existing safe categorical callback/readiness logs.

**Never:** Derive identity from cookies, headers, body fields, URLs, or any client-supplied actor value. Do not log raw error text, session/CSRF/OAuth values, raw paths or query strings, request/response bodies, profile text, account labels, brokerage/institution account identifiers or numbers, balances, positions, activities, provider payloads, idempotency keys, IP addresses, or user agents. The only account IDs allowed are UUID-shaped IDs identified as SnapTrade resource IDs; ignore malformed/unbounded client values. Do not alter session authorization behavior, OpenAPI contracts, database schema, or log retention/deployment configuration. Do not broaden this task into provider-client error logging or make extra provider calls; both need a dedicated taxonomy/integration decision.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Anonymous request | Public, unmatched, or protected request without a valid session | Completion log retains only safe request fields and omits `FINDUR_USER_ID` | Existing categorical HTTP response remains unchanged |
| Authenticated read | Protected GET successfully returns an Actor | Completion log and any handler event using that request context include the Actor UUID as `FINDUR_USER_ID` | Repository/session failure does not emit an identity |
| Authorized mutation | Valid session and all CSRF/origin/fetch defenses pass | Completion event includes `FINDUR_USER_ID` after authorization and otherwise retains safe route/status/latency fields | CSRF/origin failures omit identity because authorization did not succeed |
| Unexpected handler failure | A protected handler returns an unclassified internal error or response serialization fails | One `ERROR` record includes the request context and a stable `handler_failure` category; completion remains a `503` request record | Client receives the existing categorical unavailable response; raw error is not serialized or logged |
| Account synchronization | A claimed scheduled resource has an owner and SnapTrade account UUID | Claim and finish events include `user_id`, `claim_id`, `snaptrade_account_id`, resource, and categorical outcome | Worker failures log event/category and safe structural context, never a raw Go error |
| Inclusion request | An authenticated owner submits selected SnapTrade account UUIDs | Completion and unexpected-error events include a bounded validated `snaptrade_account_ids` field and route context | Invalid/non-UUID values are omitted from metadata; response behavior is unchanged |
| Sensitive request input | A path, query, cookie, body, or provider/account value contains a private value | No added logging field serializes that value | Existing safe route categorization continues to prevent path leakage |

</frozen-after-approval>

## Code Map

- `backend/internal/platform/httpapi/httpapi.go` -- owns request-ID creation, request context, request-completion event, safe route categories, and tests that prohibit raw unmatched paths. Add narrowly scoped request log metadata, set `FINDUR_USER_ID` once, read its snapshot at every contextual log emission and when the request completes, and use `THIS_CASE` attributes.
- `backend/internal/platform/httpapi/authorization.go` -- turns validated sessions into `auth.Actor` for all protected GETs and unsafe mutations; mark request metadata only after these successful authentication/authorization boundaries, including status/logout where applicable. The inclusion body carries candidate SnapTrade account IDs; admit only bounded UUID-shaped values as contextual targets after successful authorization.
- `backend/internal/platform/httpapi/authorization.go` -- its generated strict-handler response-error callback currently converts unexpected errors to a `503` without logging; emit one privacy-safe categorical error event using the supplied request context, but keep expected request-validation branches silent.
- `backend/internal/platform/httpapi/authorization_test.go` -- session stubs and endpoint coverage; extend fixtures so an Actor can be returned and assert identity reaches request logging only on authorized requests.
- `backend/internal/platform/httpapi/httpapi_test.go` -- unit-test safe request-completion fields and a context-aware slog handler without exposing request inputs.
- `backend/cmd/findur/main.go` -- production composition creates the JSON slog logger; wrap its handler so request-scoped metadata enriches every HTTP log record while background/process logs remain unchanged.
- `backend/internal/auth/session.go` -- reuse `Actor.UserID()` as the only identity source; do not change session policy or Actor construction.
- `backend/internal/portfolio/sync.go` -- `SyncClaim.Owner` is the server-owned user UUID and `AccountID` is the existing SnapTrade account UUID; add both to claim/outcome records and replace raw error attributes with stable failure categories.
- `backend/internal/portfolio/sync_test.go` -- pin owner/SnapTrade-account correlation and ensure worker failure logs omit raw error detail.

## Tasks & Acceptance

**Execution:**
- [ ] `backend/internal/platform/httpapi/httpapi.go` -- introduce an internal request-scoped metadata carrier and a context-aware slog handler/wrapper; record only approved fields and preserve the current safe request-completion schema.
- [ ] `backend/internal/platform/httpapi/authorization.go` -- register the server-derived Actor identity after successful session authentication/authorization; capture bounded UUID-validated selected SnapTrade account IDs after unsafe authorization; and add one safe error event at the centralized unexpected-handler boundary, without changing authorization results.
- [ ] `backend/cmd/findur/main.go` -- use the context-aware logging handler for the production JSON logger so contextual HTTP events are enriched automatically.
- [ ] `backend/internal/platform/httpapi/{httpapi_test.go,authorization_test.go}` -- prove authenticated completion and handler logs carry the expected UUID, anonymous/failed/forbidden requests omit it, and sensitive values remain absent.
- [ ] `backend/internal/portfolio/{sync.go,sync_test.go}` -- correlate each account sync claim/outcome with the server-owned user UUID and existing SnapTrade account UUID, and use only categorical worker error fields.

**Acceptance Criteria:**
- Given an authenticated owner makes a protected request, when session authentication succeeds, then its completion log contains that server-derived UUID as `FINDUR_USER_ID`, `REQUEST_ID`, route category, method, status, and latency, without raw request inputs.
- Given an HTTP handler emits an `slog` event with the request context after authentication, when it is written, then it includes the same `REQUEST_ID` and `FINDUR_USER_ID` as the completion event.
- Given an anonymous request or an invalid session/CSRF/origin check, when it is logged, then it has no `FINDUR_USER_ID` and the API’s existing response semantics are unchanged.
- Given an HTTP handler returns an unexpected error, when the strict handler converts it to an unavailable response, then one `ERROR` event has the request context and a stable failure category while no raw error text is exposed.
- Given an account synchronization claim or completion, when an operator filters logs by its owner UUID or SnapTrade account UUID, then every emitted outcome for that resource is discoverable without exposing brokerage/institution account identifiers.
- Given an authenticated inclusion request, when selected account values are valid bounded SnapTrade UUIDs, then they are attached to its request context; malformed values are not logged.
- Given attacker-controlled paths, queries, cookies, bodies, or provider/account references, when requests are handled, then none appear in request logging fields.

## Implementation Notes

## Spec Change Log

## Review Triage Log

| Finding | Verdict | Evidence |
|---|---|---|
| Static logger attributes could duplicate reserved request metadata | medium | Fixed: `RequestLogHandler.WithAttrs` filters reserved keys, and focused tests prove each reserved key appears exactly once. |
| Per-event logger attributes could spoof server-derived request metadata | medium | Fixed: request-scoped records strip nested or direct reserved attributes before the metadata snapshot is appended; spoofing tests cover all three keys. |
| SnapTrade account IDs can be overwritten in request metadata | false | The current handler has one authorized inclusion writer; the one-time requirement applies to `FINDUR_USER_ID`, which is guarded against replacement. |
| Malformed inclusion values cause unbounded log-extraction work | false | The request body is capped by `inclusionRequestLimit` (1 MiB) before generated binding, so extraction is bounded by the existing request limit. |
| The 100-ID cap and duplicate behavior lacked verification | low | Fixed: a 101-UUID helper test pins the output cap. Duplicate preservation matches the submitted operation and does not violate an approved invariant. |
| Failed unsafe authorization did not prove identity omission | low | Fixed: a real valid session with missing CSRF asserts the resulting request log omits `FINDUR_USER_ID`. |
| Successful handler-event propagation lacked direct coverage | false | The contract is conditional on a handler event being emitted; real authenticated handler-failure coverage proves context propagation, and successful requests emit the required completion event. |
| Inclusion endpoint logging did not verify actual valid/malformed request values | medium | Fixed: endpoint coverage proves valid SnapTrade UUIDs reach completion logs and malformed values do not. |
| Malformed persisted sync account IDs lacked omission coverage | medium | Fixed: buffered worker-log coverage asserts the raw malformed value and `SNAPTRADE_ACCOUNT_ID` are both absent. |

## Design Notes

`context.Context` values are immutable, so a handler cannot replace the outer request context after authentication. The request middleware will instead place one private, request-local metadata holder in the original context. Authentication code writes the immutable Actor UUID once as `FINDUR_USER_ID`; the completion middleware and a delegating `slog.Handler` read its approved snapshot. This maintains a Go-style context propagation boundary without trusting caller data or requiring each logger call site to repeat attributes. The strict generated-handler error callback is the single location for HTTP application errors that escaped endpoint-level categorical handling, so it can log a safe failure category without exposing an arbitrary Go error string.

Illustrative completion record for an authenticated request:

```json
{"msg":"http request","REQUEST_ID":"...","FINDUR_USER_ID":"findur-uuid","METHOD":"GET","ROUTE":"portfolio_showcase","STATUS":200,"LATENCY_MS":12}
```

## Verification

**Commands:**
- `cd backend && go test ./internal/platform/httpapi ./cmd/findur` -- request metadata, authorization/error boundaries, and production logger composition pass.
- `cd backend && go test ./...` -- backend regression suite passes.
- `cd backend && golangci-lint run` -- mandatory lint verification passes.
- `git diff --check` -- changed files contain no whitespace errors.
