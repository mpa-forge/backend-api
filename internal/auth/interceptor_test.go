package auth

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestAuthInterceptorMapsVerifierResults(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		verifier  Verifier
		wantCode  connect.Code
		wantCalls int
	}{
		{
			name:      "missing header",
			wantCode:  connect.CodeUnauthenticated,
			wantCalls: 0,
		},
		{
			name:      "valid principal",
			header:    "Bearer valid-token",
			verifier:  staticVerifier{principal: Principal{UserID: "user_123", Role: RoleUser}},
			wantCalls: 1,
		},
		{
			name:      "invalid token",
			header:    "Bearer invalid-token",
			verifier:  staticVerifier{err: ErrUnauthenticated},
			wantCode:  connect.CodeUnauthenticated,
			wantCalls: 1,
		},
		{
			name:      "forbidden role",
			header:    "Bearer forbidden-token",
			verifier:  staticVerifier{err: ErrForbidden},
			wantCode:  connect.CodePermissionDenied,
			wantCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			verifier := test.verifier
			if verifier == nil {
				verifier = staticVerifier{}
			}
			wrapped := NewAuthInterceptor(countingVerifier{Verifier: verifier, calls: &calls}).WrapUnary(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				principal, ok := PrincipalFromContext(ctx)
				if !ok {
					return nil, errors.New("principal missing from context")
				}
				if principal.UserID == "" {
					return nil, errors.New("principal user id is empty")
				}
				return connect.NewResponse(&struct{}{}), nil
			})

			req := connect.NewRequest(&struct{}{})
			if test.header != "" {
				req.Header().Set("Authorization", test.header)
			}

			_, err := wrapped(context.Background(), req)
			if test.wantCode == 0 {
				if err != nil {
					t.Fatalf("wrapped() error = %v, want nil", err)
				}
			} else {
				var connectErr *connect.Error
				if !errors.As(err, &connectErr) {
					t.Fatalf("wrapped() error = %T, want *connect.Error", err)
				}
				if connectErr.Code() != test.wantCode {
					t.Fatalf("wrapped() code = %v, want %v", connectErr.Code(), test.wantCode)
				}
			}

			if calls != test.wantCalls {
				t.Fatalf("verifier calls = %d, want %d", calls, test.wantCalls)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr error
	}{
		{name: "valid", header: "Bearer abc", want: "abc"},
		{name: "missing", wantErr: ErrUnauthenticated},
		{name: "malformed", header: "Token abc", wantErr: ErrUnauthenticated},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := bearerToken(test.header)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("bearerToken() error = %v, want %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("bearerToken() = %q, want %q", got, test.want)
			}
		})
	}
}

type staticVerifier struct {
	principal Principal
	err       error
}

func (v staticVerifier) VerifyToken(context.Context, string) (Principal, error) {
	if v.err != nil {
		return Principal{}, v.err
	}
	return v.principal, nil
}

type countingVerifier struct {
	Verifier
	calls *int
}

func (v countingVerifier) VerifyToken(ctx context.Context, token string) (Principal, error) {
	*v.calls++
	return v.Verifier.VerifyToken(ctx, token)
}
