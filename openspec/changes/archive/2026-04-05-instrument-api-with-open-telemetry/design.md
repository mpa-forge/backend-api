## Context

`backend-api` already initializes the shared `platform-observability`
`backendobs` runtime during startup and logs the selected telemetry mode and
profile. The sibling `platform-observability` repository is explicit that this
shared package owns exporter/provider bootstrap, resource labels, shutdown
behavior, and now the canonical `backend-http-observability` capability for
reusable inbound HTTP observability helpers.

Inside this repository, the current router stack uses `chi` middleware plus
Connect handlers and a custom auth interceptor. Request logging records
`request_id`, method, raw path, status, bytes, and duration, but there is not
yet a canonical requirement or design for server spans, HTTP metrics, request
context propagation through auth and handlers, or trace/log correlation. Phase
3 planning expects that gap to be closed without changing the runtime-mode
contract: Cloud Run uses `direct_otlp`, GKE uses `collector_gateway`, and the
API code should not fork by mode.

## Goals / Non-Goals

**Goals:**

- Emit request-level API telemetry for public endpoints and protected Connect
  procedures when telemetry is enabled.
- Preserve one request context across `chi` middleware, auth interception, and
  service handlers so traces, metrics, and logs stay correlated.
- Use low-cardinality route and procedure attributes that align with the
  telemetry budget profile rules.
- Keep instrumentation call sites mode-independent and continue to consume the
  shared backend runtime plus the canonical shared HTTP helpers for exporter,
  resource, and base request instrumentation setup.
- Add focused tests and lightweight doc updates so the behavior is verifiable.

**Non-Goals:**

- Reworking Grafana secret delivery, Cloud Run/GKE infra, dashboards, or alert
  rules
- Worker or frontend observability work
- Redesigning auth, database access, or API surface behavior beyond telemetry
- Turning `platform-observability` into a full generic middleware framework in
  this change

## Decisions

### Reuse the shared backend runtime and shared HTTP helpers, then keep only API-specific enrichment here

`backendobs` remains the source of truth for providers, exporters, propagators,
resource attributes, mode/profile startup rules, and the shared
`backend-http-observability` helper contract. `backend-api` should only own the
integration details that are specific to its `chi` router, Connect procedures,
and auth flow.

For this consumer, the expected shared API surface is:

- `Runtime.Middleware(...)` for inbound HTTP server spans and base request
  metrics
- `Runtime.StartSpan(...)` for any API-local child spans that survive the first
  pass
- `Runtime.Count(...)` and `Runtime.ObserveDuration(...)` for simple
  API-specific counters or durations if they are needed
- `Runtime.CorrelationAttrs(...)` for request completion logs
- `Runtime.Inject(...)` for any explicit outbound propagation points that arise
  later

Inbound HTTP extraction is expected to happen inside the shared
`Runtime.Middleware(...)` path, so `backend-api` should not need to call
`Runtime.Extract(...)` directly during the normal request flow.

Alternative considered: implement all request instrumentation directly in
`backend-api` first and extract later. This was rejected because the reusable
code belongs in `platform-observability`, and this is the cleaner long-term
boundary.

### Create a single inbound server span at the HTTP layer and enrich it in Connect/auth code

The router layer is still the best place to attach the canonical shared inbound
HTTP middleware because it sees every public and protected request, the final
HTTP status, and the route template. `backend-api` should adapt the shared HTTP
helper layer there, while Connect and auth code enrich the active span and
request context with procedure identity and failure semantics rather than
creating a second server span.

Alternative considered: independent HTTP and RPC server spans for the same
request. This was rejected because it would duplicate latency accounting and
make traces harder to read.

### Normalize telemetry dimensions around route templates and Connect procedure constants

Metrics and span attributes should use low-cardinality values such as service,
environment, HTTP method, route template, Connect procedure, and status code or
status class. Raw URL paths, user identifiers, request IDs, and bearer-token
claims must stay out of metric labels.

Alternative considered: logging or exporting raw request paths for convenience.
This was rejected because it conflicts with the telemetry budget profile's
cardinality rules and would age poorly once dynamic routes expand.

### Correlate structured request logs with trace context without making logs a second source of truth

The existing request logger remains the main place for human-readable request
completion logs, but it should be updated to add trace and span correlation
fields when telemetry is active. This keeps local troubleshooting and Grafana
queries aligned while avoiding duplicate business logging patterns.

Alternative considered: rely only on exported OTEL logs or trace backends for
correlation. This was rejected because the repo already uses `slog` JSON logs
for runtime diagnostics and operators still need correlation in stdout logs.

### Validate request-level observability with local tests instead of Grafana-dependent checks

Implementation should use focused handler or middleware tests and OTEL test
providers or in-memory assertions to verify propagation, error mapping, and
disabled-mode safety. Grafana Cloud evidence belongs to later environment
validation work, not to repo-local unit tests.

Alternative considered: only validate by running the service against Grafana
Cloud. This was rejected because it would make the change hard to verify in CI
and on developer machines.

## Risks / Trade-offs

- [Shared helper APIs do not cover every Connect-specific detail] -> Keep only
  procedure and auth enrichment in `backend-api` and feed any clearly reusable
  pieces back into `platform-observability`.
- [Double instrumentation or duplicate metrics] -> Make the HTTP middleware the
  only creator of inbound server spans and treat Connect/auth hooks as
  enrichment only.
- [High-cardinality labels leak into metrics] -> Restrict exported dimensions to
  route or procedure templates, methods, status classes, and stable service or
  environment labels.
- [Some logs miss correlation fields] -> Centralize request completion logging
  and derive trace or span fields from the active request context in one place.

## Migration Plan

1. Consume the released `backend-http-observability` helper surface from
   `platform-observability`.
2. Integrate those helpers behind the existing telemetry runtime contract so
   `OTEL_MODE=disabled` remains safe and local development can opt out.
3. Verify public endpoints and protected Connect procedures locally with focused
   tests and manual smoke checks.
4. Roll the change through `rc` using the existing direct OTLP path and confirm
   correlated API telemetry reaches Grafana Cloud.
5. If signal quality or cost is unacceptable, roll back by reverting the change
   or setting `OTEL_MODE=disabled` until the issue is corrected.

## Open Questions

- Confirm during implementation whether any Connect procedure enrichment beyond
  the current `backendobs` facade belongs in `platform-observability` later or
  remains a `backend-api` adapter concern.
