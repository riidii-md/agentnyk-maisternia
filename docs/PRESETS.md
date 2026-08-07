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
- `harness-profile`: read-only configuration, capability, and usage profiling;
- `session-audit`: evidence-backed correctness, trajectory, process/safety, and
  cost review plus delegated bottleneck analysis for one completed run;
- `harness-improvement`: post-run profiling, audit, repeated-pattern proposals,
  held-out replay, and human-approved installation;
- `codex-compatibility`: permanent Codex-prefixed compatibility aliases;
- `codex-resource-lab`: a safe Codex-only example with one MCP reference,
  prompt, skill, hook, and settings resource;
- `approval-standard`: the provider-neutral least-privilege allow, ask, and
  deny policy with human-only grants;
- `hook-safety`, `hook-continuity`, `hook-quality`, `hook-delegation`,
  `hook-maintenance`, and `hook-observability`: focused hook definitions;
- `hook-standard`: the recommended safety, continuity, and repository-quality
  bundle, including the approval policy;
- `hook-complete`: all six hook packs plus the approval policy for explicit
  inspection and selection.

Inspect them:

```bash
agentctl preset list
agentctl preset show idea-shaping
agentctl preset show scored-experiment
agentctl preset show harness-improvement
agentctl preset show codex-resource-lab
agentctl preset show approval-standard
agentctl preset show hook-standard
agentctl preset validate all
```

`codex-resource-lab` makes all six content counters testable in the TUI. Its
prompt and skill install to provider-native locations. Its MCP and hook files
are review fragments, not active configuration, because agentctl does not yet
merge structured TOML or JSON. Its settings resource is an opt-in named Codex
profile and does not replace the user's main `config.toml`.

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

# Install a preset into one repository instead of the provider user home.
agentctl preset apply \
  --scope project \
  --project /path/to/repository \
  --target codex \
  --yes \
  hook-quality

# Preserve customized target files and apply everything else.
agentctl preset apply \
  --target codex \
  --conflicts keep \
  --yes \
  codex-compatibility

# Back up customized files and replace them with preset versions.
agentctl preset apply \
  --target codex \
  --conflicts replace \
  --yes \
  codex-compatibility
```

These commands select only the manifest resources declared by the preset. They
delegate to the same configurator used by full-manifest operations, preserving
target allowlists, symlink checks, managed install state, drift detection,
conflict detection, backups, atomic writes, and explicit confirmation.

`--scope user` is the default and resolves targets under `--home`. Its managed
state and backups live under `~/.config/agentctl`. `--scope project` resolves
targets under `--project`; its local managed state and backups live under
`<project>/.agentctl`. State from one scope is never used to claim ownership in
the other scope.

Conflict handling defaults to `abort`. `keep` records the exact source and
target checksums behind the decision, so the target is reported as `IGNORED`
instead of blocking later applies. If either side changes, the decision becomes
stale and agentctl requires a new decision. `replace` backs up each target before
installing and managing the preset version.

`agentctl` stops after rendering or installing configuration. Claude Code,
Codex CLI, Hermes, and Antigravity own execution, sessions, approvals, history,
and runtime loops.

See [Provider-native experiment loops](PROVIDER-NATIVE-EXPERIMENTS.md) for the
scored experiment contract, provider mappings, safety model, and boundary with
Kaji.

See [Session retrospectives and harness improvement](RETROSPECTIVES.md) for the
profiling, audit, post-run artifact, replay, and human-approval contracts.

See [Hook packs and installation scopes](HOOKS.md) for hook policy,
provider-layer mappings, and the native activation boundary.

See [Approval policy](APPROVAL-POLICY.md) for the operation matrix and bounded
grant rules, and [Hook and approval roadmap](HOOK-APPROVAL-ROADMAP.md) for the
native enforcement plan.
