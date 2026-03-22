package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mpa-forge/backend-api/internal/api"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/usersvc"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := api.Run(ctx, cfg, logger, usersvc.NewServer()); err != nil {
		logger.Error("api runtime stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}
