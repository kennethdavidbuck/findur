CREATE TABLE portfolio_inventory_state (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_generation bigint NOT NULL CHECK (current_generation > 0),
    current_status text NOT NULL CHECK (current_status IN ('pending', 'ready', 'empty', 'disabled', 'unauthorized', 'rate_limited', 'unavailable', 'malformed')),
    head_generation bigint,
    retry_at timestamptz,
    claim_expires_at timestamptz,
    updated_at timestamptz NOT NULL,
    CHECK (head_generation IS NULL OR head_generation > 0),
    CHECK (head_generation IS NULL OR head_generation <= current_generation),
    CHECK ((current_status = 'pending') = (claim_expires_at IS NOT NULL))
);

CREATE TABLE portfolio_inventory_versions (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    generation bigint NOT NULL CHECK (generation > 0),
    status text NOT NULL CHECK (status IN ('ready', 'empty', 'disabled', 'unauthorized', 'rate_limited', 'unavailable', 'malformed')),
    published_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, generation)
);

CREATE TABLE portfolio_inventory_connections (
    user_id uuid NOT NULL,
    generation bigint NOT NULL,
    connection_id text NOT NULL CHECK (connection_id <> '' AND length(connection_id) <= 128),
    brokerage_label text NOT NULL CHECK (brokerage_label <> '' AND length(brokerage_label) <= 120),
    status text NOT NULL CHECK (status IN ('active', 'disabled', 'unavailable')),
    sync_mode text NOT NULL CHECK (sync_mode IN ('realtime', 'delayed', 'unknown')),
    available boolean NOT NULL,
    eligible boolean NOT NULL,
    PRIMARY KEY (user_id, generation, connection_id),
    FOREIGN KEY (user_id, generation) REFERENCES portfolio_inventory_versions(user_id, generation) ON DELETE CASCADE
);

CREATE TABLE portfolio_inventory_accounts (
    user_id uuid NOT NULL,
    generation bigint NOT NULL,
    account_id text NOT NULL CHECK (account_id <> '' AND length(account_id) <= 128),
    connection_id text NOT NULL,
    category text NOT NULL CHECK (category IN ('investment', 'deposit', 'credit', 'unknown')),
    account_type text NOT NULL CHECK (account_type <> '' AND length(account_type) <= 120),
    masked_label text NOT NULL CHECK (masked_label <> '' AND length(masked_label) <= 180),
    available boolean NOT NULL,
    eligible boolean NOT NULL,
    sync_state text NOT NULL CHECK (sync_state IN ('complete', 'pending', 'unavailable', 'unknown')),
    PRIMARY KEY (user_id, generation, account_id),
    FOREIGN KEY (user_id, generation, connection_id) REFERENCES portfolio_inventory_connections(user_id, generation, connection_id) ON DELETE CASCADE
);

CREATE INDEX portfolio_inventory_retry_idx ON portfolio_inventory_state (retry_at) WHERE current_status = 'rate_limited';
CREATE INDEX portfolio_inventory_claim_expiry_idx ON portfolio_inventory_state (claim_expires_at) WHERE current_status = 'pending';
