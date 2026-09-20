package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func discoveryHTTPClient(t *testing.T, body string, requests *atomic.Int32) *http.Client {
	t.Helper()
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.String() != "http://issuer.test/.well-known/openid-configuration" {
			t.Fatalf("discovery URL=%s", request.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}
}

func TestDiscoveryUsesValidatedProviderMetadataAndCachesProvider(t *testing.T) {
	var requests atomic.Int32
	body := `{"issuer":"http://issuer.test","authorization_endpoint":"http://issuer.test/authorize","token_endpoint":"http://issuer.test/token","jwks_uri":"http://issuer.test/keys","revocation_endpoint":"http://issuer.test/revoke","response_types_supported":["code"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}`
	client := NewDiscoveryClient("http://issuer.test", discoveryHTTPClient(t, body, &requests))
	for range 2 {
		metadata, err := client.Discover(context.Background())
		if err != nil || metadata.AuthorizationEndpoint != "http://issuer.test/authorize" {
			t.Fatalf("metadata=%+v err=%v", metadata, err)
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("discovery requests=%d, want cached provider", requests.Load())
	}
}

func TestDiscoveryRejectsIssuerMismatch(t *testing.T) {
	var requests atomic.Int32
	body := `{"issuer":"https://evil.example","authorization_endpoint":"https://evil.example/authorize","token_endpoint":"https://evil.example/token","jwks_uri":"https://evil.example/keys","response_types_supported":["code"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}`
	if _, err := NewDiscoveryClient("http://issuer.test", discoveryHTTPClient(t, body, &requests)).Discover(context.Background()); err == nil {
		t.Fatal("Discover() accepted mismatched issuer")
	}
}

func TestDiscoveryRejectsOversizedMetadataWithoutMutatingClient(t *testing.T) {
	var requests atomic.Int32
	shared := discoveryHTTPClient(t, strings.Repeat(" ", maxDiscoveryMetadataBytes+1), &requests)
	if _, err := NewDiscoveryClient("http://issuer.test", shared).Discover(context.Background()); err == nil {
		t.Fatal("Discover() accepted oversized metadata")
	}
	if _, unchanged := shared.Transport.(roundTripFunc); !unchanged {
		t.Fatal("Discover() mutated shared HTTP client")
	}
}

func TestDiscoveryRejectsHTTPSIssuerDowngrade(t *testing.T) {
	var requests atomic.Int32
	body := `{"issuer":"https://issuer.test","authorization_endpoint":"http://issuer.test/authorize","token_endpoint":"https://issuer.test/token","jwks_uri":"https://issuer.test/keys","response_types_supported":["code"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}`
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}
	if _, err := NewDiscoveryClient("https://issuer.test", client).Discover(context.Background()); err == nil {
		t.Fatal("Discover() accepted HTTP authorization endpoint for HTTPS issuer")
	}
}

func TestCanceledDiscoveryDoesNotWaitForAnotherCaller(t *testing.T) {
	started := make(chan struct{}, 1)
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	discovery := NewDiscoveryClient("http://issuer.test", client)
	firstContext, cancelFirst := context.WithCancel(context.Background())
	firstResult := make(chan error, 1)
	go func() {
		_, err := discovery.Discover(firstContext)
		firstResult <- err
	}()
	<-started

	secondContext, cancelSecond := context.WithCancel(context.Background())
	cancelSecond()
	secondResult := make(chan error, 1)
	go func() {
		_, err := discovery.Discover(secondContext)
		secondResult <- err
	}()
	select {
	case err := <-secondResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("second error=%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("canceled discovery waited for another caller")
	}
	cancelFirst()
	if err := <-firstResult; !errors.Is(err, context.Canceled) {
		t.Fatalf("first error=%v", err)
	}
}

func TestCallbackClientUsesHTTPBasicForConfidentialClientOperations(t *testing.T) {
	const clientID, clientSecret = "client-id", "client-secret"
	var tokenCalls, revokeCalls atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		response := func(status int, body string) (*http.Response, error) {
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
		}
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			return response(http.StatusOK, providerMetadata("http://issuer.test"))
		case "/token", "/revoke":
			id, secret, ok := r.BasicAuth()
			if !ok || id != clientID || secret != clientSecret {
				t.Errorf("missing confidential-client Basic authentication")
				return response(http.StatusUnauthorized, `{"error":"invalid_client"}`)
			}
			body, _ := io.ReadAll(r.Body)
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Error(err)
			}
			if form.Get("client_id") != "" || form.Get("client_secret") != "" || strings.Contains(r.URL.RawQuery, clientSecret) || strings.Contains(string(body), clientSecret) {
				t.Error("client credentials leaked outside Authorization header")
			}
			if r.URL.Path == "/token" {
				tokenCalls.Add(1)
				if form.Get("redirect_uri") != "http://127.0.0.1:8080/api/auth/snaptrade/callback" || form.Get("code_verifier") != "verifier" {
					t.Errorf("token form=%v", form)
				}
				return response(http.StatusOK, `{"access_token":"access","token_type":"Bearer","id_token":"id"}`)
			}
			revokeCalls.Add(1)
			return response(http.StatusOK, `{}`)
		default:
			return response(http.StatusNotFound, `{}`)
		}
	})}
	discovery := NewDiscoveryClient("http://issuer.test", httpClient)
	client := NewCallbackClient(discovery, clientID, clientSecret, "http://127.0.0.1:8080/api/auth/snaptrade/callback")
	if _, err := client.Exchange(context.Background(), "code", "http://127.0.0.1:8080/api/auth/snaptrade/callback", "verifier"); err != nil {
		t.Fatal(err)
	}
	if err := client.Revoke(context.Background(), "access"); err != nil {
		t.Fatal(err)
	}
	if tokenCalls.Load() != 1 || revokeCalls.Load() != 1 {
		t.Fatalf("token=%d revoke=%d", tokenCalls.Load(), revokeCalls.Load())
	}
}

func TestCallbackClientRejectsInvalidOIDCClaims(t *testing.T) {
	validKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	badKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	tests := map[string]func(map[string]any) *rsa.PrivateKey{
		"signature": func(map[string]any) *rsa.PrivateKey { return badKey },
		"issuer":    func(c map[string]any) *rsa.PrivateKey { c["iss"] = "https://wrong.example"; return validKey },
		"audience":  func(c map[string]any) *rsa.PrivateKey { c["aud"] = "wrong-client"; return validKey },
		"expiry":    func(c map[string]any) *rsa.PrivateKey { c["exp"] = now.Add(-time.Minute).Unix(); return validKey },
		"issued-at": func(c map[string]any) *rsa.PrivateKey { c["iat"] = now.Add(2 * time.Minute).Unix(); return validKey },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			client := verificationFixture(t, validKey)
			claims := map[string]any{"iss": "http://issuer.test", "aud": "client-id", "sub": "subject", "nonce": "nonce", "iat": now.Unix(), "exp": now.Add(5 * time.Minute).Unix()}
			raw := signIDToken(t, mutate(claims), claims)
			if _, err := client.Verify(context.Background(), raw); err == nil {
				t.Fatalf("Verify accepted invalid %s", name)
			}
		})
	}
}

func verificationFixture(t *testing.T, key *rsa.PrivateKey) *CallbackClient {
	t.Helper()
	jwk := jose.JSONWebKey{Key: &key.PublicKey, KeyID: "test-key", Algorithm: string(jose.RS256), Use: "sig"}
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		status, body := http.StatusOK, ""
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			body = providerMetadata("http://issuer.test")
		case "/keys":
			encoded, _ := json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}})
			body = string(encoded)
		default:
			status, body = http.StatusNotFound, `{}`
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
	return NewCallbackClient(NewDiscoveryClient("http://issuer.test", httpClient), "client-id", "client-secret", "http://127.0.0.1:8080/api/auth/snaptrade/callback")
}

func providerMetadata(issuer string) string {
	return fmt.Sprintf(`{"issuer":%q,"authorization_endpoint":%q,"token_endpoint":%q,"revocation_endpoint":%q,"jwks_uri":%q,"response_types_supported":["code"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}`, issuer, issuer+"/authorize", issuer+"/token", issuer+"/revoke", issuer+"/keys")
}

func signIDToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "test-key", Algorithm: string(jose.RS256), Use: "sig"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(claims)
	signed, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := signed.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
