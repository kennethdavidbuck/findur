package oidcfixture

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

const (
	fixtureIssuer       = "http://issuer.example"
	fixtureClientID     = "client-id"
	fixtureClientSecret = "client-secret"
	fixtureCallbackURL  = "http://127.0.0.1:8080/api/auth/snaptrade/callback"
)

func TestHandlerServesJWKS(t *testing.T) {
	handler := fixtureHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, fixtureIssuer+"/jwks", nil))

	var keys jose.JSONWebKeySet
	if err := json.Unmarshal(response.Body.Bytes(), &keys); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || len(keys.Keys) != 1 || keys.Keys[0].KeyID != keyID {
		t.Fatalf("status=%d keys=%+v", response.Code, keys.Keys)
	}
}

func TestHandlerRedirectsValidBrowserAuthorization(t *testing.T) {
	handler := fixtureHandler(t)
	query := validAuthorizationQuery()
	request := httptest.NewRequest(http.MethodGet, fixtureIssuer+"/authorize?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusSeeOther || response.Header().Get("Cache-Control") != noStore {
		t.Fatalf("status=%d cache-control=%q", response.Code, response.Header().Get("Cache-Control"))
	}
	if location.Scheme+"://"+location.Host+location.Path != fixtureCallbackURL {
		t.Fatalf("redirect location=%q", location.String())
	}
	if location.Query().Get(codeField) != "expected-nonce" || location.Query().Get(stateField) != "expected-state" || len(location.Query()) != 2 {
		t.Fatalf("redirect query=%q", location.RawQuery)
	}
}

func TestHandlerRejectsMalformedOrUnsafeAuthorizationRequests(t *testing.T) {
	handler := fixtureHandler(t)
	tests := map[string]func(url.Values){
		"missing client":       func(query url.Values) { query.Del(clientIDField) },
		"wrong client":         func(query url.Values) { query.Set(clientIDField, "wrong-client") },
		"wrong response type":  func(query url.Values) { query.Set(responseTypeField, "token") },
		"unsafe redirect":      func(query url.Values) { query.Set(redirectURIField, "https://evil.example/callback") },
		"wrong scope":          func(query url.Values) { query.Set(scopeField, "openid") },
		"missing state":        func(query url.Values) { query.Set(stateField, "") },
		"missing nonce":        func(query url.Values) { query.Del(nonceField) },
		"missing challenge":    func(query url.Values) { query.Set(codeChallengeField, "") },
		"wrong challenge mode": func(query url.Values) { query.Set(codeChallengeMethod, "plain") },
		"duplicate redirect":   func(query url.Values) { query.Add(redirectURIField, fixtureCallbackURL) },
		"unexpected parameter": func(query url.Values) { query.Set("prompt", "login") },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			query := validAuthorizationQuery()
			mutate(query)
			request := httptest.NewRequest(http.MethodGet, fixtureIssuer+"/authorize?"+query.Encode(), nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || response.Header().Get("Location") != "" || response.Header().Get("Cache-Control") != noStore {
				t.Fatalf("status=%d location=%q cache-control=%q", response.Code, response.Header().Get("Location"), response.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestHandlerIssuesSignedNonceTokenForAuthenticatedRequest(t *testing.T) {
	handler := fixtureHandler(t)
	form := url.Values{
		grantTypeField:    {authorizationCodeFlow},
		redirectURIField:  {fixtureCallbackURL},
		codeVerifierField: {"verifier"},
		codeField:         {"expected-nonce"},
	}
	request := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(fixtureClientID, fixtureClientSecret)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var token tokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &token); err != nil {
		t.Fatal(err)
	}
	signed, err := jose.ParseSigned(token.IDToken, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := signed.Verify(handler.jwks.Keys[0].Key)
	if err != nil {
		t.Fatal(err)
	}
	var claims identityClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || token.AccessToken != testAccessToken || token.RefreshToken != testRefreshToken || claims.Nonce != "expected-nonce" {
		t.Fatalf("status=%d token=%+v claims=%+v", response.Code, token, claims)
	}
}

func TestHandlerCanIssueDedicatedExpiredCredentialFixture(t *testing.T) {
	handler, err := NewWithTokenExpiry(fixtureIssuer, fixtureClientID, fixtureClientSecret, fixtureCallbackURL, 0)
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{grantTypeField: {authorizationCodeFlow}, redirectURIField: {fixtureCallbackURL}, codeVerifierField: {"verifier"}, codeField: {"nonce"}}
	request := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(fixtureClientID, fixtureClientSecret)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var token tokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &token); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || token.ExpiresIn != 0 {
		t.Fatalf("status=%d expires=%d", response.Code, token.ExpiresIn)
	}
}

func TestHandlerRotatesRefreshTokenForAuthenticatedRequest(t *testing.T) {
	handler := fixtureHandler(t)
	form := url.Values{grantTypeField: {refreshTokenGrant}, refreshTokenField: {testRefreshToken}}
	request := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(fixtureClientID, fixtureClientSecret)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var token tokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &token); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || token.AccessToken != testAccessToken+"-refreshed" || token.RefreshToken != testRefreshToken+"-rotated-1" || token.IDToken != "" || token.ExpiresIn <= 15*60 {
		t.Fatalf("status=%d token=%+v", response.Code, token)
	}
}

func TestHandlerRejectsInvalidTokenRequests(t *testing.T) {
	handler := fixtureHandler(t)
	valid := url.Values{
		grantTypeField:    {authorizationCodeFlow},
		redirectURIField:  {fixtureCallbackURL},
		codeVerifierField: {"verifier"},
		codeField:         {"nonce"},
	}
	tests := map[string]func(*http.Request){
		"missing basic authentication": func(*http.Request) {},
		"credentials in form": func(request *http.Request) {
			request.SetBasicAuth(fixtureClientID, fixtureClientSecret)
			request.Form = url.Values{clientIDField: {fixtureClientID}}
		},
		"wrong redirect": func(request *http.Request) {
			request.SetBasicAuth(fixtureClientID, fixtureClientSecret)
			request.Form = url.Values{redirectURIField: {"https://wrong.example/callback"}}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/token", strings.NewReader(valid.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mutate(request)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
		})
	}

	malformed := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/token", strings.NewReader("%"))
	malformed.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	malformed.SetBasicAuth(fixtureClientID, fixtureClientSecret)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, malformed)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("malformed form status=%d", response.Code)
	}
}

func TestHandlerRevocationAndUnknownRoutes(t *testing.T) {
	handler := fixtureHandler(t)

	revocation := httptest.NewRequest(http.MethodPost, fixtureIssuer+"/revoke", strings.NewReader(""))
	revocation.SetBasicAuth(fixtureClientID, fixtureClientSecret)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, revocation)
	if response.Code != http.StatusOK {
		t.Fatalf("authenticated revocation status=%d", response.Code)
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, fixtureIssuer+"/revoke", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated revocation status=%d", response.Code)
	}

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, fixtureIssuer+"/token", nil),
		httptest.NewRequest(http.MethodPost, fixtureIssuer+"/unknown", nil),
	} {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s %s status=%d", request.Method, request.URL.Path, response.Code)
		}
	}
}

func fixtureHandler(t *testing.T) *Handler {
	t.Helper()
	handler, err := New(fixtureIssuer, fixtureClientID, fixtureClientSecret, fixtureCallbackURL)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func validAuthorizationQuery() url.Values {
	return url.Values{
		clientIDField:       {fixtureClientID},
		responseTypeField:   {authorizationResponse},
		redirectURIField:    {fixtureCallbackURL},
		scopeField:          {authorizationScope},
		stateField:          {"expected-state"},
		nonceField:          {"expected-nonce"},
		codeChallengeField:  {"expected-challenge"},
		codeChallengeMethod: {pkceS256},
	}
}
