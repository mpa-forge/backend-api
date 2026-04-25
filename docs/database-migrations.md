# Database Migrations

This file is a compatibility entry point. The canonical migration and
persistence requirements now live in
`openspec/specs/database-migration-baseline/spec.md`.

## OpenSpec Capability

- `database-migration-baseline`

## Quick Reference

- repo-local commands remain:
  - `make migrate-up`
  - `make migrate-down`
  - `make db-seed`
  - `make db-prepare`
  - `make sqlc-generate`
- migration, seed, query, generated sqlc, and migrate entrypoint assets remain
  under `internal/database/` and `cmd/migrate/`
- migration commands select the target database from local `DATABASE_URL` when
  no split database inputs are configured, or from split `DB_HOST`, `DB_NAME`,
  `DB_USER`, and `DB_PASSWORD` values when the cloud-runtime contract is used
- the baseline schema includes `user_profiles`
- `user_profiles.clerk_user_id` remains the canonical persistence key for the
  verified Clerk `sub`
- the seed set remains deterministic and safe to re-run
- runtime startup does not auto-apply migrations in this phase

## Update Rule

When database baseline behavior changes, update the OpenSpec capability first
and keep this file as a lightweight reader-friendly summary.
