# API Runtime

This file is a compatibility entry point. The canonical runtime requirements now
live in `openspec/specs/api-runtime/spec.md`.

## OpenSpec Capability

- `api-runtime`

## Quick Reference

- startup is fail-fast and validates required environment variables before the
  API binds a port
- `OTEL_MODE` accepts `disabled`, `direct_otlp`, or `collector_gateway`
- `OBS_TELEMETRY_PROFILE` accepts `balanced`, `cost`, or `debug` and defaults
  to `balanced`
- startup now initializes the shared backend observability runtime from
  `github.com/mpa-forge/platform-observability/backendobs`
- inbound HTTP requests now go through `backendobs.Runtime.Middleware(...)` so
  public routes and protected Connect procedures share one server span and the
  baseline HTTP request metrics path
- protected Connect requests enrich the active request span with the Connect
  procedure, auth result, and Connect failure code when a request fails
- request completion logs now add `trace_id`, `span_id`, and `trace_sampled`
  when the active request is running inside a traced context
- startup diagnostics now include the resolved trace sample ratio, the initial
  high-latency force-sample threshold, and whether the `cost` profile reduces
  successful request-duration metrics
- telemetry-enabled runtimes require:
  - `OTEL_EXPORTER_OTLP_ENDPOINT`
  - `GRAFANA_CLOUD_INSTANCE_ID`
  - `GRAFANA_OTLP_INGEST_TOKEN`
- database configuration accepts either local `DATABASE_URL` or split
  `DB_HOST`, `DB_NAME`, `DB_USER`, and `DB_PASSWORD`; cloud runtimes should
  leave `DATABASE_URL` unset and the split contract wins whenever those values
  are populated; `DB_HOST` may be a Cloud SQL socket path under `/cloudsql`
- the runtime composes the OTLP Basic auth header from the Grafana instance ID
  and ingest token instead of requiring `OTEL_EXPORTER_OTLP_HEADERS`
- `make run` sources `.env` when present and starts `./cmd/api`
- public endpoints remain `GET /`, `GET /healthz`, and `GET /readyz`
- protected Connect procedures remain:
  - `POST /blueprint.user.v1.UserService/EnsureCurrentUserProfile`
  - `POST /blueprint.user.v1.UserService/GetCurrentUser`
- in `local`, the runtime now allows browser CORS requests from
  `http://localhost:3000` and `http://127.0.0.1:3000` so the frontend SPA can
  call the protected Connect procedures directly

## Local Verification

- the current workspace resolves `github.com/mpa-forge/platform-observability`
  through the sibling `../platform-observability` checkout so `backend-api`
  consumes the latest shared `backendobs` helper methods during implementation
- run `go test ./internal/api ./internal/auth` to verify:
  - disabled-mode request handling
  - OTLP export of public endpoint spans and metrics
  - OTLP export of protected Connect procedure metadata and failure mapping
  - resource labels for service identity, runtime mode, and telemetry profile
  - trace correlation fields on request completion logs

## Update Rule

When runtime behavior changes, update the OpenSpec capability first and keep
this file as a lightweight reader-friendly summary.
