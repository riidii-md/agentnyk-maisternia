---
name: herdr-worktrees
description: Add Git worktrees to exact named Herdr sessions or repair native same-repository grouping. Use isolated named groups only when explicitly requested; do not use this skill for cleanup or cross-clone workspace moves.
---

# Herdr Worktrees

Operate Git checkouts and Herdr workspaces as two separate systems with separate
identity proof. Sidebar labels and position are presentation, never repository
or workspace identity.

## Select exactly one mode

- For native `/herdr-add`, read [references/native.md](references/native.md).
- For `/herdr-regroup`, read [references/native.md](references/native.md). The
  mode remains native even when the user calls the destination a group.
- For `/herdr-add --isolated-group <name>`, read
  [references/isolated.md](references/isolated.md).

Do not load both mode references unless the same user request explicitly
contains both independent operations. Regroup cannot target an isolated group.

## Common preconditions

Require a Herdr-managed caller with `HERDR_ENV=1`. Use
`herdr --session <exact-name>` for every live operation. A context-dependent
selector such as `--current` also requires the caller's canonical inherited
socket to match the named session's canonical socket. A session-name variable
or equal-looking workspace ID is not proof.

Probe the `herdr` executable actually resolved on `PATH`, then its version,
schema, relevant help, and structured live behavior. Do not infer capabilities
from package-manager metadata. Distinguish a missing or stopped session,
protocol mismatch, unsupported behavior, permission denial, and ambiguous
identity.

Never launch or attach an interactive Herdr client, and never start or restart
a stopped target session. If the exact session is absent or stopped, report the
user action required and stop before any Git or Herdr mutation.

Use argv-safe commands and parse structured JSON. Do not interpolate an
argument into shell syntax. Default every open/create action to `--no-focus`;
use focus only when the user explicitly asks.

Run every direct Git subprocess and every Git mutation with one fail-closed
profile. The exact audited read-only Herdr registration exception is defined
below; no server-side Git mutation is allowed. Start from an allowlisted environment and remove inherited
`GIT_CONFIG_*`, `GIT_ASKPASS`, `SSH_ASKPASS`, `GIT_SSH`, `GIT_SSH_COMMAND`, and
related prompt or command overrides. Set `GIT_TERMINAL_PROMPT=0`,
`SSH_ASKPASS_REQUIRE=never`, `GIT_CONFIG_NOSYSTEM=1`, and a
`GIT_CONFIG_GLOBAL` that names a verified private controlled file. Never fall
back to the user's system or global Git configuration. Set a verified empty
`GIT_TEMPLATE_DIR` and exact command-scope configuration. After scrubbing all
inherited `GIT_CONFIG_*`, use `GIT_CONFIG_COUNT` entries to bind
`core.hooksPath` to a verified empty private directory, set
`core.fsmonitor=false`, neutralize `core.askPass`, and set or neutralize
`core.sshCommand` for the selected transport. Inspect and reject repository
configuration for URL rewrites, credential helpers, askpass, SSH commands,
external hooks, fsmonitor, filters, diffs, or merge drivers that escape this
profile. Do not inherit repository hooks or templates.

For HTTP(S), scrub proxy and libcurl/TLS environment variables and reject every
repository-local `http.*`, `credential.*`, `url.*`, or selected-remote proxy
setting unless it is an exact separately authorized value. This includes
proxies, `sslVerify=false`, scoped TLS options, extra headers, cookie files,
client certificates, custom CA paths, redirect policy, and remote-specific
proxies. Set `http.sslVerify=true` and the standard initial-only redirect policy
in command-scope configuration. Never let repository configuration add a
header, credential source, proxy, or TLS exception.

Allow HTTP(S) only through one validated fixed platform credential-helper
executable copied into the controlled configuration; refuse shell-form helpers
and every other executable helper. Allow SSH only through a verified fixed
executable invoked with `-F /dev/null`, batch mode, strict host-key checking,
disabled password and keyboard-interactive authentication, `ProxyCommand=none`,
and local commands disabled. Set `GIT_ALLOW_PROTOCOL` and
`GIT_PROTOCOL_FROM_USER=0` to the exact selected transport and refuse all other
protocols. Before a checkout or worktree add, inspect effective local
configuration and checkout attributes. Refuse when any external filter,
clean/smudge/process driver, fsmonitor, hook, or external diff/merge command
could execute unless the user separately and explicitly authorizes that exact
program. Never replay raw stderr from a remote-bearing command.

Do not assume the environment of a `herdr` client reaches the already-running
Herdr server. The supported registration profile is exact Herdr `0.8.0`,
protocol `19`, using only `workspace create --cwd` and `worktree open
--workspace ... --path ...`; the audited v0.8.0 source at commit
`346411fa21afd297f5ed3b3fa56f9e3fbf7654b7` limits this path to workspace
creation plus local read-only Git metadata and `git worktree list --porcelain`.
Canonicalize the resolved regular executable, reject unsafe ownership or write
permissions, hash its exact bytes with SHA-256, and require the official v0.8.0
release digest for the current platform: Linux aarch64
`f647ac66468d9efbc642fe534fb284468f0aea60641606fc008dfc0d82a3ca87`,
Linux x86_64
`b872ea7e40fa2cb17e857ac9b62b1bf26db7b403c622f5d2f3f5b35f6e9acd28`,
macOS aarch64
`d53a9f93fccfdfcc55632927bf51002f5add0aa7990bcdf508ffbd84ac658178`,
or macOS x86_64
`77cb5afd6c8fcaaaf3bc28e474ec01c209331ad08094e20d7f8aa9b0bb78d649`.
Bind the exact canonical target-session socket to one live server process owned
by the current user. Capture its PID and start time, then resolve the loaded
executable instead of trusting the client path: use `/proc/<pid>/exe` on Linux,
or the server's `txt` vnode path plus device/inode evidence from `lsof` on
macOS. Require that loaded image to match the platform release digest above.
Recheck socket ownership, PID/start time, executable device/inode, and digest
immediately before and after each Herdr mutation. Refuse PID reuse, ambiguous
socket holders, a replaced or deleted executable, missing process evidence, or
any mismatch; server-reported version and protocol alone are insufficient.
Never use server-side worktree create or `worktree open --branch`. Verify the
exact version, protocol, help, pre/post Git topology, and structured result. For
an unknown client or server digest, source build, other platform, or other
operation, require equivalent source/binary provenance and capability proof or
an attested fail-closed server startup profile; otherwise refuse before
mutation.
Perform every Git mutation directly under the controlled profile.

## Authority and confirmation

Inspect before mutation. Resolve the exact repository, common directory,
session, parent workspace, child paths, branch/ref state, and conflicts. Build a
redacted preview containing the exact ordered mutations and their stable
identities. Exclude remote URLs, credential material, raw provider responses,
and unrelated session state.

Do not infer authority for deletion, cleanup, reset, force, prune, stash,
workspace/session closure, or process termination. Never copy uncommitted files
between checkouts. When an operation is already satisfied, prove and report a
no-op rather than recreating it.

Obtain exact human confirmation immediately before every non-no-op Git or Herdr
mutation set. Invoking the command is not confirmation. If the harness cannot
obtain confirmation, stop without mutation. A proven no-op needs no additional
confirmation. New children, changed refs, changed paths, or changed
destinations invalidate the preview. Reinspect and ask again rather than
silently widening approval.

## Mutation and verification

Revalidate mutable identities immediately before each mutation. After each Git
change, reread Git. After each Herdr change, reread the exact session. Success
requires the mode-specific Git and Herdr postconditions, not only a zero exit
status.

For a batch, mutate sequentially. Verify one child completely before starting
the next. Stop on drift or the first failure. Reconcile observable state and
report completed, partial, pending, refused, and unchanged items. Preserve
partial success; never invent destructive rollback.

## Reporting

Report the exact session, mode, parent identity, child checkout paths, branch or
detached OIDs, Herdr workspace IDs, and verified outcome. Redact remote URLs and
sensitive command output. State explicitly when isolated creation duplicates
committed repository state and therefore does not preserve an existing pane,
process, index, or uncommitted file.
