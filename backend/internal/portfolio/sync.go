package portfolio

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	// SyncWorkerInterval controls attempts to acquire the global pass lease.
	SyncWorkerInterval = time.Minute
	// SyncRefreshAge is the minimum successful bundle age before refresh.
	SyncRefreshAge = 24 * time.Hour
	// SyncClaimLease bounds recovery after a worker disappears.
	SyncClaimLease = time.Minute
	// SyncPassLimit prevents one minute tick from running without bound.
	SyncPassLimit = 45 * time.Second
)

// SyncClaim is a durable, guarded lease for one account resource.
type SyncClaim struct {
	ID                  uuid.UUID
	Owner               uuid.UUID
	AccountID           string
	InclusionVersion    int64
	LifecycleGeneration int64
	InventoryGeneration int64
	Resource            AccountResource
	ChangeID            *uuid.UUID
}

// SyncRepository leases due resources and atomically publishes or retries them.
type SyncRepository interface {
	AcquireWorkerLease(context.Context, time.Time, time.Duration) (uuid.UUID, bool, error)
	ReleaseWorkerLease(context.Context, uuid.UUID, time.Time) error
	ClaimDue(context.Context, time.Time, time.Duration, time.Duration) (*SyncClaim, error)
	FinishSync(context.Context, SyncClaim, *AccountData, string, *time.Time, time.Time) (bool, error)
}

// SyncService drains due account work within a bounded pass.
type SyncService struct {
	repository  SyncRepository
	provider    AccountDataProvider
	credentials CredentialReader
	clock       func() time.Time
	timeout     time.Duration
	logger      *slog.Logger
}

// NewSyncService validates and constructs the scheduled synchronization service.
func NewSyncService(repository SyncRepository, provider AccountDataProvider, credentials CredentialReader, clock func() time.Time, timeout time.Duration, logger *slog.Logger) (*SyncService, error) {
	if repository == nil || provider == nil || credentials == nil || clock == nil || timeout <= 0 || logger == nil {
		return nil, errors.New("incomplete portfolio sync configuration")
	}
	return &SyncService{repository: repository, provider: provider, credentials: credentials, clock: clock, timeout: timeout, logger: logger}, nil
}

// RunPass serially drains work until empty, canceled, or the pass deadline.
func (s *SyncService) RunPass(ctx context.Context) {
	started := s.clock().UTC()
	passCtx, cancel := context.WithTimeout(ctx, SyncPassLimit)
	defer cancel()
	workerClaim, acquired, err := s.repository.AcquireWorkerLease(passCtx, started, SyncClaimLease)
	if err != nil {
		s.logger.Error("portfolio sync worker lease failed", "event", "portfolio_sync_worker_lease_failed", "failure", "worker_lease_failure")
		return
	}
	if !acquired {
		s.logger.Debug("portfolio sync pass skipped; another worker owns the lease", "event", "portfolio_sync_pass_skipped")
		return
	}
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
		defer releaseCancel()
		if err := s.repository.ReleaseWorkerLease(releaseCtx, workerClaim, s.clock().UTC()); err != nil {
			s.logger.Error("portfolio sync worker lease release failed", "event", "portfolio_sync_worker_lease_release_failed", "failure", "worker_lease_release_failure")
		}
	}()
	s.logger.Info("portfolio sync pass started", "event", "portfolio_sync_pass_started")
	processed := 0
	for passCtx.Err() == nil {
		claim, err := s.repository.ClaimDue(passCtx, s.clock().UTC(), SyncRefreshAge, SyncClaimLease)
		if err != nil {
			s.logger.Error("portfolio sync claim failed", "event", "portfolio_sync_claim_failed", "failure", "claim_failure")
			break
		}
		if claim == nil {
			s.logger.Debug("portfolio sync queue empty", "event", "portfolio_sync_queue_empty")
			break
		}
		processed++
		s.syncClaim(passCtx, *claim)
	}
	s.logger.Info("portfolio sync pass finished", "event", "portfolio_sync_pass_finished", "processed", processed, "elapsed_ms", s.clock().UTC().Sub(started).Milliseconds(), "deadline_reached", errors.Is(passCtx.Err(), context.DeadlineExceeded))
}

func (s *SyncService) syncClaim(ctx context.Context, claim SyncClaim) {
	started := s.clock().UTC()
	claimAttrs := syncClaimAttributes(claim)
	s.logger.Info("portfolio sync claimed", append([]any{"event", "portfolio_sync_claimed"}, claimAttrs...)...)
	var data *AccountData
	failure := ""
	var retryAt *time.Time
	opCtx, cancel := context.WithTimeout(ctx, s.timeout)
	read := func(callCtx context.Context, token string) error {
		loaded, loadErr := s.provider.LoadAccountResource(callCtx, token, claim.AccountID, claim.Resource, s.clock().UTC())
		if loadErr == nil {
			data = &loaded
		}
		return loadErr
	}
	providerErr := s.credentials.Read(opCtx, claim.Owner, read)
	cancel()
	if providerErr != nil {
		failure = inclusionFailureReason(providerErr)
		var categorized *ProviderError
		if errors.As(providerErr, &categorized) {
			retryAt = categorized.RetryAt
		}
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
	accepted, finishErr := s.repository.FinishSync(cleanupCtx, claim, data, failure, retryAt, s.clock().UTC())
	cancel()
	attrs := append([]any{"event", "portfolio_sync_finished"}, claimAttrs...)
	attrs = append(attrs, "accepted", accepted, "outcome", syncOutcome(failure), "elapsed_ms", s.clock().UTC().Sub(started).Milliseconds())
	if retryAt != nil {
		attrs = append(attrs, "provider_retry_at", retryAt.UTC())
	}
	switch {
	case finishErr != nil:
		s.logger.Error("portfolio sync finalization failed", append(attrs, "failure", "finalization_failure")...)
	case failure != "":
		s.logger.Warn("portfolio sync scheduled retry", attrs...)
	case !accepted:
		s.logger.Info("portfolio sync result discarded by lifecycle guard", attrs...)
	default:
		s.logger.Info("portfolio sync completed", attrs...)
	}
}

func syncClaimAttributes(claim SyncClaim) []any {
	attrs := []any{"user_id", claim.Owner.String(), "claim_id", claim.ID.String(), "resource", claim.Resource}
	if accountID, err := uuid.Parse(claim.AccountID); err == nil && accountID != uuid.Nil {
		attrs = append(attrs, "snaptrade_account_id", accountID.String())
	}
	return attrs
}

func syncOutcome(failure string) string {
	if failure == "" {
		return "succeeded"
	}
	return failure
}

// RunSyncWorker starts one immediate pass and then one pass per interval.
func RunSyncWorker(ctx context.Context, interval time.Duration, service *SyncService) {
	service.RunPass(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			service.logger.Info("portfolio sync worker stopped", "event", "portfolio_sync_worker_stopped", "reason", workerStopReason(ctx.Err()))
			return
		case <-ticker.C:
			service.RunPass(ctx)
		}
	}
}

func workerStopReason(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	return "canceled"
}
