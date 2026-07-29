# agentctl

[![CI](https://github.com/kagi-labs/agentctl/actions/workflows/ci.yml/badge.svg)](https://github.com/kagi-labs/agentctl/actions/workflows/ci.yml)

`agentctl` is a provider-neutral preset and pipeline configurator for
command-line coding agent harnesses.

It provides one version-controlled source of truth for:

- shared work phases;
- Codex, Claude, Antigravity, and Hermes provider adapters;
- reusable preset-library entries;
- workflow/pipeline DAGs inside presets;
- MCP references and neutral `/work-*` commands;
- permanent provider aliases such as `/codex-plan`;
- personal skills and policies;
- model roles, provider capability metadata, and existing provider-config inspection;
- safe configuration rendering and installation.

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
- backups before managed updates;
- drift detection using install checksums;
- atomic file writes;
- a strict, versioned preset library under `config/presets`;
- reusable standard-work, idea-shaping, and Codex-compatibility presets;
- preset DAG validation with explicit loop edges and cycle rejection;
- preset create, copy, metadata edit, delete, list, show, and validation commands;
- preset-scoped plan, staging render, and guarded apply;
- a complete initial work-phase catalog;
- complete Claude-to-Codex command adapters with explicit authority boundaries;
- direct Codex aliases for the canonical phases;
- repository tests that prevent command inventory and adapter behavior from
  silently shrinking;
- strict normalized event validation as untrusted input fixtures;
- provider-specific `/work-shape`, `/work-source`, `/work-grill`, and
  `/work-brainstorm` command templates;
- a read-only configuration TUI backed by real preset-library entries,
  workflow DAGs, provider health, per-preset plans, and managed files;
- cross-platform CI snapshot builds and tag-based releases;
- persistent configuration-repository discovery for system installations.

Structured TOML and JSON settings merging, structured preset-content and DAG
editing, TUI write actions, broader provider-native rendering, existing
provider-file classification, and configuration import are planned next.
Runtime dispatch is intentionally out of scope; existing harnesses run the
rendered commands.

## Installation

Source installation is available now. Homebrew, `go install ...@latest`, and
release archives become available after the first tagged release and the
one-time private tap setup described in
[Release process](docs/RELEASING.md).

### Homebrew

`agentctl` is currently distributed from private GitHub repositories. Authenticate
Git and provide a GitHub token to Homebrew:

```bash
gh auth login
gh auth setup-git
brew tap kagi-labs/tap
HOMEBREW_GITHUB_API_TOKEN="$(gh auth token)" \
  brew install --cask kagi-labs/tap/agentctl
```

### Go

```bash
gh auth setup-git
GOPRIVATE=github.com/kagi-labs/* \
  go install github.com/kagi-labs/agentctl/cmd/agentctl@latest
```

### Build From Source

```bash
git clone git@github.com:kagi-labs/agentctl.git
cd agentctl
make install
```

Verify any installation:

```bash
agentctl --version
```

See [Installation](docs/INSTALLATION.md) for upgrades, release downloads, and
uninstallation.

## Admin

Point the system installation at this configuration repository once:

```bash
agentctl config set-repository /path/to/agentctl
```

Then open the read-only admin interface from any directory:

```bash
agentctl admin
```

Use `1` through `5` to open Overview, Presets, State Fixtures, Providers, and
Config. Press `?` for all keys. The current TUI browses preset-library entries,
their workflow DAGs and contents, provider health, preset-scoped plans, drift,
and conflicts. Preset writes and apply remain explicit CLI operations. The TUI
must not run, observe, approve, dispatch, commit, or push.

See [Admin terminal interface](docs/ADMIN.md) for repository resolution,
controls, and configuration boundaries.

## Pipeline Configuration

`agentctl` owns the preset and pipeline configuration layer:

- reusable preset-library entries;
- provider-neutral workflow/pipeline DAGs and phase definitions inside presets;
- MCP references, provider adapters, and capability metadata;
- command, prompt, skill, hook, MCP, and settings rendering;
- existing provider-configuration inspection;
- safe preview, conflict detection, and guarded installation;
- configuration inspection through the CLI and TUI.

It does **not** own runtime execution, task observation, agent dispatch, phase
control, or harness approvals. After `agentctl` renders and installs a pipeline,
run it inside the harness you choose:

```text
Claude Code: /work-shape
Codex CLI:   /work-shape
Hermes:      work-shape skill
Antigravity: provider-native prompt/command mapping
```

Those harnesses own their own sessions, histories, approvals, and execution
loops. `agentctl` may validate and render configuration for them, but it must not
become a controller or observer of live runs.

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
go run ./cmd/agentctl doctor
```

Inspect what would happen without writing:

```bash
go run ./cmd/agentctl plan --target codex
go run ./cmd/agentctl plan --target claude
go run ./cmd/agentctl plan --target antigravity
```

`agy` remains accepted as a permanent compatibility alias for `antigravity`.

Inspect the installed provider CLIs without executing an agent:

```bash
go run ./cmd/agentctl provider list
go run ./cmd/agentctl provider inspect agy
go run ./cmd/agentctl provider doctor all
go run ./cmd/agentctl provider capabilities hermes
```

Provider doctor never invokes a provider's native doctor command.

Inspect and validate the preset library:

```bash
go run ./cmd/agentctl preset list
go run ./cmd/agentctl preset show idea-shaping
go run ./cmd/agentctl preset validate all
```

Plan or stage only the files selected by one preset:

```bash
go run ./cmd/agentctl preset plan --target hermes idea-shaping
go run ./cmd/agentctl preset render \
  --target codex \
  --output ./build/standard-work \
  standard-work
```

Preset apply uses the same conflict, drift, backup, and managed-state checks as
the full manifest apply, and still requires explicit confirmation:

```bash
go run ./cmd/agentctl preset apply --target codex --yes standard-work
```

Render a staging tree:

```bash
go run ./cmd/agentctl render \
  --target all \
  --output ./build/rendered
```

`apply` refuses unmanaged conflicts and requires explicit confirmation:

```bash
go run ./cmd/agentctl apply --target codex --yes
```

Do not run `apply` against a real home directory until the displayed plan has
been reviewed.

## Experimental State Fixtures

The repository still contains earlier state-machine commands for event
validation, manual ingestion, task inspection, and idea-shaping transitions:

```bash
go run ./cmd/agentctl event validate ./examples/events/issue-opened.json
go run ./cmd/agentctl event ingest ./examples/events/issue-opened.json
go run ./cmd/agentctl task list
go run ./cmd/agentctl work next <task-id>
agentctl pipeline start shape --title "Improve agent workflow"
```

Treat these as experimental schema fixtures, not the product runtime. They are
useful for validating untrusted event envelopes and exploring state shapes, but
new product work should focus on preset-library authoring, workflow DAG editing,
MCP/config bundle editing, existing provider-configuration inspection,
provider-native rendering, configuration previews, drift/conflict handling, and
explicit apply.

`agentctl` should not expand these commands into live observation, phase
control, approval queues, or agent dispatch. Existing harnesses own execution.

## Documentation

- [Improved workflow](docs/WORKFLOW.md)
- [Preset library](docs/PRESETS.md)
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
/work-review
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
