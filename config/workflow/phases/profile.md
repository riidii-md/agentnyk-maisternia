---
name: work-profile
description: Profile the active agent harness configuration, context cost, capabilities, duplication, and operational risks without changing it.
version: 0.1.0
---

# /work-profile - Profile The Current Harness

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Inspect the active agent configuration and produce a reproducible baseline.
Do not edit configuration while profiling it.

Input:

`$ARGUMENTS`

The input may name a provider, repository, session export, profile policy, or
existing retrospective directory. If scope is missing, profile the provider
running this command and the current repository only. Do not crawl unrelated
home-directory configuration or provider history.

## Inventory

Identify active and shadowed configuration, including:

- project and user instructions;
- commands and prompts;
- skills and their activation descriptions;
- MCP servers and exposed tools;
- hooks and automatic actions;
- settings, profiles, plugins, and model routing;
- provider capabilities required by each configured resource.

For every item record its path, scope, enabled state, provenance, managed or
unmanaged status, size in bytes, and whether it is loaded always or on demand.
Never print secret values. Report only secret names and redacted locations.

## Measure

Use deterministic tools for file counts, byte counts, duplicate content, broken
references, and configuration syntax. Record exact provider usage only when it
comes from telemetry or an explicit session export. Otherwise mark token,
cache, reasoning, and monetary cost as `unknown`; do not estimate them as exact.

Flag:

- instructions that overlap or conflict;
- large always-loaded content;
- skills with vague or duplicate activation rules;
- unused or unavailable MCP servers and tools;
- hooks that add latency, recurse, or mutate state unexpectedly;
- provider capabilities required but not available;
- stale aliases, broken paths, and unmanaged drift;
- configuration whose benefit has no observable evidence.

## Output

Write a self-contained report to:

```text
.agent-runs/retrospectives/<run-id>/profile.md
```

Create or update `record.json` in the same directory using the installed
retrospective record schema. Store profile measurements in `metrics`; do not
invent values merely to satisfy the schema.

Include:

1. provider, model, repository, and configuration scope;
2. inventory by resource type;
3. measured context and usage costs, with unknowns explicit;
4. duplication, conflict, capability, and safety findings;
5. current strengths that should be preserved;
6. a baseline table suitable for comparison after a proposed change;
7. evidence paths and commands used.

End with observations only. Do not recommend or apply changes in this phase.
