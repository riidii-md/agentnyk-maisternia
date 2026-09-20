---
name: work-plan
description: Create a reviewable implementation proposal with ordered changes, decisions, acceptance evidence, risks, and verification gates.
version: 0.5.0
---

# /work-plan - Create the Implementation Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Produce an actionable implementation proposal. When the direction gate applied,
the exact reviewed and approved direction constrains it. The implementation
proposal itself is not approved until the human reviews the final revision and
makes an explicit decision.

When `work-routing` resolves several harnesses, request independent plans and
let the current coordinating harness synthesize one plan while preserving
material disagreements and unsupported assumptions.

Input:

`$ARGUMENTS`

Discover repository rules before assuming paths, base branches, ticket formats,
tools, tests, or PR conventions.

Inspect the affected code and, when direction was required, locate the accepted direction
artifact, decision, revision, and content hash. Treat its architecture,
boundaries, constraints, and accepted risks as inputs rather than silently
redesigning them. If a required direction is missing, stale, or materially
contradicted by code, return to direction and review.

When analysis records that direction was not required, include the evidence for
that skip and propose the simplest viable direction satisfying accepted scope
and safeguards. Prefer omitting speculative work and reusing existing
repository code. Use the standard library, a native platform capability, an
already-installed dependency, or direct local code before introducing a new
abstraction. Explain the evidence and complexity avoided.

Simplicity does not authorize narrowing requested behavior or weakening
correctness, validation, security, accessibility, compatibility, data-loss
prevention, or verification. When a simpler direction materially changes behavior
or an accepted constraint, risk, user experience, compatibility, or
long-term ownership, present the concrete alternatives and ask the user before
finalizing the plan. Choose between equivalent implementation details without
asking.

## Define The Implementation Design

For non-trivial work, make the affected-system design explicit before listing
tasks. The plan is sufficiently detailed when a fresh executor can implement it
without inventing material architecture, interfaces, dependencies, storage or
state behavior, or cross-component behavior. Cover only the architecture affected
by the task; do not manufacture a whole-application design or speculative
class-by-class detail.

Describe, at risk-appropriate depth:

- the relevant current architecture, established patterns, and ownership boundaries;
- proposed components, their responsibilities, and where changed behavior belongs;
- material interfaces, schemas, protocols, and state transitions, including error
  semantics and compatibility expectations;
- the important data and control flow across boundaries;
- dependency, configuration, persistence, migration, rollout, and rollback effects;
- decisions fixed by the plan, open decisions requiring human judgment, and
  permitted executor discretion.

For a small local change, explicitly state which of these concerns are unaffected
instead of adding ceremonial design sections. Private helper names, equivalent
local control flow, and test-fixture organization may remain executor choices when
they do not alter an approved contract.

For non-trivial plans, add a compact visual lens selection table and embed the
smallest useful Mermaid portfolio. Select only evidence-supported lenses:
`flowchart` for architecture, dependencies, or data flow; `classDiagram` for
interfaces and type relationships; `erDiagram` for planned persisted entities;
`stateDiagram-v2` for lifecycle transitions; `sequenceDiagram` for ordered
cross-boundary behavior; and `requirementDiagram` for requirement-to-component
and requirement-to-verification traceability. Label every planned element and
relationship as proposed rather than verified implementation. Record why an
inapplicable lens was omitted; do not generate all diagram forms ceremonially.

Plan in dependency order. Each task should describe observable behavior, fit one
focused implementation loop, and keep the repository runnable. Identify a thin
end-to-end slice first when appropriate. For each task, identify dependencies and
prerequisite tasks, affected files or components, contract changes or an
explicit statement that there are none, expected behavior, failure and edge cases,
focused tests, and observable completion evidence. Avoid broad tasks such as
"implement the backend" that require the executor to perform hidden decomposition.

Return:

- Discovered repository rules
- Scope and exclusions
- Accepted direction artifact, decision, revision, and content hash, or the
  evidence-backed reason direction was not required
- Approved direction summary when its gate applied, or the evidence-backed skip
- Implementation approach within the accepted direction; do not choose a second
  architecture
- Simplest viable direction, evidence, and complexity avoided only when the
  direction gate was skipped
- Affected-system design, ownership boundaries, and responsibilities
- Material interfaces and cross-component contracts
- Planned data and control flow
- Visual lens selection and evidence-grounded Mermaid diagrams when applicable
- Fixed decisions and permitted executor discretion
- Material alternatives and tradeoffs
- Files and patterns to change or reuse, with inspect-only unknowns explicit
- Ordered implementation tasks
- Risk and blast-radius checks
- Migration or rollout concerns
- Acceptance contract with observable evidence and expected verification
- Stop conditions
- Open decisions requiring human judgment
- Whether `/work-prove` is needed as an optional expansion

Write the complete plan as durable Markdown at the explicit task artifact path
when one exists, otherwise under `.agent-runs/readable-output/`. For complex or
high-risk work, recommend `/work-review plan` before presentation. When no
separate plan review is needed, use `readable-output` to validate and deliver
the plan through mdmaid.desk in explicit `plan-decision` mode. Include a useful
request message, record the request ID and exact revision, and wait for the
durable result; keep the current agent turn open while the foreground waiter is
pending. `waiting_for_approval` is an intermediate update, never a final
response. If the execution tool yields a process/session ID, resume that same
process until it exits. Surface the outcome and human response text immediately
when it returns: approval continues to `/work-decide`, requested changes return
to planning, rejection stops or reshapes the work, and a stale request requires
publication of the current revision.

Registration or presentation is not approval. Do not implement code, mark a
direction or plan accepted, or claim implementation readiness.
