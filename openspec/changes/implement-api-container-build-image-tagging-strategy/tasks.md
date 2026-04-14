## 1. Establish the image publish workflow contract

- [x] 1.1 Inspect the current `backend-api` CI workflow and decide whether the
  image build logic should live in this repo or in the shared reusable workflow
  layer.
- [x] 1.2 Add CI logic that builds the canonical API Docker image on pull
  requests without publishing it.
- [x] 1.3 Add merge and semver-trigger handling that computes the required
  `sha-<git_sha_12>` tag and optional `v<semver>` tag for `backend-api`.

## 2. Publish deterministic API image artifacts

- [x] 2.1 Configure the publish job to push `backend-api` images to the
  platform GAR path using the shared naming convention and without emitting a
  `latest` tag.
- [x] 2.2 Add workflow verification that semver and SHA tags point to the same
  pushed image digest when both tags are published.
- [x] 2.3 Expose the final published image URI and digest in workflow outputs or
  logs for downstream deployment traceability.

## 3. Validate and document the rollout

- [x] 3.1 Update repo-facing CI or container documentation to describe the
  canonical image tags and publish behavior for `backend-api`.
- [x] 3.2 Run the repo validation needed for the workflow changes and capture
  any follow-up dependency on GAR auth or WIF setup required by later Phase 4
  tasks.
