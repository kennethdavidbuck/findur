package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

type memoryRepository struct {
	mu       sync.Mutex
	attempts map[string]Attempt
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{attempts: map[string]Attempt{}}
}
func key(hash []byte) string { return base64.RawStdEncoding.EncodeToString(hash) }
func (r *memoryRepository) Create(_ context.Context, attempt Attempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts[key(attempt.StateHash)] = attempt
	return nil
}
func (r *memoryRepository) Delete(_ context.Context, hash []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, key(hash))
	return nil
}
func (r *memoryRepository) Claim(_ context.Context, state, binding []byte, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.attempts[key(state)]
	if !ok || !a.ExpiresAt.After(now) || a.ReturnRoute == "claimed" || key(a.BrowserBindingHash) != key(binding) {
		return ErrNotClaimable
	}
	a.ReturnRoute = "claimed"
	r.attempts[key(state)] = a
	return nil
}
func (r *memoryRepository) Cleanup(_ context.Context, now time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for k, a := range r.attempts {
		if !a.ExpiresAt.After(now) || a.ReturnRoute == "claimed" {
			delete(r.attempts, k)
			count++
		}
	}
	return count, nil
}

type staticDiscovery struct {
	result Discovery
	err    error
	calls  int
}

type discoveryFunc func(context.Context) (Discovery, error)

func (f discoveryFunc) Discover(ctx context.Context) (Discovery, error) { return f(ctx) }

type errorRepository struct {
	*memoryRepository
	createErr  error
	cleanupErr error
}

func (r *errorRepository) Cleanup(ctx context.Context, now time.Time) (int64, error) {
	if r.cleanupErr != nil {
		return 0, r.cleanupErr
	}
	return r.memoryRepository.Cleanup(ctx, now)
}

type failAfterReader struct{ remaining int }

func (r *failAfterReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, errors.New("random unavailable")
	}
	if len(p) > r.remaining {
		n := r.remaining
		r.remaining = 0
		return n, errors.New("random unavailable")
	}
	r.remaining -= len(p)
	for i := range p {
		p[i] = byte(i + 1)
	}
	return len(p), nil
}

func (r *errorRepository) Create(ctx context.Context, attempt Attempt) error {
	if r.createErr != nil {
		return r.createErr
	}
	return r.memoryRepository.Create(ctx, attempt)
}

func (d *staticDiscovery) Discover(context.Context) (Discovery, error) {
	d.calls++
	return d.result, d.err
}

func serviceForTest(t *testing.T, repository AttemptRepository, discovery DiscoveryProvider, now *time.Time, enabled bool) *Service {
	t.Helper()
	svc, err := NewService(Config{Enabled: enabled, ClientID: "client", CallbackURL: "https://findur.example/api/auth/snaptrade/callback", AllowedReturns: []string{"/connect", "/portfolio"}, DefaultReturn: "/connect", HashKey: make([]byte, 32), EncryptionKey: bytes.Repeat([]byte{1}, 32), Random: rand.Reader, Clock: func() time.Time { return *now }, OperationTimeout: time.Second}, repository, discovery)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestBeginCreatesSecureRequestAndSanitizesReturn(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	discovery := &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}
	svc := serviceForTest(t, repo, discovery, &now, true)
	result, err := svc.Begin(context.Background(), "https://evil.example/steal")
	if err != nil {
		t.Fatal(err)
	}
	location, _ := url.Parse(result.AuthorizationURL)
	query := location.Query()
	if query.Get("scope") != "openid read" || query.Get("code_challenge_method") != "S256" || query.Get("response_type") != "code" {
		t.Fatalf("unsafe authorization query: %v", query)
	}
	if query.Get("state") == "" || query.Get("nonce") == "" || query.Get("code_challenge") == "" || result.BrowserBinding == "" {
		t.Fatal("missing independent correlation values")
	}
	if query.Get("state") == query.Get("nonce") || query.Get("state") == result.BrowserBinding || query.Get("nonce") == result.BrowserBinding {
		t.Fatal("state, nonce, and browser binding are not independent")
	}
	if result.ExpiresAt.Sub(now) != AttemptLifetime {
		t.Fatalf("expiry = %s", result.ExpiresAt)
	}
	for _, a := range repo.attempts {
		if a.ReturnRoute != "/connect" {
			t.Fatalf("return route = %q", a.ReturnRoute)
		}
		nonceSize := svc.aead.NonceSize()
		verifier, err := svc.aead.Open(nil, a.EncryptedVerifier[:nonceSize], a.EncryptedVerifier[nonceSize:], a.StateHash)
		if err != nil || oauth2.S256ChallengeFromVerifier(string(verifier)) != query.Get("code_challenge") {
			t.Fatalf("encrypted verifier cannot reproduce challenge: %v", err)
		}
		if bytes.Contains(a.EncryptedVerifier, verifier) || bytes.Equal(a.StateHash, []byte(query.Get("state"))) || bytes.Equal(a.NonceHash, []byte(query.Get("nonce"))) || bytes.Equal(a.BrowserBindingHash, []byte(result.BrowserBinding)) {
			t.Fatal("attempt persisted plaintext correlation material")
		}
	}
}

func TestBeginUsesSafeDefaultForEveryUnapprovedReturn(t *testing.T) {
	for name, candidate := range map[string]string{
		"external":          "https://evil.example/steal",
		"protocol-relative": "//evil.example/steal",
		"malformed":         "\x00/connect",
		"encoded":           "%2Fportfolio",
		"unlisted":          "/admin",
	} {
		t.Run(name, func(t *testing.T) {
			now := time.Now()
			repo := newMemoryRepository()
			svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
			if _, err := svc.Begin(context.Background(), candidate); err != nil {
				t.Fatal(err)
			}
			for _, stored := range repo.attempts {
				if stored.ReturnRoute != "/connect" {
					t.Fatalf("return route=%q", stored.ReturnRoute)
				}
			}
		})
	}
}

func TestBeginPersistsEveryAllowlistedReturnUnchanged(t *testing.T) {
	for _, route := range []string{"/connect", "/portfolio"} {
		t.Run(route, func(t *testing.T) {
			now := time.Now()
			repo := newMemoryRepository()
			svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
			if _, err := svc.Begin(context.Background(), route); err != nil {
				t.Fatal(err)
			}
			for _, stored := range repo.attempts {
				if stored.ReturnRoute != route {
					t.Fatalf("return route=%q", stored.ReturnRoute)
				}
			}
		})
	}
}

func TestVerifierGenerationFailureIsCategorized(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	svc.config.Random = &failAfterReader{remaining: randomBytes * 2}
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrInitialization) || InitializationStageOf(err) != StageGeneration || len(repo.attempts) != 0 {
		t.Fatalf("err=%v stage=%q attempts=%d", err, InitializationStageOf(err), len(repo.attempts))
	}
}

func TestCleanupFailureStopsBeforeCreate(t *testing.T) {
	now := time.Now()
	repo := &errorRepository{memoryRepository: newMemoryRepository(), cleanupErr: errors.New("cleanup unavailable")}
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrInitialization) || InitializationStageOf(err) != StageStorage || len(repo.attempts) != 0 {
		t.Fatalf("err=%v stage=%q attempts=%d", err, InitializationStageOf(err), len(repo.attempts))
	}
}

func TestClosedGateMakesNoDiscoveryOrAttempt(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	discovery := &staticDiscovery{}
	svc := serviceForTest(t, repo, discovery, &now, false)
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrUnavailable) || discovery.calls != 0 || len(repo.attempts) != 0 {
		t.Fatalf("err=%v calls=%d attempts=%d", err, discovery.calls, len(repo.attempts))
	}
}

func TestDiscoveryFailureLeavesNoAttempt(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	discovery := &staticDiscovery{err: errors.New("private provider body")}
	svc := serviceForTest(t, repo, discovery, &now, true)
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrInitialization) {
		t.Fatalf("err=%v", err)
	}
	if len(repo.attempts) != 0 {
		t.Fatal("failure left reusable attempt")
	}
	if stage := InitializationStageOf(err); stage != StageDiscovery {
		t.Fatalf("stage=%q", stage)
	}
}

func TestInvalidDiscoveredEndpointLeavesNoAttempt(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "/relative"}}, &now, true)
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrInitialization) || InitializationStageOf(err) != StageDiscovery || len(repo.attempts) != 0 {
		t.Fatalf("err=%v stage=%q attempts=%d", err, InitializationStageOf(err), len(repo.attempts))
	}
}

func TestPersistenceFailureIsCategorizedAndLeavesNoAttempt(t *testing.T) {
	now := time.Now()
	repo := &errorRepository{memoryRepository: newMemoryRepository(), createErr: errors.New("database unavailable")}
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, ErrInitialization) || InitializationStageOf(err) != StageStorage || len(repo.attempts) != 0 {
		t.Fatalf("err=%v stage=%q attempts=%d", err, InitializationStageOf(err), len(repo.attempts))
	}
}

func TestBeginHonorsOperationDeadline(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	discovery := discoveryFunc(func(ctx context.Context) (Discovery, error) {
		<-ctx.Done()
		return Discovery{}, ctx.Err()
	})
	svc := serviceForTest(t, repo, discovery, &now, true)
	svc.config.OperationTimeout = 10 * time.Millisecond
	started := time.Now()
	_, err := svc.Begin(context.Background(), "/connect")
	if !errors.Is(err, context.DeadlineExceeded) || InitializationStageOf(err) != StageDiscovery {
		t.Fatalf("err=%v stage=%q", err, InitializationStageOf(err))
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("deadline took %s", elapsed)
	}
	if len(repo.attempts) != 0 {
		t.Fatal("timed out request persisted an attempt")
	}
}

func TestAttemptCanBeClaimedOnlyOnceConcurrently(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	result, err := svc.Begin(context.Background(), "/portfolio")
	if err != nil {
		t.Fatal(err)
	}
	state, _ := url.Parse(result.AuthorizationURL)
	const workers = 16
	outcomes := make(chan error, workers)
	for range workers {
		go func() { outcomes <- svc.Claim(context.Background(), state.Query().Get("state"), result.BrowserBinding) }()
	}
	var successes int
	for range workers {
		if <-outcomes == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful claims = %d, want 1", successes)
	}
}

func TestExpiredAttemptCannotBeClaimedAndIsCleanedUp(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	result, _ := svc.Begin(context.Background(), "/connect")
	location, _ := url.Parse(result.AuthorizationURL)
	now = now.Add(AttemptLifetime)
	if !errors.Is(svc.Claim(context.Background(), location.Query().Get("state"), result.BrowserBinding), ErrNotClaimable) {
		t.Fatal("expired attempt was claimable")
	}
	count, err := svc.Cleanup(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("cleanup = %d, %v", count, err)
	}
}

func TestClaimRejectsWrongStateAndBinding(t *testing.T) {
	now := time.Now()
	repo := newMemoryRepository()
	svc := serviceForTest(t, repo, &staticDiscovery{result: Discovery{AuthorizationEndpoint: "https://provider.example/authorize"}}, &now, true)
	result, err := svc.Begin(context.Background(), "/connect")
	if err != nil {
		t.Fatal(err)
	}
	location, _ := url.Parse(result.AuthorizationURL)
	for name, pair := range map[string][2]string{
		"state":   {"wrong", result.BrowserBinding},
		"binding": {location.Query().Get("state"), "wrong"},
	} {
		t.Run(name, func(t *testing.T) {
			if !errors.Is(svc.Claim(context.Background(), pair[0], pair[1]), ErrNotClaimable) {
				t.Fatal("invalid correlation was claimable")
			}
		})
	}
}
