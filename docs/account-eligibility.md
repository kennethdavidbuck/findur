# Brokerage account eligibility

These rules govern the account chooser and admission of new account selections.
They were agreed on 2026-09-21. The server applies them; the browser receives the
filtered inventory and does not infer investment eligibility from account names.

An account is included only when **all** these conditions hold:

| Field | Required value | Treatment of missing or other values |
| --- | --- | --- |
| `account_category` | `INVESTMENT`, null, or missing | Accept null or missing provisionally. Exclude `DEPOSIT`, `LOC`, and unrecognized non-null categories. |
| `status` | `open` or null | Exclude `closed`, `archived`, and `unavailable`. Missing status is treated as null. An unrecognized non-null status violates the pinned response contract and fails validation. |
| `sync_status.holdings.initial_sync_completed` | `true` | Exclude false, null, missing flag, or missing holdings metadata. |
| `sync_status.holdings.holdings_unavailable` | Anything except `true` | Explicit true excludes the account even if initial sync is complete. The optional flag may be absent. |
| Referenced connection | Known, active, with `disabled: false` | Exclude disabled connections. Missing, duplicate, or invalid connection identity fails validation. |

Null account status is accepted only when every other rule passes. Accounts with
null or missing category remain provisionally selectable because SnapTrade can
omit the normalized category for real investment accounts; they are not asserted
as positively eligible. Explicit deposit and line-of-credit categories remain
excluded.

The following are **not** eligibility requirements: account name, `raw_type`,
deprecated `meta`, a positive balance, transaction-sync completion, opening or
funding dates, and a recent sync timestamp. An investment account holding only
cash can qualify. Paper accounts are not additionally filtered by this policy;
the integration fixture intentionally uses simulated accounts.

## Request and persistence behavior

- A returning login rotates credentials and fences stale in-flight work without
  synchronously calling SnapTrade. Independently, the minute worker repairs
  incomplete inventory and refreshes successful connection/account inventory
  every 24 hours.
- Discovery uses one `GET /authorizations` and one `GET /accounts` when an active
  connection exists. Empty or entirely disabled connections need no account call.
- Each immutable account inventory row retains the provider-denominated
  `/accounts.balance.total` amount and currency when supplied. This is distinct
  from per-currency cash and buying-power rows fetched from `/balances`.
- `/accounts` returns SnapTrade’s daily cached inventory, including on real-time
  plans. Explicit retry reloads that provider cache; it does not force a fresh
  brokerage sync.
- Filtering a valid response is normal success, not a partial-loading failure.
  If no accounts qualify, return the existing empty state; entirely disabled
  connections retain their disabled state.
- A failed or malformed bulk request does not publish a partially loaded
  inventory. This includes invalid JSON/field types, invalid or duplicate account
  IDs, mismatched connection references, and unexpected response envelopes.
- Existing persisted inventory is filtered by its normalized or provisional
  category, availability, completed holdings sync, and active connection before display.
  New additions are checked against those same requirements at confirmation so a
  legacy `selectable` flag cannot admit an excluded account.
- At most five accounts can be selected. Both the browser and API enforce the
  limit before provider work.
- A valid selection save commits membership without provider calls. Selected
  account identities appear immediately in the Portfolio; first-time datasets
  are labelled as syncing until the minute worker completes balances,
  positions, and activities for every newly selected account.
- Removing an account deletes only its inclusion membership. Its synchronized
  datasets remain retained but hidden, and re-inclusion reuses them without an
  eager provider call.
- Included accounts are refreshed when their complete bundle is at least 24
  hours old. A minute worker uses durable database leases, bounded backoff, a
  45-second pass limit, and a shared paced/circuit-broken SnapTrade client.
- Each leased worker pass refreshes at most one due user inventory before it
  drains included-account resource work. Inventory refresh preserves saved
  historical resource data by stable account identity. An included account that
  no longer meets current eligibility is removed from active inclusion and
  receives no new resource work; its retained identity and history are not
  deleted.
- Balances and positions replace their current snapshots. Activities append by
  stable provider ID from the newest 50 returned rows, and the Showcase exposes
  only its newest accumulated 50 rows.
  Accounts excluded by current eligibility are absent from the chooser.
  Eligibility filtering does not silently rewrite an existing inclusion;
  membership changes only through an explicit selection save.

## Verification

Provider table tests cover each condition and precedence, including unavailable
holdings overriding completed sync, null status with invalid category/sync,
disabled connections, malformed payloads, and duplicate identities. Repository
tests cover older persisted flags and admission of new selections.

The WireMock fixture has 1,000 synthetic accounts across 50 connections, of which
one is disabled. Exactly nine slots in each of the 49 active connections qualify:
4 through 6 and 14 through 19 (zero-based). Thus **441 accounts are returned and
selectable; 559 are excluded**. Of the returned accounts, 392 have an explicit
investment category and 49 have a provisional null category. The browser test
checks exact IDs and grouping, not just counts. See the
[integration guide](../test/integration/README.md).

## Provider references

SnapTrade documents normalized categories, nullable status, and holdings flags in
[List user accounts](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_listUserAccounts).
Connection availability is documented in
[List all connections](https://docs.snaptrade.com/reference/Connections/Connections_listBrokerageAuthorizations).
These fields support the policy above; SnapTrade does not prescribe every Findur
eligibility decision. The separate category-filtering guide uses `kind` for a
different endpoint shape; this integration uses `account_category` from `/accounts`.
