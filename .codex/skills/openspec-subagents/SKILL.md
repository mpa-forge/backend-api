---
name: openspec-subagents
description: Select the best project agent for `openspec-explore`, `openspec-propose`, and `openspec-apply-change`, and use sub-agents only where they improve OpenSpec implementation quality or feasibility. Use when an OpenSpec task benefits from explicit agent selection, stronger model escalation, or well-scoped delegated implementation slices.
---

# OpenSpec Subagents

Use this skill to choose the best project agent from `.codex/agents/*.toml`.

Treat the model selected in the CLI or app as the orchestrator by default. Its job is to route work to the best-fit project agent. Read [references/delegation-matrix.md](references/delegation-matrix.md) first.

## Rules

1. Optimize for agent fit and outcome quality before concurrency and speed.
2. For `openspec-explore`, use the best single agent for the phase and avoid sub-agents.
3. For `openspec-propose`, use the best single agent for the phase and avoid sub-agents during drafting.
4. In `openspec-propose`, decide where later apply work should stay local, use one sub-agent, or parallelize across multiple sub-agents, and encode that directly in `tasks.md`.
5. Treat `tasks.md` as the source of truth for apply-time delegation.
6. In `openspec-apply-change`, follow the delegation plan already documented in `tasks.md` instead of redesigning it.
7. Do not force parallelism. Some tasks stay local, some tasks share one agent, and some task groups justify multiple agents.
8. Keep ownership explicit for any delegated implementation slice.

## Phase Use

### Explore

- Choose the best single agent for the exploration need.
- Start with that agent right away.
- Avoid sub-agents unless there is an exceptional reason.

Typical fit:
- `openspec_scout` for codebase reconnaissance and evidence gathering
- `openspec_architect` for design-heavy exploration

### Propose

- Choose the best single agent for the proposal phase.
- Keep proposal, design, and task writing coherent in one thread.
- While writing `tasks.md`, decide the future apply strategy for each meaningful task or task group.
- Mark which work should stay local, which should use one better-fit sub-agent, and which could be parallelized across multiple sub-agents.
- Record the agent reference and ownership shape directly in the task breakdown where delegation is useful and feasible.

Typical fit for proposal task:
- `openspec_analyst` for normal proposal work
- `openspec_architect` for harder design tradeoffs
Typical fit for implementation tasks:
- `openspec_implementer` for focused implementation tasks the proposer expects to delegate later
- `openspec_refactorer` for delicate refactors or risky fixes the proposer should pre-assign carefully
- `openspec_heavy_lift` for broad but still bounded implementation slices that may justify a stronger single worker

### Apply

- Read the current task and execute the delegation plan already documented in `tasks.md`.
- If the task is marked local, keep it local.
- If the task names one sub-agent, spawn that sub-agent.
- If the task names multiple sub-agents, spawn them according to the documented ownership split.
- Keep integration, final validation, and cross-cutting judgment local unless there is a strong reason not to.


## Task Breakdown Format

During propose, build `tasks.md` so it already reflects the intended apply-time execution strategy:

```md
- [ ] 2.3 Inventory auth touchpoints across API and frontend
  Sub-agent: `openspec_scout`
  Ownership: Read-only inventory of routes, middleware, and auth guards
  Expected output: File/path list, flow summary, open questions
  Blocking: No - continue local design synthesis while the agent runs
```

For grouped or split implementation:

```md
- [ ] 4.2 Implement API token validation and frontend session refresh
  Sub-agents:
  Agent A: `openspec_implementer`, owns `backend-api/internal/auth/**`
  Agent B: `openspec_implementer`, owns `frontend-web/src/auth/**`
  Integration owner: Main agent
  Blocking: Partial - keep local work on shared contracts and integration tests
```

Do not repeat runtime, model, reasoning, or sandbox details in `tasks.md`. The agent reference and delegation matrix are the source of truth.
The apply phase should be able to execute from this plan without re-deciding delegation.

## Spawning

When spawning a sub-agent:
- state the exact subtask
- state the owned files or modules
- state the expected output
- keep the prompt narrow
- before using a pinned model from `.codex/agents/*.toml`, confirm it appears in `~/.codex/models_cache.json`; if it does not, use a supported fallback from the cache instead of stalling on spawn failure

For worker-oriented agents, explicitly say they are not alone in the codebase and must not revert others' edits.

## Keep It Local When

- the task is small enough that delegation adds overhead
- the next step is blocked on one judgment call
- the work overlaps heavily with another active slice
- the prompt would require too much hidden context
- one coherent reasoning thread matters more than throughput

## Improvement Trigger

If the workflow repeatedly struggles because the right agent is unclear, delegation boundaries are fuzzy, or task grouping is hard to express, propose updates to this skill, the delegation matrix, or the project agents.
