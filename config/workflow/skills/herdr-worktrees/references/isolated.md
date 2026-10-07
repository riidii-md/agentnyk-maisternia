# Isolated Named Herdr Groups

Use this reference only for `/herdr-add --isolated-group`. An isolated group is
an independent non-bare clone with its own Git common directory and top-level
Herdr parent. It duplicates committed repository state; it is not regrouping
and cannot preserve a live workspace, pane, process, index, or uncommitted file.

## Resolve deterministic storage

The default namespace is:

```text
${XDG_DATA_HOME:-$HOME/.local/share}/maisternia/herdr-groups/v1/
  <source-hash>/<session-hash>/<group-slug>-<group-hash>/
    repo/
    worktrees/<worktree-slug>-<worktree-hash>/
```

Hash the canonical source common-directory bytes, exact session bytes, exact
group-name bytes, and exact worktree-name bytes separately with
`git hash-object --stdin`; feed the raw bytes with no added line terminator and
retain each full digest. Slugs are sanitized ASCII readability aids only, with
`group` or `worktree` fallback. Never hash ambiguous concatenated fields.

`--group-root` names the exact container that holds `repo/` and `worktrees/`.
It does not name a higher-level storage root.

Canonicalize the proposed root and inspect every existing component without
following symlinks. The only permitted descendant relationships are the exact
layout above: `repo/` and `worktrees/` beneath the approved group root, and each
deterministic child beneath that exact `worktrees/` directory. Refuse every
other symlink, alias, non-directory component, wrong owner, unsafe permission,
unexpected occupant, or ancestor/descendant overlap with the source repository,
its common directory, any source worktree, any unrelated destination checkout
or repository, or Herdr/config state. Recheck immediately before the first
write and stop if identity changed.

## Select the remote and freeze refs

Select exactly one named fetch remote. Use `origin` only when unambiguous;
otherwise require `--remote`. Keep its URL opaque in memory. Never print,
normalize, interpolate, or place it in a preview. The validated credential-free
URL may persist only as the selected remote URL inside the private isolated
clone so compatibility can be recognized later; never place it in reports,
logs, previews, task artifacts, or any other file. Disable shell tracing and use
the common fail-closed Git profile with `GIT_TERMINAL_PROMPT=0`. Reject a remote
name or URL with a leading hyphen, NUL, newline, carriage return, control byte,
or option-like shape. Insert `--` before every untrusted operand when that Git
subcommand supports it. Refuse detectable URL user information, password/token
material, query, or fragment. Reject option injection strings such as `-q`,
`--all`, or `--upload-pack=` as data, never as flags.

Capture and discard raw stderr from URL-bearing Git operations. Report a stable
failure category without replaying Git output. This includes failures from a
remote containing a harmless redaction-test marker.

Validate branch input with `git check-ref-format --branch`. Before confirmation,
query only with non-mutating `git ls-remote --symref` under the controlled
profile; do not fetch into the source repository or update `FETCH_HEAD`,
remote-tracking refs, or its object database. When no base is supplied, resolve
the remote symbolic HEAD from that query and use its target as the deterministic
default base. If remote HEAD is absent, ambiguous, or not a returned commit,
require an explicit base and stop without mutation. Resolve the base and
requested remote branch to exact commit OIDs and freeze the remote name, ref
names, and OIDs in the redacted preview. Repeat the same non-mutating query and
compare immediately before the confirmed clone or worktree mutation. Stop for
ref movement.

Version 1 accepts committed remote refs only. If the requested source branch is
dirty, ahead, local-only, or would omit source work, refuse. Do not block on an
unrelated dirty branch and never alter it.

## Create an independent parent

Capability-probe the local Git version before relying on clone flags. Prefer an
accelerated remote clone that combines `--no-local`, `--reference-if-able`,
`--dissociate`, `--no-checkout`, and the selected origin name. If acceleration
is unavailable, use an ordinary remote clone with `--no-local` and
`--no-checkout`. A shared clone, retained object alternates, hardlink dependency,
source hook/config copy, push-URL copy, or remote URL rewrite is forbidden.

Use the verified empty template directory and disabled hooks/fsmonitor settings
for clone and every later Git or Herdr action. Inspect effective attributes
before materializing files; if an external filter or
clean/smudge/process driver would run, refuse without checkout unless
separately authorized for that exact program.

After cloning, require a non-bare repository, no `objects/info/alternates`, and
complete object connectivity. Resolve the clone's fetched base and branch refs
and require them to equal the confirmed OIDs before any checkout. Detach the
primary checkout at the exact confirmed base OID and require it to be clean.
This primary checkout is the real Herdr top-level parent; do not create a
synthetic group branch or a child review-base worktree.

The detached parent `HEAD` freezes the group's base. On reuse with no explicit
base, accept the existing clean detached `HEAD`. With an explicit base, require
the resolved OID to equal that `HEAD`. Never switch, reset, or advance a reused
parent implicitly.

Create or resolve the exact top-level Herdr workspace in the named session and
verify that its canonical checkout is `repo/` and its non-empty `repo_key`
differs from the source repository's key. Apply the shared server-boundary gate:
the exact audited Herdr 0.8.0/protocol 19 path may perform only its pinned local
read-only discovery before registration. Another build needs equivalent proof
or exact server-startup attestation; a clean Herdr client environment is not
proof.

## Recognize and resume existing state

No marker file or runtime database establishes ownership. Recognize Git clone
compatibility independently from Herdr reconciliation. Reuse of the clone
requires every Git condition below:

- the exact deterministic or explicitly approved canonical group path;
- a valid independent non-bare primary repository at `repo/`;
- selected remote name and opaque URL bytes equal the requested remote when
  compared internally;
- no alternates or hardlink dependency and successful connectivity checks;
- a clean detached parent at the compatible frozen base;
- every child contained below the exact `worktrees/` directory.

Refuse and preserve an incomplete clone, unknown directory, alternates-backed
experiment, wrong parent, unexpected child, dirty parent, or mismatched remote.
After the Git clone passes those checks, reconcile the exact named session. A
missing Herdr parent or missing child is resumable state: include only the
missing compatible open/create action in the new preview and confirmation. An
existing Herdr parent or child with conflicting path or repository provenance
is a refusal. Never treat absence as conflicting provenance, and never close or
replace an existing workspace.

## Create children

Detached review mode is the default. Create the child at the frozen remote OID
with `git worktree add --detach`, then open its canonical path beneath the
verified isolated parent. Require its Git `HEAD` to equal the frozen OID, its
Herdr `repo_key` to equal the isolated parent and differ from the source, and
Herdr to report both linked and detached provenance.

`--editable` is separate authority. Create a new local tracking branch from the
selected remote branch without force or reset. Refuse an incompatible existing
local branch. Warn and obtain exact confirmation when the same remote branch is
known to be checked out elsewhere, because independent clones can edit and push
concurrently.

For a batch, the first child is the canary. Verify its OID, mode, containment,
linked provenance, isolated `repo_key`, and parent before starting child two.
Repeat the full gate for each later child. Stop at the first failure; preserve
and report completed, partial, and pending children without rollback.

## Independence and final proof

Before reporting success, prove:

- parent and children share one isolated common directory and non-empty
  `repo_key`, while the source has a different key;
- parent and detached children remain at their exact frozen commit OIDs;
- `objects/info/alternates` is absent;
- full object connectivity and exact commit lookup succeed independently;
- repeated open returns a verified no-op with stable workspace identity; and
- all checkout, Git metadata, Herdr config, socket, and log paths remain inside
  the intended roots.

On any failure, reconcile what exists and report a stable redacted state. Never
delete the clone or worktree, remove a branch, close a workspace/session, kill a
process, or perform destructive rollback.
