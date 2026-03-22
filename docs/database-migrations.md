# Database Migrations

`backend-api` uses `golang-migrate` with repo-local embedded SQL files for the
Phase 2 schema baseline.

## Current Scope

The baseline migration creates:

- `user_profiles`

The baseline deterministic seed set inserts:

- one `user` profile row
- one `admin` profile row

`user_profiles.external_user_id` is the canonical persistence key for the
authenticated user and maps directly to the verified Clerk `sub` claim.

The seed set is intentionally generic. It establishes a repeatable database
baseline for local and CI validation without pretending that the placeholder
seed users are real Clerk identities.

## Commands

Run from the repository root with `DATABASE_URL` set:

- apply schema migrations:
  - `make migrate-up`
- roll back all schema migrations:
  - `make migrate-down`
- apply deterministic seed data:
  - `make db-seed`
- apply schema + seed baseline:
  - `make db-prepare`

For local development:

1. start shared support services:
   - `make support-up`
2. apply schema + seed baseline:
   - `make db-prepare`

## File Layout

- `internal/database/migrations/`
- `internal/database/seeds/`
- `cmd/migrate/`

## Notes

- migrations are idempotent at the schema versioning level
- seeds are deterministic and re-runnable through `ON CONFLICT`
- runtime startup does not auto-apply migrations in this phase
