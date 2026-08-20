---
name: work-plan-review
description: Adversarially review a full plan or targeted plan delta against the actual repository, verify every candidate finding, and apply confirmed corrections to the plan artifact.
version: 0.2.0
---

# /work-plan-review - Review A Plan Before Implementation

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Review a plan or a targeted decision change before handoff. Do not implement
product code.

Input:

`$ARGUMENTS`

Accepted forms:

```text
/work-plan-review plan <plan or design path and optional focus>
/work-plan-review plan-delta <changed decision, section, or task>
```

An explicit plan artifact or conversation handoff is sufficient. Never refuse
only because the repository has no implementation diff.

## Resolve And Ground The Target

Read repository instructions, the complete plan or design, relevant source
code, tests, schemas, dependencies, migrations, CI, and accepted decisions. For
`plan-delta`, identify exactly which tasks, interfaces, assumptions, acceptance
criteria, and verification steps the delta can affect. Escalate to a full plan
review only when the delta invalidates broader dependencies or scope.

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

On `pass`, preserve the final reviewed plan as durable Markdown and invoke
`readable-output` to validate and deliver that exact revision through
mdmaid.desk in explicit `plan-decision` mode. Record its path, content hash,
document revision, and review request ID, then report `waiting_for_approval`
while the request is pending. Preserve the human response text returned by the
gate. Approval continues to `/work-decide`; `changes_requested` returns to the
plan and review loop; rejection stops or reshapes the work; and a stale request
requires a fresh review of the current revision. Registration is not approval:
do not infer a decision from the document being registered, opened, marked
done, or closed, and do not begin implementation until the human decision is
recorded.
