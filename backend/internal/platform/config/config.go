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
)

const (
	defaultPort = "10000"

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
	SessionHashKey  []byte
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
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	port, err := loadPort()
	if err != nil {
		return Config{}, err
	}

	migrationURL := strings.TrimSpace(os.Getenv("MIGRATIONS_URL"))
	if migrationURL == "" {
		migrationURL = "file://db/migrations"
	}
	fixtureBaseURL, fixtureToken, err := loadFixtureConfig()
	if err != nil {
		return Config{}, err
	}

	appEnvironment := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	production := appEnvironment == "production"
	authorization, err := loadAuthorizationConfig(appEnvironment)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Address:              net.JoinHostPort("", port),
		DatabaseURL:          databaseURL,
		MigrationURL:         migrationURL,
		Production:           production,
		Integration:          appEnvironment == "integration",
		FixtureBaseURL:       fixtureBaseURL,
		FixtureProviderToken: fixtureToken,
		Authorization:        authorization,
	}, nil
}

func loadAuthorizationConfig(appEnvironment string) (AuthorizationConfig, error) {
	enabled, err := strconv.ParseBool(defaultString(strings.TrimSpace(os.Getenv("AUTH_INITIATION_ENABLED")), "false"))
	if err != nil {
		return AuthorizationConfig{}, errors.New("AUTH_INITIATION_ENABLED must be true or false")
	}
	result := AuthorizationConfig{Enabled: enabled, AllowedReturns: []string{"/connect", "/portfolio"}}
	if !enabled {
		return result, nil
	}
	result.ClientID = strings.TrimSpace(os.Getenv("SNAPTRADE_OAUTH_CLIENT_ID"))
	result.ClientSecret = strings.TrimSpace(os.Getenv("SNAPTRADE_OAUTH_CLIENT_SECRET"))
	result.CallbackURL = strings.TrimSpace(os.Getenv("SNAPTRADE_OAUTH_CALLBACK_URL"))
	result.Issuer = strings.TrimSpace(os.Getenv("SNAPTRADE_OIDC_ISSUER"))
	if result.ClientID == "" || result.ClientSecret == "" || result.CallbackURL == "" || result.Issuer == "" {
		return AuthorizationConfig{}, errors.New("enabled authorization requires complete OAuth configuration")
	}
	issuer, err := validateHTTPURL(result.Issuer, appEnvironment == "integration")
	if err != nil || issuer.Path != "" {
		return AuthorizationConfig{}, errors.New("invalid OIDC issuer URL")
	}
	callback, err := validateHTTPURL(result.CallbackURL, false)
	if err != nil {
		return AuthorizationConfig{}, errors.New("invalid OAuth callback URL")
	}
	if callback.Scheme == "http" && !isLoopback(callback.Hostname()) {
		return AuthorizationConfig{}, errors.New("HTTP OAuth callback must use a loopback host")
	}
	if callback.Path != "/api/auth/snaptrade/callback" {
		return AuthorizationConfig{}, errors.New("OAuth callback URL must use the configured callback path")
	}
	result.HashKey, err = decodeKey("OAUTH_HASH_KEY")
	if err != nil {
		return AuthorizationConfig{}, err
	}
	result.EncryptionKey, err = decodeKey("OAUTH_ENCRYPTION_KEY")
	if err != nil {
		return AuthorizationConfig{}, err
	}
	if subtle.ConstantTimeCompare(result.HashKey, result.EncryptionKey) == 1 {
		return AuthorizationConfig{}, errors.New("OAuth hashing and encryption keys must be independent")
	}
	result.SessionHashKey, err = decodeKey("SESSION_HASH_KEY")
	if err != nil {
		return AuthorizationConfig{}, err
	}
	tokenKey, err := decodeKey("OAUTH_TOKEN_KEY_V1")
	if err != nil {
		return AuthorizationConfig{}, err
	}
	for _, other := range [][]byte{result.HashKey, result.EncryptionKey, result.SessionHashKey} {
		if subtle.ConstantTimeCompare(tokenKey, other) == 1 {
			return AuthorizationConfig{}, errors.New("OAuth token encryption key must be independent")
		}
	}
	if subtle.ConstantTimeCompare(result.SessionHashKey, result.HashKey) == 1 || subtle.ConstantTimeCompare(result.SessionHashKey, result.EncryptionKey) == 1 {
		return AuthorizationConfig{}, errors.New("session hashing key must be independent")
	}
	result.TokenKeys, result.CurrentTokenKey = map[int][]byte{1: tokenKey}, 1
	return result, nil
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
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (allowNonLoopbackHTTP || isLoopback(parsed.Hostname()))) {
		return nil, errors.New("URL must use HTTPS")
	}
	return parsed, nil
}

func isLoopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func loadFixtureConfig() (*url.URL, string, error) {
	rawURL := strings.TrimSpace(os.Getenv("FIXTURE_BASE_URL"))
	token := strings.TrimSpace(os.Getenv("FIXTURE_PROVIDER_BEARER_TOKEN"))
	if rawURL == "" && token == "" {
		return nil, "", nil
	}
	if rawURL == "" || token == "" {
		return nil, "", errors.New("fixture base URL and provider bearer token must be configured together")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
		return nil, "", errors.New("FIXTURE_BASE_URL must be an HTTP origin without credentials")
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, "", errors.New("FIXTURE_BASE_URL must not contain a path, query, or fragment")
	}
	return parsed, token, nil
}

func loadPort() (string, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
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
