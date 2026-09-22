---
name: work-change-review
description: Resolve whether reviewed changes need an implementation-approval gate or unified pull-request feedback, then deliver the exact review to its selected destination.
version: 0.3.0
---

# /work-change-review - Decide And Deliver A Change Review

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

Input:

`$ARGUMENTS`

Accepted modes:

```text
/work-change-review implementation-approval <implementation snapshot>
/work-change-review pr-feedback --destination internal <PR or review artifacts>
/work-change-review pr-feedback --destination pr <PR or review artifacts>
```

`implementation-approval` is the mandatory human gate after implementation
verification and a repair-disposition `/work-review` pass. `pr-feedback`
challenges and delivers the unified findings from an existing pull request
review. These outcomes are not interchangeable: approving an implementation
is not approving a PR review comment, and publishing comments is not approving
the implementation or merge.

When the caller or current standard-work phase explicitly supplies the mode and
destination, preserve it. Otherwise, if the invocation could mean either
internal repair or PR publication, ask where the review should land and wait.
Do not infer the destination merely because a PR exists. A direct request to
review an existing PR without complete intake should continue through
`/work-start-pr-review` after that answer so the ticket, PR, comments, current
head, and CI evidence are synchronized before this final phase.

## PR-feedback mode

Require schema-valid report-only review artifacts for the exact PR head and the
complete existing-comment inventory produced by `/work-start-pr-review`.
`--destination internal` returns the confirmed findings to internal repair and
does not write to the provider. `--destination pr` means PR publication and
authorizes posting only the final review comments described here.

Review the full PR diff against its declared merge target, such as `main` or
`develop`. Resolve the target from live PR metadata, fetch it when possible,
and use the target/head merge-base so unrelated target-branch movement is not
misclassified as part of the PR. Never narrow the review to the last commit,
the last rework, or only the hunks changed since an earlier review. Earlier
review state helps classify comments; it does not redefine the source diff.

Independently challenge the unified draft against the current diff, necessary
unchanged context, ticket or accepted behavior, tests, CI, and earlier review
threads. Refute false positives, merge duplicates, remove stale or already
addressed findings, and preserve material disagreement. A short analysis detour
may propose additional findings, but include them only after grounding and
independent verification. Never add filler comments.

For every surviving finding, select the most useful destination:

- use line comments for a precise issue anchored to a current changed line;
- use a summary comment for cross-cutting findings, context outside changed
  lines, prior-thread disposition, residual risk, and the overall result;
- use both only when the summary adds coordination value without duplicating
  the line comments.

Each comment states the problem, impact, concrete evidence, and smallest useful
next action. Keep provider/model attribution in the review artifact without
turning the posted review into a transcript. When there are no confirmed
findings, publish a concise grounded no-findings summary if PR publication was
selected.

Before publication, compare the source visibility of every ticket, project,
CI, log, or artifact fact used in the draft with the destination PR audience.
Private evidence may support internal verification but must not be quoted,
linked, summarized, or inferentially disclosed to a broader audience. Screen
the exact outbound line-comment and summary bodies and anchors for secrets,
credentials, PII, private links, and restricted context; rewrite or withhold
unsafe evidence while preserving a useful public finding.

Immediately before PR publication, refresh the PR head, full diff, relevant
comments, and required checks or CI. If evidence used by the draft changed,
mark the draft stale and return to the affected review and synthesis steps.
Bind the provider review submission and line comments to the expected head SHA
through a revision precondition when supported. Without one, perform the
narrowest possible final head check immediately before writing and record the
weaker guarantee.

Prefer one provider review submission when it can atomically contain the
selected line comments and summary comment. Otherwise post deterministically,
record each remote response ID and URL, and retry only items proven not to have
been created. After writing, perform post-publication reconciliation: re-read
the live head, required checks, created review, and relevant threads. Exclude
the workflow's own newly created response IDs from concurrent-comment drift.
If the head or checks changed, or a created item is not attached to the expected
head and anchor, report the result as stale or partial and return to context
synchronization without automatically reposting. Do not approve, request
changes, merge, close, label, assign, or modify code as part of this mode.

Return the reviewed head SHA, unified confirmed and refuted findings, prior
comment disposition, exact posted bodies and anchors or internal-repair handoff,
provider response IDs/URLs, failures, and residual risk.

## Implementation-approval mode

Present the exact implementation named in the arguments. This mode must not
continue to `/work-pr`, claim delivery, or report the implementation complete
without a durable `approved` `change-decision` for the exact implementation
snapshot.

## Resolve And Freeze The Review Scope

Read repository instructions and the successful implementation `review.md` and
schema-valid `review.json`. Resolve the actual merge target from the user's
instruction, existing PR base, or trusted repository evidence; otherwise use
the remote default branch, commonly `main` or `develop`. Fetch the target when
network access permits and record whether it is current. For committed work,
resolve the merge-base between that target and the reviewed head and review the
complete merge-base-to-head change. Never substitute `HEAD^`, the last commit,
the last rework, or the previously reviewed delta for the merge-target diff.

When the working tree is the review target, start with the complete
merge-base-to-`HEAD` change and also include staged, unstaged, and untracked
implementation files. Exclude ignored files and task-owned review artifacts
from the source snapshot. If the merge target is ambiguous, stop and ask rather
than selecting the current branch or most recent commit as the base.

Record a reproducible **change fingerprint** over the exact base/head or
working-tree snapshot. The fingerprint input must include changed paths,
statuses, rename targets, mode changes, binary markers, patch bytes, and hashes
of included untracked files. It must also record the merge-target ref and SHA,
merge-base SHA, and head SHA. Record the algorithm and input classification;
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
