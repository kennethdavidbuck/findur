ALTER TABLE portfolio_inventory_connections
    ADD COLUMN diagnostic_reason text
        CHECK (diagnostic_reason IN (
            'no_accounts_returned',
            'no_supported_accounts',
            'connection_disabled',
            'authorization_required',
            'provider_unavailable',
            'sync_pending',
            'unknown'
        )),
    ADD COLUMN diagnostic_action text
        CHECK (diagnostic_action IN ('none', 'wait', 'retry', 'reconnect'));

ALTER TABLE portfolio_inventory_state
    ADD COLUMN failure_reason text
        CHECK (failure_reason IN ('authorization_required', 'provider_unavailable', 'sync_pending', 'unknown')),
    ADD COLUMN failure_action text
        CHECK (failure_action IN ('none', 'wait', 'retry', 'reconnect')),
    ADD COLUMN diagnostic_refresh_needed boolean NOT NULL DEFAULT false;

UPDATE portfolio_inventory_state
SET diagnostic_refresh_needed = true
WHERE head_generation IS NOT NULL
    AND current_status IN ('ready', 'empty');

ALTER TABLE portfolio_account_sync_state
    ADD COLUMN balances_failure_reason text
        CHECK (balances_failure_reason IN ('authorization_required', 'provider_unavailable', 'sync_pending', 'unknown')),
    ADD COLUMN balances_failure_action text
        CHECK (balances_failure_action IN ('wait', 'retry', 'reconnect')),
    ADD COLUMN balances_retry_at timestamptz,
    ADD COLUMN positions_failure_reason text
        CHECK (positions_failure_reason IN ('authorization_required', 'provider_unavailable', 'sync_pending', 'unknown')),
    ADD COLUMN positions_failure_action text
        CHECK (positions_failure_action IN ('wait', 'retry', 'reconnect')),
    ADD COLUMN positions_retry_at timestamptz,
    ADD COLUMN activities_failure_reason text
        CHECK (activities_failure_reason IN ('authorization_required', 'provider_unavailable', 'sync_pending', 'unknown')),
    ADD COLUMN activities_failure_action text
        CHECK (activities_failure_action IN ('wait', 'retry', 'reconnect')),
    ADD COLUMN activities_retry_at timestamptz;
