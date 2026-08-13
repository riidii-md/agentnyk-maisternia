---
name: work-brainstorm
description: Use after research and critical clarification to generate materially distinct options for a shape task.
version: 0.1.0
---

# /work-brainstorm - Generate Distinct Options

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Generate three to five materially different approaches for the current shape
task. Do not start while critical grill questions remain open.

Include:

- a minimal or reversible option;
- a conventional option;
- an ambitious option;
- an unconventional option when evidence supports one.

For each option record:

- intended outcome;
- supporting evidence and source IDs;
- assumptions;
- cost and dependencies;
- tradeoffs and likely failure modes;
- reversibility;
- unresolved questions.

Avoid cosmetic variations of one preferred answer. Write the result to the
task's `artifacts/brainstorm.md` and open it through `mdmaid.show` when
available.

Move to challenge when the options are genuinely distinct. Return with the
`weak-options` outcome when challenge shows that the option set is inadequate.
