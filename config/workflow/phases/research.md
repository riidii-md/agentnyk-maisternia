# /work-research - Compare Solution Directions

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Research viable solution families before detailed planning.

When `work-routing` resolves several harnesses, run independent research lanes,
verify factual conflicts against primary evidence, and let the current harness
synthesize the recommendation without treating consensus as proof.

Input:

`$ARGUMENTS`

Inspect existing repository patterns first. Use current primary documentation
when external behavior matters. Compare two to four real options when meaningful
choice exists.

Evaluate:

- scope and complexity;
- correctness and testability;
- security;
- migration and deployment impact;
- maintainability;
- user impact;
- reversibility.

Do not edit files.

Return options, recommendation, rejected directions, risks, user decisions, and
whether the task is ready for a decision.
