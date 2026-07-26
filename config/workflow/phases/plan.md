# /work-plan - Create the Implementation Plan

Produce an actionable implementation specification for the accepted direction.

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
