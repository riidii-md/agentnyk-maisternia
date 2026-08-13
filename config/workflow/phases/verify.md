# /work-verify - Run Repository-Specific Checks

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Verify the current change using the repository's actual commands.

Input:

`$ARGUMENTS`

Discover build, format, lint, typecheck, unit, integration, end-to-end,
security, migration, and review gates from repository instructions, scripts,
CI, and hooks.

Run the smallest relevant checks first, then broaden with risk. Preserve full
logs. Classify failures as regression, environment, flaky, pre-existing, or
unknown.

Return commands, results, failure classification, remaining risk, log paths, and
whether independent review can begin.
