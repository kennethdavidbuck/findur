// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// OAuthAttemptRepository persists only hashed or encrypted OAuth material.
type OAuthAttemptRepository struct{ pool *pgxpool.Pool }

func NewOAuthAttemptRepository(pool *pgxpool.Pool) *OAuthAttemptRepository {
	return &OAuthAttemptRepository{pool: pool}
}

func (r *OAuthAttemptRepository) ClaimCallback(ctx context.Context, stateHash, bindingHash []byte, now time.Time) (auth.CallbackClaim, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return auth.CallbackClaim{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var claim auth.CallbackClaim
	var status string
	var terminalOutcome, terminalRoute *string
	var userID *uuid.UUID
	err = tx.QueryRow(ctx, `SELECT state_hash, nonce_hash, pkce_verifier_encrypted, return_route, status, terminal_outcome, terminal_route, user_id
		FROM oauth_attempts WHERE state_hash=$1 AND browser_binding_hash=$2
		AND (status IN ('succeeded','restart-required') OR expires_at>$3) FOR UPDATE`, stateHash, bindingHash, now).
		Scan(&claim.StateHash, &claim.NonceHash, &claim.EncryptedVerifier, &claim.ReturnRoute, &status, &terminalOutcome, &terminalRoute, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.CallbackClaim{}, auth.ErrNotClaimable
	}
	if err != nil {
		return auth.CallbackClaim{}, err
	}
	if status == "pending" {
		if _, err = tx.Exec(ctx, `UPDATE oauth_attempts SET claimed_at=$2,status='exchanging' WHERE state_hash=$1`, stateHash, now); err != nil {
			return auth.CallbackClaim{}, err
		}
	} else if status == "succeeded" || status == "restart-required" {
		if terminalOutcome != nil {
			claim.TerminalOutcome = *terminalOutcome
		}
		if terminalRoute != nil {
			claim.TerminalRoute = *terminalRoute
		}
		if userID != nil {
			claim.UserID = *userID
		}
	} else {
		return auth.CallbackClaim{}, auth.ErrNotClaimable
	}
	if err = tx.Commit(ctx); err != nil {
		return auth.CallbackClaim{}, err
	}
	return claim, nil
}

func (r *OAuthAttemptRepository) FindActiveUser(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT u.id FROM external_identities i JOIN users u ON u.id=i.user_id WHERE i.provider=$1 AND i.subject=$2 AND u.active`, provider, subject).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return id, err == nil, err
}

func (r *OAuthAttemptRepository) FinalizeCallback(ctx context.Context, value auth.Finalization) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status string
	if err = tx.QueryRow(ctx, `SELECT status FROM oauth_attempts WHERE state_hash=$1 FOR UPDATE`, value.StateHash).Scan(&status); err != nil || status != "exchanging" {
		if err == nil {
			err = auth.ErrNotClaimable
		}
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,origin) VALUES($1,'oauth') ON CONFLICT (id) DO NOTHING`, value.UserID); err != nil {
		return err
	}
	var owner uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO external_identities(provider,subject,user_id) VALUES($1,$2,$3)
		ON CONFLICT(provider,subject) DO UPDATE SET subject=EXCLUDED.subject RETURNING user_id`, value.Provider, value.Subject, value.UserID).Scan(&owner)
	if err != nil || owner != value.UserID {
		if err == nil {
			err = errors.New("identity owner changed")
		}
		return err
	}
	var active bool
	if err = tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1 FOR UPDATE`, owner).Scan(&active); err != nil || !active {
		if err == nil {
			err = auth.ErrNotClaimable
		}
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO provider_authorizations(user_id,provider,access_token_encrypted,refresh_token_encrypted,envelope_version,token_expires_at)
		VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(user_id,provider) DO UPDATE SET access_token_encrypted=EXCLUDED.access_token_encrypted,
		refresh_token_encrypted=EXCLUDED.refresh_token_encrypted,envelope_version=EXCLUDED.envelope_version,token_expires_at=EXCLUDED.token_expires_at,updated_at=now()`,
		owner, value.Provider, value.AccessToken, nullableBytes(value.RefreshToken), value.EnvelopeVersion, nullableTime(value.TokenExpiresAt))
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO sessions(session_hash,csrf_hash,user_id,idle_expires_at,absolute_expires_at) VALUES($1,$2,$3,$4,$5)`, value.SessionHash, value.CSRFHash, owner, value.SessionIdleExpiresAt, value.SessionAbsoluteExpiresAt); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE oauth_attempts SET status='succeeded',terminal_outcome='succeeded',terminal_route='/connect/result',user_id=$2,completed_at=$3 WHERE state_hash=$1`, value.StateHash, owner, value.CompletedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *OAuthAttemptRepository) FailCallback(ctx context.Context, stateHash []byte, now time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE oauth_attempts SET status='restart-required',terminal_outcome='restart-required',terminal_route='/connect/result',completed_at=$2
		WHERE state_hash=$1 AND status='exchanging'`, stateHash, now)
	return err
}

func (r *OAuthAttemptRepository) SessionActive(ctx context.Context, sessionHash []byte, now time.Time) (bool, error) {
	var active bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.session_hash=$1 AND s.idle_expires_at>$2 AND s.absolute_expires_at>$2 AND u.active)`, sessionHash, now).Scan(&active)
	return active, err
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
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
		WHERE (status = 'pending' AND expires_at <= $1)
		   OR (status = 'exchanging' AND claimed_at <= $1 - interval '10 minutes')
		   OR (status IN ('succeeded', 'restart-required') AND completed_at <= $1 - interval '10 minutes')`, now)
	return command.RowsAffected(), err
}
