---
name: openspec-pr-prep
description: Prepare pull request summaries that combine synced OpenSpec artifacts with code or documentation changes. Use when packaging a branch, staged diff, or working tree into a PR description, especially if `openspec/specs/...`, `openspec/changes/...`, archived change records, and non-OpenSpec files changed together.
---

# OpenSpec PR Prep

Create a paste-ready PR summary for cross-cutting changes that mix OpenSpec sync
work with normal implementation or documentation updates.

## Workflow

1. Pick the diff scope that matches the PR:
   - default: current worktree versus `HEAD`
   - `--staged`: staged changes only
   - `--base <ref>`: branch summary from `<ref>...HEAD`
   - `--base <ref> --head <ref>`: explicit compare range
2. Run the helper script:

   ```powershell
   python .codex/skills/openspec-pr-prep/scripts/build_pr_summary.py
   python .codex/skills/openspec-pr-prep/scripts/build_pr_summary.py --staged
   python .codex/skills/openspec-pr-prep/scripts/build_pr_summary.py --base origin/main
   python .codex/skills/openspec-pr-prep/scripts/build_pr_summary.py --base origin/main --head HEAD
   ```

3. Read the generated markdown and tighten the lead sentence so it describes the
   outcome, not just the file movement.
4. Keep the automatically-detected OpenSpec bullets unless they are wrong.
   Reviewers benefit from seeing:
   - archived change names
   - active change names
   - canonical specs updated under `openspec/specs/...`
   - non-OpenSpec areas touched in the same PR
5. Add validation notes separately if checks were run. The script does not infer
   validation.

## Output Rules

- Prefer one short summary section plus two detail sections:
  - OpenSpec changes
  - code/docs/ops changes
- Mention synced canonical specs explicitly when an archived change and
  `openspec/specs/...` updates land together.
- Keep file lists grouped by top-level area so the PR stays scannable.
- Remove noisy bullets if a path does not help a reviewer understand the change.
- If there are no OpenSpec paths in scope, use a more general PR-writing skill
  instead of forcing OpenSpec framing.

## Helper Script

Use `scripts/build_pr_summary.py` for the first draft. The script:

- inspects the selected diff scope with `git`
- groups OpenSpec paths into active changes, archived changes, canonical specs,
  and other OpenSpec files
- groups non-OpenSpec paths by top-level area
- prints markdown you can paste into a PR description or progress update

If the output is too generic for a specific repo, improve the lead paragraph
manually instead of expanding the script with repo-specific prose.
