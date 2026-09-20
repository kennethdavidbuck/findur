CREATE TABLE users (
    id uuid PRIMARY KEY,
    origin text NOT NULL CHECK (origin = 'oauth'),
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE external_identities (
    provider text NOT NULL,
    subject text NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, subject),
    UNIQUE (user_id, provider)
);

CREATE TABLE provider_authorizations (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider text NOT NULL,
    access_token_encrypted bytea NOT NULL,
    refresh_token_encrypted bytea,
    envelope_version integer NOT NULL CHECK (envelope_version > 0),
    token_expires_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, provider)
);

CREATE TABLE sessions (
    session_hash bytea PRIMARY KEY CHECK (octet_length(session_hash) = 32),
    csrf_hash bytea NOT NULL CHECK (octet_length(csrf_hash) = 32),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (idle_expires_at <= absolute_expires_at)
);

ALTER TABLE oauth_attempts
    ADD COLUMN terminal_outcome text CHECK (terminal_outcome IN ('succeeded', 'restart-required')),
    ADD COLUMN terminal_route text CHECK (terminal_route IN ('/connect', '/connect/result', '/portfolio')),
    ADD COLUMN user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN completed_at timestamptz;

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expiry_idx ON sessions (absolute_expires_at);
