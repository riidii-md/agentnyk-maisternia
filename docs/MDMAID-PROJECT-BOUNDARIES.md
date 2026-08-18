# Mdmaid Project Boundaries and Naming

## Status

Accepted and implemented direction. The workflow product is branded
**AgentnykMaisternia**, its repository is `agentnyk-maisternia`, its CLI is
`maisternia`, and the presentation service is the separate `mdmaid.desk`
repository with the `mdmaid-desk` CLI.

## Decision Direction

The persistent presentation service should be a separate project.

`mdmaid.nvim` demonstrates useful multi-file preview behavior, but it should
not be a dependency of the CLI agent workflow.

The new service should work:

- without Neovim;
- without an active agent session;
- across multiple repositories;
- across multiple CLI agent providers;
- as a persistent local document workspace.

Later, `mdmaid.nvim` may use this same service instead of managing its own
server lifecycle.

## Proposed Project Split

```mermaid
flowchart TD
    A[Agent harness or collaboration runtime] -->|documents and artifact events| S[mdmaid.desk]
    N[mdmaid.nvim] -->|register and open| S
    X[Scripts and other tools] -->|register documents| S
    S -->|render Markdown and Mermaid| M[mdmaid]
    S --> C[(Document catalog)]
    S --> B[Browser workspace]
    B --> H[Human]
    H -->|explicit decision| W[Harness or collaboration approval state]
```

### `mdmaid`

Responsibility:

- Markdown rendering;
- Mermaid rendering;
- reusable rendering APIs;
- low-level preview primitives;
- common styles and assets.

It should not own:

- agent workflow state;
- task approvals;
- provider routing;
- persistent agent artifact policy.

### `mdmaid.desk`

Responsibility:

- persistent local daemon;
- workspace registration;
- document catalog;
- directory watching;
- stable document URLs;
- browser document library;
- document search and grouping;
- CLI and local API;
- secure filesystem boundaries;
- attention indicators.

It should remain useful outside agent workflows. Agents are important document
producers, but the registration API should be generic.

### `mdmaid.nvim`

Responsibility:

- Neovim commands and keymaps;
- registering Markdown buffers;
- opening the matching browser document;
- editor-specific file selection;
- health checks and editor notifications.

It should eventually support two modes:

```text
daemon mode:
  use mdmaid.desk

standalone compatibility mode:
  start the existing per-Neovim preview server
```

In daemon mode:

```text
BufEnter *.md
  -> mdmaid-desk register <file> --workspace <id>
  -> open the returned or existing desk URL when policy requires
```

The Neovim process no longer owns the persistent server.

### Agent Harness or Collaboration Runtime

Responsibility:

- live agent sessions and roles;
- task or room state when persistence is required;
- conversation and steering history;
- model and runner routing;
- artifact production;
- artifact events;
- presentation policy;
- approval gates;
- execution coordination.

A runtime may emit a generic event:

```json
{
  "event": "artifact.produced",
  "task_id": "PROJECT-123",
  "kind": "plan",
  "path": "/absolute/path/to/plan.md",
  "attention": "approval"
}
```

The mdmaid presentation adapter translates that event into:

```text
register document
open when policy requires
return stable URL
```

The selected harness or collaboration runtime remains the authority for live
execution and approvals. `mdmaid.desk` is the human review surface and may grow
collaboration-room capabilities under its own contract.

### AgentnykMaisternia

Responsibility:

- provider-neutral workflow definitions;
- provider adapters and capability metadata;
- preset and collection composition;
- rendering, validation, conflict inspection, and guarded installation;
- installing the `readable-output` integration and mdmaid.desk client settings.

Maisternia does not own live task state, artifact-event history, execution
coordination, approval queues, or collaboration rooms. It configures the
harness or collaboration service that owns those concerns.

## Why Not Put Everything in `mdmaid.nvim`

An editor plugin has the wrong lifecycle for agent artifacts:

- agents run without Neovim;
- the editor may close while tasks continue;
- multiple agents may produce documents concurrently;
- documents should remain available after sessions end;
- stable browser links should not depend on an editor PID;
- other editors and scripts should be able to register documents.

The editor plugin is a client. It should not be the presentation platform.

## Why a Separate Repository

The presentation workspace has its own:

- daemon lifecycle;
- persistent state;
- security model;
- web interface;
- CLI and API;
- release cycle;
- testing strategy;
- users beyond the configurator.

Keeping it separate prevents the workflow configurator from becoming a web
server and prevents the renderer from becoming an agent orchestration system.

## Recommended Naming

### Presentation Product

Preferred product name:

```text
mdmaid.desk
```

Technical names:

```text
Git repository: mdmaid.desk
Package:        mdmaid-desk
CLI:            mdmaid-desk
Product/UI:     mdmaid.desk
```

Current CLI shape includes:

```bash
mdmaid-desk daemon start
mdmaid-desk workspace add /path/to/repository --id project-id
mdmaid-desk register plan.md --workspace project-id
mdmaid-desk list --task PROJECT-123
mdmaid-desk web
```

The separate executable avoids coupling the `mdmaid` renderer/validator and
`mdmaid.desk` catalog release cycles.

### AgentnykMaisternia

The configuration product is branded **AgentnykMaisternia**, with
**Maisternia** as the short name.

```text
Product brand:      AgentnykMaisternia
Repository:         agentnyk-maisternia
CLI/config namespace: maisternia
Installed commands:  /work-brief, /work-plan, /work-run
Reading hub:         mdmaid.desk
```

The name covers its configuration workshop scope:

- manifests and provider targets;
- safe rendering and guarded installation;
- declarative workflow DAGs and provider-native resources;
- preset and collection composition;
- provider capability and existing-configuration inspection.

## Recommended Suite

```text
mdmaid
  Rendering engine and reusable preview primitives

mdmaid.desk
  Persistent local document workspace and presentation service

mdmaid.nvim
  Optional Neovim client for mdmaid.desk

maisternia
  Provider-neutral workflow configuration, rendering, validation, and installation

agent harness or collaboration runtime
  Live sessions, routing execution, shared context, steering, and approvals
```

## Relationship to Human-in-the-Loop Workflow

```text
Agent produces artifact
        |
        v
Harness or collaboration runtime produces a document or artifact event
        |
        v
mdmaid.desk registers document
        |
        +--> passive artifact: catalog only
        |
        +--> attention required: open browser
                              |
                              v
                         human reviews
                              |
                              v
                    explicit runtime decision
```

Important invariants:

1. Registration does not interrupt the user.
2. Opening does not imply approval.
3. Approval belongs to the selected harness or collaboration runtime.
4. Mdmaid remains usable without the agent workflow.
5. The agent harness remains usable without mdmaid.
6. Neovim is an optional client, not infrastructure.

## Document Wording Correction

The broader human-in-the-loop proposal should describe `mdmaid.nvim` as a
reference implementation:

> `mdmaid.nvim` validates the multi-file preview model, but is neither required
> by nor responsible for the agent presentation service. `mdmaid.desk`
> generalizes the reusable daemon and catalog layer; `mdmaid.nvim` may become
> one of its clients later.

## Current Boundary

The implemented split is:

```text
repository:   mdmaid.desk
product name: mdmaid.desk
CLI/package:  mdmaid-desk
scope:        daemon, catalog, browser workspace, TUI, CLI/API
renderer:     mdmaid
clients:      agent workflows, scripts, humans, and future editor adapters
```

The generic artifact-event ownership and any workspace-local daemon mode remain
separate design decisions; they do not change the reading-hub contract.
