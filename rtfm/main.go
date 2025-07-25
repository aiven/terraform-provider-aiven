package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aiven/terraform-provider-aiven/rtfm/config"
	"github.com/aiven/terraform-provider-aiven/rtfm/exporter/extract"
	"github.com/aiven/terraform-provider-aiven/rtfm/exporter/provision"
	"github.com/aiven/terraform-provider-aiven/rtfm/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	cfg, err := config.LoadConfig[config.Config]()
	if err != nil {
		fatal(fmt.Errorf("failed to initialize configuration: %w", err))
	}

	srv, err := newServer(cfg)
	if err != nil {
		fatal(fmt.Errorf("failed to initialize server: %w", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	logger.Info("server started")

	select {
	case err = <-serverErrors:
		fatal(fmt.Errorf("server error: %w", err))
	case <-ctx.Done():
		logger.Info("shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()

		if err = srv.Shutdown(shutdownCtx); err != nil {
			fatal(fmt.Errorf("could not gracefully shutdown server: %w", err))
		}

		logger.Info("server stopped")
	}
}

func newServer(cfg *config.Config) (*server.Server, error) {
	exp, err := extract.NewTerraformExecutor(cfg.ExportConfig)
	if err != nil {
		fatal(fmt.Errorf("failed to initialize exporter: %w", err))
	}

	extractHandler := extract.NewHandler(exp)
	provisionHandler := provision.NewHandler(cfg.ProvisionConfig)

	mux := http.NewServeMux()

	// export existing resources
	mux.Handle("GET /resources/existing", extractHandler)

	// mock non-existing resources
	mux.Handle("GET /resources/{project}/provisions", provisionHandler)

	return server.New(cfg.ServerConfig, mux), nil
}

func fatal(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
