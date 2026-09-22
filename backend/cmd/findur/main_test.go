package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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

func TestBuildAuthorizationSkipsDisabledFeature(t *testing.T) {
	components, err := buildAuthorization(config.Config{}, nil)
	if err != nil || components.initiator != nil || components.callback != nil || components.fixture != nil {
		t.Fatalf("components=%+v error=%v", components, err)
	}
}

func TestBuildInventoryRemainsAvailableWhenAuthorizationInitiationIsClosed(t *testing.T) {
	providerURL, err := url.Parse("https://api.snaptrade.example")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Session: config.SessionConfig{HashKey: bytes.Repeat([]byte{2}, 32)},
		Authorization: config.AuthorizationConfig{
			Enabled: false, ProviderBaseURL: providerURL, TokenKeys: map[int][]byte{1: bytes.Repeat([]byte{3}, 32)}, CurrentTokenKey: 1,
		},
	}
	components, err := buildPortfolioServices(cfg, &pgxpool.Pool{}, slog.Default())
	if err != nil || components.inventory == nil || components.sync == nil {
		t.Fatalf("components=%+v err=%v", components, err)
	}
}

func TestBuildOIDCFixtureFollowsIntegrationGate(t *testing.T) {
	fixtureURL, err := url.Parse("http://fixture.example")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Integration:    true,
		FixtureBaseURL: fixtureURL,
		Authorization: config.AuthorizationConfig{
			Issuer:       "http://issuer.example",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			CallbackURL:  "http://127.0.0.1:8080/api/auth/snaptrade/callback",
		},
	}
	fixture, err := buildOIDCFixture(cfg)
	if err != nil || fixture == nil {
		t.Fatalf("fixture=%v error=%v", fixture, err)
	}
	cfg.Integration = false
	fixture, err = buildOIDCFixture(cfg)
	if err != nil || fixture != nil {
		t.Fatalf("disabled fixture=%v error=%v", fixture, err)
	}
}

func TestExecuteDispatchesHealthcheckCommand(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"findur", healthcheckCommand}
	t.Cleanup(func() { os.Args = originalArgs })
	t.Setenv("PORT", "invalid")

	if err := execute(); err == nil {
		t.Fatal("execute() error = nil, want invalid healthcheck configuration")
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
