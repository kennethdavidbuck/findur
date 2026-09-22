ALTER TABLE provider_authorizations
    ADD COLUMN lifecycle_status text NOT NULL DEFAULT 'active' CHECK (lifecycle_status IN ('active', 'reauthorization-required', 'disconnecting')),
    ADD COLUMN lifecycle_generation bigint NOT NULL DEFAULT 1,
    ADD COLUMN credential_version bigint NOT NULL DEFAULT 1,
    ADD COLUMN refresh_lease_id uuid,
    ADD COLUMN refresh_lease_expires_at timestamptz;

CREATE INDEX provider_authorizations_refresh_lease_idx
    ON provider_authorizations (refresh_lease_expires_at)
    WHERE refresh_lease_id IS NOT NULL;
