package oidc

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
