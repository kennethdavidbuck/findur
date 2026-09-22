DROP TABLE portfolio_sync_worker_lease;
DROP TABLE portfolio_account_sync_state;
DROP TABLE portfolio_account_activities;
ALTER TABLE portfolio_inclusion_changes DROP COLUMN sync_account_ids;

DELETE FROM portfolio_balance_versions version
WHERE NOT EXISTS (SELECT 1 FROM portfolio_balance_heads head WHERE head.version_id=version.id);
DELETE FROM portfolio_position_versions version
WHERE NOT EXISTS (SELECT 1 FROM portfolio_position_heads head WHERE head.version_id=version.id);
DELETE FROM portfolio_activity_versions version
WHERE NOT EXISTS (SELECT 1 FROM portfolio_activity_heads head WHERE head.version_id=version.id);

ALTER TABLE portfolio_balance_versions ADD UNIQUE (user_id,account_id,inclusion_version);
ALTER TABLE portfolio_position_versions ADD UNIQUE (user_id,account_id,inclusion_version);
ALTER TABLE portfolio_activity_versions ADD UNIQUE (user_id,account_id,inclusion_version);
