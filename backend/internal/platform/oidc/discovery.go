// Package oidc adapts the selected OIDC implementation to the auth domain.
package oidc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"golang.org/x/oauth2"
)

const maxDiscoveryMetadataBytes = 64 << 10

// DiscoveryClient uses go-oidc for standards-compliant discovery and validation.
type DiscoveryClient struct {
	issuer     string
	httpClient *http.Client
	cached     atomic.Pointer[coreoidc.Provider]
}

func NewDiscoveryClient(issuer string, client *http.Client) *DiscoveryClient {
	return &DiscoveryClient{issuer: issuer, httpClient: client}
}

func (c *DiscoveryClient) Discover(ctx context.Context) (auth.Discovery, error) {
	if err := ctx.Err(); err != nil {
		return auth.Discovery{}, err
	}
	provider := c.cached.Load()
	if provider == nil {
		boundedClient := *c.httpClient
		transport := c.httpClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		boundedClient.Transport = boundedTransport{next: transport}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &boundedClient)
		discovered, err := coreoidc.NewProvider(ctx, c.issuer)
		if err != nil {
			return auth.Discovery{}, err
		}
		provider = discovered
	}
	result, err := c.validate(provider)
	if err != nil {
		return auth.Discovery{}, err
	}
	c.cached.CompareAndSwap(nil, provider)
	return result, nil
}

type boundedTransport struct{ next http.RoundTripper }

func (t boundedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.next.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(request.URL.Path, "/.well-known/openid-configuration") {
		return response, nil
	}
	if response.ContentLength > maxDiscoveryMetadataBytes {
		response.Body.Close()
		return nil, errors.New("OIDC discovery metadata exceeds size limit")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxDiscoveryMetadataBytes+1))
	response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read OIDC discovery metadata: %w", err)
	}
	if len(body) > maxDiscoveryMetadataBytes {
		return nil, errors.New("OIDC discovery metadata exceeds size limit")
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return response, nil
}

func (c *DiscoveryClient) validate(provider *coreoidc.Provider) (auth.Discovery, error) {
	authorizationEndpoint := provider.Endpoint().AuthURL
	endpoint, err := url.Parse(authorizationEndpoint)
	if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return auth.Discovery{}, errors.New("OIDC authorization endpoint is invalid")
	}
	issuer, _ := url.Parse(c.issuer)
	if issuer.Scheme == "https" && endpoint.Scheme != "https" || issuer.Scheme == "http" && endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return auth.Discovery{}, errors.New("OIDC authorization endpoint has an unsafe scheme")
	}
	return auth.Discovery{AuthorizationEndpoint: authorizationEndpoint}, nil
}
