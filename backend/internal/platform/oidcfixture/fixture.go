// Package oidcfixture provides a synthetic, integration-only OIDC token service.
package oidcfixture

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
)

type Handler struct {
	issuer, clientID, clientSecret, callbackURL string
	signer                                      jose.Signer
	jwks                                        jose.JSONWebKeySet
}

func New(issuer, clientID, clientSecret, callbackURL string) (*Handler, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	jwk := jose.JSONWebKey{Key: &key.PublicKey, KeyID: "synthetic-rs256", Algorithm: string(jose.RS256), Use: "sig"}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: jwk.KeyID, Algorithm: jwk.Algorithm, Use: jwk.Use}}, nil)
	if err != nil {
		return nil, err
	}
	return &Handler{issuer: issuer, clientID: clientID, clientSecret: clientSecret, callbackURL: callbackURL, signer: signer, jwks: jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/jwks"):
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(h.jwks)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/token"):
		clientID, clientSecret, authenticated := r.BasicAuth()
		if err := r.ParseForm(); err != nil || !authenticated || clientID != h.clientID || clientSecret != h.clientSecret || r.Form.Get("client_id") != "" || r.Form.Get("client_secret") != "" || r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("redirect_uri") != h.callbackURL || r.Form.Get("code_verifier") == "" || r.Form.Get("code") == "" {
			http.Error(w, "invalid_request", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		claims := map[string]any{"iss": h.issuer, "aud": h.clientID, "sub": "synthetic-test-subject", "nonce": r.Form.Get("code"), "iat": now.Unix(), "exp": now.Add(5 * time.Minute).Unix()}
		payload, _ := json.Marshal(claims)
		signed, err := h.signer.Sign(payload)
		if err != nil {
			http.Error(w, "server_error", http.StatusInternalServerError)
			return
		}
		raw, err := signed.CompactSerialize()
		if err != nil {
			http.Error(w, "server_error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"token_type": "Bearer", "access_token": "synthetic-access-token", "refresh_token": "synthetic-refresh-token", "expires_in": 300, "id_token": raw})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/revoke"):
		clientID, clientSecret, authenticated := r.BasicAuth()
		if err := r.ParseForm(); err != nil || !authenticated || clientID != h.clientID || clientSecret != h.clientSecret || r.Form.Get("client_id") != "" || r.Form.Get("client_secret") != "" {
			http.Error(w, "invalid_client", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}
