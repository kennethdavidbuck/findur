package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kennethdavidbuck/findur/backend/internal/platform/config"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("migration configuration is invalid", "category", "invalid_configuration")
		os.Exit(1)
	}
	if err := migrations.Up(context.Background(), cfg.MigrationURL, cfg.DatabaseURL); err != nil {
		logger.Error("database migration failed", "category", "migration_failure")
		os.Exit(1)
	}

	logger.Info("database migrations are current")
}
