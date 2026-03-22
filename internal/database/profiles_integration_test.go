package database

import (
	"context"
	"os"
	"testing"
)

func TestProfileStoreRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	migrator := NewMigrator(databaseURL)
	if err := migrator.Prepare(ctx); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	pool, err := OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("OpenPool() error = %v", err)
	}
	defer pool.Close()

	store := NewProfileStore(pool)
	provisioned, err := store.UpsertUserProfile(
		ctx,
		"clerk_integration_test_user",
		"integration.user@example.com",
		"Integration User",
		"user",
	)
	if err != nil {
		t.Fatalf("UpsertUserProfile() error = %v", err)
	}

	loaded, err := store.GetUserProfileByClerkUserID(ctx, "clerk_integration_test_user")
	if err != nil {
		t.Fatalf("GetUserProfileByClerkUserID() error = %v", err)
	}

	if loaded.ClerkUserID != provisioned.ClerkUserID {
		t.Fatalf("clerkUserID = %q, want %q", loaded.ClerkUserID, provisioned.ClerkUserID)
	}
	if loaded.Email != "integration.user@example.com" {
		t.Fatalf("email = %q, want %q", loaded.Email, "integration.user@example.com")
	}
	if loaded.DisplayName != "Integration User" {
		t.Fatalf("displayName = %q, want %q", loaded.DisplayName, "Integration User")
	}
	if loaded.Role != "user" {
		t.Fatalf("role = %q, want %q", loaded.Role, "user")
	}
}
