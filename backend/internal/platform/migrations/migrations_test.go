package migrations

import (
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func TestNormalizeUpErrorTreatsNoChangeAsSuccess(t *testing.T) {
	if err := normalizeUpError(migrate.ErrNoChange); err != nil {
		t.Fatalf("normalizeUpError(ErrNoChange) = %v, want nil", err)
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
