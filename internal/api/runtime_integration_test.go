package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/database"
	"github.com/mpa-forge/backend-api/internal/usersvc"
)

func TestProvisioningAndGetCurrentUserWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	migrator := database.NewMigrator(databaseURL)
	if err := migrator.Prepare(ctx); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	pool, err := database.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("OpenPool() error = %v", err)
	}
	defer pool.Close()

	handler := newRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
		staticVerifier{principal: auth.Principal{
			UserID:      "clerk_api_integration_user",
			Email:       "api.integration@example.com",
			DisplayName: "API Integration User",
			Role:        auth.RoleUser,
		}},
		usersvc.NewServer(database.NewProfileStore(pool)),
	)

	provisionReq := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/EnsureCurrentUserProfile", strings.NewReader("{}"))
	provisionReq.Header.Set("Authorization", "Bearer valid-token")
	provisionReq.Header.Set("Content-Type", "application/json")
	provisionResp := httptest.NewRecorder()
	handler.ServeHTTP(provisionResp, provisionReq)
	if provisionResp.Code != http.StatusOK {
		t.Fatalf("provision status = %d, want %d; body = %s", provisionResp.Code, http.StatusOK, provisionResp.Body.String())
	}
	if !strings.Contains(provisionResp.Body.String(), "API Integration User") {
		t.Fatalf("provision body = %q, want display name", provisionResp.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodPost, "/blueprint.user.v1.UserService/GetCurrentUser", strings.NewReader("{}"))
	getReq.Header.Set("Authorization", "Bearer valid-token")
	getReq.Header.Set("Content-Type", "application/json")
	getResp := httptest.NewRecorder()
	handler.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d; body = %s", getResp.Code, http.StatusOK, getResp.Body.String())
	}
	if !strings.Contains(getResp.Body.String(), "api.integration@example.com") {
		t.Fatalf("get body = %q, want stored email", getResp.Body.String())
	}
}
