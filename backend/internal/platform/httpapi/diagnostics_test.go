package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fixtureProviderFunc func(context.Context) (*http.Response, error)

func (f fixtureProviderFunc) Accounts(ctx context.Context) (*http.Response, error) { return f(ctx) }

type proxyRoundTripFunc func(*http.Request) (*http.Response, error)

func (f proxyRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDiagnosticsAreAbsentUnlessExplicitlyConfigured(t *testing.T) {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	request := httptest.NewRequest(http.MethodGet, "/api/__fixture/provider/accounts", nil)
	response := httptest.NewRecorder()

	testHandler(readiness).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestProviderDiagnosticReturnsBackendAdapterResponse(t *testing.T) {
	provider := fixtureProviderFunc(func(context.Context) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"accounts":[]}`)),
		}, nil
	})
	diagnostics := &Diagnostics{proxy: http.NotFoundHandler(), provider: provider}
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", diagnostics)
	request := httptest.NewRequest(http.MethodGet, "/api/__fixture/provider/accounts", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != `{"accounts":[]}` {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

func TestReverseProxyPreservesContractSemantics(t *testing.T) {
	target, _ := url.Parse("http://wiremock:8080")
	transport := proxyRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("ok"))}
		switch request.URL.Path {
		case "/unsafe-json":
			body, _ := io.ReadAll(request.Body)
			if string(body) != `{"text":"<script>&\""}` || request.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("unsafe request = %q, %q", body, request.Header.Get("Content-Type"))
			}
			response.Header.Set("Cache-Control", "no-store")
		case "/callback":
			if request.URL.Query().Get("code") != "a+b/c=" || request.URL.Query().Get("state") != "synthetic-state" {
				t.Fatalf("callback query = %q", request.URL.RawQuery)
			}
			response.StatusCode = http.StatusNoContent
			response.Body = io.NopCloser(strings.NewReader(""))
		case "/cookies/set":
			response.Header.Add("Set-Cookie", "findur_session=synthetic; Path=/; HttpOnly; SameSite=Lax")
			response.Header.Add("Set-Cookie", "findur_csrf=synthetic; Path=/; SameSite=Strict")
		case "/cookies/replay":
			cookie := request.Header.Get("Cookie")
			if !strings.Contains(cookie, "findur_session=synthetic") || !strings.Contains(cookie, "findur_csrf=synthetic") {
				t.Fatalf("replayed cookie = %q", cookie)
			}
		case "/cached":
			response.Header.Set("Cache-Control", "private, max-age=60")
			response.Header.Set("ETag", "synthetic-etag")
		case "/failure":
			response.StatusCode = http.StatusTooManyRequests
			response.Header.Set("Retry-After", "7")
		default:
			t.Fatalf("unexpected proxy path %q", request.URL.Path)
		}
		return response, nil
	})
	diagnostics := newDiagnostics(target, fixtureProviderFunc(func(context.Context) (*http.Response, error) {
		return nil, errors.New("provider must not be called")
	}), transport)
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", diagnostics)

	unsafe := httptest.NewRequest(http.MethodPost, "/api/__fixture/proxy/unsafe-json", strings.NewReader(`{"text":"<script>&\""}`))
	unsafe.Header.Set("Content-Type", "application/json")
	unsafeResponse := httptest.NewRecorder()
	handler.ServeHTTP(unsafeResponse, unsafe)
	if unsafeResponse.Code != http.StatusOK || unsafeResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unsafe response = %d, %q", unsafeResponse.Code, unsafeResponse.Header())
	}

	callbackResponse := httptest.NewRecorder()
	handler.ServeHTTP(callbackResponse, httptest.NewRequest(http.MethodGet, "/api/__fixture/proxy/callback?code=a%2Bb%2Fc%3D&state=synthetic-state", nil))
	if callbackResponse.Code != http.StatusNoContent {
		t.Fatalf("callback status = %d", callbackResponse.Code)
	}

	setCookiesResponse := httptest.NewRecorder()
	handler.ServeHTTP(setCookiesResponse, httptest.NewRequest(http.MethodGet, "/api/__fixture/proxy/cookies/set", nil))
	cookies := setCookiesResponse.Result().Cookies()
	if len(cookies) != 2 || cookies[0].Domain != "" || cookies[1].Domain != "" {
		t.Fatalf("host-only cookies = %#v", cookies)
	}
	replayRequest := httptest.NewRequest(http.MethodGet, "/api/__fixture/proxy/cookies/replay", nil)
	for _, cookie := range cookies {
		replayRequest.AddCookie(cookie)
	}
	replayResponse := httptest.NewRecorder()
	handler.ServeHTTP(replayResponse, replayRequest)
	if replayResponse.Code != http.StatusOK {
		t.Fatalf("replay status = %d", replayResponse.Code)
	}

	cachedResponse := httptest.NewRecorder()
	handler.ServeHTTP(cachedResponse, httptest.NewRequest(http.MethodGet, "/api/__fixture/proxy/cached", nil))
	if cachedResponse.Header().Get("Cache-Control") != "private, max-age=60" || cachedResponse.Header().Get("ETag") != "synthetic-etag" {
		t.Fatalf("cache headers = %q", cachedResponse.Header())
	}

	failureResponse := httptest.NewRecorder()
	handler.ServeHTTP(failureResponse, httptest.NewRequest(http.MethodGet, "/api/__fixture/proxy/failure", nil))
	if failureResponse.Code != http.StatusTooManyRequests || failureResponse.Header().Get("Retry-After") != "7" {
		t.Fatalf("error response = %d, %q", failureResponse.Code, failureResponse.Header())
	}
}
