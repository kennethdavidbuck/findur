ALTER TABLE sessions
    ADD COLUMN last_used_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN revoked_at timestamptz;

CREATE INDEX sessions_cleanup_idx
    ON sessions (idle_expires_at, absolute_expires_at, revoked_at);
