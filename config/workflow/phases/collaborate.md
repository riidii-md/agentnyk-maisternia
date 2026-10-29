---
name: work-collaborate
description: Discuss analysis, research, direction, or a plan with the human before advancing, using discussion, focused grill questions, or an optional human draft.
version: 0.1.0
---

# /work-collaborate - Develop the Work Together

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. Use its cleaned task and phase context;
the collaboration coordinator stays in the active conversation. Apply routes
only to eligible evidence work under its own phase contract.

Keep the coordinator in the current conversation. This checkpoint belongs to
the flow even when the agent has no unanswered questions. `/work-start` and
`/work-shape` invoke it automatically after analysis, after any research, and
when presenting direction or a plan. The human does not need a separate command.

Accepted forms:

```text
/work-collaborate <analyze|research|direction|plan> <context>
/work-collaborate <stage> discuss <context>
/work-collaborate <stage> grill <context>
/work-collaborate <stage> human-draft <context>
```

Infer the stage from the current conversation when it is unambiguous. Each
stage uses its own `<stage>-collaboration` graph node. Do not delegate the
human conversation or require a Maisternia runtime task or question queue.

## Present the Current Work

Read repository evidence, prior answers, and any selected artifacts first.
Resolve discoverable facts before asking the human. Keep one evolving discovery
brief for analysis, research, and attributed human input. Use separate stable
direction and plan artifacts. The coordinator may write these task-owned
artifacts; analysis and research themselves retain read-only authority.

Before each document-based invitation, screen sensitive material, finalize and
validate the exact revision, and register it through `readable-output` in
mdmaid.desk with `--attention review`. Require its successful receipt and local
content hash checks. Do not add `--expect` or a decision waiter for discovery
discussion or the AI-review choice. Present the safe clickable document link,
a short summary, and the useful choices at this stage:

- Analyze: correct the problem, scope, assumptions, acceptance criteria, or
  proposed research; add context; or continue.
- Research: question findings, compare tradeoffs, request more or broader
  research, describe the preferred solution, or continue.
- Direction: discuss the proposed shape and constraints, request changes or
  more evidence, then choose optional AI review and make the direction decision.
- Plan: discuss implementation and test coverage, request changes, then choose
  optional AI review and make the plan decision.

If validation or registration fails, report `delivery blocked`, preserve the
artifact, and follow readable-output's recovery contract. Do not claim the
checkpoint was delivered or advance past it.

## Collaborate Across Turns

Use `discuss` by default. Answer the human's questions and develop their ideas
with them. Do not manufacture questions when a simple invitation is sufficient.
The human may continue immediately without a discussion exercise.

Use `grill` when a material human constraint needs focused clarification. Apply
`/work-grill` to ask one useful question at a time. Answering a grill question
does not close collaboration. Resolve the answer, invite any further input,
and keep the checkpoint open until explicit continuation.

Use `human-draft` when the human wants to describe the desired behavior, user
experience, architecture, boundaries, non-negotiables, or acceptable compromises.
After research, offer this mode before direction; if research is unnecessary,
offer it at the analysis checkpoint. Accept rough notes or a conversation.
Preserve the result as attributed human input, separate from evidence and agent
recommendations. A human draft is not approval. Check conflicting facts with
the human; do not silently replace their intent. Do not require a separate draft
file or force the human to write a design.

Incorporate material answers and update the affected artifact. Register a changed
revision before asking for a response to that document. Questions, corrections,
drafts, and answers remain within collaboration; none imply permission to advance.
Record `human continues` only from an explicit continuation such as "continue"
or "go to direction". A reply may both answer a question and explicitly continue.
Silence, reading a document, or the agent having enough information is insufficient.

## Resume the Flow

When the human continues after analysis or research, advance to the applicable
next phase only when material constraints and evidence gaps are resolved. A
request to expand research authorizes that bounded research and returns to its
collaboration checkpoint. A changed problem or scope returns to analysis. Do
not repeat a completed checkpoint for the same unchanged evidence and explicit
continuation; a material revision reopens the affected checkpoint.

For direction and plan, combine discussion with the existing explicit AI-review
choice: run AI review now, skip AI review, or review later and pause. A selected
review or explicit skip can also express continuation of discussion. Only the
selected review authorizes reviewer workers. Then obtain the existing
authenticated, exact-revision direction or plan decision. Continuation, a draft,
and an AI-review skip never substitute for approval. Changed direction or plan
content returns to collaboration and a fresh AI-review choice; prior decisions
do not transfer to the changed revision.

Use the current conversation for replies today. Do not assume document chat
exists in mdmaid.desk. Future document conversation may carry the same human
input only when the selected tools explicitly support it. Maisternia configures
the workflow; the harness owns live discussion and task progress.

Only task-owned artifacts and scoped desk presentation are authorized here.
Do not modify target-project files, commit, push, publish unrelated content,
or broaden external-write authority.
