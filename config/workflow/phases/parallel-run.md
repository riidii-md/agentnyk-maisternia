---
name: work-parallel-run
description: Execute an approved dependency-aware plan in bounded parallel waves with isolated workers, deterministic integration, and verification barriers.
version: 0.1.0
---

# /work-parallel-run - Execute An Approved Parallel Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Execute a schema-valid, human-approved parallel plan. The current provider
harness owns workers and continuation; `maisternia` only installed this command.

Input:

`$ARGUMENTS`

## Preconditions

Before launching workers:

1. verify the plan approval, base commit, repository rules, and cleanly
   understood working-tree state;
2. revalidate dependencies, write-path ownership, protected paths, provider
   capabilities, concurrency, and budgets;
3. confirm every write-capable parallel task has an isolated worktree or branch;
4. downgrade write tasks to serialized execution when isolation is unavailable;
5. stop for replanning when current code invalidates the approved task graph.

Read-only research and review may run concurrently without worktrees. Never let
two workers write to the same checkout.

## Dispatch Ready Waves

Select only tasks whose dependencies are complete. Launch at most the configured
parallelism limit. Give every worker a bounded packet containing:

- task ID, objective, acceptance criteria, and dependencies;
- exact base revision and isolated working directory;
- allowed read paths, exclusive write paths, and protected paths;
- required authority, tools, and repository instructions;
- focused verification commands;
- artifact, patch, or explicitly authorized local-commit handoff format;
- token, attempt, and wall-time limits;
- stop conditions and required evidence.

Workers must not broaden scope, modify another task's write paths, integrate
other work, push, publish, deploy, or weaken acceptance criteria.

## Integrate At A Barrier

After every worker in the wave finishes:

1. collect status, changed files, diff or local commit reference, tests, logs,
   cost, blockers, and remaining uncertainty;
2. reject handoffs that changed undeclared paths or do not match the base;
3. integrate successful work in declared dependency order using the approved
   patch, cherry-pick, or serialized strategy;
4. stop on semantic or textual conflicts instead of guessing intent;
5. run the wave's integration checks against the combined tree;
6. mark tasks complete only after the barrier passes;
7. recompute ready tasks and the remaining critical path.

Do not start dependents from an unverified parent result.

## Failure And Replan Rules

- Retry only when the failure classification and next change are concrete.
- Park a task after the configured attempt limit.
- Cancel or pause dependents of a failed task.
- Continue independent branches only when doing so cannot create wasted or
  conflicting work.
- Return to `/work-parallel-plan` when interfaces, write ownership, dependencies,
  or acceptance criteria have changed.
- Fall back to sequential `/work-run` when safe parallel execution is no longer
  possible.

## Complete

Run final repository-specific verification and independent review after all
waves integrate. Report:

- completed, failed, parked, and remaining tasks;
- wave timeline and actual concurrency;
- changed files and integration order;
- worker and integration checks with results;
- measured wall time, provider usage, and coordination overhead when available;
- estimated sequential time and observed speedup, clearly labeled;
- remaining risks and exact next action.

Do not commit integrated work, push, open a PR, or perform destructive or
production actions unless separately authorized.
