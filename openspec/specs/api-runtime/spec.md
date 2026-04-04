# api-runtime Specification

## Purpose

Define the canonical runtime contract for `backend-api`, including startup
validation, local run behavior, and the mounted public and Connect endpoints.

## Requirements

### Requirement: Runtime startup validation

The `backend-api` runtime SHALL validate its required startup configuration
before binding an HTTP port. Required variables MUST include `APP_ENV`,
`LOG_LEVEL`, `HTTP_PORT`, `DATABASE_URL`, `AUTH_ISSUER_URL`,
`AUTH_AUDIENCE`, `OTEL_MODE`, `OBS_TELEMETRY_PROFILE`,
`OTEL_EXPORTER_OTLP_ENDPOINT`, `GRAFANA_CLOUD_INSTANCE_ID`, and
`GRAFANA_OTLP_INGEST_TOKEN`.
`OTEL_MODE` MUST accept only `disabled`, `direct_otlp`, or
`collector_gateway`. `OBS_TELEMETRY_PROFILE` MUST accept only `balanced`,
`cost`, or `debug`, and it MUST default to `balanced` when unset.
When telemetry is enabled through `direct_otlp` or `collector_gateway`,
the OTLP endpoint and Grafana OTLP token ingredients MUST be present and valid.

#### Scenario: Startup fails on missing required configuration

- **WHEN** the API process starts without one of the required environment
  variables
- **THEN** the runtime exits before binding the configured HTTP port

#### Scenario: Telemetry configuration is required when enabled

- **WHEN** `OTEL_MODE` is `direct_otlp` or `collector_gateway`
- **THEN** the runtime rejects startup if `OTEL_EXPORTER_OTLP_ENDPOINT`,
  `GRAFANA_CLOUD_INSTANCE_ID`, or `GRAFANA_OTLP_INGEST_TOKEN` is empty or
  malformed

### Requirement: Local run entrypoint

The repository SHALL provide a local API run flow through `make run`, and that
entrypoint MUST source `.env` when present before starting `./cmd/api`.

#### Scenario: Local run uses repo environment overrides

- **WHEN** a developer runs `make run` from the repository root with a `.env`
  file present
- **THEN** the API process starts using variables sourced from that file

### Requirement: Runtime endpoint surface

The API runtime SHALL expose metadata and health endpoints at `GET /`,
`GET /healthz`, and `GET /readyz`, and it SHALL mount the generated Connect
procedures `POST /blueprint.user.v1.UserService/EnsureCurrentUserProfile` and
`POST /blueprint.user.v1.UserService/GetCurrentUser`. In `local`, the runtime
MUST allow browser CORS requests from the documented frontend SPA origins so
the frontend can call those procedures directly.

#### Scenario: Protected Connect procedures are mounted

- **WHEN** the API runtime is started with valid configuration
- **THEN** the generated `UserService` procedures are available at their Connect
  HTTP paths

#### Scenario: Local frontend origin can preflight protected Connect procedures

- **WHEN** a browser from the documented local frontend origin sends a CORS
  preflight request to a protected Connect procedure path
- **THEN** the API runtime returns the corresponding allow-origin,
  allow-methods, and allow-headers response needed for the frontend SPA call
