# /work - Provider-Neutral Work Conductor

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
