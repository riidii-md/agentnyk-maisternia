---
name: work-brainstorm
description: Use after research and critical clarification to generate materially distinct options for shaping work.
version: 0.2.0
---

# /work-brainstorm - Generate Distinct Options

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Generate three to five materially different approaches for the current shaping
discussion. Do not start while critical clarification questions remain open.

Include:

- a minimal or reversible option;
- a conventional option;
- an ambitious option;
- an unconventional option when evidence supports one.

For each option record:

- intended outcome;
- supporting evidence with links or concise references when available;
- assumptions;
- cost and dependencies;
- tradeoffs and likely failure modes;
- reversibility;
- unresolved questions.

Avoid cosmetic variations of one preferred answer. Return the options in the
current response. When the surrounding task authorizes a Markdown artifact,
use the installed `readable-output` skill to validate it and deliver it through
`mdmaid-desk`; presentation is not approval.

Move to challenge when the options are genuinely distinct. Revisit
brainstorming explicitly when challenge shows that the option set is
inadequate.
