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
