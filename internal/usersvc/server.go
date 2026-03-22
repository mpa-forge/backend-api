// Package usersvc provides the first Connect service implementation backed by generated contracts.
package usersvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/mpa-forge/backend-api/internal/auth"
	userv1 "github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1"
	"github.com/mpa-forge/platform-contracts/gen/go/blueprint/user/v1/userv1connect"
)

var _ userv1connect.UserServiceHandler = (*Server)(nil)

// Server implements the generated UserService Connect handler contract.
type Server struct{}

// NewServer constructs the placeholder user service used by the Phase 2 runtime skeleton.
func NewServer() *Server {
	return &Server{}
}

// GetCurrentUser returns the authenticated principal extracted from the verified
// Clerk bearer token.
func (s *Server) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[userv1.GetCurrentUserRequest],
) (*connect.Response[userv1.GetCurrentUserResponse], error) {
	_ = req

	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authenticated principal missing from request context"))
	}

	return connect.NewResponse(&userv1.GetCurrentUserResponse{
		User: &userv1.UserProfile{
			UserId:      principal.UserID,
			Email:       principal.Email,
			DisplayName: principal.DisplayName,
			Role:        string(principal.Role),
		},
	}), nil
}
