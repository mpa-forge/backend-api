# Agent Context

## Local Entry Point

This file is the repo-local entry point for agent context.

## Always Load

Before making changes:

1. Read `README.md`.
2. Read `Makefile` if present.
3. Read `../platform-blueprint-specs/common/AGENTS.md`.
4. Read `../platform-blueprint-specs/.codex/skills/automated-ai-worker/SKILL.md` when the repo is being changed by an automated AI worker or when following the same autonomous workflow manually.
5. Read `../platform-blueprint-specs/implementation/phases/phase-2-contracts-service-skeletons-and-data-baseline.md`.
6. Read `../platform-blueprint-specs/implementation/phase-tasks/phase-2-contracts-service-skeletons-and-data-baseline-tasks.md`.
7. Check local repo docs under `docs/` if the task touches API runtime details.
8. Read `docs/api-runtime-paths-cloud-run-gke.md` when the task affects runtime-path selection, rollout path, or Cloud Run vs GKE behavior.

## Repo Role

- Own the browser-facing Go API service.
- Serve protobuf-defined endpoints through Connect-compatible handlers.
- Use `chi` as the HTTP routing layer for the API skeleton baseline.

## Relevant Shared Constraints

- API contract model is proto-first with Connect-compatible endpoints.
- Go HTTP baseline is `chi` with `connect-go` handlers.
- Config must fail fast on missing or malformed required environment variables once runtime startup is implemented.
- Typed DB access baseline is `sqlc` with handwritten SQL and `pgx` runtime.

## Consult Conditionally

- `../platform-blueprint-specs/platform-specification.md` only when the task needs broader platform architecture context beyond the API baseline.
- `docs/api-runtime-paths-cloud-run-gke.md` when the task affects runtime-path selection, rollout path, or Cloud Run vs GKE behavior.

## Typical Validation

- `make lint`
- `make test`
- `make format-check`

## Priority of Instructions

Repo-local instructions override shared planning docs.

If local repo docs conflict with a shared planning file, the more specific repo or task instruction wins.
