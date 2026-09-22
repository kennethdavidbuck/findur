package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCredentialSourceRefreshesDueCredentialOnce(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{1}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{2}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now.Add(RefreshEarlyWindow), Status: "active", Generation: 1, Version: 1}}
	grant := &refreshStub{result: TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}}
	var logs bytes.Buffer
	source, err := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewJSONHandler(&logs, nil)), func() time.Time { return now }, time.Second, time.Second, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := source.Access(context.Background(), owner); err != nil || got != "new" {
		t.Fatalf("token=%q err=%v", got, err)
	}
	if grant.calls != 1 || repo.c.Version != 2 {
		t.Fatalf("refreshes=%d version=%d", grant.calls, repo.c.Version)
	}
	for _, event := range []string{"credential_refresh_claimed", "credential_refresh_installed"} {
		if !strings.Contains(logs.String(), `"event":"`+event+`"`) {
			t.Fatalf("missing %s event: %s", event, logs.String())
		}
	}
	for _, secret := range []string{"old", "new", "refresh", "rotated"} {
		if strings.Contains(logs.String(), `"`+secret+`"`) {
			t.Fatalf("credential material appeared in logs: %s", logs.String())
		}
	}
}

func TestCredentialSourceReusesRotationAfterConcurrent401(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{3}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{4}, 512)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now.Add(time.Hour), Status: "active", Generation: 1, Version: 1}}
	grant := &refreshStub{result: TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}}
	source, _ := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now }, time.Second, time.Second, time.Second)
	unauthorized := &testUnauthorized{}
	var calls sync.WaitGroup
	calls.Add(2)
	for range 2 {
		go func() {
			defer calls.Done()
			attempts := 0
			_ = source.Read(context.Background(), owner, func(_ context.Context, token string) error {
				attempts++
				if token == "old" {
					return unauthorized
				}
				return nil
			})
		}()
	}
	calls.Wait()
	if grant.calls != 1 {
		t.Fatalf("refreshes=%d", grant.calls)
	}
}

func TestCredentialSourceFailsClosedWhenRefreshMayHaveBeenSent(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now, Status: "active", Generation: 1, Version: 1}}
	source, _ := NewCredentialSource(repo, cipher, &refreshStub{err: context.DeadlineExceeded}, slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now }, time.Second, time.Second, time.Second)
	if _, err := source.Access(context.Background(), owner); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("error=%v", err)
	}
	if repo.c.Status != "reauthorization-required" {
		t.Fatalf("status=%q", repo.c.Status)
	}
}

func TestCredentialSourceDisablesUnreadableAccessCredential(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 256)))
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: []byte("invalid"), EnvelopeVersion: 1, ExpiresAt: now.Add(time.Hour), Status: "active", Generation: 1, Version: 1}}
	grant := &refreshStub{}
	source, _ := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now }, time.Second, time.Second, time.Second)
	if _, err := source.Access(context.Background(), owner); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("error=%v", err)
	}
	if repo.c.Status != "reauthorization-required" || grant.calls != 0 {
		t.Fatalf("status=%q refreshes=%d", repo.c.Status, grant.calls)
	}
}

func TestCredentialSourceRetriesPreSendFailureAndReleasesLease(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now, Status: "active", Generation: 1, Version: 1}}
	grant := &refreshStub{err: &RefreshPreSendError{Cause: errors.New("discovery unavailable")}}
	source, _ := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, time.Second, time.Second, time.Second)
	if _, err := source.Access(context.Background(), owner); !errors.Is(err, ErrRefreshUnavailable) {
		t.Fatalf("error=%v", err)
	}
	if grant.calls != 3 || repo.c.Status != "active" || repo.c.LeaseID != nil {
		t.Fatalf("calls=%d status=%q lease=%v", grant.calls, repo.c.Status, repo.c.LeaseID)
	}
}

func TestCredentialSourceWaitTimeoutKeepsActiveAuthorization(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	leaseID := uuid.New()
	leaseExpiry := now.Add(time.Minute)
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, EnvelopeVersion: 1, ExpiresAt: now.Add(time.Hour), Status: "active", Generation: 1, Version: 1, LeaseID: &leaseID, LeaseExpiresAt: &leaseExpiry}}
	source, _ := NewCredentialSource(repo, cipher, &refreshStub{}, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, time.Second, time.Second, 20*time.Millisecond)
	if _, err := source.Access(context.Background(), owner); !errors.Is(err, ErrRefreshWaitTimeout) {
		t.Fatalf("error=%v", err)
	}
	if repo.c.Status != "active" {
		t.Fatalf("status=%q", repo.c.Status)
	}
}

func TestCredentialSourceRejectsExpiredRefreshResult(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now, Status: "active", Generation: 1, Version: 1}}
	grant := &refreshStub{result: TokenSet{AccessToken: "expired", RefreshToken: "rotated", Expiry: now.Add(-time.Second)}}
	source, _ := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, time.Second, time.Second, time.Second)
	if _, err := source.Access(context.Background(), owner); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("error=%v", err)
	}
	if repo.c.Status != "reauthorization-required" || repo.c.Version != 1 {
		t.Fatalf("status=%q version=%d", repo.c.Status, repo.c.Version)
	}
}

func TestCredentialSourceSecond401UsesRotatedCredentialVersion(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher, _ := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{8}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{9}, 256)))
	access, _ := cipher.EncryptAccess(owner, "old")
	refresh, _ := cipher.EncryptRefresh(owner, "refresh")
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now.Add(time.Hour), Status: "active", Generation: 1, Version: 1}}
	source, _ := NewCredentialSource(repo, cipher, &refreshStub{result: TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}}, slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now }, time.Second, time.Second, time.Second)
	if err := source.Read(context.Background(), owner, func(context.Context, string) error { return &testUnauthorized{} }); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("error=%v", err)
	}
	if repo.c.Status != "reauthorization-required" {
		t.Fatalf("status=%q", repo.c.Status)
	}
}

type testUnauthorized struct{}

func (*testUnauthorized) Error() string      { return "unauthorized" }
func (*testUnauthorized) Unauthorized() bool { return true }

type refreshStub struct {
	result TokenSet
	err    error
	calls  int
}

func (s *refreshStub) Refresh(context.Context, string) (TokenSet, error) {
	s.calls++
	return s.result, s.err
}

type credentialRepo struct {
	mu sync.Mutex
	c  Credential
}

func (r *credentialRepo) ReadCredential(_ context.Context, _ uuid.UUID) (Credential, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.c, true, nil
}
func (r *credentialRepo) ClaimRefresh(_ context.Context, _ uuid.UUID, expectedVersion int64, now time.Time, lease time.Duration) (Credential, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.c.Version != expectedVersion {
		return r.c, false, nil
	}
	if r.c.LeaseID != nil {
		return r.c, false, nil
	}
	id := uuid.New()
	expires := now.Add(lease)
	r.c.LeaseID, r.c.LeaseExpiresAt = &id, &expires
	return r.c, true, nil
}
func (r *credentialRepo) InstallRefresh(_ context.Context, claim Credential, access, refresh []byte, version int, expiry, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.c.LeaseID == nil || claim.LeaseID == nil || *r.c.LeaseID != *claim.LeaseID {
		return false, nil
	}
	r.c.AccessEnvelope, r.c.RefreshEnvelope, r.c.EnvelopeVersion, r.c.ExpiresAt, r.c.Version, r.c.LeaseID, r.c.LeaseExpiresAt = access, refresh, version, expiry, r.c.Version+1, nil, nil
	return true, nil
}
func (r *credentialRepo) ReleaseRefresh(_ context.Context, claim Credential, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if claim.LeaseID == nil || r.c.LeaseID == nil || *claim.LeaseID != *r.c.LeaseID {
		return false, nil
	}
	r.c.LeaseID, r.c.LeaseExpiresAt = nil, nil
	return true, nil
}
func (r *credentialRepo) RequireReauthorization(_ context.Context, _ uuid.UUID, _ *uuid.UUID, version int64, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.c.Version != version {
		return nil
	}
	r.c.Status = "reauthorization-required"
	return nil
}
