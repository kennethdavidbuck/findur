// Package oidcfixture provides a synthetic, integration-only OIDC token service.
package oidcfixture

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
)

const (
	keyID                             = "synthetic-rs256"
	testSubject                       = "synthetic-test-subject"
	testAccessToken                   = "synthetic-access-token"
	testRefreshToken                  = "synthetic-refresh-token"
	authorizationCodeFlow             = "authorization_code"
	clientIDField                     = "client_id"
	clientSecretField                 = "client_secret"
	codeField                         = "code"
	codeChallengeField                = "code_challenge"
	codeChallengeMethod               = "code_challenge_method"
	codeVerifierField                 = "code_verifier"
	redirectURIField                  = "redirect_uri"
	grantTypeField                    = "grant_type"
	fixtureAccessTokenLifetimeSeconds = 3600
	nonceField                        = "nonce"
	responseTypeField                 = "response_type"
	scopeField                        = "scope"
	stateField                        = "state"
	authorizationScope                = "openid read"
	authorizationResponse             = "code"
	pkceS256                          = "S256"
	bearerTokenType                   = "Bearer"
	jsonContentType                   = "application/json"
	noStore                           = "no-store"
)

// Handler serves a deterministic OIDC provider for integration tests.
type Handler struct {
	issuer, clientID, clientSecret, callbackURL string
	signer                                      jose.Signer
	jwks                                        jose.JSONWebKeySet
}

// New creates an integration-only OIDC fixture handler.
func New(issuer, clientID, clientSecret, callbackURL string) (*Handler, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	jwk := jose.JSONWebKey{Key: &key.PublicKey, KeyID: keyID, Algorithm: string(jose.RS256), Use: "sig"}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: jwk.KeyID, Algorithm: jwk.Algorithm, Use: jwk.Use}}, nil)
	if err != nil {
		return nil, err
	}
	return &Handler{issuer: issuer, clientID: clientID, clientSecret: clientSecret, callbackURL: callbackURL, signer: signer, jwks: jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", noStore)
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/authorize"):
		h.serveAuthorization(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/jwks"):
		h.serveJWKS(w)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/token"):
		h.serveToken(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/revoke"):
		h.serveRevocation(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) serveAuthorization(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !h.validAuthorizationRequest(query) {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	callback, err := url.Parse(h.callbackURL)
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	redirectQuery := callback.Query()
	redirectQuery.Set(codeField, query.Get(nonceField))
	redirectQuery.Set(stateField, query.Get(stateField))
	callback.RawQuery = redirectQuery.Encode()
	http.Redirect(w, r, callback.String(), http.StatusSeeOther)
}

func (h *Handler) validAuthorizationRequest(query url.Values) bool {
	return len(query) == 8 &&
		exactQueryValue(query, clientIDField, h.clientID) &&
		exactQueryValue(query, responseTypeField, authorizationResponse) &&
		exactQueryValue(query, redirectURIField, h.callbackURL) &&
		exactQueryValue(query, scopeField, authorizationScope) &&
		nonEmptyQueryValue(query, stateField) &&
		nonEmptyQueryValue(query, nonceField) &&
		nonEmptyQueryValue(query, codeChallengeField) &&
		exactQueryValue(query, codeChallengeMethod, pkceS256)
}

func exactQueryValue(query url.Values, key, expected string) bool {
	values := query[key]
	return len(values) == 1 && values[0] == expected
}

func nonEmptyQueryValue(query url.Values, key string) bool {
	values := query[key]
	return len(values) == 1 && values[0] != ""
}

func (h *Handler) serveJWKS(w http.ResponseWriter) {
	writeJSON(w, h.jwks)
}

func (h *Handler) serveToken(w http.ResponseWriter, r *http.Request) {
	if !h.authenticatedForm(r) {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	if !h.validTokenRequest(r) {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	idToken, err := h.signIDToken(r.Form.Get(codeField), time.Now().UTC())
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, tokenResponse{
		TokenType:    bearerTokenType,
		AccessToken:  testAccessToken,
		RefreshToken: testRefreshToken,
		ExpiresIn:    fixtureAccessTokenLifetimeSeconds,
		IDToken:      idToken,
	})
}

func (h *Handler) serveRevocation(w http.ResponseWriter, r *http.Request) {
	if !h.authenticatedForm(r) {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) validTokenRequest(r *http.Request) bool {
	return h.authenticatedForm(r) &&
		r.Form.Get(grantTypeField) == authorizationCodeFlow &&
		r.Form.Get(redirectURIField) == h.callbackURL &&
		r.Form.Get(codeVerifierField) != "" &&
		r.Form.Get(codeField) != ""
}

func (h *Handler) authenticatedForm(r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		return false
	}
	clientID, clientSecret, ok := r.BasicAuth()
	return ok && clientID == h.clientID && clientSecret == h.clientSecret &&
		r.Form.Get(clientIDField) == "" && r.Form.Get(clientSecretField) == ""
}

func (h *Handler) signIDToken(nonce string, now time.Time) (string, error) {
	payload, err := json.Marshal(identityClaims{
		Issuer: h.issuer, Audience: h.clientID, Subject: testSubject, Nonce: nonce,
		IssuedAt: now.Unix(), ExpiresAt: now.Add(5 * time.Minute).Unix(),
	})
	if err != nil {
		return "", err
	}
	signed, err := h.signer.Sign(payload)
	if err != nil {
		return "", err
	}
	return signed.CompactSerialize()
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", jsonContentType)
	_ = json.NewEncoder(w).Encode(value)
}

type identityClaims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Subject   string `json:"sub"`
	Nonce     string `json:"nonce"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type tokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}
