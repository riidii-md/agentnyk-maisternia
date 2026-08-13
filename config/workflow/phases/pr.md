# /work-pr - Check Pull Request Readiness

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Inspect the current branch and prepare an accurate PR handoff.

Input:

`$ARGUMENTS`

Discover provider, base branch, commit rules, title and body conventions, hooks,
CI requirements, and repository-specific checks.

Verify:

- staged, unstaged, and untracked files;
- accidental secrets or local configuration;
- commits and tracking status;
- checks actually run;
- documentation and migration notes;
- review findings and accepted risk.

Return a pass/fail checklist, exact fixes, suggested title and body, provider
command, and remaining approvals. Do not commit, push, or open a PR without
explicit authorization.
