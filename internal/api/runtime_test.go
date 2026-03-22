package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/usersvc"
)

func TestNewRouterServesHealthAndConnectRoutes(t *testing.T) {
	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		staticVerifier{principal: auth.Principal{UserID: "user_123", Email: "user@example.com", DisplayName: "Example User", Role: auth.RoleUser}},
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
			wantContains: "Example User",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.body != "" {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer valid-token")
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

func TestNewRouterRejectsUnauthorizedAndForbiddenConnectRequests(t *testing.T) {
	tests := []struct {
		name        string
		verifier    auth.Verifier
		authorize   string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "missing bearer token",
			verifier:    staticVerifier{},
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "unauthenticated",
		},
		{
			name:        "invalid token",
			verifier:    staticVerifier{err: auth.ErrUnauthenticated},
			authorize:   "Bearer invalid-token",
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "unauthenticated",
		},
		{
			name:        "unsupported role",
			verifier:    staticVerifier{err: auth.ErrForbidden},
			authorize:   "Bearer forbidden-token",
			wantStatus:  http.StatusForbidden,
			wantMessage: "permission_denied",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newRouter(
				config.Config{AppEnv: "test"},
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				test.verifier,
				usersvc.NewServer(),
			)

			req := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/GetCurrentUser", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			if test.authorize != "" {
				req.Header.Set("Authorization", test.authorize)
			}

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			if resp.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", resp.Code, test.wantStatus, resp.Body.String())
			}
			if !strings.Contains(resp.Body.String(), test.wantMessage) {
				t.Fatalf("body = %q, want substring %q", resp.Body.String(), test.wantMessage)
			}
		})
	}
}

type staticVerifier struct {
	principal auth.Principal
	err       error
}

func (v staticVerifier) VerifyToken(_ context.Context, _ string) (auth.Principal, error) {
	if v.err != nil {
		return auth.Principal{}, v.err
	}
	return v.principal, nil
}
