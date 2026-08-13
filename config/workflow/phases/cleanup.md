# /work-cleanup - Review Temporary Artifacts

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

List temporary prompts, outputs, rendered reports, worktrees, previews, and
generated staging files associated with the active task.

Input:

`$ARGUMENTS`

Show exact candidates, ownership, age, and whether each item is recoverable.
Delete nothing until the user explicitly approves the displayed list.

Never include repository source files, credentials, environment files, durable
task state, decisions, plans, contracts, or review records as automatic cleanup
candidates.
