---
name: work
description: Conduct provider-neutral work through the smallest useful discovery, decision, execution, verification, and review phases.
version: 0.3.0
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
4. Validate that required artifacts for the next phase exist.
5. Recommend exactly one next phase. Treat research, the conditional direction
   gate, expanded proof, handoff, and PR preparation as conditional work
   selected by evidence, risk, executor continuity, and publication intent.
   Direction and plan AI review are selected only by the human at the explicit
   review-choice checkpoint, never by inferred risk.
6. Require an architectural direction and explicit direction decision
   before the detailed implementation plan for cross-system, cross-owner,
   trust-boundary, public-contract, persistent-data, migration, rollout,
   costly-to-reverse, or materially ambiguous work, or when the human requests
   it. Record an evidence-backed skip for a small, local, reversible task.
7. After direction and plan artifacts are produced, require an explicit choice
   to run AI review now, skip AI review, or review later. Do not automatically
   invoke plan review. Before implementation, require the exact plan revision to
   be presented for human attention, record the explicit decision against its
   content hash, and pass the implementation-readiness gate.
8. After verification and independent implementation review pass, require
   `/work-change-review implementation-approval` and a durable `approved`
   decision for the exact implementation snapshot before `/work-pr`,
   publication, or completion.
9. Ask before implementation, permission escalation, commit, push, PR, or a
   destructive operation when approval is not already recorded.
10. Dispatch through the configured runner policy or honor an explicit runner.
11. Report the phase result and next action to the coordinating session.

Do not silently run or skip optional AI direction or plan review. Do not
silently skip required readiness, acceptance evidence, approval, implementation
verification, or implementation review gates. Do not manufacture separate
artifacts when the approved plan already contains sufficient evidence, and do
not require a handoff when the same agent continues in the same session.

The decision sequence is evidence and focused human context, optional user
sketch, conditional direction, explicit optional-review choice and direction
decision, detailed implementation plan, explicit optional-review choice and
plan decision, then execution.

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
