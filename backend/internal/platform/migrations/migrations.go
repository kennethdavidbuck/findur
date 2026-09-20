// Package migrations provides the idempotent database migration boundary.
package migrations

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Up applies all pending migrations. An already-current schema is success.
func Up(ctx context.Context, sourceURL, databaseURL string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	migrator, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}

	upErr := runUp(ctx, migrator.Up, migrator.GracefulStop)
	sourceCloseErr, databaseCloseErr := migrator.Close()
	return migrationResult(upErr, sourceCloseErr, databaseCloseErr)
}

func runUp(ctx context.Context, up func() error, gracefulStop chan bool) error {
	result := make(chan error, 1)
	go func() { result <- up() }()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
	}

	select {
	case gracefulStop <- true:
	case err := <-result:
		return errors.Join(ctx.Err(), err)
	}
	return errors.Join(ctx.Err(), <-result)
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
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return fmt.Errorf("apply migrations: %w", err)
}
