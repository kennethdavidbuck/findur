// Package httpapi provides the walking skeleton's transport-only HTTP endpoints.
package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

var errNotAccepting = errors.New("process is not accepting work")

type statusResponse struct {
	Status   string `json:"status"`
	BuildSHA string `json:"buildSha"`
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

type requestIDContextKey struct{}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// NewHandler builds the complete HTTP handler with safe request metadata.
// NewHandler builds the API. The optional initiator preserves the diagnostic-only composition.
func NewHandler(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics, initiators ...authorizationInitiator) http.Handler {
	var initiator authorizationInitiator
	if len(initiators) > 0 {
		initiator = initiators[0]
	}
	return newHandler(logger, readiness, buildSHA, diagnostics, initiator, nil, nil, nil, initiator != nil, "")
}

// NewHandlerWithCallback creates the API handler with optional authorization dependencies.
func NewHandlerWithCallback(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics, initiator authorizationInitiator, completer authorizationCompleter, integrationFixtures ...http.Handler) http.Handler {
	return newHandler(logger, readiness, buildSHA, diagnostics, initiator, completer, nil, nil, initiator != nil, "", integrationFixtures...)
}

// NewHandlerWithSessions composes OAuth and the independent session lifecycle.
func NewHandlerWithSessions(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics, initiator authorizationInitiator, completer authorizationCompleter, sessions sessionLifecycle, authorizationAvailable bool, publicOrigin string, integrationFixtures ...http.Handler) http.Handler {
	return newHandler(logger, readiness, buildSHA, diagnostics, initiator, completer, sessions, nil, authorizationAvailable, publicOrigin, integrationFixtures...)
}

// NewHandlerWithInventory composes OAuth, sessions, and masked portfolio inventory.
func NewHandlerWithInventory(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics, initiator authorizationInitiator, completer authorizationCompleter, sessions sessionLifecycle, inventory inventoryLifecycle, authorizationAvailable bool, publicOrigin string, integrationFixtures ...http.Handler) http.Handler {
	return newHandler(logger, readiness, buildSHA, diagnostics, initiator, completer, sessions, inventory, authorizationAvailable, publicOrigin, integrationFixtures...)
}

func newHandler(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics, initiator authorizationInitiator, completer authorizationCompleter, sessions sessionLifecycle, inventory inventoryLifecycle, authorizationAvailable bool, publicOrigin string, integrationFixtures ...http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeStatus(w, http.StatusOK, "ok", buildSHA)
	})
	mux.HandleFunc("GET /api/readyz", func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		if err := readiness.Check(r.Context()); err != nil {
			logger.WarnContext(r.Context(), "readiness check failed",
				"category", readinessCategory(err),
				"latency_ms", time.Since(started).Milliseconds(),
			)
			writeStatus(w, http.StatusServiceUnavailable, "unavailable", buildSHA)
			return
		}
		writeStatus(w, http.StatusOK, "ready", buildSHA)
	})
	if diagnostics != nil {
		diagnostics.register(mux)
	}
	if len(integrationFixtures) > 0 && integrationFixtures[0] != nil {
		mux.Handle("/api/__fixture/oidc/", integrationFixtures[0])
	}
	registerAuthorizationAPI(mux, logger, initiator, completer, sessions, inventory, authorizationAvailable, publicOrigin)

	return requestMetadata(logger, admission(readiness, buildSHA, mux))
}

func admission(readiness *Readiness, buildSHA string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !readiness.IsReady() && r.URL.Path != "/api/healthz" && r.URL.Path != "/api/readyz" {
			writeStatus(w, http.StatusServiceUnavailable, "unavailable", buildSHA)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestMetadata(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := newRequestID()
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		r = r.WithContext(context.WithValue(r.Context(), requestIDContextKey{}, requestID))
		next.ServeHTTP(recorder, r)
		logger.InfoContext(r.Context(), "http request",
			"request_id", requestID,
			"method", r.Method,
			"route", routeCategory(r.URL.Path),
			"status", recorder.status,
			"latency_ms", time.Since(started).Milliseconds(),
		)
	})
}

func routeCategory(path string) string {
	switch path {
	case "/api/healthz":
		return "healthz"
	case "/api/readyz":
		return "readyz"
	case authorizationPath:
		return "authorization_begin"
	case authorizationStatusPath:
		return "authorization_status"
	case logoutPath:
		return "session_logout"
	case portfolioInventoryPath:
		return "portfolio_inventory"
	case portfolioInventoryRetryPath:
		return "portfolio_inventory_retry"
	case auth.SnapTradeCallbackPath:
		return "authorization_callback"
	default:
		if strings.HasPrefix(path, "/api/__fixture/") {
			return "fixture"
		}
		return "unmatched"
	}
}

func writeStatus(w http.ResponseWriter, code int, status, buildSHA string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(statusResponse{Status: status, BuildSHA: buildSHA})
}

func readinessCategory(err error) string {
	switch {
	case errors.Is(err, errNotAccepting):
		return "draining"
	case errors.Is(err, context.Canceled):
		return "request_canceled"
	default:
		return "database_unavailable"
	}
}

func newRequestID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(bytes[:])
}
