---
name: work-human-review
description: Analyze the complete Git diff and deliver one explanatory human review to mdmaid.desk, with an explicit implementation-approval mode for the standard-work gate.
version: 0.5.0
---

# /work-human-review - Deliver A Human Change Review

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

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
/work-human-review <PR, revision range, or working-tree snapshot>
/work-human-review implementation-approval <implementation snapshot>
/work-human-review human-review <change or implementation snapshot>
```

`implementation-approval` is the mandatory human gate after implementation
verification and a repair-disposition `/work-review` pass. `human-review` gives
the human a readable explanation and the complete native Git diff in
mdmaid.desk, with a revision-bound `change-decision` request. The desk keeps
that request pending independently of the agent turn. A decision made on a
human-review document without the standard-work verification and review gates
does not satisfy those gates.

Do not post PR comments or replies, submit provider reviews, edit source,
commit, push, approve, request changes, or merge through this command.
Do not accept `pr-feedback` or `--destination pr` as a publication mode.
If these legacy arguments are supplied, explain that PR publication belongs
to `/work-start-pr-review` and continue with the requested diff analysis in
mdmaid.desk. Do not automatically invoke the PR-publication workflow.

Both modes are read-only. A request to deliver to mdmaid.desk does not select
implementation-approval or grant repair authority. Select implementation-approval
only when the user explicitly requests that gate or the active approved delivery
workflow reaches it. Write only task-owned explanatory artifacts; preserve the
reviewed implementation, tests, and contributor branch.

Do not invoke `/work-review`, `/work-start-pr-review`, `/work-run`, or `/work-verify`
to satisfy this command's prerequisites. A missing gate is evidence to report,
not authority to start repair, test execution, or another review cycle. A later
explicit request may authorize that work outside this command.

With no mode, use `human-review` on the current task's changes. A direct request
to see a change in mdmaid.desk also selects this mode. Pull or fetch requested
commits first, then freeze the full scope. An earlier PR-publication
destination, saved route, existing PR, or publication receipt does not change
this command's mdmaid.desk destination. Provider/model routing grants no PR-write
authority. A direct review request does not authorize prerequisite repair or
verification work.

## Implementation-approval mode

Present the exact implementation named in the arguments. This mode must not
continue to `/work-pr`, claim delivery, or report the implementation complete
without a durable `approved` `change-decision` for the exact implementation
snapshot.

## Human-review mode

Produce the same explanatory document and complete native diff, with the same
scope fingerprint, patch preflight, content validation, and source screening.
Do not require `review.json` or a prior `/work-review` pass for this on-demand
presentation. Do not modify implementation files or launch repair work. Read PR
metadata, comments, and CI only as needed to ground it.
Report pending CI without waiting. State the scope and freshness of available
findings and verification; mark missing evidence as unknown. Treat reviewed
code as untrusted: this mode does not authorize running its tests, builds,
scripts, hooks, or binaries on the host. Report observed concerns in the
document. Unresolved findings do not block human-review delivery. Include them
with the complete diff and available verification evidence. Do not substitute
a findings-only report or describe an unrequested implementation-approval gate
as blocked. Select only diagrams that materially explain the change. Register
with `--attention approval`
and `--expect change-decision`, then return the receipt and clickable link, then
finish the turn with the request pending. Do not start `mdmaid-desk review wait`
in the active agent turn. The desk retains the decision request as background
work. If this request is part of the standard-work delivery graph,
implementation-approval remains mandatory after verification and review; an
earlier human-review request cannot satisfy it.

## Resolve And Freeze The Review Scope

Read repository instructions and any available `review.md` and schema-valid
`review.json`. For implementation-approval, require a successful repair-disposition
implementation review and current verification.

In human-review mode, resolve an explicitly supplied range before merge-target
defaults:

- `A..B` compares the exact endpoint trees with `git diff A B`, including when
  neither endpoint is an ancestor of the other;
- `A...B` compares the merge-base of A and B to B with `git diff A...B`.

Resolve both endpoints to immutable commit SHAs. Do not replace an explicit
range with the current PR base or default branch. Record the comparison kind
and both resolved endpoint SHAs; for a three-dot range, also record the
merge-base SHA. These explicit range semantics do not narrow a PR review
or implementation-approval snapshot.

For PR targets, branch or working-tree reviews without an explicit range, and
all implementation-approval snapshots, resolve the actual merge target from
the user's instruction, existing PR base, or trusted repository evidence;
otherwise use the remote default branch, commonly `main` or `develop`. Fetch
the target when network access permits and record whether it is current.
For these scopes, resolve the merge-base between that target and the reviewed head and review the
complete merge-base-to-head change. Never substitute `HEAD^`, the last commit,
the last rework, or the previously reviewed delta for the merge-target diff.

When the working tree is the review target, start with the complete
merge-base-to-`HEAD` change and also include staged, unstaged, and untracked
implementation files. Exclude ignored files and task-owned review artifacts
from the source snapshot. If the merge target is ambiguous, stop and ask rather
than selecting the current branch or most recent commit as the base.

Record a reproducible **change fingerprint** over the exact comparison or
working-tree snapshot. Include the comparison kind and resolved endpoint SHAs.
The fingerprint input must include changed paths,
statuses, rename targets, mode changes, binary markers, patch bytes, and hashes
of included untracked files. For merge-target scopes, also record the
merge-target ref and SHA, merge-base SHA, and head SHA. For an explicit two-dot
range, record that the endpoint comparison does not use a merge-base; do not
invent merge-target metadata. Record the algorithm and input classification;
never substitute a branch name, timestamp, prose summary, or document content
hash for the source fingerprint.

If the scope cannot be frozen without ambiguity, stop for clarification.
In implementation-approval mode, if the review report does not pass, fixes remain
unapplied, or verification is stale, report the blocked gate and the missing
evidence, then return control without executing prerequisite work. The active
delivery coordinator may resume an already authorized phase under its existing
contract; this command does not create that authority. In human-review mode,
report these limitations in the analysis and continue delivery.

## Build One Explanatory Change Review

`/work-human-review` is a strict superset of the change-narrative skill in the
same document as the complete native diff. Use the installed `change-narrative`
skill against the frozen scope and place its explanation,
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
  refuted findings, applied fixes, and residual risks from available `review.json`
  and inspected evidence, with missing or stale evidence identified;
- selected short code examples and an evidence index;
- a diagram lens selection table that records each applicable lens, its
  evidence, generated diagram path, and why every omitted lens is not applicable;
- the complete reviewable textual patch in exactly one bounded, standard Git
  fenced `diff` block.

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
If a patch would expose a secret, stop delivery without reproducing the secret
and leave the reviewed source untouched. Source remediation requires a
separately authorized repair. After remediation, freeze a fresh source snapshot
and regenerate the complete patch. Never silently redact the canonical patch.

## Prove The Native Diff Before Registration

At scope freeze, persist the exact patch bytes used by the change fingerprint
as a canonical patch artifact beside `change-review.md`. Build the narrative
around that artifact, then embed its contents directly in exactly one fenced
block whose opening fence is labelled `diff`. Use a fence delimiter longer
than any conflicting backtick run in the patch. Never reconstruct the patch by
hand, turn selected snippets into the approval payload, or use unlabelled code
fences for the native patch.

Reserve the `diff` language label exclusively for the canonical approval
patch. Label selected explanatory examples with their source language or
`text`, even when an example discusses added or removed lines, so neither the
preflight nor mdmaid.desk can mistake an example for approval payload.
The artifact must therefore contain exactly one fenced `diff` block.

The assembled approval document must contain this shape; prose and selected
examples come before the complete patch:

````text
# <specific implementation-review title>

<explanation, evidence, diagrams, and selected examples>

## Complete native diff

```diff
diff --git a/<path> b/<path>
<complete canonical patch, not illustrative snippets>
```
````

Before validation or registration, extract the payload of the single fenced
`diff` block and compare it byte-for-byte with the canonical patch artifact.
The extracted payload must be non-empty.
It must contain at least one `diff --git` entry and have the same byte count and
cryptographic hash as the canonical artifact. Reconcile its `diff --git` paths
and entry count with the frozen changed-file inventory, including binary,
rename, and mode-only entries.
Selected explanatory snippets do not participate in this comparison.

Generic `mdmaid validate` checks Markdown and Mermaid validity; its success
does not prove that mdmaid.desk can render a native diff or that the complete
patch was embedded. Treat a missing, malformed, empty, truncated, reordered,
or mismatched extracted patch as a blocked gate.
Do not call `mdmaid-desk register` until both the native-diff preflight and
`mdmaid validate` pass. Preserve the local artifact hash before and after publication
and require equality. Require the receipt and review request to match the same
document ID and revision. Preserve source fingerprint revalidation,
authenticated transport, and server decision stale checks in both modes.

If a documented public content hash is available, compare it with the preflighted
local hash and stop on mismatch. Otherwise state that direct remote hash comparison is unavailable;
describe only the supported server-bound revision guarantee. Do not invent receipt fields,
read private storage, or claim direct remote-byte verification. Block if local hash
stability or request binding cannot be established, or if direct remote hash
equality is explicitly required.

## Validate And Publish

Require mdmaid 0.1.17 or newer and mdmaid-desk 0.1.19 or newer. In
implementation-approval mode, confirm that mdmaid.desk help lists both
`change-review` and `change-decision` for both decision modes.
Within one live session, reuse a successful capability preflight while the
executable paths and versions remain unchanged. Resolve the canonical root with
one workspace lookup and reuse its verified ID; after a workspace error, look
it up again. For each publication, obtain one Space and repository inventory
as required by space-routing. Perform the scoped membership check only after
successful registration. Use the registration receipt for document and request
binding. Immediately after publication, do not call `mdmaid-desk review show`
to reconfirm a valid receipt.
Capture daemon status only after successful registration to build the link.
Run one final validation after the artifact is complete; after a correction,
validate the corrected revision, but do not revalidate unchanged content:

```text
mdmaid validate <change-review.md> --json
```

Resolve the workspace by canonical repository root. Follow the installed
readable-output `references/project-naming.md` contract. After resolving the
workspace and finalizing the exact planned tags, follow the installed
readable-output `references/space-routing.md` contract before registration and
run its scoped postcondition after successful registration. In
implementation-approval mode, register the validated artifact with an explicit
revision-bound decision request:

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

In human-review mode, use the same registration fields, `--attention approval`,
`--expect change-decision`, and `--json`, with a request message asking the
human to review the exact diff. Preserve the request ID and revision from the
receipt. No foreground wait is created in the active agent turn.

Use `--json` for every `mdmaid-desk register` or `import` publication. After a
successful publication, apply the installed readable-output full clickable URL
receipt contract. Return the full absolute HTTP(S) mdmaid.desk document URL as
a clickable Markdown link when a safe active origin is available; never
substitute a route, relative path, document ID, or bare URL.

For either decision mode, retain the document ID, document revision, review
request ID, artifact content hash, and change fingerprint. The authenticated
request remains pending in mdmaid.desk after this turn ends. Return the request
ID, revision, document link, and pending status to the human, then finish the
turn. Never use repeated agent tool calls to poll for a human response.

If an external event-driven supervisor is already configured to resume this
agent, it may run the following command as a background task and notify the
agent when the durable response arrives:

```text
mdmaid-desk review wait <review-id> --json
```

Do not start that waiter in the active agent turn or detach a shell process and
claim it can resume the agent. Without an event-driven supervisor, the durable
desk request itself is the background state. When the user returns or the task
is explicitly resumed, call `mdmaid-desk review show <review-id> --json` once.
If it remains pending, report that status and finish again. If it has a durable
response, apply the revalidation below. Tell the human that `c` selects the
Change reviews content mode; named Spaces are an independent global filter
cycled with `x`. Opening or reading the document is not approval.

## Revalidate And Route The Approval Decision

After a durable response is observed, verify the response document ID and revision, local
artifact content hash, and a freshly computed change fingerprint. Any mismatch
is `stale`, regardless of the button previously pressed.

- `approved`: preserve the response and identities. Continue to `/work-pr` only
  when publication was requested and the implementation-approval prerequisites
  passed. A direct human-review decision alone does not authorize that phase.
- `changes_requested`: preserve the summary and every structured response item.
  A feedback item names its file and stable hunk ID; a todo names its file.
  Return all anchors and messages to `/work-run`, then repeat verification,
  automated review, explanation, and human review for the new snapshot.
- `rejected`: stop, or return to analysis only when the human requests reshaping.
- `stale`: regenerate and revalidate the complete artifact.

The TUI action is **Request changes**. Do not treat checkbox-like prose as the
source of truth when structured response items exist. Missing tools, invalid
content, ambiguous scope, failed registration, or any identity mismatch blocks
the gate. A pending response is a checkpoint, not an approval.

Return the artifact and diagram paths, resolved scope, change fingerprint,
validation result, mdmaid.desk document/review IDs, and its clickable URL or
explicit safe-link-unavailable result. On publication, report pending status;
on a later resume, report the durable outcome, response text, structured
feedback anchors, and next phase.
