# Provider-Native Experiment Loops

## Decision

`maisternia` configures experiment workflows. It does not run them.

The selected harness owns the live session, tool calls, approvals, native
continuation loop, and completion decision:

```text
maisternia preset apply --scope user --target codex --yes scored-experiment
        |
        v
start Codex normally
        |
        v
/work-experiment <objective and scorer>
        |
        v
Codex owns the scored improvement loop
```

The same preset targets Claude Code, Hermes, and Antigravity. Existing
provider-branded workflow aliases are replaced by canonical routing such as
`/work-experiment @codex -- <objective>`; the workflow definition remains
provider-neutral.

Kaji is not required for a basic loop inside one live harness. It becomes
useful when the work needs process recovery, cross-provider coordination,
global budgets, queues, or multi-host supervision.

## Why This Is Feasible

All four target harnesses expose native continuation mechanisms:

| Harness | Native continuation | Packaging and safety |
|---|---|---|
| Codex | `/goal` and blocking `Stop` hooks | Skills, hooks, sandbox, approvals, tool guards, worktrees |
| Claude Code | `/goal`, `Stop` hooks, and `/loop` | Plugins, permissions, tool guards, budgets, worktrees |
| Hermes | Persistent Ralph-style `/goal` | Skills, config, turn budgets, checkpoints, worktrees |
| Antigravity | `Stop` hook with a `continue` decision | Plugins, rules, hooks, fine-grained permissions, sandbox |

A prompt that says "never stop" is not enough. A reliable installation has
three layers:

1. a skill or command that defines the experiment protocol;
2. a tool guard that protects the scorer and restricted paths;
3. a continuation gate that checks evidence when the agent tries to stop.

The continuation decision should be deterministic when the scorer is
machine-readable. Do not use another model to decide whether a numeric target
was reached.

## Provider-Neutral Contract

A complete experiment configuration needs:

```yaml
objective: "Improve the parser benchmark"

scorer:
  command: "./scripts/score-parser --json"
  metric: "score"
  direction: "maximize"
  target: 0.92

scope:
  editable:
    - "src/parser/**"
  protected:
    - "scripts/score-parser"
    - "benchmarks/**"
    - "tests/fixtures/**"

budget:
  attempt_timeout: "10m"
  total_time: "8h"
  max_no_improvement: 12
  max_scorer_failures: 3

evidence:
  ledger: ".agent-runs/experiment.jsonl"
  inspect_diff: true

safety:
  require_worktree: true
  network: "deny"
  push: "deny"
```

The checked-in `scored-experiment` preset currently supplies the neutral
workflow and `/work-experiment` command. A later schema revision should make
the contract above structured and project-configurable.

## Runtime Shape

```mermaid
flowchart LR
    A[maisternia apply] --> B[Provider-native configuration]
    B --> C[User starts normal harness session]
    C --> D[/work-experiment objective]
    D --> E[Run baseline]
    E --> F[One focused change]
    F --> G[Run protected scorer]
    G --> H[Record score and diff]
    H --> I{Completion gate}
    I -->|continue| F
    I -->|target reached| J[Stop for review]
    I -->|budget or safety limit| K[Pause with evidence]
```

A generated provider hook may run the scorer and inspect state. That hook is
part of the installed provider configuration. It does not make `maisternia` a
daemon, runner, scheduler, or owner of the session.

## Provider Compilation

### Codex

Render:

- a neutral experiment skill or command;
- a Goal-mode completion contract;
- a `Stop` hook that continues while the target and budgets allow;
- a `PreToolUse` hook that blocks protected-path changes;
- a sandbox and approval profile;
- a persistent score ledger format.

Codex documents scored improvement loops directly: make one focused
improvement, rerun the evaluator, inspect artifacts, and keep a durable log.

### Claude Code

Render:

- a plugin or project-local skill;
- a `/goal` completion contract;
- `Stop` and `PreToolUse` hooks;
- an optional `.claude/loop.md` for longer session-scoped repetition;
- permission and budget policy;
- a persistent score ledger format.

Claude limits consecutive Stop-hook continuations, so long local runs should
combine immediate Goal-mode work with its scheduled `/loop` mechanism instead
of assuming unbounded Stop recursion.

### Hermes

Render:

- an experiment skill or bundle;
- a structured `/goal` completion contract;
- `goals.max_turns` policy;
- protected-path guards;
- checkpoint and worktree policy;
- a persistent score ledger format.

Hermes already persists active goals in its session database, continues across
turns, survives conversation compression, resumes sessions, and pauses on
interrupt or budget exhaustion.

### Antigravity

Render:

- an Antigravity plugin;
- an experiment skill;
- a `Stop` hook that returns `{"decision": "continue"}` while work remains;
- a `PreToolUse` protected-path hook;
- rules, permissions, and sandbox policy;
- a persistent score ledger format.

Capability detection must verify the installed CLI supports the required Stop
hook contract. Do not infer support from the provider name alone.

## Safe Long-Running Behavior

"Almost endless" means long-running, resumable, and evidence-driven. It must
not mean literally unbounded.

Every loop needs:

- an immutable scorer;
- one worktree per active agent;
- one focused change per attempt;
- a fixed attempt timeout;
- a score and diff ledger;
- a target or acceptance condition;
- a total wall-clock ceiling;
- a no-improvement ceiling;
- a scorer-failure ceiling;
- a user stop mechanism;
- no push, deploy, secret access, or scope expansion by default.

The completion gate should behave like this:

```text
if no active experiment in this worktree:
    allow stop

if the user requested stop:
    deactivate and allow stop

if protected files changed:
    pause and require human review

run the scorer with a timeout
append score and diff metadata to the ledger

if target reached:
    deactivate and allow stop

if a budget or failure limit was reached:
    deactivate and allow stop with status

otherwise:
    continue with the best score and next required action
```

Use one active marker per worktree to avoid affecting unrelated sessions:

```text
.agent-runs/active-experiment.json
```

## Human In The Loop

The human should not approve every attempt. Approval is required when:

- editable scope must expand;
- the scorer or protected fixtures appear wrong;
- network, secrets, push, deployment, or external mutation is requested;
- progress or failure limits are reached;
- competing branches need product judgment;
- the best final diff is ready to accept.

The normal path remains autonomous inside a narrow sandbox. The final
deliverable is the best diff, score history, failed approaches, and stop
reason, not only an agent summary.

## AgentnykMaisternia Work

The implementation should proceed in layers:

1. Keep the neutral `scored-experiment` preset and command.
2. Add a structured experiment contract to the preset schema.
3. Detect concrete provider capabilities and version constraints.
4. Implement provider compilers for skills, plugins, hooks, and settings.
5. Merge structured settings and hook files idempotently.
6. Add preview, apply, verification, drift detection, and rollback for those
   generated resources.
7. Feed the experiment ledger into `/work-brief`, `/work-showcase`, and the
   admin TUI.

Generated hooks should be self-contained. They must not call an `maisternia`
runtime service or require `maisternia` to remain running.

## Kaji Boundary

Kaji is the later supervisor for requirements outside a single live harness:

| Requirement | Harness alone | Kaji |
|---|---:|---:|
| Improve one worktree for hours | Yes | Not required |
| Native scorer gate and protected paths | Yes | Not required |
| Resume the same session manually | Usually | Not required |
| Restart after a CLI process crash | No | Yes |
| Continue through a provider outage or machine reboot | No | Yes |
| Compete branches across providers | No | Yes |
| Enforce global cost and concurrency budgets | No | Yes |
| Queue work across hosts | No | Yes |
| Rank candidates and promote a winner | No | Yes |

Kaji must consume the same provider identities, experiment contract, and
ledger format. It should launch, monitor, pause, resume, and compare configured
harnesses without reimplementing their inner scored loop.

## Sources

- [Referenced repeated-experiment pattern](https://x.com/0xcarnagee/status/2081473278376325534?s=46)
- [OpenAI: scored improvement loops](https://learn.chatgpt.com/use-cases/iterate-on-difficult-problems)
- [OpenAI: long-running work and Goal mode](https://learn.chatgpt.com/docs/long-running-work)
- [OpenAI: Codex hooks](https://learn.chatgpt.com/docs/hooks)
- [Anthropic: Claude Code hooks and Goal mode](https://code.claude.com/docs/en/hooks)
- [Anthropic: Claude Code scheduled tasks and `/loop`](https://code.claude.com/docs/en/scheduled-tasks)
- [Anthropic: Claude Code plugins](https://code.claude.com/docs/en/plugins)
- [Google: Antigravity hooks](https://antigravity.google/docs/hooks?app=antigravity)
- [Google: Antigravity plugins](https://www.antigravity.google/docs/plugins)
