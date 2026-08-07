---
name: parallel-work
description: Use when an approved implementation can be decomposed into independent tasks and the user wants a faster work loop through safe parallel execution.
---

# Parallel Work

Use this skill to reduce wall-clock delivery time with a dependency-aware plan.

1. `/work-parallel-plan` discovers write ownership, dependencies, waves,
   integration, verification, and expected speedup.
2. `/work-parallel-run` dispatches approved ready tasks with isolated scopes and
   integrates them at deterministic barriers.
3. `/work-speed-loop` coordinates planning and execution with safe fallback.

Do not parallelize overlapping write paths, shared generated state, or tasks on
the same critical dependency chain. Require worktrees or equivalent isolation
for concurrent writers. Use parallel read-only work and serialized writes when
the provider cannot isolate write-capable workers.

Speed is measured after verification. Report extra token and coordination cost,
and fall back to sequential work when parallel overhead exceeds the likely gain.
