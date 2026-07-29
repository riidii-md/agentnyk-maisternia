# Agent Workflow Configuration

## Status

This document defines the target configuration workflow for `agentctl`.

`agentctl` is a pipeline configurator for existing command-line agent
harnesses. It defines reusable provider-neutral pipeline templates and renders
those templates into provider-native commands, prompts, skills, hooks, and
settings.

Runtime dispatch and live observation are out of scope. Claude Code, Codex CLI,
Hermes, Antigravity, Kaji, and future harnesses own their own execution loops,
sessions, approvals, histories, and runtime state.

## Purpose

`agentctl` should let a user manage many providers and harnesses without having
to remember:

- which reusable pipelines and phases exist;
- which command, prompt, skill, hook, or settings files each provider needs;
- which provider mappings are installed, missing, or drifting;
- which generated files would change before apply;
- which authority boundary a generated command asks the harness to preserve.

The same configuration model should support bugs, features, refactors,
investigations, operational work, existing repositories, and new projects. The
actual work happens after the user invokes the rendered command inside the
chosen harness.

## Product Boundary

### `agentctl` owns

- Provider-neutral pipeline definitions.
- Phase prompt/command templates.
- Provider adapters and capability metadata.
- Provider-native rendering and staging trees.
- Path allowlists, ownership rules, drift detection, conflict detection, and
  backups.
- Explicit preview/apply flows.
- A TUI for designing and reviewing configuration.

### `agentctl` does not own

- Running Claude, Codex, Hermes, Antigravity, Kaji, or other harnesses.
- Watching live agent runs.
- Runtime task queues.
- Runtime phase transitions.
- Agent session history.
- Approval queues.
- Commit, push, PR, deploy, or release actions.

If the TUI observes live work, it becomes a controller. Keep it focused on
configuration authoring, render previews, provider mappings, drift, conflicts,
and explicit apply.

## Design Principles

### Pipelines Are Configuration Artifacts

A pipeline in `agentctl` is not a running job. It is a reusable declarative
artifact that can be rendered into one or more provider harnesses.

Example shape:

```yaml
id: feature-delivery
phases:
  - scout
  - plan
  - run
  - verify
  - review
targets:
  - codex
  - claude
  - hermes
```

### Commands Describe Work, Not Providers

Canonical commands use the `/work-*` namespace:

```text
/work-shape
/work-source
/work-grill
/work-brainstorm
/work-brief
/work-scout
/work-analyze
/work-research
/work-decide
/work-ready
/work-plan
/work-prove
/work-handoff
/work-run
/work-verify
/work-review
/work-pr
/work-showcase
/work-cleanup
```

The command identifies the phase. Provider rendering decides how that command is
installed in each harness. The harness decides how to execute it at runtime.

### Provider Aliases Remain First-Class

Provider-prefixed aliases are useful when the user wants an explicit harness:

```text
/codex-shape
/codex-analyze
/codex-plan
/codex-work-loop
/codex-review
/codex-pr-check
```

These aliases preserve canonical phase semantics while rendering provider-aware
instructions for a specific harness. Optional aliases such as `/claude-plan` or
`/antigravity-research` can be added when there is recurring need. The repository
should not generate every possible provider/phase combination by default.

### Rendered Files Are Configuration, Not Runtime State

`agentctl` owns declarative pipeline definitions and rendered provider files it
manages. It should not become the source of truth for live runs, phase progress,
approvals, or agent session history.

Provider runtime directories contain mixed user configuration, caches, sessions,
and generated artifacts. `agentctl` must manage only the files it rendered and
must preserve conflict, drift, backup, path, and symlink protections.

### Authority Never Expands Silently

Generated commands can describe requested authority, but only the external
harness can enforce it at runtime. Changing providers or render targets must not
silently broaden:

- filesystem access;
- network access;
- tool access;
- destructive authority;
- production access;
- permission to commit, push, or open a PR.

## End-to-End Configuration Flow

```mermaid
flowchart TD
    D[Pipeline definitions] --> C[agentctl configurator]
    A[Provider adapters] --> C
    C --> P[Preview render plan]
    P --> X[Conflict and drift checks]
    X --> I[Explicit apply]
    I --> CC[Claude commands]
    I --> CX[Codex commands]
    I --> H[Hermes skills]
    I --> AGY[Antigravity config]
    CC --> R[Harness-owned execution]
    CX --> R
    H --> R
    AGY --> R
```

Plain-text fallback:

```text
pipeline definitions + provider adapters
  -> preview generated provider files
  -> inspect conflicts and drift
  -> explicit apply
  -> run inside the chosen external harness
```

`agentctl` stops at configuration. The external harness owns running, observing,
pausing, resuming, approving, and recording live work.

## TUI Role

The TUI is a configuration studio. It should help users:

1. browse pipeline templates;
2. create or edit pipeline definitions;
3. inspect provider adapters and render targets;
4. preview generated files;
5. inspect drift and conflicts;
6. apply reviewed configuration changes explicitly.

It should not:

- dispatch agents;
- display live runs as active pipelines;
- choose the next runtime phase;
- show approval queues;
- observe or tail harness sessions;
- commit, push, or open PRs.

## Relationship To Kaji

Kaji is an execution harness: it can run multi-agent spec-driven loops. Under the
`agentctl` boundary, Kaji is a possible render target or downstream harness, not
logic to duplicate inside `agentctl`.

Clean split:

```text
agentctl = define, render, validate, install configuration
kaji     = execute multi-agent pipelines when chosen as the harness
```

If Claude Code, Codex CLI, Hermes, or Antigravity can run a pipeline from their
own native configuration, `agentctl` should configure those harnesses directly
instead of introducing a new runner.

## Current Experimental State Commands

The repository currently contains `event`, `task`, `work next`, and
`pipeline start/transition` commands from an earlier state-machine slice. Treat
these as experimental schema fixtures while the configuration model settles.
They should not be presented as the product runtime, and new work should avoid
expanding them into dispatch, observation, approval, or control features.

Future changes should either:

1. narrow those commands to configuration authoring and schema validation; or
2. move runtime execution concerns into a separate harness such as Kaji.

## Implementation Priorities

1. Define first-class pipeline template objects.
2. Add pipeline authoring commands and TUI screens.
3. Render provider-native commands, prompts, skills, hooks, and settings from
   templates.
4. Improve structured settings merge and key ownership.
5. Improve render previews, conflict explanations, and drift explanations.
6. Keep runtime dispatch and live observation out of `agentctl`.
