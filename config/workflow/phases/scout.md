---
name: work-scout
description: Gather repository, environment, and task facts with read-only inspection before analysis or implementation.
version: 0.3.0
---

# /work-scout - Gather Facts

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

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

## Project document map

Find existing equivalents for product goals, current architecture, design and
interaction rules, security, code conventions, and testing. Start with the
repository root instructions and documentation index. Apply relevant nested
instructions to their directory scope. Do not require fixed filenames such as
`PRD.md` or `DESIGN_SYSTEM.md`. Code, component libraries, ADRs, and trusted
project references may supply the context.

For each topic, return its source path or URL, relevant sections, task relevance,
and status: current, proposed, historical, missing, inaccessible, or contradicted
by inspected evidence. Cite the evidence for a contradiction. A document's name
or date alone does not prove that it describes current behavior. Separate the
current implementation from approved intent. Preserve unknowns when verification
is unavailable; do not invent product decisions or silently resolve conflicts.

Do not read every document. Read the relevant sections and follow references
only when the task requires them. Discovery remains read-only. Return the map
to the coordinator for inclusion in an existing task artifact or progress file;
do not create or update project documents during scout.

Return:

- Handoff summary
- Facts with source paths or URLs
- Project document map
- Unknowns
- Scope boundaries
- Likely affected subsystems
- Cross-system context map: ownership, interfaces, data or control flows,
  trust boundaries, and operational dependencies
- Systems checked and evidence-supported affected, unaffected, or unknown status
- Missing evidence
- Recommended next phase
