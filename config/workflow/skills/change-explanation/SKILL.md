---
name: change-explanation
description: Build an evidence-grounded explanation of a pull request, commit, range, or working-tree change with selected code and PR Lens architecture or data-flow visuals.
version: 0.1.0
---

# Change Explanation

Use this skill when the reader needs to understand what changed without reading
the entire diff. The result combines a compact narrative with the smallest
visual forms that make the change's shape clear.

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

Do not include credentials, tokens, personal data, private environment values,
or large raw patches in any artifact. Redact sensitive literals while
preserving their technical meaning.

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

## Representation gate

Choose the smallest visual that answers each important question. Use several
only when they explain materially different aspects; do not create every form:

- pseudocode for an algorithm, decision, or state transition;
- a call tree for runtime control flow;
- a component tree for UI ownership, state, and module boundaries;
- a shallow file tree for responsibility moves or broad refactors;
- a diff-shaped tree or pseudocode block when the surrounding shape already
  exists and the delta is the point;
- a short complete code block when most of the unit is new, omitted context
  would hide ownership or order, or the reader needs a copyable target shape;
- PR Lens for system relationships, blast radius, or an ordered interaction
  whose motion materially improves understanding.

Place each visual next to the text it supports. Remove incidental calls, files,
props, states, and boundaries. A small function-only change may need a
diff-shaped sketch and no PR Lens graph; say why the graph was omitted rather
than manufacturing architecture.

## PR Lens boundary

PR Lens is the visualization layer, not the reviewer. Do not report findings in
the graph. Put risks, uncertainties, or review observations in clearly labelled
prose backed by evidence, and use `/work-review` when the user asks for defect
finding or merge advice.

When the representation gate selects PR Lens, author `graph.json` from the
inspected evidence, then run:

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

Keep the graph small enough to read: usually three or four lanes, changed nodes
plus necessary unchanged neighbours, and at most two hero edges. A function or
class belongs in the diagram only when it represents an important changed
boundary; list the rest in the abstraction inventory.

## Verification

The bundle is complete only when:

- resolved scope and diff agree;
- every material narrative claim has code, test, metadata, or explicit intent
  evidence;
- when PR Lens was selected, `pr-lens validate` accepts the authored and drawn
  graph documents and manifest;
- every generated Markdown image refers to an asset named by `manifest.json`;
- selected code is short, relevant, and non-sensitive;
- the text explains animation content for readers who cannot see motion;
- `mdmaid validate` accepts the final Markdown before desk registration.

If rendering fails, preserve `graph.json` and the diagnostics, correct only
evidence-supported contract errors, and retry. If a required tool remains
unavailable, deliver the narrative and exact blocker without pretending the
animated artifact exists.
