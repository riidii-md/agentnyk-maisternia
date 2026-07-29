# Preset Library

## Model

A preset is a reusable configuration bundle. A pipeline is a declarative
workflow DAG inside that bundle, not a running job.

Version 1 preset files live under:

```text
config/presets/<preset-id>.json
```

Each file declares:

- an ID, display name, and description;
- zero or more workflow DAGs;
- managed resource references grouped as MCPs, commands, prompts, skills,
  hooks, and settings;
- canonical provider targets.

Every content value is a resource ID from `config/manifest.json`. The manifest
remains the source of truth for source files and provider-native target paths.
This keeps presets compositional without duplicating file ownership rules.

## Included Presets

The repository starts with:

- `standard-work`: the provider-neutral delivery workflow;
- `idea-shaping`: source intake, research, grill, brainstorm, challenge,
  decision, and planning;
- `scored-experiment`: a provider-native baseline, focused change, scoring,
  evidence, and bounded continuation loop;
- `codex-compatibility`: permanent Codex-prefixed compatibility aliases.

Inspect them:

```bash
agentctl preset list
agentctl preset show idea-shaping
agentctl preset show scored-experiment
agentctl preset validate all
```

Validation rejects unknown fields, unsupported schema versions, invalid IDs,
duplicate resources, unknown manifest resources, unknown providers, dangling
edges, duplicate graph elements, and cycles that are not marked as explicit
loop edges.

## Authoring

Create an empty preset:

```bash
agentctl preset create \
  --name "Team Workflow" \
  --description "Shared delivery configuration" \
  team-work
```

Copy an existing preset when a working bundle is a better starting point:

```bash
agentctl preset copy \
  --name "Team Standard Work" \
  standard-work team-standard-work
```

Metadata can be changed without rewriting the file manually:

```bash
agentctl preset edit \
  --description "Delivery workflow for the platform team" \
  team-standard-work
```

Structured content and DAG editing is not yet exposed as a command. Edit the
JSON file, then run:

```bash
agentctl preset validate team-standard-work
```

Deletion is intentionally explicit:

```bash
agentctl preset delete --yes team-work
```

## DAG Rules

Each pipeline declares `entry_phases`, `phases`, and `edges`. A normal edge must
remain part of an acyclic graph. A back edge is allowed only when it is visibly
declared as a loop:

```json
{
  "from": "verify",
  "to": "analyze",
  "condition": "failed",
  "loop": true
}
```

Conditions and loops describe configuration that an external harness may use.
They do not make `agentctl` a runtime and do not cause it to execute phases.

## Plan, Render, Apply

Use preset-scoped operations when only one bundle should be considered:

```bash
agentctl preset plan --target hermes idea-shaping

agentctl preset render \
  --target codex \
  --output ./build/rendered-standard-work \
  standard-work

agentctl preset apply --target codex --yes standard-work
```

These commands select only the manifest resources declared by the preset. They
delegate to the same configurator used by full-manifest operations, preserving
target allowlists, symlink checks, managed install state, drift detection,
conflict detection, backups, atomic writes, and explicit confirmation.

`agentctl` stops after rendering or installing configuration. Claude Code,
Codex CLI, Hermes, and Antigravity own execution, sessions, approvals, history,
and runtime loops.

See [Provider-native experiment loops](PROVIDER-NATIVE-EXPERIMENTS.md) for the
scored experiment contract, provider mappings, safety model, and boundary with
Kaji.
