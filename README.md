# CLI Agent Configurator

`cli-agent-configurator` is a provider-neutral configuration manager and work
conductor for command-line coding agents.

It provides one version-controlled source of truth for:

- shared work phases;
- Codex, Claude, `agy`, and future Hermes adapters;
- neutral `/work-*` commands;
- permanent provider aliases such as `/codex-plan`;
- personal skills and policies;
- model roles and runner routing;
- safe configuration rendering and installation.

## Status

The repository contains the first safe configurator foundation:

- manifest validation;
- provider and path allowlists;
- traversal and symlink protection;
- read-only inventory and planning;
- staging-tree rendering;
- conflict detection;
- guarded apply with explicit `--yes`;
- backups before managed updates;
- drift detection using install checksums;
- atomic file writes;
- a complete initial work-phase catalog;
- Codex alias adapters for Claude.

Structured TOML and JSON settings merging, task-state persistence, dynamic
runner dispatch, and configuration import are planned next.

## Build

```bash
go build ./cmd/agent-config
```

## Test

```bash
go test ./...
```

## Safe First Run

Validate the repository:

```bash
go run ./cmd/agent-config doctor
```

Inspect what would happen without writing:

```bash
go run ./cmd/agent-config plan --target codex
go run ./cmd/agent-config plan --target claude
go run ./cmd/agent-config plan --target agy
```

Render a staging tree:

```bash
go run ./cmd/agent-config render \
  --target all \
  --output ./build/rendered
```

`apply` refuses unmanaged conflicts and requires explicit confirmation:

```bash
go run ./cmd/agent-config apply --target codex --yes
```

Do not run `apply` against a real home directory until the displayed plan has
been reviewed.

## Documentation

- [Improved workflow](docs/WORKFLOW.md)
- [Configurator architecture](docs/CONFIGURATOR.md)
- [Security](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## Command Model

Neutral commands describe the work:

```text
/work
/work-plan
/work-research
/work-run
/work-review
```

Provider aliases force a runner:

```text
/codex-plan
/codex-research
/codex-work-loop
/codex-review
```

The neutral workflow remains canonical. Aliases are provider-specific adapters,
not independent prompt copies.
