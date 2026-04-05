# api-observability Specification

## Purpose

Define the canonical request-level observability contract for `backend-api`,
including shared inbound HTTP instrumentation, protected Connect telemetry
enrichment, and trace-correlated request logging.

## Requirements

### Requirement: API requests emit request-level telemetry when observability is enabled

The `backend-api` runtime SHALL emit request-level telemetry for public HTTP
endpoints and protected Connect procedures whenever `OTEL_MODE` is
`direct_otlp` or `collector_gateway`. The emitted telemetry MUST include one
inbound server span and HTTP request metrics that share the same request
context and use the shared backend observability runtime plus the canonical
shared `backend-http-observability` helper contract for exporters, resource
labels, and inbound HTTP propagation.

#### Scenario: Public endpoint requests produce server telemetry

- **WHEN** the API handles a request to a public endpoint such as `GET /` or
  `GET /healthz` with telemetry enabled
- **THEN** the runtime records a server span and HTTP request metrics for that
  request using the active route template, HTTP method, and response status

#### Scenario: Disabled mode does not break request handling

- **WHEN** the API runs with `OTEL_MODE=disabled`
- **THEN** the same public and protected request paths continue to work without
  requiring exporter credentials or request-specific observability setup

### Requirement: Protected Connect procedures preserve telemetry context and failure mapping

Protected Connect procedures SHALL preserve the active request telemetry context
through auth interception and handler execution. The runtime MUST annotate the
active request telemetry with the Connect procedure identity and failure status
for authentication and handler errors without creating duplicate inbound server
spans.

#### Scenario: Authenticated Connect request keeps one telemetry context

- **WHEN** a protected Connect procedure is called with a valid bearer token
- **THEN** the auth interceptor and service handler execute within the same
  request telemetry context created for the inbound HTTP request

#### Scenario: Auth or handler failures are reflected in telemetry

- **WHEN** a protected Connect request fails because authentication is missing,
  forbidden, or the handler returns an error
- **THEN** the active request telemetry records the Connect procedure identity
  and the mapped failure status for the resulting response

### Requirement: API observability uses low-cardinality labels and log correlation

The `backend-api` runtime SHALL use normalized route and procedure identifiers
for API observability and SHALL avoid high-cardinality metric dimensions such as
raw request paths, request IDs, or principal identifiers. Request completion
logs emitted by the runtime MUST include request correlation fields that let
operators relate logs back to the active trace when telemetry is enabled,
including the correlation fields exposed by the shared `backendobs` helper
surface.

#### Scenario: Metrics use normalized route or procedure identifiers

- **WHEN** the API exports telemetry for a protected or public request
- **THEN** the exported dimensions use stable route templates or Connect
  procedure names instead of raw URL paths or unique identifiers

#### Scenario: Request logs carry trace correlation fields

- **WHEN** the runtime emits its structured request completion log for a handled
  request while telemetry is enabled
- **THEN** the log includes the request ID and active trace correlation fields
  in addition to the existing method, status, bytes, and duration data
