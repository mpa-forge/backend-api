// Package usersvc provides the first Connect service implementation backed by generated contracts.
package usersvc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/mpa-forge/backend-api/internal/auth"
	"github.com/mpa-forge/backend-api/internal/database"
	userv1 "github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1"
	"github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1/userv1connect"
)

var _ userv1connect.UserServiceHandler = (*Server)(nil)

// ProfileStore captures the profile persistence operations needed by the user
// service without exposing transport-specific concerns.
type ProfileStore interface {
	GetUserProfileByClerkUserID(ctx context.Context, clerkUserID string) (database.UserProfile, error)
	UpsertUserProfile(ctx context.Context, clerkUserID, email, displayName, role string) (database.UserProfile, error)
}

// Server implements the generated UserService Connect handler contract.
type Server struct {
	profiles ProfileStore
}

// NewServer constructs the user service backed by the typed profile store.
func NewServer(profiles ProfileStore) *Server {
	return &Server{profiles: profiles}
}

// EnsureCurrentUserProfile provisions or refreshes the local profile row for the
// authenticated Clerk user before normal reads rely on persisted data.
func (s *Server) EnsureCurrentUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.EnsureCurrentUserProfileRequest],
) (*connect.Response[userv1.EnsureCurrentUserProfileResponse], error) {
	_ = req

	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.profiles.UpsertUserProfile(
		ctx,
		principal.UserID,
		principal.Email,
		normalizeDisplayName(principal),
		string(principal.Role),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("provision current user profile: %w", err))
	}

	return connect.NewResponse(&userv1.EnsureCurrentUserProfileResponse{
		User: mapProfile(profile),
	}), nil
}

// GetCurrentUser returns the persisted local profile associated with the
// authenticated Clerk bearer token.
func (s *Server) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[userv1.GetCurrentUserRequest],
) (*connect.Response[userv1.GetCurrentUserResponse], error) {
	_ = req

	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.profiles.GetUserProfileByClerkUserID(ctx, principal.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user profile not found for authenticated subject"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("load current user profile: %w", err))
	}

	return connect.NewResponse(&userv1.GetCurrentUserResponse{
		User: mapProfile(profile),
	}), nil
}

func requirePrincipal(ctx context.Context) (auth.Principal, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return auth.Principal{}, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authenticated principal missing from request context"))
	}

	return principal, nil
}

func normalizeDisplayName(principal auth.Principal) string {
	displayName := strings.TrimSpace(principal.DisplayName)
	if displayName != "" {
		return displayName
	}
	if email := strings.TrimSpace(principal.Email); email != "" {
		return email
	}
	return principal.UserID
}

func mapProfile(profile database.UserProfile) *userv1.UserProfile {
	return &userv1.UserProfile{
		UserId:      profile.ClerkUserID,
		Email:       profile.Email,
		DisplayName: profile.DisplayName,
		Role:        profile.Role,
	}
}
