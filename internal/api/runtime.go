// Package api wires the chi router, Connect handlers, and HTTP server lifecycle.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1/userv1connect"
)

const shutdownTimeout = 10 * time.Second

// Run starts the API server and blocks until the listener stops or the context is cancelled.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, userService userv1connect.UserServiceHandler) error {
	server := &http.Server{
		Addr:              net.JoinHostPort("", strconv.Itoa(cfg.HTTPPort)),
		Handler:           newRouter(cfg, logger, auth.NewClerkVerifier(cfg.AuthIssuerURL, cfg.AuthAudience), userService),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info(
			"starting api server",
			slog.String("address", server.Addr),
			slog.String("environment", cfg.AppEnv),
			slog.String("telemetry_mode", string(cfg.Telemetry.Mode)),
			slog.String("telemetry_profile", cfg.Telemetry.Profile),
		)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutting down api server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown api server: %w", err)
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped with error: %w", err)
	}

	return nil
}

func newRouter(cfg config.Config, logger *slog.Logger, verifier auth.Verifier, userService userv1connect.UserServiceHandler) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(browserCORS(cfg.AppEnv))
	router.Use(requestLogger(logger))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"service":     "backend-api",
			"environment": cfg.AppEnv,
			"procedures": []string{
				userv1connect.UserServiceEnsureCurrentUserProfileProcedure,
				userv1connect.UserServiceGetCurrentUserProcedure,
			},
		})
	})

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "backend-api",
			"status":  "ok",
		})
	})

	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Readiness only reflects in-process startup here; dependency checks land in later tasks.
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "backend-api",
			"status":  "ready",
		})
	})

	path, handler := userv1connect.NewUserServiceHandler(userService, connect.WithInterceptors(auth.NewAuthInterceptor(verifier)))
	router.Mount(path, handler)

	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
