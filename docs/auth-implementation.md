# Auth Implementation

## Purpose

This document explains how `backend-api` enforces the Phase 2 auth baseline.

Use this file when working on the API runtime itself. Consumers in other repos
should prefer the contract-facing auth documentation in
`../platform-contracts/docs/consumer-auth-usage.md`.

## Runtime Shape

Protected API procedures are enforced through a Connect interceptor, not a plain
`chi` middleware chain.

That keeps auth enforcement attached to generated Connect handlers while still
running inside the shared HTTP server built with `chi`.

Current flow:

1. `chi` hosts the HTTP server and public endpoints.
2. The generated `UserService` Connect handler is created with an auth interceptor.
3. The auth interceptor extracts and verifies the bearer token.
4. The verified principal is stored in request context.
5. The provisioning handler upserts the local profile row from that principal.
6. The read handler loads the local profile from Postgres by Clerk subject.

Relevant code:

- `internal/api/runtime.go`
- `internal/auth/interceptor.go`
- `internal/auth/verifier.go`
- `internal/auth/jwks.go`
- `internal/database/profiles.go`
- `internal/database/sqlc/`
- `internal/usersvc/server.go`

## Verification Rules

The verifier currently:

- fetches JWKS from `AUTH_ISSUER_URL/.well-known/jwks.json`
- caches signing keys in memory for a short TTL
- validates:
  - signature
  - issuer
  - audience
- maps claims into the internal principal used by handlers

Current baseline claim usage:

- required:
  - `sub`
- optional profile fields:
  - `email`
  - `display_name`
  - `given_name`
  - `family_name`
- optional role fields:
  - `role`
  - `roles`

The verified Clerk `sub` is the backend's stable external identity key. The API
uses that `sub` directly as `user_profiles.clerk_user_id` both when
provisioning the row and when reading it later.

## Role Mapping

The verifier maps token claims to these internal roles:

- `user`
- `admin`

Behavior:

- if no recognized role claim is present, the caller defaults to `user`
- if a role claim is present but does not map to `user` or `admin`, the request
  is rejected with `403`

## Error Mapping

The auth interceptor returns Connect errors so generated clients get protocol-
appropriate responses.

Current mapping:

- invalid or missing bearer token -> `connect.CodeUnauthenticated` -> HTTP `401`
- valid token with unsupported role -> `connect.CodePermissionDenied` -> HTTP `403`
- valid token with no local row on `GetCurrentUser` -> `connect.CodeNotFound` -> HTTP `404`
- unexpected verifier failure -> `connect.CodeInternal`

## Tests

Auth behavior is covered by:

- `internal/auth/interceptor_test.go`
- `internal/auth/verifier_test.go`
- `internal/api/runtime_test.go`

These tests cover:

- missing and malformed bearer tokens
- invalid issuer/audience
- JWKS-backed signature verification
- unsupported role handling
- authenticated route success through the mounted Connect handlers
- explicit profile provisioning from the verified principal
- DB-backed profile reads after provisioning
