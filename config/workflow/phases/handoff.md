---
name: work-handoff
description: Compile approved findings and decisions into a self-contained, implementation-ready execution contract.
version: 0.1.0
---

# /work-handoff - Compile a Self-Contained Execution Contract

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

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
