# /work-adapt-for-reader - Adapt Text to Its Reader

Routing gate (explicit adaptation): load `work-routing` because this deliberate command asks where to run when no preference exists. Automatic or nested reader adaptation remains local unless its caller already resolved a route. Continue only with the cleaned task.

Transform supplied text, referenced files, or current conversation context for
the reader and use described in:

`$ARGUMENTS`

Use the installed `adapt-for-reader` skill and its mode, preference, and design
principle references. Preserve meaning, evidence, uncertainty, constraints, and
source provenance.

Resolve preferences by the skill's precedence rules; the current request wins.
If the reader or intended use is missing and plausible choices
would materially change the output, ask one focused question:

> Who will use this text, and what should they be able to do after reading it?

Do not ask the reader-contract question when active instructions, a matching
situation preference, or the request already resolves it.

Resolve the profile's `view_selection` gate. An explicitly supplied view or mode
skips it. For `always-ask`, ask within its configured scope even when inference
is possible. For `ask-when-ambiguous`, ask only when plausible outcomes differ.
Offer: big picture, decision support, explanation, action brief, lookup, and
story/rationale; describe each in one line and mark the inferred choice as
recommended. Also resolve conceptual depth independently as high-level, working,
or deep.

The shared `work-routing` skill owns harness selection. An explicit `@harness`
route wins; otherwise it resolves the general workflow-routing profile and the
legacy reader delegation migration rule. Do not run a separate adaptation-only
delegation prompt. Configure durable execution choices with
`/work-routing-preferences`.

The coordinating harness must verify returned content and remains responsible
for the final Markdown artifact and mdmaid.desk registration.

Apply plain-language, accessibility, density, evidence, and visual preferences
as modifiers. Use tables or diagrams only when they materially reduce
comparison or relationship-building work.

Always write the complete standalone result to
`.agent-runs/readability/<timestamp>-adapted.md`; do not edit the supplied source
unless the user requested an edit. Resolve the current mdmaid.desk workspace
from `MDMAID_DESK_WORKSPACE` or by matching the canonical project root in
`mdmaid-desk workspace list`. When missing, add the current root once with a
stable collision-safe workspace ID.

Send the artifact to the desk with `mdmaid-desk register`, selecting the closest
document kind for the mode and using `--attention review`. Registration is a
presentation action, not approval. If mdmaid.desk is unavailable or rejects the
document, keep the Markdown and return its path plus an exact retry command.
Return only a short summary, the artifact path, and registration status in the
terminal unless the user explicitly asks for the full text inline.
