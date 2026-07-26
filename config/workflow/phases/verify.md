# /work-verify - Run Repository-Specific Checks

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
