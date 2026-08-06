# Hook Packs And Installation Scopes

## Purpose

Hooks are a configuration surface for small, bounded actions at provider
lifecycle events. They are useful for deterministic safety checks, session
continuity, repository-specific quality reminders, delegation contracts, local
maintenance, and redacted metrics.

They are not an autonomous runtime. A hook must have a narrow trigger, declared
authority, bounded timeout, explicit failure behavior, and a provider-native
event mapping. Model-backed hooks require a recursion guard and should be the
exception.

The initial library is an original provider-neutral design based on common
lifecycle responsibilities. It does not import another repository's scripts,
wording, paths, model policy, or project conventions.

## Included Packs

| Pack | Default scope | Activation | Responsibility |
| --- | --- | --- | --- |
| `safety` | user | global | Deny destructive actions and policy bypass; ask before sensitive reads |
| `continuity` | user | global | Restore a brief, checkpoint before compaction, capture bounded session status |
| `quality` | project | repository opt-in | Surface repository-owned checks after changes and at work boundaries |
| `delegation` | user | global | Require a delegation contract and record bounded outcome metadata |
| `maintenance` | project | repository opt-in | Detect documentation impact and refresh explicitly configured local indexes |
| `observability` | user | global | Record redacted tool/session metrics without prompts or output bodies |

`hook-standard` combines safety, continuity, and quality. `hook-complete`
contains all six packs for users who want to inspect and select the complete
surface. Each focused pack also has a matching `hook-<pack>` preset.

## Inspect And Validate

```bash
agentctl hook list
agentctl hook show safety
agentctl hook validate all
agentctl doctor
```

Hook pack files are strict JSON under `config/hooks`. Unknown fields, invalid
scopes, unsupported effects, unsafe fail-closed combinations, unbounded
timeouts, unknown providers, and model-cost rules without recursion guards are
rejected.

## Choose The Scope

Install globally for one provider user:

```bash
agentctl hook plan \
  --scope user \
  --target codex \
  hook-standard

agentctl hook apply \
  --scope user \
  --target codex \
  --yes \
  hook-standard
```

Install into a specific repository:

```bash
agentctl hook plan \
  --scope project \
  --project /path/to/repository \
  --target claude \
  hook-quality

agentctl hook apply \
  --scope project \
  --project /path/to/repository \
  --target claude \
  --yes \
  hook-quality
```

User scope resolves target paths below `--home` and stores state and backups in
`~/.config/agentctl`. Project scope resolves target paths below `--project` and
stores state and backups in `<project>/.agentctl`. The planner always prints the
resolved scope and root before file actions.

The same conflict policy applies in both scopes:

```bash
# Leave customized target files in place and remember that decision.
agentctl hook apply --conflicts keep --yes hook-standard

# Back up customized target files and replace them.
agentctl hook apply --conflicts replace --yes hook-standard
```

## Inheritance

User-global and repository-local policy are merged conceptually:

```text
effective hooks = user-global hooks + repository hooks
```

Agentctl's policy merge is asymmetric for safety. Repository policy may add
rules or tighten a user-global restriction, but it must not remove or weaken a
global deny. Advisory packs use normal additive merge behavior.
Repository-opt-in packs stay dormant unless the repository explicitly installs
or enables them.

This is a desired agentctl invariant, not a blanket claim about native provider
precedence. Some providers let project configuration disable user hooks, and
some treat a failed hook process as nonblocking. The native renderer must check
those capabilities and either use a stronger managed/system layer, preserve the
guarantee through a stable dispatcher, or reject activation. It must never label
a best-effort hook as fail-closed.

## Provider Layers

The providers expose different native configuration layers:

| Provider | User layer | Project layer or status |
| --- | --- | --- |
| Codex | `~/.codex/hooks.json` or inline config | `<repo>/.codex/hooks.json` or inline config; additive and trust-gated |
| Claude | `~/.claude/settings.json` | `<repo>/.claude/settings.json` |
| Antigravity | Adapter-managed user definition | Native settings merge and precedence still require provider-specific verification |
| Hermes | `~/.hermes/config.yaml` and user hook scripts | Project plugins are opt-in; shell-hook failures are best-effort |

Provider support is capability-driven. A missing lifecycle event is omitted
instead of approximated with an unrelated event. For example, the continuity
pack does not claim a Hermes pre-compaction event where no equivalent is
declared.

Primary provider references:

- [Codex hooks](https://learn.chatgpt.com/docs/hooks)
- [Claude Code hooks](https://code.claude.com/docs/en/hooks)
- [Gemini CLI hook reference](https://geminicli.com/docs/hooks/reference/)
- [Gemini CLI hook security and best practices](https://geminicli.com/docs/hooks/best-practices/)
- [Hermes shell hooks](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/features/hooks.md)
- [Hermes project plugin discovery](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/features/built-in-plugins.md)

## Current Activation Boundary

The current implementation validates and installs managed hook definitions at
provider-specific paths:

```text
.codex/agentctl/hook-packs/<pack>.json
.claude/agentctl/hook-packs/<pack>.json
.config/agy/agentctl/hook-packs/<pack>.json
.hermes/agentctl/hook-packs/<pack>.json
```

These files are inspectable configuration inputs, not active native hooks yet.
Activating them safely requires a structured JSON, TOML, and YAML merger that:

1. preserves unmanaged provider settings;
2. registers one stable `agentctl hook run` dispatcher per provider and scope;
3. resolves the current repository from the hook working directory;
4. merges global and repository policy with tighten-only safety semantics;
5. generates provider-native event/matcher entries;
6. verifies that native precedence and failure behavior satisfy each rule;
7. supports plan, conflict resolution, backup, rollback, and doctor checks.

Until that renderer exists, `agentctl` does not silently edit native provider
settings or imply that copied definitions execute. This boundary keeps the
first feature useful for authoring and inspection without risking existing
agent configuration.

Project-managed `.agentctl` state is local operational metadata. Repositories
should ignore it unless a future team-state format explicitly separates
shareable policy from machine-specific ownership and backups.
