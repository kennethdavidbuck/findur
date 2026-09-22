package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// CredentialRepository coordinates one rotating refresh token across processes.
type CredentialRepository struct{ pool *pgxpool.Pool }

// NewCredentialRepository creates the PostgreSQL credential coordination adapter.
func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

// ReadCredential reads the current encrypted credential without a lease claim.
func (r *CredentialRepository) ReadCredential(ctx context.Context, owner uuid.UUID) (auth.Credential, bool, error) {
	return r.read(ctx, r.pool, owner, false)
}

func (r *CredentialRepository) read(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, owner uuid.UUID, lock bool) (auth.Credential, bool, error) {
	query := `SELECT
            user_id,
            access_token_encrypted,
            refresh_token_encrypted,
            envelope_version,
            token_expires_at,
            lifecycle_status,
            lifecycle_generation,
            credential_version,
            refresh_lease_id,
            refresh_lease_expires_at
        FROM provider_authorizations
        WHERE user_id = $1
            AND provider = $2`
	if lock {
		query += `
        FOR UPDATE`
	}
	var c auth.Credential
	var expiry *time.Time
	err := q.QueryRow(ctx, query, owner, auth.SnapTradeProvider).Scan(&c.Owner, &c.AccessEnvelope, &c.RefreshEnvelope, &c.EnvelopeVersion, &expiry, &c.Status, &c.Generation, &c.Version, &c.LeaseID, &c.LeaseExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Credential{}, false, nil
	}
	if expiry != nil {
		c.ExpiresAt = expiry.UTC()
	}
	return c, err == nil, err
}

// ClaimRefresh obtains the sole short-lived refresh lease, or fails closed after expiry.
func (r *CredentialRepository) ClaimRefresh(ctx context.Context, owner uuid.UUID, expectedVersion int64, now time.Time, lease time.Duration) (auth.Credential, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return auth.Credential{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	c, found, err := r.read(ctx, tx, owner, true)
	if err != nil || !found {
		return c, false, err
	}
	if c.Status != auth.CredentialStatusActive {
		return c, false, nil
	}
	if c.Version != expectedVersion {
		if err = tx.Commit(ctx); err != nil {
			return c, false, err
		}
		return c, false, nil
	}
	// A lease that has run out may have sent a refresh request. Its refresh token
	// could have rotated, so recovery is reauthorization rather than replay.
	if c.LeaseID != nil {
		if c.LeaseExpiresAt != nil && !c.LeaseExpiresAt.After(now) {
			_, err = tx.Exec(ctx, `UPDATE provider_authorizations
                SET lifecycle_status = 'reauthorization-required',
                    lifecycle_generation = lifecycle_generation + 1,
                    refresh_lease_id = NULL,
                    refresh_lease_expires_at = NULL,
                    updated_at = $2
                WHERE user_id = $1
                    AND provider = $3
                    AND credential_version = $4`, owner, now, auth.SnapTradeProvider, c.Version)
			if err != nil {
				return c, false, err
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return c, false, err
		}
		return c, false, nil
	}
	id := uuid.New()
	expires := now.Add(lease)
	command, err := tx.Exec(ctx, `UPDATE provider_authorizations
        SET refresh_lease_id = $3,
            refresh_lease_expires_at = $4,
            updated_at = $2
        WHERE user_id = $1
            AND provider = $5
            AND credential_version = $6
            AND lifecycle_generation = $7
            AND lifecycle_status = 'active'
            AND refresh_lease_id IS NULL`, owner, now, id, expires, auth.SnapTradeProvider, c.Version, c.Generation)
	if err != nil {
		return c, false, err
	}
	if command.RowsAffected() == 1 {
		c.LeaseID, c.LeaseExpiresAt = &id, &expires
	}
	if err = tx.Commit(ctx); err != nil {
		return c, false, err
	}
	return c, command.RowsAffected() == 1, nil
}

// ReleaseRefresh clears a lease only when no token exchange was sent.
func (r *CredentialRepository) ReleaseRefresh(ctx context.Context, claim auth.Credential, now time.Time) (bool, error) {
	if claim.LeaseID == nil {
		return false, nil
	}
	command, err := r.pool.Exec(ctx, `UPDATE provider_authorizations
        SET refresh_lease_id = NULL,
            refresh_lease_expires_at = NULL,
            updated_at = $6
        WHERE user_id = $1
            AND provider = $2
            AND lifecycle_status = 'active'
            AND lifecycle_generation = $3
            AND credential_version = $4
            AND refresh_lease_id = $5`, claim.Owner, auth.SnapTradeProvider, claim.Generation, claim.Version, *claim.LeaseID, now)
	return command.RowsAffected() == 1, err
}

// InstallRefresh atomically installs rotated envelopes only for the active lease.
func (r *CredentialRepository) InstallRefresh(ctx context.Context, claim auth.Credential, access, refresh []byte, envelopeVersion int, expiry, now time.Time) (bool, error) {
	if claim.LeaseID == nil {
		return false, nil
	}
	command, err := r.pool.Exec(ctx, `UPDATE provider_authorizations
        SET access_token_encrypted = $4,
            refresh_token_encrypted = $5,
            envelope_version = $6,
            token_expires_at = $7,
            credential_version = credential_version + 1,
            refresh_lease_id = NULL,
            refresh_lease_expires_at = NULL,
            updated_at = $8
        WHERE user_id = $1
            AND provider = $2
            AND lifecycle_status = 'active'
            AND lifecycle_generation = $3
            AND credential_version = $9
            AND refresh_lease_id = $10
            AND refresh_lease_expires_at > $8`, claim.Owner, auth.SnapTradeProvider, claim.Generation, access, refresh, envelopeVersion, expiry, now, claim.Version, *claim.LeaseID)
	return command.RowsAffected() == 1, err
}

// RequireReauthorization disables an authorization after an unsafe refresh outcome.
func (r *CredentialRepository) RequireReauthorization(ctx context.Context, owner uuid.UUID, lease *uuid.UUID, version int64, now time.Time) error {
	query := `UPDATE provider_authorizations
        SET lifecycle_status = 'reauthorization-required',
            lifecycle_generation = lifecycle_generation + 1,
            refresh_lease_id = NULL,
            refresh_lease_expires_at = NULL,
            updated_at = $2
        WHERE user_id = $1
            AND provider = $3
            AND lifecycle_status = 'active'`
	args := []any{owner, now, auth.SnapTradeProvider}
	if lease != nil {
		query += `
            AND refresh_lease_id = $4`
		args = append(args, *lease)
	}
	if version > 0 {
		query += `
            AND credential_version = $` + strconv.Itoa(len(args)+1)
		args = append(args, version)
	}
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
