# /work-scout - Gather Facts

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Collect evidence before analysis or planning.

Input:

`$ARGUMENTS`

Read repository instructions, relevant code, CI, hooks, build files, ticket
content, URLs, logs, and existing task artifacts. Separate discovered facts from
inferences. Do not choose a final solution or edit files.

Return:

- Handoff summary
- Facts with source paths or URLs
- Unknowns
- Scope boundaries
- Likely affected subsystems
- Missing evidence
- Recommended next phase
