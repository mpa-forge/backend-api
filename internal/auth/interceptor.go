package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	traceapi "go.opentelemetry.io/otel/trace"
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
		procedureAttrs := connectSpanAttributes(req)
		setSpanAttributes(ctx, procedureAttrs...)

		token, err := bearerToken(req.Header().Get("Authorization"))
		if err != nil {
			connectErr := connect.NewError(connect.CodeUnauthenticated, err)
			recordConnectError(ctx, connectErr, append(procedureAttrs, attribute.String("auth.result", "missing_bearer_token"))...)
			return nil, connectErr
		}

		principal, err := i.verifier.VerifyToken(ctx, token)
		if err != nil {
			switch {
			case errors.Is(err, ErrUnauthenticated):
				connectErr := connect.NewError(connect.CodeUnauthenticated, err)
				recordConnectError(ctx, connectErr, append(procedureAttrs, attribute.String("auth.result", "unauthenticated"))...)
				return nil, connectErr
			case errors.Is(err, ErrForbidden):
				connectErr := connect.NewError(connect.CodePermissionDenied, err)
				recordConnectError(ctx, connectErr, append(procedureAttrs, attribute.String("auth.result", "forbidden"))...)
				return nil, connectErr
			default:
				connectErr := connect.NewError(connect.CodeInternal, fmt.Errorf("verify bearer token: %w", err))
				recordConnectError(ctx, connectErr, append(procedureAttrs, attribute.String("auth.result", "error"))...)
				return nil, connectErr
			}
		}

		ctx = WithPrincipal(ctx, principal)
		setSpanAttributes(ctx, append(procedureAttrs, attribute.String("auth.result", "authenticated"))...)

		resp, err := next(ctx, req)
		if err != nil {
			recordConnectError(ctx, err, procedureAttrs...)
			return resp, err
		}

		return resp, nil
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

func connectSpanAttributes(req connect.AnyRequest) []attribute.KeyValue {
	procedure := strings.TrimSpace(req.Spec().Procedure)
	attrs := []attribute.KeyValue{attribute.String("rpc.system", "connect")}
	if procedure == "" {
		return attrs
	}

	attrs = append(attrs, attribute.String("rpc.connect.procedure", procedure))

	trimmed := strings.TrimPrefix(procedure, "/")
	service, method, found := strings.Cut(trimmed, "/")
	if found {
		attrs = append(attrs,
			attribute.String("rpc.service", service),
			attribute.String("rpc.method", method),
		)
	}

	return attrs
}

func setSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := traceapi.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() || len(attrs) == 0 {
		return
	}

	span.SetAttributes(attrs...)
}

func recordConnectError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := traceapi.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}

	code := connect.CodeOf(err).String()
	span.SetAttributes(append(attrs, attribute.String("rpc.connect.code", code))...)
	span.RecordError(err)
	span.SetStatus(codes.Error, code)
}
