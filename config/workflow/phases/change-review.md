---
name: work-change-review
description: Explain and present the exact reviewed implementation in mdmaid.desk, then require a revision-bound human decision before publication or completion.
version: 0.2.0
---

# /work-change-review - Explain And Approve An Implementation

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

Present the exact implementation named in:

`$ARGUMENTS`

This is the mandatory human gate after implementation verification and
`/work-review` pass. It must not continue to `/work-pr`, claim delivery, or
report the implementation complete without a durable `approved`
`change-decision` for the exact implementation snapshot.

## Resolve And Freeze The Review Scope

Read repository instructions and the successful implementation `review.md` and
schema-valid `review.json`. Resolve base and head commits when the change is
committed. Include staged, unstaged, and untracked implementation files when
the working tree is the review target. Exclude ignored files and task-owned
review artifacts from the source snapshot.

Record a reproducible **change fingerprint** over the exact base/head or
working-tree snapshot. The fingerprint input must include changed paths,
statuses, rename targets, mode changes, binary markers, patch bytes, and hashes
of included untracked files. Record the algorithm and input classification;
never substitute a branch name, timestamp, prose summary, or document content
hash for the source fingerprint.

If the review report does not pass, fixes remain unapplied, verification is
stale, or the scope cannot be frozen without ambiguity, stop and return to the
appropriate review, run, or verify phase.

## Build One Explanatory Change Review

`/work-change-review` is a strict superset of `/work-explain-change`, not a raw
patch wrapper and not a second disconnected document. Use the installed
`change-explanation` skill against the frozen scope and place its explanation,
evidence, selected code, and diagrams in this same approval artifact:

```text
.agent-runs/change-reviews/<timestamp>-<change-id>/change-review.md
```

Give the first level-one heading a specific document title: the grounded
ticket ID when available, the task name, and the document purpose. For example,
`TASK-123 — Worker scaling: implementation review`. If there is no ticket ID,
start with the task name. Derive a concise task name from the reviewed change
when none was supplied. Pass this title explicitly to mdmaid.desk; the file
name `change-review` is only a document role.

The artifact must include:

- a 60-second summary and user-visible or operator-visible outcome;
- requested intent separated from verified implementation and inference;
- important behavior before and after the change;
- a changed-file inventory with add, modify, delete, rename, binary, and line
  statistics derived deterministically from Git;
- an abstraction inventory covering changed interfaces, types, functions,
  comments, data shapes, ownership boundaries, and important consumers;
- compatibility, migration, operations, tests, uncertainties, confirmed and
  refuted findings, applied fixes, and residual risks from `review.json`;
- selected short code examples and an evidence index;
- a diagram lens selection table that records each applicable lens, its
  evidence, generated diagram path, and why every omitted lens is not applicable;
- the complete reviewable textual patch in one or more bounded, standard Git
  fenced `diff` blocks.

The explanation must remain understandable without opening the native diff.
The diff remains the exact approval payload rather than an illustrative excerpt.

## Analyze The Diff Through Visual Lenses

Use Mermaid by default. An explicit `presentation=animated-web` may use PR Lens
for supported architecture or interaction views, but it does not replace the
Mermaid-only lens types or the native diff. Generate the smallest
evidence-complete portfolio; do not manufacture every diagram type.

Evaluate these lenses against changed interfaces and their necessary unchanged
neighbours:

- architecture, dependency, blast-radius, or data-flow relationships with a
  Mermaid `flowchart`;
- interface, type, class, implementation, composition, and inheritance changes
  with `classDiagram`;
- persisted entities, records, keys, ownership, and cardinality changes with
  `erDiagram`;
- lifecycle, status, mode, retry, or approval transition changes with
  `stateDiagram-v2`;
- ordered calls, messages, async work, errors, or cross-boundary interactions
  with `sequenceDiagram`;
- requested behavior traced to changed elements and verification evidence with
  `requirementDiagram`, with an equivalent traceability `flowchart` when the
  selected terminal renderer falls back to source.

Label nodes and prose as added, changed, removed, unchanged context, verified,
or inferred where the notation permits. A relation belongs in a diagram only
when supported by the diff, inspected source, tests, schema, or explicit intent.
Record `unknown` instead of inventing a relation. Verify every Mermaid source
with `mdmaid render-mermaid`; retain each `.mmd` file and embed its source once
in `change-review.md`.

## Preserve The Complete Native Diff

Patch blocks must cover every textual hunk in the frozen scope. Preserve
`diff --git`, file metadata, and hunk headers so mdmaid.desk can derive native
file and hunk navigation, line numbers, and stable hunk IDs. Include binary,
rename, and mode-only entries even when no textual hunk exists. Generate Git
paths with `core.quotePath=false` so the bounded parser receives readable UTF-8.

Bound each file and line and keep the validated artifact within mdmaid.desk
limits: at most 256 files, 2,048 hunks, 50,000 diff lines, 16 KiB per line, and
1,024 characters per path. If the exact change cannot fit, reduce or split the
implementation scope before requesting approval. Never silently truncate code.
If a patch would expose a secret, stop and remove the secret from the
implementation instead of redacting approval evidence.

## Validate And Publish The Gate

Require mdmaid 0.1.17 or newer and mdmaid-desk 0.1.19 or newer. Confirm that
mdmaid.desk help lists both `change-review` and `change-decision`, then validate:

```text
mdmaid validate <change-review.md> --json
```

Resolve the workspace by canonical repository root. Follow the installed
readable-output `references/project-naming.md` contract. Register the validated
artifact with an explicit revision-bound decision request:

```text
mdmaid-desk register <change-review.md> \
  --workspace <id> \
  --producer <current-provider> \
  --kind change-review \
  --title "<document title>" \
  --task <jira-id> \
  --feature-name "<minimal feature text>" \
  --attention approval \
  --expect change-decision \
  --request-message "Review this exact implementation before publication." \
  --json
```

Omit both project options when no grounded Jira ID exists.

Retain the document ID, document revision, review request ID, artifact content
hash, and change fingerprint. Keep the current turn open and wait:

```text
mdmaid-desk review wait <review-id> --json
```

Resume the same yielded process until it exits. `waiting_for_approval` is only
an intermediate update. Tell the human that `mdmaid-desk tui` opens the Changes
space and `c` selects it. Opening or reading the document is not approval.

## Revalidate And Route The Decision

After the waiter returns, verify the response document ID and revision, local
artifact content hash, and a freshly computed change fingerprint. Any mismatch
is `stale`, regardless of the button previously pressed.

- `approved`: preserve the response and identities. Continue to `/work-pr` only
  when publication was requested.
- `changes_requested`: preserve the summary and every structured response item.
  A feedback item names its file and stable hunk ID; a todo names its file.
  Return all anchors and messages to `/work-run`, then repeat verification,
  automated review, explanation, and human review for the new snapshot.
- `rejected`: stop, or return to analysis only when the human requests reshaping.
- `stale`: regenerate and revalidate the complete artifact.

The TUI action is **Request changes**. Do not treat checkbox-like prose as the
source of truth when structured response items exist. Missing tools, invalid
content, ambiguous scope, failed registration, an exited waiter without a
durable response, or any identity mismatch blocks the gate.

Return the artifact and diagram paths, resolved scope, change fingerprint,
validation result, mdmaid.desk document/review IDs, durable outcome, response
text, structured feedback anchors, and next phase.
