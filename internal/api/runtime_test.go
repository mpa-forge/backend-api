package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/usersvc"
)

func TestNewRouterServesHealthAndConnectRoutes(t *testing.T) {
	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		usersvc.NewServer(),
	)

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "root metadata",
			method:       http.MethodGet,
			path:         "/",
			wantStatus:   http.StatusOK,
			wantContains: "backend-api",
		},
		{
			name:         "health",
			method:       http.MethodGet,
			path:         "/healthz",
			wantStatus:   http.StatusOK,
			wantContains: "\"status\":\"ok\"",
		},
		{
			name:         "ready",
			method:       http.MethodGet,
			path:         "/readyz",
			wantStatus:   http.StatusOK,
			wantContains: "\"status\":\"ready\"",
		},
		{
			name:         "connect procedure",
			method:       http.MethodPost,
			path:         "/blueprint.user.v1.UserService/GetCurrentUser",
			body:         "{}",
			wantStatus:   http.StatusOK,
			wantContains: "Local Placeholder User",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			if resp.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", resp.Code, test.wantStatus, resp.Body.String())
			}
			if !strings.Contains(resp.Body.String(), test.wantContains) {
				t.Fatalf("body = %q, want substring %q", resp.Body.String(), test.wantContains)
			}
		})
	}
}
