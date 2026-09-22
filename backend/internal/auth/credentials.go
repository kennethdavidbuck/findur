package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	// RefreshEarlyWindow is the shared buffer before access-token expiry.
	RefreshEarlyWindow = 15 * time.Minute
	// CredentialStatusActive permits provider reads and refreshes.
	CredentialStatusActive = "active"
	// CredentialStatusReauthorizationRequired blocks provider reads until a new grant.
	CredentialStatusReauthorizationRequired = "reauthorization-required"
	refreshMaxAttempts                      = 3
	refreshRetryBaseDelay                   = 100 * time.Millisecond
	refreshWaitInitialDelay                 = 10 * time.Millisecond
	refreshWaitMaxDelay                     = 100 * time.Millisecond
)

// ErrReauthorizationRequired means no safely usable provider credential remains.
var ErrReauthorizationRequired = errors.New("provider authorization must be restarted")

// ErrRefreshUnavailable is a temporary failure before a refresh exchange was sent.
var ErrRefreshUnavailable = errors.New("provider credential refresh temporarily unavailable")

// ErrRefreshWaitTimeout means another caller still owns the refresh lease.
var ErrRefreshWaitTimeout = errors.New("provider credential refresh still in progress")

// RefreshPreSendError identifies a failure known to precede the token exchange.
type RefreshPreSendError struct{ Cause error }

func (e *RefreshPreSendError) Error() string { return "refresh exchange not sent" }
func (e *RefreshPreSendError) Unwrap() error { return e.Cause }

// Credential is the encrypted, version-guarded state of one provider authorization.
type Credential struct {
	Owner                           uuid.UUID
	AccessEnvelope, RefreshEnvelope []byte
	EnvelopeVersion                 int
	ExpiresAt                       time.Time
	Status                          string
	Generation, Version             int64
	LeaseID                         *uuid.UUID
	LeaseExpiresAt                  *time.Time
}

// CredentialRepository performs only short database operations. Network calls
// are deliberately made by Source after a claim transaction has committed.
type CredentialRepository interface {
	ReadCredential(context.Context, uuid.UUID) (Credential, bool, error)
	ClaimRefresh(context.Context, uuid.UUID, int64, time.Time, time.Duration) (Credential, bool, error)
	InstallRefresh(context.Context, Credential, []byte, []byte, int, time.Time, time.Time) (bool, error)
	ReleaseRefresh(context.Context, Credential, time.Time) (bool, error)
	RequireReauthorization(context.Context, uuid.UUID, *uuid.UUID, int64, time.Time) error
}

// RefreshGrant exchanges a rotating refresh token outside database transactions.
type RefreshGrant interface {
	Refresh(context.Context, string) (TokenSet, error)
}

// Source is the sole server-side source of usable provider credentials.
type Source struct {
	repo                 CredentialRepository
	cipher               *TokenCipher
	refresh              RefreshGrant
	logger               *slog.Logger
	clock                func() time.Time
	timeout, lease, wait time.Duration
}

// NewCredentialSource constructs the shared credential and safe-read boundary.
func NewCredentialSource(repo CredentialRepository, cipher *TokenCipher, refresh RefreshGrant, logger *slog.Logger, clock func() time.Time, timeout, lease, wait time.Duration) (*Source, error) {
	if repo == nil || cipher == nil || refresh == nil || logger == nil || clock == nil || timeout <= 0 || lease <= 0 || wait <= 0 {
		return nil, errors.New("incomplete credential source configuration")
	}
	return &Source{repo: repo, cipher: cipher, refresh: refresh, logger: logger, clock: clock, timeout: timeout, lease: lease, wait: wait}, nil
}

// Read executes one safe provider read and recovers one 401 with one forced
// refresh. A repeated unauthorized response disables local authorization.
func (s *Source) Read(ctx context.Context, owner uuid.UUID, read func(context.Context, string) error) error {
	token, version, err := s.access(ctx, owner, false, 0)
	if err != nil {
		return err
	}
	err = read(ctx, token)
	if !isUnauthorized(err) {
		return err
	}
	s.logger.Info("provider read retrying after unauthorized", "event", "provider_401_retry", "user_id", owner.String())
	token, retryVersion, err := s.access(ctx, owner, true, version)
	if err != nil {
		return err
	}
	err = read(ctx, token)
	if isUnauthorized(err) {
		s.requireReauthorization(ctx, owner, nil, retryVersion)
		return ErrReauthorizationRequired
	}
	return err
}

func isUnauthorized(err error) bool {
	var categorized interface{ Unauthorized() bool }
	return errors.As(err, &categorized) && categorized.Unauthorized()
}

func (s *Source) access(ctx context.Context, owner uuid.UUID, force bool, failedVersion int64) (string, int64, error) {
	deadline := s.clock().UTC().Add(s.wait)
	delay := refreshWaitInitialDelay
	for {
		now := s.clock().UTC()
		credential, found, err := s.repo.ReadCredential(ctx, owner)
		if err != nil {
			return "", 0, err
		}
		if !found || credential.Status != CredentialStatusActive {
			return "", 0, ErrReauthorizationRequired
		}
		if canUseAccess(credential, now, force, failedVersion) {
			token, decryptErr := s.decryptAccess(credential)
			if decryptErr != nil {
				s.requireReauthorization(ctx, owner, nil, credential.Version)
			}
			return token, credential.Version, decryptErr
		}
		claim, claimed, err := s.repo.ClaimRefresh(ctx, owner, credential.Version, now, s.lease)
		if err != nil {
			return "", 0, err
		}
		if claimed {
			s.logger.Info("provider credential refresh claimed", "event", "credential_refresh_claimed", "user_id", owner.String())
			token, refreshErr := s.refreshClaim(ctx, claim)
			return token, claim.Version + 1, refreshErr
		}
		if !s.clock().UTC().Before(deadline) {
			s.logger.Warn("provider credential refresh wait expired", "event", "credential_refresh_wait_expired", "user_id", owner.String())
			return "", 0, ErrRefreshWaitTimeout
		}
		// Contenders never repeat the exchange. They use bounded, cancellable
		// backoff while the lease holder installs its rotated credential.
		select {
		case <-ctx.Done():
			return "", 0, ctx.Err()
		case <-time.After(delay):
			s.logger.Debug("provider credential refresh contention wait", "event", "credential_refresh_wait", "user_id", owner.String())
		}
		if delay < refreshWaitMaxDelay {
			delay *= 2
		}
	}
}

func canUseAccess(c Credential, now time.Time, afterUnauthorized bool, failedVersion int64) bool {
	if c.LeaseID != nil {
		return false
	}
	if afterUnauthorized {
		return c.Version != failedVersion && c.ExpiresAt.After(now)
	}
	return c.ExpiresAt.After(now.Add(RefreshEarlyWindow))
}

func (s *Source) decryptAccess(c Credential) (string, error) {
	token, err := s.cipher.DecryptAccess(c.Owner, c.EnvelopeVersion, c.AccessEnvelope)
	if err != nil || token == "" {
		return "", ErrReauthorizationRequired
	}
	return token, nil
}

func (s *Source) refreshClaim(ctx context.Context, claim Credential) (string, error) {
	if claim.LeaseID == nil || len(claim.RefreshEnvelope) == 0 {
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", ErrReauthorizationRequired
	}
	refreshToken, err := s.cipher.DecryptRefresh(claim.Owner, claim.EnvelopeVersion, claim.RefreshEnvelope)
	if err != nil || refreshToken == "" {
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", ErrReauthorizationRequired
	}
	opCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	var result TokenSet
	for attempt := range refreshMaxAttempts {
		result, err = s.refresh.Refresh(opCtx, refreshToken)
		var preSend *RefreshPreSendError
		if !errors.As(err, &preSend) {
			break
		}
		if attempt == refreshMaxAttempts-1 || opCtx.Err() != nil {
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
			released, releaseErr := s.repo.ReleaseRefresh(releaseCtx, claim, s.clock().UTC())
			releaseCancel()
			if releaseErr == nil && released {
				s.logger.Warn("provider refresh temporarily unavailable", "event", "credential_refresh_unavailable", "user_id", claim.Owner.String())
				return "", ErrRefreshUnavailable
			}
			s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
			return "", ErrReauthorizationRequired
		}
		select {
		case <-opCtx.Done():
			continue
		case <-time.After(time.Duration(1<<attempt) * refreshRetryBaseDelay):
		}
	}
	if err != nil || result.AccessToken == "" || result.RefreshToken == "" || !result.Expiry.After(s.clock().UTC()) {
		s.logger.Warn("provider credential refresh failed", "event", "credential_refresh_failed", "user_id", claim.Owner.String())
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", ErrReauthorizationRequired
	}
	finalizeCtx, finalizeCancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
	defer finalizeCancel()
	access, err := s.cipher.EncryptAccess(claim.Owner, result.AccessToken)
	if err != nil {
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", err
	}
	rotated, err := s.cipher.EncryptRefresh(claim.Owner, result.RefreshToken)
	if err != nil {
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", err
	}
	ok, err := s.repo.InstallRefresh(finalizeCtx, claim, access, rotated, s.cipher.CurrentVersion(), result.Expiry.UTC(), s.clock().UTC())
	if err != nil || !ok {
		s.logger.Warn("provider credential refresh finalization failed", "event", "credential_refresh_failed", "user_id", claim.Owner.String())
		s.requireReauthorization(ctx, claim.Owner, claim.LeaseID, claim.Version)
		return "", ErrReauthorizationRequired
	}
	s.logger.Info("provider credential refresh installed", "event", "credential_refresh_installed", "user_id", claim.Owner.String())
	return result.AccessToken, nil
}

func (s *Source) requireReauthorization(ctx context.Context, owner uuid.UUID, lease *uuid.UUID, version int64) {
	s.logger.Warn("provider authorization requires reauthorization", "event", "credential_reauthorization_required", "user_id", owner.String())
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
	defer cancel()
	if err := s.repo.RequireReauthorization(cleanupCtx, owner, lease, version, s.clock().UTC()); err != nil {
		s.logger.Error("provider authorization transition failed", "event", "credential_reauthorization_transition_failed", "user_id", owner.String())
	}
}
