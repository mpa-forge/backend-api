## ADDED Requirements

### Requirement: Pull requests verify the canonical API container build

The `backend-api` repository SHALL run a container build verification step in
CI for pull requests using the repository-root Dockerfile and build context that
package `cmd/api`.

#### Scenario: Pull request validates the API image build

- **WHEN** a pull request triggers the repository CI workflow
- **THEN** CI builds the API container image from the canonical Dockerfile
- **AND** the build completes without publishing a registry tag

### Requirement: Merge builds publish immutable API image tags

The `backend-api` merge pipeline SHALL publish the API container image to the
configured GAR repository using the shared naming pattern
`${REGION}-docker.pkg.dev/${PROJECT_ID}/${GAR_REPOSITORY}/backend-api:${TAG}`,
and every published merge artifact MUST include an immutable
`sha-<git_sha_12>` tag derived from the source commit. The pipeline MUST NOT
publish a mutable `latest` tag for deployment use.

#### Scenario: Main merge publishes a deterministic SHA tag

- **WHEN** code is merged and the publish pipeline runs for the resulting
  commit
- **THEN** the API image is pushed with the tag `sha-<git_sha_12>` for that
  commit
- **AND** no `latest` deployment tag is emitted

### Requirement: Release tags alias the immutable merge artifact

The publish pipeline SHALL, when running for a semver Git ref named
`v<semver>`, publish the matching semver tag in addition to the required
immutable SHA tag, and both tags MUST reference the same pushed image digest.

#### Scenario: Semver release publishes a release alias

- **WHEN** CI runs for a Git tag that matches `v<semver>`
- **THEN** the API image is published with both `sha-<git_sha_12>` and
  `v<semver>` tags
- **AND** the published tags resolve to the same image digest

### Requirement: Published image metadata is traceable in CI output

The publish workflow SHALL expose the final published image reference and digest
in CI output so downstream automation and operators can map a merge or release
event to the exact container artifact.

#### Scenario: Publish job reports the pushed artifact identity

- **WHEN** the publish workflow completes successfully
- **THEN** the job output or logs include the published `backend-api` image URI
- **AND** the output includes the pushed manifest digest for that artifact
