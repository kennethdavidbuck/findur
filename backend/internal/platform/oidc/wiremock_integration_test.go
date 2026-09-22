package oidc_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/oidc"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/provider"
)

const (
	credentialTestClientID     = "wiremock-refresh-client"
	credentialTestClientSecret = "wiremock-refresh-secret"
	credentialTestOldAccess    = "wiremock-old-access"
	credentialTestNewAccess    = "wiremock-new-access"
	credentialTestRefresh      = "wiremock-original-refresh"
	credentialTestRotated      = "wiremock-rotated-refresh"
)

// This test runs against the actual WireMock service in Compose. Repository
// lease/CAS behavior is independently exercised by PostgreSQL repository tests.
func TestWireMockCredentialRefresh(t *testing.T) {
	base := os.Getenv("WIREMOCK_URL")
	if base == "" {
		t.Skip("WIREMOCK_URL is required for the Compose integration scenario")
	}
	base = strings.TrimRight(base, "/")
	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("valid token reads without refresh", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.inventoryResponse(credentialTestOldAccess, http.StatusOK)
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(time.Hour), time.Second)
		if err := loadCredentialInventory(source, inventory, owner); err != nil {
			t.Fatal(err)
		}
		if repository.status() != "active" || repository.version() != 1 {
			t.Fatalf("authorization state=%s version=%d", repository.status(), repository.version())
		}
		wiremock.assertCounts(0, 1, 0)
	})

	t.Run("discovery outage retries without spending refresh token", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.add(map[string]any{
			"priority": 1,
			"request":  map[string]any{"method": "GET", "urlPath": "/.well-known/openid-configuration"},
			"response": map[string]any{"status": http.StatusServiceUnavailable},
		})
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(-time.Minute), time.Second)
		if err := loadCredentialInventory(source, inventory, owner); !errors.Is(err, auth.ErrRefreshUnavailable) {
			t.Fatalf("error=%v", err)
		}
		if repository.status() != "active" || repository.version() != 1 {
			t.Fatalf("authorization state=%s version=%d", repository.status(), repository.version())
		}
		wiremock.assertCounts(0, 0, 0)
	})

	t.Run("expired token refreshes before provider read", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.refreshResponse(http.StatusOK, 0, credentialTokenResponse())
		wiremock.inventoryResponse(credentialTestOldAccess, http.StatusUnauthorized)
		wiremock.inventoryResponse(credentialTestNewAccess, http.StatusOK)
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(-time.Minute), time.Second)
		if err := loadCredentialInventory(source, inventory, owner); err != nil {
			t.Fatal(err)
		}
		if repository.status() != "active" || repository.version() != 2 {
			t.Fatalf("authorization state=%s version=%d", repository.status(), repository.version())
		}
		wiremock.assertCounts(1, 0, 1)
	})

	t.Run("provider 401 refreshes and retries once", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.refreshResponse(http.StatusOK, 0, credentialTokenResponse())
		wiremock.inventoryResponse(credentialTestOldAccess, http.StatusUnauthorized)
		wiremock.inventoryResponse(credentialTestNewAccess, http.StatusOK)
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(time.Hour), time.Second)
		if err := loadCredentialInventory(source, inventory, owner); err != nil {
			t.Fatal(err)
		}
		if repository.status() != "active" {
			t.Fatalf("authorization state=%s", repository.status())
		}
		wiremock.assertCounts(1, 1, 1)
	})

	t.Run("second 401 requires reauthorization", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.refreshResponse(http.StatusOK, 0, credentialTokenResponse())
		wiremock.inventoryResponse(credentialTestOldAccess, http.StatusUnauthorized)
		wiremock.inventoryResponse(credentialTestNewAccess, http.StatusUnauthorized)
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(time.Hour), time.Second)
		if err := loadCredentialInventory(source, inventory, owner); !errors.Is(err, auth.ErrReauthorizationRequired) {
			t.Fatalf("error=%v", err)
		}
		if repository.status() != "reauthorization-required" {
			t.Fatalf("authorization state=%s", repository.status())
		}
		wiremock.assertCounts(1, 1, 1)
	})

	for _, scenario := range []struct {
		name   string
		status int
		delay  int
		body   string
	}{
		{name: "invalid refresh grant", status: http.StatusBadRequest, body: `{"error":"invalid_grant"}`},
		{name: "ambiguous token timeout", status: http.StatusOK, delay: 300, body: credentialTokenResponse()},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			wiremock := newCredentialWireMock(t, base, client)
			wiremock.refreshResponse(scenario.status, scenario.delay, scenario.body)
			timeout := time.Second
			if scenario.delay > 0 {
				timeout = 50 * time.Millisecond
			}
			source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(-time.Minute), timeout)
			if err := loadCredentialInventory(source, inventory, owner); !errors.Is(err, auth.ErrReauthorizationRequired) {
				t.Fatalf("error=%v", err)
			}
			if repository.status() != "reauthorization-required" {
				t.Fatalf("authorization state=%s", repository.status())
			}
			wiremock.assertCounts(1, 0, 0)
		})
	}

	t.Run("concurrent callers share one rotating refresh", func(t *testing.T) {
		wiremock := newCredentialWireMock(t, base, client)
		wiremock.refreshResponse(http.StatusOK, 200, credentialTokenResponse())
		wiremock.inventoryResponse(credentialTestNewAccess, http.StatusOK)
		source, repository, inventory, owner := credentialFixture(t, base, client, time.Now().Add(-time.Minute), 2*time.Second)
		const callers = 6
		var group sync.WaitGroup
		results := make(chan error, callers)
		for range callers {
			group.Add(1)
			go func() {
				defer group.Done()
				results <- loadCredentialInventory(source, inventory, owner)
			}()
		}
		group.Wait()
		close(results)
		for err := range results {
			if err != nil {
				t.Fatal(err)
			}
		}
		if repository.version() != 2 {
			t.Fatalf("credential version=%d", repository.version())
		}
		wiremock.assertCounts(1, 0, callers)
	})
}

func credentialTokenResponse() string {
	return `{"access_token":"` + credentialTestNewAccess + `","refresh_token":"` + credentialTestRotated + `","token_type":"Bearer","expires_in":3600}`
}

func credentialFixture(t *testing.T, base string, client *http.Client, expiry time.Time, timeout time.Duration) (*auth.Source, *memoryCredentialRepository, *provider.InventoryClient, uuid.UUID) {
	t.Helper()
	owner := uuid.New()
	cipher, err := auth.NewTokenCipher(auth.SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{7}, 32)}, 1, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	access, err := cipher.EncryptAccess(owner, credentialTestOldAccess)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := cipher.EncryptRefresh(owner, credentialTestRefresh)
	if err != nil {
		t.Fatal(err)
	}
	repository := &memoryCredentialRepository{credential: auth.Credential{
		Owner: owner, AccessEnvelope: access, RefreshEnvelope: refresh,
		EnvelopeVersion: 1, ExpiresAt: expiry, Status: "active", Generation: 1, Version: 1,
	}}
	discovery := oidc.NewDiscoveryClient(base, client)
	grant := oidc.NewCallbackClient(discovery, credentialTestClientID, credentialTestClientSecret, "")
	source, err := auth.NewCredentialSource(repository, cipher, grant, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Now, timeout, 2*time.Second, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := provider.NewInventoryClient(parsed, client, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	return source, repository, inventory, owner
}

func loadCredentialInventory(source *auth.Source, inventory *provider.InventoryClient, owner uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return source.Read(ctx, owner, func(callCtx context.Context, bearer string) error {
		_, err := inventory.Load(callCtx, bearer)
		return err
	})
}

type credentialWireMock struct {
	t      *testing.T
	base   string
	client *http.Client
	ids    []string
}

func newCredentialWireMock(t *testing.T, base string, client *http.Client) *credentialWireMock {
	t.Helper()
	w := &credentialWireMock{t: t, base: base, client: client}
	t.Cleanup(func() {
		for _, id := range w.ids {
			request, err := http.NewRequest(http.MethodDelete, base+"/__admin/mappings/"+id, nil)
			if err == nil {
				response, requestErr := client.Do(request)
				if requestErr == nil {
					_ = response.Body.Close()
				}
			}
		}
	})
	w.admin(http.MethodDelete, "/__admin/requests", nil, nil)
	w.add(map[string]any{
		"priority": 2,
		"request":  map[string]any{"method": "GET", "urlPath": "/.well-known/openid-configuration"},
		"response": map[string]any{"status": 200, "jsonBody": map[string]any{
			"issuer": base, "authorization_endpoint": base + "/__credential_test/authorize",
			"token_endpoint":           base + "/__credential_test/token",
			"revocation_endpoint":      base + "/__credential_test/revoke",
			"jwks_uri":                 base + "/__credential_test/jwks",
			"response_types_supported": []string{"code"}, "subject_types_supported": []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}},
	})
	return w
}

func (w *credentialWireMock) refreshResponse(status, delay int, body string) {
	w.t.Helper()
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(credentialTestClientID+":"+credentialTestClientSecret))
	w.add(map[string]any{
		"priority": 1,
		"request": map[string]any{
			"method": "POST", "urlPath": "/__credential_test/token",
			"headers":      map[string]any{"Authorization": map[string]any{"equalTo": basic}},
			"bodyPatterns": []any{map[string]any{"contains": "grant_type=refresh_token"}, map[string]any{"contains": "refresh_token=" + credentialTestRefresh}},
		},
		"response": map[string]any{
			"status": status, "fixedDelayMilliseconds": delay,
			"headers": map[string]string{"Content-Type": "application/json"}, "body": body,
		},
	})
}

func (w *credentialWireMock) inventoryResponse(bearer string, status int) {
	w.t.Helper()
	w.add(map[string]any{
		"priority": 1,
		"request": map[string]any{
			"method": "GET", "urlPath": "/authorizations",
			"headers": map[string]any{"Authorization": map[string]any{"equalTo": "Bearer " + bearer}},
		},
		"response": map[string]any{
			"status": status, "headers": map[string]string{"Content-Type": "application/json"}, "body": "[]",
		},
	})
}

func (w *credentialWireMock) add(mapping map[string]any) {
	w.t.Helper()
	var created struct {
		ID string `json:"id"`
	}
	w.admin(http.MethodPost, "/__admin/mappings", mapping, &created)
	if created.ID == "" {
		w.t.Fatal("WireMock did not return a mapping ID")
	}
	w.ids = append(w.ids, created.ID)
}

func (w *credentialWireMock) admin(method, path string, payload, target any) {
	w.t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			w.t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, w.base+path, body)
	if err != nil {
		w.t.Fatal(err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := w.client.Do(request)
	if err != nil {
		w.t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		w.t.Fatalf("WireMock admin %s %s returned %d", method, path, response.StatusCode)
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			w.t.Fatal(err)
		}
	}
}

func (w *credentialWireMock) assertCounts(refresh, oldReads, newReads int) {
	w.t.Helper()
	var journal struct {
		Requests []struct {
			Request struct {
				URL     string            `json:"url"`
				Headers map[string]string `json:"headers"`
			} `json:"request"`
		} `json:"requests"`
	}
	w.admin(http.MethodGet, "/__admin/requests", nil, &journal)
	var actualRefresh, actualOld, actualNew int
	for _, entry := range journal.Requests {
		switch entry.Request.URL {
		case "/__credential_test/token":
			actualRefresh++
		case "/authorizations":
			switch entry.Request.Headers["Authorization"] {
			case "Bearer " + credentialTestOldAccess:
				actualOld++
			case "Bearer " + credentialTestNewAccess:
				actualNew++
			}
		}
	}
	if actualRefresh != refresh || actualOld != oldReads || actualNew != newReads {
		w.t.Fatalf("WireMock requests: refresh=%d old_reads=%d new_reads=%d; want %d,%d,%d", actualRefresh, actualOld, actualNew, refresh, oldReads, newReads)
	}
}

type memoryCredentialRepository struct {
	mu         sync.Mutex
	credential auth.Credential
}

func (r *memoryCredentialRepository) ReadCredential(_ context.Context, _ uuid.UUID) (auth.Credential, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.credential, true, nil
}

func (r *memoryCredentialRepository) ClaimRefresh(_ context.Context, _ uuid.UUID, expectedVersion int64, now time.Time, lease time.Duration) (auth.Credential, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.credential.Version != expectedVersion || r.credential.Status != "active" || r.credential.LeaseID != nil {
		return r.credential, false, nil
	}
	id := uuid.New()
	expires := now.Add(lease)
	r.credential.LeaseID, r.credential.LeaseExpiresAt = &id, &expires
	return r.credential, true, nil
}

func (r *memoryCredentialRepository) InstallRefresh(_ context.Context, claim auth.Credential, access, refresh []byte, version int, expiry, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if claim.LeaseID == nil || r.credential.LeaseID == nil || *claim.LeaseID != *r.credential.LeaseID || r.credential.Version != claim.Version || r.credential.Status != "active" {
		return false, nil
	}
	r.credential.AccessEnvelope, r.credential.RefreshEnvelope = access, refresh
	r.credential.EnvelopeVersion, r.credential.ExpiresAt = version, expiry
	r.credential.Version++
	r.credential.LeaseID, r.credential.LeaseExpiresAt = nil, nil
	return true, nil
}

func (r *memoryCredentialRepository) ReleaseRefresh(_ context.Context, claim auth.Credential, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if claim.LeaseID == nil || r.credential.LeaseID == nil || *claim.LeaseID != *r.credential.LeaseID {
		return false, nil
	}
	r.credential.LeaseID, r.credential.LeaseExpiresAt = nil, nil
	return true, nil
}

func (r *memoryCredentialRepository) RequireReauthorization(_ context.Context, _ uuid.UUID, lease *uuid.UUID, version int64, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.credential.Version != version || lease != nil && (r.credential.LeaseID == nil || *lease != *r.credential.LeaseID) {
		return nil
	}
	r.credential.Status = "reauthorization-required"
	r.credential.Generation++
	r.credential.LeaseID, r.credential.LeaseExpiresAt = nil, nil
	return nil
}

func (r *memoryCredentialRepository) status() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.credential.Status
}

func (r *memoryCredentialRepository) version() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.credential.Version
}

var _ auth.CredentialRepository = (*memoryCredentialRepository)(nil)
