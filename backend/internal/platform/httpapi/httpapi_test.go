package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type pingFunc func(context.Context) error

func (f pingFunc) Ping(ctx context.Context) error { return f(ctx) }

func testHandler(readiness *Readiness) http.Handler {
	return NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil)
}

func TestLivenessDoesNotTouchDatabase(t *testing.T) {
	var calls atomic.Int32
	readiness := NewReadiness(pingFunc(func(context.Context) error {
		calls.Add(1)
		return errors.New("database should not be called")
	}), time.Second)

	request := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	response := httptest.NewRecorder()
	testHandler(readiness).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if calls.Load() != 0 {
		t.Fatalf("database calls = %d, want 0", calls.Load())
	}
	if body := response.Body.String(); body != "{\"status\":\"ok\",\"buildSha\":\"development\"}\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestReadinessSucceedsWhenDatabaseResponds(t *testing.T) {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)

	request := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	response := httptest.NewRecorder()
	testHandler(readiness).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID is empty")
	}
}

func TestReadinessFailsSafelyWhenDatabaseIsUnavailable(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	readiness := NewReadiness(pingFunc(func(context.Context) error {
		return errors.New("postgres://user:secret@private-host/findur")
	}), time.Second)
	readiness.SetReady(true)

	request := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	response := httptest.NewRecorder()
	NewHandler(logger, readiness, "development", nil).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if body := response.Body.String(); body != "{\"status\":\"unavailable\",\"buildSha\":\"development\"}\n" {
		t.Fatalf("body = %q", body)
	}
	if strings.Contains(response.Body.String(), "secret") || strings.Contains(response.Body.String(), "private-host") {
		t.Fatalf("response exposes database details: %q", response.Body.String())
	}
	if !strings.Contains(logs.String(), `"category":"database_unavailable"`) {
		t.Fatalf("logs lack safe failure category: %s", logs.String())
	}
	if strings.Contains(logs.String(), "secret") || strings.Contains(logs.String(), "private-host") {
		t.Fatalf("logs expose database details: %s", logs.String())
	}
}

func TestBothProbesReportIndependentlyInjectedBuildSHA(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, sha, nil)
	for _, path := range []string{"/api/healthz", "/api/readyz"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if !strings.Contains(response.Body.String(), `"buildSha":"`+sha+`"`) {
			t.Fatalf("%s body = %q, want build SHA", path, response.Body.String())
		}
	}
}

func TestReadinessTimesOut(t *testing.T) {
	readiness := NewReadiness(pingFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}), 10*time.Millisecond)
	readiness.SetReady(true)

	request := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	response := httptest.NewRecorder()
	started := time.Now()
	testHandler(readiness).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("readiness took %s, want a bounded timeout", elapsed)
	}
}

func TestShutdownTransitionFailsReadinessWithoutDatabaseCall(t *testing.T) {
	var calls atomic.Int32
	readiness := NewReadiness(pingFunc(func(context.Context) error {
		calls.Add(1)
		return nil
	}), time.Second)
	readiness.SetReady(true)
	readiness.SetReady(false)

	request := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	response := httptest.NewRecorder()
	testHandler(readiness).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if calls.Load() != 0 {
		t.Fatalf("database calls = %d, want 0 after drain starts", calls.Load())
	}
}

func TestReadinessFailsWhenDrainBeginsDuringPing(t *testing.T) {
	pingStarted := make(chan struct{})
	releasePing := make(chan struct{})
	readiness := NewReadiness(pingFunc(func(context.Context) error {
		close(pingStarted)
		<-releasePing
		return nil
	}), time.Second)
	readiness.SetReady(true)

	result := make(chan error, 1)
	go func() {
		result <- readiness.Check(context.Background())
	}()
	<-pingStarted
	readiness.SetReady(false)
	close(releasePing)

	if err := <-result; !errors.Is(err, errNotAccepting) {
		t.Fatalf("Check() error = %v, want errNotAccepting", err)
	}
}

func TestClientCancellationUsesSafeCategory(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	readiness := NewReadiness(pingFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}), time.Second)
	readiness.SetReady(true)

	requestCtx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodGet, "/api/readyz", nil).WithContext(requestCtx)
	response := httptest.NewRecorder()
	NewHandler(logger, readiness, "development", nil).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(logs.String(), `"category":"request_canceled"`) {
		t.Fatalf("logs lack request cancellation category: %s", logs.String())
	}
}

func TestStatusWriterUnwrapPreservesResponseControllerCapabilities(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusWriter{ResponseWriter: recorder}

	if writer.Unwrap() != recorder {
		t.Fatal("Unwrap() did not return the wrapped response writer")
	}
	if err := http.NewResponseController(writer).Flush(); err != nil {
		t.Fatalf("Flush() through statusWriter error = %v", err)
	}
}

func TestRequestLogsDoNotContainUnmatchedPathDetails(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)

	request := httptest.NewRequest(http.MethodGet, "/secret-value", nil)
	response := httptest.NewRecorder()
	NewHandler(logger, readiness, "development", nil).ServeHTTP(response, request)

	if strings.Contains(logs.String(), "secret-value") {
		t.Fatalf("request log exposes unmatched path: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"route":"unmatched"`) {
		t.Fatalf("request log lacks safe route category: %s", logs.String())
	}
}

func TestRequestLogHandlerAddsOnlyApprovedRequestMetadata(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(NewRequestLogHandler(slog.NewJSONHandler(&logs, nil))).With("request_id", "caller-request-id", "user_id", "caller-user-id", "snaptrade_account_ids", []string{"caller-account-id"})
	metadata := &requestLogMetadata{
		requestID:           "request-opaque-id",
		findurUserID:        "b24b69c1-d6c6-4ae3-83a5-35b39950dc2e",
		snapTradeAccountIDs: []string{"6f1ee24e-4f23-4fdd-8d1c-93b51f55580a"},
	}
	ctx := context.WithValue(context.Background(), requestLogMetadataContextKey{}, metadata)
	logger.InfoContext(ctx, "handler event", "resource", "portfolio", "request_id", "event-request-id", "user_id", "event-user-id", "snaptrade_account_ids", []string{"event-account-id"})

	output := logs.String()
	for _, expected := range []string{`"request_id":"request-opaque-id"`, `"user_id":"b24b69c1-d6c6-4ae3-83a5-35b39950dc2e"`, `"snaptrade_account_ids":["6f1ee24e-4f23-4fdd-8d1c-93b51f55580a"]`, `"resource":"portfolio"`} {
		if !strings.Contains(output, expected) {
			t.Fatalf("log missing %s: %s", expected, output)
		}
	}
	for _, forbidden := range []string{"caller-request-id", "caller-user-id", "caller-account-id", "event-request-id", "event-user-id", "event-account-id"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("caller supplied reserved metadata reached log: %s", output)
		}
	}
	for _, key := range []string{`"request_id"`, `"user_id"`, `"snaptrade_account_ids"`} {
		if strings.Count(output, key) != 1 {
			t.Fatalf("reserved key %s was duplicated: %s", key, output)
		}
	}
}
