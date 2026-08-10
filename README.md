# AgentnykMaisternia

[![CI](https://github.com/kagi-labs/agentnyk-maisternia/actions/workflows/ci.yml/badge.svg)](https://github.com/kagi-labs/agentnyk-maisternia/actions/workflows/ci.yml)

**AgentnykMaisternia** is an opinionated, provider-neutral workshop for
configuring command-line coding-agent harnesses. Its `maisternia` CLI provides
one version-controlled source of truth for:

- shared work phases;
- Codex, Claude, Antigravity, and Hermes provider adapters;
- reusable preset-library entries;
- immutable external preset sources from local folders or GitHub repositories;
- workflow/pipeline DAGs inside presets;
- MCP references and neutral `/work-*` commands;
- permanent provider aliases such as `/codex-plan`;
- personal skills and policies;
- reusable hook packs with explicit user or project installation scope;
- declarative environment requirements referenced by presets;
- a provider-neutral allow, ask, and deny approval policy;
- model roles, provider capability metadata, and existing provider-config inspection;
- safe configuration rendering and installation.

## Naming

**AgentnykMaisternia** is the product brand, **Maisternia** is its short name,
`maisternia` is the executable and local state namespace, and
`agentnyk-maisternia` is the repository slug. *Maisternia* transliterates the
Ukrainian `майстерня`: a workshop where things are made, assembled, and tuned.
That workshop metaphor covers the project's curation, composition,
configuration, rendering, and guarded installation responsibilities without
implying that it controls agent runtimes.

## Status

The repository contains the first safe configurator foundation:

- manifest validation;
- provider and path allowlists;
- canonical provider identities with compatibility aliases;
- checked-in provider capability and safety contracts;
- executable, version, and configuration-root inspection;
- read-only provider health reporting;
- traversal and symlink protection;
- read-only inventory and planning;
- staging-tree rendering;
- conflict detection;
- guarded apply with explicit `--yes`;
- backups before managed updates and removals;
- drift detection using install checksums;
- atomic file writes;
- a strict, versioned preset library under `config/presets`;
- reusable standard-work, idea-shaping, scored-experiment, parallel-work,
  multi-lens-review, harness-profile, session-audit, harness-improvement,
  terminal-orchestration, Codex-compatibility, and Codex resource-lab presets;
- preset DAG validation with explicit loop edges and cycle rejection;
- preset create, copy, metadata edit, delete, list, show, and validation commands;
- preset-scoped plan, staging render, guarded apply, ownership reconciliation,
  and uninstall;
- local-folder and GitHub preset-source registration, immutable snapshots,
  explicit refresh/removal, qualified IDs, and Admin source addition;
- strict environment-pack validation, read-only planning, and guarded typed installation;
- six validated hook packs and eight hook presets spanning safety, continuity,
  quality, delegation, maintenance, and redacted local observability;
- a strict human-only approval policy definition with bounded grants, deny
  precedence, CLI explanation, and a standalone installation preset;
- user-global and repository-local plan/apply with isolated state and backups;
- a complete initial work-phase catalog;
- complete Claude-to-Codex command adapters with explicit authority boundaries;
- direct Codex aliases for the canonical phases;
- repository tests that prevent command inventory and adapter behavior from
  silently shrinking;
- strict normalized event validation as untrusted input fixtures;
- provider-specific `/work-shape`, `/work-source`, `/work-grill`, and
  `/work-brainstorm` command templates;
- a configuration TUI backed by real preset-library entries, workflow DAGs,
  provider health, per-preset plans, managed files, and guarded preset apply;
- cross-platform CI snapshot builds and tag-based releases;
- a self-installing configuration catalog embedded in every binary;
- automatic current-Git-project suggestions for scoped preset installation.

Structured TOML, JSON, and YAML settings merging, native hook and approval
activation, structured preset-content and DAG editing, broader provider-native
rendering, existing provider-file classification, and configuration import are
planned next.
Runtime dispatch is intentionally out of scope; existing harnesses run the
rendered commands.

The `scored-experiment` preset establishes the provider-native experiment
workflow and capability contract. Native Stop/tool-guard hook rendering still
depends on the planned structured settings merge; see
[Provider-native experiment loops](docs/PROVIDER-NATIVE-EXPERIMENTS.md).

The `parallel-work` preset adds `/work-parallel-plan`, `/work-parallel-run`, and
`/work-speed-loop` for dependency-safe concurrent execution. See
[Parallel work and the speed loop](docs/PARALLEL-WORK.md).

The separate, provider-neutral `terminal-orchestration` preset installs or
verifies Zellij, Tatami, Herdr, Mdmaid, and two pinned Herdr plugins without
coupling machine setup to a workflow preset. See
[Environment requirements](docs/ENVIRONMENT-REQUIREMENTS.md).

The retrospective presets add read-only harness profiling, evidence-backed run
audits, and proposal-only improvement with held-out replay and human approval.
`/work-session-analysis` provides the direct end-of-session bottleneck review
for token cost, repetition, skills, user friction, setup, commands, and
delegated subagents.
See [Session retrospectives and harness improvement](docs/RETROSPECTIVES.md).

The `multi-lens-review` preset adds separate plan and implementation gates,
independent review lenses, per-finding refutation, coordinator-applied fixes,
and explicit cross-provider delegation. See
[Multi-lens review workflow](docs/REVIEW-WORKFLOW.md).

## Installation

Source installation is available now. Homebrew, `go install ...@latest`, and
release archives become available after the first tagged release and the
one-time private tap setup described in
[Release process](docs/RELEASING.md).

### Homebrew

AgentnykMaisternia is currently distributed from private GitHub repositories.
Authenticate Git and provide a GitHub token to Homebrew:

```bash
gh auth login
gh auth setup-git
brew tap kagi-labs/tap
HOMEBREW_GITHUB_API_TOKEN="$(gh auth token)" \
  brew install --cask kagi-labs/tap/maisternia
```

### Go

```bash
gh auth setup-git
GOPRIVATE=github.com/kagi-labs/* \
  go install github.com/kagi-labs/agentnyk-maisternia/cmd/maisternia@latest
```

### Build From Source

```bash
git clone git@github.com:kagi-labs/agentnyk-maisternia.git
cd agentnyk-maisternia
make install
```

Verify any installation:

```bash
maisternia --version
```

See [Installation](docs/INSTALLATION.md) for upgrades, release downloads, and
uninstallation.

## Admin

Open the Admin interface with no subcommand:

```bash
maisternia
```

`maisternia admin` is the explicit equivalent. On first use, the binary installs
its embedded, versioned catalog under `~/.config/maisternia/catalogs/`; no source
checkout or `config set-repository` step is required.

Use `1` through `4` to open Overview, Presets, Providers, and Config. Press `?`
for all keys. The current TUI browses preset-library entries, their workflow
DAGs and contents, provider health, drift, and conflicts. In Presets, use `/` to
search, `f` to filter/group by resource type, and `s` to add a validated local
folder or GitHub preset source. On a selected preset, press `i`
to install it, choose one provider, then choose user-global or a specific project
folder scope. When Maisternia is launched inside a Git repository, that project
is prefilled and recommended. Only that scoped plan is inspected; any conflicts
require an explicit keep-existing or replace-from-preset decision followed by
confirmation. Overview and Config can open the same scoped installer for a
conflicting preset. The TUI must not run, observe, dispatch, commit, or push.

See [Admin terminal interface](docs/ADMIN.md) for repository resolution,
controls, and configuration boundaries.

## Pipeline Configuration

`maisternia` owns the preset and pipeline configuration layer:

- reusable preset-library entries;
- provider-neutral workflow/pipeline DAGs and phase definitions inside presets;
- MCP references, provider adapters, and capability metadata;
- command, prompt, skill, hook, MCP, and settings rendering;
- existing provider-configuration inspection;
- safe preview, conflict detection, and guarded installation;
- configuration inspection through the CLI and TUI.

It does **not** own runtime execution, task observation, agent dispatch, phase
control, or harness approvals. After `maisternia` renders and installs a pipeline,
run it inside the harness you choose:

```text
Claude Code: /work-shape
Codex CLI:   /work-shape
Hermes:      work-shape skill
Antigravity: provider-native prompt/command mapping
```

Those harnesses own their own sessions, histories, live approval prompts, and
execution loops. `maisternia` may define approval policy and render configuration
for them, but it must not become a controller or observer of live runs.

The current `event`, `task`, `work next`, and `pipeline start/transition`
commands are retained as experimental state-model fixtures while the
configuration model settles. They should not be presented as the product
runtime, and new TUI work should focus on structured preset and workflow DAG
editing, MCP/config bundle editing, existing provider-file classification,
render previews, provider mappings, drift, and conflicts.

## Test

```bash
make verify
```

## CI/CD

The `CI` workflow runs on pull requests, pushes to `main`, and manual dispatch:

1. module verification, formatting, vet, race tests, coverage, and a local
   build;
2. GoReleaser configuration validation;
3. release-equivalent snapshot archives for macOS, Linux, and Windows on
   `amd64` and `arm64`;
4. upload of archives, checksums, and coverage for 14 days.

Pushing a `v*` tag reruns `make verify`, publishes a GitHub release through
GoReleaser, and updates the Homebrew tap when its token is configured.

## Safe First Run

Validate the repository:

```bash
go run ./cmd/maisternia doctor
```

Inspect what would happen without writing:

```bash
go run ./cmd/maisternia plan --target codex
go run ./cmd/maisternia plan --target claude
go run ./cmd/maisternia plan --target antigravity
```

`agy` remains accepted as a permanent compatibility alias for `antigravity`.

Inspect the installed provider CLIs without executing an agent:

```bash
go run ./cmd/maisternia provider list
go run ./cmd/maisternia provider inspect agy
go run ./cmd/maisternia provider doctor all
go run ./cmd/maisternia provider capabilities hermes
```

Provider doctor never invokes a provider's native doctor command.

Inspect and validate the preset library:

```bash
go run ./cmd/maisternia preset list
go run ./cmd/maisternia preset show idea-shaping
go run ./cmd/maisternia preset show scored-experiment
go run ./cmd/maisternia preset show parallel-work
go run ./cmd/maisternia preset show terminal-orchestration
go run ./cmd/maisternia preset show harness-improvement
go run ./cmd/maisternia preset show codex-resource-lab
go run ./cmd/maisternia preset validate all
```

Inspect the external tools referenced by presets without running installers:

```bash
go run ./cmd/maisternia environment list
go run ./cmd/maisternia environment show terminal-orchestration
go run ./cmd/maisternia environment plan terminal-orchestration
go run ./cmd/maisternia preset plan terminal-orchestration
```

Environment detection only checks command presence on `PATH`; it does not run
the tools or any suggested installer.

After reviewing that plan, install missing requirements explicitly:

```bash
go run ./cmd/maisternia preset apply --yes terminal-orchestration
# Equivalent direct pack command:
go run ./cmd/maisternia environment install --yes terminal-orchestration
```

Environment install uses typed commands only, verifies each requirement, and
runs only after explicit confirmation in the CLI or Admin install review.

Inspect hook packs and preview a user-global or repository-local installation:

```bash
go run ./cmd/maisternia hook list
go run ./cmd/maisternia hook show safety
go run ./cmd/maisternia hook validate all
go run ./cmd/maisternia hook plan --scope user --target codex hook-standard
go run ./cmd/maisternia hook plan \
  --scope project \
  --project /path/to/repository \
  --target claude \
  hook-quality
```

Hook apply uses the normal explicit confirmation and conflict controls. The
current implementation installs managed, provider-neutral definitions. It does
not yet modify provider settings to activate native hooks; that requires the
planned structured settings merger.

Inspect and install the standard approval definition:

```bash
go run ./cmd/maisternia approval list
go run ./cmd/maisternia approval explain git.push
go run ./cmd/maisternia approval validate
go run ./cmd/maisternia approval plan --scope user --target codex
```

The current implementation installs a managed policy input; it does not yet
activate native enforcement. See [Approval policy](docs/APPROVAL-POLICY.md) and
the [hook and approval roadmap](docs/HOOK-APPROVAL-ROADMAP.md).

Plan or stage only the files selected by one preset:

```bash
go run ./cmd/maisternia preset plan --scope user --target hermes idea-shaping
go run ./cmd/maisternia preset render \
  --target codex \
  --output ./build/standard-work \
  standard-work
```

Preset apply uses the same conflict, drift, backup, and managed-state checks as
the full manifest apply, records per-preset target ownership, and still requires
explicit confirmation. Applying a changed preset removes targets it previously
owned only after drift and shared-ownership checks:

```bash
go run ./cmd/maisternia preset apply --scope user --target codex --yes standard-work
go run ./cmd/maisternia preset apply \
  --scope user \
  --conflicts keep \
  --yes \
  codex-compatibility

go run ./cmd/maisternia preset uninstall \
  --scope user \
  --target codex \
  --yes \
  standard-work
```

Uninstall also works by remembered preset ID after its catalog definition has
been deleted. It covers all managed preset resource categories. Environment
packs remain presence-based host requirements and are not automatically removed
through package managers or plugin hosts.

Render a staging tree:

```bash
go run ./cmd/maisternia render \
  --target all \
  --output ./build/rendered
```

`apply` aborts on conflicts by default and requires explicit confirmation.
Choose `keep` to preserve customized files and remember that decision, or
`replace` to back them up and install the repository version:

```bash
go run ./cmd/maisternia apply --target codex --yes
go run ./cmd/maisternia apply --target codex --conflicts keep --yes
go run ./cmd/maisternia apply --target codex --conflicts replace --yes
```

Do not run `apply` against a real home directory until the displayed plan has
been reviewed.

## Experimental State Fixtures

The repository still contains earlier state-machine commands for event
validation, manual ingestion, task inspection, and idea-shaping transitions:

```bash
go run ./cmd/maisternia event validate ./examples/events/issue-opened.json
go run ./cmd/maisternia event ingest ./examples/events/issue-opened.json
go run ./cmd/maisternia task list
go run ./cmd/maisternia work next <task-id>
maisternia pipeline start shape --title "Improve agent workflow"
```

Treat these as experimental schema fixtures, not the product runtime. They are
useful for validating untrusted event envelopes and exploring state shapes, but
new product work should focus on preset-library authoring, workflow DAG editing,
MCP/config bundle editing, existing provider-configuration inspection,
provider-native rendering, configuration previews, drift/conflict handling, and
explicit apply.

`maisternia` should not expand these commands into live observation, phase
control, approval queues, or agent dispatch. Existing harnesses own execution.

## Documentation

- [Improved workflow](docs/WORKFLOW.md)
- [Preset library](docs/PRESETS.md)
- [Environment requirements](docs/ENVIRONMENT-REQUIREMENTS.md)
- [Parallel work and the speed loop](docs/PARALLEL-WORK.md)
- [Multi-lens review workflow](docs/REVIEW-WORKFLOW.md)
- [Hook packs and installation scopes](docs/HOOKS.md)
- [Session retrospectives and harness improvement](docs/RETROSPECTIVES.md)
- [Idea-shaping pipeline](docs/IDEA-SHAPING-PIPELINE.md)
- [Admin terminal interface](docs/ADMIN.md)
- [Event-driven workflow](docs/EVENT-WORKFLOW.md)
- [Configurator architecture](docs/CONFIGURATOR.md)
- [Configuration boundary](docs/CONFIGURATION-BOUNDARY.md)
- [Provider adapters](docs/PROVIDERS.md)
- [Installation](docs/INSTALLATION.md)
- [Release process](docs/RELEASING.md)
- [Mdmaid human-in-the-loop integration](docs/MDMAID-HUMAN-IN-THE-LOOP.md)
- [Mdmaid project boundaries and naming](docs/MDMAID-PROJECT-BOUNDARIES.md)
- [Security](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## Command Model

Neutral commands describe the work:

```text
/work
/work-shape
/work-source
/work-grill
/work-brainstorm
/work-plan
/work-research
/work-run
/work-plan-review
/work-review
/work-delegated-review
```

Provider aliases force a runner:

```text
/codex-scout
/codex-shape
/codex-analyze
/codex-plan
/codex-research
/codex-decision
/codex-ready
/codex-work-loop
/codex-review
/codex-pr-check
/codex-showcase
/codex-cleanup
```

The neutral workflow remains canonical. Aliases are provider-specific adapters,
not independent workflow definitions. Claude adapters are intentionally longer
because they contain the executable Codex handoff, model/profile selection,
sandbox authority, temporary-output handling, and result synthesis.
