# Agent Context

## Local Entry Point

This file is the repo-local entry point for agent context.

## Always Load

Before making changes:

1. Read `README.md`.
2. Read `Makefile` if present.
3. Run `make sync-agent-skills` before starting major changes or when shared skill guidance may have changed.
4. Read `../platform-blueprint-specs/common/AGENTS.md`.

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

- Check local repo docs under `docs/` when the task touches API runtime details.
- `docs/api-runtime-paths-cloud-run-gke.md` when the task affects runtime-path selection, rollout path, or Cloud Run vs GKE behavior.
- `../platform-blueprint-specs/platform-specification.md` only when the task needs broader platform architecture context or locked platform decisions beyond the API baseline.
- `../platform-blueprint-specs/implementation/implementation-plan.md` when the task depends on roadmap sequencing, phase ownership, or baseline MVP scope.
- `../platform-blueprint-specs/implementation/phases/phase-2-contracts-service-skeletons-and-data-baseline.md` when the task needs Phase 2 acceptance criteria, unfinished baseline alignment, or historical implementation intent.
- `../platform-blueprint-specs/implementation/phase-tasks/phase-2-contracts-service-skeletons-and-data-baseline-tasks.md` when the task needs Phase 2 task alignment, ownership, dependencies, or evidence context.
- `../platform-blueprint-specs/common/standards/environment-variable-strategy.md` when the task touches `.env.example`, environment-variable naming, startup validation, secret placeholders, or config contract design.
- `../platform-blueprint-specs/ops/observability-telemetry-budget-profile.md` when the task affects OpenTelemetry wiring, `OBS_TELEMETRY_PROFILE`, Grafana export behavior, or Cloud Run vs GKE observability-path behavior.
- `../platform-contracts/docs/go-server-usage.md` when the task touches
  generated Go contract consumption, Connect handler wiring, or server-side
  contract package usage.
- `../platform-contracts/docs/consumer-auth-usage.md` when the task needs the
  consumer-facing protected API auth contract shared with frontend callers.
- `../platform-contracts/docs/contract-release-workflow.md` when the task
  touches released contract versions, tag semantics, or how `backend-api`
  should consume versioned `platform-contracts` releases.

## Typical Validation

- `make lint`
- `make test`
- `make format-check`

## Priority of Instructions

Repo-local instructions override shared planning docs.

If local repo docs conflict with a shared planning file, the more specific repo or task instruction wins.
