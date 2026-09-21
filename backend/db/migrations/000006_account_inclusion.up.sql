ALTER TABLE portfolio_inventory_accounts
    ADD COLUMN selectable boolean NOT NULL DEFAULT false,
    ADD COLUMN usability_reason text NOT NULL DEFAULT 'account_unavailable'
        CHECK (usability_reason IN ('ready', 'provisional_status', 'provisional_category', 'sync_pending', 'connection_disabled', 'connection_unavailable', 'account_closed', 'account_unavailable', 'unsupported_category', 'sync_unavailable'));

CREATE TABLE portfolio_account_identities (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id text NOT NULL CHECK (account_id <> '' AND length(account_id) <= 128),
    connection_id text NOT NULL CHECK (connection_id <> '' AND length(connection_id) <= 128),
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, account_id)
);

WITH account_sightings AS (
    SELECT
        account.user_id,
        account.account_id,
        account.connection_id,
        min(version.published_at) OVER (PARTITION BY account.user_id, account.account_id) AS first_seen_at,
        max(version.published_at) OVER (PARTITION BY account.user_id, account.account_id) AS last_seen_at,
        row_number() OVER (PARTITION BY account.user_id, account.account_id ORDER BY account.generation DESC) AS recency
    FROM portfolio_inventory_accounts account
    JOIN portfolio_inventory_versions version
      ON version.user_id = account.user_id
     AND version.generation = account.generation
)
INSERT INTO portfolio_account_identities (user_id, account_id, connection_id, first_seen_at, last_seen_at)
SELECT user_id, account_id, connection_id, first_seen_at, last_seen_at
FROM account_sightings
WHERE recency = 1;

UPDATE portfolio_inventory_accounts account
SET
    selectable = connection.status = 'active'
        AND connection.available
        AND account.available
        AND account.category IN ('investment', 'unknown')
        AND account.sync_state = 'complete',
    usability_reason = CASE
        WHEN connection.status = 'disabled' THEN 'connection_disabled'
        WHEN connection.status = 'unavailable' OR NOT connection.available THEN 'connection_unavailable'
        WHEN account.sync_state = 'unavailable' THEN 'sync_unavailable'
        WHEN NOT account.available THEN 'account_unavailable'
        WHEN account.category IN ('deposit', 'credit') THEN 'unsupported_category'
        WHEN account.sync_state <> 'complete' THEN 'sync_pending'
        WHEN account.eligible AND account.category = 'investment' THEN 'ready'
        WHEN account.category = 'unknown' THEN 'provisional_category'
        WHEN account.category = 'investment' THEN 'provisional_status'
        ELSE 'account_unavailable'
    END
FROM portfolio_inventory_connections connection
WHERE connection.user_id = account.user_id
  AND connection.generation = account.generation
  AND connection.connection_id = account.connection_id;

CREATE TABLE portfolio_inclusion_state (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    version bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    lifecycle_generation bigint NOT NULL DEFAULT 1 CHECK (lifecycle_generation > 0),
    updated_at timestamptz NOT NULL
);

CREATE TABLE portfolio_inclusion_changes (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idempotency_key text NOT NULL CHECK (idempotency_key <> '' AND length(idempotency_key) <= 200),
    expected_version bigint NOT NULL CHECK (expected_version >= 0),
    result_version bigint NOT NULL CHECK (result_version >= 0),
    inventory_generation bigint NOT NULL CHECK (inventory_generation > 0),
    lifecycle_generation bigint NOT NULL CHECK (lifecycle_generation > 0),
    target_account_ids text[] NOT NULL,
    addition_account_ids text[] NOT NULL,
    removal_account_ids text[] NOT NULL,
    status text NOT NULL CHECK (status IN ('pending', 'committed', 'failed')),
    failure_reason text CHECK (failure_reason IN ('authorization_required', 'rate_limited', 'provider_unavailable', 'unusable_data', 'stale_guard')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (user_id, idempotency_key)
);

CREATE INDEX portfolio_inclusion_changes_user_idx ON portfolio_inclusion_changes (user_id, created_at DESC);

CREATE TABLE portfolio_included_accounts (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id text NOT NULL,
    inclusion_version bigint NOT NULL CHECK (inclusion_version > 0),
    included_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, account_id),
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE
);

CREATE TABLE portfolio_balance_versions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    inclusion_version bigint NOT NULL CHECK (inclusion_version > 0),
    observed_at timestamptz,
    retrieved_at timestamptz NOT NULL,
    published_at timestamptz NOT NULL,
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE,
    UNIQUE (user_id, account_id, inclusion_version)
);

CREATE TABLE portfolio_balance_heads (
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    version_id uuid NOT NULL REFERENCES portfolio_balance_versions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, account_id),
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE
);

CREATE TABLE portfolio_balance_rows (
    version_id uuid NOT NULL REFERENCES portfolio_balance_versions(id) ON DELETE CASCADE,
    row_number integer NOT NULL CHECK (row_number >= 0),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    cash numeric,
    buying_power numeric,
    PRIMARY KEY (version_id, row_number)
);

CREATE TABLE portfolio_position_versions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    inclusion_version bigint NOT NULL CHECK (inclusion_version > 0),
    observed_at timestamptz NOT NULL,
    retrieved_at timestamptz NOT NULL,
    published_at timestamptz NOT NULL,
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE,
    UNIQUE (user_id, account_id, inclusion_version)
);

CREATE TABLE portfolio_position_heads (
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    version_id uuid NOT NULL REFERENCES portfolio_position_versions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, account_id),
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE
);

CREATE TABLE portfolio_position_rows (
    version_id uuid NOT NULL REFERENCES portfolio_position_versions(id) ON DELETE CASCADE,
    row_number integer NOT NULL CHECK (row_number >= 0),
    instrument_id text NOT NULL CHECK (instrument_id <> '' AND length(instrument_id) <= 128),
    symbol text NOT NULL CHECK (symbol <> '' AND length(symbol) <= 120),
    kind text NOT NULL CHECK (kind <> '' AND length(kind) <= 60),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    units numeric,
    price numeric,
    cost_basis numeric,
    PRIMARY KEY (version_id, row_number)
);

CREATE TABLE portfolio_activity_versions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    inclusion_version bigint NOT NULL CHECK (inclusion_version > 0),
    observed_at timestamptz,
    retrieved_at timestamptz NOT NULL,
    published_at timestamptz NOT NULL,
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE,
    UNIQUE (user_id, account_id, inclusion_version)
);

CREATE TABLE portfolio_activity_heads (
    user_id uuid NOT NULL,
    account_id text NOT NULL,
    version_id uuid NOT NULL REFERENCES portfolio_activity_versions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, account_id),
    FOREIGN KEY (user_id, account_id) REFERENCES portfolio_account_identities(user_id, account_id) ON DELETE CASCADE
);

CREATE TABLE portfolio_activity_rows (
    version_id uuid NOT NULL REFERENCES portfolio_activity_versions(id) ON DELETE CASCADE,
    row_number integer NOT NULL CHECK (row_number >= 0),
    activity_id text NOT NULL CHECK (activity_id <> '' AND length(activity_id) <= 160),
    activity_type text NOT NULL CHECK (activity_type <> '' AND length(activity_type) <= 80),
    trade_date timestamptz,
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    amount numeric,
    fee numeric,
    price numeric,
    units numeric,
    PRIMARY KEY (version_id, row_number)
);
