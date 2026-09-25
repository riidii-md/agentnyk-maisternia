# Architectural Direction Gate

`/work-direction` creates a high-level decision contract between problem
analysis and a detailed implementation plan. It aggregates repository and
bounded cross-system scouting, research, focused human answers, and an optional
user-authored sketch. It decides what shape the solution should have;
`/work-plan` specifies how to implement that accepted shape.

## When It Applies

Require the direction gate for changes spanning systems, owners, trust
boundaries, public contracts, persistent data, migrations, rollout models, or
materially different viable designs; for costly-to-reverse decisions; or when
the human requests architectural review. A small, local, reversible change
with one evident approach can bypass it if `/work-analyze` records why.
Unknown impact calls for more scouting or research, not an assumed bypass.

```text
SCOUT -> ANALYZE -> [RESEARCH <-> GRILL] -> DIRECTION
                                            -> DIRECTION REVIEW
                                            -> HUMAN DIRECTION DECISION
                                            -> DETAILED WORK PLAN
```

The bracketed research and grill steps are conditional. `/work-grill sketch`
may, after presenting a short evidence summary, invite the human to describe
their preferred architecture, boundaries, non-negotiables, acceptable
tradeoffs, and concerns in their own words. The response is attributed input,
not approval.

If drafting the direction exposes a boundary, evidence, or human-constraint
gap, return directly to scout, research, or grill before direction review.

## Parallel Analysis, Single Authorship

When the direction gate applies and native subagents are available,
`/work-direction` runs three independent read-only analysis lanes in parallel:
boundaries/interfaces/trust, constraints/migration/operations, and
alternatives/tradeoffs. The coordinator alone writes the canonical direction
and preserves material disagreement instead of deciding by vote.
Routed external workers and native same-harness workers fill the same graph,
without duplicating every lane at each provider.

`/work-plan` uses the same pattern for non-trivial work: code impact,
verification evidence, and delivery risk/sequencing run as independent lanes,
then one coordinator writes the dependency-ordered plan. Small local work may
skip this graph with a recorded reason. If advertised subagent spawning fails,
the workflow asks before sequential fallback rather than silently reducing
independence. The installed `design-graph-policy.json` carries this shared
provider-neutral contract.

## Direction Artifact

The artifact records outcomes and exclusions; system boundaries, ownership,
and sources of truth; components and data/control flows; interfaces and trust
boundaries; the implementation shape of representative functions, methods,
types, classes, interfaces, or modules; and the system-design and code-level
patterns that are established, proposed, adapted, or rejected. Pattern decisions
identify their participants, the concrete problem they solve, and their
consequences rather than applying labels ceremonially. The artifact also records
the selected approach and rationale; rejected alternatives and tradeoffs;
implementation constraints; rollout, compatibility, observability, and recovery
implications; risks, assumptions, unresolved questions, and acceptance signals.
It cites its evidence and attributed human input.

It intentionally excludes file-by-file edits, exhaustive symbol inventories,
final private-helper details, ordered coding tasks, and exhaustive test commands.
Those belong to the detailed implementation plan, which must be executable by a
fresh agent without inventing material design.

## Representation And Decision Readability

For a non-trivial direction, architecture and sequence views are strongly weighted
candidates rather than mandatory deliverables. Architecture is evaluated when
components, boundaries, ownership, interfaces, dependencies, sources of truth,
or trust matter. A sequence is evaluated when ordered calls, messages,
asynchronous work, retries, failures, or lifecycle behavior matter.
A table or structured prose is preferred when it presents the same evidence more simply
and completely. The artifact records why a weighted view was omitted, and no
diagram may be the only carrier of essential information.

After review findings are resolved, the coordinator performs a final in-place
reader pass before the exact revision is hashed. It optimizes decision readiness,
completeness, and reading effort rather than word count. The pass may restructure
prose or add a useful table or visual, but it must preserve facts, decisions,
evidence status, uncertainty, constraints, alternatives, risks, interfaces, and
open questions. Meaning-changing edits return to review.

## Review And Decision

`/work-plan-review direction` checks the exact artifact against repository
and adjacent-system evidence at architectural altitude. Confirmed corrections
are applied within scope; product decisions return to the human. The reviewed
revision is validated and delivered through mdmaid.desk. The currently
supported authenticated transport is `plan-decision` with
`kind=decision`; its title, request
message, and local receipt identify semantic mode `direction`.

`/work-decide direction` records the explicit human response against document
revision and content hash. Approval lets detailed planning start. Requested
changes return to direction and review; rejection stops or reshapes; stale
requires a new request for the current revision. Reading or registering a
document never implies approval.

## Fewer, Stable Documents

Prefer one evolving discovery brief when a durable checkpoint is useful, one
architectural direction, one detailed plan, and one exact implementation change
review. Reuse a stable task-and-role path for revisions rather than registering
one new document per phase. Within a live session, reuse a successful
mdmaid.desk version/capability and workspace preflight while executable and
repository root are unchanged; retry the same operation after any failure.
This reduces repeated setup, not validation or human-decision checks.

`/work-scout` follows concrete task references into adjacent repositories,
services, interfaces, data stores, deployment paths, and trust boundaries.
It records evidence-backed affected, unaffected, or unknown status. It does
not scan unrelated repositories or treat unavailable evidence as proof of no
impact.
