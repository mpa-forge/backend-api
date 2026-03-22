// Package usersvc provides the first Connect service implementation backed by generated contracts.
package usersvc

import (
	"context"

	"connectrpc.com/connect"
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

// GetCurrentUser returns deterministic placeholder profile data until auth and persistence land.
func (s *Server) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[userv1.GetCurrentUserRequest],
) (*connect.Response[userv1.GetCurrentUserResponse], error) {
	_ = ctx
	_ = req

	return connect.NewResponse(&userv1.GetCurrentUserResponse{
		User: &userv1.UserProfile{
			UserId:      "user_local_placeholder",
			Email:       "user@example.com",
			DisplayName: "Local Placeholder User",
			Role:        "user",
		},
	}), nil
}
