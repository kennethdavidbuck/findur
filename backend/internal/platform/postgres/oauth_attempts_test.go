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

	expired := attempt(2, now.Add(-time.Minute))
	if err := repository.Create(ctx, expired); err != nil {
		t.Fatal(err)
	}
	if err := repository.Claim(ctx, expired.StateHash, expired.BrowserBindingHash, now); !errors.Is(err, auth.ErrNotClaimable) {
		t.Fatalf("expired claim error=%v", err)
	}
	deleted, err := repository.Cleanup(ctx, now)
	if err != nil || deleted != 1 {
		t.Fatalf("first cleanup deleted=%d error=%v", deleted, err)
	}
	deleted, err = repository.Cleanup(ctx, now.Add(11*time.Minute))
	if err != nil || deleted != 1 {
		t.Fatalf("claimed cleanup deleted=%d error=%v", deleted, err)
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
