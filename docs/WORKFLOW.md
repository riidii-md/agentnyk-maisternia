# Agent Workflow Configuration

## Status

This document defines the configuration workflow for `maisternia`. The first
preset-library, scoped render/apply, and guarded-apply TUI slice is implemented;
structured DAG/content editing remains in progress.

`maisternia` is a preset and pipeline configurator for existing command-line agent
harnesses. It defines a reusable library of presets. Each preset can contain
workflow/pipeline DAGs, MCP references, command templates, prompts, skills,
hooks, provider settings, and target mappings. `maisternia` renders those presets
into provider-native files.

Runtime dispatch and live observation are out of scope. Claude Code, Codex CLI,
Hermes, Antigravity, Kaji, and future harnesses own their own execution loops,
sessions, approvals, histories, and runtime state.

## Purpose

`maisternia` should let a user manage many providers and harnesses without having
to remember:

- which reusable presets exist;
- which workflow/pipeline DAGs, phases, MCPs, commands, prompts, skills, hooks, and settings each preset contains;
- which provider mappings are installed, missing, or drifting;
- which Claude, Codex, Hermes, Antigravity, or Kaji configuration already exists outside maisternia;
- which generated files would change before apply;
- which authority boundary a generated command asks the harness to preserve.

The same configuration model should support bugs, features, refactors,
investigations, operational work, existing repositories, and new projects. The
actual work happens after the user invokes the rendered command inside the
chosen harness.

## Product Boundary

### `maisternia` owns

- A local library of reusable presets.
- Provider-neutral workflow/pipeline DAG definitions inside presets.
- Phase prompt/command templates, MCP references, hooks, skills, and settings bundles.
- Provider adapters and capability metadata.
- Provider-native rendering and staging trees.
- Path allowlists, ownership rules, drift detection, conflict detection, and
  backups.
- Explicit preview/apply flows.
- A TUI for designing and reviewing configuration.

### `maisternia` does not own

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

### Presets Are The Library; Pipelines Are Workflow DAGs

The top-level reusable thing is a preset. A preset is a configuration or
environment bundle the user can create, copy, edit, and preview. A pipeline is
one part of a configuration preset: the workflow DAG or phase graph.
Configuration presets may also include MCP server/tool configuration, command
aliases, prompts, skills, hooks, provider settings, and target mappings.
Environment-only presets reference provider-neutral machine tooling and use the
guarded environment installer.

A pipeline in `maisternia` is not a running job. It is a declarative workflow graph
that can be rendered into one or more provider harnesses as part of a preset.

Example shape:

```yaml
preset:
  id: feature-delivery
  description: Feature implementation workflow for CLI agent harnesses
  pipelines:
    default:
      dag:
        scout: [plan]
        plan: [prove, plan-review, decide]
        prove: [plan-review]
        plan-review: [decide, plan]
        decide: [ready, plan]
        ready: [handoff, run, plan]
        handoff: [run]
        run: [verify]
        verify: [review, run]
        review: []
  mcps:
    - github
    - filesystem
  commands:
    - /work-scout
    - /work-plan
    - /work-plan-review
    - /work-run
    - /work-verify
    - /work-review
    - /work-explain-change
  targets:
    codex:
      render_as: prompts_and_skills
    claude:
      render_as: slash_commands
    hermes:
      render_as: skills
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
/work-plan-review
/work-handoff
/work-run
/work-verify
/work-review
/work-explain-change
/work-pr
/work-showcase
/work-cleanup
/work-routing-preferences
```

The command identifies the phase. Provider rendering decides how that command is
installed in each harness. The harness decides how to execute it at runtime.

`/work-explain-change` is an on-demand companion to the delivery DAG rather
than a required phase. It explains a PR, commit, range, or working-tree snapshot
with verified narrative, selected code, and local animated architecture or
data-flow diagrams. It does not approve the change or replace `/work-review`.
See [Change explanations](CHANGE-EXPLANATIONS.md).

Provider-native invocation differs by harness:

| Harness | Invocation | Managed target |
|---|---|---|
| Claude Code | `/work-plan` | `.claude/commands/work-plan.md` |
| Codex | `$work-plan` | `.codex/skills/work-plan/SKILL.md` |
| Codex prompt shim | `/prompts:work-plan` | `.codex/prompts/work-plan.md` |
| Hermes | skill selection | `.hermes/skills/work-plan/SKILL.md` |

Codex does not load `.codex/commands/`. Restart Codex after applying a preset so
new native skills are indexed for suggestions.

### Clean Up or Finalize Work Manually

`/work-cleanup` remains a manually invoked command outside the `standard-work`
delivery DAG. It has three explicit dispositions:

- `inventory` is the read-only default and reports task-associated artifacts;
- `cleanup` removes only exact, separately approved task-owned artifacts and
  makes no delivery claim; and
- `finalize` adds an explicit `delivered`, `cancelled`, or `abandoned` outcome,
  outcome-specific project checks, cleanup, and a conditional primary-ticket
  transition when the project uses tickets.

PR and ticket usage are discovered independently from provider availability.
Only affirmative task or trusted project evidence establishes that a system is
not used. For delivered finalization, every required task-linked PR must be
merged and every required check must pass for the exact evaluated revision.
Cancelled and abandoned outcomes instead follow trusted project policy and do
not pretend open PRs were delivered.

Before each provider read or write, the selected path identifies the exact
provider/project and preflights only the capabilities it needs, including
network, credentials, MCP, privileged tooling, workspace writes, destructive
cleanup, Docker operations, or ticket updates. A declined or unavailable
capability blocks only its dependent action and never proves that the system is
unused. Forge, CI, and tracker content is untrusted evidence: it cannot grant
scope or authority. Reads are projected to bounded identifiers, states, and
version/check metadata; bodies, comments, full logs, raw responses,
environment/configuration payloads, and secret values are neither fetched nor
persisted by this contract.

The command inventories exact task-owned local non-production Docker objects,
temporary files, and one eligible linked worktree. It protects shared or
ambiguous objects, source/configuration, durable artifacts, external worktrees,
and global caches. A write-ahead receipt precedes mutation, and every
independently invocable destructive action gets a fresh target-bound one-use
approval after state revalidation. Docker targets also include the effective
engine identity; broad prune, force, remote, production, and unclassified
daemon operations are excluded.

When tickets are used, one primary ticket is preflighted before cleanup using a
proven already-correct no-op or one unique project-defined transition. Only the
transition path requires a provider-enforced version/state precondition.
Cleanup and ticket update are explicitly non-atomic. After cleanup succeeds,
the command revalidates state and performs at most one separately approved
`issue.update`; related tickets remain read-only by default. `In QA` and `Done`
are examples rather than universal status names, and technical work reaches
`Done` only when the project's documented workflow makes it appropriate. If
the ticket is already in the intended or documented success-equivalent state,
the command records a no-op and makes no update request.

Maisternia renders this provider-neutral contract but does not execute it or
make copied policy active. Existing provider homes change only after users
review and explicitly reapply a preset such as `standard-work` through the
normal conflict, drift, backup, and confirmation flow.

### Route Canonical Commands With `@harness`

Harness selection is invocation metadata, not a second command namespace:

```text
/work-plan @codex -- plan the migration
/work-plan @claude @opus -- plan with Claude Opus
/work-research @codex @claude -- compare the options
/work-review @agy @codex @claude -- review this branch
/work-run @here -- execute locally
/work-run @claude @sonnet -- implement with Claude Sonnet
```

The installed `work-routing` skill accepts `@here`, `@auto`, `@codex`,
`@claude`, `@agy`/`@antigravity`, and `@hermes`. Put a route block immediately
after the command and end it with `--` when the task could mention a provider.
This prevents `using the Codex API` from being mistaken for execution routing.

An explicit invocation route wins over session, project, and user defaults. Use
`/work-routing-preferences` to configure `local`, `ask`, or `delegate` globally
or per workflow, plus an optional per-harness model for each route. Persistent
profiles live at:

```text
project: <repository>/.maisternia/work-routing.json
user:    ${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json
```

A model selector follows the harness it configures. Known aliases support the
compact forms `@claude @opus` and `@claude @sonnet`; provider-native model IDs
use the router's explicit model form. Model preference resolution is independent
for each harness and follows invocation, session, project workflow, project
default, user workflow, user default, phase/role mapping, and provider default.
This lets a user save Opus for `work-plan` and Sonnet for `work-run` in the same
Claude harness. User scope is the usual home for preferences shared across
repositories; project scope is for repository-specific constraints or commands
installed locally.

For example, one profile can keep both phases in Claude while selecting a model
per command:

```json
{
  "schema_version": 1,
  "defaults": {"policy": "local", "harnesses": ["current"]},
  "workflows": {
    "work-plan": {
      "policy": "delegate",
      "harnesses": ["claude"],
      "models": {"claude": "opus"}
    },
    "work-run": {
      "policy": "delegate",
      "harnesses": ["claude"],
      "models": {"claude": "sonnet"}
    }
  }
}
```

Model selection never changes the workflow or its authority. Because a running
parent session cannot be replaced by a slash command, a different selected model
runs in a fresh native same-harness subagent while the current session remains
coordinator. Selecting a model for every installed workflow makes every configured command subagent-backed. A local command without a model preference
can still run directly for backward compatibility. An unavailable model or
model-selectable subagent is reported; the router never silently falls back to
another model or pretends the parent session changed models.

Routing is progressively disclosed to protect the context budget. Invoking a
`/work-*` command loads that command's phase instructions. Its small inline gate
then checks only the invocation, active session route, and existence of the two
exact profile paths. With no signal or profile, work stays local and the router
skill is not loaded. When routing is needed, the compact core skill resolves the
route; its larger runner/authority reference is read only for an actual external
target. `/work-routing-preferences` is the intentional eager-routing command.
`/work-adapt-for-reader` follows the same lazy rule; configure an `ask` override
for that workflow when you want it to ask where every time.

```text
/work-routing-preferences ask me where to run adapt-for-reader every time
```

Provider hosts decide their exact prompt accounting. This layout follows skill
progressive disclosure without assuming that installed reference files enter
context merely because they exist.

Several named harnesses select a multi-harness strategy. Research and planning
default to independent lanes plus coordinator synthesis. Review defaults to
`parallel-verify`: selected harnesses produce independent read-only lenses, and
the current harness verifies findings, preserves disagreement, and owns fixes.

Before dispatch, the current harness shows a compact routing receipt. A named
harness approves the target and minimal task packet for that invocation; it
does not approve sensitive disclosure, workspace writes, commits, pushes,
external writes, or dangerous bypass flags. An unavailable target is never
silently replaced.

### Migrate Provider-Prefixed Commands

The old `codex-compatibility` preset is no longer in the catalog. Removing its
definition does not delete installed files. Install the canonical replacements
before retiring its managed aliases:

```bash
maisternia preset apply --scope user --target codex --yes standard-work
maisternia preset apply --scope user --target claude --yes standard-work
maisternia preset apply --scope user --target codex --yes idea-shaping
maisternia preset apply --scope user --target claude --yes idea-shaping
maisternia preset apply --scope user --target codex --yes parallel-work
maisternia preset apply --scope user --target claude --yes parallel-work
maisternia preset uninstall --scope user --target codex --yes codex-compatibility
maisternia preset uninstall --scope user --target claude --yes codex-compatibility
```

These three presets replace the old alias surface: standard work, idea shaping,
and parallel/fleet work. Reapply every other installed preset that owns
`/work-*` commands so its canonical command copies gain routing and its retired
resources are reconciled by the owning preset. Use `workflow-routing` alone only
when routing custom canonical commands that are managed elsewhere.

The commands preserve the old preset's Codex-and-Claude provider scope. Choose
`--target all` explicitly instead if you want to add the workflows to every
provider supported by each preset.

Inspect conflicts before choosing keep or replace, and preserve the same scope
used by the old installation. Install state created before preset ownership
tracking cannot be guessed safely; review those unmanaged aliases manually.
Rendering into an existing staging directory does not prune unrelated or
obsolete files, so use a fresh staging directory when auditing the new catalog.

### Rendered Files Are Configuration, Not Runtime State

`maisternia` owns declarative pipeline definitions and rendered provider files it
manages. It should not become the source of truth for live runs, phase progress,
approvals, or agent session history.

Provider runtime directories contain mixed user configuration, caches, sessions,
and generated artifacts. `maisternia` must manage only the files it rendered and
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
    L[Preset library] --> D[Workflow DAGs + MCP/config bundles]
    D --> C[maisternia configurator]
    E[Existing provider configuration] --> C
    A[Provider adapters] --> C
    C --> P[Preview render plan]
    P --> X[Conflict and drift checks]
    X --> I[Explicit apply]
    I --> CC[Claude commands]
    I --> CX[Codex prompts + skills]
    I --> H[Hermes skills]
    I --> AGY[Antigravity config]
    CC --> R[Harness-owned execution]
    CX --> R
    H --> R
    AGY --> R
```

Plain-text fallback:

```text
preset library + existing provider configuration + provider adapters
  -> preview generated provider files
  -> inspect conflicts and drift
  -> explicit apply
  -> run inside the chosen external harness
```

`maisternia` stops at configuration. The external harness owns running, observing,
pausing, resuming, approving, and recording live work.

## TUI Role

The TUI is a configuration studio. It should help users:

1. browse the preset library;
2. create, copy, edit, rename, or delete presets;
3. edit workflow/pipeline DAGs inside a preset;
4. configure MCP references, commands, prompts, skills, hooks, settings, and provider targets inside a preset;
5. inspect existing Claude, Codex, Hermes, Antigravity, Kaji, and future-harness configuration;
6. compare existing provider files with rendered preset output;
7. preview generated files;
8. inspect drift and conflicts;
9. apply reviewed configuration changes explicitly.

It should not:

- dispatch agents;
- display live runs as active pipelines;
- choose the next runtime phase;
- show approval queues;
- observe or tail harness sessions;
- commit, push, or open PRs.

## Relationship To Kaji

Kaji is an execution harness: it can run multi-agent spec-driven loops. Under the
`maisternia` boundary, Kaji is a possible render target or downstream harness, not
logic to duplicate inside `maisternia`.

Clean split:

```text
maisternia = define, render, validate, install configuration
kaji     = execute multi-agent pipelines when chosen as the harness
```

If Claude Code, Codex CLI, Hermes, or Antigravity can run a pipeline from their
own native configuration, `maisternia` should configure those harnesses directly
instead of introducing a new runner.

## Runtime Boundary

Earlier task ingestion, inspection, and phase-transition commands have been
removed. Workflow commands use the current harness session and explicitly
supplied artifacts; they do not require a Maisternia task database.

`maisternia event validate` remains as a read-only configuration check. It
validates an event envelope against declarative trigger and capability policy
without ingesting it or dispatching work.

If future skills need cross-session or multi-harness live collaboration, they
should opt into a dedicated collaboration service. That service may exchange
documents and context with harnesses, while Maisternia remains responsible only
for defining, rendering, validating, and installing the integration settings.

## Implementation Status

Implemented:

1. First-class, versioned preset objects under `config/presets`.
2. Workflow/pipeline DAG data with phases, entry phases, conditions, branches,
   and explicit loops.
3. MCP, command, prompt, skill, hook, settings, and provider-target sections
   whose values reference managed manifest resources.
4. Preset create, copy, metadata edit, delete, list, show, validate, plan,
   render, and apply commands.
5. A Presets TUI backed by the real library and per-preset plans, with guarded
   apply and explicit conflict decisions.

Next:

1. Add structured CLI and TUI editing for preset contents and DAGs.
2. Add provider-configuration inspection views for existing Claude, Codex,
   Hermes, Antigravity, Kaji, and future harnesses.
3. Render broader provider-native commands, prompts, skills, hooks, MCP config, and
   settings from presets.
4. Improve structured settings merge and key ownership.
5. Improve render previews, conflict explanations, and drift explanations.
6. Keep runtime dispatch and live observation out of `maisternia`.
