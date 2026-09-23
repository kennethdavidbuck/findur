DROP INDEX portfolio_inventory_due_idx;

ALTER TABLE portfolio_inventory_accounts
    DROP COLUMN total_balance_amount,
    DROP COLUMN total_balance_currency;

ALTER TABLE portfolio_inventory_state
    DROP COLUMN failure_count;
