# Admin Terminal Interface

## Purpose

`agentctl admin` is a read-only terminal control surface for the local
agentctl installation. It brings together:

- configuration repository health;
- provider installation and configuration health;
- durable workflow tasks and approval state;
- pipeline phases, trigger branches, and verification loops;
- managed configuration drift and conflicts.

The interface does not approve work, apply configuration, dispatch an agent,
commit, push, open a pull request, or perform external writes.

## Configure The Repository

Save the configuration repository once:

```bash
agentctl config set-repository /path/to/agentctl
```

Inspect the saved value:

```bash
agentctl config show
```

After that, launch the interface from any directory:

```bash
agentctl admin
```

Repository resolution order is:

1. `--repo`;
2. `AGENTCTL_REPO`;
3. `~/.config/agentctl/settings.json`;
4. the current directory and its ancestors.

An explicit one-time override does not change saved settings:

```bash
agentctl admin --repo /path/to/another/agentctl
```

Clear the saved value:

```bash
agentctl config clear-repository
```

Settings are stored with mode `0600` under a mode `0700` agentctl directory.
The settings file contains only the repository path and schema version.

## Views

### Overview

Shows repository readiness, healthy provider count, active and waiting tasks,
configuration drift, and current attention items.

### Pipelines

Shows the selected task's known current phase and the configured policy
topology:

```text
DISCOVERY  BRIEF -> SCOUT -> ANALYZE -> RESEARCH -> DECIDE
READINESS  READY -> PLAN -> PROVE -> HANDOFF -> APPROVAL
EXECUTION  RUN -> VERIFY -> REVIEW -> PR
```

The view also shows trigger entry branches and the planned control loops:

```text
READY not ready -> RESEARCH
VERIFY failed   -> ANALYZE
REVIEW changes  -> RUN
```

For `shape` tasks, the same view switches to the idea-shaping topology and adds
source inbox and grill summaries:

```text
DISCOVERY  INTAKE -> RESEARCH -> GRILL -> BRAINSTORM
CONVERGE   CHALLENGE -> DECIDE -> PLAN -> FINAL -> HUMAN FINALIZE

SOURCE INBOX  4 total  1 unread  1 material
GRILL STATE   3 total  2 open    1 critical
```

Shape loops are guarded by recorded outcomes and a pass budget. Other pipeline
loops still describe policy topology and are not automatically executed. Only
the current phase is marked active because completed-phase history is not yet
projected into the TUI.

### Tasks

Lists durable task state from `~/.agent-workflow` and shows the selected task's
repository, phase, authority, approval status, next action, and update time.

### Providers

Shows Codex, Claude, Antigravity, and Hermes executable health, version,
runner support, configuration roots, capabilities, and inspection issues.
Refresh runs the same bounded, read-only version inspection used by
`agentctl provider doctor`.

### Config

Summarizes managed resources as unchanged, create, update, or conflict actions.
Conflicts can be inspected, but applying changes still requires a separate:

```bash
agentctl plan
agentctl apply --yes
```

## Keys

| Key | Action |
|---|---|
| `1` through `5` | Open a view |
| `Tab`, `Shift+Tab` | Switch views |
| Left/Right or `h`/`l` | Switch views |
| Up/Down or `j`/`k` | Move selection |
| `g`, `G` | First or last item |
| `r` | Refresh all read-only state |
| `?`, `Esc` | Open or close help |
| `q`, `Ctrl+C` | Quit |

The minimum supported terminal size is 48 columns by 12 rows. Use
`--no-alt-screen` when terminal scrollback is preferable to a full-screen
session.

## Current Boundary

The TUI is an observational admin surface. Controlled automation still
requires additional runtime work:

- concrete runner capability resolution;
- generalized durable outcomes beyond the shape transition history;
- explicit approval records;
- bounded read-only dispatch;
- write-phase leases, budgets, and cancellation;
- artifact registration with `mdmaid.show`.

Those features must extend the workflow state machine. They must not be hidden
inside terminal key handlers.
