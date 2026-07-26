# /work-handoff - Compile a Self-Contained Execution Contract

Prepare a fresh agent to execute the approved task without relying on chat
history.

Input:

`$ARGUMENTS`

Require an accepted definition, decision, plan, and proof contract.

Compile:

- goal and scope;
- repository rules;
- ordered tasks;
- acceptance contract;
- verification commands;
- guardrails and approval boundaries;
- retry, parking, and stop behavior;
- worktree and branch;
- progress and event locations.

The handoff is immutable during a run. If a criterion is wrong or impossible,
park the task and report the issue instead of weakening the contract.
