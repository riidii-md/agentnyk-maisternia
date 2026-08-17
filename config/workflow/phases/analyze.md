---
name: work-analyze
description: Define the accepted task, constraints, risks, unknowns, and acceptance criteria before solution work begins.
version: 0.1.0
---

# /work-analyze - Define the Task

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Turn gathered facts into a clear bug, feature, refactor, investigation, or
operations definition.

Input:

`$ARGUMENTS`

Separate facts, assumptions, inferences, and unknowns. Define observed and
expected behavior for bugs. Define user goal, desired behavior, constraints,
scope exclusions, and draft acceptance criteria for features.

Do not edit files or choose a solution prematurely.

Return:

- Concise task statement
- Evidence summary
- Scope and exclusions
- Acceptance criteria draft
- Constraints and risks
- Questions needing user decisions
- Recommended next phase
