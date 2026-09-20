// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// OAuthAttemptRepository persists only hashed or encrypted OAuth material.
type OAuthAttemptRepository struct{ pool *pgxpool.Pool }

func NewOAuthAttemptRepository(pool *pgxpool.Pool) *OAuthAttemptRepository {
	return &OAuthAttemptRepository{pool: pool}
}

func (r *OAuthAttemptRepository) Create(ctx context.Context, attempt auth.Attempt) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oauth_attempts
			(state_hash, nonce_hash, browser_binding_hash, pkce_verifier_encrypted, return_route, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		attempt.StateHash, attempt.NonceHash, attempt.BrowserBindingHash, attempt.EncryptedVerifier,
		attempt.ReturnRoute, attempt.ExpiresAt)
	return err
}

func (r *OAuthAttemptRepository) Delete(ctx context.Context, stateHash []byte) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_attempts WHERE state_hash = $1`, stateHash)
	return err
}

func (r *OAuthAttemptRepository) Claim(ctx context.Context, stateHash, bindingHash []byte, now time.Time) error {
	var claimed time.Time
	err := r.pool.QueryRow(ctx, `
		UPDATE oauth_attempts
		SET claimed_at = $3, status = 'exchanging'
		WHERE state_hash = $1 AND browser_binding_hash = $2
		  AND status = 'pending' AND claimed_at IS NULL AND expires_at > $3
		RETURNING claimed_at`, stateHash, bindingHash, now).Scan(&claimed)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrNotClaimable
	}
	return err
}

func (r *OAuthAttemptRepository) Cleanup(ctx context.Context, now time.Time) (int64, error) {
	command, err := r.pool.Exec(ctx, `
		DELETE FROM oauth_attempts
		WHERE expires_at <= $1 OR (claimed_at IS NOT NULL AND claimed_at <= $1 - interval '10 minutes')`, now)
	return command.RowsAffected(), err
}
