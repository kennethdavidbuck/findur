package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
	"github.com/kennethdavidbuck/findur/backend/internal/profile"
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

type sessionRepositoryStub struct {
	session auth.Session
	err     error
}

func (s sessionRepositoryStub) FindActive(context.Context, []byte, time.Time) (auth.Session, error) {
	return s.session, s.err
}

func (s sessionRepositoryStub) AuthenticateAndTouch(context.Context, []byte, time.Time, time.Time) (auth.Session, error) {
	return s.session, s.err
}

func (s sessionRepositoryStub) RevokeCurrent(context.Context, []byte, time.Time) (bool, error) {
	return s.err == nil, s.err
}

func (s sessionRepositoryStub) CleanupSessions(context.Context, time.Time) (int64, error) {
	return 0, s.err
}

type inventoryLifecycleStub struct {
	snapshot   portfolio.Snapshot
	getCalls   int
	retryCalls int
}

type inclusionLifecycleStub struct {
	snapshot     portfolio.InclusionSnapshot
	err          error
	getCalls     int
	confirmCalls int
	expected     int64
	key          string
	accountIDs   []string
}
type showcaseLifecycleStub struct {
	snapshot portfolio.Showcase
	calls    int
}

func (s *showcaseLifecycleStub) Get(context.Context, auth.Actor) (portfolio.Showcase, error) {
	s.calls++
	return s.snapshot, nil
}

type profileLifecycleStub struct {
	snapshot profile.Snapshot
	saved    profile.Input
	saves    int
	err      error
}

func (s *profileLifecycleStub) Get(context.Context, auth.Actor) (profile.Snapshot, error) {
	return s.snapshot, s.err
}
func (s *profileLifecycleStub) Save(_ context.Context, _ auth.Actor, input profile.Input) (profile.Profile, error) {
	s.saved, s.saves = input, s.saves+1
	return profile.Profile{DisplayName: input.DisplayName, AdultAttestedAt: time.Now(), LocationKey: input.LocationKey, RelationshipIntent: input.RelationshipIntent, Biography: input.Biography, AvatarKey: input.AvatarKey, Locale: input.Locale, Theme: input.Theme, Version: input.ExpectedVersion + 1}, s.err
}

func (s *inclusionLifecycleStub) Get(context.Context, auth.Actor) (portfolio.InclusionSnapshot, error) {
	s.getCalls++
	return s.snapshot, s.err
}

func (s *inclusionLifecycleStub) Confirm(_ context.Context, _ auth.Actor, expected int64, key string, accountIDs []string) (portfolio.InclusionSnapshot, error) {
	s.confirmCalls++
	s.expected, s.key, s.accountIDs = expected, key, accountIDs
	return s.snapshot, s.err
}

func (s *inventoryLifecycleStub) Get(context.Context, auth.Actor) (portfolio.Snapshot, error) {
	s.getCalls++
	return s.snapshot, nil
}

func (s *inventoryLifecycleStub) Retry(context.Context, auth.Actor) (portfolio.Snapshot, error) {
	s.retryCalls++
	return s.snapshot, nil
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

func inventoryHandler(sessions sessionLifecycle, inventory inventoryLifecycle, publicOrigin string) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithInventory(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, nil, sessions, inventory, false, publicOrigin)
}

func portfolioHandler(sessions sessionLifecycle, inventory inventoryLifecycle, inclusion inclusionLifecycle, publicOrigin string) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithPortfolio(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, nil, sessions, inventory, inclusion, false, publicOrigin)
}
func showcaseHandler(sessions sessionLifecycle, showcase showcaseLifecycle) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithShowcase(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, nil, sessions, nil, nil, showcase, false, "https://findur.example")
}

func TestShowcaseGETRequiresSessionAndSerializesOnlySafeFields(t *testing.T) {
	showcase := &showcaseLifecycleStub{snapshot: portfolio.Showcase{Accounts: []portfolio.ShowcaseAccount{{Label: "Account (•••• 1234)", Brokerage: "Broker", SyncMode: portfolio.SyncModeRealtime, Balances: portfolio.ShowcaseDataset{Context: portfolio.DatasetContext{Source: "SnapTrade", Coverage: "included account", Currency: "CAD", Freshness: portfolio.FreshnessCurrent}, Balances: []portfolio.ShowcaseBalance{{Currency: "CAD", Cash: stringPtr("1.2300")}}}}}}}
	request := httptest.NewRequest(http.MethodGet, "/api/portfolio/showcase", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "owner-session"})
	response := httptest.NewRecorder()
	showcaseHandler(&sessionLifecycleStub{}, showcase).ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != privateNoStoreDirective || showcase.calls != 1 {
		t.Fatalf("status=%d cache=%q calls=%d", response.Code, response.Header().Get("Cache-Control"), showcase.calls)
	}
	body := response.Body.String()
	if strings.Contains(body, "owner-session") || !strings.Contains(body, "1.2300") || !strings.Contains(body, "•••• 1234") {
		t.Fatalf("unsafe or missing body %q", body)
	}
	unauthenticated := httptest.NewRecorder()
	showcaseHandler(&sessionLifecycleStub{authorizeErr: auth.ErrUnauthenticated}, showcase).ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/portfolio/showcase", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", unauthenticated.Code)
	}
}
func stringPtr(value string) *string { return &value }

func profileHandler(sessions sessionLifecycle, profiles profileLifecycle, publicOrigin string) http.Handler {
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	return NewHandlerWithProfile(slog.New(slog.NewTextHandler(io.Discard, nil)), readiness, "development", nil, nil, nil, sessions, nil, nil, nil, profiles, false, publicOrigin)
}

func TestProfileGETIsPrivateAndOmitsCatalogueCoordinates(t *testing.T) {
	profiles := &profileLifecycleStub{snapshot: profile.Snapshot{Locations: []profile.Location{{Key: "halifax-ns", CityEN: "Halifax", CityFR: "Halifax", ProvinceEN: "Nova Scotia", ProvinceFR: "Nouvelle-Écosse"}}}}
	request := httptest.NewRequest(http.MethodGet, profilePath, nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	response := httptest.NewRecorder()
	profileHandler(&sessionLifecycleStub{}, profiles, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get(cacheControlHeader) != privateNoStoreDirective {
		t.Fatalf("status=%d cache=%q", response.Code, response.Header().Get(cacheControlHeader))
	}
	if strings.Contains(response.Body.String(), "latitude") || strings.Contains(response.Body.String(), "longitude") {
		t.Fatalf("coordinates leaked: %s", response.Body.String())
	}
}

func TestProfilePUTRejectsCrossOriginBeforePersistence(t *testing.T) {
	profiles := &profileLifecycleStub{}
	request := httptest.NewRequest(http.MethodPut, profilePath, strings.NewReader(`{"displayName":"Alex","adultAttested":true,"locationKey":"halifax-ns","relationshipIntent":"long-term","biography":"Hello","avatarKey":"aurora","locale":"en","theme":"system","expectedVersion":0}`))
	request.Header.Set(contentTypeHeader, jsonMediaType)
	request.Header.Set(csrfHeaderName, "csrf")
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set("Sec-Fetch-Site", secFetchSameOrigin)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
	response := httptest.NewRecorder()
	profileHandler(&sessionLifecycleStub{}, profiles, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || profiles.saves != 0 {
		t.Fatalf("status=%d saves=%d", response.Code, profiles.saves)
	}
}

func TestProfilePUTMapsCreateValidationConflictAndSessionFailures(t *testing.T) {
	validation := &profile.ValidationError{Fields: []string{"displayName", "biography"}}
	for _, test := range []struct {
		name         string
		profileErr   error
		authorizeErr error
		withSession  bool
		wantStatus   int
		wantSaves    int
		wantBody     string
	}{
		{name: "create", withSession: true, wantStatus: http.StatusOK, wantSaves: 1, wantBody: `"version":1`},
		{name: "validation", profileErr: validation, withSession: true, wantStatus: http.StatusBadRequest, wantSaves: 1, wantBody: `"fields":["displayName","biography"]`},
		{name: "conflict", profileErr: profile.ErrConflict, withSession: true, wantStatus: http.StatusConflict, wantSaves: 1, wantBody: `"code":"conflict"`},
		{name: "invalid csrf", authorizeErr: auth.ErrForbidden, withSession: true, wantStatus: http.StatusForbidden, wantBody: `"code":"forbidden"`},
		{name: "missing session", wantStatus: http.StatusUnauthorized, wantBody: `"code":"unauthenticated"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			profiles := &profileLifecycleStub{err: test.profileErr}
			sessions := &sessionLifecycleStub{authorizeErr: test.authorizeErr}
			request := httptest.NewRequest(http.MethodPut, profilePath, strings.NewReader(`{"displayName":"Alex","adultAttested":true,"locationKey":"halifax-ns","relationshipIntent":"long-term","biography":"Hello","avatarKey":"aurora","locale":"en","theme":"system","expectedVersion":0}`))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			request.Header.Set(csrfHeaderName, "csrf")
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", secFetchSameOrigin)
			if test.withSession {
				request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			}
			response := httptest.NewRecorder()
			profileHandler(sessions, profiles, "https://findur.example").ServeHTTP(response, request)
			if response.Code != test.wantStatus || profiles.saves != test.wantSaves || response.Header().Get(cacheControlHeader) != privateNoStoreDirective || !strings.Contains(response.Body.String(), test.wantBody) {
				t.Fatalf("status=%d saves=%d cache=%q body=%q", response.Code, profiles.saves, response.Header().Get(cacheControlHeader), response.Body.String())
			}
			if test.name == "create" {
				want := profile.Input{DisplayName: "Alex", AdultAttested: true, LocationKey: "halifax-ns", RelationshipIntent: "long-term", Biography: "Hello", AvatarKey: "aurora", Locale: "en", Theme: "system", ExpectedVersion: 0}
				if profiles.saved != want {
					t.Fatalf("mapped input=%+v want=%+v", profiles.saved, want)
				}
				for _, field := range []string{`"displayName":"Alex"`, `"locationKey":"halifax-ns"`, `"relationshipIntent":"long-term"`, `"biography":"Hello"`, `"avatarKey":"aurora"`, `"locale":"en"`, `"theme":"system"`, `"version":1`} {
					if !strings.Contains(response.Body.String(), field) {
						t.Fatalf("response omitted %s: %s", field, response.Body.String())
					}
				}
			}
		})
	}
}

func TestProfilePUTMapsSchemaFailuresToFieldValidation(t *testing.T) {
	valid := `{"displayName":"Alex","adultAttested":true,"locationKey":"halifax-ns","relationshipIntent":"long-term","biography":"Hello","avatarKey":"aurora","locale":"en","theme":"system","expectedVersion":0}`
	tests := []struct {
		name, body, field string
	}{
		{name: "missing display name", body: strings.Replace(valid, `"displayName":"Alex",`, "", 1), field: "displayName"},
		{name: "adult attestation false", body: strings.Replace(valid, `"adultAttested":true`, `"adultAttested":false`, 1), field: "adultAttested"},
		{name: "invalid relationship intent", body: strings.Replace(valid, `"relationshipIntent":"long-term"`, `"relationshipIntent":"unsupported"`, 1), field: "relationshipIntent"},
		{name: "negative version", body: strings.Replace(valid, `"expectedVersion":0`, `"expectedVersion":-1`, 1), field: "expectedVersion"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profiles := &profileLifecycleStub{}
			request := httptest.NewRequest(http.MethodPut, profilePath, strings.NewReader(test.body))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			request.Header.Set(csrfHeaderName, "csrf")
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", secFetchSameOrigin)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			response := httptest.NewRecorder()
			profileHandler(&sessionLifecycleStub{}, profiles, "https://findur.example").ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || profiles.saves != 0 || response.Header().Get(cacheControlHeader) != privateNoStoreDirective || !strings.Contains(response.Body.String(), `"code":"invalid_profile"`) || !strings.Contains(response.Body.String(), `"`+test.field+`"`) {
				t.Fatalf("status=%d saves=%d cache=%q body=%q", response.Code, profiles.saves, response.Header().Get(cacheControlHeader), response.Body.String())
			}
		})
	}
}

func TestEveryProfileFailureUsesPrivateNoStore(t *testing.T) {
	tests := []struct {
		name     string
		profiles profileLifecycle
		request  *http.Request
		want     int
	}{
		{name: "unsupported content type", profiles: &profileLifecycleStub{}, request: httptest.NewRequest(http.MethodPut, profilePath, strings.NewReader("profile")), want: http.StatusUnsupportedMediaType},
		{name: "service unavailable", profiles: &profileLifecycleStub{err: errors.New("database unavailable")}, request: httptest.NewRequest(http.MethodGet, profilePath, nil), want: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid"})
			if test.want == http.StatusUnsupportedMediaType {
				test.request.Header.Set(contentTypeHeader, "text/plain")
				test.request.Header.Set(csrfHeaderName, "csrf")
			}
			response := httptest.NewRecorder()
			profileHandler(&sessionLifecycleStub{}, test.profiles, "https://findur.example").ServeHTTP(response, test.request)
			if response.Code != test.want || response.Header().Get(cacheControlHeader) != privateNoStoreDirective {
				t.Fatalf("status=%d cache=%q body=%q", response.Code, response.Header().Get(cacheControlHeader), response.Body.String())
			}
		})
	}
}

func TestInclusionGETIsOwnerPrivateAndStartsEmpty(t *testing.T) {
	inclusion := &inclusionLifecycleStub{snapshot: portfolio.InclusionSnapshot{Version: 0, Committed: []string{}}}
	request := httptest.NewRequest(http.MethodGet, "/api/portfolio/inclusion", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "opaque-session"})
	response := httptest.NewRecorder()
	portfolioHandler(&sessionLifecycleStub{}, &inventoryLifecycleStub{}, inclusion, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != privateNoStoreDirective || inclusion.getCalls != 1 || !strings.Contains(response.Body.String(), `"committed":[]`) {
		t.Fatalf("status=%d headers=%v calls=%d body=%q", response.Code, response.Header(), inclusion.getCalls, response.Body.String())
	}
}

func TestInclusionPOSTRequiresAllSessionDefensesBeforeMutation(t *testing.T) {
	for _, test := range []struct {
		name, origin, fetchSite, csrf string
		want                          int
	}{
		{name: "accepted", origin: "https://findur.example", fetchSite: "same-origin", csrf: "csrf", want: http.StatusOK},
		{name: "cross origin", origin: "https://evil.example", fetchSite: "same-origin", csrf: "csrf", want: http.StatusForbidden},
		{name: "cross site", origin: "https://findur.example", fetchSite: "cross-site", csrf: "csrf", want: http.StatusForbidden},
		{name: "missing csrf", origin: "https://findur.example", fetchSite: "same-origin", want: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			inclusion := &inclusionLifecycleStub{snapshot: portfolio.InclusionSnapshot{Version: 1, Committed: []string{"account"}}}
			request := httptest.NewRequest(http.MethodPost, "/api/portfolio/inclusion", strings.NewReader(`{"accountIds":["account"]}`))
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Sec-Fetch-Site", test.fetchSite)
			request.Header.Set("X-CSRF-Token", test.csrf)
			request.Header.Set("X-Inclusion-Version", "0")
			request.Header.Set("Idempotency-Key", "key-1")
			response := httptest.NewRecorder()
			portfolioHandler(&sessionLifecycleStub{}, &inventoryLifecycleStub{}, inclusion, "https://findur.example").ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
			if test.want == http.StatusOK && (inclusion.confirmCalls != 1 || inclusion.expected != 0 || inclusion.key != "key-1" || len(inclusion.accountIDs) != 1) {
				t.Fatalf("inclusion=%+v", inclusion)
			}
			if test.want != http.StatusOK && inclusion.confirmCalls != 0 {
				t.Fatalf("unsafe calls=%d", inclusion.confirmCalls)
			}
		})
	}
}

func TestInclusionPOSTReturnsSafeConflictWithoutMutationDetail(t *testing.T) {
	inclusion := &inclusionLifecycleStub{err: portfolio.ErrInvalidAccountSelection}
	request := httptest.NewRequest(http.MethodPost, "/api/portfolio/inclusion", strings.NewReader(`{"accountIds":["foreign-account"]}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://findur.example")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("X-CSRF-Token", "csrf")
	request.Header.Set("X-Inclusion-Version", "0")
	request.Header.Set("Idempotency-Key", "key-2")
	response := httptest.NewRecorder()
	portfolioHandler(&sessionLifecycleStub{}, &inventoryLifecycleStub{}, inclusion, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusConflict || strings.Contains(response.Body.String(), "foreign-account") || !strings.Contains(response.Body.String(), "invalid_selection") {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestInclusionPOSTEnforcesFiveAccountLimitBeforeService(t *testing.T) {
	for _, count := range []int{5, 6} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			ids := make([]string, count)
			for index := range ids {
				ids[index] = fmt.Sprintf(`"synthetic-account-%04d"`, index)
			}
			inclusion := &inclusionLifecycleStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/portfolio/inclusion", strings.NewReader(`{"accountIds":[`+strings.Join(ids, ",")+`]}`))
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
			request.Header.Set("X-CSRF-Token", "csrf")
			request.Header.Set("X-Inclusion-Version", "0")
			request.Header.Set("Idempotency-Key", "large-selection")
			response := httptest.NewRecorder()
			portfolioHandler(&sessionLifecycleStub{}, &inventoryLifecycleStub{}, inclusion, "https://findur.example").ServeHTTP(response, request)
			if count > portfolio.MaxIncludedAccounts {
				if response.Code != http.StatusBadRequest || inclusion.confirmCalls != 0 {
					t.Fatalf("oversized selection: status=%d calls=%d", response.Code, inclusion.confirmCalls)
				}
			} else if response.Code != http.StatusOK || inclusion.confirmCalls != 1 || len(inclusion.accountIDs) != count {
				t.Fatalf("bounded selection: status=%d calls=%d accounts=%d", response.Code, inclusion.confirmCalls, len(inclusion.accountIDs))
			}
		})
	}
}

func TestInclusionPOSTEnforcesItsByteLimit(t *testing.T) {
	for _, size := range []int{inclusionRequestLimit, inclusionRequestLimit + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			body := `{"accountIds":["synthetic-account"]}`
			body += strings.Repeat(" ", size-len(body))
			inclusion := &inclusionLifecycleStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/portfolio/inclusion", strings.NewReader(body))
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", "https://findur.example")
			request.Header.Set("Sec-Fetch-Site", "same-origin")
			request.Header.Set("X-CSRF-Token", "csrf")
			request.Header.Set("X-Inclusion-Version", "0")
			request.Header.Set("Idempotency-Key", "large-body")
			response := httptest.NewRecorder()
			portfolioHandler(&sessionLifecycleStub{}, &inventoryLifecycleStub{}, inclusion, "https://findur.example").ServeHTTP(response, request)
			if size > inclusionRequestLimit {
				if response.Code != http.StatusBadRequest || inclusion.confirmCalls != 0 {
					t.Fatalf("oversized body: status=%d calls=%d", response.Code, inclusion.confirmCalls)
				}
			} else if response.Code != http.StatusOK || inclusion.confirmCalls != 1 {
				t.Fatalf("bounded body: status=%d calls=%d", response.Code, inclusion.confirmCalls)
			}
		})
	}
}

func TestInventoryGETIsOwnerDerivedMinimizedAndPrivateNoStore(t *testing.T) {
	lastSuccessfulAt := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	inventory := &inventoryLifecycleStub{snapshot: portfolio.Snapshot{
		State: portfolio.StateReady, Generation: 2, UpdatedAt: time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC),
		Connections: []portfolio.Connection{{ID: "connection", BrokerageLabel: "Synthetic Broker", Status: "disabled", SyncMode: "unknown", Diagnostic: &portfolio.ResourceDiagnostic{Reason: portfolio.DiagnosticConnectionDisabled, RecommendedAction: portfolio.DiagnosticActionReconnect, LastSuccessfulAt: &lastSuccessfulAt}, Accounts: []portfolio.Account{{ID: "account", Category: "investment", Type: "Margin", MaskedLabel: "Retirement (•••• 8443)", Available: false, Eligible: false, SyncState: "complete"}}}},
	}}
	request := httptest.NewRequest(http.MethodGet, "/api/portfolio/inventory", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "opaque-session"})
	response := httptest.NewRecorder()
	inventoryHandler(&sessionLifecycleStub{}, inventory, "https://findur.example").ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != privateNoStoreDirective || inventory.getCalls != 1 {
		t.Fatalf("status=%d headers=%v calls=%d body=%q", response.Code, response.Header(), inventory.getCalls, response.Body.String())
	}
	for _, forbidden := range []string{"balance", "position", "activity", "access-token", "userId"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("response exposed %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(response.Body.String(), `"maskedLabel":"Retirement (•••• 8443)"`) {
		t.Fatalf("body=%q", response.Body.String())
	}
	for _, diagnosticField := range []string{`"reason":"connection_disabled"`, `"recommendedAction":"reconnect"`, `"lastSuccessfulAt":"2026-09-19T12:00:00Z"`} {
		if !strings.Contains(response.Body.String(), diagnosticField) {
			t.Fatalf("response omitted persisted connection diagnostic %s: %s", diagnosticField, response.Body.String())
		}
	}
}

func TestInventoryGETRequiresAnActiveSession(t *testing.T) {
	inventory := &inventoryLifecycleStub{}
	response := httptest.NewRecorder()
	inventoryHandler(&sessionLifecycleStub{}, inventory, "https://findur.example").ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/portfolio/inventory", nil))
	if response.Code != http.StatusUnauthorized || response.Header().Get("Cache-Control") != privateNoStoreDirective || inventory.getCalls != 0 {
		t.Fatalf("status=%d headers=%v calls=%d", response.Code, response.Header(), inventory.getCalls)
	}
}

func TestInventoryRetryRequiresSameOriginSessionBoundCSRF(t *testing.T) {
	for _, test := range []struct {
		name, origin, fetchSite string
		want                    int
	}{
		{name: "same origin", origin: "https://findur.example", fetchSite: "same-origin", want: http.StatusOK},
		{name: "cross origin", origin: "https://evil.example", fetchSite: "same-origin", want: http.StatusForbidden},
		{name: "cross site", origin: "https://findur.example", fetchSite: "cross-site", want: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			inventory := &inventoryLifecycleStub{snapshot: portfolio.Snapshot{State: portfolio.StateEmpty, Generation: 2, UpdatedAt: time.Now()}}
			lifecycle := &sessionLifecycleStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/portfolio/inventory/retry", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "opaque-session"})
			request.Header.Set("X-CSRF-Token", "csrf")
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Sec-Fetch-Site", test.fetchSite)
			response := httptest.NewRecorder()
			inventoryHandler(lifecycle, inventory, "https://findur.example").ServeHTTP(response, request)
			if response.Code != test.want || response.Header().Get("Cache-Control") != privateNoStoreDirective {
				t.Fatalf("status=%d headers=%v body=%q", response.Code, response.Header(), response.Body.String())
			}
			if test.want == http.StatusOK && (inventory.retryCalls != 1 || lifecycle.authorizeCalls != 1) {
				t.Fatalf("retry=%d authorization=%d", inventory.retryCalls, lifecycle.authorizeCalls)
			}
			if test.want != http.StatusOK && inventory.retryCalls != 0 {
				t.Fatalf("unsafe retry calls=%d", inventory.retryCalls)
			}
		})
	}
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
		return auth.CallbackResult{Route: "/onboarding/accounts", Session: "session-secret", CSRF: "csrf-secret", Success: true}, nil
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/auth/snaptrade/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: "findur_oauth_attempt", Value: "binding"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/onboarding/accounts" || response.Header().Get("Cache-Control") != "no-store" {
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
		return auth.CallbackResult{Route: "/onboarding/accounts"}, errors.New("private provider detail")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/snaptrade/callback?error=access_denied&error_description=private-provider-detail", nil))
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/onboarding/accounts" || strings.Contains(response.Body.String(), "provider") || len(response.Result().Cookies()) != 1 {
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

func TestBoundedSnapTradeAccountIDsExcludeMalformedValues(t *testing.T) {
	valid := "6f1ee24e-4f23-4fdd-8d1c-93b51f55580a"
	actual := boundedSnapTradeAccountIDs([]string{"sensitive-account-number", valid, "00000000-0000-0000-0000-000000000000"})
	if len(actual) != 1 || actual[0] != valid {
		t.Fatalf("account IDs = %#v, want only valid UUID", actual)
	}
}

func TestUnexpectedHandlerFailureLogsSafeCategoryOnce(t *testing.T) {
	var logs bytes.Buffer
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandlerWithProfile(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, &sessionLifecycleStub{}, nil, nil, nil, &profileLifecycleStub{err: errors.New("private database detail")}, false, "https://findur.example")
	request := httptest.NewRequest(http.MethodGet, profilePath, nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	output := logs.String()
	if response.Code != http.StatusServiceUnavailable || strings.Count(output, `"failure":"handler_failure"`) != 1 || !strings.Contains(output, `"request_id":"`) || strings.Contains(output, "private database detail") {
		t.Fatalf("unsafe or incomplete unexpected-failure logging: status=%d logs=%s", response.Code, output)
	}
}

func TestAuthenticatedRequestLogsServerDerivedFindurUserID(t *testing.T) {
	owner := uuid.New()
	sessions, err := auth.NewSessionService(auth.SessionConfig{
		HashKey: bytes.Repeat([]byte{1}, 32),
		Clock:   time.Now,
	}, sessionRepositoryStub{session: auth.Session{UserID: owner}})
	if err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	handler := NewHandlerWithProfile(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, sessions, nil, nil, nil, &profileLifecycleStub{err: errors.New("unavailable")}, false, "https://findur.example")
	request := httptest.NewRequest(http.MethodGet, profilePath, nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	output := logs.String()
	if response.Code != http.StatusServiceUnavailable || strings.Count(output, `"user_id":"`+owner.String()+`"`) != 2 || strings.Count(output, `"request_id":"`) != 2 {
		t.Fatalf("authenticated logs do not share server-derived context: status=%d logs=%s", response.Code, output)
	}
}

func TestAnonymousAndForbiddenRequestsOmitFindurUserID(t *testing.T) {
	owner := uuid.New()
	sessions, err := auth.NewSessionService(auth.SessionConfig{
		HashKey: bytes.Repeat([]byte{2}, 32),
		Clock:   time.Now,
	}, sessionRepositoryStub{session: auth.Session{UserID: owner}})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		profile bool
		session bool
		request *http.Request
		status  int
	}{
		{
			name:    "anonymous read",
			profile: true,
			request: httptest.NewRequest(http.MethodGet, profilePath, nil),
			status:  http.StatusUnauthorized,
		},
		{
			name:    "valid session rejected before authorization",
			session: true,
			request: httptest.NewRequest(http.MethodPost, logoutPath, nil),
			status:  http.StatusForbidden,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
			readiness.SetReady(true)
			var handler http.Handler
			if test.profile {
				handler = NewHandlerWithProfile(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, sessions, nil, nil, nil, &profileLifecycleStub{}, false, "https://findur.example")
			} else {
				handler = NewHandlerWithSessions(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, sessions, false, "https://findur.example")
			}
			response := httptest.NewRecorder()
			if test.session {
				test.request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
			}
			handler.ServeHTTP(response, test.request)
			if response.Code != test.status {
				t.Fatalf("status=%d, want %d", response.Code, test.status)
			}
			if strings.Contains(logs.String(), owner.String()) || strings.Contains(logs.String(), `"user_id"`) {
				t.Fatalf("unauthorized request logged identity: %s", logs.String())
			}
		})
	}
}

func TestInclusionRequestLogsOnlyBoundedValidSnapTradeAccountIDs(t *testing.T) {
	owner := uuid.New()
	hashKey := bytes.Repeat([]byte{3}, 32)
	sessions, err := auth.NewSessionService(auth.SessionConfig{HashKey: hashKey, Clock: time.Now}, sessionRepositoryStub{session: sessionForLogging(owner, hashKey)})
	if err != nil {
		t.Fatal(err)
	}
	valid := uuid.New().String()
	var logs bytes.Buffer
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), time.Second)
	readiness.SetReady(true)
	inclusion := &inclusionLifecycleStub{snapshot: portfolio.InclusionSnapshot{Version: 1, Committed: []string{}}}
	handler := NewHandlerWithPortfolio(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, sessions, nil, inclusion, false, "https://findur.example")
	request := inclusionLoggingRequest(`{"accountIds":["` + valid + `","private-not-a-uuid"]}`)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	output := logs.String()
	if response.Code != http.StatusOK || !strings.Contains(output, `"user_id":"`+owner.String()+`"`) || !strings.Contains(output, `"snaptrade_account_ids":["`+valid+`"]`) || strings.Contains(output, "private-not-a-uuid") {
		t.Fatalf("inclusion request log was unsafe or incomplete: status=%d logs=%s", response.Code, output)
	}
}

func TestInclusionRequestLogCapsSnapTradeAccountIDs(t *testing.T) {
	accountIDs := make([]string, maxLoggedSnapTradeAccountIDs+1)
	for index := range accountIDs {
		accountIDs[index] = uuid.New().String()
	}
	actual := boundedSnapTradeAccountIDs(accountIDs)
	if len(actual) != maxLoggedSnapTradeAccountIDs || actual[0] != accountIDs[0] || strings.Contains(strings.Join(actual, ","), accountIDs[maxLoggedSnapTradeAccountIDs]) {
		t.Fatalf("account ID logging cap not enforced: %#v", actual)
	}
}

func sessionForLogging(owner uuid.UUID, hashKey []byte) auth.Session {
	mac := hmac.New(sha256.New, hashKey)
	_, _ = mac.Write([]byte("csrf"))
	return auth.Session{UserID: owner, CSRFHash: mac.Sum(nil)}
}

func inclusionLoggingRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, portfolioInclusionPath, strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	request.Header.Set("Content-Type", jsonMediaType)
	request.Header.Set("Origin", "https://findur.example")
	request.Header.Set("Sec-Fetch-Site", secFetchSameOrigin)
	request.Header.Set(csrfHeaderName, "csrf")
	request.Header.Set("X-Inclusion-Version", "0")
	request.Header.Set("Idempotency-Key", "key")
	return request
}
