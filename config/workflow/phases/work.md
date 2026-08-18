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

1. Discover the active task from explicit arguments, the current conversation,
   ticket, branch, worktree, repository, PR, or user-supplied artifact.
2. Infer completed and pending work from evidence available in the current
   session; do not create bookkeeping merely to run the workflow.
3. Report the current phase, status, blockers, approvals, and next action.
4. Validate that required inputs for the next phase exist.
5. Recommend exactly one next phase.
6. Ask before implementation, permission escalation, commit, push, PR, or a
   destructive operation when approval is not already recorded.
7. Dispatch through the configured runner policy or honor an explicit runner.
8. Report the phase result and next action to the coordinating session.

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

Maisternia may configure this command and its routing policy, but it is not the
runtime task store or phase controller.
