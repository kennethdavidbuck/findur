package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type sessionRepositoryStub struct {
	session       Session
	authErr       error
	revoked       bool
	revokeErr     error
	receivedHash  []byte
	receivedNow   time.Time
	receivedTouch time.Time
	findCalls     int
	touchCalls    int
}

func (r *sessionRepositoryStub) FindActive(_ context.Context, hash []byte, now time.Time) (Session, error) {
	r.findCalls++
	r.receivedHash, r.receivedNow = hash, now
	return r.session, r.authErr
}
func (r *sessionRepositoryStub) AuthenticateAndTouch(_ context.Context, hash []byte, now, touch time.Time) (Session, error) {
	r.touchCalls++
	r.receivedHash, r.receivedNow, r.receivedTouch = hash, now, touch
	return r.session, r.authErr
}
func (r *sessionRepositoryStub) RevokeCurrent(_ context.Context, hash []byte, now time.Time) (bool, error) {
	r.receivedHash, r.receivedNow = hash, now
	return r.revoked, r.revokeErr
}
func (*sessionRepositoryStub) CleanupSessions(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestSessionServiceDerivesActorAndUsesCheckedInIdlePolicy(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	owner := uuid.New()
	repository := &sessionRepositoryStub{session: Session{UserID: owner, CSRFHash: keyedHash(bytes.Repeat([]byte{8}, 32), "csrf")}}
	service, err := NewSessionService(SessionConfig{HashKey: bytes.Repeat([]byte{8}, 32), Clock: func() time.Time { return now }}, repository)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := service.Authenticate(context.Background(), "session")
	if err != nil || actor.UserID() != owner {
		t.Fatalf("actor=%v err=%v", actor.UserID(), err)
	}
	if !repository.receivedTouch.Equal(now.Add(12*time.Hour)) || bytes.Equal(repository.receivedHash, []byte("session")) {
		t.Fatalf("touch=%v hash=%x", repository.receivedTouch, repository.receivedHash)
	}
	if _, err := service.AuthorizeUnsafe(context.Background(), "session", "csrf"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AuthorizeUnsafe(context.Background(), "session", "wrong"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("wrong CSRF error=%v", err)
	}
	if repository.touchCalls != 2 {
		t.Fatalf("invalid CSRF touched session: touch calls=%d", repository.touchCalls)
	}
	if _, err := service.AuthorizeUnsafe(context.Background(), "session", ""); !errors.Is(err, ErrForbidden) || repository.findCalls != 2 || repository.touchCalls != 2 {
		t.Fatalf("missing CSRF error=%v find calls=%d touch calls=%d", err, repository.findCalls, repository.touchCalls)
	}
}

func TestSessionServiceFailsClosedAndRevokesOnlyPresentedSession(t *testing.T) {
	repository := &sessionRepositoryStub{authErr: ErrUnauthenticated, revoked: true}
	service, _ := NewSessionService(SessionConfig{HashKey: bytes.Repeat([]byte{9}, 32), Clock: time.Now}, repository)
	if _, err := service.Authenticate(context.Background(), ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("empty session error=%v", err)
	}
	if _, err := service.Authenticate(context.Background(), "expired"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expired session error=%v", err)
	}
	if err := service.RevokeCurrent(context.Background(), "current"); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(repository.receivedHash, []byte("current")) {
		t.Fatal("plaintext session reached repository")
	}
}
