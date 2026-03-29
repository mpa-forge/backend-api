# api-runtime-path-selection Specification

## Purpose

Define the canonical runtime-path selection contract for `backend-api`,
including the Cloud Run baseline, the GKE alternative, and the shared
deployment and observability expectations across both paths.

## Requirements

### Requirement: Runtime path selection is explicit

The platform SHALL support two API runtime paths, `cloud_run` and `gke`, and
runtime-path selection MUST be represented explicitly in environment
configuration through `API_RUNTIME_PATH=cloud_run|gke`. The first-iteration
default SHALL be `cloud_run`.

#### Scenario: Cloud Run remains the baseline path

- **WHEN** the platform uses the default runtime-path selection
- **THEN** the API is targeted for the `cloud_run` path

#### Scenario: Runtime-path changes are explicit

- **WHEN** an environment chooses the GKE alternative
- **THEN** the configuration sets `API_RUNTIME_PATH=gke` rather than relying on
  implicit infrastructure behavior

### Requirement: Infrastructure keeps both runtime paths available

`platform-infra` SHALL keep both runtime-path modules available with enable
flags so Cloud Run and GKE can be created or removed without restructuring the
root Terraform layout. Shared dependencies such as networking, Cloud SQL,
artifact registry, and secret or IAM wiring MUST remain reusable across both
paths.

#### Scenario: Cloud Run module is enabled by default

- **WHEN** the first-iteration baseline infrastructure is applied
- **THEN** the Cloud Run API module is enabled and the GKE path can remain
  disabled without deleting its module definition

### Requirement: Delivery and observability behavior stays path-aware

Delivery pipelines SHALL provide equivalent API smoke-test coverage for both
runtime paths after deployment. The Cloud Run path SHALL use direct OTLP export,
the GKE path SHALL use the collector or alloy gateway path, and both paths
SHALL use the same shared observability library contract and
`OBS_TELEMETRY_PROFILE`.

#### Scenario: Cloud Run deployments use direct OTLP

- **WHEN** the API is deployed on the Cloud Run path
- **THEN** telemetry is configured for direct OTLP export using the shared
  observability contract

#### Scenario: GKE deployments use collector gateway telemetry

- **WHEN** the API is deployed on the GKE path
- **THEN** telemetry is configured through the collector or alloy gateway using
  the shared observability contract

### Requirement: Runtime-path switches are runbook-driven and reversible

Changing the selected runtime path SHALL be performed through an explicit
runbook that updates runtime selection, routing, secrets, database
connectivity, observability settings, smoke tests, and rollback checks.

#### Scenario: Runtime-path switch verifies operational dependencies

- **WHEN** operators switch the API from one supported runtime path to the other
- **THEN** they verify routing, secret wiring, database connectivity,
  observability mode, smoke tests, and rollback readiness as part of the
  change
