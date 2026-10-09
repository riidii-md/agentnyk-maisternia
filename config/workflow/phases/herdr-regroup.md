---
name: herdr-regroup
description: Repair Herdr provenance for existing linked worktrees inside one Git common directory without recreating or moving them.
version: 0.1.0
---

# /herdr-regroup - Repair Native Herdr Grouping

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Use the installed `herdr-worktrees` skill to handle:

`$ARGUMENTS`

Supported selectors are:

```text
/herdr-regroup <session> --current [--parent-label <label>]
/herdr-regroup <session> --path <absolute-checkout> ... [--parent-label <label>]
/herdr-regroup <session> --workspace <id> ... [--parent-label <label>]
/herdr-regroup <session> --all-repo [--parent-label <label>]
```

This command is restricted to existing linked worktrees in the same Git common directory
as their real non-bare primary checkout. It authorizes Herdr open or
provenance repair and an explicitly requested parent display-label change. It
never creates a Git checkout or branch, never creates an isolated group, and
never moves a live workspace across repository identities or sessions.

Resolve absolute paths and structured workspace identity before using labels or
branch names as hints. Freeze batch membership in the approved preview,
revalidate each item immediately before opening it, and stop on drift or first
failure. Default to no focus. If the requested destination is an isolated
group, refuse and explain that `/herdr-add --isolated-group` creates another
committed-state checkout and cannot preserve live workspace state.
