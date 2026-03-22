package auth

import "errors"

var (
	// ErrUnauthenticated signals that the caller did not present a valid token.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrForbidden signals that the caller authenticated successfully but does
	// not map to an allowed internal role.
	ErrForbidden = errors.New("forbidden")
)
