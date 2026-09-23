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

const (
	attemptPending         = "pending"
	attemptExchanging      = "exchanging"
	attemptSucceeded       = "succeeded"
	attemptRestartRequired = "restart-required"
	userOriginOAuth        = "oauth"
)

// NewOAuthAttemptRepository creates a PostgreSQL OAuth persistence adapter.
func NewOAuthAttemptRepository(pool *pgxpool.Pool) *OAuthAttemptRepository {
	return &OAuthAttemptRepository{pool: pool}
}

// ClaimCallback atomically claims a pending callback or reads its terminal result.
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
        AND (status IN ($4,$5) OR expires_at>$3) FOR UPDATE`, stateHash, bindingHash, now, attemptSucceeded, attemptRestartRequired).
		Scan(&claim.StateHash, &claim.NonceHash, &claim.EncryptedVerifier, &claim.ReturnRoute, &status, &terminalOutcome, &terminalRoute, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.CallbackClaim{}, auth.ErrNotClaimable
	}
	if err != nil {
		return auth.CallbackClaim{}, err
	}
	switch status {
	case attemptPending:
		if _, err = tx.Exec(ctx, `UPDATE oauth_attempts SET claimed_at=$2,status=$3 WHERE state_hash=$1`, stateHash, now, attemptExchanging); err != nil {
			return auth.CallbackClaim{}, err
		}
	case attemptSucceeded, attemptRestartRequired:
		if terminalOutcome != nil {
			claim.TerminalOutcome = *terminalOutcome
		}
		if terminalRoute != nil {
			claim.TerminalRoute = *terminalRoute
		}
		if userID != nil {
			claim.UserID = *userID
		}
	default:
		return auth.CallbackClaim{}, auth.ErrNotClaimable
	}
	if err = tx.Commit(ctx); err != nil {
		return auth.CallbackClaim{}, err
	}
	return claim, nil
}

// FindActiveUser resolves an active user by verified external identity.
func (r *OAuthAttemptRepository) FindActiveUser(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT u.id FROM external_identities i JOIN users u ON u.id=i.user_id WHERE i.provider=$1 AND i.subject=$2 AND u.active`, provider, subject).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return id, err == nil, err
}

// FinalizeCallback atomically persists identity, authorization, session, and terminal state.
func (r *OAuthAttemptRepository) FinalizeCallback(ctx context.Context, value auth.Finalization) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = requireAttemptStatus(ctx, tx, value.StateHash, attemptExchanging); err != nil {
		return err
	}
	owner, err := upsertActiveIdentity(ctx, tx, value)
	if err != nil {
		return err
	}
	if err = storeAuthorization(ctx, tx, owner, value); err != nil {
		return err
	}
	if err = storeSession(ctx, tx, owner, value); err != nil {
		return err
	}
	if err = completeAttempt(ctx, tx, owner, value); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func requireAttemptStatus(ctx context.Context, tx pgx.Tx, stateHash []byte, expected string) error {
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM oauth_attempts WHERE state_hash=$1 FOR UPDATE`, stateHash).Scan(&status); err != nil {
		return err
	}
	if status != expected {
		return auth.ErrNotClaimable
	}
	return nil
}

func upsertActiveIdentity(ctx context.Context, tx pgx.Tx, value auth.Finalization) (uuid.UUID, error) {
	if _, err := tx.Exec(ctx, `INSERT INTO users(id,origin) VALUES($1,$2) ON CONFLICT (id) DO NOTHING`, value.UserID, userOriginOAuth); err != nil {
		return uuid.Nil, err
	}
	var owner uuid.UUID
	err := tx.QueryRow(ctx, `INSERT INTO external_identities(provider,subject,user_id) VALUES($1,$2,$3)
        ON CONFLICT(provider,subject) DO UPDATE SET subject=EXCLUDED.subject RETURNING user_id`, value.Provider, value.Subject, value.UserID).Scan(&owner)
	if err != nil {
		return uuid.Nil, err
	}
	if owner != value.UserID {
		return uuid.Nil, errors.New("identity owner changed")
	}
	var active bool
	if err = tx.QueryRow(ctx, `SELECT active FROM users WHERE id=$1 FOR UPDATE`, owner).Scan(&active); err != nil {
		return uuid.Nil, err
	}
	if !active {
		return uuid.Nil, auth.ErrNotClaimable
	}
	return owner, nil
}

func storeAuthorization(ctx context.Context, tx pgx.Tx, owner uuid.UUID, value auth.Finalization) error {
	_, err := tx.Exec(ctx, `INSERT INTO provider_authorizations (
            user_id,
            provider,
            access_token_encrypted,
            refresh_token_encrypted,
            envelope_version,
            token_expires_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id, provider) DO UPDATE
        SET access_token_encrypted = EXCLUDED.access_token_encrypted,
            refresh_token_encrypted = EXCLUDED.refresh_token_encrypted,
            envelope_version = EXCLUDED.envelope_version,
            token_expires_at = EXCLUDED.token_expires_at,
            lifecycle_status = 'active',
            lifecycle_generation = provider_authorizations.lifecycle_generation + 1,
            credential_version = provider_authorizations.credential_version + 1,
            refresh_lease_id = NULL,
            refresh_lease_expires_at = NULL,
            updated_at = now()`,
		owner, value.Provider, value.AccessToken, nullableBytes(value.RefreshToken), value.EnvelopeVersion, nullableTime(value.TokenExpiresAt))
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE portfolio_inclusion_state
        SET lifecycle_generation = lifecycle_generation + 1,
            updated_at = $2
        WHERE user_id = $1`, owner, value.CompletedAt); err != nil {
		return err
	}
	// The requested selection remains durable across reauthorization. Move its
	// pending worker change to the new lifecycle while old in-flight finalizers
	// retain their previous generation and fail their publication guard.
	if _, err = tx.Exec(ctx, `UPDATE portfolio_inclusion_changes change
        SET lifecycle_generation = inclusion.lifecycle_generation,
            updated_at = $2
        FROM portfolio_inclusion_state inclusion
        WHERE change.user_id = $1
            AND inclusion.user_id = change.user_id
            AND change.status = 'pending'
            AND change.result_version = inclusion.version`, owner, value.CompletedAt); err != nil {
		return err
	}
	// A prior authorization failure may have left account resources in backoff.
	// Fresh credentials make that delay obsolete; the worker may claim them again.
	if _, err = tx.Exec(ctx, `UPDATE portfolio_account_sync_state
        SET next_attempt_at = NULL,
            failure_count = 0,
            claim_id = NULL,
            claim_expires_at = NULL,
            claimed_resource = NULL,
            claimed_change_id = NULL,
            updated_at = $2
        WHERE user_id = $1
            AND (next_attempt_at IS NOT NULL OR claim_id IS NOT NULL)`, owner, value.CompletedAt); err != nil {
		return err
	}
	// Authorization-required inventory work stays paused while credentials are
	// inactive. Fresh credentials make it immediately eligible for the worker.
	if _, err = tx.Exec(ctx, `UPDATE portfolio_inventory_state
        SET retry_at = NULL,
            failure_count = 0,
            claim_expires_at = NULL,
            updated_at = $2
        WHERE user_id = $1
			AND (current_status = 'unauthorized' OR retry_at IS NOT NULL)`, owner, value.CompletedAt); err != nil {
		return err
	}
	// Returning logins rotate credentials and fence in-flight portfolio work, but
	// preserve the last successfully published account inventory.
	return nil
}

func storeSession(ctx context.Context, tx pgx.Tx, owner uuid.UUID, value auth.Finalization) error {
	_, err := tx.Exec(ctx, `INSERT INTO sessions(session_hash,csrf_hash,user_id,idle_expires_at,absolute_expires_at) VALUES($1,$2,$3,$4,$5)`, value.SessionHash, value.CSRFHash, owner, value.SessionIdleExpiresAt, value.SessionAbsoluteExpiresAt)
	return err
}

func completeAttempt(ctx context.Context, tx pgx.Tx, owner uuid.UUID, value auth.Finalization) error {
	_, err := tx.Exec(ctx, `UPDATE oauth_attempts SET status=$4,terminal_outcome=$4,terminal_route=$5,user_id=$2,completed_at=$3 WHERE state_hash=$1`, value.StateHash, owner, value.CompletedAt, attemptSucceeded, auth.AuthorizationResultRoute)
	return err
}

// FailCallback makes an exchanging callback terminal and restartable.
func (r *OAuthAttemptRepository) FailCallback(ctx context.Context, stateHash []byte, now time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE oauth_attempts SET status=$3,terminal_outcome=$3,terminal_route=$4,completed_at=$2
        WHERE state_hash=$1 AND status=$5`, stateHash, now, attemptRestartRequired, auth.AuthorizationResultRoute, attemptExchanging)
	return err
}

// SessionActive reports whether a hashed session belongs to an active user and is unexpired.
func (r *OAuthAttemptRepository) SessionActive(ctx context.Context, sessionHash []byte, now time.Time) (bool, error) {
	var active bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.session_hash=$1 AND s.revoked_at IS NULL AND s.idle_expires_at>$2 AND s.absolute_expires_at>$2 AND u.active)`, sessionHash, now).Scan(&active)
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

// Create persists a new encrypted authorization attempt.
func (r *OAuthAttemptRepository) Create(ctx context.Context, attempt auth.Attempt) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO oauth_attempts
            (state_hash, nonce_hash, browser_binding_hash, pkce_verifier_encrypted, return_route, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)`,
		attempt.StateHash, attempt.NonceHash, attempt.BrowserBindingHash, attempt.EncryptedVerifier,
		attempt.ReturnRoute, attempt.ExpiresAt)
	return err
}

// Delete removes an authorization attempt by hashed state.
func (r *OAuthAttemptRepository) Delete(ctx context.Context, stateHash []byte) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_attempts WHERE state_hash = $1`, stateHash)
	return err
}

// Claim atomically marks a matching unexpired attempt as exchanging.
func (r *OAuthAttemptRepository) Claim(ctx context.Context, stateHash, bindingHash []byte, now time.Time) error {
	var claimed time.Time
	err := r.pool.QueryRow(ctx, `
        UPDATE oauth_attempts
        SET claimed_at = $3, status = $4
        WHERE state_hash = $1 AND browser_binding_hash = $2
          AND status = $5 AND claimed_at IS NULL AND expires_at > $3
        RETURNING claimed_at`, stateHash, bindingHash, now, attemptExchanging, attemptPending).Scan(&claimed)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrNotClaimable
	}
	return err
}

// Cleanup removes expired, abandoned, and old terminal authorization attempts.
func (r *OAuthAttemptRepository) Cleanup(ctx context.Context, now time.Time) (int64, error) {
	command, err := r.pool.Exec(ctx, `
        DELETE FROM oauth_attempts
        WHERE (status = $2 AND expires_at <= $1)
           OR (status = $3 AND claimed_at <= $1 - interval '10 minutes')
           OR (status IN ($4, $5) AND completed_at <= $1 - interval '10 minutes')`, now, attemptPending, attemptExchanging, attemptSucceeded, attemptRestartRequired)
	return command.RowsAffected(), err
}
