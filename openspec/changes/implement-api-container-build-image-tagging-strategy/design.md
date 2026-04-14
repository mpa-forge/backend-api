## Context

`backend-api` currently ships a repo-local Dockerfile and a single reusable CI
workflow entrypoint, but the repository does not yet define how container
artifacts are built, tagged, and published for deployment use. Phase 4 task
P4-T04 requires deterministic API image outputs that align with the shared GAR
naming convention and support both merge-based traceability and semver-based
release references.

The design must fit the current repository shape:

- GitHub Actions is the CI entrypoint.
- The Docker build context is the repository root and packages `cmd/api`.
- Shared naming guidance requires immutable `sha-<git_sha_12>` tags and allows
  optional `v<semver>` tags.
- Follow-on tasks will handle GAR permissions and Workload Identity Federation,
  so this change should define the publishing contract without assuming static
  credentials.

## Goals / Non-Goals

**Goals:**

- Define one canonical API image build flow for CI.
- Make merge builds publish immutable tags that map directly to the source
  commit.
- Support semver release tags without changing the underlying image content.
- Capture enough verification behavior that later implementation can prove tag
  reproducibility.

**Non-Goals:**

- Building or publishing `backend-worker` images.
- Redesigning the existing Dockerfile beyond what is required for deterministic
  CI builds.
- Finalizing GAR IAM, Workload Identity Federation, or deployment rollout
  configuration.
- Adding vulnerability scanning or release orchestration outside image build and
  tagging behavior.

## Decisions

### Use one canonical Docker build definition

CI will build from the repository root Dockerfile so local and CI image outputs
stay aligned. This avoids drift between local debugging and merge artifacts and
keeps the build surface anchored to the existing API packaging path.

Alternative considered: a dedicated CI-only Dockerfile.
This was rejected because it would duplicate build logic and create avoidable
divergence between local and CI images.

### Separate verification builds from publish events

Pull requests should prove the Docker build remains healthy, but publishing
should happen only on merge-oriented events and semver release tags. That keeps
review workflows safe while still guaranteeing that every merged revision has a
published immutable image reference.

Alternative considered: publish preview images for every pull request.
This was rejected for the baseline because it increases registry churn and is
not required by P4-T04.

### Tag every published image with immutable commit identity first

Every published API image will receive a required `sha-<git_sha_12>` tag. When
the triggering ref is a semver release tag, CI will additionally publish the
matching `v<semver>` tag to the same image digest. Deployments and automation
must treat the SHA tag or digest as canonical, while the semver tag remains a
human-friendly release alias.

Alternative considered: publishing `latest` or branch-name tags.
This was rejected because mutable tags weaken traceability and conflict with the
shared naming standard.

### Verify tag aliasing by digest

When CI emits more than one tag for the same build, the workflow should confirm
those tags resolve to the same pushed manifest digest. This catches accidental
double-build drift and preserves the promise that semver tags alias the exact
immutable merge artifact.

Alternative considered: trust the build-and-push tool without explicit digest
verification.
This was rejected because the change goal is reproducible artifacts, not only
successful pushes.

## Risks / Trade-offs

- [Registry auth is owned by follow-up tasks] -> Keep this change scoped to the
  publishing contract and implementation hooks, while letting P4-T05/P4-T06
  provide the final credential path.
- [Current Dockerfile hardcodes `linux/amd64`] -> Document that platform choice
  as the initial baseline so later multi-arch work can be an explicit change.
- [Reusable workflow changes may be needed outside this repo] -> Keep the spec
  centered on `backend-api` behavior and allow implementation to choose whether
  repo-local glue or shared workflow updates are the cleanest delivery path.

## Migration Plan

1. Extend the merge pipeline so `main` pushes build and publish the API image.
2. Add semver-tag handling that reuses the same build/tag logic.
3. Expose the published image URI and digest in workflow logs or outputs for
   downstream deployment consumers.
4. Validate that the merge and release tags resolve to the expected digest
   before making the workflow the canonical publish path.

Rollback is straightforward: disable the publish step while retaining the
non-publishing build verification job.

## Open Questions

None. The first implementation will extend the shared reusable workflow path
used by `backend-api`, and the rollout will stay minimal by deferring optional
OCI metadata labels until a follow-up change requires them.
