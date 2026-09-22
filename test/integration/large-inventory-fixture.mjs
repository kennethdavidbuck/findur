// Deterministic, entirely invented provider data. No captured customer payloads.
// Imported by the composed browser test, or run directly to export the fixture.
import { pathToFileURL } from 'node:url'

const id = (kind, index) => `${kind}0000000-0000-4000-8000-${String(index + 1).padStart(12, '0')}`

export function largeInventoryFixture() {
  const connections = Array.from({ length: 50 }, (_, index) => ({
    id: id('a', index),
    disabled: index === 49,
    brokerage: { display_name: `Synthetic Broker ${String(index + 1).padStart(2, '0')}` },
    data_freshness_mode: { institution: 'delayed', snaptrade: 'delayed' },
  }))
  const accounts = connections.flatMap((connection, connectionIndex) => Array.from({ length: 20 }, (_, slot) => {
    const index = connectionIndex * 20 + slot
    return {
      id: id('b', index),
      brokerage_authorization: connection.id,
      portfolio_group: id('c', index),
      name: `Synthetic Account ${String(index + 1).padStart(4, '0')}`,
      number: slot === 4 ? `*****${index + 4001}` : `FAKE-${String(index + 1).padStart(8, '0')}`,
      institution_account_id: `INVENTED-${index + 1}`,
      institution_name: connection.brokerage.display_name,
      meta: slot === 6 ? null : { currency: 'USD', institution_name: connection.brokerage.display_name },
      cash_restrictions: [],
      created_date: '2026-01-01T00:00:00Z', funding_date: null, opening_date: null,
      sync_status: {
        holdings: slot === 11 ? null : { last_successful_sync: '2026-09-20T00:00:00Z', initial_sync_completed: slot === 9 ? null : slot === 13 ? undefined : slot !== 3, holdings_unavailable: slot === 8 },
        transactions: { last_successful_sync: '2026-09-20', first_transaction_date: null, initial_sync_completed: slot !== 14 },
      },
      balance: { total: { amount: 100000 + index, currency: 'USD' } },
      raw_data: null,
      raw_type: slot === 6 ? null : slot === 15 ? 'Cash' : 'NP',
      status: slot === 0 ? 'closed' : slot === 7 ? 'archived' : slot === 10 ? 'unavailable' : slot === 4 ? null : 'open',
      is_paper: true,
      account_category: slot === 1 ? 'LOC' : slot === 2 ? 'DEPOSIT' : slot === 5 ? null : slot === 12 ? 'FUTURE_CATEGORY' : 'INVESTMENT',
    }
  }))
  return {
    connections, accounts,
    expected: { connections: 50, accounts: 1000, normalized: 392, eligible: 392, selectable: 392, visible: 392, disabled: 0, hidden: 608,
      // Independent oracle: these slots alone satisfy the documented rules.
      accountIds: connections.slice(0, 49).flatMap((_, index) => [4, 6, 14, 15, 16, 17, 18, 19].map((slot) => id('b', index * 20 + slot))) },
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  process.stdout.write(`${JSON.stringify(largeInventoryFixture(), null, 2)}\n`)
}
