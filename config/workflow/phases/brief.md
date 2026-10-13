---
name: work-brief
description: Produce a concise task and session refresher from current evidence, decisions, progress, and remaining work.
version: 0.2.0
---

# /work-brief - Quick Task and Session Refresher

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Give a concise plain-language reminder of what the active task or session is
about, where it stands, and what happened so far. The result must be useful to
a reader who has not seen the current conversation.

Input:

`$ARGUMENTS`

Use the current conversation, repository, branch, Git status, recent relevant
commits, and explicitly referenced artifacts. Infer a task identifier from
arguments, branch, path, or conversation when possible. Separate facts from
assumptions and use read-only checks only; do not require an external task-state
store.

Read the task's selected `progress.md` when present. Locate it from an explicit
path, existing task artifacts, or the matching task ID under `.agent-runs/work/`.
If several files could match, resolve task identity before selecting one. Treat
the saved summary as context, not approval. Use read-only inspection to
revalidate the task identity, branch, worktree, current files, and referenced
decisions. Use the installed `readable-output` skill's `references/task-context.md`
for recovery safeguards. Mark unsupported or stale status as unknown and report
the next evidence check. Include the progress path and relevant project document sources
in the refresher. Do not update the summary; the coordinator owns persistence.

Return:

```text
Ticket: <ticket, inferred short name, or "unknown">
What this is about: <2-4 simple sentences>
Current status: <not started / investigating / planned / implementing / verifying / blocked / ready for review, with why>
What happened so far:
- <3-6 short events>
Important files/places:
- <only the most relevant items>
Next step: <one concrete action>
Open questions: <none or 1-3 questions>
```

Keep the answer readable in under one minute. Do not edit files, change external
state, expose secrets, or dump raw logs.
