---
name: herdr-add
description: Create or open a Git worktree in an exact Herdr session, using native repository grouping by default or an explicit independent isolated group.
version: 0.1.0
---

# /herdr-add - Add a Herdr Worktree

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Use the installed `herdr-worktrees` skill to handle:

`$ARGUMENTS`

The friendly native form is:

```text
/herdr-add <session> <worktree> <branch> [--parent-label <label>]
```

An independent named group requires explicit authority:

```text
/herdr-add <session> <worktree> <branch> --isolated-group <name>
```

Accept grounded conversational values, an explicit base ref, absolute checkout
path, child label, named fetch remote, focus choice, exact `--group-root`, and
`--editable`. Explicit arguments win. If a material repository, session,
remote, branch, path, group, or editability value remains ambiguous, ask one
concise question and do not mutate.

This command authorizes only the add/create/open behavior selected by the
request. Native mode may create or open a same-repository linked worktree.
Isolated mode may create or reuse an independent clone and its linked child
worktrees. It never authorizes cleanup, deletion, reset, dirty-state copying,
cross-session moves, or closing an existing workspace.

Load only the mode-specific reference selected by the request. Show the exact
redacted mutation preview and obtain exact confirmation for every non-no-op
mutation set as required by the shared skill before changing Git or Herdr.
Default to no focus and report only state verified after mutation.
