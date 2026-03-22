# backend-api

Go API service repository for the platform blueprint.

## Structure

- `cmd/`: service entrypoints
- `internal/`: private application code
- `pkg/`: shareable public packages
- `deploy/`: deployment manifests and charts
- `docs/`: service-specific documentation
- `scripts/`: local utility and developer scripts

## Toolchain

- GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Go `1.25.1`
- Version pin source: `.tool-versions` and `go.mod`

Windows note:

- use a POSIX-friendly GNU Make such as `ezwinports.make` or MSYS2 `make`
- ensure Git for Windows `bash.exe` is on `PATH`
- do not use `GnuWin32` make for this repo

## Setup

Before running bootstrap:

- Shared workspace requirement: keep `platform-blueprint-specs` checked out as a sibling directory if you want to use `make doctor`.
- Required: GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Recommended: `mise` or `asdf` for automatic tool installation from `.tool-versions`
- Fallback: manually install the pinned tool versions listed above

Run the setup commands from the repository root:

- Workstation checks: `make doctor`
- Bootstrap: `make bootstrap`

Bootstrap validates the pinned Go toolchain and runs `go mod download`.
If `mise` or `asdf` is available, the script will use it to install the pinned toolchain automatically.

## Lint and Format

- Install git hooks: `make precommit-install`
- Run all pre-commit checks manually: `make precommit-run`
- Run repo lint checks: `make lint`
- Run the Go test suite: `make test`
- Apply formatting: `make format`
- Check formatting only: `make format-check`

## Environment

- Copy `.env.example` to `.env` for local development
- Required local baseline variables:
  - `APP_ENV`
  - `LOG_LEVEL`
  - `HTTP_PORT`
  - `DATABASE_URL`
- Planned now and enforced once the runtime exists in Phase 2:
  - `AUTH_ISSUER_URL`
  - `AUTH_AUDIENCE`
  - `OTEL_MODE`
  - `OTEL_EXPORTER_OTLP_ENDPOINT`
  - `OTEL_EXPORTER_OTLP_HEADERS`

The API runtime now validates all of the variables above at startup and exits
before binding a port when required values are missing or malformed.
See `docs/api-runtime.md` for endpoint and runtime details.

Protected API procedures expect a Clerk bearer token and currently map token
claims into the baseline internal profile as follows:

- required identity: `sub`
- optional profile fields: `email`, `display_name`, `given_name`, `family_name`
- optional role fields: `role` or `roles`

Recognized internal roles are `user` and `admin`. Requests with a valid token
but an unsupported role claim receive `403 Forbidden`.

## Run

For native API work:

- Start support services from this repo: `make support-up`
- Run the API locally: `make run`
- Stop support services: `make support-down`

`make run` sources `.env` when present and starts the API on the configured
`HTTP_PORT`.
Support services come from the centralized compose stack in `../platform-infra`.
After code changes, rerun `make run` to restart the native API process.

## Container

- Build a local image: `docker build -t backend-api:local .`
- The image packages the `cmd/api` runtime skeleton
- Runtime configuration is still supplied through environment variables at container start
- The container healthcheck follows the configured `HTTP_PORT`

## Local Stack

- API-focused mode:
  - run `make support-up`
  - run `make run`
  - compose provides `frontend-web` on `http://localhost:3000` and Postgres on `localhost:5432`
- Frontend-focused mode is orchestrated from `frontend-web`, where compose provides the containerized API on `http://localhost:8080`

## Test

The repo now includes focused Go tests for config validation and HTTP routing.

- Run the test suite: `make test`
- Run lint + tests + formatting checks: `make lint`, `make test`, `make format-check`
