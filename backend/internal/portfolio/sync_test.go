package portfolio

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

func TestSyncServiceRunPassDrainsClaimsAndLogsSafeOutcome(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	owner := uuid.New()
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{2}, 64)))
	encrypted, _ := tokens.EncryptAccess(owner, "secret-access-token")
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), Owner: owner, AccountID: "account-sensitive", EncryptedToken: encrypted, TokenVersion: 1, Resource: AccountResourceBalances}}}
	provider := &accountDataProviderStub{data: completeTestAccountData(now)}
	var logs bytes.Buffer
	service, err := NewSyncService(repository, provider, tokens, func() time.Time { return now }, time.Second, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}

	service.RunPass(context.Background())
	if provider.calls != 1 || repository.finishes != 1 || repository.failure != "" || repository.data == nil {
		t.Fatalf("providerCalls=%d finishes=%d failure=%q data=%v", provider.calls, repository.finishes, repository.failure, repository.data)
	}
	output := logs.String()
	if !strings.Contains(output, `"msg":"portfolio sync completed"`) || strings.Contains(output, "secret-access-token") || strings.Contains(output, "account-sensitive") {
		t.Fatalf("unsafe or incomplete logs: %s", output)
	}
}

func TestSyncServiceCategorizesProviderFailureForDurableRetry(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{3}, 64)))
	encrypted, _ := tokens.EncryptAccess(owner, "access-token")
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), Owner: owner, AccountID: "account", EncryptedToken: encrypted, TokenVersion: 1, Resource: AccountResourceActivities}}}
	retryAt := now.Add(7 * time.Minute)
	provider := &accountDataProviderStub{failAt: 1, err: &ProviderError{State: StateRateLimited, RetryAt: &retryAt}}
	service, _ := NewSyncService(repository, provider, tokens, func() time.Time { return now }, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))

	service.RunPass(context.Background())
	if repository.failure != "rate_limited" || repository.data != nil || repository.finishes != 1 || repository.retryAt == nil || !repository.retryAt.Equal(retryAt) {
		t.Fatalf("failure=%q retryAt=%v data=%v finishes=%d", repository.failure, repository.retryAt, repository.data, repository.finishes)
	}
}

func TestRunSyncWorkerStopsPromptlyOnCancellation(t *testing.T) {
	now := time.Now().UTC()
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{4}, 64)))
	repository := &syncRepositoryStub{}
	service, _ := NewSyncService(repository, &accountDataProviderStub{}, tokens, func() time.Time { return now }, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); RunSyncWorker(ctx, time.Millisecond, service) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
}

type syncRepositoryStub struct {
	mu       sync.Mutex
	claims   []SyncClaim
	finishes int
	data     *AccountData
	failure  string
	retryAt  *time.Time
}

func (r *syncRepositoryStub) AcquireWorkerLease(context.Context, time.Time, time.Duration) (uuid.UUID, bool, error) {
	return uuid.New(), true, nil
}

func (r *syncRepositoryStub) ReleaseWorkerLease(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func (r *syncRepositoryStub) ClaimDue(context.Context, time.Time, time.Duration, time.Duration) (*SyncClaim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.claims) == 0 {
		return nil, nil
	}
	claim := r.claims[0]
	r.claims = r.claims[1:]
	return &claim, nil
}

func (r *syncRepositoryStub) FinishSync(_ context.Context, _ SyncClaim, data *AccountData, failure string, retryAt *time.Time, _ time.Time) (bool, error) {
	r.finishes++
	r.data, r.failure, r.retryAt = data, failure, retryAt
	return failure == "", nil
}
