# /work-run - Execute the Approved Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Implement an approved handoff in small, restartable steps.

Input:

`$ARGUMENTS`

For each pass:

1. Read durable progress and select the first unfinished unblocked task.
2. Reconfirm the task against current code and repository rules.
3. Implement only that task.
4. Add or update focused tests.
5. Run the task criteria and smallest relevant checks.
6. Record result, files, checks, attempt count, and next action.
7. Repeat.

Park a task after three failed verification attempts. Stop when three tasks are
parked, all remaining work is blocked, or the configured run cap is reached.

Do not commit, push, open a PR, weaken criteria, touch production data, or
perform destructive actions without explicit authorization.
