// Package portfolio owns minimized, owner-scoped portfolio inventory behavior.
package portfolio

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// State is a browser-safe inventory lifecycle category.
type State string

const (
	// StatePending means one bounded provider load has been claimed but not finalized.
	StatePending State = "pending"
	// StateReady means a non-empty masked inventory version was published.
	StateReady State = "ready"
	// StateEmpty means an empty inventory version was published.
	StateEmpty State = "empty"
	// StateDisabled means all connections are disabled or provider access is disabled.
	StateDisabled State = "disabled"
	// StateUnauthorized means the stored bearer is missing, invalid, or rejected.
	StateUnauthorized State = "unauthorized"
	// StateRateLimited means provider retry timing may restrict explicit retry.
	StateRateLimited State = "rate_limited"
	// StateUnavailable means a transport or transient provider failure occurred.
	StateUnavailable State = "unavailable"
	// StateMalformed means the provider response could not be safely normalized.
	StateMalformed State = "malformed"
)

// ConnectionStatus is a normalized provider connection lifecycle category.
type ConnectionStatus string

// Normalized connection status values.
const (
	ConnectionStatusActive      ConnectionStatus = "active"
	ConnectionStatusDisabled    ConnectionStatus = "disabled"
	ConnectionStatusUnavailable ConnectionStatus = "unavailable"
)

// SyncMode is the effective freshness of a connection.
type SyncMode string

// Effective connection freshness values.
const (
	SyncModeRealtime SyncMode = "realtime"
	SyncModeDelayed  SyncMode = "delayed"
	SyncModeUnknown  SyncMode = "unknown"
)

// ResourceDiagnosticReason is a bounded, browser-safe explanation for a
// degraded successful resource response.
type ResourceDiagnosticReason string

// Resource diagnostic reasons shared by inventory connections and datasets.
const (
	DiagnosticNoAccountsReturned    ResourceDiagnosticReason = "no_accounts_returned"
	DiagnosticNoSupportedAccounts   ResourceDiagnosticReason = "no_supported_accounts"
	DiagnosticConnectionDisabled    ResourceDiagnosticReason = "connection_disabled"
	DiagnosticAuthorizationRequired ResourceDiagnosticReason = "authorization_required"
	DiagnosticProviderUnavailable   ResourceDiagnosticReason = "provider_unavailable"
	DiagnosticSyncPending           ResourceDiagnosticReason = "sync_pending"
	DiagnosticUnknown               ResourceDiagnosticReason = "unknown"
)

// ResourceDiagnosticAction is the only recovery action a browser may offer.
type ResourceDiagnosticAction string

// Resource diagnostic actions.
const (
	DiagnosticActionNone      ResourceDiagnosticAction = "none"
	DiagnosticActionWait      ResourceDiagnosticAction = "wait"
	DiagnosticActionRetry     ResourceDiagnosticAction = "retry"
	DiagnosticActionReconnect ResourceDiagnosticAction = "reconnect"
)

// ResourceDiagnostic contains only safe categorical recovery facts.
type ResourceDiagnostic struct {
	Reason                    ResourceDiagnosticReason
	RecommendedAction         ResourceDiagnosticAction
	RetryAt, LastSuccessfulAt *time.Time
}

// AccountCategory is the provider's normalized account category.
type AccountCategory string

// Normalized account category values.
const (
	AccountCategoryInvestment AccountCategory = "investment"
	AccountCategoryDeposit    AccountCategory = "deposit"
	AccountCategoryCredit     AccountCategory = "credit"
	AccountCategoryUnknown    AccountCategory = "unknown"
)

// AccountSyncState is the minimum normalized holdings synchronization state.
type AccountSyncState string

// Normalized account holdings synchronization values.
const (
	AccountSyncStateComplete    AccountSyncState = "complete"
	AccountSyncStatePending     AccountSyncState = "pending"
	AccountSyncStateUnavailable AccountSyncState = "unavailable"
	AccountSyncStateUnknown     AccountSyncState = "unknown"
)

// UsabilityReason explains whether an account can be selected and whether the
// later probe still needs to prove it usable.
type UsabilityReason string

const (
	// UsabilityReady means inventory already proves basic usability.
	UsabilityReady UsabilityReason = "ready"
	// UsabilityProvisionalStatus means status needs the later probe.
	UsabilityProvisionalStatus UsabilityReason = "provisional_status"
	// UsabilityProvisionalCategory means category needs the later probe.
	UsabilityProvisionalCategory UsabilityReason = "provisional_category"
	// UsabilitySyncPending means missing initial holdings make selection unavailable.
	UsabilitySyncPending UsabilityReason = "sync_pending"
	// UsabilityConnectionDisabled means the connection must be repaired.
	UsabilityConnectionDisabled UsabilityReason = "connection_disabled"
	// UsabilityConnectionUnavailable means the connection could not be read.
	UsabilityConnectionUnavailable UsabilityReason = "connection_unavailable"
	// UsabilityAccountClosed means a closed account cannot be selected.
	UsabilityAccountClosed UsabilityReason = "account_closed"
	// UsabilityAccountUnavailable means provider account access is unavailable.
	UsabilityAccountUnavailable UsabilityReason = "account_unavailable"
	// UsabilityUnsupportedCategory means the account is not an investment account.
	UsabilityUnsupportedCategory UsabilityReason = "unsupported_category"
	// UsabilitySyncUnavailable means holdings synchronization is unavailable.
	UsabilitySyncUnavailable UsabilityReason = "sync_unavailable"
)

// Connection is the complete persisted subset of one provider connection.
type Connection struct {
	ID, BrokerageLabel  string
	Status              ConnectionStatus
	SyncMode            SyncMode
	Available, Eligible bool
	Diagnostic          *ResourceDiagnostic
	LastSuccessfulAt    *time.Time
	Accounts            []Account
}

// Account is the complete persisted subset of one provider account.
type Account struct {
	ID, Type, MaskedLabel           string
	TotalBalanceAmount              *string
	TotalBalanceCurrency            *string
	Category                        AccountCategory
	SyncState                       AccountSyncState
	Available, Eligible, Selectable bool
	UsabilityReason                 UsabilityReason
}

// Snapshot is safe to return to an authenticated browser.
type Snapshot struct {
	State       State
	Generation  int64
	RetryAt     *time.Time
	UpdatedAt   time.Time
	Diagnostic  *ResourceDiagnostic
	Connections []Connection
}

// Preparation is returned after a short atomic claim transaction has completed.
type Preparation struct {
	Snapshot
	Claimed        bool
	ClaimExpiresAt *time.Time
}

// ScheduledInventoryClaim is a worker-owned lease for one user's inventory.
type ScheduledInventoryClaim struct {
	Owner                   uuid.UUID
	Generation              int64
	AuthorizationGeneration int64
}

// Repository persists claims, immutable normalized versions, and guarded heads.
type Repository interface {
	Prepare(context.Context, uuid.UUID, bool, time.Time) (Preparation, error)
	ClaimDue(context.Context, time.Time, time.Duration, time.Duration) (*ScheduledInventoryClaim, error)
	Finalize(context.Context, uuid.UUID, int64, State, *time.Time, []Connection, time.Time) (Snapshot, bool, error)
	FinalizeScheduled(context.Context, ScheduledInventoryClaim, State, *time.Time, []Connection, time.Time) (Snapshot, bool, error)
}

// Provider performs the only allowlisted provider inventory operation.
type Provider interface {
	Load(context.Context, string) ([]Connection, error)
}

// CredentialReader centrally supplies tokens and retries one safe unauthorized read.
type CredentialReader interface {
	Read(context.Context, uuid.UUID, func(context.Context, string) error) error
}

// ProviderError contains only a categorical failure and safe retry timing.
type ProviderError struct {
	State   State
	RetryAt *time.Time
}

func (e *ProviderError) Error() string { return "portfolio provider " + string(e.State) }

// Unauthorized lets the credential source classify a safe provider-read 401.
func (e *ProviderError) Unauthorized() bool { return e != nil && e.State == StateUnauthorized }

// Service coordinates claim, provider work outside a transaction, and guarded publication.
type Service struct {
	repository  Repository
	provider    Provider
	credentials CredentialReader
	clock       func() time.Time
	timeout     time.Duration
}

// NewService validates and constructs the portfolio inventory orchestrator.
func NewService(repository Repository, provider Provider, credentials CredentialReader, clock func() time.Time, timeout time.Duration) (*Service, error) {
	if repository == nil || provider == nil || credentials == nil || clock == nil || timeout <= 0 {
		return nil, errors.New("incomplete portfolio service configuration")
	}
	return &Service{repository: repository, provider: provider, credentials: credentials, clock: clock, timeout: timeout}, nil
}

// Get returns a persisted snapshot, claiming the first bootstrap only when none exists.
func (s *Service) Get(ctx context.Context, actor auth.Actor) (Snapshot, error) {
	return s.load(ctx, actor, false)
}

// Retry explicitly claims one new generation when safe retry timing permits it.
func (s *Service) Retry(ctx context.Context, actor auth.Actor) (Snapshot, error) {
	return s.load(ctx, actor, true)
}

// RefreshDue claims and refreshes at most one worker-owned inventory.
func (s *Service) RefreshDue(ctx context.Context, refreshAge, lease time.Duration) (bool, error) {
	claim, err := s.repository.ClaimDue(ctx, s.clock().UTC(), refreshAge, lease)
	if err != nil || claim == nil {
		return false, err
	}
	_, err = s.refresh(ctx, claim.Owner, claim.Generation, claim)
	return true, err
}

func (s *Service) load(ctx context.Context, actor auth.Actor, retry bool) (Snapshot, error) {
	now := s.clock().UTC()
	preparation, err := s.repository.Prepare(ctx, actor.UserID(), retry, now)
	if err != nil || !preparation.Claimed {
		return preparation.Snapshot, err
	}
	return s.refresh(ctx, actor.UserID(), preparation.Generation, nil)
}

func (s *Service) refresh(ctx context.Context, owner uuid.UUID, generation int64, scheduled *ScheduledInventoryClaim) (Snapshot, error) {
	opCtx, cancel := context.WithTimeout(ctx, s.timeout)
	var connections []Connection
	read := func(callCtx context.Context, token string) error {
		var loadErr error
		connections, loadErr = s.provider.Load(callCtx, token)
		return loadErr
	}
	providerErr := s.credentials.Read(opCtx, owner, read)
	cancel()
	if providerErr != nil {
		state, retryAt := StateUnavailable, (*time.Time)(nil)
		if errors.Is(providerErr, auth.ErrReauthorizationRequired) {
			state = StateUnauthorized
		}
		var categorized *ProviderError
		if errors.As(providerErr, &categorized) && validFailureState(categorized.State) {
			state, retryAt = categorized.State, categorized.RetryAt
		}
		snapshot, err := s.finish(ctx, owner, generation, state, retryAt, nil, scheduled)
		if err == nil && scheduled != nil {
			err = providerErr
		}
		return snapshot, err
	}
	state := StateEmpty
	allDisabled := len(connections) > 0
	for _, connection := range connections {
		allDisabled = allDisabled && connection.Status == ConnectionStatusDisabled
		if len(connection.Accounts) > 0 {
			state = StateReady
		}
	}
	if allDisabled {
		state = StateDisabled
	}
	return s.finish(ctx, owner, generation, state, nil, connections, scheduled)
}

func (s *Service) finish(ctx context.Context, owner uuid.UUID, generation int64, state State, retryAt *time.Time, connections []Connection, scheduled *ScheduledInventoryClaim) (Snapshot, error) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
	defer cancel()
	var snapshot Snapshot
	var err error
	if scheduled != nil {
		snapshot, _, err = s.repository.FinalizeScheduled(cleanupCtx, *scheduled, state, retryAt, connections, s.clock().UTC())
	} else {
		snapshot, _, err = s.repository.Finalize(cleanupCtx, owner, generation, state, retryAt, connections, s.clock().UTC())
	}
	return snapshot, err
}

func validFailureState(state State) bool {
	switch state {
	case StateDisabled, StateUnauthorized, StateRateLimited, StateUnavailable, StateMalformed:
		return true
	default:
		return false
	}
}

// SafeLabel trims and bounds provider display strings before they reach persistence.
func SafeLabel(value, fallback string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		value = fallback
	}
	runes := []rune(value)
	if len(runes) > 120 {
		value = string(runes[:120])
	}
	return value
}
