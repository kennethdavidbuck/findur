package portfolio

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

func TestServiceClaimsOncePublishesAndServesPersistedSnapshot(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
	encrypted, _ := tokens.EncryptAccess(uuid.Nil, "access-token")
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1, UpdatedAt: now}, Claimed: true, EncryptedToken: encrypted, TokenVersion: 1}}
	provider := &providerStub{connections: []Connection{{ID: "connection", BrokerageLabel: "Broker", Status: "active", SyncMode: "delayed", Available: true}}}
	service, err := NewService(repository, provider, tokens, func() time.Time { return now }, time.Second)
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
			tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
			encrypted, _ := tokens.EncryptAccess(uuid.Nil, "access-token")
			retryAt := now.Add(time.Minute)
			repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true, EncryptedToken: encrypted, TokenVersion: 1}}
			service, _ := NewService(repository, &providerStub{err: &ProviderError{State: state, RetryAt: &retryAt}}, tokens, func() time.Time { return now }, time.Second)
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

func TestServicePublishesTrustworthyPartialRowsWithCategoricalFailure(t *testing.T) {
	now := time.Now().UTC()
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
	encrypted, _ := tokens.EncryptAccess(uuid.Nil, "access-token")
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true, EncryptedToken: encrypted, TokenVersion: 1}}
	partial := []Connection{{ID: "connection", BrokerageLabel: "Broker", Status: "unavailable", SyncMode: "delayed"}}
	service, _ := NewService(repository, &providerStub{connections: partial, err: &ProviderError{State: StateMalformed}}, tokens, func() time.Time { return now }, time.Second)
	got, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || got.State != StateMalformed || len(got.Connections) != 1 || got.Connections[0].BrokerageLabel != "Broker" {
		t.Fatalf("snapshot=%+v err=%v", got, err)
	}
}

func TestServiceTreatsBadTokenEnvelopeAsUnauthorizedWithoutProviderCall(t *testing.T) {
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
	repository := &memoryRepository{preparation: Preparation{Snapshot: Snapshot{State: StatePending, Generation: 1}, Claimed: true, EncryptedToken: []byte("bad"), TokenVersion: 1}}
	provider := &providerStub{}
	service, _ := NewService(repository, provider, tokens, time.Now, time.Second)
	got, err := service.Get(context.Background(), auth.Actor{})
	if err != nil || got.State != StateUnauthorized || provider.calls != 0 {
		t.Fatalf("snapshot=%+v calls=%d err=%v", got, provider.calls, err)
	}
}

type memoryRepository struct {
	preparation Preparation
	finalized   int
}

func (r *memoryRepository) Prepare(context.Context, uuid.UUID, bool, time.Time) (Preparation, error) {
	return r.preparation, nil
}

func (r *memoryRepository) Finalize(_ context.Context, _ uuid.UUID, generation int64, state State, retryAt *time.Time, connections []Connection, now time.Time) (Snapshot, bool, error) {
	if generation != r.preparation.Generation {
		return Snapshot{}, false, errors.New("stale generation")
	}
	r.finalized++
	r.preparation = Preparation{Snapshot: Snapshot{State: state, Generation: generation, RetryAt: retryAt, UpdatedAt: now, Connections: connections}}
	return r.preparation.Snapshot, true, nil
}

type providerStub struct {
	connections []Connection
	err         error
	calls       int
	token       string
}

func (p *providerStub) Load(_ context.Context, token string) ([]Connection, error) {
	p.calls++
	p.token = token
	return p.connections, p.err
}
