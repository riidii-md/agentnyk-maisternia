# /work-ready - Readiness Gate

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Evaluate whether the task can safely move to planning or implementation.

Input:

`$ARGUMENTS`

Check:

- task statement exists;
- facts and assumptions are separated;
- scope and exclusions are clear;
- acceptance criteria exist;
- solution direction is decided;
- repository rules are known or explicitly unknown;
- important risks are resolved or accepted;
- required user approval exists.

Return pass, conditional pass, or fail with exact missing inputs and the next
phase. Do not proceed past unresolved critical ambiguity.
