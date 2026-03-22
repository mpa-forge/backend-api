package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	dbsqlc "github.com/mpa-forge/backend-api/internal/database/sqlc"
)

// UserProfile is the application-facing profile model loaded from Postgres.
type UserProfile struct {
	ID          int64
	ClerkUserID string
	Email       string
	DisplayName string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProfileStore wraps the generated sqlc query layer with app-facing methods.
type ProfileStore struct {
	queries *dbsqlc.Queries
}

// NewProfileStore constructs the DB-backed profile store.
func NewProfileStore(db dbsqlc.DBTX) *ProfileStore {
	return &ProfileStore{queries: dbsqlc.New(db)}
}

// GetUserProfileByClerkUserID returns the persisted profile for a verified
// Clerk subject.
func (s *ProfileStore) GetUserProfileByClerkUserID(ctx context.Context, clerkUserID string) (UserProfile, error) {
	profile, err := s.queries.GetUserProfileByClerkUserID(ctx, clerkUserID)
	if err != nil {
		return UserProfile{}, err
	}

	return mapUserProfile(profile), nil
}

// UpsertUserProfile creates or refreshes the persisted profile for a verified
// Clerk subject.
func (s *ProfileStore) UpsertUserProfile(ctx context.Context, clerkUserID, email, displayName, role string) (UserProfile, error) {
	profile, err := s.queries.UpsertUserProfile(ctx, dbsqlc.UpsertUserProfileParams{
		ClerkUserID: clerkUserID,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	})
	if err != nil {
		return UserProfile{}, err
	}

	return mapUserProfile(profile), nil
}

func mapUserProfile(profile dbsqlc.UserProfile) UserProfile {
	return UserProfile{
		ID:          profile.ID,
		ClerkUserID: profile.ClerkUserID,
		Email:       profile.Email,
		DisplayName: profile.DisplayName,
		Role:        profile.Role,
		CreatedAt:   timestamptzValue(profile.CreatedAt),
		UpdatedAt:   timestamptzValue(profile.UpdatedAt),
	}
}

func timestamptzValue(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
