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

// InclusionPreparation is the short-transaction handoff for provider work.
type InclusionPreparation struct {
	InclusionSnapshot
	Claimed             bool
	ChangeID            uuid.UUID
	InventoryGeneration int64
	LifecycleGeneration int64
	Additions           []string
	EncryptedToken      []byte
	TokenVersion        int
}

// InclusionRepository fences changes and atomically publishes complete datasets.
type InclusionRepository interface {
	GetInclusion(context.Context, uuid.UUID) (InclusionSnapshot, error)
	PrepareInclusion(context.Context, uuid.UUID, int64, string, []string, time.Time) (InclusionPreparation, error)
	FinalizeInclusion(context.Context, uuid.UUID, uuid.UUID, int64, int64, int64, map[string]AccountData, string, time.Time) (InclusionSnapshot, bool, error)
}

// AccountDataProvider performs only the allowlisted account-data reads.
type AccountDataProvider interface {
	LoadAccountData(context.Context, string, string, time.Time) (AccountData, error)
}

// InclusionService coordinates removal-first, call-outside-transaction changes.
type InclusionService struct {
	repository InclusionRepository
	provider   AccountDataProvider
	tokens     *auth.TokenCipher
	clock      func() time.Time
	timeout    time.Duration
}

// NewInclusionService validates and constructs the inclusion orchestrator.
func NewInclusionService(repository InclusionRepository, provider AccountDataProvider, tokens *auth.TokenCipher, clock func() time.Time, timeout time.Duration) (*InclusionService, error) {
	if repository == nil || provider == nil || tokens == nil || clock == nil || timeout <= 0 {
		return nil, errors.New("incomplete inclusion service configuration")
	}
	return &InclusionService{
		repository: repository,
		provider:   provider,
		tokens:     tokens,
		clock:      clock,
		timeout:    timeout,
	}, nil
}

// Get returns the authenticated owner's persisted inclusion state.
func (s *InclusionService) Get(ctx context.Context, actor auth.Actor) (InclusionSnapshot, error) {
	return s.repository.GetInclusion(ctx, actor.UserID())
}

// Confirm immediately commits removals, probes additions outside a transaction,
// then publishes every required dataset and membership in one guarded transaction.
func (s *InclusionService) Confirm(ctx context.Context, actor auth.Actor, expectedVersion int64, idempotencyKey string, requested []string) (InclusionSnapshot, error) {
	targets := canonicalAccountIDs(requested)
	preparation, err := s.repository.PrepareInclusion(ctx, actor.UserID(), expectedVersion, idempotencyKey, targets, s.clock().UTC())
	if err != nil || !preparation.Claimed {
		return preparation.InclusionSnapshot, err
	}

	token, err := s.tokens.DecryptAccess(actor.UserID(), preparation.TokenVersion, preparation.EncryptedToken)
	if err != nil || token == "" {
		return s.finish(ctx, actor.UserID(), preparation, nil, "authorization_required")
	}

	data := make(map[string]AccountData, len(preparation.Additions))
	for _, accountID := range preparation.Additions {
		opCtx, cancel := context.WithTimeout(ctx, s.timeout)
		accountData, providerErr := s.provider.LoadAccountData(opCtx, token, accountID, s.clock().UTC())
		cancel()
		if providerErr != nil {
			return s.finish(ctx, actor.UserID(), preparation, nil, inclusionFailureReason(providerErr))
		}
		data[accountID] = accountData
	}
	return s.finish(ctx, actor.UserID(), preparation, data, "")
}

func (s *InclusionService) finish(ctx context.Context, owner uuid.UUID, preparation InclusionPreparation, data map[string]AccountData, failure string) (InclusionSnapshot, error) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
	defer cancel()

	snapshot, _, err := s.repository.FinalizeInclusion(cleanupCtx, owner, preparation.ChangeID, preparation.Version, preparation.InventoryGeneration, preparation.LifecycleGeneration, data, failure, s.clock().UTC())
	return snapshot, err
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
