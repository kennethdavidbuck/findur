package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/config"

	"github.com/kennethdavidbuck/findur/backend/internal/platform/httpapi"
	"github.com/kennethdavidbuck/findur/backend/internal/platform/lifecycle"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(rootCtx, logger); err != nil {
		logger.Error("findur stopped", "category", failureCategory(err))
		os.Exit(1)
	}
}

func run(rootCtx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return errInvalidConfiguration
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

	server := newServer(cfg.Address, httpapi.NewHandler(logger, readiness), logger)
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "address", cfg.Address)
		serverErrors <- server.ListenAndServe()
	}()
	readiness.SetReady(true)

	select {
	case err := <-serverErrors:
		readiness.SetReady(false)
		pool.Close()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errHTTPServer
	case <-rootCtx.Done():
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.ShutdownDrain)
	defer cancelShutdown()
	if err := lifecycle.Drain(shutdownCtx, readiness, server, pool.Close); err != nil {
		return errShutdown
	}
	return nil
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
	w.logger.Error("http server error", "category", "standard_library")
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
	errHTTPServer           = errors.New("http server failure")
	errShutdown             = errors.New("shutdown deadline exceeded")
)

func failureCategory(err error) string {
	switch {
	case errors.Is(err, errInvalidConfiguration):
		return "invalid_configuration"
	case errors.Is(err, errDatabase):
		return "database_unavailable"
	case errors.Is(err, errHTTPServer):
		return "http_server_failure"
	case errors.Is(err, errShutdown):
		return "shutdown_timeout"
	default:
		return "internal_failure"
	}
}
