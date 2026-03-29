# API Runtime Paths: Cloud Run Baseline and GKE Alternative

## Purpose

Define the two supported API runtime paths and how Terraform/CI select between them.

## Runtime Paths

- Path A (`cloud_run`, baseline):
  - API runs as Cloud Run service revisions.
  - Scale-to-zero enabled for low-traffic periods.
  - No GKE cluster required for first iteration.
- Path B (`gke`, alternative):
  - API runs in GKE Autopilot with Helm deployment.
  - Enabled later if product needs cluster capabilities.

## Selection Contract

- Runtime selection must be explicit in environment config.
- Suggested selector:
  - `API_RUNTIME_PATH=cloud_run|gke`
- First iteration default:
  - `API_RUNTIME_PATH=cloud_run`

## Terraform Requirements

- Keep both modules in `platform-infra`:
  - `cloudrun_api` (enabled by default)
  - `gke` (disabled by default)
- Use enable flags so either path can be created/removed without changing root structure.
- Keep shared dependencies reusable:
  - VPC/private access where required
  - Cloud SQL
  - GAR
  - GSM/IAM

## CI/CD Requirements

- Cloud Run baseline pipeline:
  - build/push image
  - deploy new Cloud Run revision
  - run smoke tests against `/api/*`
- GKE alternative pipeline:
  - build/push image
  - Helm upgrade/install
  - run equivalent smoke tests

## Observability Requirements

- Cloud Run path:
  - direct OTLP/HTTP export to Grafana Cloud.
- GKE path:
  - collector/alloy gateway path.
- Both must use the same shared observability library contract and `OBS_TELEMETRY_PROFILE`.

## Switching Guidance

- Runtime switch must be runbook-driven and reversible:
  - update Terraform runtime selector/enable flags
  - apply routing updates for `/api/*`
  - confirm secret wiring and DB connectivity
  - confirm observability mode/profile
  - execute smoke tests and rollback checks
