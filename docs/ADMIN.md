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

Press `Enter` to open the preset's managed prompt/resource source browser.
Use `j`/`k` to move between resources and `Page Up`/`Page Down` to scroll
source text. The browser shows the canonical repository source and its rendered
provider targets; it does not read content back from provider session storage.

It must not present a preset pipeline as a live run, mark phases active, choose
the next agent, or dispatch work. If legacy task-state fixtures are present,
they may be shown only as schema/debug information, not as runtime control.

### Legacy Fixtures

Shows experimental local state files only when they exist. This view is for
schema/debug inspection while configuration authoring matures. It is not a task
monitor, approval queue, provider status view, session browser, or run observer.

### Providers / Existing Provider Config

Shows existing Codex, Claude, Antigravity, Hermes, Kaji, and future-harness
configuration roots and existing manifest target paths. This is separate from the
preset library. The view shows executable health, version, exact declared
configuration roots, render capabilities, inspection issues, and the current
unchanged/update/conflict classification of existing manifest targets. It does
not recursively enumerate or copy provider runtime caches, sessions, transcripts,
histories, credentials, or secrets. Refresh runs the same bounded, read-only
inspection used by `agentctl provider doctor` and configuration planning.

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

An unmanaged conflict means a file already exists at a declared target but
agentctl has no install-state ownership record for it. A changed managed conflict
means agentctl previously installed the file and its current checksum no longer
matches that record. Both are preserved until the user explicitly resolves the
difference.

## Keys

| Key | Action |
|---|---|
| `1` through `5` | Open a view |
| `Tab`, `Shift+Tab` | Switch views |
| Left/Right or `h`/`l` | Switch views |
| Up/Down or `j`/`k` | Move selection |
| `g`, `G` | First or last item |
| `Enter` | Inspect a preset's prompt/resource source |
| `Page Up`, `Page Down` | Scroll prompt/resource source |
| `r` | Refresh all read-only state |
| `?`, `Esc` | Open or close help |
| `q`, `Ctrl+C` | Quit |

The minimum supported terminal size is 48 columns by 12 rows. Use
`--no-alt-screen` when terminal scrollback is preferable to a full-screen
session.

## Current Boundary

The TUI is a configuration studio. Runtime automation is deliberately outside
its scope. The preset library, DAG browser, source preview, provider inspection,
and conflict explanation are implemented.

Pipeline and step editing should be delivered in a separate change because it
introduces writes. That editor should:

1. Edit repository preset data, never provider files directly.
2. Work on an in-memory draft until the user asks to save.
3. Validate preset schema, phase references, edges, entry phases, loop markers,
   manifest resource references, and provider targets before save.
4. Show the exact JSON/file diff and resulting provider plan before confirmation.
5. Write atomically and retain a recoverable previous version.
6. Keep applying provider configuration as a separate explicit action.

Other future work should add:

- create, copy, edit, and delete actions in the TUI;
- structured workflow/pipeline DAG editing inside presets;
- MCP/config bundle editing inside presets;
- provider-native render previews;
- structured settings merge and ownership;
- safe explicit apply flows.

Do not add hidden run control, live observation, approval queues, dispatch, or
agent-session management to terminal key handlers.
