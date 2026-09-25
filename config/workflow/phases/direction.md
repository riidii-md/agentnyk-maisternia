---
name: work-direction
description: Synthesize reviewed evidence and human input into a high-level architectural direction before detailed implementation planning.
version: 0.2.0
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

## Run The Complexity-Gated Agent Graph

Read the installed `design-graph-policy`. When the direction gate applies, use
its complexity-gated agent graph. If the current harness exposes native
subagents, they are required for lanes assigned to that harness. Routed external
workers count toward the same graph. Dispatch one independent read-only worker
for each bounded lane, in parallel up to the available capacity:

- `boundaries-interfaces-trust`: system ownership, interfaces, sources of truth,
  data or control flow, and trust boundaries;
- `constraints-migration-operations`: compatibility, migration, rollout,
  observability, recovery, and operational constraints;
- `alternatives-tradeoffs`: materially different approaches, reversibility,
  cost, risks, and rejected alternatives.

Give each worker the accepted task, attributed human input, and the minimum
evidence needed for its lane. Workers inspect evidence and return proposals;
they do not edit the direction artifact. The coordinator is the single
canonical writer. It reconciles disagreements and records unsupported
assumptions instead of deciding by vote.

Do not silently collapse an advertised multi-agent graph into coordinator-only
work. If spawning is unexpectedly unavailable or fails, ask before sequential
fallback. A provider that does not expose subagents may run the same
lanes sequentially only after disclosing that limitation. For a small local
task where `/work-analyze` recorded that no direction gate is needed, skip this
graph and continue to `/work-plan`.

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
- the affected implementation shape: the current and proposed responsibilities
  of representative functions, methods, types, classes, interfaces, or modules,
  and enough of their composition and collaboration to make the direction
  concrete without prematurely fixing incidental code details;
- pattern decisions across system design and code design: patterns already
  established in the affected code, patterns to reuse or introduce, and material
  patterns considered but rejected; identify the problem each pattern solves, its
  participants, evidence, and consequences or constraints;
- selected approach, rationale, key decisions, and implementation constraints;
- rejected alternatives and material tradeoffs;
- migration, rollout, compatibility, observability, and recovery implications
  when relevant;
- risks, assumptions, unresolved questions, and acceptance signals;
- evidence and attributed human input supporting the decisions.

Keep the direction at architectural altitude. Representative code elements may
explain implementation shape and pattern decisions. Do not produce implementation tasks
or an exhaustive test-command list. Avoid file-by-file edits, an exhaustive symbol
inventory, and final private-helper names or signatures; those belong to
`/work-plan`. Do not force pattern names onto ordinary decomposition:
name a pattern only when repository evidence or the proposed design gives it
specific participants, behavior, and consequences.

Write one durable Markdown direction artifact at an explicit task artifact path
when provided, otherwise under `.agent-runs/readable-output/`. Reuse its stable
task-and-role path for revisions and record its content hash. Present it through
`/work-plan-review direction`. An agent-produced direction is not approved:
do not start detailed planning until the exact reviewed revision has an
explicit human direction decision.
