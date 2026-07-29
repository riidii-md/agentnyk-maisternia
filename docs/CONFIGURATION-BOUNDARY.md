# Configuration Boundary

`agentctl` is a preset and pipeline configurator, not a workflow runtime.

## Owns

- A local library of reusable presets.
- Provider-neutral workflow/pipeline DAG definitions inside those presets.
- Phase prompt/command templates, MCP references, hooks, skills, and settings bundles.
- Provider adapter metadata.
- Render previews and staging trees.
- Safe apply, backups, conflict checks, and drift checks.
- Configuration TUI views for presets, workflow DAGs, providers, existing provider files, generated files, and settings.

## Does Not Own

- Running Claude, Codex, Hermes, Antigravity, Kaji, or other harnesses.
- Live task observation.
- Runtime phase transitions.
- Agent-session history.
- Approval queues.
- Commit, push, PR, or release actions.

If the TUI observes live work, it becomes a controller. Keep it focused on
authoring and applying configuration.

## Presets And Pipelines

A preset is the user-facing library item. It is a reusable configuration bundle
that can be created, copied, edited, previewed, and applied later.

A pipeline is one part of a preset: the workflow DAG or phase graph. The same
preset may also include MCP server/tool configuration, command aliases, prompts,
skills, hooks, provider settings, and target mappings.

A pipeline in `agentctl` is therefore not a running job. It describes what
provider-native files should be generated so an external harness can run the
workflow in its own environment.

```text
agentctl preset
  -> workflow DAG + MCP/config bundles
  -> provider render plan
  -> staged files
  -> explicit apply
  -> harness-owned execution
```

## Harness Relationship

Claude Code, Codex CLI, Hermes, Antigravity, Kaji, and future tools are
harnesses. `agentctl` may render configuration for them, but it should not
duplicate their runtime loops.

## Existing Provider Configuration

`agentctl` should also have an inspection surface for existing Claude, Codex,
Hermes, Antigravity, Kaji, and future-harness configuration. This is separate
from the preset library:

- **Preset library:** what `agentctl` can create, edit, preview, and apply.
- **Existing provider config:** what is already installed in each harness home.

The inspection surface should classify files as managed, unmanaged, conflicting,
or unknown. It should help the user decide what to import or overwrite, but it
should not treat provider runtime caches, sessions, transcripts, or histories as
agentctl-owned configuration.
