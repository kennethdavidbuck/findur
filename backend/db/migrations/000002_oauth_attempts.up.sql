CREATE TABLE oauth_attempts (
    state_hash bytea PRIMARY KEY CHECK (octet_length(state_hash) = 32),
    nonce_hash bytea NOT NULL CHECK (octet_length(nonce_hash) = 32),
    browser_binding_hash bytea NOT NULL CHECK (octet_length(browser_binding_hash) = 32),
    pkce_verifier_encrypted bytea NOT NULL CHECK (octet_length(pkce_verifier_encrypted) > 32),
    return_route text NOT NULL CHECK (return_route IN ('/connect', '/portfolio')),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'exchanging', 'succeeded', 'restart-required')),
    expires_at timestamptz NOT NULL,
    claimed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((status = 'pending' AND claimed_at IS NULL) OR (status <> 'pending' AND claimed_at IS NOT NULL))
);

CREATE INDEX oauth_attempts_cleanup_idx ON oauth_attempts (expires_at);
