package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// SessionRepository persists session lifecycle state independently of OAuth initiation.
type SessionRepository struct{ pool *pgxpool.Pool }

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// FindActive resolves session-bound authorization material without renewing the
// idle deadline, so a failed CSRF check cannot mutate session state.
func (r *SessionRepository) FindActive(ctx context.Context, hash []byte, now time.Time) (auth.Session, error) {
	var session auth.Session
	err := r.pool.QueryRow(ctx, `SELECT s.user_id, s.csrf_hash
		FROM sessions s
		JOIN users u ON u.id=s.user_id
		WHERE s.session_hash=$1 AND u.active
		  AND s.revoked_at IS NULL AND s.idle_expires_at>$2 AND s.absolute_expires_at>$2`, hash, now).Scan(&session.UserID, &session.CSRFHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Session{}, auth.ErrUnauthenticated
	}
	return session, err
}

// AuthenticateAndTouch atomically rejects inactive sessions and owners, then
// advances idle expiry without ever moving the absolute deadline.
func (r *SessionRepository) AuthenticateAndTouch(ctx context.Context, hash []byte, now, requestedIdle time.Time) (auth.Session, error) {
	var session auth.Session
	err := r.pool.QueryRow(ctx, `UPDATE sessions s
		SET idle_expires_at=LEAST($3, s.absolute_expires_at), last_used_at=$2
		FROM users u
		WHERE s.session_hash=$1 AND u.id=s.user_id AND u.active
		  AND s.revoked_at IS NULL AND s.idle_expires_at>$2 AND s.absolute_expires_at>$2
		RETURNING s.user_id, s.csrf_hash`, hash, now, requestedIdle).Scan(&session.UserID, &session.CSRFHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Session{}, auth.ErrUnauthenticated
	}
	return session, err
}

// RevokeCurrent marks only the presented active session as revoked.
func (r *SessionRepository) RevokeCurrent(ctx context.Context, hash []byte, now time.Time) (bool, error) {
	command, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked_at=$2
		WHERE session_hash=$1 AND revoked_at IS NULL AND idle_expires_at>$2 AND absolute_expires_at>$2`, hash, now)
	return command.RowsAffected() == 1, err
}

// CleanupSessions removes state that can no longer authenticate.
func (r *SessionRepository) CleanupSessions(ctx context.Context, now time.Time) (int64, error) {
	command, err := r.pool.Exec(ctx, `DELETE FROM sessions
		WHERE idle_expires_at<=$1 OR absolute_expires_at<=$1 OR revoked_at IS NOT NULL`, now)
	return command.RowsAffected(), err
}
