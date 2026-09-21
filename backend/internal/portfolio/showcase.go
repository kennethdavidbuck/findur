package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// Freshness is the honest usability category of a persisted dataset.
type Freshness string

// Freshness values describe whether a persisted dataset is safe to present.
const (
	FreshnessCurrent     Freshness = "current"
	FreshnessStale       Freshness = "stale_usable"
	FreshnessExpired     Freshness = "expired"
	FreshnessUnavailable Freshness = "unavailable"
)

// DatasetContext explains the source and timing of one immutable dataset head.
type DatasetContext struct {
	Source, Coverage, Currency           string
	ObservedAt, RetrievedAt, PublishedAt *time.Time
	Freshness                            Freshness
}

// ShowcaseBalance is one persisted cash balance row for an included account.
type ShowcaseBalance struct {
	Currency          string
	Cash, BuyingPower *string
}

// ShowcasePosition is one persisted position row for an included account.
type ShowcasePosition struct {
	Symbol, Kind, Currency  string
	Units, Price, CostBasis *string
}

// ShowcaseActivity is one persisted activity row for an included account.
type ShowcaseActivity struct {
	Type, Currency            string
	TradeDate                 *time.Time
	Amount, Fee, Price, Units *string
}

// ShowcaseDataset pairs persisted rows with their source and freshness context.
type ShowcaseDataset struct {
	Context    DatasetContext
	Balances   []ShowcaseBalance
	Positions  []ShowcasePosition
	Activities []ShowcaseActivity
}

// ShowcaseAccount groups the independently persisted datasets for one account.
type ShowcaseAccount struct {
	Label, Brokerage                string
	SyncMode                        SyncMode
	Balances, Positions, Activities ShowcaseDataset
}

// Showcase is the private, owner-scoped portfolio evidence view.
type Showcase struct{ Accounts []ShowcaseAccount }

// ShowcaseRepository reads only committed owner-scoped dataset heads.
type ShowcaseRepository interface {
	GetShowcase(context.Context, uuid.UUID) (Showcase, error)
}

// ShowcaseService has intentionally no provider or credential dependency.
type ShowcaseService struct {
	repository ShowcaseRepository
	clock      func() time.Time
}

// NewShowcaseService creates a service for owner-scoped persisted showcase data.
func NewShowcaseService(repository ShowcaseRepository, clock func() time.Time) (*ShowcaseService, error) {
	if repository == nil || clock == nil {
		return nil, ErrInvalidAccountSelection
	}
	return &ShowcaseService{repository: repository, clock: clock}, nil
}

// Get returns only the calling actor's committed showcase datasets.
func (s *ShowcaseService) Get(ctx context.Context, actor auth.Actor) (Showcase, error) {
	return s.repository.GetShowcase(ctx, actor.UserID())
}

// ClassifyFreshness applies the published per-dataset policy without refetching.
func ClassifyFreshness(mode SyncMode, activities bool, observed, retrieved *time.Time, now time.Time) Freshness {
	if mode != SyncModeRealtime && mode != SyncModeDelayed {
		return FreshnessUnavailable
	}

	at := observed
	if at == nil {
		at = retrieved
	}
	if at == nil {
		return FreshnessUnavailable
	}
	if activities {
		return activityFreshness(*at, now)
	}

	age := now.Sub(*at)
	if age < 0 {
		age = 0
	}
	current := 36 * time.Hour
	if mode == SyncModeRealtime {
		current = 15 * time.Minute
	}
	stale := 72 * time.Hour
	if age <= current {
		return FreshnessCurrent
	}
	if age <= stale {
		return FreshnessStale
	}
	return FreshnessExpired
}

func activityFreshness(at, now time.Time) Freshness {
	days := calendarDaysBetween(at, now)
	if days <= 2 {
		return FreshnessCurrent
	}
	if days <= 7 {
		return FreshnessStale
	}
	return FreshnessExpired
}

func calendarDaysBetween(earlier, later time.Time) int {
	earlier = earlier.UTC()
	later = later.UTC()
	earlierDay := time.Date(earlier.Year(), earlier.Month(), earlier.Day(), 0, 0, 0, 0, time.UTC)
	laterDay := time.Date(later.Year(), later.Month(), later.Day(), 0, 0, 0, 0, time.UTC)
	return int(laterDay.Sub(earlierDay) / (24 * time.Hour))
}
