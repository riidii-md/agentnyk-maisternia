---
name: work-plan-review
description: Adversarially review a full plan or targeted plan delta against the actual repository, verify every candidate finding, and apply confirmed corrections to the plan artifact.
version: 0.6.0
---

# /work-plan-review - Review A Plan Before Implementation

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Review an architectural direction, an implementation plan, or a targeted plan
change before handoff. Do not implement product code.

Input:

`$ARGUMENTS`

Accepted forms:

```text
/work-plan-review direction <direction artifact and optional focus>
/work-plan-review plan <plan or design path and optional focus>
/work-plan-review plan-delta <changed decision, section, or task>
```

An explicit plan artifact or conversation handoff is sufficient. Never refuse
only because the repository has no implementation diff.

## Resolve And Ground The Target

Read repository instructions, the complete direction, plan, or design, relevant source
code, tests, schemas, dependencies, migrations, CI, and accepted decisions. For
`plan-delta`, identify exactly which tasks, interfaces, assumptions, acceptance
criteria, and verification steps the delta can affect. Escalate to a full plan
review only when the delta invalidates broader dependencies or scope.

For `direction`, review at architectural altitude: system boundaries,
ownership, interfaces, data or control flows, trust boundaries, rejected
alternatives, migration and operational implications, and constraints handed
to implementation planning. Verify that its representative implementation shape
and pattern decisions are supported by repository evidence, identify concrete
participants and consequences, and remain detailed enough to constrain planning
without prematurely fixing incidental symbols. Do not demand file-by-file tasks
or detailed test commands from a direction. The fresh-executor criterion below
applies to implementation plans, not architectural directions.

Apply the fresh-executor criterion at risk-appropriate depth: a fresh executor
must be able to implement the plan without inventing material architecture,
interfaces, dependencies, storage or state behavior, current and proposed
implementation composition, material pattern decisions, or cross-component
behavior. Verify that the pattern register names only evidence-backed or
intentional patterns and records their participants, purpose, status, and
tradeoffs. A small local change may explicitly state that these concerns are
unaffected. For non-trivial work, missing material design, ownership boundaries,
contracts, composition, pattern decisions, data or control flow, task
dependencies, or design-changing open decisions is High and blocking.
Do not invent the missing design during review; return it to planning and human
decision.

## Run Independent Lenses

Run one read-only reviewer per base lens:

- `correctness-vs-code`: proposed behavior, paths, symbols, APIs, and assumptions
  match the current repository;
- `internal-consistency`: tasks, dependencies, decisions, terminology, and
  acceptance criteria do not contradict each other;
- `completeness-edge-cases`: required behavior, failure paths, rollout,
  rollback, and verification are covered;
- `architecture-simplicity`: the design fits existing boundaries without
  accidental coupling or unnecessary abstraction;
- `best-practices`: repository and domain practices are followed without
  cargo-cult additions;
- `acceptance-testability`: every important claim has observable proof.

The `architecture-simplicity` reviewer must challenge every proposed feature,
dependency, configuration surface, layer, abstraction, and group of new files.
It checks whether the plan can apply YAGNI or reuse existing repository code.
It then considers the standard library, a native platform capability, an
already-installed dependency, and direct control flow. A new abstraction is
justified only by proven repeated knowledge, clear ownership, or materially
lower coupling. A simpler candidate must identify the accepted behavior and
safeguards it preserves, the plan tasks it replaces or shrinks, and the
evidence that the simpler direction fits the current repository. If it changes
an accepted requirement or decision, leave it for the user instead of applying
it silently.

Add domain lenses when warranted: accessibility, privacy/PII/SOC 2 evidence,
performance/scalability, migration safety, API compatibility, or data integrity.
Do not claim compliance without repository requirements and evidence.

Every reviewer reads the actual code and returns zero or more candidates with
severity, claim, impact, proposed fix, and concrete `file:line`, short verbatim
quote, command/test, or authoritative-document grounding.

## Refute Every Candidate

Spawn a separate verifier per candidate. Its first objective is to disprove the
claim against the plan, code, tests, and repository rules. It must return:

```text
is_real: true|false
grounded: true|false
rationale
supporting or refuting evidence
corrected severity
```

Keep only `is_real && grounded`. Record refuted findings and why. Deduplicate
confirmed findings and rank them Critical, High, Medium, then Low.

## Report And Apply

Write the report before mutation. Then the coordinator, not a reviewer, edits
the plan or design document to apply every confirmed fix within scope. Do not
silently resolve product decisions or accepted risk; mark those fixes blocked
and ask the user. Critical and High findings are blocking.

Re-read the edited plan, rerun affected consistency and acceptance checks, and
set the gate to `pass`, `fail`, or `blocked`. Write `review.md` and schema-valid
`review.json` under `.agent-runs/reviews/<run-id>/`, including confirmed,
refuted, applied, and blocked findings.

## Build The Visual Plan Review

For a plan review that passes, build a standalone approval artifact at:

```text
.agent-runs/plan-reviews/<run-id>/plan-review.md
```

It must contain the complete reviewed plan—the final reviewed plan revision—not
a summary that requires the reader to open another file, plus:

- a 60-second summary, scope, decisions, tradeoffs, and open questions;
- current repository evidence separated from proposed rather than verified
  architecture and behavior;
- planned interface, type, schema, ownership, dependency, state, and interaction
  inventories at the abstraction level needed by a fresh executor;
- the current and proposed implementation composition, distinguishing verified repository evidence from reviewed proposals;
- the pattern register, including material participants, evidence, and tradeoffs;
- confirmed, refuted, applied, blocked, and residual review findings;
- a visual lens selection table recording evidence and why omitted lenses are
  not applicable;
- the smallest evidence-complete Mermaid portfolio.

Use the installed `change-explanation` skill in plan evidence mode. Evaluate
`flowchart`, `classDiagram`, `erDiagram`, `stateDiagram-v2`,
`sequenceDiagram`, and `requirementDiagram`. Do not force every visual lens,
invent interfaces, or present planned relationships as implemented. Retain each
`.mmd` source under the same run directory, verify it with
`mdmaid render-mermaid`, and embed it once beside the plan section it explains.
When `requirementDiagram` is not supported by the terminal backend, preserve it
for compatible web readers and add an equivalent traceability `flowchart`.

Preserve the source plan path and content hash in `plan-review.md`. Then invoke
`readable-output` to validate and deliver that exact standalone artifact through
mdmaid.desk as `kind=plan` in explicit `plan-decision` mode. Record its path,
content hash, document revision, and review request ID. Then keep the current agent turn open on the foreground waiter. Report `waiting_for_approval` only as
an intermediate update, never as a final response. Resume a yielded process or
session ID until it exits. Surface the outcome and human response text immediately. Approval continues to `/work-decide`; `changes_requested`
returns to the plan and review loop; rejection stops or reshapes the work; and
a stale request requires a fresh review and visual artifact for the current
revision. Registration is not approval: do not infer a decision from the
document being registered, opened, marked done, or closed, and do not begin
implementation until the human decision is recorded.

For a passing `direction` review, preserve the complete reviewed direction
revision in a standalone direction review artifact, with a concise evidence
summary, confirmed/refuted/applied/blocked findings, and the exact source path
and content hash. Use only architecture-level visuals that materially clarify
the choice; do not run the implementation-plan visual inventory or call the
direction implementation-ready. Invoke `readable-output` to validate and
deliver the exact reviewed direction through mdmaid.desk with an unambiguous
direction title and request message. Use the currently supported authenticated
`plan-decision` transport with `kind=decision`, record semantic decision mode `direction` with
document ID, revision, review request ID, path, and content hash, and keep the
current agent turn open on the foreground waiter. Approval continues to
`/work-decide direction`; requested changes return to direction and review;
rejection stops or reshapes; and stale returns to current-revision review.
Registration, opening, or closing is never direction approval.
