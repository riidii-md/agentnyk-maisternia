---
name: work-brief
description: Produce a concise task and session refresher from current evidence, decisions, progress, and remaining work.
version: 0.1.0
---

# /work-brief - Quick Task and Session Refresher

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Give a concise plain-language reminder of what the active task or session is
about, where it stands, and what happened so far. The result must be useful to
a reader who has not seen the current conversation.

Input:

`$ARGUMENTS`

Use the current conversation, durable task state, event history, repository,
branch, Git status, recent relevant commits, and explicitly referenced
artifacts. Infer a task identifier from arguments, branch, path, or conversation
when possible. Separate facts from assumptions and use read-only checks only.

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
