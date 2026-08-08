# Admin Terminal Interface

## Purpose

`agentctl admin` is a terminal configuration surface for the local
agentctl installation. It is primarily a read-only browser for configuration
and plans, with one guarded write path: applying the selected preset after an
explicit conflict decision and final confirmation. It is not a surface for
running or observing agent work. It brings together:

- configuration repository health;
- provider installation and configuration health;
- the preset library and each preset's workflow/pipeline DAGs;
- external environment requirements referenced by presets;
- MCP references, commands, prompts, skills, hooks, settings, and provider targets inside presets;
- existing Claude, Codex, Hermes, Antigravity, Kaji, and future-harness configuration;
- managed configuration drift and conflicts;
- provider mappings for generated commands, prompts, skills, MCP config, and settings.

The interface does not dispatch an agent, watch live runs, infer next actions,
commit, push, or open a pull request. Applying configuration remains an explicit
preview, conflict-decision, and confirmation operation.

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

Shows the reusable preset library. A preset can be a provider configuration
bundle or a provider-neutral environment bundle. Configuration presets contain
workflow DAGs/pipelines plus MCP references, commands, prompts, skills, hooks,
settings, and provider targets. Environment-only presets reference machine
tooling without provider targets. The current view loads
`config/presets/*.json`, computes safe plans, and answers:

```text
Which presets exist?
Which workflow DAGs/pipelines are inside this preset?
Which MCPs, commands, prompts, skills, hooks, and settings belong to it?
Which provider targets are declared?
Which external commands are satisfied, missing, blocked, or unsupported?
What will change if this preset is applied?
```

Environment detection is path-only and read-only. Opening or refreshing the
Presets view never runs a tool, package manager, plugin host, or remote
installer. The view displays typed installation commands or official manual
links before any install is confirmed. The same pack can also be installed by
the guarded CLI command:

```bash
agentctl environment install --yes <pack>
```

Environment presets are searchable and grouped under `environments`. They show
the guarded install command. Press `i` (or `a`) to review the exact environment
plan, then `y` to run it. This environment-specific flow does not ask for a
provider or project scope because its target is the local machine. Satisfied
requirements are skipped, command output is bounded and filtered, and status is
refreshed after completion or partial failure.

Press `Enter` to open the preset's managed prompt/resource source browser.
Use `j`/`k` to move between resources and `Page Up`/`Page Down` to scroll
source text. The browser shows the canonical repository source and its rendered
provider targets; it does not read content back from provider session storage.

For configuration presets, press `a` to open the preset apply panel. If the
plan has conflicts, choose:

- `k` keeps customized files, remembers the exact decision, and applies all
  remaining preset changes;
- `x` backs up conflicting files and replaces them with preset versions;
- `Esc` cancels without writing.

After choosing, press `y` for final confirmation. Abort remains the default;
agentctl never chooses keep or replace automatically.

It must not present a preset pipeline as a live run, mark phases active, choose
the next agent, or dispatch work.

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

Summarizes managed resources as unchanged, create, update, kept-existing, or
conflict actions. When conflicts belong to a preset, press `a` from Config to
open that preset's guarded apply panel. Resolve one preset at a time; after an
apply, the refreshed plan shows whether another preset still needs a decision.
Presets can also be applied through:

```bash
agentctl plan
agentctl apply --yes
agentctl apply --conflicts keep --yes
agentctl apply --conflicts replace --yes
```

For a single preset, use:

```bash
agentctl preset plan standard-work
agentctl preset apply --yes standard-work
agentctl preset apply --conflicts keep --yes codex-compatibility
```

An unmanaged conflict means a file already exists at a declared target but
agentctl has no install-state ownership record for it. A changed managed conflict
means agentctl previously installed the file and its current checksum no longer
matches that record. Both are preserved until the user explicitly resolves the
difference.

A keep-existing decision records both source and target checksums. It appears as
`IGNORED`/kept rather than a conflict on later plans. If the repository source or
customized target changes, agentctl marks that decision stale and asks again.
Replace decisions create timestamped backups under
`~/.config/agentctl/backups/`.

The `codex-resource-lab` preset provides one visible example of every Codex
resource category:

- an MCP configuration fragment;
- a deprecated custom prompt kept for compatibility testing;
- a preferred Codex skill;
- an inactive hook example;
- an opt-in named settings profile.

The prompt, skill, and named profile use their provider-native paths. MCP and
hook examples are installed under `.codex/agentctl/fragments/` for review
because structured merge into active `config.toml` and `hooks.json` files is not
implemented yet. Applying the preset therefore cannot silently activate a
server or hook.

## Keys

| Key | Action |
|---|---|
| `1` through `4` | Open a view |
| `Tab`, `Shift+Tab` | Switch views |
| Left/Right or `h`/`l` | Switch views |
| Up/Down or `j`/`k` | Move selection |
| `g`, `G` | First or last item |
| `Enter` | Inspect a preset's prompt/resource source, or accept an installer choice |
| `i` (`a` remains an alias) | Install the selected configuration or environment preset |
| `/` | Search presets by ID, name, description, target, or resource |
| `f` | Cycle preset resource filters and groups |
| `u`, `p` | Choose user-global or project-folder scope in the installer |
| `k` | Keep existing files in the apply decision panel |
| `x` | Replace conflicts from the preset in the apply decision panel |
| `y` | Confirm the reviewed configuration apply or environment install |
| `Page Up`, `Page Down` | Scroll prompt/resource source |
| `r` | Refresh all read-only state |
| `?`, `Esc` | Open or close help |
| `q`, `Ctrl+C` | Quit |

The minimum supported terminal size is 48 columns by 12 rows. Use
`--no-alt-screen` when terminal scrollback is preferable to a full-screen
session.

## Current Boundary

The TUI is a configuration studio. Runtime automation is deliberately outside
its scope. The preset library, resource search/filter/grouping, DAG browser,
source preview, provider inspection, conflict explanation, and guarded scoped
preset install are implemented. Installation always chooses one provider and
either user-global or an explicit existing project folder before building the
plan. Conflict decisions affect only that scoped plan.

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

Do not add hidden run control, live observation, approval queues, dispatch, or
agent-session management to terminal key handlers.
