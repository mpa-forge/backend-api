# api-authentication Specification

## Purpose

Define the canonical authentication and authorization contract for protected
`backend-api` procedures, including Clerk token verification, claim mapping,
and local profile identity behavior.

## Requirements

### Requirement: Protected procedures require verified Clerk bearer tokens

Protected Connect procedures SHALL enforce authentication through a Connect
interceptor that extracts the `Authorization: Bearer <token>` header, loads
signing keys from `AUTH_ISSUER_URL/.well-known/jwks.json`, validates the JWT
signature, and enforces both the configured issuer and audience.

#### Scenario: Missing or invalid bearer tokens are rejected

- **WHEN** a protected procedure is called without a valid bearer token
- **THEN** the interceptor returns `connect.CodeUnauthenticated` and the HTTP
  response is `401`

#### Scenario: Valid Clerk tokens are accepted

- **WHEN** a protected procedure receives a token with a valid signature,
  matching issuer, and matching audience
- **THEN** the interceptor stores the verified principal in request context for
  downstream handlers

### Requirement: Verified principals use the baseline claim contract

The verified principal SHALL require the `sub` claim, MAY consume `email`,
`display_name`, `given_name`, and `family_name`, and SHALL derive roles from
the `role` or `roles` claims using the internal role set `user` and `admin`.
If no recognized role claim is present, the caller MUST default to `user`.

#### Scenario: Supported roles map into internal roles

- **WHEN** the verified token includes `role` or `roles` values that map to
  `user` or `admin`
- **THEN** the principal uses the mapped internal role for authorization

#### Scenario: Missing role claims fall back to user

- **WHEN** the verified token omits recognized role claims
- **THEN** the principal is assigned the internal `user` role

### Requirement: Unsupported roles are denied consistently

Tokens with role claims that do not map to `user` or `admin` MUST be rejected
as an authorization failure.

#### Scenario: Unsupported role claims receive permission denied

- **WHEN** a verified token contains an unsupported role value
- **THEN** the interceptor returns `connect.CodePermissionDenied` and the HTTP
  response is `403`

### Requirement: Authenticated profile procedures use the verified Clerk subject

`EnsureCurrentUserProfile` and `GetCurrentUser` SHALL use the verified Clerk
`sub` as the stable external identity key for provisioning and reading the local
profile row.

#### Scenario: Profile provisioning uses the verified subject

- **WHEN** `EnsureCurrentUserProfile` runs for an authenticated caller
- **THEN** the handler provisions or refreshes the local profile using the
  verified Clerk `sub`

#### Scenario: Missing local profiles return not found on read

- **WHEN** `GetCurrentUser` is called for an authenticated subject without a
  local profile row
- **THEN** the procedure returns `connect.CodeNotFound` and the HTTP response is
  `404`
