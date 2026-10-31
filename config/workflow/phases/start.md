---
name: work-start
description: Start or resume standard work and advance through its phases until a human response or an external blocker is required.
version: 0.4.0
---

# /work-start - Guided Standard Work

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

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

## Git workspace preflight

Before task-specific workspace writes, identify the repository and run this
preflight; recheck the base and worktree before implementation. If this is
not a Git repository, record that these checks do not apply and continue under
the available workspace rules.

1. Determine the task's base branch from the user's instructions, repository
   rules, or an existing PR target. Otherwise inspect the remote default branch
   and use that branch (often `main` or `develop`). Do not assume that the
   current checkout is the right base. Fetch the chosen remote branch when
   network access permits. Compare its tip with the current task branch and
   report whether the task branch is current, ahead, behind, or diverged. If
   fetch fails or no remote exists, report that freshness is unverified; do not
   describe a cached ref as latest.
2. Use `git rev-parse --show-toplevel` and `git worktree list --porcelain` to
   verify that this task has a dedicated worktree and branch. Reuse an existing
   task worktree when its identity is clear. Otherwise create a uniquely named
   linked worktree and task branch from the selected base, within permitted
   paths, before editing task files. If only a local base ref is available,
   record its freshness limit. Do not treat the primary checkout or a
   worktree used by another task as dedicated to this task.
3. Preserve uncommitted changes in any checkout. Do not reset, discard, or
   silently move them. If relevant edits are in a different checkout or the
   task branch is behind or diverged, use the repository's safe update path;
   resolve any ambiguity with the human before changing shared history or
   moving work. Record the selected base, its verified tip or freshness limit,
   task branch, worktree path, and any unresolved blocker in task evidence.

## Persistent task summary

Read the installed `readable-output` skill's `references/task-context.md`
before saving or resuming task progress. The coordinator maintains the selected
`progress.md`, defaulting to `.agent-runs/work/<task-id>/progress.md`, after Git
preflight, after each completed phase and implementation step, before pausing,
and before context compaction when the harness gives advance notice. Follow the
shared contract for task identity, fields, privacy, safe writes, and recovery.
Report the path so another session can resume. The coordinator saves read-only
phase results outside those phases; brief and scout remain read-only.

## Advance the workflow

1. Read repository instructions, the task, current conversation, existing
   artifacts, the selected progress file, and installed phase contracts. Run the
   Git workspace preflight and identify completed work and the next unmet gate. Do not repeat a
   completed phase or infer a decision from a document's existence. Apply the
   persistent task summary's recovery rules, then refresh it at each checkpoint.
2. Gather facts through `/work-brief`, `/work-scout`, and `/work-analyze`.
   Enter `/work-collaborate analyze` after analysis, even when no human fact is
   missing. Research when evidence is missing or the human requests it, then
   enter `/work-collaborate research` after research. Classify the conditional
   architectural direction gate;
   record evidence for a skip or return to scout or research when impact is
   unknown. Record why expanded proof or handoff applies or can be skipped.
   Evidence may support an AI-review recommendation, but only the human selects
   or skips direction or plan AI review. Apply each phase contract in this
   session; do not merely print the next command or follow a fixed checklist.
3. At each discovery collaboration checkpoint, resolve what available evidence
   can answer first. Then write or update one durable discovery brief at a
   task-specific artifact path, otherwise under `.agent-runs/readable-output/`.
   Include the task, bounded evidence, assumptions, options when useful, the
   invitation to discuss, correct scope, expand research, provide a human draft,
   or continue. Include focused questions and what their answers change when
   needed. Before writing or presenting the brief, screen and redact
   credentials, tokens, raw logs, transcripts,
   private configuration, and sensitive source bodies. Summarize necessary
   evidence and cite safe source locations instead of copying raw content.
   Present the document through mdmaid.desk and wait for the human response.
   Use discussion by default, `/work-grill` inside collaboration for focused
   questions, and optional human-draft mode before direction. A question's answer
   keeps collaboration open. Advance only when the human continues explicitly;
   a research request returns to research and its checkpoint. A scope change
   returns to analysis. Reuse the same brief across discovery revisions.
4. When architectural direction is required, create and present the exact
   `/work-direction` revision through `/work-collaborate direction`. Discuss it
   until the human continues, combining continuation with an explicit AI review
   choice when the reply supplies both. AI review is optional. Ask whether to
   `run AI review now`, `skip AI review` so
   the human can inspect and decide on the current revision, or `review later`
   and pause. Do not automatically invoke `/work-plan-review`. Run
   `/work-plan-review direction` only after the first choice. After a passing
   review, or after recording an explicit AI review skip, wait for the direction
   decision and record it with `/work-decide direction` before detailed
   planning. Requested changes return to direction and the review choice; a
   stale revision also returns to that choice.
5. Create the `/work-plan` document and expand proof when needed. Carry direction's
   material assurance constraints and ensure the consolidated test plan is visible
   before presentation, including a justified no-test-change disposition when
   appropriate. Present the exact candidate plan through
   `/work-collaborate plan`, discuss it until the
   human continues, combining continuation with the same explicit choice when
   supplied together: `run AI review now`, `skip AI review`, or `review later`.
   AI review is optional. Do not
   automatically invoke `/work-plan-review`; run `/work-plan-review plan` only
   when selected. A skipped AI review does not approve the plan: wait for the
   human plan decision, record the skip and decision with `/work-decide plan`,
   and check `/work-ready` before implementation.
6. After approval, execute `/work-run` in the same session unless a fresh
   executor makes `/work-handoff` necessary. Read and carry the agreed implementation
   review cadence from the plan or current session under `/work-run`'s contract.
   Do not add a review after each task, phase, wave, or commit. If cadence is
   absent, ask before launching implementation review workers; continue independent
   authorized implementation and focused checks. At the agreed review boundary,
   continue through `/work-verify` and `/work-review` in repository-required gate
   order. Apply only the authorized repair batch, run its focused recheck,
   and report completion or remaining blockers. Return control for the next action;
   do not start another repair batch from a review result alone.
   Rerun only affected review obligations invalidated by fixes; do not start
   unrelated review cycles or change cadence without agreement.
7. Generate the mandatory `/work-human-review implementation-approval`
   document for the exact
   implementation snapshot. Present it with a revision-bound `change-decision`,
   then finish the turn at the pending human checkpoint. The desk retains the
   request; the agent does not poll for it. On an explicit resume, inspect the
   durable response once and continue. If changes are requested,
   return to run, verify, review, and change review. Prepare `/work-pr` only
   when publication was requested and the change decision is approved.

The delivery route is Git workspace preflight → brief → scout → analyze →
analysis collaboration → optional research and research collaboration →
conditional architectural direction → direction collaboration and explicit AI
review choice → optional direction review → direction decision →
plan → optional proof → plan collaboration and explicit AI review choice → optional plan review →
plan decision → readiness → optional handoff → run → verify →
implementation review → change review →
optional PR preparation. Failed verification or review returns to the relevant
earlier phase; requested changes return to the affected document or code phase.
Update this command whenever the `standard-work` delivery graph gains a gate.

## Human checkpoints

Generic continuation and status questions do not authorize review. Bind each
review to its explicit choice or agreed boundary, target, repair authority, and
exit. The default limit is one repair batch and one focused recheck. On completion
or a residual blocker, report the result and return control to the human for the
next action. Do not interpret a complaint about review as a new review request.
Follow `/work-review`'s review exit contract before dispatching more workers.

Collaboration is a required part of the flow after analysis and after research,
and accompanies direction and plan presentation. It applies even when the agent
has no questions. Follow `/work-collaborate` in the current conversation, retain
the human's draft as attributed input, and keep discussion open until the human
continues. Do not require another `/work-*` invocation. A review selection or
explicit skip may express continuation, but it never replaces a direction or
plan approval. Do not repeat an unchanged checkpoint whose explicit continuation
is already present in the conversation; material revisions reopen it.

Before a document-bound checkpoint, complete readable-output's successful
exact-revision registration and receipt checks. Deliver the stable discovery
brief before asking its question; a local link alone does not prove delivery.
At every required human checkpoint, provide the durable Markdown document or
the full absolute HTTP(S) mdmaid.desk document URL as a clickable Markdown link,
a short summary, the exact decision or question, and the consequence of each
response. Never substitute a bare route, relative path, document ID, or bare URL.
When the readable-output safe-origin contract cannot produce a link, report its
explicit safe-link-unavailable result and recovery guidance. Use the installed
`readable-output` and phase-specific review contract where applicable. Keep
one stable task-and-role artifact path across revisions. The discovery brief,
direction, plan, and change review are separate roles; do not create a file
for every internal phase.

At a direction or plan review-choice checkpoint, do not infer a preference from
risk, document complexity, or earlier use of `/work-start`. The human may run
AI review now, skip AI review and inspect the artifact themselves, or review
later. Only the first choice authorizes reviewer or candidate-refutation
workers. The skip is not approval and does not replace the later exact-revision
direction or plan decision.

For a live mdmaid.desk decision request, keep the current agent turn open and
wait in the foreground as the phase contract requires. If the harness instead
returns control at a conversational question, treat the next relevant user
answer as the response and resume this workflow in the same conversation. Do
not replace an unanswered required decision with elapsed time, a recommendation,
or a status message. Do not infer approval from silence, document registration,
opening a document, or a general request to start work. If an artifact or
implementation changed while waiting, mark the old decision stale and present
the current revision again.

Apply the installed `readable-output` quiet-wait contract to every live
mdmaid.desk decision without weakening its foreground-wait or approval rules.

On an answer, incorporate the human's words and update affected artifacts.
Discussion, grill answers, and drafts remain at the collaboration checkpoint
until explicit continuation; only then resume the next valid phase.
Return every changed direction or plan to its explicit AI review choice; do not
automatically rerun optional AI review. Rerun only mandatory implementation
verification or the authorized focused recheck invalidated by a repair, within
the same review episode. After delivery, resume only the human-selected next
phase. A request for changes loops to the appropriate phase;
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
