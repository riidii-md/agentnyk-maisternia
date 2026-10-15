# Mandatory Human Change Review

Plain `/work-human-review` analyzes the complete Git diff and publishes the
explanation and native patch to mdmaid.desk with a revision-bound decision
request. It can report missing or stale verification without claiming that the
standard-work implementation gate has passed. An existing PR or earlier
publication destination cannot redirect this command to PR comments or replies.
Both modes preserve source and tests. Unresolved findings do not block the
complete explanatory artifact or justify a findings-only report. Failed or
pending CI remains evidence in the complete diff document and does not turn
delivery into a CI wait or a short blocker-only report.

`/work-human-review implementation-approval` is the required implementation
gate between successful automated review and PR preparation. It publishes one
first-class mdmaid.desk Change Review containing both a readable explanation
and the exact native diff.

`/work-start-pr-review` owns unified PR-feedback synthesis and publication when
the user selects that destination. The old `pr-feedback` and `--destination pr`
arguments do not authorize PR writes through `/work-human-review`; they produce
the desk analysis and identify the separate PR workflow. Posting feedback never
constitutes implementation or merge approval.

A direct `/work-human-review` request uses `human-review` mode. It still embeds
the complete Git diff and creates a revision-bound desk decision request.
It can run without a prior automated implementation review; the document says
which checks have not been established. Standard-work still requires its
verification and automated review gates before an approval permits PR work.

## One artifact, two reading depths

The Change Review is a strict superset of the internal `change-narrative` skill. Its narrative
contains the summary, intent versus verified behavior, before/after model,
changed abstractions and interfaces, selected code, compatibility and
operational notes, tests, findings, risks, diagrams, and evidence index. The
same document contains the complete bounded Git patch that drives the native
diff viewer, file and hunk navigation, and anchored feedback.

The reader can understand the design first and then inspect exact lines without
switching to a disconnected explanation document.

The producer also keeps a canonical patch artifact beside the review and
compares the payload extracted from the document's single fenced `diff` block
with that patch byte-for-byte before registration. Generic Markdown/Mermaid
validation is not evidence that the native diff viewer has a complete patch.
A missing, hand-written, truncated, or mismatched patch blocks registration;
selected code snippets never substitute for the approval payload and never use
the `diff` fence label reserved for that payload.

## Diff visual lenses

Mermaid is the default. The workflow evaluates, but does not blindly generate:

- architecture, dependency, blast-radius, and data-flow charts;
- UML-style interface and class diagrams;
- entity-relationship diagrams for persisted data;
- state diagrams for lifecycle and approval transitions;
- sequence diagrams for ordered interactions;
- requirement diagrams tracing intent to changed elements and tests.

Each visual is grounded in diff, source, schema, test, or explicit intent
evidence. Omitted class, entity-relationship, state, sequence, requirement, or
data-flow lenses are marked not applicable or unknown. Diagrams explain; they
do not independently discover findings.

## Exact decision boundary

The producer fingerprints the exact source snapshot, validates the Markdown,
registers it as `kind=change-review`, and requests a revision-bound
`change-decision`. The desk holds the pending request after the agent reports
the link and finishes its turn. An event-driven supervisor may wait for the
decision in the background and resume the agent. Without one, the agent checks
the durable request once when the user resumes the task. Repeated agent polling
is unnecessary. Approval applies only while the document revision, content
hash, and source fingerprint still match. Any implementation change makes the
decision stale.

Explicit human-review ranges preserve their requested comparison: `A..B` compares
the exact endpoint trees, while `A...B` compares their merge-base to B. The source
fingerprint records that comparison kind and its resolved commit SHAs.

PR targets, branch or working-tree reviews without an explicit range, and every
implementation-approval snapshot retain the complete change against the actual
merge target: the PR base or selected destination branch such as `main` or
`develop`. For those scopes, the workflow resolves the target/head merge-base
and never substitutes the last commit, latest rework, or delta since a previous
review.

If the exact patch would expose a secret, delivery stops without reproducing
it or editing the reviewed source. Source remediation needs a separately
authorized repair. After that repair, the producer freezes a fresh snapshot
and regenerates its full patch rather than silently redacting evidence.

The human may approve, Request Changes with file or hunk anchors, or reject.
Only approval permits `/work-pr`. The gate remains read-only: mdmaid.desk does
not stage, edit, or revert source files.

## Visual plan review

Plans use the same vocabulary through `/work-plan-review`. Its standalone
`plan-review.md` contains the complete reviewed plan, high-level interfaces and
abstractions, selected Mermaid diagrams, review findings, and a `plan-decision`.
Plan diagrams label relationships as proposed rather than verified. This avoids
inventing implementation while still making architecture, data, state,
sequence, dependency, and requirement relationships reviewable before coding.
