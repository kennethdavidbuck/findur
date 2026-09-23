#!/bin/sh
set -eu

scenario=${1:-}

case "$scenario" in
  declined)
    curl --fail --silent --show-error \
      --request POST \
      http://127.0.0.1:8080/api/__fixture/oidc/deny-next
    echo "The next SnapTrade login will be declined once."
    ;;
  findur-access-revoked|revoked)
    scenario_sql="
      UPDATE provider_authorizations
      SET lifecycle_status = 'reauthorization-required';

      UPDATE portfolio_inventory_state
      SET current_status = 'unauthorized',
          retry_at = NULL,
          claim_expires_at = NULL,
          failure_reason = 'authorization_required',
          failure_action = 'reconnect',
          updated_at = now();
    "
    docker compose exec -T postgres psql --username findur --dbname findur --command "$scenario_sql"
    echo "Findur's SnapTrade permission is revoked. Reload the app to end the Findur session and see the reconnect message."
    ;;
  connection-disconnected|disabled)
    scenario_sql="
      UPDATE portfolio_inventory_connections connection
      SET status = 'disabled',
          available = false,
          eligible = false,
          diagnostic_reason = 'connection_disabled',
          diagnostic_action = 'reconnect'
      FROM portfolio_inventory_state state
      WHERE connection.user_id = state.user_id
        AND connection.generation = state.head_generation
        AND connection.connection_id = '0eadf3ab-8daa-4357-9e38-325dbe82f003';

      UPDATE portfolio_inventory_accounts account
      SET available = false,
          eligible = false,
          selectable = false,
          usability_reason = 'connection_disabled'
      FROM portfolio_inventory_state state
      WHERE account.user_id = state.user_id
        AND account.generation = state.head_generation
        AND account.account_id = '03867fbb-41b4-4a05-8815-c96f94f8ba6b';

      UPDATE portfolio_inventory_state
      SET updated_at = now();
    "
    docker compose exec -T postgres psql --username findur --dbname findur --command "$scenario_sql"
    echo "Synthetic Self-Directed is now disconnected. Reload /portfolio to see the error for Healthy Realtime — Full Data (•••• X001)."
    ;;
  reset)
    scenario_sql="
      UPDATE provider_authorizations
      SET lifecycle_status = 'active';

      UPDATE portfolio_inventory_connections connection
      SET status = 'active',
          available = true,
          eligible = true,
          diagnostic_reason = NULL,
          diagnostic_action = NULL,
          brokerage_label = CASE
              WHEN connection.connection_id = 'a0000000-0000-4000-8000-000000000020'
                  THEN 'Synthetic Secondary Brokerage'
              ELSE connection.brokerage_label
          END
      FROM portfolio_inventory_state state
      WHERE connection.user_id = state.user_id
        AND connection.generation = state.head_generation
        AND connection.connection_id IN (
            '0eadf3ab-8daa-4357-9e38-325dbe82f003',
            'a0000000-0000-4000-8000-000000000020'
        );

      UPDATE portfolio_inventory_accounts account
      SET available = true,
          eligible = true,
          selectable = true,
          usability_reason = 'ready',
          masked_label = CASE
              WHEN account.account_id = 'b0000000-0000-4000-8000-000000000020'
                  THEN 'Secondary Investment Account (•••• 0023)'
              ELSE account.masked_label
          END
      FROM portfolio_inventory_state state
      WHERE account.user_id = state.user_id
        AND account.generation = state.head_generation
        AND account.account_id IN (
            '03867fbb-41b4-4a05-8815-c96f94f8ba6b',
            'b0000000-0000-4000-8000-000000000020'
        );

      UPDATE portfolio_inventory_state state
      SET current_status = COALESCE(
              (
                SELECT version.status
                FROM portfolio_inventory_versions version
                WHERE version.user_id = state.user_id
                  AND version.generation = state.head_generation
              ),
              'empty'
          ),
          retry_at = NULL,
          claim_expires_at = NULL,
          failure_reason = NULL,
          failure_action = NULL,
          updated_at = now();
    "
    docker compose exec -T postgres psql --username findur --dbname findur --command "$scenario_sql"
    echo "Authorization recovery scenarios were reset. Reload the page to see the healthy synthetic connection."
    ;;
  *)
    echo "Usage: $0 {declined|findur-access-revoked|connection-disconnected|reset}" >&2
    exit 2
    ;;
esac
