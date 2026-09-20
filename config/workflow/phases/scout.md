---
name: work-scout
description: Gather repository, environment, and task facts with read-only inspection before analysis or implementation.
version: 0.2.0
---

# /work-scout - Gather Facts

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Collect evidence before analysis or planning.

Input:

`$ARGUMENTS`

Read repository instructions, relevant code, CI, hooks, build files, ticket
content, URLs, logs, and existing task artifacts. Follow concrete references
into potentially affected adjacent systems: repositories, services, interfaces,
data stores, deployment and operational paths, compatibility consumers,
ownership, sources of truth, and trust boundaries.

Keep cross-system discovery bounded by the task and evidence. Inspect a
referenced system only when code, configuration, documentation, deployment
metadata, a ticket, or the user links it to the task. Do not search unrelated repositories,
the whole machine, or unbounded external systems speculatively.
Record inaccessible systems as unknown rather than claiming they are unaffected.
Separate discovered facts from inferences. Do not choose a final solution or
edit files.

Return:

- Handoff summary
- Facts with source paths or URLs
- Unknowns
- Scope boundaries
- Likely affected subsystems
- Cross-system context map: ownership, interfaces, data or control flows,
  trust boundaries, and operational dependencies
- Systems checked and evidence-supported affected, unaffected, or unknown status
- Missing evidence
- Recommended next phase
