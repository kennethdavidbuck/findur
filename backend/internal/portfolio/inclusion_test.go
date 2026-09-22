package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

func TestInclusionServiceDurablySavesPendingAdditionsWithoutProviderWork(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	repository := &inclusionRepositoryStub{preparation: InclusionPreparation{InclusionSnapshot: InclusionSnapshot{
		Version:   3,
		Committed: []string{"kept"},
		Change:    &InclusionChange{ID: uuid.New(), Status: InclusionPending, Additions: []string{"a", "b"}},
	}, Claimed: true, Additions: []string{"a", "b"}}}
	service, _ := NewInclusionService(repository, func() time.Time { return now })

	result, err := service.Confirm(context.Background(), auth.Actor{}, 2, "one", []string{"a", "b", "kept"})
	if err != nil || result.Change == nil || result.Change.Status != InclusionPending {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if repository.finalizeCalls != 0 {
		t.Fatalf("finalizeCalls=%d, save performed synchronization work", repository.finalizeCalls)
	}
}

func TestInclusionServiceRejectsMoreThanFiveUniqueAccountsBeforeRepositoryOrProvider(t *testing.T) {
	now := time.Now().UTC()
	repository := &inclusionRepositoryStub{}
	service, _ := NewInclusionService(repository, func() time.Time { return now })

	_, err := service.Confirm(context.Background(), auth.Actor{}, 0, "too-many", []string{"a", "b", "c", "d", "e", "f", "a"})
	if !errors.Is(err, ErrInvalidAccountSelection) || repository.prepareCalls != 0 {
		t.Fatalf("err=%v prepareCalls=%d", err, repository.prepareCalls)
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
	preparation   InclusionPreparation
	data          map[string]AccountData
	failure       string
	prepareCalls  int
	finalizeCalls int
}

func (r *inclusionRepositoryStub) GetInclusion(context.Context, uuid.UUID) (InclusionSnapshot, error) {
	return r.preparation.InclusionSnapshot, nil
}

func (r *inclusionRepositoryStub) PrepareInclusion(context.Context, uuid.UUID, int64, string, []string, time.Time) (InclusionPreparation, error) {
	r.prepareCalls++
	return r.preparation, nil
}

func (r *inclusionRepositoryStub) FinalizeInclusion(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _, _ int64, data map[string]AccountData, failure string, _ *time.Time, _ time.Time) (InclusionSnapshot, bool, error) {
	r.finalizeCalls++
	r.data, r.failure = data, failure
	status := InclusionCommitted
	if failure == "provider_unavailable" || failure == "rate_limited" {
		status = InclusionPending
	} else if failure != "" {
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
	err    error
}

func (p *accountDataProviderStub) LoadAccountData(context.Context, string, string, time.Time) (AccountData, error) {
	p.calls++
	if p.calls == p.failAt {
		if p.err != nil {
			return AccountData{}, p.err
		}
		return AccountData{}, errors.New("unavailable")
	}
	return p.data, nil
}

func (p *accountDataProviderStub) LoadAccountResource(_ context.Context, _ string, _ string, resource AccountResource, _ time.Time) (AccountData, error) {
	p.calls++
	if p.calls == p.failAt {
		if p.err != nil {
			return AccountData{}, p.err
		}
		return AccountData{}, errors.New("unavailable")
	}
	switch resource {
	case AccountResourceBalances:
		return AccountData{Balances: p.data.Balances}, nil
	case AccountResourcePositions:
		return AccountData{Positions: p.data.Positions}, nil
	case AccountResourceActivities:
		return AccountData{Activities: p.data.Activities}, nil
	default:
		return AccountData{}, errors.New("unknown resource")
	}
}
