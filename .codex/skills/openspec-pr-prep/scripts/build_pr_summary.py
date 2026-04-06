#!/usr/bin/env python3
"""Build a PR-ready markdown summary for mixed OpenSpec and code changes."""

from __future__ import annotations

import argparse
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
import re


@dataclass(frozen=True)
class ChangedPath:
    status: str
    path: str
    old_path: str | None = None


def run_git(repo_root: Path, *args: str) -> str:
    result = subprocess.run(
        ["git", *args],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip() or "git command failed")
    return result.stdout


def repo_root_from_cwd() -> Path:
    output = run_git(Path.cwd(), "rev-parse", "--show-toplevel").strip()
    return Path(output)


def parse_name_status(raw: str) -> list[ChangedPath]:
    changes: list[ChangedPath] = []
    for line in raw.splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        status = parts[0]
        code = status[0]
        if code == "R" and len(parts) >= 3:
            changes.append(ChangedPath(status=code, path=parts[2], old_path=parts[1]))
            continue
        if len(parts) >= 2:
            changes.append(ChangedPath(status=code, path=parts[1]))
    return changes


def parse_untracked(raw: str) -> list[ChangedPath]:
    paths = []
    for line in raw.splitlines():
        if not line.strip():
            continue
        paths.append(ChangedPath(status="A", path=line.strip()))
    return paths


def load_changed_paths(repo_root: Path, args: argparse.Namespace) -> tuple[list[ChangedPath], str]:
    if args.base or args.head:
        base = args.base or "HEAD"
        head = args.head or "HEAD"
        scope = f"{base}...{head}"
        raw = run_git(repo_root, "diff", "--name-status", "--find-renames", scope)
        return parse_name_status(raw), scope

    if args.staged:
        raw = run_git(repo_root, "diff", "--cached", "--name-status", "--find-renames", "HEAD")
        changes = parse_name_status(raw)
        if args.include_untracked:
            changes.extend(parse_untracked(run_git(repo_root, "ls-files", "--others", "--exclude-standard")))
        return dedupe_changes(changes), "staged + untracked"

    raw = run_git(repo_root, "diff", "--name-status", "--find-renames", "HEAD")
    changes = parse_name_status(raw)
    if args.include_untracked:
        changes.extend(parse_untracked(run_git(repo_root, "ls-files", "--others", "--exclude-standard")))
    return dedupe_changes(changes), "worktree + index vs HEAD"


def dedupe_changes(changes: list[ChangedPath]) -> list[ChangedPath]:
    deduped: dict[str, ChangedPath] = {}
    for change in changes:
        deduped[change.path] = change
    return list(deduped.values())


def normalize_archived_change_name(raw_name: str) -> str:
    match = re.match(r"^\d{4}-\d{2}-\d{2}-(.+)$", raw_name)
    if match:
        return match.group(1)
    return raw_name


def artifact_name(parts: list[str]) -> str:
    if parts == ["proposal.md"]:
        return "proposal"
    if parts == ["design.md"]:
        return "design"
    if parts == ["tasks.md"]:
        return "tasks"
    if len(parts) >= 3 and parts[0] == "specs" and parts[2] == "spec.md":
        return f"delta spec `{parts[1]}`"
    return "`" + "/".join(parts) + "`"


def status_suffix(change: ChangedPath) -> str:
    if change.status == "M":
        return ""
    if change.status == "A":
        return " (added)"
    if change.status == "D":
        return " (deleted)"
    if change.status == "R" and change.old_path:
        return f" (renamed from `{change.old_path}`)"
    return f" ({change.status.lower()})"


def pluralize(count: int, singular: str, plural: str | None = None) -> str:
    if count == 1:
        return f"{count} {singular}"
    return f"{count} {plural or singular + 's'}"


def summarize_changes(repo_root: Path, changes: list[ChangedPath], scope: str) -> str:
    archived_changes: dict[str, list[str]] = defaultdict(list)
    active_changes: dict[str, list[str]] = defaultdict(list)
    canonical_specs: dict[str, str] = {}
    other_openspec: list[str] = []
    non_openspec: dict[str, list[str]] = defaultdict(list)

    for change in sorted(changes, key=lambda item: item.path):
        path = change.path.replace("\\", "/")
        parts = path.split("/")
        if parts[:3] == ["openspec", "changes", "archive"] and len(parts) >= 5:
            change_name = normalize_archived_change_name(parts[3])
            archived_changes[change_name].append(artifact_name(parts[4:]) + status_suffix(change))
            continue
        if parts[:2] == ["openspec", "changes"] and len(parts) >= 4:
            change_name = parts[2]
            active_changes[change_name].append(artifact_name(parts[3:]) + status_suffix(change))
            continue
        if parts[:2] == ["openspec", "specs"] and len(parts) >= 4 and parts[3] == "spec.md":
            canonical_specs[parts[2]] = status_suffix(change)
            continue
        if parts[0] == "openspec":
            other_openspec.append(f"`{path}`{status_suffix(change)}")
            continue

        area = parts[0] if len(parts) > 1 else "repo root"
        display_path = "/".join(parts[1:]) if area != "repo root" else parts[0]
        non_openspec[area].append(f"`{display_path}`{status_suffix(change)}")

    repo_name = repo_root.name
    lines: list[str] = []
    lines.append(f"# PR Summary for `{repo_name}`")
    lines.append("")
    lines.append(f"Generated from `{scope}`.")
    lines.append("")
    lines.append("## Summary")

    summary_bits: list[str] = []
    if archived_changes:
        summary_bits.append(f"archives {pluralize(len(archived_changes), 'OpenSpec change')}")
    if active_changes:
        summary_bits.append(
            f"updates {pluralize(len(active_changes), 'active OpenSpec change')}"
        )
    if canonical_specs:
        summary_bits.append(f"syncs {pluralize(len(canonical_specs), 'canonical spec')}")
    if other_openspec:
        summary_bits.append(f"touches {pluralize(len(other_openspec), 'other OpenSpec file')}")
    if non_openspec:
        summary_bits.append(f"updates {pluralize(len(non_openspec), 'non-OpenSpec area')}")

    if summary_bits:
        lines.append("- This change set " + ", ".join(summary_bits) + ".")
    else:
        lines.append("- This change set only touches a narrow file subset.")

    if archived_changes:
        lines.append("- Archived changes: " + ", ".join(f"`{name}`" for name in archived_changes) + ".")
    if active_changes:
        lines.append("- Active changes: " + ", ".join(f"`{name}`" for name in active_changes) + ".")
    if canonical_specs:
        lines.append("- Canonical specs: " + ", ".join(f"`{name}`" for name in canonical_specs) + ".")
    if non_openspec:
        area_list = ", ".join(f"`{name}`" for name in sorted(non_openspec))
        lines.append(f"- Other touched areas: {area_list}.")

    if archived_changes or active_changes or canonical_specs or other_openspec:
        lines.append("")
        lines.append("## OpenSpec Changes")
        if archived_changes:
            lines.append("- Archived changes:")
            for change_name in sorted(archived_changes):
                detail = ", ".join(sorted(archived_changes[change_name]))
                lines.append(f"  - `{change_name}`: {detail}")
        if active_changes:
            lines.append("- Active changes:")
            for change_name in sorted(active_changes):
                detail = ", ".join(sorted(active_changes[change_name]))
                lines.append(f"  - `{change_name}`: {detail}")
        if canonical_specs:
            lines.append("- Canonical specs:")
            for spec_name in sorted(canonical_specs):
                suffix = canonical_specs[spec_name]
                lines.append(f"  - `{spec_name}`{suffix}")
        if other_openspec:
            lines.append("- Other OpenSpec paths:")
            for entry in sorted(other_openspec):
                lines.append(f"  - {entry}")

    if non_openspec:
        lines.append("")
        lines.append("## Code And Docs Changes")
        for area in sorted(non_openspec):
            lines.append(f"- `{area}`:")
            for entry in sorted(non_openspec[area]):
                lines.append(f"  - {entry}")

    lines.append("")
    lines.append("## Notes To Refine Before Posting")
    lines.append("- Replace the first summary bullet with reviewer-facing outcome language.")
    lines.append("- Add validation commands or test notes separately if they ran.")
    lines.append("- Drop low-signal file bullets if the PR is already easy to scan.")
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Generate a PR summary that groups OpenSpec and non-OpenSpec changes.",
    )
    parser.add_argument("--base", help="Base ref for a PR-style compare")
    parser.add_argument("--head", help="Head ref for a PR-style compare")
    parser.add_argument(
        "--staged",
        action="store_true",
        help="Summarize staged changes instead of the full worktree",
    )
    parser.add_argument(
        "--include-untracked",
        action="store_true",
        default=True,
        help="Include untracked files for worktree or staged summaries (default: true)",
    )
    parser.add_argument(
        "--no-include-untracked",
        dest="include_untracked",
        action="store_false",
        help="Ignore untracked files for worktree or staged summaries",
    )
    args = parser.parse_args()

    if args.staged and (args.base or args.head):
        print("Cannot combine --staged with --base/--head.", file=sys.stderr)
        return 2

    try:
        repo_root = repo_root_from_cwd()
        changes, scope = load_changed_paths(repo_root, args)
    except RuntimeError as exc:
        print(str(exc), file=sys.stderr)
        return 1

    if not changes:
        print("No changed files found for the selected diff scope.", file=sys.stderr)
        return 1

    print(summarize_changes(repo_root, changes, scope))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
