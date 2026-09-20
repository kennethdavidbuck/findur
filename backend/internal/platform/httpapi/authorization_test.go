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
type completeFunc func(context.Context, auth.CallbackInput) (auth.CallbackResult, error)
type statusCompleter struct {
	status          auth.AuthorizationStatus
	receivedSession string
}

type sessionLifecycleStub struct {
	authorizeErr   error
	revokeErr      error
	revoked        []string
	authorizeCalls int
}

func (s *sessionLifecycleStub) Authenticate(context.Context, string) (auth.Actor, error) {
	return auth.Actor{}, s.authorizeErr
}
func (s *sessionLifecycleStub) AuthorizeUnsafe(context.Context, string, string) (auth.Actor, error) {
	s.authorizeCalls++
	return auth.Actor{}, s.authorizeErr
}

func TestLogoutRejectsCrossOriginBeforeTouchingTheSession(t *testing.T) {
	lifecycle := &sessionLifecycleStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	request.Header.Set("X-CSRF-Token", "csrf")
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	response := httptest.NewRecorder()
	sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || lifecycle.authorizeCalls != 0 {
		t.Fatalf("status=%d authenticate calls=%d", response.Code, lifecycle.authorizeCalls)
	}
}
func (s *sessionLifecycleStub) RevokeCurrent(_ context.Context, session string) error {
	s.revoked = append(s.revoked, session)
	return s.revokeErr
}

func (s *statusCompleter) Complete(context.Context, auth.CallbackInput) (auth.CallbackResult, error) {
	return auth.CallbackResult{}, nil
}
func (s *statusCompleter) Status(_ context.Context, session string) (auth.AuthorizationStatus, error) {
	s.receivedSession = session
	return s.status, nil
}

func (f beginFunc) Begin(ctx context.Context, route string) (auth.BeginResult, error) {
	return f(ctx, route)
}
func (f completeFunc) Complete(ctx context.Context, input auth.CallbackInput) (auth.CallbackResult, error) {
	return f(ctx, input)
}

func callbackHandler(completer authorizationCompleter) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithCallback(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, completer)
}

func sessionHandler(sessions sessionLifecycle, publicOrigin string) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithSessions(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, nil, sessions, false, publicOrigin)
}

func TestLogoutRequiresSessionBoundCSRFOriginAndFetchMetadata(t *testing.T) {
	for name, configure := range map[string]func(*http.Request, *sessionLifecycleStub){
		"missing session": func(request *http.Request, _ *sessionLifecycleStub) {
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
		},
		"invalid session": func(request *http.Request, lifecycle *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "expired"})
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
			lifecycle.authorizeErr = auth.ErrUnauthenticated
		},
		"invalid csrf": func(request *http.Request, lifecycle *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
			lifecycle.authorizeErr = auth.ErrForbidden
		},
		"cross origin": func(request *http.Request, _ *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			request.Header.Set("Origin", "https://evil.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
		},
		"cross site metadata": func(request *http.Request, _ *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "cross-site")
		},
		"missing origin and referer": func(request *http.Request, _ *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			request.Header.Set("Sec-Fetch-Site", "same-origin")
		},
		"missing fetch metadata": func(request *http.Request, _ *sessionLifecycleStub) {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			request.Header.Set("Origin", "https://findur.example")
		},
	} {
		t.Run(name, func(t *testing.T) {
			lifecycle := &sessionLifecycleStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			request.Header.Set("X-CSRF-Token", "csrf")
			configure(request, lifecycle)
			response := httptest.NewRecorder()
			sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
			want := http.StatusForbidden
			if name == "missing session" || name == "invalid session" {
				want = http.StatusUnauthorized
			}
			if response.Code != want || response.Header().Get("Cache-Control") != privateNoStoreDirective || len(lifecycle.revoked) != 0 {
				t.Fatalf("status=%d headers=%v revoked=%v body=%q", response.Code, response.Header(), lifecycle.revoked, response.Body.String())
			}
			if want == http.StatusUnauthorized && len(response.Result().Cookies()) != 2 {
				t.Fatalf("401 did not expire both cookies: %v", response.Result().Cookies())
			}
		})
	}
}

func TestLogoutDuplicateCSRFHeaderUsesSafeCategoricalBindingError(t *testing.T) {
	lifecycle := &sessionLifecycleStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	request.Header.Add("X-CSRF-Token", "first")
	request.Header.Add("X-CSRF-Token", "second")
	request.Header.Set("Origin", "https://findur.example")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	response := httptest.NewRecorder()
	sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || response.Header().Get("Cache-Control") != noStoreDirective || !strings.Contains(response.Body.String(), `"code":"invalid_request"`) || lifecycle.authorizeCalls != 0 {
		t.Fatalf("status=%d headers=%v calls=%d body=%q", response.Code, response.Header(), lifecycle.authorizeCalls, response.Body.String())
	}
}

func TestLogoutPropagatesSessionStorageFailures(t *testing.T) {
	lifecycle := &sessionLifecycleStub{authorizeErr: errors.New("database unavailable")}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	request.Header.Set("X-CSRF-Token", "csrf")
	request.Header.Set("Origin", "https://findur.example")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	response := httptest.NewRecorder()
	sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), `"code":"initialization_failed"`) {
		t.Fatalf("status=%d headers=%v body=%q", response.Code, response.Header(), response.Body.String())
	}
}

func TestLogoutRejectsMissingCSRFWithForbiddenBeforeRevocation(t *testing.T) {
	lifecycle := &sessionLifecycleStub{authorizeErr: auth.ErrForbidden}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	request.Header.Set("Origin", "https://findur.example")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	response := httptest.NewRecorder()
	sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || len(lifecycle.revoked) != 0 {
		t.Fatalf("status=%d revoked=%v body=%q", response.Code, lifecycle.revoked, response.Body.String())
	}
}

func TestLogoutRevokesOnlyPresentedSessionAndExpiresBothCookies(t *testing.T) {
	lifecycle := &sessionLifecycleStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "current-browser"})
	request.Header.Set("X-CSRF-Token", "session-bound-csrf")
	request.Header.Set("Referer", "https://findur.example/portfolio")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	response := httptest.NewRecorder()
	sessionHandler(lifecycle, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Header().Get("Cache-Control") != privateNoStoreDirective || len(lifecycle.revoked) != 1 || lifecycle.revoked[0] != "current-browser" {
		t.Fatalf("status=%d headers=%v revoked=%v", response.Code, response.Header(), lifecycle.revoked)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 || cookies[0].Name != sessionCookieName || cookies[1].Name != csrfCookieName || cookies[0].MaxAge != -1 || cookies[1].MaxAge != -1 || !cookies[0].HttpOnly || cookies[1].HttpOnly {
		t.Fatalf("cookies=%+v", cookies)
	}
}

func TestAuthorizationStatusIsNoStoreAndServerAuthoritative(t *testing.T) {
	for name, completer := range map[string]*statusCompleter{
		"available unauthenticated": {status: auth.AuthorizationStatus{AuthorizationAvailable: true}},
		"available authenticated":   {status: auth.AuthorizationStatus{AuthorizationAvailable: true, Authenticated: true}},
	} {
		t.Run(name, func(t *testing.T) {
			handler := callbackHandler(completer)
			request := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
			request.AddCookie(&http.Cookie{Name: "findur_session", Value: "opaque"})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" || completer.receivedSession != "opaque" {
				t.Fatalf("status=%d headers=%v session=%q", response.Code, response.Header(), completer.receivedSession)
			}
			wantAuthenticated := strings.Contains(name, "authenticated") && !strings.Contains(name, "unauthenticated")
			if strings.Contains(response.Body.String(), `"authenticated":true`) != wantAuthenticated || strings.Contains(response.Body.String(), "opaque") {
				t.Fatalf("body=%q", response.Body.String())
			}
		})
	}
	t.Run("gate closed", func(t *testing.T) {
		handler := callbackHandler(nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/status", nil))
		if response.Code != http.StatusOK || response.Body.String() != `{"authenticated":false,"authorizationAvailable":false}`+"\n" {
			t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
		}
	})
	t.Run("authenticated session remains available while initiation is closed", func(t *testing.T) {
		lifecycle := &sessionLifecycleStub{}
		handler := sessionHandler(lifecycle, "https://findur.example")
		request := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "active"})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != `{"authenticated":true,"authorizationAvailable":false}`+"\n" {
			t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
		}
	})
}

func TestAuthorizationCallbackRotatesSecureCookiesAndRedirectsCleanly(t *testing.T) {
	handler := callbackHandler(completeFunc(func(_ context.Context, input auth.CallbackInput) (auth.CallbackResult, error) {
		if input.State != "state" || input.Code != "code" || input.Binding != "binding" {
			t.Fatalf("input=%+v", input)
		}
		return auth.CallbackResult{Route: "/connect/result", Session: "session-secret", CSRF: "csrf-secret", Success: true}, nil
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/auth/snaptrade/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: "findur_oauth_attempt", Value: "binding"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/connect/result" || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response=%d %v", response.Code, response.Header())
	}
	cookies := response.Result().Cookies()
	wantSessionSeconds := 7 * 24 * 60 * 60
	if len(cookies) != 3 || cookies[0].MaxAge != -1 || cookies[1].MaxAge != wantSessionSeconds || cookies[2].MaxAge != wantSessionSeconds || !cookies[1].Secure || !cookies[1].HttpOnly || cookies[1].Domain != "" || cookies[2].HttpOnly || cookies[2].SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookies=%+v", cookies)
	}
}

func TestAuthorizationCallbackFailureIsCategorical(t *testing.T) {
	handler := callbackHandler(completeFunc(func(context.Context, auth.CallbackInput) (auth.CallbackResult, error) {
		return auth.CallbackResult{Route: "/connect/result"}, errors.New("private provider detail")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/snaptrade/callback?error=access_denied&error_description=private-provider-detail", nil))
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/connect/result" || strings.Contains(response.Body.String(), "provider") || len(response.Result().Cookies()) != 1 {
		t.Fatalf("status=%d headers=%v body=%q", response.Code, response.Header(), response.Body.String())
	}
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
