---
name: work-handoff
description: Compile an approved plan for a fresh executor only when execution context will change.
version: 0.2.0
---

# /work-handoff - Compile a Self-Contained Execution Contract

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Prepare a fresh executor to run the approved task without relying on chat
history. A fresh executor may be a different agent, provider, worktree, or later
session. When the same agent continues in the same session, the approved plan
is already the execution contract; do not require a handoff.

Input:

`$ARGUMENTS`

Require an accepted definition, human decision, approved plan, and sufficient
acceptance contract.

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

The handoff may summarize but must not expand scope, weaken evidence, or alter
the approved plan. A material change invalidates approval and returns to plan
review and human decision. Otherwise the handoff is immutable during a run. If
a criterion is wrong or impossible, park the task and report the issue instead
of weakening the contract.
