# /work-plan - Create the Implementation Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Produce an actionable implementation specification for the accepted direction.

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
- Files and patterns to inspect
- Ordered implementation tasks
- Risk and blast-radius checks
- Migration or rollout concerns
- Expected verification
- Stop conditions
- Inputs for `/work prove`

Do not edit files.
