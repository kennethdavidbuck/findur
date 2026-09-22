package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCredentialSourceAccessTable(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name             string
		expiresAt        time.Time
		invalidAccess    bool
		missingRefresh   bool
		invalidRefresh   bool
		leased           bool
		grantErr         error
		grantResult      TokenSet
		wantToken        string
		wantErr          error
		wantStatus       string
		wantRefreshCalls int
		wantVersion      int64
	}{
		{name: "fresh credential", expiresAt: now.Add(time.Hour), wantToken: "old", wantStatus: CredentialStatusActive, wantVersion: 1},
		{name: "early refresh boundary", expiresAt: now.Add(RefreshEarlyWindow), wantToken: "new", wantStatus: CredentialStatusActive, wantRefreshCalls: 1, wantVersion: 2},
		{name: "expired credential", expiresAt: now.Add(-time.Minute), wantToken: "new", wantStatus: CredentialStatusActive, wantRefreshCalls: 1, wantVersion: 2},
		{name: "discovery failure retries and releases claim", expiresAt: now, grantErr: &RefreshPreSendError{Cause: errors.New("discovery unavailable")}, wantErr: ErrRefreshUnavailable, wantStatus: CredentialStatusActive, wantRefreshCalls: refreshMaxAttempts, wantVersion: 1},
		{name: "ambiguous exchange fails closed", expiresAt: now, grantErr: context.DeadlineExceeded, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantRefreshCalls: 1, wantVersion: 1},
		{name: "unreadable access fails closed", expiresAt: now.Add(time.Hour), invalidAccess: true, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantVersion: 1},
		{name: "missing refresh fails closed", expiresAt: now, missingRefresh: true, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantVersion: 1},
		{name: "unreadable refresh fails closed", expiresAt: now, invalidRefresh: true, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantVersion: 1},
		{name: "expired refresh result fails closed", expiresAt: now, grantResult: TokenSet{AccessToken: "expired", RefreshToken: "rotated", Expiry: now.Add(-time.Second)}, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantRefreshCalls: 1, wantVersion: 1},
		{name: "held lease times out without disabling authorization", expiresAt: now.Add(time.Hour), leased: true, wantErr: ErrRefreshWaitTimeout, wantStatus: CredentialStatusActive, wantVersion: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner := uuid.New()
			cipher := credentialTestCipher(t)
			access, err := cipher.EncryptAccess(owner, "old")
			if err != nil {
				t.Fatal(err)
			}
			refresh, err := cipher.EncryptRefresh(owner, "refresh")
			if err != nil {
				t.Fatal(err)
			}
			if tc.invalidAccess {
				access = []byte("invalid")
			}
			if tc.missingRefresh {
				refresh = nil
			}
			if tc.invalidRefresh {
				refresh = []byte("invalid")
			}
			repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: tc.expiresAt, Status: CredentialStatusActive, Generation: 1, Version: 1}}
			if tc.leased {
				id, expiry := uuid.New(), now.Add(time.Minute)
				repo.c.LeaseID, repo.c.LeaseExpiresAt = &id, &expiry
			}
			result := tc.grantResult
			if result == (TokenSet{}) {
				result = TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}
			}
			grant := &refreshStub{result: result, err: tc.grantErr}
			var logs bytes.Buffer
			source, err := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewJSONHandler(&logs, nil)), time.Now, time.Second, time.Second, 20*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			var got string
			err = source.Read(context.Background(), owner, func(_ context.Context, token string) error {
				got = token
				return nil
			})
			if got != tc.wantToken || !errors.Is(err, tc.wantErr) {
				t.Fatalf("token=%q error=%v; want token=%q error=%v", got, err, tc.wantToken, tc.wantErr)
			}
			if grant.calls != tc.wantRefreshCalls || repo.c.Status != tc.wantStatus || repo.c.Version != tc.wantVersion {
				t.Fatalf("refreshes=%d status=%q version=%d", grant.calls, repo.c.Status, repo.c.Version)
			}
			if errors.Is(tc.wantErr, ErrRefreshUnavailable) && repo.c.LeaseID != nil {
				t.Fatal("pre-send failure retained the refresh lease")
			}
			if tc.wantRefreshCalls == 1 && tc.wantErr == nil {
				for _, event := range []string{"credential_refresh_claimed", "credential_refresh_installed"} {
					if !strings.Contains(logs.String(), `"event":"`+event+`"`) {
						t.Fatalf("missing %s event: %s", event, logs.String())
					}
				}
			}
			for _, secret := range []string{"old", "new", "refresh", "rotated"} {
				if strings.Contains(logs.String(), `"`+secret+`"`) {
					t.Fatalf("credential material appeared in logs: %s", logs.String())
				}
			}
		})
	}
}

func TestCredentialSourceReadTable(t *testing.T) {
	now := time.Now().UTC()
	providerFailure := errors.New("provider unavailable")
	cases := []struct {
		name             string
		read             func(string) error
		wantTokens       []string
		wantErr          error
		wantStatus       string
		wantRefreshCalls int
	}{
		{name: "successful read", read: func(string) error { return nil }, wantTokens: []string{"old"}, wantStatus: CredentialStatusActive},
		{name: "non-auth failure is not retried", read: func(string) error { return providerFailure }, wantTokens: []string{"old"}, wantErr: providerFailure, wantStatus: CredentialStatusActive},
		{name: "401 refreshes and retries once", read: func(token string) error {
			if token == "old" {
				return &testUnauthorized{}
			}
			return nil
		}, wantTokens: []string{"old", "new"}, wantStatus: CredentialStatusActive, wantRefreshCalls: 1},
		{name: "second 401 requires reauthorization", read: func(string) error { return &testUnauthorized{} }, wantTokens: []string{"old", "new"}, wantErr: ErrReauthorizationRequired, wantStatus: CredentialStatusReauthorizationRequired, wantRefreshCalls: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner := uuid.New()
			cipher := credentialTestCipher(t)
			access, err := cipher.EncryptAccess(owner, "old")
			if err != nil {
				t.Fatal(err)
			}
			refresh, err := cipher.EncryptRefresh(owner, "refresh")
			if err != nil {
				t.Fatal(err)
			}
			repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now.Add(time.Hour), Status: CredentialStatusActive, Generation: 1, Version: 1}}
			grant := &refreshStub{result: TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}}
			source, err := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, time.Second, time.Second, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			var tokens []string
			err = source.Read(context.Background(), owner, func(_ context.Context, token string) error {
				tokens = append(tokens, token)
				return tc.read(token)
			})
			if !errors.Is(err, tc.wantErr) || !equalStrings(tokens, tc.wantTokens) || grant.calls != tc.wantRefreshCalls || repo.c.Status != tc.wantStatus {
				t.Fatalf("tokens=%v error=%v refreshes=%d status=%q", tokens, err, grant.calls, repo.c.Status)
			}
		})
	}
}

func TestCredentialSourceConcurrentCallersShareOneRefresh(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	cipher := credentialTestCipher(t)
	access, err := cipher.EncryptAccess(owner, "old")
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := cipher.EncryptRefresh(owner, "refresh")
	if err != nil {
		t.Fatal(err)
	}
	repo := &credentialRepo{c: Credential{Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh, EnvelopeVersion: 1, ExpiresAt: now.Add(-time.Minute), Status: CredentialStatusActive, Generation: 1, Version: 1}}
	grant := &blockingRefreshStub{started: make(chan struct{}), release: make(chan struct{}), result: TokenSet{AccessToken: "new", RefreshToken: "rotated", Expiry: now.Add(time.Hour)}}
	newSource := func() *Source {
		source, err := NewCredentialSource(repo, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, time.Second, time.Second, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		return source
	}
	sources := []*Source{newSource(), newSource()}
	const callers = 6
	results := make(chan string, callers)
	errorsFromCallers := make(chan error, callers)
	var group sync.WaitGroup
	for i := range callers {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			var token string
			err := sources[index%len(sources)].Read(context.Background(), owner, func(_ context.Context, access string) error {
				token = access
				return nil
			})
			results <- token
			errorsFromCallers <- err
		}(i)
		if i == 0 {
			<-grant.started
		}
	}
	close(grant.release)
	group.Wait()
	close(results)
	close(errorsFromCallers)
	for token := range results {
		if token != "new" {
			t.Errorf("token=%q", token)
		}
	}
	for err := range errorsFromCallers {
		if err != nil {
			t.Errorf("access: %v", err)
		}
	}
	if calls := grant.calls.Load(); calls != 1 || repo.c.Version != 2 {
		t.Fatalf("refreshes=%d version=%d", calls, repo.c.Version)
	}
}

func credentialTestCipher(t *testing.T) *TokenCipher {
	t.Helper()
	cipher, err := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{5}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{6}, 4096)))
	if err != nil {
		t.Fatal(err)
	}
	return cipher
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

type blockingRefreshStub struct {
	started chan struct{}
	release chan struct{}
	result  TokenSet
	calls   atomic.Int32
}

func (s *blockingRefreshStub) Refresh(ctx context.Context, _ string) (TokenSet, error) {
	if s.calls.Add(1) == 1 {
		close(s.started)
	}
	select {
	case <-s.release:
		return s.result, nil
	case <-ctx.Done():
		return TokenSet{}, ctx.Err()
	}
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
	r.c.Status = CredentialStatusReauthorizationRequired
	return nil
}
