# Preset Library

## Model

A preset is a reusable configuration or environment bundle. A pipeline is a
declarative workflow DAG inside a configuration bundle, not a running job.

Version 1 preset files live under:

```text
config/presets/<preset-id>.json
```

Each file declares:

- an ID, display name, and description;
- zero or more workflow DAGs;
- managed resource references grouped as MCPs, commands, prompts, skills,
  hooks, and settings;
- optional environment-pack references for tools the workflow needs outside
  provider configuration directories;
- canonical provider targets.

An environment-only preset has no provider targets or managed manifest
resources. It references an environment pack and uses the guarded environment
installer when the preset is applied.

Every content value is a resource ID from `config/manifest.json`. The manifest
remains the source of truth for source files and provider-native target paths.
This keeps presets compositional without duplicating file ownership rules.

Environment-pack references resolve against `config/environments`. They model
machine or shared tooling such as Zellij, Tatami, Herdr, and documented host
plugins; they are not provider target paths and are not copied into every
harness home.

## Included Presets

The repository starts with:

- `standard-work`: the provider-neutral delivery workflow;
- `idea-shaping`: source intake, research, grill, brainstorm, challenge,
  decision, and planning;
- `scored-experiment`: a provider-native baseline, focused change, scoring,
  evidence, and bounded continuation loop;
- `parallel-work`: dependency-aware parallel planning and bounded execution
  waves with isolated writes, integration barriers, and sequential fallback;
- `terminal-orchestration`: provider-neutral machine setup for Zellij, Tatami,
  Herdr, Mdmaid, and three pinned Herdr plugins;
- `multi-lens-review`: plan and implementation review with independent lenses,
  per-finding refutation, applied fixes, and optional provider delegation;
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
maisternia preset list
maisternia preset show idea-shaping
maisternia preset show scored-experiment
maisternia preset show parallel-work
maisternia preset show terminal-orchestration
maisternia preset show multi-lens-review
maisternia preset show harness-improvement
maisternia preset show codex-resource-lab
maisternia preset show approval-standard
maisternia preset show hook-standard
maisternia preset validate all
```

`codex-resource-lab` makes all six content counters testable in the TUI. Its
prompt and skill install to provider-native locations. Its MCP and hook files
are review fragments, not active configuration, because maisternia does not yet
merge structured TOML or JSON. Its settings resource is an opt-in named Codex
profile and does not replace the user's main `config.toml`.

Validation rejects unknown fields, unsupported schema versions, invalid IDs,
duplicate resources, unknown manifest resources, unknown providers, dangling
edges, duplicate graph elements, and cycles that are not marked as explicit
loop edges.

## Authoring

Create an empty preset:

```bash
maisternia preset create \
  --name "Team Workflow" \
  --description "Shared delivery configuration" \
  team-work
```

Copy an existing preset when a working bundle is a better starting point:

```bash
maisternia preset copy \
  --name "Team Standard Work" \
  standard-work team-standard-work
```

Metadata can be changed without rewriting the file manually:

```bash
maisternia preset edit \
  --description "Delivery workflow for the platform team" \
  team-standard-work
```

Structured content and DAG editing is not yet exposed as a command. Edit the
JSON file, then run:

```bash
maisternia preset validate team-standard-work
```

Deletion is intentionally explicit:

```bash
maisternia preset delete --yes team-work
```

Deleting the catalog definition does not choose an installation scope and does
not mutate provider homes. Remove an installed preset before or after deleting
its definition with the ID retained in install state:

```bash
maisternia preset uninstall \
  --scope user \
  --target codex \
  --yes \
  team-work
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
They do not make `maisternia` a runtime and do not cause it to execute phases.

## Plan, Render, Apply

Use preset-scoped operations when only one bundle should be considered:

```bash
maisternia preset plan --scope user --target hermes idea-shaping

maisternia preset render \
  --target codex \
  --output ./build/rendered-standard-work \
  standard-work

maisternia preset apply --scope user --target codex --yes standard-work

# Remove all configuration targets owned only by this preset. This also works
# when the preset definition is no longer present in the catalog.
maisternia preset uninstall --scope user --target codex --yes standard-work

# Install a preset into one repository instead of the provider user home.
maisternia preset apply \
  --scope project \
  --project /path/to/repository \
  --target codex \
  --yes \
  hook-quality

# Preserve customized target files and apply everything else.
maisternia preset apply \
  --scope user \
  --target codex \
  --conflicts keep \
  --yes \
  codex-compatibility

# Back up customized files and replace them with preset versions.
maisternia preset apply \
  --scope user \
  --target codex \
  --conflicts replace \
  --yes \
  codex-compatibility
```

These commands select only the manifest resources declared by the preset. They
delegate to the same configurator used by full-manifest operations, preserving
target allowlists, symlink checks, managed install state, drift detection,
conflict detection, backups, atomic writes, and explicit confirmation.

Preset apply also reconciles previously recorded preset ownership. Any managed
target removed from the preset becomes `REMOVE` when that preset was its last
owner, or `RELEASE` when another installed preset still declares the same
target. Removal backs up the file first. Locally modified files become
conflicts; `--conflicts keep` leaves the file and relinquishes ownership, while
`--conflicts replace` backs it up and removes it. This lifecycle is shared by
all six managed content categories: MCP references, commands, prompts, skills,
hooks, and settings.

When a preset references environment packs, `preset plan` and the admin Presets
view also show a read-only environment plan. Detection checks whether each
declared command exists on `PATH`; it does not run tools or installers. Missing
requirements do not cause configuration-preset apply to install packages
implicitly. Applying an environment-only preset skips provider/scope selection,
shows the exact environment plan, and requires confirmation. The direct
`maisternia environment install --yes <pack>` command remains available. See
[Environment requirements](ENVIRONMENT-REQUIREMENTS.md).

Environment packs are host requirements rather than provider-file resources.
Their current installer remains presence-based and does not claim package
manager ownership, upgrade an already-present tool, or uninstall it when a
preset changes. Host-tool removal must therefore remain an explicit operation
through its package manager or plugin host until environment install state is
implemented.

`--scope user` is the default and resolves targets under `--home`. Its managed
state and backups live under `~/.config/maisternia`. `--scope project` resolves
targets under `--project`; its local managed state and backups live under
`<project>/.maisternia`. State from one scope is never used to claim ownership in
the other scope.

Install state schema version 3 records preset-to-target ownership. Older state
is read safely, but its targets are not retroactively attributed to a preset.
Apply a preset once after upgrading to establish ownership for future removal.

Conflict handling defaults to `abort`. `keep` records the exact source and
target checksums behind the decision, so the target is reported as `IGNORED`
instead of blocking later applies. If either side changes, the decision becomes
stale and maisternia requires a new decision. `replace` backs up each target before
installing and managing the preset version.

`maisternia` stops after rendering or installing configuration. Claude Code,
Codex CLI, Hermes, and Antigravity own execution, sessions, approvals, history,
and runtime loops.

See [Provider-native experiment loops](PROVIDER-NATIVE-EXPERIMENTS.md) for the
scored experiment contract, provider mappings, safety model, and boundary with
Kaji.

See [Parallel work and the speed loop](PARALLEL-WORK.md) for dependency-aware
planning, write isolation, execution waves, integration barriers, provider
fallback, and speed/cost reporting.

See [Session retrospectives and harness improvement](RETROSPECTIVES.md) for the
profiling, audit, post-run artifact, replay, and human-approval contracts.

See [Multi-lens review workflow](REVIEW-WORKFLOW.md) for plan and implementation
gates, evidence rules, verifier/refutation passes, applied fixes, domain lenses,
and controlled cross-provider delegation.

See [Hook packs and installation scopes](HOOKS.md) for hook policy,
provider-layer mappings, and the native activation boundary.

See [Approval policy](APPROVAL-POLICY.md) for the operation matrix and bounded
grant rules, and [Hook and approval roadmap](HOOK-APPROVAL-ROADMAP.md) for the
native enforcement plan.
