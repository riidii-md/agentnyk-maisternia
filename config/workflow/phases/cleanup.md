# /work-cleanup - Review Temporary Artifacts

List temporary prompts, outputs, rendered reports, worktrees, previews, and
generated staging files associated with the active task.

Input:

`$ARGUMENTS`

Show exact candidates, ownership, age, and whether each item is recoverable.
Delete nothing until the user explicitly approves the displayed list.

Never include repository source files, credentials, environment files, durable
task state, decisions, plans, contracts, or review records as automatic cleanup
candidates.
