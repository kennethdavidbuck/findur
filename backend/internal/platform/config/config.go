// Package config owns the process configuration and bounded operational policy.
package config

import (
	"errors"
	"net"
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
	ShutdownDrain     = 25 * time.Second
)

// Config contains only values required to start the walking skeleton.
type Config struct {
	Address      string
	DatabaseURL  string
	MigrationURL string
}

// Load reads and validates configuration from environment variables.
func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = defaultPort
	}
	for _, character := range port {
		if character < '0' || character > '9' {
			return Config{}, errors.New("PORT must be a decimal integer from 1 through 65535")
		}
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, errors.New("PORT must be a decimal integer from 1 through 65535")
	}

	migrationURL := strings.TrimSpace(os.Getenv("MIGRATIONS_URL"))
	if migrationURL == "" {
		migrationURL = "file://db/migrations"
	}

	return Config{
		Address:      net.JoinHostPort("", strconv.Itoa(portNumber)),
		DatabaseURL:  databaseURL,
		MigrationURL: migrationURL,
	}, nil
}
