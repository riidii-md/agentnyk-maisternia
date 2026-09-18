---
name: work-explain-change
description: Explain a pull request, commit, revision range, or working-tree change through evidence, selected code, and Mermaid-first visual lenses.
version: 0.1.0
---

# /work-explain-change - Explain a Change

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

Explain the pull request, commit, revision range, or working tree named in:

`$ARGUMENTS`

Use the installed `change-explanation` skill. This is an understanding and
presentation workflow. It does not approve the change and does not replace
`/work-review`; if the user also wants findings or merge advice, run that as a
separate review contract.

Default to an engineering reader with about five minutes when no reader or
time budget is given. If the installed `adapt-for-reader` skill is available,
use it to shape density, conceptual depth, terminology, and ordering. It may
change presentation, but it must not change the evidence set or what the diff
means.

Resolve a diagram presentation once for the run. Precedence is:
**Explicit current request**, project
`workflows.work-explain-change.presentation` preference, user preference, then
the default `mermaid`. `animated-web` is the explicit PR Lens alternative;
`static-tui` is a backward-compatible alias for `mermaid`. Validate stored
profiles with `reader-profile.schema.json`; report and ignore invalid values.
Do not ask when no preference exists because the Mermaid default is defined.
Use exactly one presentation: fenced Mermaid for `mermaid`, or PR Lens animated
SVGs for `animated-web`, never both.

Keep all generated files under
`.agent-runs/change-explanations/<timestamp>-<change-id>/`. Do not alter source
files, the selected revisions, `.gitignore`, or remote state. Do not run
`pr-lens analyze` in the default workflow: it sends the diff to a separately
configured model provider. Do not upload assets, post a PR comment, or call
`pr-lens comment` unless the user explicitly requests that additional external
action.

The durable output is `explanation.md`. Evaluate architecture/data flow,
interfaces/classes, entity relationships, state, ordered interaction,
requirements traceability, and dependencies through the installed skill's
lens selection gate. Generate only evidence-supported diagrams. In
`animated-web`, accompany it with `graph.json`, validate the graph with
`pr-lens validate`, render it with `pr-lens render`, retain
`rendered/manifest.json`, `rendered/drawn.graph.json`, and the selected local
SVG assets, and inspect the manifest rather than predicting asset names. In
`mermaid`, author Mermaid directly from the same inspected evidence, retain
each `.mmd` source, and embed it once in a fenced Mermaid block. Do not create a
PR Lens graph solely for the Mermaid path. For a small local change, prefer a
compact code-shape visual and state why richer lenses were not applicable.

Before presentation, require mdmaid 0.1.17 or newer and mdmaid-desk 0.1.16 or
newer. Check both installed versions. Then run:

```text
mdmaid validate <artifact.md> --json
```

Treat validation as a hard gate. Fix invalid Markdown and revalidate. If a
compatible validator is unavailable or validation reports a runtime failure,
preserve the complete local bundle and do not register it.

Only after validation succeeds, resolve the current mdmaid.desk workspace from
`MDMAID_DESK_WORKSPACE` or by matching the canonical current project root in
`mdmaid-desk workspace list`. Add the workspace once when it is missing, using
a stable collision-safe id, then run:

```text
mdmaid-desk register <artifact.md> --workspace <id> --kind showcase --attention review
```

Registration is presentation, not approval. If mdmaid-desk is missing, older
than 0.1.16, or rejects the document or local media, preserve the bundle and
report an exact retry or upgrade command. Never claim that the animated view is
available until compatible registration succeeds.

Return a concise summary, the artifact and generated diagram paths, the
explained base and head/snapshot, chosen presentation, validation status, and
mdmaid.desk registration status. For `mermaid`, also return the exact
`mdmaid tui` command. Do not duplicate the full explanation in the terminal
unless the user asks for it.
