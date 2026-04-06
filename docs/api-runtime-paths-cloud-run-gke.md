# API Runtime Paths: Cloud Run Baseline and GKE Alternative

This file is a compatibility entry point. The canonical runtime-path selection
requirements now live in `openspec/specs/api-runtime-path-selection/spec.md`.

## OpenSpec Capability

- `api-runtime-path-selection`

## Quick Reference

- supported runtime paths remain `cloud_run` and `gke`
- runtime-path selection remains explicit through
  `API_RUNTIME_PATH=cloud_run|gke`
- the first-iteration default remains `cloud_run`
- `platform-infra` keeps both runtime-path modules available with enable flags
- Cloud Run deployments use direct OTLP telemetry
- GKE deployments use the collector or alloy gateway telemetry path
- both paths use the shared backend observability package and
  `OBS_TELEMETRY_PROFILE`
- both paths now rely on the same shared policy contract for trace sampling,
  force-sample rules, and startup diagnostics even though the collector path
  can apply extra downstream processors
- runtime-path switches remain runbook-driven and reversible

## Update Rule

When runtime-path behavior changes, update the OpenSpec capability first and
keep this file as a lightweight reader-friendly summary.
