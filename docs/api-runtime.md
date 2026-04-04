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
- telemetry-enabled runtimes require:
  - `OTEL_EXPORTER_OTLP_ENDPOINT`
  - `GRAFANA_CLOUD_INSTANCE_ID`
  - `GRAFANA_OTLP_INGEST_TOKEN`
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

## Update Rule

When runtime behavior changes, update the OpenSpec capability first and keep
this file as a lightweight reader-friendly summary.
