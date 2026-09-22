DO $$
DECLARE table_name text;
DECLARE constraint_name name;
BEGIN
    FOREACH table_name IN ARRAY ARRAY['portfolio_balance_versions','portfolio_position_versions','portfolio_activity_versions'] LOOP
        SELECT conname INTO STRICT constraint_name
        FROM pg_constraint
        WHERE conrelid=table_name::regclass AND contype='u';
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I',table_name,constraint_name);
    END LOOP;
END $$;

ALTER TABLE portfolio_inclusion_changes ADD COLUMN sync_account_ids text[] NOT NULL DEFAULT '{}';

CREATE TABLE portfolio_account_activities (
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    activity_id text NOT NULL CHECK (activity_id <> '' AND length(activity_id) <= 160),
    activity_type text NOT NULL CHECK (activity_type <> '' AND length(activity_type) <= 80),
    trade_date timestamptz,
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    amount numeric,
    fee numeric,
    price numeric,
    units numeric,
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, account_id, activity_id),
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE
);

INSERT INTO portfolio_account_activities
    (user_id,account_id,activity_id,activity_type,trade_date,currency,amount,fee,price,units,first_seen_at,last_seen_at)
SELECT DISTINCT ON (version.user_id,version.account_id,row.activity_id)
       version.user_id,version.account_id,row.activity_id,row.activity_type,row.trade_date,row.currency,row.amount,row.fee,row.price,row.units,
       min(version.published_at) OVER (PARTITION BY version.user_id,version.account_id,row.activity_id),
       max(version.published_at) OVER (PARTITION BY version.user_id,version.account_id,row.activity_id)
FROM portfolio_activity_rows row
JOIN portfolio_activity_versions version ON version.id=row.version_id
ORDER BY version.user_id,version.account_id,row.activity_id,version.published_at DESC;

CREATE INDEX portfolio_account_activities_newest_idx
    ON portfolio_account_activities (user_id,account_id,trade_date DESC NULLS LAST,activity_id DESC);

CREATE TABLE portfolio_account_sync_state (
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    initialized_at timestamptz,
    last_success_at timestamptz,
	balances_success_at timestamptz,
	positions_success_at timestamptz,
	activities_success_at timestamptz,
    next_attempt_at timestamptz,
    failure_count integer NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    claim_id uuid,
    claim_expires_at timestamptz,
    claimed_inclusion_version bigint,
    claimed_lifecycle_generation bigint,
    claimed_inventory_generation bigint,
	claimed_resource text CHECK (claimed_resource IN ('balances','positions','activities')),
	claimed_change_id uuid,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (user_id,account_id),
    FOREIGN KEY (user_id,account_id) REFERENCES portfolio_account_identities(user_id,account_id) ON DELETE CASCADE
);

INSERT INTO portfolio_account_sync_state
    (user_id,account_id,initialized_at,last_success_at,balances_success_at,positions_success_at,activities_success_at,updated_at)
SELECT balance.user_id,balance.account_id,
       LEAST(balance.published_at,position.published_at,activity.published_at),
       GREATEST(balance.published_at,position.published_at,activity.published_at),
       GREATEST(balance.published_at,position.published_at,activity.published_at),
       GREATEST(balance.published_at,position.published_at,activity.published_at),
       GREATEST(balance.published_at,position.published_at,activity.published_at),
       GREATEST(balance.published_at,position.published_at,activity.published_at)
FROM portfolio_balance_heads balance_head
JOIN portfolio_balance_versions balance
  ON balance.id=balance_head.version_id
 AND balance.user_id=balance_head.user_id
 AND balance.account_id=balance_head.account_id
JOIN portfolio_position_heads position_head
  ON position_head.user_id=balance_head.user_id
 AND position_head.account_id=balance_head.account_id
JOIN portfolio_position_versions position
  ON position.id=position_head.version_id
 AND position.user_id=position_head.user_id
 AND position.account_id=position_head.account_id
JOIN portfolio_activity_heads activity_head
  ON activity_head.user_id=balance_head.user_id
 AND activity_head.account_id=balance_head.account_id
JOIN portfolio_activity_versions activity
  ON activity.id=activity_head.version_id
 AND activity.user_id=activity_head.user_id
 AND activity.account_id=activity_head.account_id;

INSERT INTO portfolio_account_sync_state (user_id,account_id,updated_at)
SELECT user_id,account_id,last_seen_at FROM portfolio_account_identities
ON CONFLICT (user_id,account_id) DO NOTHING;

CREATE INDEX portfolio_account_sync_due_idx
    ON portfolio_account_sync_state (last_success_at,next_attempt_at,claim_expires_at);

CREATE TABLE portfolio_sync_worker_lease (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    claim_id uuid,
    claim_expires_at timestamptz,
    updated_at timestamptz NOT NULL
);
