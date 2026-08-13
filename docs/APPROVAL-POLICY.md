# Approval Policy

## Purpose

Agent approval must be a deterministic policy decision, not a model judgment.
The standard maisternia policy classifies a provider-neutral operation as:

- `allow`: execute automatically only when every declared requirement holds;
- `ask`: require an explicit human decision for a bounded operation;
- `deny`: refuse the operation, including requests to bypass or rewrite policy.

Unknown operations and unmet allow-rule requirements resolve to `ask`. Rule
precedence is always `deny`, then `ask`, then `allow`, so a narrower deny cannot
be weakened by a broader rule.

The canonical definition is
[`config/policy/approval.json`](../config/policy/approval.json). Its schema and
Go validator reject unknown fields, duplicate operations, unsafe grant shapes,
model reviewers, delegable approvals, unbounded lifetimes, and unrecorded
decisions.

## Standard Policy

| Decision | Examples | Conditions |
| --- | --- | --- |
| Allow | trusted repository reads; public research; offline repository checks | Must remain inside declared boundaries and satisfy every requirement |
| Allow | approved workspace edits; redacted local metrics; bounded read-only delegation | Must have approved task scope or a complete delegation contract |
| Ask | begin implementation; expand scope; use network or MCP; access credentials | Human present, reason required, target-bound and time-bounded grant |
| Ask | dependency, hook, CI, security configuration, commit, push, PR, external write | Preview required; normally one use |
| Ask | destructive local change; reversible production action; write-capable delegation | Preview required; narrow operation and target |
| Deny | bypass policy, hooks, sandbox, or approvals | Cannot be approved by an agent or delegated reviewer |
| Deny | export credentials, raw environment, secrets, or private keys | Sensitive material must not become an agent artifact |
| Deny | destructive production action, outside-workspace deletion, history rewrite | Outside the standard agent authority envelope |
| Deny | self-grant, grant widening, grant delegation, approval-record mutation | Prevents the subject from authorizing itself |
| Deny | wildcard tool, network, MCP, or permission grants | Authority must be explicit and narrow |

Use the CLI to inspect the exact operation list and rationale:

```bash
maisternia approval list
maisternia approval explain repository.read
maisternia approval explain git.push
maisternia approval explain approval.self_grant
maisternia approval validate
```

Install the managed definition for one provider at user or project scope:

```bash
maisternia approval plan --scope user --target codex
maisternia approval apply --scope user --target codex --yes

maisternia approval plan \
  --scope project \
  --project /path/to/repository \
  --target claude
```

The `approval-standard` preset contains only this policy. The safety,
delegation, standard, and complete hook presets include it because those hooks
refer to the same authority boundary.

## Approval Grants

An approval is not a reusable phrase such as "the user said yes." A valid grant
must be:

- issued by a human, never by a model or subagent;
- non-delegable;
- bound to operation, target, repository, worktree, task, and policy digest;
- limited by expiry and use count;
- invalidated when any bound scope changes;
- written to a redacted approval record.

The standard default is one use for 15 minutes. Individual rules can narrow or
expand that within a one-hour hard limit. The implementation gate is task-scoped
so approved work can continue, while publication, sensitive access, and
destructive actions require fresh one-use decisions.

If a requirement such as `inside_workspace`, `network_disabled`, or
`approved_task_scope` cannot be proved, the decision becomes `ask`. A future
headless renderer must convert an unanswerable `ask` to deny; it must never
infer consent.

## Provider Mapping

The portable policy is stricter than a prompt, but it is not yet active native
configuration. Each renderer must compile only guarantees supported by the
selected provider and report any gap.

| Provider | Native controls to use | Required maisternia stance |
| --- | --- | --- |
| Codex | sandbox profile plus approval policy, including granular categories | Keep durable approval human-only; never let automatic review mint or widen a grant |
| Claude Code | deny/ask/allow rules, permission modes, managed policy, and `PreToolUse` hooks | Preserve deny precedence and disable modes that bypass or automatically approve protected actions |
| Antigravity | deny/ask/allow permission rules | Override broad workspace defaults where needed and show unsupported operation mappings |
| Hermes | manual/smart/off approvals, command checks, and hooks | Prefer manual approval; treat smart review as advisory; reject off/yolo for this policy |

Primary references:

- [Codex configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference)
- [Codex permissions](https://learn.chatgpt.com/docs/permissions)
- [Claude Code permissions](https://code.claude.com/docs/en/permissions)
- [Claude Code hooks](https://code.claude.com/docs/en/hooks)
- [Antigravity permissions](https://antigravity.google/docs/cli/permissions)
- [Hermes security](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/security.md)
- [Hermes hooks](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/features/hooks.md)

## Current Boundary

This change validates, explains, plans, and installs the provider-neutral
policy. It does not yet edit a provider's native settings or intercept tool
calls. A managed file under `maisternia/policy` is therefore a policy input, not
proof of enforcement.

Activation requires the settings merger, capability compiler, native renderer,
and hook decision engine described in
[Hook and approval roadmap](HOOK-APPROVAL-ROADMAP.md). Doctor and the TUI must
eventually distinguish `installed`, `compiled`, `active`, `degraded`, and
`unenforceable`; they must never report a copied definition as active.

The `developer-context` and `goreleaser-validation` presets provide narrow,
provider-native review fragments within this boundary. They demonstrate exact
MCP tool allow entries and one repository-scoped validation command without
claiming that copied fragments are active. See [Preset library](PRESETS.md) for
their activation requirements and pinned environment packs.
