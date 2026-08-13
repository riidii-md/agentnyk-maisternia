# /work-routing-preferences - Configure Workflow Routing

Create, revise, or migrate reusable harness-routing preferences from:

`$ARGUMENTS`

Use the installed `work-routing` skill and
`work-routing-profile.schema.json`. This configuration command always runs in
the current harness; a leading `@harness` names a preference target rather than
delegating the configuration task.

Resolve whether the preference is session-only, project, or user scoped. Accept
plain language such as:

```text
always run plans in Codex
ask me where to run reviews
keep implementation in the current harness
use Codex and Claude for independent review
```

Configure only the material fields: policy (`local`, `ask`, or `delegate`),
ordered harnesses, and an optional multi-harness strategy. Reuse answers already
provided and ask in short rounds.

Use these persistent destinations:

```text
project: <repository>/.maisternia/work-routing.json
user:    ${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json
```

Inspect an existing destination before proposing an update and preserve
unrelated entries. If a reader profile contains the legacy `delegation` field,
offer a migration to the `work-adapt-for-reader` workflow entry and show the
reader-profile removal separately.

Return:

1. a plain-language routing summary;
2. global defaults and workflow overrides;
3. a schema-valid candidate profile when persistence is wanted;
4. the exact destination and diff;
5. any legacy migration proposed.

Do not write a profile, remove a legacy preference, or modify user/project
instructions without explicit approval of the exact scope and diff. Never
persist inferred preferences, credentials, sensitive context, or transcripts.
If persistence is declined, return a session-only routing block.
