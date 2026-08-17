---
name: work
description: Conduct provider-neutral work through the smallest useful discovery, decision, execution, verification, and review phases.
version: 0.1.0
---

# /work - Provider-Neutral Work Conductor

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Coordinate the current task without assuming a specific agent provider.

Input:

`$ARGUMENTS`

## Behavior

1. Discover the active task from explicit arguments, ticket, branch, worktree,
   repository, PR, or durable task state.
2. Read `state.yaml` and `events.jsonl` when they exist.
3. Report the current phase, status, blockers, approvals, and next action.
4. Validate that required artifacts for the next phase exist.
5. Recommend exactly one next phase.
6. Ask before implementation, permission escalation, commit, push, PR, or a
   destructive operation when approval is not already recorded.
7. Dispatch through the configured runner policy or honor an explicit runner.
8. Record the phase result and next action in durable state.

Do not silently skip readiness, proof, approval, verification, or independent
review gates.

## Output

- Task
- Current phase and status
- What happened most recently
- Blocking issues
- Recommended next phase
- Approval required
- Selected runner and reason
