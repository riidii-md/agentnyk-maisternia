---
name: work-grill
description: Use when shaping work needs focused human clarification before options can converge.
version: 0.2.0
---

# /work-grill - Ask the Next Useful Question

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Run focused clarification in the current conversation.

Before asking anything:

1. read the repository, supplied sources, and answers already present in the
   conversation or explicit artifacts;
2. resolve questions that available evidence can answer;
3. identify the single unanswered question with the highest decision value.

Ask only that question and explain why its answer matters. Accept a direct
answer, deferral, unknown, research request, or rejection of the premise.
Interpret the human reply in the next turn without requiring a separate queue
or recording command.

Do not repeat an answered question unless new evidence directly conflicts with
the answer. Critical open questions block convergence, but they do not require
an external task-state service.
