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

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// NewHandler builds the complete HTTP handler with safe request metadata.
func NewHandler(logger *slog.Logger, readiness *Readiness, buildSHA string, diagnostics *Diagnostics) http.Handler {
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
