package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/mpa-forge/backend-api/internal/api"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/database"
	"github.com/mpa-forge/backend-api/internal/usersvc"
	"github.com/mpa-forge/platform-observability/backendobs"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	observabilityRuntime, err := newObservabilityRuntime(ctx, cfg)
	if err != nil {
		logger.Error("observability startup failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := observabilityRuntime.Shutdown(shutdownCtx); err != nil {
			logger.Error("observability shutdown failed", slog.Any("error", err))
		}
	}()

	metadata := observabilityRuntime.Metadata()
	logger.Info(
		"initialized observability runtime",
		slog.String("service", metadata.ServiceName),
		slog.String("service_version", metadata.ServiceVersion),
		slog.String("environment", metadata.Environment),
		slog.String("telemetry_mode", string(metadata.Mode)),
		slog.String("telemetry_profile", string(metadata.Profile)),
	)

	pool, err := database.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database startup failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := api.Run(ctx, cfg, logger, usersvc.NewServer(database.NewProfileStore(pool))); err != nil {
		logger.Error("api runtime stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func newObservabilityRuntime(ctx context.Context, cfg config.Config) (*backendobs.Runtime, error) {
	otlpEndpoint := ""
	if cfg.Telemetry.OTLPEndpoint != nil {
		otlpEndpoint = cfg.Telemetry.OTLPEndpoint.String()
	}

	return backendobs.Init(ctx, backendobs.Config{
		ServiceName:    "backend-api",
		ServiceVersion: buildVersion(),
		Environment:    cfg.AppEnv,
		Mode:           backendobs.Mode(cfg.Telemetry.Mode),
		Profile:        backendobs.Profile(cfg.Telemetry.Profile),
		OTLPEndpoint:   otlpEndpoint,
		OTLPHeaders:    cfg.Telemetry.OTLPHeaders,
	})
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}

	return info.Main.Version
}
