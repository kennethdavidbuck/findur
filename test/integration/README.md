The composed integration suite uses real Chromium through Selenium, the Go API,
PostgreSQL, and WireMock. Run it from the repository root:

```sh
./scripts/compose-test.sh
```

The account inventory stress scenario runs automatically after the profile browser
journey. `large-inventory-fixture.mjs` deterministically creates 1,000 invented
accounts across 50 connections. The scenario temporarily installs those complete
arrays in WireMock for `GET /accounts` and `GET /authorizations`, then restores the
normal mappings and inventory. All identifiers, labels, and balances are invented;
the generator does not use the reported customer-like example.

Export the complete fixture for inspection or reuse:

```sh
node test/integration/large-inventory-fixture.mjs > /tmp/findur-large-inventory.json
```

The server rules are defined in [account eligibility](../../docs/account-eligibility.md).
Each connection contains 20 accounts; the final connection is disabled. The same
slot matrix repeats in each connection so every expected inclusion is explicit:

| Slot (zero-based) | Circumstance | Included on an active connection? |
| --- | --- | --- |
| 0 | Closed investment account | No |
| 1 | Credit account | No |
| 2 | Deposit account | No |
| 3 | Initial holdings sync false | No |
| 4 | Null status, investment, holdings ready, provider-masked number | Yes |
| 5 | Null category, otherwise ready | Yes, provisionally |
| 6 | Null raw type and deprecated metadata | Yes |
| 7 | Archived investment account | No |
| 8 | Holdings unavailable, despite completed sync | No |
| 9 | Initial holdings sync null | No |
| 10 | Account status unavailable | No |
| 11 | Holdings metadata missing/null | No |
| 12 | Unrecognized account category | No |
| 13 | Initial holdings sync flag missing | No |
| 14 | Transaction sync incomplete, holdings ready | Yes |
| 15 | Investment account with raw type Cash | Yes |
| 16–19 | Ordinary ready investment accounts | Yes |

The independent expected-ID list contains slots 4–6 and 14–19 in each of the
49 active connections: **441 accepted accounts**. All 20 accounts in the disabled
connection are excluded.

| Measurement | Expected count |
| --- | ---: |
| Provider accounts | 1,000 |
| Connections | 50 |
| Returned accounts | 441 |
| Positively eligible accounts | 392 |
| Selectable / visible accounts | 441 |
| Visible disabled rows | 0 |
| Accounts excluded on the server | 559 |

The test first verifies that active connections containing only excluded accounts
publish an empty inventory and display the matching empty-state message. It then
loads the mixed fixture, checks the persisted normalized response,
checks every account's connection assignment, complete masked label, and brokerage, loads
the chooser, selects all, deselects all, scrolls to and clicks the last selectable
account through WebDriver, and reviews all 441 selections. It verifies the request-journal delta is exactly
two provider calls throughout this journey, with no per-connection account calls
or financial-data calls. The smaller journey still covers saving through the UI
and rendering the Showcase. Large-selection confirmation and its provider-call
budget are a separate follow-up; this scenario stops at reviewing the draft.

The console prints `large inventory performance` with API fetch/normalization/
persistence duration, browser navigation/render duration, select-all, deselect-all,
individual-toggle, and review durations. Individual-toggle timing includes the
WebDriver round trip. Smoke budgets are 10 seconds for load/render and 5 seconds
for interactions. These generous thresholds detect stalls on local/CI machines;
the reported measurements are not a production load test or latency guarantee.
There is no pagination or virtualization change to the chooser in this work.

SnapTrade documents [list accounts](https://docs.snaptrade.com/reference/Account%20Information/AccountInformation_listUserAccounts)
and [list connections](https://docs.snaptrade.com/reference/Connections/Connections_listBrokerageAuthorizations)
as complete array responses without pagination parameters. The fixture therefore
returns every account in one response and asserts exactly two inventory calls.
Account activities use a different, paginated endpoint; their existing tests are
separate from this inventory scenario.
