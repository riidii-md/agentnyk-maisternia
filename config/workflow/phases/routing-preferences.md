---
name: work-routing-preferences
description: Configure explicit provider-neutral workflow routing preferences without silently delegating current work.
version: 0.1.0
---

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
ask me where to run adapt-for-reader every time
keep implementation in the current harness
use Codex and Claude for independent review
use Opus for plans in Claude and Sonnet for runs in Claude
```

Configure only the material fields: policy (`local`, `ask`, or `delegate`),
ordered harnesses, an optional multi-harness strategy, and optional per-harness
model choices. `models` keys are canonical harness IDs; a model choice does not
grant that harness or change authority. It opts that command into a fresh
model-selectable subagent while the current session remains coordinator.

When no narrow change was requested, offer guided setup. Inventory each installed canonical `/work-*` command except this preference command. For each
command, ask where it should run, then ask for the model in each specific
harness: a known eligible model, provider default, or no override. Ask in short
rounds, allow the user to reuse one answer across remaining commands, and show
the final command-by-command matrix before proposing persistence. Do not require
a model override for every command. If the user chooses a model for every
command, state clearly that every command will be subagent-backed.

Choose persistence after the matrix is known: session-only, user-global, or
repository-local; usually recommend user-global for command sets installed for
the user or intended across repositories. Recommend repository-local when the
commands were installed at project scope or the choice expresses a repository
constraint. An explicit user choice wins; never infer scope from the current
directory alone.

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
2. global defaults and workflow overrides, including per-harness model choices;
3. a schema-valid candidate profile when persistence is wanted;
4. the exact destination and diff;
5. any legacy migration proposed.

Do not write a profile, remove a legacy preference, or modify user/project
instructions without explicit approval of the exact scope and diff. Never
persist inferred preferences, credentials, sensitive context, or transcripts.
If persistence is declined, return a session-only routing block.
