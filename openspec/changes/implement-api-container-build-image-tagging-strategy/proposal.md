## Why

The API repository already defines a runnable container image, but Phase 4
still lacks a committed CI contract for building and publishing that image with
deterministic tags. We need that contract now so merges can produce traceable
artifacts that downstream deployment automation can consume without relying on
mutable tags or ad hoc build behavior.

## What Changes

- Define the `backend-api` image publishing contract for merge pipelines.
- Require immutable commit-based image tags and optional semver-aligned release
  tags that follow the shared platform naming standard.
- Document the CI build inputs, tag calculation rules, and publication outputs
  needed for reproducible API image artifacts.
- Establish the baseline verification steps that confirm published tags point to
  the same built image content.

## Capabilities

### New Capabilities

- `api-image-publishing`: Defines how `backend-api` container images are built,
  tagged, and published from CI with immutable and release-friendly tags.

### Modified Capabilities

## Impact

- `backend-api` GitHub Actions workflows and supporting CI scripts
- Container build inputs such as the repo Dockerfile and build context
- Image consumers in deployment automation that must reference immutable tags
- Shared GAR naming and release-tag expectations used across the platform
