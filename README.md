# agentctl

`agentctl` is a provider-neutral workflow and configuration manager for
command-line coding agents.

It provides one version-controlled source of truth for:

- shared work phases;
- Codex, Claude, Antigravity, and Hermes provider adapters;
- neutral `/work-*` commands;
- permanent provider aliases such as `/codex-plan`;
- personal skills and policies;
- model roles and runner routing;
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
- a complete initial work-phase catalog;
- complete Claude-to-Codex command adapters with explicit authority boundaries;
- direct Codex aliases for the canonical phases;
- repository tests that prevent command inventory and adapter behavior from
  silently shrinking;
- strict normalized event validation;
- idempotent manual event ingestion;
- private durable task state and append-only history;
- read-only task and prepared-context inspection;
- a read-only admin TUI for providers, pipelines, tasks, and configuration;
- persistent configuration-repository discovery for system installations.

Structured TOML and JSON settings merging, runtime capability resolution,
controlled dispatch, and configuration import are planned next.

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

Use `1` through `5` to open Overview, Pipelines, Tasks, Providers, and Config.
Press `?` for all keys. Pipeline branches, approval gates, and verification
loops are visible, but the TUI cannot approve, apply, dispatch, commit, or push.

See [Admin terminal interface](docs/ADMIN.md) for repository resolution,
controls, and runtime boundaries.

## Test

```bash
go test ./...
```

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

## Event Workflow

Validate an event without creating state:

```bash
go run ./cmd/agentctl event validate ./examples/events/issue-opened.json
```

Prepare durable task state without executing an agent:

```bash
go run ./cmd/agentctl event ingest ./examples/events/issue-opened.json
go run ./cmd/agentctl task list
go run ./cmd/agentctl work next <task-id>
```

All initial triggers are read-only. Ingestion does not dispatch an agent or
grant implementation, commit, push, PR, or external-write authority.

## Shape Workflow

Start a durable idea-shaping task:

```bash
agentctl pipeline start shape \
  --title "Improve agent workflow" \
  --repository kagi-labs/agentctl
```

Add sources and record focused human questions:

```bash
agentctl source add <task-id> https://example.com/source
agentctl source list <task-id>
agentctl grill ask --why "This determines the viable options." \
  --critical <task-id> "What cannot change?"
agentctl grill next <task-id>
```

Move through guarded phases:

```bash
agentctl pipeline transition <task-id> research
agentctl pipeline transition <task-id> grill
agentctl pipeline transition <task-id> brainstorm
agentctl pipeline transition --outcome weak-options <task-id> brainstorm
agentctl pipeline transition --finalize <task-id> final
```

The shape pipeline keeps target-project access read-only. Generated workflow
artifacts may be written under private agentctl task state. Critical grill
questions, invalid edges, exhausted loop budgets, and implicit finalization are
rejected.

## Documentation

- [Improved workflow](docs/WORKFLOW.md)
- [Idea-shaping pipeline](docs/IDEA-SHAPING-PIPELINE.md)
- [Admin terminal interface](docs/ADMIN.md)
- [Event-driven workflow](docs/EVENT-WORKFLOW.md)
- [Configurator architecture](docs/CONFIGURATOR.md)
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
