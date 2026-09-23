package portfolio

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

func TestSyncServiceRunPassDrainsClaimsAndLogsSafeOutcome(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	owner := uuid.New()
	accountID := uuid.New()
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), Owner: owner, AccountID: accountID.String(), Resource: AccountResourceBalances}}}
	provider := &accountDataProviderStub{data: completeTestAccountData(now)}
	var logs bytes.Buffer
	service, err := NewSyncService(repository, &inventoryRefresherStub{}, provider, &credentialStub{token: "secret-access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}

	service.RunPass(context.Background())
	if provider.calls != 1 || repository.finishes != 1 || repository.failure != "" || repository.data == nil {
		t.Fatalf("providerCalls=%d finishes=%d failure=%q data=%v", provider.calls, repository.finishes, repository.failure, repository.data)
	}
	output := logs.String()
	if !strings.Contains(output, `"msg":"portfolio sync completed"`) || !strings.Contains(output, `"user_id":"`+owner.String()+`"`) || !strings.Contains(output, `"snaptrade_account_id":"`+accountID.String()+`"`) || strings.Contains(output, "secret-access-token") {
		t.Fatalf("unsafe or incomplete logs: %s", output)
	}
}

func TestSyncServiceRunsOneInventoryBeforeDrainingAccountResources(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	events := make([]string, 0, 3)
	inventory := &inventoryRefresherStub{processed: true, onRefresh: func() { events = append(events, "inventory") }}
	repository := &syncRepositoryStub{
		claims: []SyncClaim{
			{ID: uuid.New(), AccountID: "first", Resource: AccountResourceBalances},
			{ID: uuid.New(), AccountID: "second", Resource: AccountResourcePositions},
		},
		onClaim: func() { events = append(events, "account") },
	}
	service, err := NewSyncService(repository, inventory, &accountDataProviderStub{data: completeTestAccountData(now)}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	service.RunPass(context.Background())

	if inventory.calls != 1 || len(events) != 3 || events[0] != "inventory" || events[1] != "account" || events[2] != "account" {
		t.Fatalf("inventory calls=%d events=%v", inventory.calls, events)
	}
}

func TestSyncServiceDrainsAccountResourcesAfterInventoryFailure(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), AccountID: "account", Resource: AccountResourceBalances}}}
	inventory := &inventoryRefresherStub{processed: true, err: errors.New("scheduled inventory retry")}
	service, err := NewSyncService(repository, inventory, &accountDataProviderStub{data: completeTestAccountData(now)}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	service.RunPass(context.Background())

	if inventory.calls != 1 || repository.finishes != 1 {
		t.Fatalf("inventory calls=%d account finishes=%d", inventory.calls, repository.finishes)
	}
}

func TestSyncServiceCategorizesProviderFailureForDurableRetry(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), Owner: owner, AccountID: "account", Resource: AccountResourceActivities}}}
	retryAt := now.Add(7 * time.Minute)
	provider := &accountDataProviderStub{failAt: 1, err: &ProviderError{State: StateRateLimited, RetryAt: &retryAt}}
	service, _ := NewSyncService(repository, &inventoryRefresherStub{}, provider, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))

	service.RunPass(context.Background())
	if repository.failure != "rate_limited" || repository.data != nil || repository.finishes != 1 || repository.retryAt == nil || !repository.retryAt.Equal(retryAt) {
		t.Fatalf("failure=%q retryAt=%v data=%v finishes=%d", repository.failure, repository.retryAt, repository.data, repository.finishes)
	}
}

func TestSyncServiceOmitsMalformedPersistedAccountIDFromLogs(t *testing.T) {
	now := time.Now().UTC()
	owner := uuid.New()
	const malformedAccountID = "private-persisted-account-reference"
	repository := &syncRepositoryStub{claims: []SyncClaim{{ID: uuid.New(), Owner: owner, AccountID: malformedAccountID, Resource: AccountResourceBalances}}}
	var logs bytes.Buffer
	service, err := NewSyncService(repository, &inventoryRefresherStub{}, &accountDataProviderStub{data: completeTestAccountData(now)}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}

	service.RunPass(context.Background())
	output := logs.String()
	if strings.Contains(output, malformedAccountID) || strings.Contains(output, `"snaptrade_account_id"`) {
		t.Fatalf("malformed account ID leaked into logs: %s", output)
	}
}

func TestRunSyncWorkerStopsPromptlyOnCancellation(t *testing.T) {
	now := time.Now().UTC()
	repository := &syncRepositoryStub{}
	var logs bytes.Buffer
	service, _ := NewSyncService(repository, &inventoryRefresherStub{}, &accountDataProviderStub{}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second, slog.New(slog.NewJSONHandler(&logs, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); RunSyncWorker(ctx, time.Millisecond, service) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
	if !strings.Contains(logs.String(), `"reason":"canceled"`) {
		t.Fatalf("worker stop reason missing from logs: %s", logs.String())
	}
}

func TestWorkerStopReasonCategorizesDeadline(t *testing.T) {
	if actual := workerStopReason(context.DeadlineExceeded); actual != "deadline_exceeded" {
		t.Fatalf("deadline reason = %q", actual)
	}
}

type syncRepositoryStub struct {
	mu       sync.Mutex
	claims   []SyncClaim
	finishes int
	data     *AccountData
	failure  string
	retryAt  *time.Time
	onClaim  func()
}

type inventoryRefresherStub struct {
	processed bool
	err       error
	calls     int
	onRefresh func()
}

func (r *inventoryRefresherStub) RefreshDue(context.Context, time.Duration, time.Duration) (bool, error) {
	r.calls++
	if r.onRefresh != nil {
		r.onRefresh()
	}
	return r.processed, r.err
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
	if r.onClaim != nil {
		r.onClaim()
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
