# Provider Adapters

## Purpose

`agentctl` needs one provider-neutral workflow while still respecting the
different configuration layouts, execution modes, output protocols, and safety
behavior of each CLI agent.

Provider adapters are checked-in capability contracts. They answer:

- what the provider is called;
- which aliases resolve to it;
- where its configuration lives;
- which resource types it understands;
- how to detect its executable and version;
- whether headless execution metadata exists for downstream harnesses;
- which authority modes may be requested;
- which output formats can be parsed.

The first registry covers:

| Canonical ID | Executable | Alias |
|---|---|---|
| `claude` | `claude` | none |
| `codex` | `codex` | none |
| `antigravity` | `agy` | `agy` |
| `hermes` | `hermes` | none |

`antigravity` is the stable identity used in manifests, state, routing, and
future events. `agy` remains a permanent user-facing compatibility alias and
the executable name.

## Commands

List providers and inspect local health:

```bash
agentctl provider list
agentctl provider inspect antigravity
agentctl provider inspect --json agy
```

Run agentctl's own provider checks:

```bash
agentctl provider doctor all
agentctl provider doctor codex
```

Inspect the declared contract without requiring the executable:

```bash
agentctl provider capabilities claude
agentctl provider capabilities --json hermes
```

The commands are read-only. They do not dispatch a model, edit provider
configuration, or run native doctor commands.

## Contract Layout

Each provider has:

```text
config/providers/<provider>/adapter.json
```

The format is described by:

```text
config/schema/provider-adapter.schema.json
```

The Go validator additionally enforces canonical identities, registered aliases,
normalized relative roots, sorted unique capability lists, compatible runner and
parser formats, and safety-capability consistency.

An adapter has four operational parts:

| Part | Responsibility |
|---|---|
| Renderer | Configuration roots and resource kinds |
| Inspector | Executable discovery, version parsing, native doctor metadata |
| Runner metadata | Headless safety metadata, authority boundaries, output formats |
| Parser | Accepted result formats and structured-output support |

Capabilities are facts, not preferences. Workflow routing may eventually
require capabilities, but it must not invent them or broaden authority because a
preferred provider is unavailable.

## Current Safety Baseline

| Provider | Headless | Safe for automated headless use | Initial authority |
|---|---:|---:|---|
| Claude Code | yes | yes | read-only, artifact write, workspace write |
| Codex CLI | yes | yes | read-only, artifact write, workspace write |
| Antigravity | yes | yes | read-only only |
| Hermes | yes | no | none |

Antigravity starts read-only because its bounded write behavior has not yet been
proven by the adapter test suite.

Hermes one-shot mode bypasses dangerous-operation approvals. Therefore
`agentctl` declares `headless=true` but `safe_headless=false`, exposes no
automated authority mode, and does not advertise it as an eligible dispatcher.
Safe mode alone is not evidence that unattended one-shot execution preserves
approval boundaries.

These are conservative declarations. Adding a capability requires an executable
adapter, a bounded authority mapping, parser behavior, and tests.

## Configuration Roots

Provider inspection checks paths and file types only. It does not read settings,
credentials, sessions, or caches.

| Provider | Inspected roots |
|---|---|
| Claude Code | `~/.claude/` |
| Codex CLI | `~/.codex/` |
| Antigravity | `~/.gemini/antigravity-cli/`, `~/.gemini/config/` |
| Hermes | `~/.hermes/` |

Roots are marked with ownership:

- `managed`: agentctl may own the rendered content;
- `mixed`: stable user configuration and provider runtime content coexist;
- `runtime`: the provider owns the directory and agentctl only inspects it.

Missing optional roots produce a degraded health result. Missing executables,
required roots, non-directory roots, or symlink traversal produce errors.

## Antigravity Compatibility

The existing workflow manifest still renders phase prompts under:

```text
~/.config/agy/prompts/
```

This is retained to avoid silently relocating existing managed files. It is not
listed as an Antigravity-owned configuration root, and `agentctl` does not claim
that the current Antigravity CLI consumes it.

The next renderer increment should:

1. define the exact mapping for skills, plugins, settings, instructions, hooks,
   and MCP configuration under `~/.gemini/config/`;
2. render that mapping into staging;
3. validate it against Antigravity behavior;
4. show a migration plan from the legacy prompt tree;
5. require explicit confirmation before moving or removing files.

## Native Doctors

The adapter records native doctor metadata separately from agentctl inspection.

Current policy:

| Provider | Native doctor | Automatic execution |
|---|---|---|
| Claude Code | available | never |
| Codex CLI | available | never |
| Antigravity | not declared | never |
| Hermes | available | never |

Claude and Hermes native doctors may initialize integrations or perform broader
dependency checks. Codex's JSON doctor is marked safe in the contract, but still
requires an explicit future command before agentctl may invoke it.

## Next Increment

Provider configuration and provider execution remain separate layers.

The next useful implementation step is renderer mapping:

1. define canonical resource objects;
2. add provider-specific path and format renderers;
3. support structured settings merge with key ownership;
4. render all output to staging before apply;
5. add provider-specific validation fixtures;
6. keep runtime dispatch out of `agentctl`; provider execution belongs to the
   external harness after configuration is rendered and applied.
