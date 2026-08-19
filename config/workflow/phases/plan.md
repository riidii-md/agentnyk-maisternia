---
name: work-plan
description: Create a reviewable implementation proposal with ordered changes, decisions, acceptance evidence, risks, and verification gates.
version: 0.2.0
---

# /work-plan - Create the Implementation Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Produce an actionable implementation proposal. An earlier direction decision
may constrain it, but the implementation proposal itself is not approved until
the human reviews the final revision and makes an explicit decision.

When `work-routing` resolves several harnesses, request independent plans and
let the current coordinating harness synthesize one plan while preserving
material disagreements and unsupported assumptions.

Input:

`$ARGUMENTS`

Discover repository rules before assuming paths, base branches, ticket formats,
tools, tests, or PR conventions.

Plan in dependency order. Each task should describe observable behavior, fit one
focused implementation loop, and keep the repository runnable. Identify a thin
end-to-end slice first when appropriate.

Return:

- Discovered repository rules
- Scope and exclusions
- Proposed direction and rationale
- Material alternatives and tradeoffs
- Files and patterns to inspect
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
the plan through mdmaid.desk with attention `approval`, then report
`waiting_for_approval`.

Registration or presentation is not approval. Do not implement code, mark the
direction accepted, or claim implementation readiness.
