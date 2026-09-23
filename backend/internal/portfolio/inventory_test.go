package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

func TestServiceClaimsOncePublishesAndServesPersistedSnapshot(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1, UpdatedAt: now}, Claimed: true}}
	provider := &providerStub{connections: []Connection{{ID: "connection", BrokerageLabel: "Broker", Status: "active", SyncMode: "delayed", Available: true, Accounts: []Account{{ID: "account"}}}}}
	service, err := NewService(repository, provider, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	first, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || first.State != StateReady || provider.calls != 1 {
		t.Fatalf("first=%+v calls=%d err=%v", first, provider.calls, err)
	}
	second, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || second.State != StateReady || provider.calls != 1 {
		t.Fatalf("second=%+v calls=%d err=%v", second, provider.calls, err)
	}
	if provider.token != "access-token" || repository.finalized != 1 {
		t.Fatalf("token=%q finalized=%d", provider.token, repository.finalized)
	}
}

func TestServicePersistsCategoricalProviderFailures(t *testing.T) {
	for _, state := range []State{StateDisabled, StateUnauthorized, StateRateLimited, StateUnavailable, StateMalformed} {
		t.Run(string(state), func(t *testing.T) {
			now := time.Now().UTC()
			retryAt := now.Add(time.Minute)
			repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true}}
			service, _ := NewService(repository, &providerStub{err: &ProviderError{State: state, RetryAt: &retryAt}}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second)
			got, err := service.Get(context.Background(), auth.Actor{})
			if err != nil || got.State != state {
				t.Fatalf("snapshot=%+v err=%v", got, err)
			}
			if state == StateRateLimited && got.RetryAt == nil {
				t.Fatal("rate limit retry timing was lost")
			}
		})
	}
}

func TestServiceDiscardsPartialRowsOnBulkFailure(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true}}
	partial := []Connection{{ID: "connection", BrokerageLabel: "Broker", Status: "unavailable", SyncMode: "delayed"}}
	service, _ := NewService(repository, &providerStub{connections: partial, err: &ProviderError{State: StateMalformed}}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second)
	got, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || got.State != StateMalformed || len(got.Connections) != 0 {
		t.Fatalf("snapshot=%+v err=%v", got, err)
	}
}

func TestServicePublishesEmptyWhenNoAccountsPassTheServerRules(t *testing.T) {
	now := time.Now().UTC()
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true}}
	provider := &providerStub{connections: []Connection{{ID: "connection", Status: ConnectionStatusActive, Available: true}}}
	service, _ := NewService(repository, provider, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second)
	got, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || got.State != StateEmpty || len(got.Connections) != 1 {
		t.Fatalf("snapshot=%+v err=%v", got, err)
	}
}

func TestServiceTreatsCredentialFailureAsUnauthorizedWithoutProviderCall(t *testing.T) {
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true}}
	provider := &providerStub{}
	service, _ := NewService(repository, provider, &credentialStub{err: auth.ErrReauthorizationRequired}, time.Now, time.Second)
	got, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || got.State != StateUnauthorized || provider.calls != 0 {
		t.Fatalf("snapshot=%+v calls=%d err=%v", got, provider.calls, err)
	}
}

func TestServiceRefreshDueFinalizesScheduledFailureBeforeReturningProviderError(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	retryAt := now.Add(7 * time.Minute)
	providerErr := &ProviderError{State: StateRateLimited, RetryAt: &retryAt}
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 4}, Claimed: true}}
	service, err := NewService(repository, &providerStub{err: providerErr}, &credentialStub{token: "access-token"}, func() time.Time { return now }, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	processed, err := service.RefreshDue(context.Background(), 24*time.Hour, time.Minute)
	if !processed || !errors.Is(err, providerErr) {
		t.Fatalf("processed=%v err=%v", processed, err)
	}
	if repository.scheduledFinalized != 1 || repository.lastState != StateRateLimited || repository.lastRetryAt == nil || !repository.lastRetryAt.Equal(retryAt) {
		t.Fatalf("scheduled=%d state=%s retryAt=%v", repository.scheduledFinalized, repository.lastState, repository.lastRetryAt)
	}
}

type memoryRepository struct {
	preparation        Preparation
	finalized          int
	scheduledFinalized int
	lastState          State
	lastRetryAt        *time.Time
}

func (r *memoryRepository) Prepare(context.Context, uuid.UUID, bool, time.Time) (Preparation, error) {
	return r.preparation, nil
}

func (r *memoryRepository) ClaimDue(context.Context, time.Time, time.Duration, time.Duration) (*ScheduledInventoryClaim, error) {
	if !r.preparation.Claimed {
		return nil, nil
	}
	return &ScheduledInventoryClaim{Generation: r.preparation.Generation}, nil
}

func (r *memoryRepository) Finalize(_ context.Context, _ uuid.UUID, generation int64, state State, retryAt *time.Time, connections []Connection, now time.Time) (Snapshot, bool, error) {
	if generation != r.preparation.Generation {
		return Snapshot{}, false, errors.New("stale generation")
	}
	r.finalized++
	r.lastState, r.lastRetryAt = state, retryAt
	r.preparation = Preparation{Snapshot: Snapshot{State: state, Generation: generation, RetryAt: retryAt, UpdatedAt: now, Connections: connections}}
	return r.preparation.Snapshot, true, nil
}

func (r *memoryRepository) FinalizeScheduled(ctx context.Context, claim ScheduledInventoryClaim, state State, retryAt *time.Time, connections []Connection, now time.Time) (Snapshot, bool, error) {
	r.scheduledFinalized++
	return r.Finalize(ctx, claim.Owner, claim.Generation, state, retryAt, connections, now)
}

type providerStub struct {
	connections []Connection
	err         error
	calls       int
	token       string
}

type credentialStub struct {
	token string
	err   error
}

func (s *credentialStub) Read(ctx context.Context, _ uuid.UUID, read func(context.Context, string) error) error {
	if s.err != nil {
		return s.err
	}
	return read(ctx, s.token)
}

func (p *providerStub) Load(_ context.Context, token string) ([]Connection, error) {
	p.calls++
	p.token = token
	return p.connections, p.err
}
