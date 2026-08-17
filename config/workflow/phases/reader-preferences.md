---
name: work-reader-preferences
description: Calibrate durable readability preferences for audience, structure, density, tone, and delivery medium.
version: 0.1.0
---

# /work-reader-preferences - Calibrate Readability Preferences

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Create or revise reusable reader-output preferences from:

`$ARGUMENTS`

Use the installed `adapt-for-reader` skill, its preference reference, and the
installed `reader-profile.schema.json`.

First resolve whether the preferences are for the current session, the current
project, or the user. Then identify recurring situations and ask only the
highest-impact missing questions: desired view, conceptual depth, first-pass
time, density, terminology, answer position, evidence placement, visuals, and
whether to ask or infer when intent is ambiguous. Separately configure the
view-selection policy (`infer`, `ask-when-ambiguous`, or `always-ask`) and its
scope (`explicit-command` or `all-invocations`). Ask in short rounds and reuse
supplied answers.

Do not configure harness delegation in the reader profile. Direct the user to
`/work-routing-preferences`, which is the single routing system. If the existing
reader profile contains a legacy `delegation` object, offer a migration to the
`work-adapt-for-reader` route and show both diffs before any write.

Produce:

1. a plain-language preference summary;
2. general defaults;
3. situation overrides;
4. a schema-valid candidate profile when persistence is wanted;
5. the exact proposed destination and diff.

Do not write a profile or modify user/project instructions
without explicit approval of the exact scope and content. Never persist
inferred preferences, sensitive personal attributes, or a conversation
transcript. Inspect any existing destination before proposing an update, and
preserve unrelated user configuration.

If persistence is declined, return a session-only preference block that the
current conversation can apply.
