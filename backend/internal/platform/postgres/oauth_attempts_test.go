package postgres_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/migrations"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOAuthAttemptRepositoryAgainstPostgres(t *testing.T) {
	fixture := newRepositoryFixture(t)
	t.Run("rejects mismatched browser binding", fixture.testMismatchedBinding)
	t.Run("claims an attempt exactly once", fixture.testConcurrentClaim)
	t.Run("finalizes and replays a callback", fixture.testFinalization)
	t.Run("rolls back partial finalization", fixture.testFinalizationRollback)
	t.Run("expires and cleans attempts", fixture.testExpirationAndCleanup)
	t.Run("honors canceled cleanup", fixture.testCanceledCleanup)
}

type repositoryFixture struct {
	ctx        context.Context
	pool       *pgxpool.Pool
	repository *postgresadapter.OAuthAttemptRepository
	now        time.Time
}

func newRepositoryFixture(t *testing.T) *repositoryFixture {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	container, err := tcpostgres.Run(ctx, "postgres:18.6-alpine",
		tcpostgres.WithDatabase("findur"),
		tcpostgres.WithUsername("findur"),
		tcpostgres.WithPassword("synthetic-repository-test"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		if err := container.Terminate(cleanupCtx); err != nil {
			t.Errorf("terminate PostgreSQL container: %v", err)
		}
	})
	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	_, sourceFile, _, _ := runtime.Caller(0)
	migrationPath, err := filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "../../../db/migrations"))
	if err != nil {
		t.Fatal(err)
	}
	if err := migrations.Up(ctx, "file://"+filepath.ToSlash(migrationPath), databaseURL); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return &repositoryFixture{
		ctx:        ctx,
		pool:       pool,
		repository: postgresadapter.NewOAuthAttemptRepository(pool),
		now:        time.Now().UTC().Truncate(time.Microsecond),
	}
}

func (f *repositoryFixture) testMismatchedBinding(t *testing.T) {
	f.reset(t)
	pending := f.createAttempt(t, 1, f.now.Add(10*time.Minute))
	wrongBinding := bytes.Repeat([]byte{9}, 32)
	if err := f.repository.Claim(f.ctx, pending.StateHash, wrongBinding, f.now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("wrong binding claim error=%v", err)
	}
	if _, err := f.repository.ClaimCallback(f.ctx, pending.StateHash, wrongBinding, f.now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("callback accepted mismatched binding: %v", err)
	}
}

func (f *repositoryFixture) testConcurrentClaim(t *testing.T) {
	f.reset(t)
	pending := f.createAttempt(t, 1, f.now.Add(10*time.Minute))
	const claimers = 16
	results := make(chan error, claimers)
	var start sync.WaitGroup
	start.Add(1)
	for range claimers {
		go func() {
			start.Wait()
			results <- f.repository.Claim(f.ctx, pending.StateHash, pending.BrowserBindingHash, f.now)
		}()
	}
	start.Done()
	var successes int
	for range claimers {
		err := <-results
		if err == nil {
			successes++
		} else if !errors.Is(err, auth.ErrNotClaimable) {
			t.Fatalf("claim error=%v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful claims=%d, want 1", successes)
	}
}

func (f *repositoryFixture) testFinalization(t *testing.T) {
	f.reset(t)
	completed := f.createAttempt(t, 20, f.now.Add(10*time.Minute))
	claim, err := f.repository.ClaimCallback(f.ctx, completed.StateHash, completed.BrowserBindingHash, f.now)
	if err != nil || !bytes.Equal(claim.NonceHash, completed.NonceHash) {
		t.Fatalf("callback claim=%+v err=%v", claim, err)
	}
	userID := uuid.New()
	final := finalization(completed.StateHash, userID, f.now, 30)
	if err := f.repository.FinalizeCallback(f.ctx, final); err != nil {
		t.Fatal(err)
	}
	if active, err := f.repository.SessionActive(f.ctx, final.SessionHash, f.now); err != nil || !active {
		t.Fatalf("active=%v err=%v", active, err)
	}
	replay, err := f.repository.ClaimCallback(f.ctx, completed.StateHash, completed.BrowserBindingHash, f.now)
	if err != nil || replay.TerminalOutcome != "succeeded" || replay.TerminalRoute != auth.AuthorizationResultRoute || replay.UserID != userID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	f.assertRowCounts(t, 1, 1, 1, 1, 1)
}

func (f *repositoryFixture) testFinalizationRollback(t *testing.T) {
	f.reset(t)
	baseline := f.createAttempt(t, 20, f.now.Add(10*time.Minute))
	f.claimCallback(t, baseline)
	baselineFinal := finalization(baseline.StateHash, uuid.New(), f.now, 30)
	if err := f.repository.FinalizeCallback(f.ctx, baselineFinal); err != nil {
		t.Fatal(err)
	}
	rollback := f.createAttempt(t, 40, f.now.Add(10*time.Minute))
	f.claimCallback(t, rollback)
	failedFinal := finalization(rollback.StateHash, uuid.New(), f.now, 30)
	failedFinal.Subject = "distinct-subject"
	if err := f.repository.FinalizeCallback(f.ctx, failedFinal); err == nil {
		t.Fatal("finalization with duplicate session hash succeeded")
	}
	f.assertRowCounts(t, 1, 1, 1, 1, 2)
	f.assertAttemptState(t, rollback.StateHash, "exchanging", false)
}

func (f *repositoryFixture) testExpirationAndCleanup(t *testing.T) {
	f.reset(t)
	expired := f.createAttempt(t, 2, f.now.Add(-time.Minute))
	if err := f.repository.Claim(f.ctx, expired.StateHash, expired.BrowserBindingHash, f.now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("expired claim error=%v", err)
	}
	if _, err := f.repository.ClaimCallback(f.ctx, expired.StateHash, expired.BrowserBindingHash, f.now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("callback accepted expired attempt: %v", err)
	}
	deleted, err := f.repository.Cleanup(f.ctx, f.now)
	if err != nil || deleted != 1 {
		t.Fatalf("first cleanup deleted=%d error=%v", deleted, err)
	}
	exchanging := f.createAttempt(t, 3, f.now.Add(10*time.Minute))
	if err := f.repository.Claim(f.ctx, exchanging.StateHash, exchanging.BrowserBindingHash, f.now); err != nil {
		t.Fatal(err)
	}
	terminal := f.createAttempt(t, 20, f.now.Add(10*time.Minute))
	f.claimCallback(t, terminal)
	if err := f.repository.FailCallback(f.ctx, terminal.StateHash, f.now); err != nil {
		t.Fatal(err)
	}
	abandoned := f.createAttempt(t, 40, f.now.Add(10*time.Minute))
	f.claimCallback(t, abandoned)
	retainedPending := f.createAttempt(t, 60, f.now.Add(20*time.Minute))
	retainedExchanging := f.createAttempt(t, 70, f.now.Add(20*time.Minute))
	if err := f.repository.Claim(f.ctx, retainedExchanging.StateHash, retainedExchanging.BrowserBindingHash, f.now.Add(5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	retainedTerminal := f.createAttempt(t, 80, f.now.Add(20*time.Minute))
	if _, err := f.repository.ClaimCallback(f.ctx, retainedTerminal.StateHash, retainedTerminal.BrowserBindingHash, f.now.Add(5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := f.repository.FailCallback(f.ctx, retainedTerminal.StateHash, f.now.Add(5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	deleted, err = f.repository.Cleanup(f.ctx, f.now.Add(11*time.Minute))
	if err != nil || deleted != 3 {
		t.Fatalf("claimed, terminal, and abandoned-exchange cleanup deleted=%d error=%v", deleted, err)
	}
	for _, removed := range []auth.Attempt{exchanging, terminal, abandoned} {
		f.assertAttemptMissing(t, removed.StateHash)
	}
	f.assertAttemptState(t, retainedPending.StateHash, "pending", false)
	f.assertAttemptState(t, retainedExchanging.StateHash, "exchanging", false)
	f.assertAttemptState(t, retainedTerminal.StateHash, "restart-required", true)
}

func (f *repositoryFixture) testCanceledCleanup(t *testing.T) {
	f.reset(t)
	canceled, cancelOperation := context.WithCancel(context.Background())
	cancelOperation()
	if _, err := f.repository.Cleanup(canceled, f.now); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled cleanup error=%v", err)
	}
}

func (f *repositoryFixture) reset(t *testing.T) {
	t.Helper()
	_, err := f.pool.Exec(f.ctx, `TRUNCATE sessions, provider_authorizations, external_identities, users, oauth_attempts CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
}

func (f *repositoryFixture) createAttempt(t *testing.T, marker byte, expiresAt time.Time) auth.Attempt {
	t.Helper()
	value := attempt(marker, expiresAt)
	if err := f.repository.Create(f.ctx, value); err != nil {
		t.Fatal(err)
	}
	return value
}

func (f *repositoryFixture) claimCallback(t *testing.T, value auth.Attempt) auth.CallbackClaim {
	t.Helper()
	claim, err := f.repository.ClaimCallback(f.ctx, value.StateHash, value.BrowserBindingHash, f.now)
	if err != nil {
		t.Fatal(err)
	}
	return claim
}

func (f *repositoryFixture) assertRowCounts(t *testing.T, wantUsers, wantIdentities, wantAuthorizations, wantSessions, wantAttempts int) {
	t.Helper()
	var users, identities, authorizations, sessions, attempts int
	err := f.pool.QueryRow(f.ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM external_identities),(SELECT count(*) FROM provider_authorizations),(SELECT count(*) FROM sessions),(SELECT count(*) FROM oauth_attempts)`).Scan(&users, &identities, &authorizations, &sessions, &attempts)
	if err != nil {
		t.Fatal(err)
	}
	if users != wantUsers || identities != wantIdentities || authorizations != wantAuthorizations || sessions != wantSessions || attempts != wantAttempts {
		t.Fatalf("row counts users=%d identities=%d authorizations=%d sessions=%d attempts=%d", users, identities, authorizations, sessions, attempts)
	}
}

func (f *repositoryFixture) assertAttemptState(t *testing.T, stateHash []byte, wantStatus string, wantTerminal bool) {
	t.Helper()
	var status string
	var outcome, route *string
	var userID *uuid.UUID
	var completedAt *time.Time
	err := f.pool.QueryRow(f.ctx, `SELECT status,terminal_outcome,terminal_route,user_id,completed_at FROM oauth_attempts WHERE state_hash=$1`, stateHash).
		Scan(&status, &outcome, &route, &userID, &completedAt)
	if err != nil {
		t.Fatal(err)
	}
	terminal := outcome != nil || route != nil || userID != nil || completedAt != nil
	if status != wantStatus || terminal != wantTerminal {
		t.Fatalf("attempt status=%q terminal=%v, want status=%q terminal=%v", status, terminal, wantStatus, wantTerminal)
	}
}

func (f *repositoryFixture) assertAttemptMissing(t *testing.T, stateHash []byte) {
	t.Helper()
	var status string
	if err := f.pool.QueryRow(f.ctx, `SELECT status FROM oauth_attempts WHERE state_hash=$1`, stateHash).Scan(&status); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("removed attempt status=%q error=%v", status, err)
	}
}

func finalization(stateHash []byte, userID uuid.UUID, now time.Time, sessionMarker byte) auth.Finalization {
	return auth.Finalization{
		StateHash: stateHash, Provider: auth.SnapTradeProvider, Subject: "isolated-subject", UserID: userID,
		AccessToken: []byte("encrypted-access"), RefreshToken: []byte("encrypted-refresh"), EnvelopeVersion: 1, TokenExpiresAt: now.Add(time.Hour),
		SessionHash: bytes.Repeat([]byte{sessionMarker}, 32), CSRFHash: bytes.Repeat([]byte{sessionMarker + 1}, 32),
		SessionIdleExpiresAt: now.Add(30 * time.Minute), SessionAbsoluteExpiresAt: now.Add(12 * time.Hour), CompletedAt: now, ReturnRoute: "/portfolio",
	}
}

func attempt(marker byte, expiresAt time.Time) auth.Attempt {
	return auth.Attempt{
		StateHash:          bytes.Repeat([]byte{marker}, 32),
		NonceHash:          bytes.Repeat([]byte{marker + 1}, 32),
		BrowserBindingHash: bytes.Repeat([]byte{marker + 2}, 32),
		EncryptedVerifier:  bytes.Repeat([]byte{marker + 3}, 64),
		ReturnRoute:        "/connect",
		ExpiresAt:          expiresAt,
	}
}
