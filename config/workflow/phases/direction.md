---
name: work-direction
description: Synthesize reviewed evidence and human input into a high-level architectural direction before detailed implementation planning.
version: 0.1.0
---

# /work-direction - Define The Architectural Direction

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Create a high-level architectural direction from the accepted task definition,
repository and bounded cross-system evidence, research, material grill answers,
and any user-authored sketch. This is the bridge from solution shaping to a
detailed implementation plan.

Use this conditional gate when a task crosses systems, owners, trust boundaries,
public contracts, persistent data, migration or rollout models; is costly to
reverse; has several materially different viable designs; or the human asks for
an architectural decision. For a small, local, reversible change with one
evident approach, record an evidence-backed skip in `/work-analyze` and proceed
to `/work-plan`.

Input:

`$ARGUMENTS`

## Direction Contract

Ground decisions in evidence. Preserve a user-authored sketch as attributed
input, not automatic approval. Resolve discoverable facts before asking the
human. Return to scout, research, or grill when a missing fact or preference
could change the architecture.

The durable Markdown architectural direction must contain:

- problem, outcomes, scope, and exclusions;
- system boundaries, ownership, and sources of truth;
- main components and data or control flows;
- interfaces, contracts, trust boundaries, and cross-system effects;
- selected approach, rationale, key decisions, and implementation constraints;
- rejected alternatives and material tradeoffs;
- migration, rollout, compatibility, observability, and recovery implications
  when relevant;
- risks, assumptions, unresolved questions, and acceptance signals;
- evidence and attributed human input supporting the decisions.

Keep the direction at architectural altitude. Do not produce implementation tasks,
file-by-file edits, symbol-level changes, or an exhaustive test-command
list; those belong to `/work-plan`.

Write one durable Markdown direction artifact at an explicit task artifact path
when provided, otherwise under `.agent-runs/readable-output/`. Reuse its stable
task-and-role path for revisions and record its content hash. Present it through
`/work-plan-review direction`. An agent-produced direction is not approved:
do not start detailed planning until the exact reviewed revision has an
explicit human direction decision.
