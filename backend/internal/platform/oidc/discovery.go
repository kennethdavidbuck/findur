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
	"time"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"golang.org/x/oauth2"
)

const (
	maxDiscoveryMetadataBytes = 64 << 10
	discoveryPath             = "/.well-known/openid-configuration"
	accessTokenTypeHint       = "access_token"
	httpScheme                = "http"
	httpsScheme               = "https"
)

// DiscoveryClient uses go-oidc for standards-compliant discovery and validation.
type DiscoveryClient struct {
	issuer     string
	httpClient *http.Client
	cached     atomic.Pointer[coreoidc.Provider]
}

// CallbackClient performs the standards-sensitive token exchange, ID-token
// verification and optional compensation revocation using discovered metadata.
type CallbackClient struct {
	discovery                           *DiscoveryClient
	clientID, clientSecret, callbackURL string
}

// NewCallbackClient creates an OAuth/OIDC callback adapter from discovered metadata.
func NewCallbackClient(discovery *DiscoveryClient, clientID, clientSecret, callbackURL string) *CallbackClient {
	return &CallbackClient{discovery: discovery, clientID: clientID, clientSecret: clientSecret, callbackURL: callbackURL}
}

func (c *CallbackClient) provider(ctx context.Context) (*coreoidc.Provider, error) {
	if _, err := c.discovery.Discover(ctx); err != nil {
		return nil, err
	}
	p := c.discovery.cached.Load()
	if p == nil {
		return nil, errors.New("OIDC provider unavailable")
	}
	return p, nil
}

// Exchange trades an authorization code and PKCE verifier for provider tokens.
func (c *CallbackClient) Exchange(ctx context.Context, code, redirectURI, verifier string) (auth.TokenSet, error) {
	p, err := c.provider(ctx)
	if err != nil {
		return auth.TokenSet{}, err
	}
	ctx = coreoidc.ClientContext(ctx, c.discovery.httpClient)
	endpoint := p.Endpoint()
	endpoint.AuthStyle = oauth2.AuthStyleInHeader
	token, err := (&oauth2.Config{ClientID: c.clientID, ClientSecret: c.clientSecret, RedirectURL: redirectURI, Endpoint: endpoint}).Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return auth.TokenSet{}, err
	}
	result := auth.TokenSet{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, Expiry: token.Expiry}
	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		return result, errors.New("token response omitted ID token")
	}
	result.IDToken = rawID
	return result, nil
}

// Verify validates an ID token and returns only the identity claims Findur uses.
func (c *CallbackClient) Verify(ctx context.Context, raw string) (auth.Identity, error) {
	p, err := c.provider(ctx)
	if err != nil {
		return auth.Identity{}, err
	}
	ctx = coreoidc.ClientContext(ctx, c.discovery.httpClient)
	token, err := p.Verifier(&coreoidc.Config{ClientID: c.clientID, SupportedSigningAlgs: []string{coreoidc.RS256}}).Verify(ctx, raw)
	if err != nil {
		return auth.Identity{}, err
	}
	var claims struct {
		Nonce    string `json:"nonce"`
		IssuedAt int64  `json:"iat"`
	}
	if err := token.Claims(&claims); err != nil {
		return auth.Identity{}, err
	}
	now := time.Now().UTC()
	issued := time.Unix(claims.IssuedAt, 0)
	if claims.IssuedAt == 0 || issued.After(now.Add(time.Minute)) || issued.Before(now.Add(-24*time.Hour)) {
		return auth.Identity{}, errors.New("implausible issued-at")
	}
	return auth.Identity{Subject: token.Subject, Nonce: claims.Nonce}, nil
}

// Revoke performs best-effort access-token compensation after callback failure.
func (c *CallbackClient) Revoke(ctx context.Context, token string) error {
	p, err := c.provider(ctx)
	if err != nil {
		return err
	}
	var metadata struct {
		RevocationEndpoint string `json:"revocation_endpoint"`
	}
	if err := p.Claims(&metadata); err != nil {
		return err
	}
	if metadata.RevocationEndpoint == "" {
		return errors.New("OIDC revocation endpoint unavailable")
	}
	form := url.Values{"token": {token}, "token_type_hint": {accessTokenTypeHint}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metadata.RevocationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)
	response, err := c.discovery.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("token revocation failed")
	}
	return nil
}

// NewDiscoveryClient creates a validated, cacheable OIDC discovery adapter.
func NewDiscoveryClient(issuer string, client *http.Client) *DiscoveryClient {
	return &DiscoveryClient{issuer: issuer, httpClient: client}
}

// Discover returns the validated subset of provider metadata used by authorization.
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
		ctx = coreoidc.ClientContext(ctx, &boundedClient)
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
	if !strings.Contains(request.URL.Path, discoveryPath) {
		return response, nil
	}
	if response.ContentLength > maxDiscoveryMetadataBytes {
		_ = response.Body.Close()
		return nil, errors.New("OIDC discovery metadata exceeds size limit")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxDiscoveryMetadataBytes+1))
	_ = response.Body.Close()
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
	issuer, _ := url.Parse(c.issuer)
	if err := validateDiscoveredEndpoint(authorizationEndpoint, issuer.Scheme == httpsScheme); err != nil {
		return auth.Discovery{}, errors.New("OIDC authorization endpoint is unsafe")
	}
	var metadata struct {
		JWKSURI            string `json:"jwks_uri"`
		RevocationEndpoint string `json:"revocation_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return auth.Discovery{}, errors.New("OIDC metadata is invalid")
	}
	for _, endpoint := range []string{provider.Endpoint().TokenURL, metadata.JWKSURI, metadata.RevocationEndpoint} {
		if err := validateDiscoveredEndpoint(endpoint, issuer.Scheme == httpsScheme); err != nil {
			return auth.Discovery{}, errors.New("OIDC metadata contains an unsafe endpoint")
		}
	}
	return auth.Discovery{AuthorizationEndpoint: authorizationEndpoint}, nil
}

func validateDiscoveredEndpoint(raw string, requireHTTPS bool) error {
	endpoint, err := url.Parse(raw)
	if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return errors.New("invalid endpoint")
	}
	if requireHTTPS && endpoint.Scheme != httpsScheme || !requireHTTPS && endpoint.Scheme != httpScheme && endpoint.Scheme != httpsScheme {
		return errors.New("unsafe endpoint scheme")
	}
	return nil
}
