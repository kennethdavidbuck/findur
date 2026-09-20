package config

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://findur:secret@localhost/findur?sslmode=disable")
	t.Setenv("PORT", "8080")
	t.Setenv("MIGRATIONS_URL", "file://testdata/migrations")
	t.Setenv("APP_ENV", "production")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Address != ":8080" {
		t.Fatalf("Address = %q, want :8080", got.Address)
	}
	if got.MigrationURL != "file://testdata/migrations" {
		t.Fatalf("MigrationURL = %q", got.MigrationURL)
	}
	if !got.Production {
		t.Fatal("Production = false, want true")
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing DATABASE_URL error")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"0", "65536", "http", "+80", "-1", "80.0"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost/findur")
			t.Setenv("PORT", port)

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil for PORT %q, want invalid PORT error", port)
			}
		})
	}
}

func TestLoadHealthcheckUsesValidatedPort(t *testing.T) {
	t.Setenv("PORT", "18080")

	got, err := LoadHealthcheck()
	if err != nil {
		t.Fatalf("LoadHealthcheck() error = %v", err)
	}
	if got.URL != "http://127.0.0.1:18080/api/healthz" {
		t.Fatalf("URL = %q", got.URL)
	}
}

func TestLoadHealthcheckRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	if _, err := LoadHealthcheck(); err == nil {
		t.Fatal("LoadHealthcheck() error = nil, want invalid PORT error")
	}
}

func TestOperationalTimeoutsStayWithinDrainBudget(t *testing.T) {
	if ReadinessTimeout >= ShutdownDrain {
		t.Fatalf("readiness timeout %s must be below drain budget %s", ReadinessTimeout, ShutdownDrain)
	}
	for name, timeout := range map[string]time.Duration{
		"read header":   ReadHeaderTimeout,
		"read":          ReadTimeout,
		"write":         WriteTimeout,
		"provider":      ProviderTimeout,
		"authorization": AuthorizationTimeout,
	} {
		if timeout <= 0 || timeout >= ShutdownDrain {
			t.Fatalf("%s timeout %s must be positive and below drain budget %s", name, timeout, ShutdownDrain)
		}
	}
	if ProviderTimeout >= AuthorizationTimeout || AuthorizationTimeout >= WriteTimeout {
		t.Fatalf("timeouts must nest provider %s < authorization %s < write %s", ProviderTimeout, AuthorizationTimeout, WriteTimeout)
	}
}

func TestLoadValidatesSyntheticFixtureConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/findur")
	t.Setenv("FIXTURE_BASE_URL", "http://wiremock:8080")
	t.Setenv("FIXTURE_PROVIDER_BEARER_TOKEN", "synthetic-token")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.FixtureBaseURL.String() != "http://wiremock:8080" || got.FixtureProviderToken != "synthetic-token" {
		t.Fatalf("fixture config = %v, %q", got.FixtureBaseURL, got.FixtureProviderToken)
	}
}

func TestLoadRejectsPartialOrUnsafeFixtureConfiguration(t *testing.T) {
	for _, test := range []struct{ baseURL, token string }{
		{baseURL: "http://wiremock:8080"},
		{token: "synthetic-token"},
		{baseURL: "http://user:password@wiremock:8080", token: "synthetic-token"},
		{baseURL: "http://wiremock:8080/provider", token: "synthetic-token"},
	} {
		t.Run(test.baseURL+test.token, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost/findur")
			t.Setenv("FIXTURE_BASE_URL", test.baseURL)
			t.Setenv("FIXTURE_PROVIDER_BEARER_TOKEN", test.token)
			if _, err := Load(); err == nil {
				t.Fatal("Load() error = nil, want invalid fixture configuration")
			}
		})
	}
}

func TestAuthorizationGateIsIndependentAndClosedInProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/findur")
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_INITIATION_ENABLED", "true")
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Authorization.Enabled {
		t.Fatal("production authorization gate opened before callback story")
	}
}

func TestEnabledAuthorizationRequiresSafeExplicitConfiguration(t *testing.T) {
	key := base64.RawStdEncoding.EncodeToString(make([]byte, 32))
	encryptionKey := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	for name, callback := range map[string]string{"https": "https://findur.example/api/auth/snaptrade/callback", "loopback": "http://127.0.0.1:8080/api/auth/snaptrade/callback"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost/findur")
			t.Setenv("APP_ENV", "integration")
			t.Setenv("AUTH_INITIATION_ENABLED", "true")
			t.Setenv("SNAPTRADE_OAUTH_CLIENT_ID", "synthetic-client")
			t.Setenv("SNAPTRADE_OAUTH_CALLBACK_URL", callback)
			t.Setenv("SNAPTRADE_OIDC_ISSUER", "http://wiremock:8080")
			t.Setenv("OAUTH_HASH_KEY", key)
			t.Setenv("OAUTH_ENCRYPTION_KEY", encryptionKey)
			got, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if !got.Authorization.Enabled {
				t.Fatal("gate disabled")
			}
		})
	}
}

func TestEnabledAuthorizationRejectsUnexpectedCallbackPath(t *testing.T) {
	setValidAuthorizationEnvironment(t)
	t.Setenv("SNAPTRADE_OAUTH_CALLBACK_URL", "https://findur.example/oauth/callback")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a callback path that disagrees with the handler cookie")
	}
}

func TestHTTPAuthorizationIssuerPolicy(t *testing.T) {
	t.Run("remote rejected outside integration", func(t *testing.T) {
		setValidAuthorizationEnvironment(t)
		t.Setenv("SNAPTRADE_OIDC_ISSUER", "http://wiremock:8080")
		if _, err := Load(); err == nil {
			t.Fatal("Load() accepted remote HTTP issuer")
		}
	})
	t.Run("remote allowed in integration", func(t *testing.T) {
		setValidAuthorizationEnvironment(t)
		t.Setenv("APP_ENV", "integration")
		t.Setenv("SNAPTRADE_OIDC_ISSUER", "http://wiremock:8080")
		if _, err := Load(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("loopback allowed locally", func(t *testing.T) {
		setValidAuthorizationEnvironment(t)
		t.Setenv("SNAPTRADE_OIDC_ISSUER", "http://127.0.0.1:8080")
		if _, err := Load(); err != nil {
			t.Fatal(err)
		}
	})
}

func setValidAuthorizationEnvironment(t *testing.T) {
	t.Helper()
	key := base64.RawStdEncoding.EncodeToString(make([]byte, 32))
	encryptionKey := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	t.Setenv("DATABASE_URL", "postgres://localhost/findur")
	t.Setenv("AUTH_INITIATION_ENABLED", "true")
	t.Setenv("SNAPTRADE_OAUTH_CLIENT_ID", "synthetic-client")
	t.Setenv("SNAPTRADE_OAUTH_CALLBACK_URL", "https://findur.example/api/auth/snaptrade/callback")
	t.Setenv("SNAPTRADE_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("OAUTH_HASH_KEY", key)
	t.Setenv("OAUTH_ENCRYPTION_KEY", encryptionKey)
}

func TestEnabledAuthorizationRejectsNonLoopbackHTTPCallback(t *testing.T) {
	key := base64.RawStdEncoding.EncodeToString(make([]byte, 32))
	encryptionKey := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	t.Setenv("DATABASE_URL", "postgres://localhost/findur")
	t.Setenv("AUTH_INITIATION_ENABLED", "true")
	t.Setenv("SNAPTRADE_OAUTH_CLIENT_ID", "synthetic-client")
	t.Setenv("SNAPTRADE_OAUTH_CALLBACK_URL", "http://backend:10000/api/auth/snaptrade/callback")
	t.Setenv("SNAPTRADE_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("OAUTH_HASH_KEY", key)
	t.Setenv("OAUTH_ENCRYPTION_KEY", encryptionKey)
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted non-loopback HTTP callback")
	}
}

func TestEnabledAuthorizationRejectsReusedCryptoKey(t *testing.T) {
	key := base64.RawStdEncoding.EncodeToString(make([]byte, 32))
	t.Setenv("DATABASE_URL", "postgres://localhost/findur")
	t.Setenv("AUTH_INITIATION_ENABLED", "true")
	t.Setenv("SNAPTRADE_OAUTH_CLIENT_ID", "synthetic-client")
	t.Setenv("SNAPTRADE_OAUTH_CALLBACK_URL", "https://findur.example/api/auth/snaptrade/callback")
	t.Setenv("SNAPTRADE_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("OAUTH_HASH_KEY", key)
	t.Setenv("OAUTH_ENCRYPTION_KEY", key)
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted one key for hashing and encryption")
	}
}
