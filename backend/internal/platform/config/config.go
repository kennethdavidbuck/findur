// Package config owns the process configuration and bounded operational policy.
package config

import (
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

	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 15 * time.Second
	IdleTimeout       = 60 * time.Second
	ReadinessTimeout  = 2 * time.Second
	ProviderTimeout   = 10 * time.Second
	ShutdownDrain     = 25 * time.Second
)

// Config contains only values required to start the walking skeleton.
type Config struct {
	Address              string
	DatabaseURL          string
	MigrationURL         string
	Production           bool
	FixtureBaseURL       *url.URL
	FixtureProviderToken string
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

	return Config{
		Address:              net.JoinHostPort("", port),
		DatabaseURL:          databaseURL,
		MigrationURL:         migrationURL,
		Production:           strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production"),
		FixtureBaseURL:       fixtureBaseURL,
		FixtureProviderToken: fixtureToken,
	}, nil
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
