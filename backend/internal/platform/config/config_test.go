package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://findur:secret@localhost/findur?sslmode=disable")
	t.Setenv("PORT", "8080")
	t.Setenv("MIGRATIONS_URL", "file://testdata/migrations")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Address != ":8080" {
		t.Fatalf("Address = %q, want :8080", got.Address)
	}
	if got.MigrationURL != "file://testdata/migrations" {
		t.Fatalf("MigrationURL = %q", got.MigrationURL)
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing DATABASE_URL error")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"0", "65536", "http", "+80", "-1", "80.0"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost/findur")
			t.Setenv("PORT", port)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil for PORT %q, want invalid PORT error", port)
			}
		})
	}
}

func TestOperationalTimeoutsStayWithinDrainBudget(t *testing.T) {
	if ReadinessTimeout >= ShutdownDrain {
		t.Fatalf("readiness timeout %s must be below drain budget %s", ReadinessTimeout, ShutdownDrain)
	}
	for name, timeout := range map[string]time.Duration{
		"read header": ReadHeaderTimeout,
		"read":        ReadTimeout,
		"write":       WriteTimeout,
	} {
		if timeout <= 0 || timeout >= ShutdownDrain {
			t.Fatalf("%s timeout %s must be positive and below drain budget %s", name, timeout, ShutdownDrain)
		}
	}
}
