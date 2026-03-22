package usersvc

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/database"
	userv1 "github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1"
)

func TestEnsureCurrentUserProfileUpsertsFromPrincipal(t *testing.T) {
	store := &fakeProfileStore{
		upsertProfile: database.UserProfile{
			ClerkUserID: "user_123",
			Email:       "user@example.com",
			DisplayName: "Example User",
			Role:        "user",
		},
	}
	server := NewServer(store)
	ctx := auth.WithPrincipal(context.Background(), auth.Principal{
		UserID:      "user_123",
		Email:       "user@example.com",
		DisplayName: "Example User",
		Role:        auth.RoleUser,
	})

	resp, err := server.EnsureCurrentUserProfile(ctx, connect.NewRequest(&userv1.EnsureCurrentUserProfileRequest{}))
	if err != nil {
		t.Fatalf("EnsureCurrentUserProfile() error = %v", err)
	}

	if store.upsertArgs.clerkUserID != "user_123" {
		t.Fatalf("clerkUserID = %q, want %q", store.upsertArgs.clerkUserID, "user_123")
	}
	if got := resp.Msg.GetUser().GetDisplayName(); got != "Example User" {
		t.Fatalf("display name = %q, want %q", got, "Example User")
	}
}

func TestEnsureCurrentUserProfileFallsBackToEmailForDisplayName(t *testing.T) {
	store := &fakeProfileStore{
		upsertProfile: database.UserProfile{
			ClerkUserID: "user_123",
			Email:       "user@example.com",
			DisplayName: "user@example.com",
			Role:        "user",
		},
	}
	server := NewServer(store)
	ctx := auth.WithPrincipal(context.Background(), auth.Principal{
		UserID: "user_123",
		Email:  "user@example.com",
		Role:   auth.RoleUser,
	})

	if _, err := server.EnsureCurrentUserProfile(ctx, connect.NewRequest(&userv1.EnsureCurrentUserProfileRequest{})); err != nil {
		t.Fatalf("EnsureCurrentUserProfile() error = %v", err)
	}

	if got := store.upsertArgs.displayName; got != "user@example.com" {
		t.Fatalf("displayName = %q, want %q", got, "user@example.com")
	}
}

func TestGetCurrentUserLoadsPersistedProfile(t *testing.T) {
	store := &fakeProfileStore{
		getProfile: database.UserProfile{
			ClerkUserID: "user_123",
			Email:       "stored@example.com",
			DisplayName: "Stored User",
			Role:        "admin",
		},
	}
	server := NewServer(store)
	ctx := auth.WithPrincipal(context.Background(), auth.Principal{
		UserID:      "user_123",
		Email:       "token@example.com",
		DisplayName: "Token User",
		Role:        auth.RoleUser,
	})

	resp, err := server.GetCurrentUser(ctx, connect.NewRequest(&userv1.GetCurrentUserRequest{}))
	if err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}

	if got := resp.Msg.GetUser().GetEmail(); got != "stored@example.com" {
		t.Fatalf("email = %q, want %q", got, "stored@example.com")
	}
	if got := resp.Msg.GetUser().GetRole(); got != "admin" {
		t.Fatalf("role = %q, want %q", got, "admin")
	}
}

func TestGetCurrentUserReturnsNotFoundWhenProfileMissing(t *testing.T) {
	server := NewServer(&fakeProfileStore{getErr: pgx.ErrNoRows})
	ctx := auth.WithPrincipal(context.Background(), auth.Principal{UserID: "missing_user", Role: auth.RoleUser})

	_, err := server.GetCurrentUser(ctx, connect.NewRequest(&userv1.GetCurrentUserRequest{}))
	if err == nil {
		t.Fatal("GetCurrentUser() error = nil, want not found")
	}

	connectErr := new(connect.Error)
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Fatalf("connect code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

type fakeProfileStore struct {
	getProfile    database.UserProfile
	getErr        error
	upsertProfile database.UserProfile
	upsertErr     error
	upsertArgs    struct {
		clerkUserID string
		email       string
		displayName string
		role        string
	}
}

func (s *fakeProfileStore) GetUserProfileByClerkUserID(_ context.Context, clerkUserID string) (database.UserProfile, error) {
	if s.getErr != nil {
		return database.UserProfile{}, s.getErr
	}
	if s.getProfile.ClerkUserID == "" {
		s.getProfile.ClerkUserID = clerkUserID
	}
	return s.getProfile, nil
}

func (s *fakeProfileStore) UpsertUserProfile(_ context.Context, clerkUserID, email, displayName, role string) (database.UserProfile, error) {
	s.upsertArgs.clerkUserID = clerkUserID
	s.upsertArgs.email = email
	s.upsertArgs.displayName = displayName
	s.upsertArgs.role = role
	if s.upsertErr != nil {
		return database.UserProfile{}, s.upsertErr
	}
	if s.upsertProfile.ClerkUserID == "" {
		s.upsertProfile.ClerkUserID = clerkUserID
	}
	return s.upsertProfile, nil
}
