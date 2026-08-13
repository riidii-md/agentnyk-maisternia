---
name: work-speed-loop
description: Coordinate dependency-aware planning, bounded parallel execution, integration, and verification to reduce wall-clock delivery time without sacrificing correctness.
version: 0.1.0
---

# /work-speed-loop - Fast Parallel Work Loop

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Reduce delivery wall time by executing independent plan tasks concurrently.
More workers are not automatically faster; parallelize only demonstrated
independent work.

Input:

`$ARGUMENTS`

## Coordinate

1. Discover provider capabilities, repository rules, current state, authority,
   and budgets.
2. If no schema-valid plan exists, run the `/work-parallel-plan` procedure and
   stop for approval.
3. If the plan is not safely parallelizable, explain why and use sequential
   `/work-run` behavior.
4. For an approved plan, run `/work-parallel-run` in dependency-ready waves.
5. Keep one coordinator responsible for task state, worker scope, integration,
   barriers, failure classification, and replanning.
6. After each verified barrier, launch newly ready tasks up to the configured
   concurrency limit.
7. Run final integration verification and review before declaring completion.

## Provider Degradation

- With subagents and worktree isolation, parallelize independent write tasks.
- With subagents but no write isolation, parallelize read-only tasks and
  serialize writes.
- Without subagents, execute the same dependency plan sequentially and preserve
  restartable task state.
- Never request broader authority merely to increase concurrency.

## Optimize The Right Metric

Primary objective: reduce verified wall-clock completion time.

Also report:

- total tokens and model calls;
- coordination and integration overhead;
- retries and discarded worker output;
- critical-path duration;
- sequential baseline estimate;
- observed speedup and confidence;
- correctness and verification results.

A faster run that costs materially more or weakens correctness must be visible,
not silently labeled an improvement.

## Stop Conditions

Stop and request a decision when the plan lacks approval, write sets overlap,
isolation fails, integration conflicts, protected paths change, the budget is
exhausted, verification repeatedly fails, or scope expands.

Do not push, deploy, expose secrets, weaken checks, or mutate durable harness
configuration without explicit authorization.
