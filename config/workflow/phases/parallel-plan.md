---
name: work-parallel-plan
description: Decompose an implementation goal into a dependency-safe parallel plan with isolated write ownership, execution waves, integration barriers, and measurable acceptance criteria.
version: 0.1.0
---

# /work-parallel-plan - Create A Parallel Implementation Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Create an implementation plan that can safely run more than one task at a time.
Do not edit files or launch implementation workers in this phase.

Input:

`$ARGUMENTS`

## Discover Before Decomposing

Read repository instructions, architecture, ownership boundaries, current Git
state, build and test commands, generated-file rules, migration constraints, and
provider capabilities. Identify the requested behavior, acceptance criteria,
scope exclusions, authority, budget, and protected paths.

## Decide Whether Parallelism Helps

Parallel execution is useful only when at least two ready tasks can proceed
without shared mutable state. Estimate the critical path and coordination cost.
Choose sequential `/work-run` when:

- all meaningful tasks depend on the same unfinished change;
- tasks need overlapping write paths or the same generated files;
- isolation is unavailable for write-capable workers;
- integration or review overhead is likely to exceed saved wall time;
- the task is too small to justify delegation.

Do not invent parallel work to satisfy this command.

## Build The Task Graph

For each task define:

- stable task ID, title, objective, and task kind;
- explicit `depends_on` task IDs;
- read paths and exclusive write paths;
- required authority and isolation mode;
- focused acceptance criteria and verification commands;
- expected artifact, patch, or local commit handoff;
- estimated duration and important risks.

Prefer end-to-end slices with one owner. Split by independently verifiable
ownership boundaries, not by arbitrary file counts. Keep integration, shared
interfaces, generated files, global dependency files, migrations, and final
verification as explicit serialized tasks when they cannot be isolated.

## Validate The Decomposition

Before presenting the plan, prove:

1. the dependency graph is acyclic and every task is reachable;
2. every dependency names a declared task;
3. tasks in the same wave have no overlapping write paths;
4. parent and child paths count as overlapping;
5. generated files and shared manifests have one owner per wave;
6. each write-capable task has an isolated worktree, branch, or serialized mode;
7. every task has observable acceptance and focused verification;
8. integration order follows dependencies;
9. final verification covers the combined result;
10. concurrency and budget limits are explicit.

If any check fails, revise the graph or serialize the conflicting tasks.

## Output

Write:

```text
.agent-runs/parallel/<run-id>/parallel-plan.md
.agent-runs/parallel/<run-id>/parallel-plan.json
```

`parallel-plan.json` must conform to the installed parallel-plan schema. The
Markdown view must include:

- scope and repository rules;
- why parallel execution is or is not beneficial;
- task dependency graph;
- wave table with task owners and write paths;
- critical path and estimated sequential versus parallel duration;
- isolation and integration strategy;
- per-wave and final verification barriers;
- cost and concurrency limits;
- failure, replan, and stop conditions;
- exact human approval requested.

Leave approval status as `draft` until the user explicitly approves the plan.
