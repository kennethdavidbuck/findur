DROP INDEX sessions_cleanup_idx;

ALTER TABLE sessions
    DROP COLUMN revoked_at,
    DROP COLUMN last_used_at;
