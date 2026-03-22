# API Runtime

`backend-api` now runs a minimal `chi` + Connect skeleton instead of the Phase 1
placeholder server.

## Startup Contract

Startup is fail-fast. The process exits before binding a port when required
environment variables are missing or malformed.

Required variables:

- `APP_ENV`
- `LOG_LEVEL`
- `HTTP_PORT`
- `DATABASE_URL`
- `AUTH_ISSUER_URL`
- `AUTH_AUDIENCE`
- `OTEL_MODE`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `OTEL_EXPORTER_OTLP_HEADERS`

`OTEL_MODE` accepts:

- `disabled`
- `direct_otlp`
- `collector_gateway`

When telemetry is enabled (`direct_otlp` or `collector_gateway`), both OTLP
exporter variables must be non-empty and valid.

## Local Run

`make run` sources `.env` when present, then starts `./cmd/api`.

Typical local flow:

1. Copy `.env.example` to `.env`.
2. Adjust placeholder auth or telemetry values as needed.
3. Run `make run`.

## Validation

- `make test` runs the runtime/config test suite.
- `make lint` runs the Go lint configuration used by CI and pre-commit.
- `make precommit-run` runs the full repo-local validation stack.

## Endpoints

- `GET /`
  - returns basic service metadata and the mounted Connect procedure list
- `GET /healthz`
  - liveness response
- `GET /readyz`
  - readiness response for the in-process runtime skeleton
- `POST /blueprint.user.v1.UserService/GetCurrentUser`
  - generated Connect handler mounted through `chi`
  - requires `Authorization: Bearer <Clerk session token>`

## Authentication

Protected Connect procedures use a Connect interceptor that:

- extracts the bearer token from the `Authorization` header
- verifies the JWT signature against Clerk JWKS at `AUTH_ISSUER_URL/.well-known/jwks.json`
- enforces `iss` against `AUTH_ISSUER_URL`
- enforces `aud` against `AUTH_AUDIENCE`
- maps token role claims into internal `user` or `admin` roles

The runtime accepts the following token claim shape for the baseline:

- required identity: `sub`
- optional profile fields: `email`, `display_name`, `given_name`, `family_name`
- optional role fields: `role` or `roles`

If no recognized role claim is present, the API defaults the caller to `user`.
If a role claim is present but does not map to `user` or `admin`, the API
rejects the request with `403`.

`GetCurrentUser` now reflects the authenticated principal from the verified
token. Database-backed user enrichment still lands in later Phase 2 tasks.

Implementation details:

- `docs/auth-implementation.md`
