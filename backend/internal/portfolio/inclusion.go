package portfolio

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// InclusionChangeStatus is the durable outcome of one idempotent selection change.
type InclusionChangeStatus string

const (
	// InclusionPending means required provider probes have not finalized.
	InclusionPending InclusionChangeStatus = "pending"
	// InclusionCommitted means every requested addition published atomically.
	InclusionCommitted InclusionChangeStatus = "committed"
	// InclusionFailed means additions were excluded while removals stayed final.
	InclusionFailed InclusionChangeStatus = "failed"
)

var (
	// ErrInclusionConflict means the expected version is stale.
	ErrInclusionConflict = errors.New("inclusion version conflict")
	// ErrInvalidAccountSelection hides whether an account belongs to another owner.
	ErrInvalidAccountSelection = errors.New("invalid account selection")
	// ErrIdempotencyConflict means a key was reused for a different request.
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
)

const (
	// MaxIncludedAccounts bounds provider load and scheduled synchronization work.
	MaxIncludedAccounts = 5
	// MaxIdempotencyKeyLength matches the public API and persistence contract.
	MaxIdempotencyKeyLength = 200
)

// Balance is a minimized currency-specific account balance.
type Balance struct {
	Currency          string
	Cash, BuyingPower *string
}

// Position is the minimized supported subset of an account position.
type Position struct {
	InstrumentID, Symbol, Kind, Currency string
	Units, Price, CostBasis              *string
}

// Activity is the minimized supported subset of a recent account activity.
type Activity struct {
	ID, Type, Currency        string
	TradeDate                 *time.Time
	Amount, Fee, Price, Units *string
}

// BalanceDataset is one complete, transient balance probe result. A nil
// observation time means the provider supplied retrieval context only.
type BalanceDataset struct {
	ObservedAt  *time.Time
	RetrievedAt time.Time
	Rows        []Balance
}

// PositionDataset is one complete, transient position probe result. Positions
// require provider-supplied observation time as proof of usable holdings data.
type PositionDataset struct {
	ObservedAt  time.Time
	RetrievedAt time.Time
	Rows        []Position
}

// ActivityDataset is one complete, transient recent-activity probe result. A
// nil observation time means the provider supplied retrieval context only.
type ActivityDataset struct {
	ObservedAt  *time.Time
	RetrievedAt time.Time
	Rows        []Activity
}

// AccountData is the three independently versioned dataset probe results for
// one account.
type AccountData struct {
	Balances   BalanceDataset
	Positions  PositionDataset
	Activities ActivityDataset
}

// AccountResource identifies one independently fetched and published dataset.
type AccountResource string

const (
	// AccountResourceBalances selects the balance snapshot.
	AccountResourceBalances AccountResource = "balances"
	// AccountResourcePositions selects the position snapshot.
	AccountResourcePositions AccountResource = "positions"
	// AccountResourceActivities selects the bounded activity page.
	AccountResourceActivities AccountResource = "activities"
)

// InclusionChange describes the current or most recent owner-scoped change.
type InclusionChange struct {
	ID                  uuid.UUID
	Status              InclusionChangeStatus
	Additions, Removals []string
	FailureReason       string
}

// InclusionSnapshot is safe for an authenticated owner response.
type InclusionSnapshot struct {
	Version   int64
	Committed []string
	Change    *InclusionChange
}

// InclusionPreparation describes a saved selection and any first-sync work the
// background worker must finish.
type InclusionPreparation struct {
	InclusionSnapshot
	Claimed             bool
	ChangeID            uuid.UUID
	InventoryGeneration int64
	LifecycleGeneration int64
	Additions           []string
}

// InclusionRepository fences changes and atomically publishes complete datasets.
type InclusionRepository interface {
	GetInclusion(context.Context, uuid.UUID) (InclusionSnapshot, error)
	PrepareInclusion(context.Context, uuid.UUID, int64, string, []string, time.Time) (InclusionPreparation, error)
	FinalizeInclusion(context.Context, uuid.UUID, uuid.UUID, int64, int64, int64, map[string]AccountData, string, *time.Time, time.Time) (InclusionSnapshot, bool, error)
}

// AccountDataProvider performs only the allowlisted account-data reads.
type AccountDataProvider interface {
	LoadAccountData(context.Context, string, string, time.Time) (AccountData, error)
	LoadAccountResource(context.Context, string, string, AccountResource, time.Time) (AccountData, error)
}

// InclusionService durably saves account membership without provider calls.
type InclusionService struct {
	repository InclusionRepository
	clock      func() time.Time
}

// NewInclusionService validates and constructs the inclusion orchestrator.
func NewInclusionService(repository InclusionRepository, clock func() time.Time) (*InclusionService, error) {
	if repository == nil || clock == nil {
		return nil, errors.New("incomplete inclusion service configuration")
	}
	return &InclusionService{repository: repository, clock: clock}, nil
}

// Get returns the authenticated owner's persisted inclusion state.
func (s *InclusionService) Get(ctx context.Context, actor auth.Actor) (InclusionSnapshot, error) {
	return s.repository.GetInclusion(ctx, actor.UserID())
}

// Confirm durably saves the intended selection. First-time additions remain
// pending for the scheduled synchronization worker; retained additions commit
// immediately without provider work.
func (s *InclusionService) Confirm(ctx context.Context, actor auth.Actor, expectedVersion int64, idempotencyKey string, requested []string) (InclusionSnapshot, error) {
	targets := canonicalAccountIDs(requested)
	if len(targets) > MaxIncludedAccounts {
		return InclusionSnapshot{}, ErrInvalidAccountSelection
	}
	preparation, err := s.repository.PrepareInclusion(ctx, actor.UserID(), expectedVersion, idempotencyKey, targets, s.clock().UTC())
	return preparation.InclusionSnapshot, err
}

func inclusionFailureReason(err error) string {
	var categorized *ProviderError
	if errors.As(err, &categorized) {
		switch categorized.State {
		case StateUnauthorized:
			return "authorization_required"
		case StateRateLimited:
			return "rate_limited"
		case StateMalformed:
			return "unusable_data"
		}
	}
	return "provider_unavailable"
}

func canonicalAccountIDs(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			unique[value] = struct{}{}
		}
	}

	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
