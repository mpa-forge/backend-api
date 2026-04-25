# database-migration-baseline Specification

## Purpose

Define the canonical database migration and persistence baseline for
`backend-api`, including repo-local migration commands, typed query generation,
and the `user_profiles` identity contract.

## Requirements

### Requirement: Repo-local migration workflow

The repository SHALL provide repo-local database lifecycle commands for the API
service. At minimum, `make migrate-up`, `make migrate-down`, `make db-seed`,
`make db-prepare`, and `make sqlc-generate` MUST be available from the
repository root when either local `DATABASE_URL` or split `DB_HOST`, `DB_NAME`,
`DB_USER`, and `DB_PASSWORD` inputs are configured.

#### Scenario: Operators can apply the schema baseline

- **WHEN** an operator runs `make migrate-up`
- **THEN** the repository applies the embedded Postgres schema migrations using
  the resolved database URL from either local `DATABASE_URL` or the split
  cloud-runtime database contract

#### Scenario: Developers can prepare a local database baseline

- **WHEN** a developer runs `make db-prepare`
- **THEN** the repository applies schema migrations and the deterministic seed
  set in one workflow

### Requirement: Migration assets remain repo-local and typed

The API service SHALL keep migration, seed, handwritten query, generated sqlc,
and migration entrypoint assets inside the repository under
`internal/database/migrations/`, `internal/database/seeds/`,
`internal/database/queries/`, `internal/database/sqlc/`, and `cmd/migrate/`.

#### Scenario: Typed query generation stays aligned with repo-local SQL

- **WHEN** a developer runs `make sqlc-generate`
- **THEN** typed accessors are generated from the repository's handwritten SQL
  sources

### Requirement: Baseline profile persistence contract

The baseline schema SHALL include `user_profiles`, and
`user_profiles.clerk_user_id` SHALL be the canonical persistence key for the
verified Clerk `sub`. The deterministic seed set SHALL insert one `user` row
and one `admin` row, and repeated seed runs MUST remain safe through conflict-
aware writes.

#### Scenario: Baseline seed data is deterministic

- **WHEN** the seed workflow is run multiple times
- **THEN** the baseline user and admin rows are preserved without creating
  duplicate logical records

#### Scenario: Authenticated profile reads use Clerk subject identity

- **WHEN** the API reads a local profile for an authenticated user
- **THEN** it uses `user_profiles.clerk_user_id` derived from the verified Clerk
  `sub`

### Requirement: Runtime startup does not auto-apply schema changes

The Phase 2 API runtime SHALL require explicit migration or prepare commands and
MUST NOT auto-apply schema migrations during normal startup.

#### Scenario: API startup remains separate from migration execution

- **WHEN** the API process starts normally
- **THEN** schema migration behavior is not triggered implicitly by runtime
  startup
