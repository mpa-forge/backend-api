package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
)

// Verifier validates a bearer token and returns the authenticated principal.
type Verifier interface {
	VerifyToken(ctx context.Context, token string) (Principal, error)
}

// NewAuthInterceptor constructs a Connect interceptor that enforces bearer-token
// authentication for protected RPCs.
func NewAuthInterceptor(verifier Verifier) connect.Interceptor {
	return authInterceptor{verifier: verifier}
}

type authInterceptor struct {
	verifier Verifier
}

func (i authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		token, err := bearerToken(req.Header().Get("Authorization"))
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		principal, err := i.verifier.VerifyToken(ctx, token)
		if err != nil {
			switch {
			case errors.Is(err, ErrUnauthenticated):
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			case errors.Is(err, ErrForbidden):
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			default:
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("verify bearer token: %w", err))
			}
		}

		return next(WithPrincipal(ctx, principal), req)
	}
}

func (authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func bearerToken(authorization string) (string, error) {
	if authorization == "" {
		return "", fmt.Errorf("%w: missing bearer token", ErrUnauthenticated)
	}

	scheme, token, found := strings.Cut(authorization, " ")
	if !found || !strings.EqualFold(strings.TrimSpace(scheme), "Bearer") || strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("%w: malformed Authorization header", ErrUnauthenticated)
	}

	return strings.TrimSpace(token), nil
}
