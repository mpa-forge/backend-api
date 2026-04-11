# API Alerting

This file defines the Phase 3 alert contract owned by `backend-api`.

The repo owns service-specific signal choices, thresholds, and Prometheus-style
queries. Shared Grafana contact-point and routing policy stays in
`../platform-blueprint-specs/docs/operations/grafana-alert-routing-bootstrap-runbook.md`
so other services can follow the same severity model later without copying API
details.

## Source Artifacts

- service rule manifest: `docs/grafana-alert-rules.phase3.yaml`
- dashboard dependencies:
  - `../platform-infra/docs/grafana-dashboards/api-golden-signals.json`
  - `../platform-infra/docs/grafana-dashboards/runtime-path-status.json`
  - `../platform-infra/docs/grafana-dashboards/db-connectivity-symptoms.json`
- common routing runbook:
  - `../platform-blueprint-specs/docs/operations/grafana-alert-routing-bootstrap-runbook.md`

## Signal Boundaries

- Phase 3 availability is an application-success proxy, not full black-box
  uptime. The current source-controlled telemetry baseline does not yet publish
  a dedicated uptime-probe signal, so availability is derived from sustained 5xx
  behavior on live traffic.
- Error-rate and latency rules focus on the authenticated Connect procedures
  because those routes carry the product path that depends on Postgres and user
  profile state.
- The alert expressions assume the same translated metric names and label keys
  already used by the Phase 3 Grafana dashboards.

## Alert Catalog

- `backend-api-availability-proxy-p1`
  - severity: `P1`
  - condition: global backend-api 5xx-derived availability proxy falls below
    `99.0%` for `5m`
  - intent: catch acute service-wide failures quickly and route to the immediate
    page or webhook path
- `backend-api-error-rate-p2`
  - severity: `P2`
  - condition: authenticated Connect procedure 5xx ratio exceeds `5%` for `10m`
  - intent: highlight sustained user-facing failures without paging immediately
- `backend-api-latency-p2`
  - severity: `P2`
  - condition: authenticated Connect procedure p95 latency exceeds `1s` for
    `15m`
  - intent: catch slow degradation aligned with the provisional rc latency
    budget

Each rule also carries a minimum traffic floor so sparse rc traffic does not
trip alerts on single-request noise.

## Bootstrap Use

1. Import the dashboard JSON from `platform-infra` first so operators can verify
   the query contract visually.
2. Create matching alert rules in Grafana Cloud from the expressions in
   `docs/grafana-alert-rules.phase3.yaml`.
3. Apply the shared Phase 3 routing policy from the common runbook.
4. Record the final Grafana rule UID, contact points, and validation date in
   Phase 3 evidence once the bootstrap import is complete.

## Synthetic Validation

The current API runtime does not ship a dedicated fault-injection endpoint, so
synthetic alert validation is done through controlled rc changes and replay
traffic:

1. availability and error-rate validation:
   temporarily break the Postgres-backed user-service path in rc, then replay
   authenticated Connect traffic long enough to exceed the manifest threshold
2. latency validation:
   use a short-lived rc canary or controlled dependency slowdown that pushes the
   protected user-service path above `1s` p95 for at least `15m`
3. routing validation:
   confirm `P1` reaches Slack, the alert-to-AI webhook, and the immediate page
   or webhook destination, then confirm `P2` notifies immediately and only
   pages after `15m` without acknowledgement

## Update Rule

If API route labels, metric names, or service thresholds change, update the
rule manifest in this repo and the shared routing runbook in
`platform-blueprint-specs` in the same task.
