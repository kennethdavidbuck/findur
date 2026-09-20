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
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

type beginFunc func(context.Context, string) (auth.BeginResult, error)

func (f beginFunc) Begin(ctx context.Context, route string) (auth.BeginResult, error) {
	return f(ctx, route)
}

func authorizationHandler(initiator authorizationInitiator) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, initiator)
}

func TestAuthorizationBeginRedirectsWithHostOnlySecureCookie(t *testing.T) {
	handler := authorizationHandler(beginFunc(func(_ context.Context, route string) (auth.BeginResult, error) {
		if route != "/portfolio" {
			t.Fatalf("route=%q", route)
		}
		return auth.BeginResult{AuthorizationURL: "https://provider.example/authorize?scope=openid+read", BrowserBinding: "one-time-secret", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader("returnTo=%2Fportfolio"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") == "" {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies=%d", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Domain != "" || cookie.Path != "/api/auth/snaptrade/callback" {
		t.Fatalf("unsafe cookie: %+v", cookie)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("redirect is cacheable")
	}
}

func TestAuthorizationClosedAndFailuresAreCategoricalAndNoStore(t *testing.T) {
	private := "private-provider-secret"
	for name, handler := range map[string]http.Handler{
		"closed": authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) {
			return auth.BeginResult{}, auth.ErrUnavailable
		})),
		"failed": authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) {
			return auth.BeginResult{}, errors.New(private)
		})),
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", bytes.NewBufferString(`{"returnTo":"/connect"}`))
			request.Header.Set("Content-Type", "application/json")
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusServiceUnavailable || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("response=%d %v", response.Code, response.Header())
			}
			if strings.Contains(response.Body.String(), private) {
				t.Fatal("private failure leaked")
			}
			if len(response.Result().Cookies()) != 0 || response.Header().Get("Location") != "" {
				t.Fatal("failure created browser correlation")
			}
		})
	}
}

func TestAuthorizationGETCannotInitiate(t *testing.T) {
	var called bool
	handler := authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) { called = true; return auth.BeginResult{}, nil }))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/snaptrade/authorize", nil))
	if called || response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodPost || response.Header().Get("Cache-Control") != "no-store" || !strings.Contains(response.Body.String(), "invalid_request") {
		t.Fatalf("called=%v status=%d headers=%v body=%q", called, response.Code, response.Header(), response.Body.String())
	}
}

func TestAuthorizationRejectsUnsupportedContentType(t *testing.T) {
	var called bool
	handler := authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) { called = true; return auth.BeginResult{}, nil }))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader("returnTo=/connect"))
	request.Header.Set("Content-Type", "text/plain")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if called || response.Code != http.StatusUnsupportedMediaType || response.Header().Get("Cache-Control") != "no-store" || !strings.Contains(response.Body.String(), "invalid_request") {
		t.Fatalf("called=%v status=%d body=%q", called, response.Code, response.Body.String())
	}
}

func TestAuthorizationCookieCannotOutliveAttempt(t *testing.T) {
	expires := time.Now().Add(90 * time.Second)
	handler := authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) {
		return auth.BeginResult{AuthorizationURL: "https://provider.example/authorize", BrowserBinding: "binding", ExpiresAt: expires}, nil
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge < 1 || cookies[0].MaxAge > 90 {
		t.Fatalf("cookies=%+v", cookies)
	}
}

func TestAuthorizationPassesUnlistedReturnToDomainPolicy(t *testing.T) {
	var received string
	handler := authorizationHandler(beginFunc(func(_ context.Context, route string) (auth.BeginResult, error) {
		received = route
		return auth.BeginResult{AuthorizationURL: "https://provider.example/authorize", BrowserBinding: "binding", ExpiresAt: time.Now().Add(time.Minute)}, nil
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader(`{"returnTo":"//evil.example"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if received != "//evil.example" || response.Code != http.StatusSeeOther {
		t.Fatalf("received=%q status=%d body=%q", received, response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("response is cacheable")
	}
}

func TestAuthorizationBoundsRequestBodyBeforeInitiation(t *testing.T) {
	var called bool
	handler := authorizationHandler(beginFunc(func(context.Context, string) (auth.BeginResult, error) {
		called = true
		return auth.BeginResult{}, nil
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader(`{"returnTo":"`+strings.Repeat("x", authorizationRequestLimit)+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if called || response.Code != http.StatusBadRequest {
		t.Fatalf("called=%v status=%d body=%q", called, response.Code, response.Body.String())
	}
}

func TestAuthorizationLogsCorrelatedSafeFailure(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandler(logger, readiness, "development", nil, beginFunc(func(ctx context.Context, _ string) (auth.BeginResult, error) {
		<-ctx.Done()
		return auth.BeginResult{}, ctx.Err()
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/snaptrade/authorize", strings.NewReader(`{}`)).WithContext(ctx)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	requestID := response.Header().Get("X-Request-ID")
	if requestID == "" || strings.Count(logs.String(), requestID) != 2 {
		t.Fatalf("request ID not shared by warning and completion logs: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"category":"request_canceled"`) {
		t.Fatalf("missing safe cancellation category: %s", logs.String())
	}
}
