ALTER TABLE portfolio_inventory_state
    ADD COLUMN failure_count integer NOT NULL DEFAULT 0 CHECK (failure_count >= 0);

ALTER TABLE portfolio_inventory_accounts
    ADD COLUMN total_balance_amount numeric,
    ADD COLUMN total_balance_currency char(3)
        CHECK (total_balance_currency ~ '^[A-Z]{3}$'),
    ADD CHECK (
        (total_balance_amount IS NULL AND total_balance_currency IS NULL)
        OR (total_balance_amount IS NOT NULL AND total_balance_currency IS NOT NULL)
    );

CREATE INDEX portfolio_inventory_due_idx
    ON portfolio_inventory_state (retry_at, updated_at, claim_expires_at);
