# Configurator Architecture

## Goal

Manage useful CLI-agent configuration from one version-controlled repository
without copying credentials, sessions, caches, or runtime databases.

## Source and Targets

```mermaid
flowchart LR
    REPO[Configuration repository] --> MANIFEST[Validated manifest]
    MANIFEST --> PLAN[Plan and conflict detection]
    PLAN --> RENDER[Staging render]
    PLAN --> APPLY[Guarded apply]

    APPLY --> CODEX[Codex home]
    APPLY --> CLAUDE[Claude home]
    APPLY --> AGY[Antigravity prompts and configuration]
    APPLY --> HERMES[Hermes home]

    LOCAL[Machine-local overlay] --> PLAN
    SECRET[Environment or keychain references] --> CODEX
    SECRET --> CLAUDE
    SECRET --> AGY
```

## Current Commands

### Doctor

```bash
maisternia doctor
```

Validates:

- manifest schema version;
- resource IDs;
- source files;
- provider names;
- target roots;
- duplicate destinations;
- path traversal;
- managed file size.

### Inventory and Plan

```bash
maisternia inventory --target all
maisternia plan --target codex
```

Actions:

- `CREATE`: target does not exist;
- `UNCHANGED`: target matches source;
- `UPDATE`: target matches the last installed checksum and source changed;
- `REMOVE`: a preset no longer declares a target that it exclusively owns;
- `RELEASE`: a preset no longer declares a shared target, but another preset
  still owns it;
- `CONFLICT`: target is unmanaged, drifted, unsafe, or a symlink.

`REMOVE` and `RELEASE` are preset-scoped actions. Full-manifest planning does
not infer deletion ownership.

### Render

```bash
maisternia render --target all --output ./build/rendered
```

Renders provider files into an isolated staging directory.

### Apply

```bash
maisternia apply --target codex --yes
```

Apply:

- refuses conflicts;
- requires explicit `--yes`;
- rechecks source and target checksums;
- backs up managed updates and removals;
- writes atomically;
- stores checksums in an install-state file.

Preset apply records ownership per preset and target. On the next apply of the
same preset, targets removed from that preset are reconciled across MCP
references, commands, prompts, skills, hooks, and settings. An unchanged
exclusive target is backed up and removed. A locally changed target becomes a
conflict. A target shared by another applied preset remains installed and only
the obsolete ownership edge is released.

To remove every target owned by a preset, including after its catalog
definition has been deleted, use:

```bash
maisternia preset uninstall --scope user --target codex --yes <preset-id>
```

Deletion remains explicit and scope-specific; editing or pulling the catalog
does not mutate provider homes in the background. State written before schema
version 3 has no preset identity, so it is never guessed as removable. Applying
the current preset once records ownership for later reconciliation.

Install state:

```text
~/.config/maisternia/install-state.json
```

Backups:

```text
~/.config/maisternia/backups/<timestamp>/
```

For rename compatibility, `maisternia` reads the legacy
`~/.config/cli-agent-configurator/install-state.json` when the new state file
does not exist. The next successful apply writes the state under
`~/.config/maisternia/`; the legacy file is left untouched.

### Provider Inspection

```bash
maisternia provider list
maisternia provider inspect antigravity
maisternia provider doctor all
maisternia provider capabilities hermes
```

Provider inspection:

- resolves canonical identities and aliases;
- locates configured executables;
- reads bounded version output;
- checks provider configuration roots without reading their contents;
- rejects roots that traverse symlinks;
- reports static runner and parser safety contracts.

`provider doctor` does not execute native provider doctor commands. It reports
whether each native doctor exists and whether its checked-in contract marks it
safe for an explicit future invocation.

## Provider Roots

The first manifest schema allows targets only under:

| Provider | Root |
|---|---|
| Codex | `~/.codex/` |
| Claude | `~/.claude/` |
| Antigravity (`agy` alias) | `~/.config/agy/` |
| Hermes | `~/.hermes/` |

The allowlist prevents a manifest from writing arbitrary home-directory paths.

The Antigravity row is a compatibility target used by the current phase-prompt
manifest. It is not Antigravity's provider-owned settings root. Provider
inspection uses the current CLI roots:

| Purpose | Antigravity root |
|---|---|
| CLI settings and runtime | `~/.gemini/antigravity-cli/` |
| Global customizations, skills, and plugins | `~/.gemini/config/` |

A later renderer migration will map canonical workflow resources into provider
recognized structures under `~/.gemini/config/`. Until that mapping is
implemented and verified, the compatibility prompt tree remains managed but is
not treated as proof that Antigravity consumes those files.

## Prompt Source Boundaries

The command catalog has two workflow layers.

### Canonical Work Phases

Files under `config/workflow/phases/` define provider-neutral behavior for
`/work-*`. They describe inputs, gates, outputs, and authority without embedding
a provider invocation.

The same source can be rendered for Codex, Claude, Antigravity, and supported
Hermes skills.

### Shared Work Routing

`config/workflow/skills/work-routing/SKILL.md` resolves optional leading
`@harness` route blocks for canonical commands. It owns:

- explicit, session, project, and user routing precedence;
- independent per-harness model precedence at those same scopes;
- provider availability and authority checks;
- minimal redacted handoff packets;
- safe single- and multi-harness strategies;
- unavailable-target behavior and routing receipts;
- coordinator-owned verification and delivery.

The installed harness owns the actual native subagent or provider-runner call.
The `maisternia` binary still does not dispatch runtime work.

`/work-routing-preferences` proposes schema-valid global and per-workflow
profiles. Its guided setup walks each installed canonical command, asks for the
harness and optional per-harness model, then offers session-only, user-global,
or repository-local persistence with an exact diff. User-global is normally
recommended for user-installed command sets; repository-local is recommended
for project-installed commands or repository constraints. A model choice also
selects a fresh native subagent in the current harness, so choosing models for
the complete command matrix makes the workflow subagent-backed while the parent
session coordinates it. Provider-prefixed
workflow commands are not generated; for example, use
`/work-plan @claude @opus -- <task>` and
`/work-run @claude @sonnet -- <approved-plan>`.

Repository tests verify the routing inventory, canonical command integration,
multi-harness review contract, provider-branded command removal, showcase
presentation integration, and absence of personal absolute paths.

Users migrating from the removed `codex-compatibility` preset must first apply
or reapply the canonical workflow presets they use, including `standard-work`,
`idea-shaping`, and `parallel-work` for replacement coverage, and only then
uninstall the remembered alias preset for Codex and Claude. Reapplication lets
each owning preset refresh its `/work-*` files and reconcile retired resources;
`workflow-routing` alone cannot update another preset's files. Catalog changes
never delete provider-home files in the background. Preserve the old Codex and
Claude target scope by default; `--target all` is an explicit choice to add the
workflows to every supported provider. A render output directory is likewise
additive; use a fresh staging directory when checking that retired aliases are
absent.

Showcase writes durable Markdown under `.agent-runs/showcase/` and registers it
with `mdmaid-desk`; the terminal environment pack manages that CLI.

## Managed Data

Good repository-managed candidates:

- commands;
- personal skills;
- policies;
- global instructions;
- model roles;
- runner routing;
- MCP definitions using secret references;
- desired plugin and skill sources;
- helper scripts;
- non-secret machine and repository profiles.

## Unmanaged Data

Never commit or overwrite:

- tokens and credentials;
- auth databases;
- session transcripts;
- runtime state databases;
- checkpoints;
- logs;
- caches;
- downloaded plugin caches;
- raw environment values;
- temporary prompt and output files.

## Structured Settings Merge

Whole-file replacement is not sufficient for every provider:

- Codex `config.toml` mixes stable preferences with project trust and marketplace
  metadata.
- Claude `settings.json` mixes stable permission and UI preferences with values
  that Claude may update.

A later manifest version will declare managed keys and use format-aware merge:

```text
parse installed file
  -> overlay declared managed keys
  -> preserve unknown and runtime-owned keys
  -> show diff
  -> backup
  -> atomic write
  -> provider doctor
```

## Profiles

Planned precedence:

```text
base
  -> machine
  -> repository
  -> local untracked override
```

Profiles may control:

- model roles;
- runner order;
- enabled MCPs;
- provider timeouts;
- optional skills and plugins;
- repository-specific profiles;
- non-secret machine paths.

Secrets remain external and are referenced by name.

## One-Way Synchronization

Initial direction:

```text
configuration repository -> provider targets
```

Changes made directly in provider homes are reported as drift. Importing them
back into the repository will be explicit and reviewable.

Automatic bidirectional synchronization is intentionally out of scope until
ownership and conflict behavior are proven.

## Next Engineering Milestones

1. Add schema files for the configuration manifest and install state.
2. Map canonical resources to each provider's recognized configuration layout.
3. Add JSON and TOML managed-key merge.
4. Add `import` into a staging directory.
5. Add `diff`, `backup`, and rollback subcommands.
6. Add profiles, local overlays, and plugin or skill lockfiles.
7. Resolve render requirements against provider capability contracts.
8. Add preset-library authoring, workflow DAG editing, existing provider-config inspection, and provider-native render previews.

Untrusted event envelopes can be checked against the declarative workflow
policy without creating runtime state. See [Event validation](EVENT-VALIDATION.md).
