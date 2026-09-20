---
name: work-start
description: Start or resume standard work and advance through its phases until a human response or an external blocker is required.
version: 0.1.0
---

# /work-start - Guided Standard Work

Routing gate (lazy): load `work-routing` for phase-scoped preferences only when
`$ARGUMENTS` has a plausible explicit route, an active session route exists,
or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. Use the router to resolve the cleaned task
and phase preferences. Do not delegate `/work-start` as a whole: its coordinator
must remain in the active conversation to collect human responses and advance
the workflow. Apply a requested route only to an eligible bounded phase, under
that phase's authority and runner contract. If the route cannot support a
required phase, report the mismatch and ask before changing the route.

Input:

`$ARGUMENTS`

Start the `standard-work` delivery workflow for the task in the arguments or
current conversation. Keep the active harness and same conversation as the
coordinator. This command is a continuing workflow: after each completed
phase, automatically select and perform the next applicable phase until a
human response, missing external capability, or terminal outcome stops
progress. The user does not need to invoke each `/work-*` command or repeat
`/work-start` after replying. A later explicit `/work-start` may resume the
same task from its evidence.

## Advance the workflow

1. Read repository instructions, the task, current conversation, existing
   artifacts, and installed phase contracts. Identify work already completed
   and the next unmet gate. Do not repeat a completed phase or infer a decision
   from a document's existence.
2. Gather facts through `/work-brief`, `/work-scout`, and `/work-analyze`.
   Research when evidence is missing and use `/work-grill` when focused human
   context is needed. Classify the conditional architectural direction gate;
   record evidence for a skip or return to scout or research when impact is
   unknown. Record why expanded proof, plan review, or handoff applies or can
   be skipped. Apply each phase contract in this session; do not merely print
   the next command or follow a fixed checklist.
3. When a material human fact is needed, resolve what available evidence can
   answer first. Then write or update one durable discovery brief at a
   task-specific artifact path, otherwise under `.agent-runs/readable-output/`.
   Include the task, bounded evidence, assumptions, options when useful, the
   exact question, and what its answer changes. Before writing or presenting
   the brief, screen and redact credentials, tokens, raw logs, transcripts,
   private configuration, and sensitive source bodies. Summarize necessary
   evidence and cite safe source locations instead of copying raw content.
   Present the document and wait for the human response.
4. When architectural direction is required, create the `/work-direction`
   document, run `/work-plan-review direction`, and present the exact reviewed
   revision. Wait for an explicit direction decision and record it with
   `/work-decide direction` before detailed planning. Requested changes return
   to direction and review; a stale revision requires a fresh review.
5. Create the `/work-plan` document, expand proof when needed, and run the
   applicable `/work-plan-review plan` path. Present the exact reviewed revision
   for a plan decision. Wait for the human response, record it with
   `/work-decide plan`, and check `/work-ready` before implementation.
6. After approval, execute `/work-run` in the same session unless a fresh
   executor makes `/work-handoff` necessary. Continue through `/work-verify`
   and `/work-review`. Apply requested fixes and repeat affected verification
   and review phases until they pass or a real blocker remains.
7. Generate the mandatory `/work-change-review` document for the exact
   implementation snapshot. Present it, wait for the human response, and
   preserve its revision-bound `change-decision`. If changes are requested,
   return to run, verify, review, and change review. Prepare `/work-pr` only
   when publication was requested and the change decision is approved.

The delivery route is brief → scout → analyze → optional research or human
question → conditional architectural direction, direction review, and direction
decision → plan → optional proof and plan review → plan decision → readiness →
optional handoff → run → verify → implementation review → change review →
optional PR preparation. Failed verification or review returns to the relevant
earlier phase; requested changes return to the affected document or code phase.
Update this command whenever the `standard-work` delivery graph gains a gate.

## Human checkpoints

At every required human checkpoint, provide the durable Markdown document or
validated mdmaid.desk review link, a short summary, the exact decision or
question, and the consequence of each response. Use the installed
`readable-output` and phase-specific review contract where applicable. Keep
one stable task-and-role artifact path across revisions. The discovery brief,
direction, plan, and change review are separate roles; do not create a file
for every internal phase.

For a live mdmaid.desk decision request, keep the current agent turn open and
wait in the foreground as the phase contract requires. If the harness instead
returns control at a conversational question, treat the next relevant user
answer as the response and resume this workflow in the same conversation. Do
not replace an unanswered required decision with elapsed time, a recommendation,
or a status message. Do not infer approval from silence, document registration,
opening a document, or a general request to start work. If an artifact or
implementation changed while waiting, mark the old decision stale and present
the current revision again.

On an answer, incorporate the human's words, update affected artifacts, rerun
their required review when content changed, and automatically resume from the
next valid phase. A request for changes loops to the appropriate phase;
rejection stops the task unless the human explicitly requests reshaping, in
which case return to analysis. Cancellation stops the task. Ask only for
information or authority that cannot be derived from existing evidence. Do all
authorized, reviewable work before requesting a decision.

## Boundaries and output

Respect phase authority and repository rules. Do not plan in detail before an
approved direction when its gate applies. Do not implement before the approved
plan and readiness gate. Do not publish, claim delivery, or complete
before the approved exact-snapshot change decision. Commit, push, PR, deploy,
and other external writes require their own authority when applicable.
Maisternia configures this command; the active harness owns session state and
execution. Do not create a Maisternia runtime task or approval queue.

At a pause, report the current phase, completed evidence, artifact link and
revision, the response needed, and the phase that will resume. At completion,
report the delivered result, verification, review decisions, and any remaining
limitations. Do not stop merely to recommend the next routine phase.
