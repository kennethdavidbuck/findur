DROP INDEX provider_authorizations_refresh_lease_idx;

ALTER TABLE provider_authorizations
    DROP COLUMN refresh_lease_expires_at,
    DROP COLUMN refresh_lease_id,
    DROP COLUMN credential_version,
    DROP COLUMN lifecycle_generation,
    DROP COLUMN lifecycle_status;
