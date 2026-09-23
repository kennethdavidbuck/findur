# Resource failure modes

## Domain contract

Add a reusable diagnostic shape for successful resource responses. Names may be adjusted to fit generated-code conventions, but semantics and closed values are normative.

```yaml
ResourceDiagnostic:
  type: object
  additionalProperties: false
  required: [reason, recommendedAction]
  properties:
    reason:
      $ref: '#/components/schemas/ResourceDiagnosticReason'
    recommendedAction:
      type: string
      enum: [none, wait, retry, reconnect]
    retryAt:
      type: string
      format: date-time
    lastSuccessfulAt:
      type: string
      format: date-time
```

Attach an optional `diagnostic` to `InventoryConnection` and `DatasetContext`. The status or freshness remains the primary lifecycle field; the diagnostic explains the safe cause and action. Do not duplicate the failed-request `BrowserError` envelope inside successful payloads.

## Connection matrix

| Reason | Meaning | Action | Required user information |
|---|---|---|---|
| `no_accounts_returned` | Connection succeeded but returned no accounts | `retry` or `none` according to provider semantics | Connected; no accounts were returned; do not imply an outage. |
| `no_supported_accounts` | Accounts were returned but none meet current Findur eligibility | `none` | Connection works; Findur cannot use the returned account types. |
| `connection_disabled` | Provider marks the connection as requiring repair | `reconnect` | Name the affected institution and say repair is required. |
| `authorization_required` | Authorization is absent, expired, or rejected | `reconnect` | State that renewed authorization is required; do not call this a transient outage. |
| `provider_unavailable` | Provider or brokerage access failed transiently | `retry` or `wait` | State that the connection could not be checked; show authoritative retry timing when present. |
| `sync_pending` | Provider data is still being prepared | `wait` | State that preparation continues and give a check-again action or timing. |
| `unknown` | Safe cause cannot be determined | `retry` | Admit that Findur could not determine the cause; do not guess. |

An active connection with usable accounts normally has no diagnostic. Account-level `usabilityReason` remains authoritative for individual account selection and must retain distinct copy.

## Account-setup summary

Derive a summary from the complete response:

- `ready` with selectable accounts: state how many accounts are ready.
- mixed results: state that some connections need attention, give checked/affected counts, and avoid “couldn’t load your accounts.”
- no returned accounts: state that connections were checked but no accounts were returned.
- returned but unsupported accounts: state that connected accounts are not supported by current eligibility policy.
- all repair or authorization failures: state that reconnection is required.
- all transient failures: state that no connections could be checked and provide global retry or wait guidance.
- retained committed choices: state explicitly that saved choices or previously saved evidence have not been removed when true.

Display `PortfolioInventory.updatedAt` as “Last checked” using the active locale. If `retryAt` is present, display the exact localized time. Never infer that a provider will recover by a particular time without an authoritative value.

Per-connection actions are preferred. A global action is allowed only when it correctly applies to every affected connection or the request itself failed.

## Dataset matrix

Every balances, positions, and activities dataset renders its state independently, including when it contains zero rows.

| Freshness / reason | Values | Presentation | Action |
|---|---|---|---|
| `current` | Show | Current with observation or retrieval time | None |
| `stale_usable` | Show | Saved values are older but still usable; show last success | Optional retry/check |
| `expired` | Hide under current policy | Saved values expired and are hidden; name the affected dataset | Reconnect only when diagnostic requires it; otherwise retry |
| `unavailable` + `sync_pending` | Hide | Dataset is still syncing | Wait/check again |
| `unavailable` + `authorization_required` | Hide | Authorization must be renewed | Reconnect |
| `unavailable` + `provider_unavailable` | Hide | Dataset could not be refreshed; show last success and retry timing when present | Retry/wait |
| `unavailable` + `unknown` | Hide | Cause could not be determined | Retry |

Remove row-count gating that suppresses a dataset section before its freshness is evaluated. In particular, unavailable or expired positions with zero rows must still render the diagnostic.

## Inclusion matrix

Use the existing `failureReason` values rather than changing the durable inclusion contract unless implementation evidence requires added metadata.

| Failure reason | Explanation | Action |
|---|---|---|
| `authorization_required` | One or more additions require renewed provider authorization; unchanged saved choices remain intact | Reconnect |
| `rate_limited` | The provider deferred the update | Wait or retry at an authoritative time |
| `provider_unavailable` | The provider could not complete one or more additions | Retry |
| `unusable_data` | One or more selected accounts no longer have usable data | Review affected current choices; do not blindly resubmit hidden accounts |
| `stale_guard` | Accounts changed while the save was being processed | Review refreshed choices and save again |

The UI must identify affected accounts when the current response safely contains those account IDs and labels. It must distinguish failed additions from effective removals and preserve the current durable selection.

## Persistence and lifecycle

- First determine which facts already exist in inventory versions, account sync state, inclusion changes, provider normalization, and publication timestamps.
- Add persistence only for facts that must survive reload or scheduled execution and cannot be derived safely.
- Any new columns have bounded values or checks, are owner-scoped through their parent resource, and include reversible up/down migrations.
- A scheduled transient failure preserves the last successful snapshot and records the new diagnostic and retry information without presenting retained data as fresh.
- A later success clears obsolete transient diagnostics atomically with publication.
- Authorization generation and existing lease guards continue preventing stale workers from publishing misleading state.

## Required tests

- Backend table-driven normalization tests cover every connection and dataset reason.
- Repository tests cover persistence, clearing on success, scheduled retries, stale publication guards, and owner isolation.
- OpenAPI response tests prove valid diagnostics and reject unknown or unsafe fields.
- Frontend table-driven tests cover every connection, dataset, usability, and inclusion reason in English and French.
- Component tests cover mixed connection results, all-failed results, retained choices, timestamps, per-item actions, zero-row unavailable positions, and safe unknown fallbacks.
- Accessibility tests verify blocking changes use alerts, progress uses status semantics, affected controls have contextual names, and focus reaches the appropriate recovery region.
- Browser integration reproduces the supplied mixed-state scenario and demonstrates understandable summary, explanations, and recovery actions.

## Verification

Run and report, at minimum:

```text
cd frontend && npm run generate:api
cd frontend && npm run typecheck
cd frontend && npm run lint
cd frontend && npm test -- --run
cd frontend && npm run build
cd backend && go test ./...
cd backend && golangci-lint run
```

Also run the relevant inventory, inclusion, synchronization, Showcase, and browser integration scenarios. Report any test not run rather than implying coverage.
