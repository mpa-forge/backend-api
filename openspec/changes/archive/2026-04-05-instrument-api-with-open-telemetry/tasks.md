## 1. Request Instrumentation

- [x] 1.1 Update `backend-api` to consume the shared inbound HTTP observability
  helpers from `platform-observability`'s `backend-http-observability`
  capability for server spans, request metrics, and normalized route
  attributes.
- [x] 1.2 Enrich the protected Connect flow with procedure metadata, context
  propagation, and auth or handler failure mapping without creating duplicate
  server spans.
- [x] 1.3 Update request completion logging so it keeps the current structured
  fields and adds the shared `backendobs.CorrelationAttrs(...)` trace or span
  correlation fields when telemetry is enabled.

## 2. Runtime Wiring

- [x] 2.1 Integrate the new observability helpers into router startup so the API
  behavior stays the same across `disabled`, `direct_otlp`, and
  `collector_gateway` modes from the service code's perspective.
- [x] 2.2 Update the `platform-observability` dependency and any local adapter
  code needed to consume the shared helper APIs without duplicating startup or
  exporter wiring in `backend-api`, including the canonical helper methods that
  now exist in `backendobs`.

## 3. Validation And Docs

- [x] 3.1 Add focused tests for disabled-mode safety, public endpoint
  instrumentation, protected Connect instrumentation, and auth or error status
  mapping.
- [x] 3.2 Refresh compatibility docs and local verification notes so developers
  know how API observability behaves and how to confirm the new telemetry path.
