package migrations

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func TestNormalizeUpErrorTreatsNoChangeAsSuccess(t *testing.T) {
	if err := normalizeUpError(migrate.ErrNoChange); err != nil {
		t.Fatalf("normalizeUpError(ErrNoChange) = %v, want nil", err)
	}
}

func TestRunUpSignalsGracefulStopOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	gracefulStop := make(chan bool)
	done := make(chan error, 1)
	go func() {
		done <- runUp(ctx, func() error {
			close(started)
			<-gracefulStop
			return nil
		}, gracefulStop)
	}()

	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("runUp() error = %v, want context cancellation", err)
	}
}

func TestUpRejectsCanceledContextBeforeOpeningDrivers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Up(ctx, "invalid-source", "invalid-database"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Up() error = %v, want context cancellation", err)
	}
}

func TestNormalizeUpErrorPreservesFailureCategory(t *testing.T) {
	want := errors.New("migration failed")
	got := normalizeUpError(want)
	if !errors.Is(got, want) {
		t.Fatalf("normalizeUpError() = %v, want wrapped %v", got, want)
	}
}

func TestMigrationResultTreatsNoChangeWithCleanCloseAsSuccess(t *testing.T) {
	if err := migrationResult(migrate.ErrNoChange, nil, nil); err != nil {
		t.Fatalf("migrationResult() = %v, want nil", err)
	}
}

func TestMigrationResultJoinsMigrationAndCloseFailures(t *testing.T) {
	upErr := errors.New("migration failed")
	sourceCloseErr := errors.New("source close failed")
	databaseCloseErr := errors.New("database close failed")

	got := migrationResult(upErr, sourceCloseErr, databaseCloseErr)
	for _, want := range []error{upErr, sourceCloseErr, databaseCloseErr} {
		if !errors.Is(got, want) {
			t.Fatalf("migrationResult() = %v, want it to include %v", got, want)
		}
	}
}
