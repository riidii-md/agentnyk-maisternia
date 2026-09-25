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

## Advance the workflow

1. Read repository instructions, the task, current conversation, existing
   artifacts, and installed phase contracts. Run the Git workspace preflight,
   identify work already completed and the next unmet gate. Do not repeat a
   completed phase or infer a decision from a document's existence.
2. Gather facts through `/work-brief`, `/work-scout`, and `/work-analyze`.
   Research when evidence is missing and use `/work-grill` when focused human
   context is needed. Classify the conditional architectural direction gate;
   record evidence for a skip or return to scout or research when impact is
   unknown. Record why expanded proof or handoff applies or can be skipped.
   Evidence may support an AI-review recommendation, but only the human selects
   or skips direction or plan AI review. Apply each phase contract in this
   session; do not merely print the next command or follow a fixed checklist.
3. When a material human fact is needed, resolve what available evidence can
   answer first. Then write or update one durable discovery brief at a
   task-specific artifact path, otherwise under `.agent-runs/readable-output/`.
   Include the task, bounded evidence, assumptions, options when useful, the
   exact question, and what its answer changes. Before writing or presenting
   the brief, screen and redact credentials, tokens, raw logs, transcripts,
   private configuration, and sensitive source bodies. Summarize necessary
   evidence and cite safe source locations instead of copying raw content.
   Present the document and wait for the human response.
4. When architectural direction is required, create and present the exact
   `/work-direction` revision, then stop for an explicit AI review choice. AI
   review is optional. Ask whether to `run AI review now`, `skip AI review` so
   the human can inspect and decide on the current revision, or `review later`
   and pause. Do not automatically invoke `/work-plan-review`. Run
   `/work-plan-review direction` only after the first choice. After a passing
   review, or after recording an explicit AI review skip, wait for the direction
   decision and record it with `/work-decide direction` before detailed
   planning. Requested changes return to direction and the review choice; a
   stale revision also returns to that choice.
5. Create the `/work-plan` document and expand proof when needed. Present the
   exact candidate plan, then stop for the same explicit choice: `run AI review
   now`, `skip AI review`, or `review later`. AI review is optional. Do not
   automatically invoke `/work-plan-review`; run `/work-plan-review plan` only
   when selected. A skipped AI review does not approve the plan: wait for the
   human plan decision, record the skip and decision with `/work-decide plan`,
   and check `/work-ready` before implementation.
6. After approval, execute `/work-run` in the same session unless a fresh
   executor makes `/work-handoff` necessary. Continue through `/work-verify`
   and `/work-review`. Apply requested fixes and repeat affected verification
   and review phases until they pass or a real blocker remains.
7. Generate the mandatory `/work-change-review implementation-approval`
   document for the exact
   implementation snapshot. Present it, wait for the human response, and
   preserve its revision-bound `change-decision`. If changes are requested,
   return to run, verify, review, and change review. Prepare `/work-pr` only
   when publication was requested and the change decision is approved.

The delivery route is Git workspace preflight → brief → scout → analyze →
optional research or human question → conditional architectural direction →
explicit AI review choice → optional direction review → direction decision →
plan → optional proof → explicit AI review choice → optional plan review →
plan decision → readiness → optional handoff → run → verify →
implementation review → change review →
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

On an answer, incorporate the human's words and update affected artifacts.
Return every changed direction or plan to its explicit AI review choice; do not
automatically rerun optional AI review. Rerun only mandatory implementation
verification or review invalidated by an implementation change, then resume
from the next valid phase. A request for changes loops to the appropriate phase;
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
