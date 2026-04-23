# Delegation Matrix

Use this matrix to choose the best project agent in `.codex/agents/*.toml`. The matrix and agent TOML files are the source of truth for runtime, model, reasoning, and sandbox settings.

## Defaults

- `openspec-explore`: choose one primary agent for the phase and avoid sub-agents
- `openspec-propose`: choose one primary agent for the phase, avoid sub-agents while drafting, and design the future apply delegation plan directly in `tasks.md`
- `openspec-apply-change`: execute the delegation plan already documented in `tasks.md`

## Keep Work Local When

- the next local action depends on the answer immediately
- the task is mostly synthesis or judgment
- the work overlaps the same files as another active slice
- the prompt would require too much hidden context

## Agent Guide

| Situation | Agent | Why this agent |
| --- | --- | --- |
| Codebase reconnaissance, inventories, evidence gathering | `openspec_scout` | Best default for read-only investigation |
| Normal proposal drafting, bounded comparisons, artifact shaping | `openspec_analyst` | Best default for coherent proposal work |
| Hard design tradeoffs or architecture-heavy proposal work | `openspec_architect` | Better fit when design reasoning depth matters |
| Focused implementation in one module or small file cluster | `openspec_implementer` | Best default for straightforward code changes |
| Delicate refactors, subtle bug fixes, test repair | `openspec_refactorer` | Better fit when correctness risk is higher |
| Broad but still bounded implementation slice | `openspec_heavy_lift` | Use when the slice is self-contained but too large for the default implementer |

## Ownership Rules

When writing `tasks.md`, reference only:
- the agent name
- task-specific ownership
- expected output
- blocking notes

Do not restate runtime, model, reasoning, or sandbox details there.

For worker-backed agents:
- assign exclusive file or module ownership
- warn that other agents may be editing elsewhere
- require the agent to list changed files in its final response

## Waiting Guidance

- Do not wait by reflex.
- Wait only when the delegated result is needed for the next critical-path action.
- While agents run, continue with non-overlapping local work.
