# Configuration Boundary

`agentctl` is a pipeline configurator, not a workflow runtime.

## Owns

- Provider-neutral pipeline definitions.
- Phase prompt/command templates.
- Provider adapter metadata.
- Render previews and staging trees.
- Safe apply, backups, conflict checks, and drift checks.
- Configuration TUI views for pipelines, providers, generated files, and settings.

## Does Not Own

- Running Claude, Codex, Hermes, Antigravity, Kaji, or other harnesses.
- Live task observation.
- Runtime phase transitions.
- Agent-session history.
- Approval queues.
- Commit, push, PR, or release actions.

If the TUI observes live work, it becomes a controller. Keep it focused on
authoring and applying configuration.

## Pipeline Meaning

A pipeline in `agentctl` is a reusable configuration artifact. It describes what
provider-native files should be generated so an external harness can run the
workflow in its own environment.

```text
agentctl pipeline definition
  -> provider render plan
  -> staged files
  -> explicit apply
  -> harness-owned execution
```

## Harness Relationship

Claude Code, Codex CLI, Hermes, Antigravity, Kaji, and future tools are
harnesses. `agentctl` may render configuration for them, but it should not
duplicate their runtime loops.
