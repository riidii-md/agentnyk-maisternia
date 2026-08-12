# /work-reader-preferences - Calibrate Readability Preferences

Create or revise reusable reader-output preferences from:

`$ARGUMENTS`

Use the installed `adapt-for-reader` skill, its preference reference, and the
installed `reader-profile.schema.json`.

First resolve whether the preferences are for the current session, the current
project, or the user. Then identify recurring situations and ask only the
highest-impact missing questions: reader task, first-pass time, density,
terminology, answer position, evidence placement, visuals, and whether to ask
or infer when intent is ambiguous. Ask in short rounds and reuse supplied
answers.

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
