package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// SessionIdleLifetime is the rolling inactivity window for an active session.
	SessionIdleLifetime = 12 * time.Hour
	// SessionAbsoluteLifetime is the fixed maximum lifetime that renewal cannot extend.
	SessionAbsoluteLifetime = 7 * 24 * time.Hour
)

var (
	// ErrUnauthenticated indicates that no active session could be derived.
	ErrUnauthenticated = errors.New("session is not authenticated")
	// ErrForbidden indicates that an authenticated request defense failed.
	ErrForbidden = errors.New("request is forbidden")
)

// Actor is the immutable owner identity derived from an authenticated session.
// Its identifier is deliberately unexported so transports cannot construct it.
type Actor struct{ userID uuid.UUID }

// UserID exposes the server-derived identifier to owner-scoped application code.
func (a Actor) UserID() uuid.UUID { return a.userID }

// Session is the storage-safe material resolved for one active browser session.
type Session struct {
	UserID   uuid.UUID
	CSRFHash []byte
}

// SessionRepository owns atomic session lifecycle persistence.
type SessionRepository interface {
	FindActive(context.Context, []byte, time.Time) (Session, error)
	AuthenticateAndTouch(context.Context, []byte, time.Time, time.Time) (Session, error)
	RevokeCurrent(context.Context, []byte, time.Time) (bool, error)
	CleanupSessions(context.Context, time.Time) (int64, error)
}

// SessionConfig contains checked-in session policy dependencies.
type SessionConfig struct {
	HashKey []byte
	Clock   func() time.Time
}

// SessionService is the only boundary that turns browser secrets into an Actor.
type SessionService struct {
	repository SessionRepository
	hashKey    []byte
	clock      func() time.Time
}

// NewSessionService validates session policy and dependencies.
func NewSessionService(cfg SessionConfig, repository SessionRepository) (*SessionService, error) {
	if repository == nil || len(cfg.HashKey) < 32 || cfg.Clock == nil {
		return nil, errors.New("incomplete session service configuration")
	}
	return &SessionService{repository: repository, hashKey: append([]byte(nil), cfg.HashKey...), clock: cfg.Clock}, nil
}

// Authenticate validates and atomically rolls the idle deadline, capped by the
// fixed absolute deadline. It never accepts an actor identifier from a caller.
func (s *SessionService) Authenticate(ctx context.Context, secret string) (Actor, error) {
	session, err := s.authenticate(ctx, secret)
	if err != nil {
		return Actor{}, err
	}
	return Actor{userID: session.UserID}, nil
}

// AuthorizeUnsafe authenticates and verifies the session-bound CSRF secret.
func (s *SessionService) AuthorizeUnsafe(ctx context.Context, secret, csrf string) (Actor, error) {
	if secret == "" {
		return Actor{}, ErrUnauthenticated
	}
	if csrf == "" {
		return Actor{}, ErrForbidden
	}
	now := s.clock().UTC()
	hash := s.hash(secret)
	session, err := s.repository.FindActive(ctx, hash, now)
	if err != nil {
		return Actor{}, err
	}
	if subtle.ConstantTimeCompare(session.CSRFHash, s.hash(csrf)) != 1 {
		return Actor{}, ErrForbidden
	}
	session, err = s.repository.AuthenticateAndTouch(ctx, hash, now, now.Add(SessionIdleLifetime))
	if err != nil {
		return Actor{}, err
	}
	return Actor{userID: session.UserID}, nil
}

// RevokeCurrent ends only the presented browser session.
func (s *SessionService) RevokeCurrent(ctx context.Context, secret string) error {
	if secret == "" {
		return ErrUnauthenticated
	}
	revoked, err := s.repository.RevokeCurrent(ctx, s.hash(secret), s.clock().UTC())
	if err != nil {
		return err
	}
	if !revoked {
		return ErrUnauthenticated
	}
	return nil
}

// Cleanup removes expired and revoked session rows without exposing secrets.
func (s *SessionService) Cleanup(ctx context.Context) (int64, error) {
	return s.repository.CleanupSessions(ctx, s.clock().UTC())
}

func (s *SessionService) authenticate(ctx context.Context, secret string) (Session, error) {
	if secret == "" {
		return Session{}, ErrUnauthenticated
	}
	now := s.clock().UTC()
	return s.repository.AuthenticateAndTouch(ctx, s.hash(secret), now, now.Add(SessionIdleLifetime))
}

func (s *SessionService) hash(value string) []byte {
	digest := hmac.New(sha256.New, s.hashKey)
	_, _ = digest.Write([]byte(value))
	return digest.Sum(nil)
}
