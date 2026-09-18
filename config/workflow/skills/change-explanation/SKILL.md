---
name: change-explanation
description: Build an evidence-grounded explanation of an implementation change or reviewed plan with selected code, abstractions, and visual lenses.
---

# Change Explanation

Use this skill when the reader needs to understand what changed without reading
the entire diff, or needs to understand a reviewed plan before implementation.
The result combines a compact narrative with the smallest visual forms that
make the implementation or proposal clear.

Read [references/graph-contract.md](references/graph-contract.md) before writing
`graph.json`.

## Scope and evidence

1. Resolve the target exactly: pull request, one commit, `base...head` range, or
   staged/unstaged/untracked working tree. Record the resolved refs and commit
   ids. For a pull request, retain its stated request, title, and description as
   intent evidence; do not treat them as proof of implementation.
2. Read repository instructions and inspect the diff stat, changed files,
   relevant tests, and the smallest surrounding source context needed to
   understand the behavior. Follow renamed files and important callers or
   consumers when the diff alone would misstate an abstraction.
3. Separate verified behavior, author-stated intent, and inference. Cite
   repository-relative paths and symbols. Never invent runtime behavior from a
   filename or ticket title.
4. For a working tree, describe the snapshot timestamp and whether evidence is
   staged, unstaged, or untracked. PR Lens provenance can name the nearest
   committed anchor, but the prose must say that the diagram also describes
   uncommitted evidence.

Do not include credentials, tokens, personal data, or private environment
values. A standalone explanation does not include a large raw patch. When
`/work-change-review` invokes this skill, that caller separately owns the
complete bounded approval patch and must stop rather than expose a secret.

## Explanation contract

Write a standalone `explanation.md` that includes, when supported by evidence:

- a 60-second summary and the user-visible or operator-visible outcome;
- requested intent versus verified implementation;
- the important behavior before and after the change;
- abstractions, modules, functions, data shapes, or boundaries added, removed,
  or modified, with concise descriptions;
- the smallest useful visual form, including an architecture diagram only when
  relationships matter and a data-flow diagram only when an ordered sequence
  changed;
- selected code examples that clarify a key contract or transition, normally
  no more than 20 lines per example and never the whole diff;
- compatibility, migration, operational, test, and uncertainty notes;
- an evidence index with the most useful files, symbols, tests, and revisions.

Use the reader's language and time budget when supplied. If `adapt-for-reader`
is installed, apply its resolved profile to the narrative only after the facts
are stable. Preserve evidence, caveats, and change meaning.

## Presentation preference

When a change diagram is useful, resolve its presentation once for the whole
run. Use this precedence, from highest to lowest:

1. **Explicit current request**, including `presentation=animated-web`,
   `presentation=mermaid`, legacy `presentation=static-tui`, or an equally
   clear request for animated web or Mermaid diagrams.
2. The project preference at
   `.maisternia/reader-profile.json` under
   `workflows.work-explain-change.presentation`.
3. The user preference at
   `${XDG_CONFIG_HOME:-~/.config}/maisternia/reader-profile.json` under the
   same key.
4. `mermaid` when no preference exists.

Validate a stored profile against `reader-profile.schema.json` before using
it. An invalid stored value does not become a preference; report it and fall
through to the next source. Do not ask merely because the preference is
absent, and do not persist a choice inferred from one request. Use
`/work-reader-preferences` when the user wants to save or change it.

Both modes start from the same inspected evidence and explain the same relevant
relationships or ordered interaction. Choose exactly one presentation in a
run:

- `animated-web` authors a PR Lens graph, embeds selected local SVG assets, and
  is intended for mdmaid.desk in a browser. Do not also include Mermaid
  versions.
- `mermaid` authors Mermaid directly in `explanation.md` so mdmaid and
  mdmaid.desk can render it in web and terminal readers. Do not create or link
  PR Lens assets for this mode. Treat `static-tui` as a backward-compatible
  alias and normalize it to `mermaid` before generation.

The preference changes diagram representation and viewing medium only. It never
changes the evidence set, diagram meaning, narrative claims, or review scope.

## Representation gate

First write a lens selection table with the question, supporting evidence,
status (`selected`, `not applicable`, or `unknown`), and output path. Choose the
smallest visual that answers each important question. Use several only when
they explain materially different aspects; do not create every form:

- pseudocode for an algorithm, decision, or state transition;
- a call tree for runtime control flow;
- a component tree for UI ownership, state, and module boundaries;
- a shallow file tree for responsibility moves or broad refactors;
- a diff-shaped tree or pseudocode block when the surrounding shape already
  exists and the delta is the point;
- a short complete code block when most of the unit is new, omitted context
  would hide ownership or order, or the reader needs a copyable target shape;
- a Mermaid `flowchart` for architecture, dependency, blast radius, or
  data-flow relationships;
- `classDiagram` for changed interfaces, types, implementations, composition,
  inheritance, or other UML-style static relationships;
- `erDiagram` for persisted entities, records, ownership, keys, and cardinality;
- `stateDiagram-v2` for lifecycle, approval, retry, mode, or status transitions;
- `sequenceDiagram` for ordered calls, messages, asynchronous work, errors, or
  cross-boundary interactions;
- `requirementDiagram` to trace requested behavior to changed elements and
  verification evidence when the selected renderer supports it, otherwise a
  traceability flowchart carrying the same relationships;
- PR Lens for system relationships, blast radius, or an ordered interaction
  whose motion materially improves understanding in `animated-web`;
- compact prose or `unknown` when the evidence cannot establish a relation.

Place each visual next to the text it supports. Remove incidental calls, files,
props, states, and boundaries. A small function-only change may need a
diff-shaped sketch and no PR Lens graph; say why the graph was omitted rather
than manufacturing architecture.

## Diagram generation boundary

PR Lens and Mermaid are visualization layers, not reviewers. Do not report findings
in a graph or diagram. Put risks, uncertainties, or review observations in
clearly labelled prose backed by evidence, and use `/work-review` when the user
asks for defect finding or merge advice.

For `animated-web`, author `graph.json` from the inspected evidence, validate
it, render it, and validate the rendered contract:

```text
pr-lens validate <run-dir>/graph.json
pr-lens render <run-dir>/graph.json --out <run-dir>/rendered --theme dark
pr-lens validate <run-dir>/rendered/manifest.json <run-dir>/rendered/drawn.graph.json
```

`validate` and `render` are local deterministic operations. Do not use
`pr-lens analyze` by default. It invokes another model provider and requires a
separate disclosure decision. Do not use `comment`, an asset host, Canvas, or
any upload in this workflow.

Inspect `rendered/manifest.json`. Select the highest useful architecture asset
and at most one changed data-flow asset. Link their exact relative SVG paths
from `explanation.md` with descriptive alt text. Do not guess the
content-addressed filename. The `animated` edges and messages should illuminate
the changed interaction, not decorate every connection.

For `mermaid`, author Mermaid directly from the inspected evidence. Select
among `flowchart`, `classDiagram`, `erDiagram`, `stateDiagram-v2`,
`sequenceDiagram`, and `requirementDiagram` using the lens table. Keep labels
concise, distinguish new, changed, removed, and necessary unchanged context
when the diff supports that distinction, and omit incidental nodes.
Save each diagram as `<run-dir>/<diagram-id>.mmd`, then verify terminal
rendering:

```text
mdmaid render-mermaid <run-dir>/<diagram-id>.mmd --backend beautiful-mermaid --width 120 --no-color
```

Inspect the rendered output, not only the exit code. If mdmaid reports a
renderer fallback or returns the fenced source instead of a diagram, retain the
native source for compatible web readers and create an equivalent supported
traceability flowchart for terminal readers. Record the fallback; do not claim
native terminal rendering for that diagram type.

Place the same source in a fenced `mermaid` block next to the prose it supports.
Keep the `.mmd` file as reproducible output, but do not add an SVG link or an
animated duplicate to the Markdown. This path has no dependency on a PR Lens
Mermaid converter or an unpublished PR Lens release.

Keep the selected diagram small enough to read. A PR Lens graph usually needs
only three or four lanes, changed nodes plus necessary unchanged neighbours,
and at most two hero edges; use equivalent restraint in Mermaid. A function or
class belongs in the diagram only when it represents an important changed
boundary; list the rest in the abstraction inventory.

## Plan evidence mode

When `/work-plan-review` reuses these lenses, treat plan interfaces,
relationships, requirements, and flows as proposed rather than verified.
Ground them in the complete reviewed plan and cite current repository evidence
separately. Never describe a planned class, entity, state, sequence, dependency,
or requirement as implemented. Embed selected Mermaid sources in the standalone
plan-review artifact and record why each other lens is not applicable.

## Verification

The bundle is complete only when:

- in change mode, resolved scope and diff agree; in plan evidence mode, the
  source plan hash and complete reviewed plan in the artifact agree;
- every material narrative claim has code, test, metadata, or explicit intent
  evidence;
- in `animated-web`, `pr-lens validate` accepts the source and drawn graphs and
  manifest, and every Markdown image refers to an asset named by
  `manifest.json`;
- in `mermaid`, each selected Mermaid diagram is present exactly once in the
  Markdown and renders through mdmaid's terminal backend;
- selected code is short, relevant, and non-sensitive;
- in `animated-web`, the text explains animation content for readers who
  cannot see motion;
- `mdmaid validate` accepts the final Markdown before desk registration.

If rendering fails, preserve the mode-specific source and diagnostics, correct
only evidence-supported contract errors, and retry. If a required tool remains
unavailable, deliver the narrative and exact blocker without pretending the
visual artifact exists.
