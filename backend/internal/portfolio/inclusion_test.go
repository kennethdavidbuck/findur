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

func TestInclusionServicePublishesAllAdditionsTogether(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
	encrypted, _ := tokens.EncryptAccess(uuid.Nil, "access-token")
	repository := &inclusionRepositoryStub{preparation: InclusionPreparation{InclusionSnapshot: InclusionSnapshot{Version: 3, Committed: []string{"kept"}}, Claimed: true, ChangeID: uuid.New(), InventoryGeneration: 4, LifecycleGeneration: 1, Additions: []string{"a", "b"}, EncryptedToken: encrypted, TokenVersion: 1}}
	provider := &accountDataProviderStub{data: completeTestAccountData(now)}
	service, _ := NewInclusionService(repository, provider, tokens, func() time.Time { return now }, time.Second)

	result, err := service.Confirm(context.Background(), auth.Actor{}, 2, "one", []string{"a", "b", "kept"})
	if err != nil || result.Change == nil || result.Change.Status != InclusionCommitted {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if provider.calls != 2 || repository.failure != "" || len(repository.data) != 2 {
		t.Fatalf("calls=%d failure=%q data=%v", provider.calls, repository.failure, repository.data)
	}
}

func TestInclusionServiceNeverPublishesPartialAdditions(t *testing.T) {
	now := time.Now().UTC()
	tokens, _ := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
	encrypted, _ := tokens.EncryptAccess(uuid.Nil, "access-token")
	repository := &inclusionRepositoryStub{preparation: InclusionPreparation{InclusionSnapshot: InclusionSnapshot{Version: 8, Committed: []string{"kept"}}, Claimed: true, ChangeID: uuid.New(), InventoryGeneration: 6, LifecycleGeneration: 2, Additions: []string{"a", "b"}, EncryptedToken: encrypted, TokenVersion: 1}}
	provider := &accountDataProviderStub{data: completeTestAccountData(now), failAt: 2}
	service, _ := NewInclusionService(repository, provider, tokens, func() time.Time { return now }, time.Second)

	result, err := service.Confirm(context.Background(), auth.Actor{}, 7, "two", []string{"a", "b", "kept"})
	if err != nil || result.Change == nil || result.Change.Status != InclusionFailed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if repository.data != nil || repository.failure != "provider_unavailable" {
		t.Fatalf("partial data reached finalizer: data=%v failure=%q", repository.data, repository.failure)
	}
}

func completeTestAccountData(now time.Time) AccountData {
	return AccountData{
		Balances:   BalanceDataset{RetrievedAt: now, Rows: []Balance{}},
		Positions:  PositionDataset{ObservedAt: now, RetrievedAt: now, Rows: []Position{}},
		Activities: ActivityDataset{RetrievedAt: now, Rows: []Activity{}},
	}
}

type inclusionRepositoryStub struct {
	preparation InclusionPreparation
	data        map[string]AccountData
	failure     string
}

func (r *inclusionRepositoryStub) GetInclusion(context.Context, uuid.UUID) (InclusionSnapshot, error) {
	return r.preparation.InclusionSnapshot, nil
}

func (r *inclusionRepositoryStub) PrepareInclusion(context.Context, uuid.UUID, int64, string, []string, time.Time) (InclusionPreparation, error) {
	return r.preparation, nil
}

func (r *inclusionRepositoryStub) FinalizeInclusion(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _, _ int64, data map[string]AccountData, failure string, _ time.Time) (InclusionSnapshot, bool, error) {
	r.data, r.failure = data, failure
	status := InclusionCommitted
	if failure != "" {
		status = InclusionFailed
	}
	result := r.preparation.InclusionSnapshot
	result.Change = &InclusionChange{ID: r.preparation.ChangeID, Status: status, Additions: r.preparation.Additions, FailureReason: failure}
	return result, failure == "", nil
}

type accountDataProviderStub struct {
	data   AccountData
	failAt int
	calls  int
}

func (p *accountDataProviderStub) LoadAccountData(context.Context, string, string, time.Time) (AccountData, error) {
	p.calls++
	if p.calls == p.failAt {
		return AccountData{}, errors.New("unavailable")
	}
	return p.data, nil
}
