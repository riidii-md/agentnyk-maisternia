# Herdr Worktree Commands

The opt-in `herdr-worktrees` preset installs `/herdr-add` and
`/herdr-regroup` for Codex, Claude Code, Antigravity, and Hermes. Maisternia
installs the command and skill definitions; the active harness executes Git and
Herdr operations after the command is invoked.

The preset is not part of `standard-work`, `terminal-orchestration`, or the
`software-engineer` collection. It neither installs Herdr nor changes a live
session during preset application.

## Install

Preview the exact provider files before applying:

```bash
maisternia preset show herdr-worktrees
maisternia preset plan --scope user --target codex herdr-worktrees
maisternia preset apply --scope user --target codex --yes herdr-worktrees
```

Replace `codex` with `claude`, `antigravity`, or `hermes` as needed. Restart a
harness when it requires a restart to discover newly installed commands or
skills.

| Provider | Invocation |
|---|---|
| Codex | `$herdr-add`, `$herdr-regroup`, or the compatibility prompts |
| Claude Code | `/herdr-add`, `/herdr-regroup` |
| Antigravity | provider-native prompt mappings |
| Hermes | `herdr-add` and `herdr-regroup` skills |

The resolved runtime must already provide Git and Herdr. Both commands require
a Herdr-managed caller and always target the exact supplied named session. They
never start, restart, or attach an interactive client to a stopped session.
Every non-no-op Git or Herdr mutation set receives a redacted preview and needs
fresh exact human confirmation; invoking the command is not confirmation.

## Native grouping

Use native mode to create or open a worktree beneath the primary checkout for
the same Git common directory:

```text
/herdr-add agentic SA-3500-report SA-3500-report
/herdr-add agentic SA-3500-report SA-3500-report --parent-label "Audits"
```

The label changes presentation only. Repository identity comes from Git and
Herdr provenance. A linked worktree cannot become a parent for sibling
worktrees.

Use regroup to repair Herdr provenance for an existing linked worktree:

```text
/herdr-regroup agentic --current
/herdr-regroup agentic --path /absolute/path/to/checkout
/herdr-regroup agentic --workspace w9
/herdr-regroup agentic --all-repo --parent-label "Audits"
```

Regroup never creates a Git checkout or moves a workspace across repositories
or sessions. A plain workspace is adopted only when the installed Herdr behavior
proves that its workspace, tabs, panes, and processes remain intact; otherwise
the command refuses the operation.

## Separate named groups

Herdr groups worktrees by Git common directory. To create another top-level
Review or Features group for the same upstream project, request an isolated
group explicitly:

```text
/herdr-add agentic SA-3500-report SA-3500-report \
  --isolated-group "October Review"
```

This mode creates or reuses an independent non-bare clone and opens that clone
as the top-level Herdr parent. Its children are linked worktrees of that clone,
not of the source repository. The default child is detached at the exact fetched
remote commit, which is appropriate for review.

Without an explicit base, the isolated parent uses the selected remote's
fetched symbolic HEAD. If that HEAD is absent or ambiguous, the command requires
an explicit base and makes no change. A compatible independent clone can resume
with a missing Herdr parent or child; conflicting existing provenance is
refused.

Use `--editable` only when the isolated child must have a local tracking branch:

```text
/herdr-add agentic SA-3500-report SA-3500-report \
  --isolated-group "October Review" --editable
```

Independent clones can edit and push the same remote branch concurrently. The
command warns and asks before a known concurrent checkout.

The default group root is under:

```text
${XDG_DATA_HOME:-$HOME/.local/share}/maisternia/herdr-groups/v1/
```

Its deterministic path uses separate full Git object hashes for source common
directory, exact session, group name, and worktree name. An explicit
`--group-root` names the exact container that will hold `repo/` and
`worktrees/`. Symlinks, unsafe permissions, path overlap, unexpected occupants,
and unrecognized partial clones are refused.

An isolated group copies committed remote state only. It does not move or copy
the source workspace, panes, processes, index, uncommitted files, stashes, hooks,
or arbitrary Git configuration. `/herdr-regroup` cannot target it.

## Safety and recovery

- Remote URLs stay opaque and are excluded from previews and reports. Embedded
  user information, tokens, query data, fragments, controls, and option-like
  leading-hyphen values are refused. The validated credential-free URL is
  stored only in the private isolated clone's selected remote for later reuse.
- Remote discovery and drift checks use non-mutating `ls-remote` queries. They
  do not fetch into the source repository before confirmation.
- Git runs with interactive askpass and terminal prompting disabled, a verified
  controlled system/global configuration, empty template and hooks directories,
  disabled filesystem monitors, and a restricted transport policy. Checkouts
  that would execute an external Git filter or driver are refused without
  separate exact authority.
- A clean Herdr client environment is not treated as proof about an existing
  server. Git creation runs directly under the controlled profile. Exact Herdr
  0.8.0/protocol 19 registration requires the official platform release digest
  for both the resolved client and the live process that owns the exact session
  socket, and uses only the pinned source-audited local read-only discovery
  path; an unknown or custom build requires equivalent proof or server startup
  attestation.
- Clone creation must produce an independent object store. Shared clones,
  retained alternates, and hardlink dependencies are forbidden.
- The isolated parent is clean and detached at a frozen base commit. Detached
  children are created at frozen branch commits.
- A batch verifies its first child as a canary before creating the second.
- Ref movement, identity drift, wrong provenance, dirty/ahead/local-only
  requested work, or an unknown partial destination stops mutation.
- Recovery is non-destructive. The command reports completed, partial, pending,
  refused, and unchanged state without deleting or rolling back repositories,
  worktrees, branches, workspaces, panes, or processes.

## Uninstall and lifecycle

Uninstall removes only the provider files managed by Maisternia:

```bash
maisternia preset uninstall --scope user --target codex --yes herdr-worktrees
```

It does not remove runtime clones, worktrees, branches, Herdr workspaces, or
sessions. Removing those durable artifacts is a separate task requiring exact
inventory and explicit authority.

Static all-provider rendering proves installation parity. It does not prove
that every provider harness executed the runtime workflow. Runtime conformance
must identify the actual harness, resolved Herdr executable, and disposable
session that were exercised.
