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
- Go `1.24.12`
- Version pin source: `.tool-versions` and `go.mod`

Windows note:
- use a POSIX-friendly GNU Make such as `ezwinports.make` or MSYS2 `make`
- ensure Git for Windows `bash.exe` is on `PATH`
- do not use `GnuWin32` make for this repo

## Setup
Before running bootstrap:
- Required: GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Recommended: `mise` or `asdf` for automatic tool installation from `.tool-versions`
- Fallback: manually install the pinned tool versions listed above

Run the bootstrap command from the repository root:
- Make: `make bootstrap`

Bootstrap validates the pinned Go toolchain and runs `go mod download`.
If `mise` or `asdf` is available, the script will use it to install the pinned toolchain automatically.

## Lint and Format
- Install git hooks: `make precommit-install`
- Run all pre-commit checks manually: `make precommit-run`
- Run repo lint checks: `make lint`
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

## Run
For native API work:

- Start support services from this repo: `make support-up`
- Run the API locally: `make run`
- Stop support services: `make support-down`

The placeholder API serves on `http://localhost:8080`.
Support services come from the centralized compose stack in `../platform-infra`.
After code changes, rerun `make run` to restart the native API process.

## Container
- Build placeholder image: `docker build -t backend-api:local .`
- The image packages a minimal placeholder HTTP server for the Docker baseline
- Placeholder routes currently serve `/` and `/healthz` only, pending the real API skeleton in Phase 2

## Local Stack

- API-focused mode:
  - run `make support-up`
  - run `make run`
  - compose provides `frontend-web` on `http://localhost:3000` and Postgres on `localhost:5432`
- Frontend-focused mode is orchestrated from `frontend-web`, where compose provides the containerized API on `http://localhost:8080`

## Test
No automated test suite is configured yet.
Linting, formatting, and test commands will be introduced incrementally in later tasks.
