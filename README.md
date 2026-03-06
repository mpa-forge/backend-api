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

## Run
No runnable API entrypoint exists yet.
Service bootstrap and local run commands will be added in later Phase 1 tasks.

## Test
No automated test suite is configured yet.
Linting, formatting, and test commands will be introduced incrementally in later tasks.
