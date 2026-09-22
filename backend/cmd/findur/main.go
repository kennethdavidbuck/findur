// Package main composes and runs the Findur API process.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/buildinfo"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/config"

	"github.com/kennethdavidbuck/findur/backend/internal/platform/httpapi"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/lifecycle"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/migrations"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/oidc"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/oidcfixture"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/provider"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
	"github.com/kennethdavidbuck/findur/backend/internal/profile"
)

const (
	healthcheckCommand           = "healthcheck"
	invalidConfigurationCategory = "invalid_configuration"
	databaseUnavailableCategory  = "database_unavailable"
	migrationFailureCategory     = "migration_failure"
	httpServerFailureCategory    = "http_server_failure"
	shutdownTimeoutCategory      = "shutdown_timeout"
	internalFailureCategory      = "internal_failure"
	standardLibraryCategory      = "standard_library"
)

func main() {
	if err := execute(); err != nil {
		os.Exit(1)
	}
}

func execute() error {
	if len(os.Args) == 2 && os.Args[1] == healthcheckCommand {
		healthcheckConfig, err := config.LoadHealthcheck()
		client := &http.Client{Timeout: 2 * time.Second}
		if err != nil || localHealthcheck(healthcheckConfig, client) != nil {
			return errors.New("healthcheck failed")
		}
		return nil
	}
	logger := slog.New(httpapi.NewRequestLogHandler(slog.NewJSONHandler(os.Stdout, nil)))
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(rootCtx, logger); err != nil {
		logger.Error("findur stopped", "category", failureCategory(err))
		return err
	}
	return nil
}

type healthcheckClient interface {
	Do(*http.Request) (*http.Response, error)
}

func localHealthcheck(cfg config.HealthcheckConfig, client healthcheckClient) error {
	request, err := http.NewRequest(http.MethodGet, cfg.URL, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode != http.StatusOK {
		return errors.New("health endpoint unavailable")
	}
	return nil
}

func run(rootCtx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return errInvalidConfiguration
	}
	if cfg.Production && buildinfo.ValidateProduction() != nil {
		return errInvalidConfiguration
	}
	if err := migrations.Up(rootCtx, cfg.MigrationURL, cfg.DatabaseURL); err != nil {
		if rootCtx.Err() != nil {
			return nil
		}
		return errMigration
	}
	pool, err := pgxpool.New(rootCtx, cfg.DatabaseURL)
	if err != nil {
		return errDatabase
	}

	readiness := httpapi.NewReadiness(pool, config.ReadinessTimeout)
	checkCtx, cancelCheck := context.WithTimeout(rootCtx, config.ReadinessTimeout)
	err = pool.Ping(checkCtx)
	cancelCheck()
	startupResult := startupPingResult(rootCtx, err)
	if err != nil || rootCtx.Err() != nil {
		pool.Close()
		return startupResult
	}

	var diagnostics *httpapi.Diagnostics
	if cfg.FixtureBaseURL != nil {
		providerHTTPClient := &http.Client{Timeout: config.ProviderTimeout}
		providerClient := provider.NewClient(cfg.FixtureBaseURL, cfg.FixtureProviderToken, providerHTTPClient)
		diagnostics = httpapi.NewDiagnostics(cfg.FixtureBaseURL, providerClient)
	}
	authorization, err := buildAuthorization(cfg, pool)
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	sessions, err := buildSessions(cfg, pool)
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	portfolioServices, err := buildPortfolioServices(cfg, pool, logger)
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	showcaseService, err := portfolio.NewShowcaseService(postgresadapter.NewShowcaseRepository(pool, time.Now), time.Now)
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	profiles, err := profile.NewService(postgresadapter.NewProfileRepository(pool), time.Now)
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	preferences, err := profile.NewPreferenceService(postgresadapter.NewPreferenceRepository(pool, time.Now))
	if err != nil {
		pool.Close()
		return errInvalidConfiguration
	}
	server := newServer(cfg.Address, httpapi.NewHandlerWithProfilePreferences(logger, readiness, buildinfo.SHA, diagnostics, authorization.initiator, authorization.callback, sessions, portfolioServices.inventory, portfolioServices.inclusion, showcaseService, profiles, preferences, cfg.Authorization.Enabled, cfg.Session.PublicOrigin, authorization.fixture), logger)
	serverErrors := make(chan error, 1)
	workerCtx, stopWorker := context.WithCancel(rootCtx)
	defer stopWorker()
	workerDone := make(chan struct{})
	go func() {
		logger.Info("http server starting", "address", cfg.Address)
		serverErrors <- server.ListenAndServe()
	}()
	if portfolioServices.sync != nil {
		go func() {
			defer close(workerDone)
			portfolio.RunSyncWorker(workerCtx, portfolio.SyncWorkerInterval, portfolioServices.sync)
		}()
	} else {
		close(workerDone)
	}
	readiness.SetReady(true)

	select {
	case err := <-serverErrors:
		readiness.SetReady(false)
		stopWorker()
		waitForWorker(logger, workerDone, config.ShutdownDrain)
		pool.Close()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errHTTPServer
	case <-rootCtx.Done():
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.ShutdownDrain)
	defer cancelShutdown()
	stopWorker()
	if err := lifecycle.Drain(shutdownCtx, readiness, server, func() {
		select {
		case <-workerDone:
			logger.Info("portfolio sync worker drained", "event", "portfolio_sync_worker_drained")
		case <-shutdownCtx.Done():
			logger.Warn("portfolio sync worker drain deadline reached", "event", "portfolio_sync_worker_drain_timeout")
		}
		pool.Close()
	}); err != nil {
		return errShutdown
	}
	return nil
}

func waitForWorker(logger *slog.Logger, done <-chan struct{}, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		logger.Info("portfolio sync worker drained", "event", "portfolio_sync_worker_drained")
	case <-timer.C:
		logger.Warn("portfolio sync worker drain deadline reached", "event", "portfolio_sync_worker_drain_timeout")
	}
}

type portfolioComponents struct {
	inventory *portfolio.Service
	inclusion *portfolio.InclusionService
	sync      *portfolio.SyncService
}

func buildPortfolioServices(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) (portfolioComponents, error) {
	if len(cfg.Session.HashKey) == 0 || len(cfg.Authorization.TokenKeys) == 0 || cfg.Authorization.ProviderBaseURL == nil {
		return portfolioComponents{}, nil
	}
	tokens, err := auth.NewTokenCipher(auth.SnapTradeProvider, cfg.Authorization.TokenKeys, cfg.Authorization.CurrentTokenKey, rand.Reader)
	if err != nil {
		return portfolioComponents{}, err
	}
	providerClient, err := provider.NewInventoryClient(cfg.Authorization.ProviderBaseURL, &http.Client{Timeout: config.ProviderTimeout}, time.Now)
	if err != nil {
		return portfolioComponents{}, err
	}
	discovery := oidc.NewDiscoveryClient(cfg.Authorization.Issuer, &http.Client{Timeout: config.ProviderTimeout})
	credentials, err := auth.NewCredentialSource(postgresadapter.NewCredentialRepository(pool), tokens, oidc.NewCallbackClient(discovery, cfg.Authorization.ClientID, cfg.Authorization.ClientSecret, cfg.Authorization.CallbackURL), logger, time.Now, config.AuthorizationTimeout, 30*time.Second, config.AuthorizationTimeout)
	if err != nil {
		return portfolioComponents{}, err
	}
	inventory, err := portfolio.NewService(postgresadapter.NewInventoryRepository(pool), providerClient, credentials, time.Now, config.ProviderTimeout)
	if err != nil {
		return portfolioComponents{}, err
	}
	inclusion, err := portfolio.NewInclusionService(postgresadapter.NewInclusionRepository(pool), time.Now)
	if err != nil {
		return portfolioComponents{}, err
	}
	syncService, err := portfolio.NewSyncService(postgresadapter.NewSyncRepository(pool), providerClient, credentials, time.Now, config.ProviderTimeout, logger)
	if err != nil {
		return portfolioComponents{}, err
	}
	return portfolioComponents{inventory: inventory, inclusion: inclusion, sync: syncService}, nil
}

func buildInventory(cfg config.Config, pool *pgxpool.Pool) (*portfolio.Service, error) {
	if len(cfg.Session.HashKey) == 0 || len(cfg.Authorization.TokenKeys) == 0 || cfg.Authorization.ProviderBaseURL == nil {
		return nil, nil
	}
	tokens, err := auth.NewTokenCipher(auth.SnapTradeProvider, cfg.Authorization.TokenKeys, cfg.Authorization.CurrentTokenKey, rand.Reader)
	if err != nil {
		return nil, err
	}
	providerClient, err := provider.NewInventoryClient(cfg.Authorization.ProviderBaseURL, &http.Client{Timeout: config.ProviderTimeout}, time.Now)
	if err != nil {
		return nil, err
	}
	discovery := oidc.NewDiscoveryClient(cfg.Authorization.Issuer, &http.Client{Timeout: config.ProviderTimeout})
	credentials, err := auth.NewCredentialSource(postgresadapter.NewCredentialRepository(pool), tokens, oidc.NewCallbackClient(discovery, cfg.Authorization.ClientID, cfg.Authorization.ClientSecret, cfg.Authorization.CallbackURL), slog.Default(), time.Now, config.AuthorizationTimeout, 30*time.Second, config.AuthorizationTimeout)
	if err != nil {
		return nil, err
	}
	return portfolio.NewService(postgresadapter.NewInventoryRepository(pool), providerClient, credentials, time.Now, config.ProviderTimeout)
}

type authorizationComponents struct {
	initiator *auth.Service
	callback  *auth.CallbackService
	fixture   http.Handler
}

func buildSessions(cfg config.Config, pool *pgxpool.Pool) (*auth.SessionService, error) {
	if len(cfg.Session.HashKey) == 0 {
		return nil, nil
	}
	return auth.NewSessionService(auth.SessionConfig{HashKey: cfg.Session.HashKey, Clock: time.Now}, postgresadapter.NewSessionRepository(pool))
}

func buildAuthorization(cfg config.Config, pool *pgxpool.Pool) (authorizationComponents, error) {
	if !cfg.Authorization.Enabled {
		return authorizationComponents{}, nil
	}
	discovery := oidc.NewDiscoveryClient(cfg.Authorization.Issuer, &http.Client{Timeout: config.ProviderTimeout})
	repository := postgresadapter.NewOAuthAttemptRepository(pool)
	initiator, err := auth.NewService(auth.Config{
		Enabled:          true,
		ClientID:         cfg.Authorization.ClientID,
		CallbackURL:      cfg.Authorization.CallbackURL,
		AllowedReturns:   cfg.Authorization.AllowedReturns,
		DefaultReturn:    auth.DefaultReturnRoute,
		HashKey:          cfg.Authorization.HashKey,
		EncryptionKey:    cfg.Authorization.EncryptionKey,
		Random:           rand.Reader,
		Clock:            time.Now,
		OperationTimeout: config.AuthorizationTimeout,
	}, repository, discovery)
	if err != nil {
		return authorizationComponents{}, err
	}
	callback, err := auth.NewCallbackService(auth.CallbackConfig{
		Provider:         auth.SnapTradeProvider,
		CallbackURL:      cfg.Authorization.CallbackURL,
		AttemptHashKey:   cfg.Authorization.HashKey,
		SessionHashKey:   cfg.Session.HashKey,
		VerifierKey:      cfg.Authorization.EncryptionKey,
		TokenKeys:        cfg.Authorization.TokenKeys,
		CurrentTokenKey:  cfg.Authorization.CurrentTokenKey,
		Random:           rand.Reader,
		Clock:            time.Now,
		OperationTimeout: config.AuthorizationTimeout,
	}, repository, oidc.NewCallbackClient(discovery, cfg.Authorization.ClientID, cfg.Authorization.ClientSecret, cfg.Authorization.CallbackURL))
	if err != nil {
		return authorizationComponents{}, err
	}
	fixture, err := buildOIDCFixture(cfg)
	if err != nil {
		return authorizationComponents{}, err
	}
	return authorizationComponents{initiator: initiator, callback: callback, fixture: fixture}, nil
}

func buildOIDCFixture(cfg config.Config) (http.Handler, error) {
	if !cfg.Integration || cfg.FixtureBaseURL == nil {
		return nil, nil
	}
	return oidcfixture.New(cfg.Authorization.Issuer, cfg.Authorization.ClientID, cfg.Authorization.ClientSecret, cfg.Authorization.CallbackURL)
}

func newServer(address string, handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ErrorLog:          log.New(safeServerErrorWriter{logger: logger}, "", 0),
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}
}

type safeServerErrorWriter struct {
	logger *slog.Logger
}

func (w safeServerErrorWriter) Write(message []byte) (int, error) {
	w.logger.Error("http server error", "category", standardLibraryCategory)
	return len(message), nil
}

func startupPingResult(rootCtx context.Context, pingErr error) error {
	if rootCtx.Err() != nil {
		return nil
	}
	if pingErr != nil {
		return errDatabase
	}
	return nil
}

var (
	errInvalidConfiguration = errors.New("invalid configuration")
	errDatabase             = errors.New("database unavailable")
	errMigration            = errors.New("database migration failed")
	errHTTPServer           = errors.New("http server failure")
	errShutdown             = errors.New("shutdown deadline exceeded")
)

func failureCategory(err error) string {
	switch {
	case errors.Is(err, errInvalidConfiguration):
		return invalidConfigurationCategory
	case errors.Is(err, errDatabase):
		return databaseUnavailableCategory
	case errors.Is(err, errMigration):
		return migrationFailureCategory
	case errors.Is(err, errHTTPServer):
		return httpServerFailureCategory
	case errors.Is(err, errShutdown):
		return shutdownTimeoutCategory
	default:
		return internalFailureCategory
	}
}
