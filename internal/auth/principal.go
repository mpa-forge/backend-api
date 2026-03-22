// Package auth verifies Clerk-issued bearer tokens and exposes the authenticated
// principal to Connect handlers.
package auth

import "context"

type contextKey string

const principalContextKey contextKey = "authenticated-principal"

// Role is the internal authorization role used by backend-api.
type Role string

const (
	// RoleUser is the default baseline role for authenticated callers.
	RoleUser Role = "user"
	// RoleAdmin is the elevated role accepted by the Phase 2 baseline.
	RoleAdmin Role = "admin"
)

// Principal is the authenticated identity extracted from a verified Clerk token.
type Principal struct {
	UserID      string
	Email       string
	DisplayName string
	Role        Role
}

// WithPrincipal stores the verified principal in the request context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

// PrincipalFromContext retrieves the verified principal from context.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(Principal)
	return principal, ok
}
