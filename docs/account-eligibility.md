# Brokerage account eligibility

These rules govern the account chooser and admission of new account selections.
They were agreed on 2026-09-21. The server applies them; the browser receives the
filtered inventory and does not infer investment eligibility from account names.

An account is included only when **all** these conditions hold:

| Field | Required value | Treatment of missing or other values |
| --- | --- | --- |
| `account_category` | `INVESTMENT` | Exclude `DEPOSIT`, `LOC`, null, missing, and unrecognized categories. |
| `status` | `open` or null | Exclude `closed`, `archived`, and `unavailable`. Missing status is treated as null. An unrecognized non-null status violates the pinned response contract and fails validation. |
| `sync_status.holdings.initial_sync_completed` | `true` | Exclude false, null, missing flag, or missing holdings metadata. |
| `sync_status.holdings.holdings_unavailable` | Anything except `true` | Explicit true excludes the account even if initial sync is complete. The optional flag may be absent. |
| Referenced connection | Known, active, with `disabled: false` | Exclude disabled connections. Missing, duplicate, or invalid connection identity fails validation. |

Null account status is accepted only when every other rule passes. It does not
make an unknown account category acceptable. Accepted accounts are both eligible
and selectable, including accepted null-status accounts.

The following are **not** eligibility requirements: account name, `raw_type`,
deprecated `meta`, a positive balance, transaction-sync completion, opening or
funding dates, and a recent sync timestamp. An investment account holding only
cash can qualify. Paper accounts are not additionally filtered by this policy;
the integration fixture intentionally uses simulated accounts.

## Request and persistence behavior

- Discovery uses one `GET /authorizations` and one `GET /accounts` when an active
  connection exists. Empty or entirely disabled connections need no account call.
- `/accounts` returns SnapTrade’s daily cached inventory, including on real-time
  plans. Explicit retry reloads that provider cache; it does not force a fresh
  brokerage sync.
- Filtering a valid response is normal success, not a partial-loading failure.
  If no accounts qualify, return the existing empty state; entirely disabled
  connections retain their disabled state.
- A failed or malformed bulk request does not publish a partially loaded
  inventory. This includes invalid JSON/field types, invalid or duplicate account
  IDs, mismatched connection references, and unexpected response envelopes.
- Existing persisted inventory is filtered by its normalized investment category,
  availability, completed holdings sync, and active connection before display.
  New additions are checked against those same requirements at confirmation so a
  legacy `selectable` flag cannot admit an excluded account.
- Previously committed membership and API removal behavior are unchanged. This
  does not automatically delete existing selections or financial datasets.
  Excluded accounts are absent from the chooser, including previously committed
  ones; removing a committed account absent from inventory through the UI is an
  existing deferred lifecycle concern, not solved by this filter.

## Verification

Provider table tests cover each condition and precedence, including unavailable
holdings overriding completed sync, null status with invalid category/sync,
disabled connections, malformed payloads, and duplicate identities. Repository
tests cover older persisted flags and admission of new selections.

The WireMock fixture has 1,000 synthetic accounts across 50 connections, of which
one is disabled. Exactly eight slots in each of the 49 active connections qualify:
4, 6, and 14 through 19 (zero-based). Thus **392 accounts are returned and
selectable; 608 are excluded**. The browser test checks exact IDs and grouping,
not just counts. See [the integration guide](../test/integration/README.md).

## Provider references

SnapTrade documents normalized categories, nullable status, and holdings flags in
[List user accounts](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_listUserAccounts).
Connection availability is documented in
[List all connections](https://docs.snaptrade.com/reference/Connections/Connections_listBrokerageAuthorizations).
These fields support the policy above; SnapTrade does not prescribe every Findur
eligibility decision. The separate category-filtering guide uses `kind` for a
different endpoint shape; this integration uses `account_category` from `/accounts`.
