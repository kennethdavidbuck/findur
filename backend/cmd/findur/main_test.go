package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/kennethdavidbuck/findur/backend/internal/platform/buildinfo"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/config"
)

func TestNewServerUsesBoundedTimeouts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := newServer(":8080", http.NewServeMux(), logger)

	if server.ReadHeaderTimeout != config.ReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != config.ReadTimeout {
		t.Fatalf("ReadTimeout = %s", server.ReadTimeout)
	}
	if server.WriteTimeout != config.WriteTimeout {
		t.Fatalf("WriteTimeout = %s", server.WriteTimeout)
	}
	if server.IdleTimeout != config.IdleTimeout {
		t.Fatalf("IdleTimeout = %s", server.IdleTimeout)
	}
}

func TestRunRejectsMalformedProductionBuildBeforeMigrations(t *testing.T) {
	originalSHA := buildinfo.SHA
	buildinfo.SHA = "malformed"
	t.Cleanup(func() { buildinfo.SHA = originalSHA })
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "not-a-database-url")
	t.Setenv("MIGRATIONS_URL", "not-a-migration-source")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := run(context.Background(), logger); !errors.Is(err, errInvalidConfiguration) {
		t.Fatalf("run() error = %v, want invalid configuration before migrations", err)
	}
}

func TestNewServerLogsStandardLibraryErrorsSafelyAsJSON(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	server := newServer(":8080", http.NewServeMux(), logger)

	server.ErrorLog.Print("postgres://user:secret@private-host/findur")

	if !strings.Contains(logs.String(), `"category":"standard_library"`) {
		t.Fatalf("server error log lacks safe category: %s", logs.String())
	}
	if strings.Contains(logs.String(), "secret") || strings.Contains(logs.String(), "private-host") {
		t.Fatalf("server error log exposes raw error: %s", logs.String())
	}
}

func TestStartupPingResultTreatsRootCancellationAsCleanShutdown(t *testing.T) {
	rootCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := startupPingResult(rootCtx, context.Canceled); err != nil {
		t.Fatalf("startupPingResult() error = %v, want nil", err)
	}
}

func TestStartupPingResultClassifiesDatabaseFailure(t *testing.T) {
	if err := startupPingResult(context.Background(), errors.New("connection failed")); !errors.Is(err, errDatabase) {
		t.Fatalf("startupPingResult() error = %v, want errDatabase", err)
	}
}

func TestFailureCategoryDoesNotExposeUnderlyingError(t *testing.T) {
	if got := failureCategory(errors.New("postgres://user:secret@host/db")); got != "internal_failure" {
		t.Fatalf("failureCategory() = %q, want internal_failure", got)
	}
}

func TestLocalHealthcheckUsesInjectedTypedAddress(t *testing.T) {
	client := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "http://127.0.0.1:18080/api/healthz" {
			t.Fatalf("URL = %q", request.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, nil
	})

	if err := localHealthcheck(config.HealthcheckConfig{URL: "http://127.0.0.1:18080/api/healthz"}, client); err != nil {
		t.Fatalf("localHealthcheck() error = %v", err)
	}
}

func TestLocalHealthcheckRejectsUnhealthyStatus(t *testing.T) {
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	if err := localHealthcheck(config.HealthcheckConfig{URL: "http://127.0.0.1/api/healthz"}, client); err == nil {
		t.Fatal("localHealthcheck() error = nil, want unhealthy status error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(request *http.Request) (*http.Response, error) { return f(request) }
