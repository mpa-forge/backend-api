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
- GNU Make (or a compatible make implementation)
- Go `1.24.12`
- Version pin source: `.tool-versions` and `go.mod`

## Setup
Run one of the following bootstrap commands from the repository root:
- Make: `make bootstrap`

Bootstrap validates the pinned Go toolchain and runs `go mod download`.
If `mise` or `asdf` is available, the script will use it to install the pinned toolchain automatically.

## Run
No runnable API entrypoint exists yet.
Service bootstrap and local run commands will be added in later Phase 1 tasks.

## Test
No automated test suite is configured yet.
Linting, formatting, and test commands will be introduced in `P1-T04` and later tasks.


