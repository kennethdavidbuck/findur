// Package provider contains the narrow server-side provider transport.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client sends only server-held bearer authentication to allowlisted provider paths.
type Client struct {
	baseURL *url.URL
	bearer  string
	http    doer
}

func NewClient(baseURL *url.URL, bearer string, httpClient doer) *Client {
	return &Client{baseURL: baseURL, bearer: bearer, http: httpClient}
}

// Accounts calls the sole endpoint allowlisted for the Story 0.1 fixture gate.
func (c *Client) Accounts(ctx context.Context) (*http.Response, error) {
	target := c.baseURL.ResolveReference(&url.URL{Path: "/provider/accounts"})
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create provider request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.bearer)
	return c.http.Do(request)
}
