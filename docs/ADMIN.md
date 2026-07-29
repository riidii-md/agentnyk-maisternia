# Admin Terminal Interface

## Purpose

`agentctl admin` is a terminal configuration surface for the local
agentctl installation. Its current implementation is a read-only browser for
configuration and plans, not a writer and not a surface for running or observing
agent work. It brings together:

- configuration repository health;
- provider installation and configuration health;
- the preset library and each preset's workflow/pipeline DAGs;
- MCP references, commands, prompts, skills, hooks, settings, and provider targets inside presets;
- existing Claude, Codex, Hermes, Antigravity, Kaji, and future-harness configuration;
- managed configuration drift and conflicts;
- provider mappings for generated commands, prompts, skills, MCP config, and settings.

The interface does not approve work, dispatch an agent, watch live runs, infer
next actions, commit, push, or open a pull request. Applying configuration must
remain an explicit preview-and-confirm operation.

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

Shows repository readiness, healthy provider count, available presets, existing
provider health, configuration drift, and current render/apply attention items.

### Presets

Shows the reusable preset library. A preset is a configuration bundle: workflow
DAGs/pipelines plus MCP references, commands, prompts, skills, hooks, settings,
and provider targets. The current view loads `config/presets/*.json`, computes a
safe plan for each selected preset, and answers:

```text
Which presets exist?
Which workflow DAGs/pipelines are inside this preset?
Which MCPs, commands, prompts, skills, hooks, and settings belong to it?
Which provider targets are declared?
What will change if this preset is applied?
```

It must not present a preset pipeline as a live run, mark phases active, choose
the next agent, or dispatch work. If legacy task-state fixtures are present,
they may be shown only as schema/debug information, not as runtime control.

### State Fixtures

Shows experimental local state files only when they exist. This view is for
schema/debug inspection while configuration authoring matures. It is not a task
monitor, approval queue, or run observer.

### Providers / Existing Provider Config

Shows existing Codex, Claude, Antigravity, Hermes, Kaji, and future-harness
configuration roots and files. This is separate from the preset library. The view
should classify files as managed, unmanaged, conflicting, unknown, or ignored so
the user can decide what to import, preserve, or overwrite. It must not inspect
or copy provider runtime caches, sessions, transcripts, histories, credentials,
or secrets.

Shows Codex, Claude, Antigravity, Hermes, Kaji, and future-harness executable
health, version, configuration roots, render capabilities, and inspection issues.
Refresh runs the same bounded, read-only version inspection used by
`agentctl provider doctor`.

### Config

Summarizes managed resources as unchanged, create, update, or conflict actions.
Conflicts can be inspected, but applying changes still requires a separate:

```bash
agentctl plan
agentctl apply --yes
```

For a single preset, use:

```bash
agentctl preset plan standard-work
agentctl preset apply --yes standard-work
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

The TUI is a configuration studio. Runtime automation is deliberately outside
its scope. The preset library and DAG browser are implemented. Future work
should add:

- create, copy, edit, and delete actions in the TUI;
- structured workflow/pipeline DAG editing inside presets;
- MCP/config bundle editing inside presets;
- existing provider-configuration inspection;
- provider-native render previews;
- structured settings merge and ownership;
- drift and conflict explanation;
- safe explicit apply flows.

Do not add hidden run control, live observation, approval queues, dispatch, or
agent-session management to terminal key handlers.
