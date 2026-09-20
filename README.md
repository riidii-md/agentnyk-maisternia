# AgentnykMaisternia

[![CI](https://github.com/riidii-md/agentnyk-maisternia/actions/workflows/ci.yml/badge.svg)](https://github.com/riidii-md/agentnyk-maisternia/actions/workflows/ci.yml)

**One workshop for configuring coding-agent workflows across Codex, Claude
Code, Antigravity, and Hermes.**

AgentnykMaisternia turns a version-controlled catalog of workflows, skills,
hooks, policies, and presets into provider-native configuration. Its
`maisternia` CLI previews every change, detects conflicts and drift, and keeps
installation explicit.

[Quick start](#quick-start) · [How it works](#how-it-works) ·
[Project status](#project-status) · [Documentation](#documentation) ·
[Contributing](CONTRIBUTING.md)

## Why Maisternia?

Coding-agent tools solve similar problems but store their configuration in
different formats and locations. Maintaining the same workflow separately for
every provider is repetitive and easy to get wrong.

Maisternia provides:

- one declarative catalog for shared workflows and policies;
- provider-native output for four coding-agent harnesses;
- reusable presets for common engineering workflows;
- project-local or user-global installation;
- a readable plan before any configuration is changed;
- conflict, drift, backup, path, and symlink safeguards.

The name *Maisternia* transliterates the Ukrainian `майстерня`: a workshop where
things are made, assembled, and tuned.

## Project status

AgentnykMaisternia is **pre-release**. The configurator and embedded preset
catalog work today; packaged distribution is still being moved to the current
public repository.

| Area | Status | What this means |
|---|---|---|
| Source build | Available | Clone this repository and run `make install`. |
| Admin and CLI | Available | Browse, validate, plan, render, apply, and uninstall presets. |
| Provider targets | Available | Codex, Claude Code, Antigravity, and Hermes are supported. |
| Safe installation | Available | Apply is opt-in and guarded by conflict, drift, backup, and path checks. |
| Release archives | Pending | No tagged release has been published from this repository. |
| Homebrew | Pending | There is no public `riidii-md` tap yet. |
| `go install ...@latest` | Pending | The Go module path still needs to move from the former repository namespace. |
| Open-source license | Decision needed | The repository does not yet contain a `LICENSE` file. |

Older installation commands that reference `kagi-labs` are obsolete. Do not
use them for this repository.

## Quick start

### 1. Build from source

Requirements: Git and Go 1.25.8 or newer.

```bash
git clone https://github.com/riidii-md/agentnyk-maisternia.git
cd agentnyk-maisternia
make install
```

Make sure the Go binary directory is on `PATH`, then verify the installation:

```bash
maisternia --version
maisternia doctor
```

The binary contains its versioned configuration catalog. A source checkout is
not required after installation.

### 2. Explore the catalog

Open the terminal interface:

```bash
maisternia
```

Or stay in the CLI:

```bash
maisternia preset list
maisternia preset show standard-work
maisternia provider doctor all
```

These commands inspect configuration; they do not run an agent or apply a
preset.

### 3. Preview a project installation

Choose a repository and a provider, then inspect the plan:

```bash
maisternia preset plan \
  --scope project \
  --project /path/to/repository \
  --target codex \
  standard-work
```

Only after reviewing the plan, apply it explicitly:

```bash
maisternia preset apply \
  --scope project \
  --project /path/to/repository \
  --target codex \
  --yes \
  standard-work
```

`apply` stops on conflicts by default. Use `--conflicts keep` or
`--conflicts replace` only after reviewing the affected paths. Replacements are
backed up.

## How it works

```text
Versioned catalog
      │
      ▼
Select a preset ──► Preview the plan ──► Confirm apply ──► Use it in your harness
      │                    │                    │
      └─ workflows         ├─ conflicts        ├─ backups
         skills            ├─ drift            └─ managed ownership
         hooks             └─ exact paths
         policies
```

Maisternia renders shared definitions into each provider's native layout:

| Provider | Typical invocation after installation |
|---|---|
| Codex | `$work-shape` or `/prompts:work-shape` through the compatibility prompt |
| Claude Code | `/work-shape` |
| Hermes | `work-shape` skill |
| Antigravity | Provider-native prompt mapping |

Restart Codex after installing or updating workflow presets so newly installed
skills appear in suggestions.

### Start guided work

Install `standard-work`, then start a task with `$work-start <task>` in Codex
or `/work-start <task>` in Claude Code. The command gathers evidence and advances
through the applicable workflow phases in the current conversation. At a human
checkpoint it creates a reviewable document, presents the exact question or
decision, and waits for your response. Your reply resumes the workflow; you do
not need to invoke the next phase yourself. It pauses for material questions,
architectural direction and plan decisions, and implementation review.

### Route work to a harness or model

Canonical work commands accept an optional route before the task:

```text
/work-plan @codex -- plan the migration
/work-plan @claude @opus -- plan with Claude Opus
/work-plan @claude @opus @reasoning:high -- plan with high reasoning
/work-run @claude @sonnet -- implement the approved plan
/work-run-simplify @codex -- implement with the opt-in simplicity profile
/work-review @codex @claude -- review with both harnesses
/work-test-review @codex -- review test evidence for this change
```

A per-harness model selector and optional per-harness reasoning level follow
each harness. Reasoning accepts `low`, `medium`, or `high`; for example,
`@reasoning:high`. Explicit routes override
saved preferences, never widen authority, and never silently substitute another
model or reasoning level. The current harness remains the coordinator for routed work. See
[workflow routing](docs/WORKFLOW.md#route-canonical-commands-with-harness) for configuration
and invocation details.

## Included workflows

The catalog contains focused presets that can be installed independently:

- `standard-work` — review architectural direction when needed, visually review
  detailed plans, implement, verify, explain and
  approve the exact change, then prepare a PR;
- `idea-shaping` — turn an incomplete idea into an explicit decision and plan;
  its `work-question` utility finds one high-leverage question and one owned,
  time-bounded next move;
- `parallel-work` — create dependency-safe parallel plans and execution waves;
- `multi-lens-review` — review plans and implementations from independent lenses;
- `workflow-routing` — route work across supported harnesses and model roles;
- `adaptive-readability` — adapt technical material to its reader and purpose;
- `session-audit` and `harness-improvement` — review completed work and propose
  controlled improvements;
- `hook-standard`, `hook-complete`, and `approval-standard` — install reusable
  safety and policy definitions;
- `terminal-orchestration` — declare and verify the external terminal tools used
  by orchestration workflows.

Run `maisternia preset list` for the complete catalog. See the
[preset guide](docs/PRESETS.md) for contents, scopes, and limitations.

## Product decisions and boundaries

The project deliberately keeps configuration management separate from agent
runtime control.

| Decision | Consequence |
|---|---|
| Maisternia is a configurator, not an orchestrator | It installs workflows but does not run, dispatch, supervise, or observe agent sessions. |
| Apply remains opt-in | Planning and rendering are safe to explore; changes require explicit confirmation. |
| Managed files are tracked individually | Provider home directories are never synchronized as a whole because they contain mixed runtime and user state. |
| Conflicts and drift stop installation | Existing or locally changed files require a visible keep-or-replace decision. |
| Harnesses retain approval authority | Maisternia may install approval definitions, but live prompts remain owned by the selected harness. |
| External preset sources are immutable snapshots | Refreshing or removing a source is a separate, explicit operation. |

This boundary means Maisternia stores its catalog, installation ownership,
backups, and reviewed configuration. It does not store workflow task history,
agent transcripts, credentials, or runtime databases.

## Safety model

Provider configuration lives near valuable user state, so safe path handling is
a core feature rather than a convenience. Maisternia includes:

- provider-root and relative-path allowlists;
- traversal and destination-symlink rejection;
- unmanaged-file conflict detection;
- installed-checksum drift detection;
- source and target revalidation before apply;
- backups before managed updates and removals;
- atomic file writes;
- explicit `--yes` confirmation for apply and uninstall.

Read [Security](SECURITY.md) for the complete model. Please never include
credentials, real user configuration, transcripts, or runtime databases in an
issue or fixture.

## Common commands

```bash
# Validate the embedded catalog and manifest
maisternia doctor

# Inspect providers without running them
maisternia provider list
maisternia provider inspect codex
maisternia provider capabilities hermes

# Validate and render without touching provider configuration
maisternia preset validate all
maisternia preset render --target all --output ./build/rendered standard-work

# Remove a previously managed preset
maisternia preset uninstall --scope user --target codex --yes standard-work
```

Provider doctor is read-only and never invokes a provider's own doctor command.

## Development

Clone the repository, make the behavior observable in tests first, then run the
full verification suite:

```bash
make verify
go run ./cmd/maisternia doctor
go run ./cmd/maisternia render --target all --output ./build/rendered
```

See [Contributing](CONTRIBUTING.md) for the expected workflow. Pull requests
should explain behavior changes, tests, security implications, migration impact,
and remaining limitations.

## Documentation

### Get started

- [Installation and upgrades](docs/INSTALLATION.md)
- [Admin terminal interface](docs/ADMIN.md)
- [Preset library](docs/PRESETS.md)
- [Preset collections](docs/PRESET-COLLECTIONS.md)
- [Provider adapters](docs/PROVIDERS.md)

### Understand the design

- [Configuration boundary](docs/CONFIGURATION-BOUNDARY.md)
- [Configurator architecture](docs/CONFIGURATOR.md)
- [Runtime-boundary migration](docs/RUNTIME-BOUNDARY-MIGRATION.md)
- [Approval policy](docs/APPROVAL-POLICY.md)
- [Hook packs and installation scopes](docs/HOOKS.md)

### Explore workflows

- [Standard workflow](docs/WORKFLOW.md)
- [Architectural direction gate](docs/ARCHITECTURAL-DIRECTION.md)
- [Idea shaping](docs/IDEA-SHAPING-PIPELINE.md)
- [Parallel work](docs/PARALLEL-WORK.md)
- [Multi-lens review](docs/REVIEW-WORKFLOW.md)
- [Session retrospectives](docs/RETROSPECTIVES.md)
- [Change explanations](docs/CHANGE-EXPLANATIONS.md)
- [Mandatory human change review](docs/CHANGE-REVIEW-GATE.md)
- [Environment requirements](docs/ENVIRONMENT-REQUIREMENTS.md)

## License

A license has not been selected yet. Until a `LICENSE` file is added, the code
is publicly readable but is not licensed for reuse or redistribution. Selecting
an open-source license is a release-readiness decision.
