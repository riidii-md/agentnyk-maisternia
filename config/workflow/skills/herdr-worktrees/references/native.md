# Native Herdr Worktree Grouping

Use this reference for native add and regroup only. Every selected checkout must
belong to one Git common directory whose real primary worktree is a non-bare
checkout. A linked worktree cannot be the parent of sibling worktrees.

## Inventory and identity

1. Resolve the current or explicit repository with Git and canonicalize its
   common directory and every worktree path. Read `git worktree list` using its
   machine-readable NUL-delimited form.
2. Identify the primary checkout from Git topology, not from the current path,
   branch name, Herdr label, or sidebar position. Refuse bare repositories or a
   topology with no accessible primary checkout.
3. Query the exact named Herdr session. Match the primary by canonical checkout
   path and structured repository provenance. Existing non-empty `repo_key`
   evidence must agree with Git.
4. Resolve every selected child to one canonical linked checkout. For regroup,
   a missing checkout is a refusal with guidance to use `/herdr-add`.

For `--current`, require caller/session socket identity before resolving the
caller workspace. For `--path`, the canonical absolute checkout is authority.
For `--workspace`, resolve the structured ID inside the exact session and then
verify its checkout. For `--all-repo`, enumerate only linked checkouts from the
same common directory. A branch name is usable only when it selects exactly one
checkout.

## Native add

Before creation, validate the proposed branch with Git, confirm the base ref and
checkout path, and prove that neither the branch nor destination is owned by a
different worktree. Prefer an existing compatible checkout and open it instead
of creating a duplicate.

When the verified primary parent is absent from the exact live session, include
its creation in the redacted preview and confirmation. Create one workspace at
the canonical primary checkout with `workspace create --cwd <primary>
--no-focus`, capture the returned workspace ID, and reread the exact session.
Require the new workspace to retain the canonical primary path and a non-empty
repository key before creating or opening any child. If a conflicting parent
already exists, refuse; never close or replace it.

Create a missing linked checkout directly with Git under the common fail-closed
profile, then use Herdr worktree open with the verified parent workspace and
canonical path. When Git already owns a compatible linked checkout, skip
creation and open it. Capture returned IDs and do not depend on sidebar
ordering. Never use Herdr's server-side worktree create. For the exact audited
Herdr 0.8.0/protocol 19 profile, use only `worktree open --path`; its
server-side Git operations are limited to the pinned read-only discovery path
defined by the shared skill. A clean client environment alone is insufficient.
For any other build or operation without equivalent proof, refuse before the
Git mutation or Herdr action.

`--parent-label` may rename only the verified primary workspace and changes
display metadata only. Keep label outcome separate from grouping outcome.

## Regroup

Regroup never creates a Git checkout or branch. Freeze the exact selected
checkout paths, workspace IDs when present, primary parent ID, common directory,
and requested label into the preview. Revalidate each child immediately before
its open action and stop on drift.

If the child already has correct provenance, return a verified no-op. If Herdr
reports it as a plain workspace, adopt it only when live behavior proves that
open preserves the same workspace, tab, pane, terminal, and process identity.
Compare structured IDs before and after adoption; do not infer preservation
from a label or workspace ID alone. Herdr 0.8.0 must be tested rather than
assumed. When preservation is unsupported or uncertain, refuse; do not close
and recreate the workspace.

## Postconditions

After each action, require all of the following:

- the primary and child paths still match Git's inventory;
- the primary and child share the same non-empty Herdr `repo_key`;
- the child reports `is_linked_worktree: true`;
- reopening an already-open child reports `already_open: true` and retains its
  workspace, tab, pane, terminal, and process identity;
- a detached child, when intentionally supplied, retains `is_detached: true`;
- default behavior did not focus another workspace.

If any postcondition fails, stop on drift, reread both systems, and report the
actual partial state. Do not remove the checkout, close either workspace, or
attempt a label-based repair.
