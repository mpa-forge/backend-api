package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mpa-forge/backend-api/internal/runtime"
)

func main() {
	cfg, err := runtime.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup validation failed: %v\n", err)
		os.Exit(1)
	}

	logger := runtime.NewLogger(cfg.LogLevel)
	server := runtime.NewServer(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
