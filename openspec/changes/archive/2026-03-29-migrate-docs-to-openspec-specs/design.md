## Context

`backend-api` currently keeps runtime behavior, auth details, database migration
guidance, and deployment-path guidance in `docs/`. Those documents are useful
for human readers, but they are not represented as OpenSpec capabilities, which
means later requirement changes cannot build on them through the normal
proposal/spec/task flow.

This migration needs to preserve two things at once:

- the current repository entry points in `README.md`, `AGENTS.md`, and local
  docs references
- a clean OpenSpec capability structure that future changes can modify without
  re-parsing narrative documentation

## Goals / Non-Goals

**Goals:**

- Capture the current `/docs` behavior as OpenSpec capability specs.
- Split the migrated content into capability-sized specs instead of a single
  oversized documentation dump.
- Keep the existing `docs/` paths as compatibility entry points that redirect
  readers to the canonical specs.
- End the change with canonical specs under `openspec/specs/`, not only delta
  specs inside an active change.

**Non-Goals:**

- Changing runtime, auth, database, or deployment behavior.
- Rewriting unrelated README or planning documentation.
- Expanding the documented scope beyond the current `/docs` content.

## Decisions

### Decision: Map one capability per current documentation theme

The migration uses four capabilities:

- `api-runtime`
- `api-authentication`
- `database-migration-baseline`
- `api-runtime-path-selection`

This keeps each spec narrow enough to evolve independently while preserving the
subject boundaries that already exist in `docs/`.

Alternative considered:

- One umbrella `backend-api-docs` capability. Rejected because it would create a
  large mixed spec that is difficult to change incrementally.

### Decision: Convert narrative docs into normative requirements, not prose copies

The migrated specs restate the documented behavior as SHALL-based requirements
with concrete scenarios. Commands, file paths, and behavior that remain useful
for quick human navigation stay in compatibility docs, but the canonical
behavior contract moves into OpenSpec.

Alternative considered:

- Copy the markdown content almost verbatim into spec files. Rejected because it
  would preserve prose but not create testable or reviewable requirements.

### Decision: Keep `docs/` files as compatibility shims

Repo-local agent instructions and the README still reference `docs/` paths
directly. Removing those files would create broken guidance immediately, so each
existing doc path will remain with a short summary and a pointer to the new
canonical spec.

Alternative considered:

- Delete `docs/` entirely and update every reference in the same change.
  Rejected because it adds avoidable churn and weakens backward compatibility
  for humans and tooling that still expect those paths.

### Decision: Archive the completed change to sync canonical specs

OpenSpec keeps change deltas under `openspec/changes/`, but the migration is
not complete until the canonical capability specs exist under `openspec/specs/`.
After implementation and validation, the change will be archived so OpenSpec can
update the main specs.

Alternative considered:

- Stop after generating change artifacts. Rejected because it leaves the repo in
  a transitional state where the migration exists only as a pending proposal.

## Risks / Trade-offs

- Losing useful procedural context while converting docs to requirements ->
  compatibility docs keep the operator-facing shortcuts and point to the specs.
- Future contributors may update the compatibility docs but forget the specs ->
  the compatibility docs will explicitly mark OpenSpec as the canonical source.
- Capability boundaries may still overlap at auth/database edges -> the specs
  keep runtime auth enforcement in `api-authentication` and persistence workflow
  guarantees in `database-migration-baseline` to minimize ambiguity.
- Archive-time sync could fail if delta specs are malformed -> validate the
  change before archive and keep requirements/scenarios in strict OpenSpec
  format.

## Migration Plan

1. Create the proposal, design, and capability delta specs for the four current
   documentation themes.
2. Add implementation tasks covering spec authoring, compatibility-doc updates,
   and validation/archive.
3. Apply the tasks by updating the `docs/` files to point at the new canonical
   specs.
4. Validate the change with `openspec validate`.
5. Archive the completed change so the canonical specs are written into
   `openspec/specs/`.

## Open Questions

- None.
