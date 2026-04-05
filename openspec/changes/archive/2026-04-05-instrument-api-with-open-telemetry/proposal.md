## Why

`backend-api` now boots through the shared `platform-observability` runtime,
and the sibling repository now carries the canonical
`backend-http-observability` capability for reusable HTTP observability
helpers. Phase 3 task `P3-T03` stays open until `backend-api` consumes that
shared helper surface and produces correlated API telemetry in Grafana Cloud
across the supported runtime modes without introducing a competing repo-local
observability abstraction.

## What Changes

- Define a new `api-observability` capability for request-level API telemetry
  consumption in `backend-api`.
- Consume reusable request instrumentation helpers from
  `platform-observability`'s canonical `backend-http-observability`
  capability so the `chi` + Connect request path emits correlated server spans,
  HTTP metrics, and request-scoped log context through shared observability
  code.
- Adopt the shared `backendobs` facade methods that now exist for inbound HTTP
  middleware, child spans, simple metric recording, and trace-correlation
  fields instead of inventing local equivalents.
- Normalize route and procedure attributes, error mapping, and correlation
  fields so `direct_otlp` and `collector_gateway` behave the same from the API
  code's point of view and keep telemetry cardinality under control.
- Add focused validation coverage and lightweight runtime-documentation updates
  so operators and developers can verify the new behavior locally and in `rc`.

## Capabilities

### New Capabilities

- `api-observability`: request-level telemetry behavior for `backend-api`,
  including traces, metrics, propagation, and correlation rules for public and
  protected API flows.

### Modified Capabilities

None.

## Impact

- API runtime wiring in `cmd/api` and `internal/api`
- Connect/auth request flow in `internal/auth`
- Runtime validation and observability-focused tests
- Compatibility docs such as `docs/api-runtime.md`
- Continued consumption of the shared `platform-observability` package for
  runtime/exporter ownership and request-level helper APIs
- Alignment with `../platform-observability/openspec/specs/backend-http-observability/spec.md`
