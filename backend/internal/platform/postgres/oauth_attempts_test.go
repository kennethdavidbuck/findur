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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/migrations"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOAuthAttemptRepositoryAgainstPostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
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
	repository := postgresadapter.NewOAuthAttemptRepository(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	pending := attempt(1, now.Add(10*time.Minute))
	if err := repository.Create(ctx, pending); err != nil {
		t.Fatal(err)
	}
	if err := repository.Claim(ctx, pending.StateHash, bytes.Repeat([]byte{9}, 32), now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("wrong binding claim error=%v", err)
	}
	if _, err := repository.ClaimCallback(ctx, pending.StateHash, bytes.Repeat([]byte{9}, 32), now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("callback accepted mismatched binding: %v", err)
	}

	const claimers = 16
	results := make(chan error, claimers)
	var start sync.WaitGroup
	start.Add(1)
	for range claimers {
		go func() {
			start.Wait()
			results <- repository.Claim(ctx, pending.StateHash, pending.BrowserBindingHash, now)
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

	completed := attempt(20, now.Add(10*time.Minute))
	if err := repository.Create(ctx, completed); err != nil {
		t.Fatal(err)
	}
	claim, err := repository.ClaimCallback(ctx, completed.StateHash, completed.BrowserBindingHash, now)
	if err != nil || !bytes.Equal(claim.NonceHash, completed.NonceHash) {
		t.Fatalf("callback claim=%+v err=%v", claim, err)
	}
	userID := uuid.New()
	final := auth.Finalization{StateHash: completed.StateHash, Provider: "snaptrade", Subject: "isolated-subject", UserID: userID,
		AccessToken: []byte("encrypted-access"), RefreshToken: []byte("encrypted-refresh"), EnvelopeVersion: 1, TokenExpiresAt: now.Add(time.Hour),
		SessionHash: bytes.Repeat([]byte{30}, 32), CSRFHash: bytes.Repeat([]byte{31}, 32), SessionIdleExpiresAt: now.Add(30 * time.Minute), SessionAbsoluteExpiresAt: now.Add(12 * time.Hour), CompletedAt: now, ReturnRoute: "/portfolio"}
	if err := repository.FinalizeCallback(ctx, final); err != nil {
		t.Fatal(err)
	}
	if active, err := repository.SessionActive(ctx, final.SessionHash, now); err != nil || !active {
		t.Fatalf("active=%v err=%v", active, err)
	}
	replay, err := repository.ClaimCallback(ctx, completed.StateHash, completed.BrowserBindingHash, now)
	if err != nil || replay.TerminalOutcome != "succeeded" || replay.TerminalRoute != "/connect/result" || replay.UserID != userID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	var users, identities, authorizations, sessions int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM external_identities),(SELECT count(*) FROM provider_authorizations),(SELECT count(*) FROM sessions)`).Scan(&users, &identities, &authorizations, &sessions); err != nil {
		t.Fatal(err)
	}
	if users != 1 || identities != 1 || authorizations != 1 || sessions != 1 {
		t.Fatalf("partial finalization: %d %d %d %d", users, identities, authorizations, sessions)
	}
	rollback := attempt(40, now.Add(10*time.Minute))
	if err := repository.Create(ctx, rollback); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.ClaimCallback(ctx, rollback.StateHash, rollback.BrowserBindingHash, now); err != nil {
		t.Fatal(err)
	}
	failedFinal := final
	failedFinal.StateHash, failedFinal.Subject, failedFinal.UserID = rollback.StateHash, "rollback-subject", uuid.New()
	if err := repository.FinalizeCallback(ctx, failedFinal); err == nil {
		t.Fatal("finalization with duplicate session hash succeeded")
	}
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM external_identities),(SELECT count(*) FROM provider_authorizations),(SELECT count(*) FROM sessions)`).Scan(&users, &identities, &authorizations, &sessions); err != nil {
		t.Fatal(err)
	}
	if users != 1 || identities != 1 || authorizations != 1 || sessions != 1 {
		t.Fatalf("rollback left partial state: %d %d %d %d", users, identities, authorizations, sessions)
	}

	expired := attempt(2, now.Add(-time.Minute))
	if err := repository.Create(ctx, expired); err != nil {
		t.Fatal(err)
	}
	if err := repository.Claim(ctx, expired.StateHash, expired.BrowserBindingHash, now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("expired claim error=%v", err)
	}
	if _, err := repository.ClaimCallback(ctx, expired.StateHash, expired.BrowserBindingHash, now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("callback accepted expired attempt: %v", err)
	}
	deleted, err := repository.Cleanup(ctx, now)
	if err != nil || deleted != 1 {
		t.Fatalf("first cleanup deleted=%d error=%v", deleted, err)
	}
	deleted, err = repository.Cleanup(ctx, now.Add(11*time.Minute))
	if err != nil || deleted != 3 {
		t.Fatalf("claimed, terminal, and abandoned-exchange cleanup deleted=%d error=%v", deleted, err)
	}

	canceled, cancelOperation := context.WithCancel(context.Background())
	cancelOperation()
	if _, err := repository.Cleanup(canceled, now); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled cleanup error=%v", err)
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
