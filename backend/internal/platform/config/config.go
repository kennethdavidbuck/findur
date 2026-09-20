// Package config owns the process configuration and bounded operational policy.
package config

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/auth"
)

// Process timeouts bound HTTP, provider, readiness, and shutdown work.
const (
	defaultPort = "10000"

	envAppEnvironment    = "APP_ENV"
	envAuthorizationGate = "AUTH_INITIATION_ENABLED"
	envOAuthClientID     = "SNAPTRADE_OAUTH_CLIENT_ID"
	envOAuthClientSecret = "SNAPTRADE_OAUTH_CLIENT_SECRET"
	envOAuthCallbackURL  = "SNAPTRADE_OAUTH_CALLBACK_URL"
	envOIDCIssuer        = "SNAPTRADE_OIDC_ISSUER"
	envOAuthHashKey      = "OAUTH_HASH_KEY"
	envOAuthVerifierKey  = "OAUTH_ENCRYPTION_KEY"
	envSessionHashKey    = "SESSION_HASH_KEY"
	envPublicOrigin      = "PUBLIC_ORIGIN"
	envOAuthTokenKeyV1   = "OAUTH_TOKEN_KEY_V1"
	envDatabaseURL       = "DATABASE_URL"
	envMigrationsURL     = "MIGRATIONS_URL"
	envFixtureBaseURL    = "FIXTURE_BASE_URL"
	envFixtureToken      = "FIXTURE_PROVIDER_BEARER_TOKEN"
	envPort              = "PORT"

	environmentProduction  = "production"
	environmentIntegration = "integration"
	schemeHTTP             = "http"
	schemeHTTPS            = "https"

	ReadHeaderTimeout    = 5 * time.Second
	ReadTimeout          = 10 * time.Second
	WriteTimeout         = 15 * time.Second
	IdleTimeout          = 60 * time.Second
	ReadinessTimeout     = 2 * time.Second
	ProviderTimeout      = 10 * time.Second
	AuthorizationTimeout = 12 * time.Second
	ShutdownDrain        = 25 * time.Second
)

// Config contains only values required to start the walking skeleton.
type Config struct {
	Address              string
	DatabaseURL          string
	MigrationURL         string
	Production           bool
	Integration          bool
	FixtureBaseURL       *url.URL
	FixtureProviderToken string
	Authorization        AuthorizationConfig
	Session              SessionConfig
}

// SessionConfig is independent of the OAuth initiation feature gate.
type SessionConfig struct {
	HashKey      []byte
	PublicOrigin string
}

// AuthorizationConfig is explicit product policy and validated OIDC configuration.
type AuthorizationConfig struct {
	Enabled         bool
	ClientID        string
	ClientSecret    string
	CallbackURL     string
	Issuer          string
	HashKey         []byte
	EncryptionKey   []byte
	TokenKeys       map[int][]byte
	CurrentTokenKey int
	AllowedReturns  []string
}

// HealthcheckConfig contains the validated local probe target used by container health checks.
type HealthcheckConfig struct {
	URL string
}

// LoadHealthcheck reads the healthcheck process environment exactly once at startup.
func LoadHealthcheck() (HealthcheckConfig, error) {
	port, err := loadPort()
	if err != nil {
		return HealthcheckConfig{}, err
	}
	return HealthcheckConfig{URL: "http://127.0.0.1:" + port + "/api/healthz"}, nil
}

// Load reads and validates configuration from environment variables.
func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv(envDatabaseURL))
	if databaseURL == "" {
		return Config{}, errors.New(envDatabaseURL + " is required")
	}

	port, err := loadPort()
	if err != nil {
		return Config{}, err
	}

	migrationURL := strings.TrimSpace(os.Getenv(envMigrationsURL))
	if migrationURL == "" {
		migrationURL = "file://db/migrations"
	}
	fixtureBaseURL, fixtureToken, err := loadFixtureConfig()
	if err != nil {
		return Config{}, err
	}

	appEnvironment := strings.ToLower(strings.TrimSpace(os.Getenv(envAppEnvironment)))
	production := appEnvironment == environmentProduction
	authorization, err := loadAuthorizationConfig(appEnvironment)
	if err != nil {
		return Config{}, err
	}
	session, err := loadSessionConfig(appEnvironment, authorization)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Address:              net.JoinHostPort("", port),
		DatabaseURL:          databaseURL,
		MigrationURL:         migrationURL,
		Production:           production,
		Integration:          appEnvironment == environmentIntegration,
		FixtureBaseURL:       fixtureBaseURL,
		FixtureProviderToken: fixtureToken,
		Authorization:        authorization,
		Session:              session,
	}, nil
}

func loadAuthorizationConfig(appEnvironment string) (AuthorizationConfig, error) {
	enabled, err := strconv.ParseBool(defaultString(strings.TrimSpace(os.Getenv(envAuthorizationGate)), "false"))
	if err != nil {
		return AuthorizationConfig{}, errors.New(envAuthorizationGate + " must be true or false")
	}
	result := AuthorizationConfig{Enabled: enabled, AllowedReturns: []string{auth.DefaultReturnRoute, auth.PortfolioReturnRoute}}
	if !enabled {
		return result, nil
	}
	if err := loadOAuthSettings(&result, appEnvironment); err != nil {
		return AuthorizationConfig{}, err
	}
	if err := loadAuthorizationKeys(&result); err != nil {
		return AuthorizationConfig{}, err
	}
	return result, nil
}

func loadOAuthSettings(result *AuthorizationConfig, appEnvironment string) error {
	result.ClientID = strings.TrimSpace(os.Getenv(envOAuthClientID))
	result.ClientSecret = strings.TrimSpace(os.Getenv(envOAuthClientSecret))
	result.CallbackURL = strings.TrimSpace(os.Getenv(envOAuthCallbackURL))
	result.Issuer = strings.TrimSpace(os.Getenv(envOIDCIssuer))
	if result.ClientID == "" || result.ClientSecret == "" || result.CallbackURL == "" || result.Issuer == "" {
		return errors.New("enabled authorization requires complete OAuth configuration")
	}
	issuer, err := validateHTTPURL(result.Issuer, appEnvironment == environmentIntegration)
	if err != nil || issuer.Path != "" {
		return errors.New("invalid OIDC issuer URL")
	}
	callback, err := validateHTTPURL(result.CallbackURL, false)
	if err != nil {
		return errors.New("invalid OAuth callback URL")
	}
	if callback.Scheme == schemeHTTP && !isLoopback(callback.Hostname()) {
		return errors.New("HTTP OAuth callback must use a loopback host")
	}
	if callback.Path != auth.SnapTradeCallbackPath {
		return errors.New("OAuth callback URL must use the configured callback path")
	}
	return nil
}

func loadAuthorizationKeys(result *AuthorizationConfig) error {
	var err error
	result.HashKey, err = decodeKey(envOAuthHashKey)
	if err != nil {
		return err
	}
	result.EncryptionKey, err = decodeKey(envOAuthVerifierKey)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare(result.HashKey, result.EncryptionKey) == 1 {
		return errors.New("OAuth hashing and encryption keys must be independent")
	}
	tokenKey, err := decodeKey(envOAuthTokenKeyV1)
	if err != nil {
		return err
	}
	for _, other := range [][]byte{result.HashKey, result.EncryptionKey} {
		if subtle.ConstantTimeCompare(tokenKey, other) == 1 {
			return errors.New("OAuth token encryption key must be independent")
		}
	}
	result.TokenKeys, result.CurrentTokenKey = map[int][]byte{1: tokenKey}, 1
	return nil
}

func loadSessionConfig(appEnvironment string, authorization AuthorizationConfig) (SessionConfig, error) {
	rawOrigin := strings.TrimSpace(os.Getenv(envPublicOrigin))
	if rawOrigin == "" && authorization.CallbackURL != "" {
		callback, _ := url.Parse(authorization.CallbackURL)
		rawOrigin = callback.Scheme + "://" + callback.Host
	}
	rawKey := strings.TrimSpace(os.Getenv(envSessionHashKey))
	if rawKey == "" && rawOrigin == "" && appEnvironment != environmentProduction {
		return SessionConfig{}, nil
	}
	if rawKey == "" || rawOrigin == "" {
		return SessionConfig{}, errors.New("session lifecycle requires " + envSessionHashKey + " and " + envPublicOrigin)
	}
	origin, err := validateHTTPURL(rawOrigin, appEnvironment == environmentIntegration)
	if err != nil || origin.Path != "" {
		return SessionConfig{}, errors.New("invalid PUBLIC_ORIGIN")
	}
	hashKey, err := decodeKey(envSessionHashKey)
	if err != nil {
		return SessionConfig{}, err
	}
	for _, other := range [][]byte{authorization.HashKey, authorization.EncryptionKey} {
		if len(other) > 0 && subtle.ConstantTimeCompare(hashKey, other) == 1 {
			return SessionConfig{}, errors.New("session hashing key must be independent")
		}
	}
	for _, other := range authorization.TokenKeys {
		if subtle.ConstantTimeCompare(hashKey, other) == 1 {
			return SessionConfig{}, errors.New("session hashing key must be independent")
		}
	}
	return SessionConfig{HashKey: hashKey, PublicOrigin: origin.Scheme + "://" + origin.Host}, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func decodeKey(name string) ([]byte, error) {
	key, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(os.Getenv(name)))
	if err != nil || len(key) != 32 {
		return nil, errors.New(name + " must be an unpadded base64 32-byte key")
	}
	return key, nil
}

func validateHTTPURL(raw string, allowNonLoopbackHTTP bool) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("invalid URL")
	}
	if parsed.Scheme != schemeHTTPS && (parsed.Scheme != schemeHTTP || (!allowNonLoopbackHTTP && !isLoopback(parsed.Hostname()))) {
		return nil, errors.New("URL must use HTTPS")
	}
	return parsed, nil
}

func isLoopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func loadFixtureConfig() (*url.URL, string, error) {
	rawURL := strings.TrimSpace(os.Getenv(envFixtureBaseURL))
	token := strings.TrimSpace(os.Getenv(envFixtureToken))
	if rawURL == "" && token == "" {
		return nil, "", nil
	}
	if rawURL == "" || token == "" {
		return nil, "", errors.New("fixture base URL and provider bearer token must be configured together")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != schemeHTTP && parsed.Scheme != schemeHTTPS) || parsed.Host == "" || parsed.User != nil {
		return nil, "", errors.New("FIXTURE_BASE_URL must be an HTTP origin without credentials")
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, "", errors.New("FIXTURE_BASE_URL must not contain a path, query, or fragment")
	}
	return parsed, token, nil
}

func loadPort() (string, error) {
	port := strings.TrimSpace(os.Getenv(envPort))
	if port == "" {
		port = defaultPort
	}
	for _, character := range port {
		if character < '0' || character > '9' {
			return "", errors.New("PORT must be a decimal integer from 1 through 65535")
		}
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", errors.New("PORT must be a decimal integer from 1 through 65535")
	}
	return strconv.Itoa(portNumber), nil
}
