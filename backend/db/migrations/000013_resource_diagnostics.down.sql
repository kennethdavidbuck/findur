ALTER TABLE portfolio_account_sync_state
    DROP COLUMN activities_retry_at,
    DROP COLUMN activities_failure_action,
    DROP COLUMN activities_failure_reason,
    DROP COLUMN positions_retry_at,
    DROP COLUMN positions_failure_action,
    DROP COLUMN positions_failure_reason,
    DROP COLUMN balances_retry_at,
    DROP COLUMN balances_failure_action,
    DROP COLUMN balances_failure_reason;

ALTER TABLE portfolio_inventory_connections
    DROP COLUMN diagnostic_action,
    DROP COLUMN diagnostic_reason;

ALTER TABLE portfolio_inventory_state
    DROP COLUMN diagnostic_refresh_needed,
    DROP COLUMN failure_action,
    DROP COLUMN failure_reason;
