# /work-pr - Check Pull Request Readiness

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
