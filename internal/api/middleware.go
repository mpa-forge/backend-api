package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mpa-forge/platform-observability/backendobs"
)

var localFrontendOrigins = map[string]struct{}{
	"http://localhost:3000": {},
	"http://127.0.0.1:3000": {},
}

func requestLogger(logger *slog.Logger, obsRuntime *backendobs.Runtime) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrapped, r)

			attrs := []any{
				slog.String("request_id", chimiddleware.GetReqID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.Status()),
				slog.Int("bytes", wrapped.BytesWritten()),
				slog.Duration("duration", time.Since(startedAt)),
			}
			for _, attr := range obsRuntime.CorrelationAttrs(r.Context()) {
				attrs = append(attrs, attr)
			}

			logger.Info("http request completed", attrs...)
		})
	}
}

func observabilityRouteLabel(r *http.Request) string {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		if routePattern := routeCtx.RoutePattern(); routePattern != "" {
			return routePattern
		}
	}

	if r.URL == nil || r.URL.Path == "" {
		return "unknown"
	}

	return r.URL.Path
}

func browserCORS(appEnv string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !isAllowedBrowserOrigin(appEnv, origin) {
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					http.Error(w, "cors origin forbidden", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			headers.Set("Vary", "Origin")
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")

			if requestedHeaders := r.Header.Get("Access-Control-Request-Headers"); requestedHeaders != "" {
				headers.Set("Access-Control-Allow-Headers", requestedHeaders)
			} else {
				headers.Set(
					"Access-Control-Allow-Headers",
					"Authorization,Content-Type,Connect-Protocol-Version,Connect-Timeout-Ms,Grpc-Timeout,X-Grpc-Web,X-User-Agent",
				)
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedBrowserOrigin(appEnv, origin string) bool {
	if appEnv != "local" {
		return false
	}

	_, ok := localFrontendOrigins[origin]
	return ok
}
