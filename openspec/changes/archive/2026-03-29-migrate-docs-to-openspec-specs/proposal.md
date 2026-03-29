## Why

The repository documents important runtime, auth, database, and deployment-path
behavior in `docs/`, but none of that behavior is represented in OpenSpec yet.
Migrating that guidance into specs makes the behavior reviewable, versionable,
and ready for future spec-driven changes without losing the existing doc entry
points that the repo still references.

## What Changes

- Add OpenSpec capability specs for the API runtime contract, auth enforcement,
  database migration baseline, and runtime-path selection.
- Preserve the current `docs/` file paths as lightweight compatibility guides
  that point readers to the new canonical OpenSpec specs.
- Establish the migrated specs as the normative source for future requirement
  updates in this repository.

## Capabilities

### New Capabilities

- `api-runtime`: API startup validation, local run expectations, and mounted
  HTTP and Connect endpoints.
- `api-authentication`: Clerk-backed bearer-token verification, claim mapping,
  and protected procedure authorization behavior.
- `database-migration-baseline`: Repo-local schema migration, seed, and typed
  query baseline behavior for the API service.
- `api-runtime-path-selection`: Supported Cloud Run and GKE runtime paths plus
  the explicit selection contract used by infrastructure and delivery flows.

### Modified Capabilities

- None.

## Impact

- `openspec/changes/migrate-docs-to-openspec-specs/`
- `openspec/specs/`
- `docs/api-runtime.md`
- `docs/auth-implementation.md`
- `docs/database-migrations.md`
- `docs/api-runtime-paths-cloud-run-gke.md`
