// Package migrations provides the idempotent database migration boundary.
package migrations

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Up applies all pending migrations. An already-current schema is success.
func Up(sourceURL, databaseURL string) error {
	migrator, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}

	upErr := migrator.Up()
	sourceCloseErr, databaseCloseErr := migrator.Close()
	return migrationResult(upErr, sourceCloseErr, databaseCloseErr)
}

func migrationResult(upErr, sourceCloseErr, databaseCloseErr error) error {
	var sourceResult error
	if sourceCloseErr != nil {
		sourceResult = fmt.Errorf("close migration source: %w", sourceCloseErr)
	}
	var databaseResult error
	if databaseCloseErr != nil {
		databaseResult = fmt.Errorf("close migration database: %w", databaseCloseErr)
	}
	return errors.Join(normalizeUpError(upErr), sourceResult, databaseResult)
}

func normalizeUpError(err error) error {
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return fmt.Errorf("apply migrations: %w", err)
}
