# Failure contract

## Ownership

- The API owns stable categories, retry facts, and safe correlation metadata.
- The frontend owns localized title, explanation, severity, announcement behavior, and action labels.
- Domain resources own degraded state returned in successful responses; they do not use this envelope as an embedded substitute.

## OpenAPI envelope

Replace the current generic `Error` response body with one shared schema used by all documented non-2xx application responses:

```yaml
BrowserError:
  type: object
  additionalProperties: false
  required: [code, retryable, recommendedAction]
  properties:
    code:
      $ref: '#/components/schemas/BrowserErrorCode'
    retryable:
      type: boolean
    recommendedAction:
      type: string
      enum: [none, retry, reconnect, reload, sign_in, review_input]
    retryAt:
      type: string
      format: date-time
    supportReference:
      type: string
      minLength: 1
      maxLength: 64
    invalidFields:
      type: array
      maxItems: 32
      uniqueItems: true
      items:
        type: string
        minLength: 1
        maxLength: 64
```

`retryAt` is present only when the server can provide an authoritative bound. `supportReference` must be opaque, non-sensitive, and useful to operators. `invalidFields` is present only for bounded validation failures; endpoint behavior further restricts its allowed values. `recommendedAction` is advice, not authority: the frontend suppresses an action it cannot safely execute in the current context.

Retain existing codes initially: `invalid_request`, `authorization_unavailable`, `initialization_failed`, `restart_required`, `unauthenticated`, `forbidden`, `conflict`, `invalid_selection`, and `invalid_profile`. Rename the schema without changing these wire values. Any code addition requires synchronized backend behavior, generated artifacts, frontend presentation, and tests.

Initial defaults are:

| Code | Retryable | Recommended action | Additional facts |
|---|---:|---|---|
| `invalid_request` | No | `none` | None |
| `authorization_unavailable` | Yes | `retry` | `retryAt` when authoritative |
| `initialization_failed` | Yes | `retry` | `supportReference` when available |
| `restart_required` | Yes | `sign_in` | None |
| `unauthenticated` | Yes | `sign_in` | None |
| `forbidden` | Yes | `sign_in` | None |
| `conflict` | Yes | `reload` | None |
| `invalid_selection` | No | `review_input` | `invalidFields` only if safely applicable |
| `invalid_profile` | No | `review_input` | Bounded `invalidFields` |

Here, `retryable` means the user can safely attempt the operation again after following the recommended action; it does not authorize automatic replay of unsafe requests.

## Client-only failures

The frontend adds typed local codes that are never claimed to have come from the server:

| Code | Meaning | Default action |
|---|---|---|
| `network_unavailable` | Fetch rejected before a valid response | `retry` |
| `request_timeout` | A bounded client request timed out | `retry` |
| `malformed_response` | A response violated its declared contract | `retry` |
| `unknown_failure` | No safe known classification applies | `retry` when the operation is idempotent; otherwise `none` |

Request helpers return or throw typed failures carrying only the normalized fields. Page components must not branch on free-form exception messages.

## Frontend presentation

Create one presentation catalog keyed exhaustively by the union of OpenAPI and client-only codes. Each entry supplies:

- English and French title and explanation keys;
- severity (`info`, `warning`, or `error`);
- default announcement behavior;
- default action label and intent;
- whether existing form or page state must be retained.

The catalog must use an exhaustive TypeScript construct such as `satisfies Record<UiFailureCode, ErrorPresentation>`. Runtime parsing must tolerate an unknown server code and map it to `unknown_failure`.

Do not build one universal visual component that erases context. Pages may render their own layout, but classification and presentation lookup must be shared.

## Required flow changes

| Flow | Required behavior |
|---|---|
| Authorization status | Distinguish known unavailability from transport, timeout, and malformed response; retain a safe login-unavailable fallback. |
| Authorization start | Submit in-app, handle `400`, `415`, and `503`, render a localized alert, retain the consent page, and follow the provider redirect only on success. Raw JSON navigation is forbidden. |
| Profile load/save | Keep field validation and conflict recovery; use normalized request failures for all other outcomes and preserve entries. |
| Display preferences | Show a visible, localized failure with an explicit retry path while retaining the locally selected value and marking it unsaved. |
| Inventory and inclusion | Normalize failed requests without replacing successful degraded domain states. |
| Showcase | Treat `401` and `403` consistently with authenticated-session recovery and classify other request failures. |
| Logout | Preserve the current active-session warning and retry behavior through the shared classification path. |

## Accessibility rules

- A newly completed failure that blocks the current task uses `role="alert"`; ongoing, pending, or non-blocking information uses `role="status"` or a polite live region.
- Focus moves to a blocking error summary or recovery control only when that movement helps the user act and does not interrupt an in-progress control.
- Actions have specific accessible names such as “Retry loading profile”; repeated generic “Try again” controls require contextual labelling.
- Color and icons never carry the category alone.
- Form values and durable selections remain intact unless security or confirmed session expiry requires leaving the screen.

## Enforcement and documentation

- Regenerate backend bindings and `frontend/src/generated/api.ts` from OpenAPI and verify the generated diff.
- Add backend table-driven tests for every error code’s status, envelope fields, cache headers, and absence of unsafe detail.
- Add frontend table-driven tests for every presentation entry in both locales and runtime fallback for an unknown code.
- Add component tests for the flows in the required-flow matrix, including focus and accessibility roles.
- Add a browser test proving authorization-start failure remains on the consent page and does not display raw JSON.
- Add a short project document under `docs/` describing ownership, schema, and extension steps.
- Update the managed `AGENTS.md` context through `bmad-project-context`: new user-visible failure modes must update OpenAPI or client-only types, both locales, exhaustive mappings, and tests.
- Update `.github/pull_request_template.md` with a failure-mode section covering new codes, degraded states, localization, accessibility, and verification; use `Not applicable` when justified.

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

Run the relevant browser integration test for authorization initiation. If generation is owned by a Go `generate` command, use the repository command and confirm both Go and TypeScript generated artifacts are current.
