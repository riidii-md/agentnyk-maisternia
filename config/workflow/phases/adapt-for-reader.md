---
name: work-adapt-for-reader
description: Adapt existing text to a specific reader, purpose, medium, and time budget while preserving meaning and evidence.
version: 0.1.0
---

# /work-adapt-for-reader - Adapt Text to Its Reader

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

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
legacy reader delegation migration rule. With no route or routing preference,
run locally without asking where. If preference resolution later discovers a
legacy reader-profile `delegation` object, load `work-routing` then as a
compatibility migration path. Do not run a separate adaptation-only delegation
prompt. Configure durable choices, including **ask every time for this
workflow**, with `/work-routing-preferences`.

The coordinating harness must verify returned content and remains responsible
for the final Markdown artifact and mdmaid.desk registration.

Apply plain-language, accessibility, density, evidence, and visual preferences
as modifiers. Use tables or diagrams only when they materially reduce
comparison or relationship-building work.

Always write the complete standalone result to
`.agent-runs/readability/<timestamp>-adapted.md`; do not edit the supplied source
unless the user requested an edit. Before resolving or writing to mdmaid.desk,
require mdmaid 0.1.17 or newer and check `mdmaid --version`. An older CLI treats
`validate` as a filename, so do not treat that failure as invalid document
content. If mdmaid is missing or older, preserve the Markdown artifact, do not
register it, and report `npm install --global mdmaid@0.1.17` as the exact
upgrade command.

With a compatible version, run `mdmaid validate <artifact.md> --json`. Treat
validation as a hard gate: exit 0 may continue; invalid content at exit 1 must
be fixed using the source-located diagnostics and revalidated until exit 0;
exit 2 reports validation runtime unavailable and blocks registration. If
validation never reaches exit 0, do not register. Preserve the Markdown
artifact and return its path, the blocker, and the exact validation retry
command.

Only after validation succeeds, resolve the current mdmaid.desk workspace from
`MDMAID_DESK_WORKSPACE` or by matching the canonical project root in
`mdmaid-desk workspace list`. When missing, add the current root once with a
stable collision-safe workspace ID.

Use the final document's first level-one heading, without the Markdown marker,
as the semantic `--title`; if no level-one heading exists, derive a concise
title from the reader contract. Never use the timestamped filename as the
catalog title. Send the artifact to the desk with
`mdmaid-desk register <artifact.md>`, selecting the closest document kind for
the mode and using `--attention review`. Add `--task <id>` when the source has
an explicit stable task ID. Add up to three short lowercase `--tag <tag>` values
only for grounded subject matter not already represented by workspace, task,
or kind; never tag timestamps, filenames, workspace IDs, document kinds, or
storage modes. Registration is a presentation action, not approval. If
mdmaid.desk is unavailable or rejects the document, preserve the Markdown
artifact and return its path plus an exact retry command.
Return only a short summary, the artifact path, and registration status in the
terminal unless the user explicitly asks for the full text inline.
