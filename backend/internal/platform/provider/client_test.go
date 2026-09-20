package provider

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(request *http.Request) (*http.Response, error) { return f(request) }

func TestAccountsUsesAllowlistedPathAndBearerOnlyAuthentication(t *testing.T) {
	baseURL, _ := url.Parse("http://wiremock:8080")
	client := NewClient(baseURL, "synthetic-server-token", roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.String() != "http://wiremock:8080/provider/accounts" {
			t.Fatalf("request = %s %s", request.Method, request.URL)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer synthetic-server-token" {
			t.Fatalf("Authorization = %q", got)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("unexpected provider query = %q", request.URL.RawQuery)
		}
		for _, forbidden := range []string{
			"clientId", "consumerKey", "userId", "userSecret", "timestamp", "Signature",
			"SnapTrade-Client-Id", "SnapTrade-Consumer-Key",
		} {
			if got := request.Header.Get(forbidden); got != "" {
				t.Fatalf("forbidden header %s = %q", forbidden, got)
			}
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"accounts":[]}`))}, nil
	}))

	response, err := client.Accounts(context.Background())
	if err != nil {
		t.Fatalf("Accounts() error = %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}
