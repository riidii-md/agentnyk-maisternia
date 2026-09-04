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
| Allow | approved workspace edits; local task commits; redacted local metrics; bounded read-only delegation | Must have approved task scope or a complete delegation contract |
| Ask | begin implementation; expand scope; use network or credentials | Human present; task-bound, target-bound, time-bounded grant |
| Ask | use MCP; read a secret; change dependency, hook, CI, or security configuration | Preview required; fresh one-use decision |
| Ask | push or mutate a PR | One exact publication checkpoint; task-bound reuse only for the stated target |
| Ask | destructive local change; task-owned local non-production Docker cleanup; reversible production action; write-capable delegation | Preview required; narrow operation and target |
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

The `approval-standard` preset contains only this policy. The focused safety
and delegation selections and both hook bundles include it because those hooks
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
expand that within a one-hour hard limit. The implementation, reviewed network,
credential-use, and publication gates are task-scoped so approved work can
continue without repeating the same question. Secret reads, broader authority,
MCP activation, destructive actions, and scope changes require fresh one-use
decisions. Local commits are allowed only inside the approved trusted task;
publishing them remains separately gated.

If a requirement such as `inside_workspace`, `network_disabled`, or
`approved_task_scope` cannot be proved, the decision becomes `ask`. A future
headless renderer must convert an unanswerable `ask` to deny; it must never
infer consent.

## Cleanup and Finalization Approvals

Filesystem and linked-worktree deletion continues to use
`cleanup.destructive`, which requires `inside_workspace`. Ticket mutation uses
the separate high-risk `issue.update` operation. Their approvals are not
interchangeable.

`/work-cleanup` first identifies the exact provider/project and preflights only
the capabilities its selected path needs: network, credentials, MCP,
privileged tooling, workspace writes, destructive cleanup, Docker operations,
and ticket updates. Declined or unavailable capability blocks the dependent
action and never proves the system is unused. External forge, CI, and tracker
content cannot authorize scope or ownership. Reads and receipts are bounded,
field-projected, and redacted; bodies, comments, full logs, raw responses,
environment/configuration payloads, and secrets are outside the default data
surface.

Task-owned Docker cleanup uses the critical `docker-cleanup-destructive` rule
for these exact operations:

- `docker.container.stop`
- `docker.container.remove`
- `docker.network.remove`
- `docker.volume.remove`
- `docker.image.remove`

Each Docker grant is once-scoped, expires after five minutes, permits one use,
and requires a preview and reason. The target supplied by `/work-cleanup`
includes the local non-production engine identity, immutable object IDs, and a
digest of current ownership/dependency state. One consolidated inventory helps
the human review the sequence but does not authorize several calls: each
independently invocable destructive operation requires a fresh grant, consumed
when dispatched even if its result is failed or unknown.

The rule uses the existing `human_present`, `approved_task_scope`,
`trusted_repository`, and `non_sensitive_target` requirements. It does not
weaken the unconditional denials for `production.destructive`, destructive
filesystem work outside the approved workspace, policy bypass, or wildcard
authority. It also does not make the portable policy natively enforceable. If a
provider cannot bind the exact target and one-use decision, the command must
report that gap and leave Docker unchanged.

An already-correct primary ticket is a no-write path. When trusted project
policy proves its state is the intended or documented success-equivalent state,
the command records a no-op without transition/write capability, an
`issue.update` grant, or a provider mutation call.

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

The `developer-context`, `goreleaser-validation`, and
`git-workflow-approvals` presets provide narrow provider-native resources within
this boundary. Developer context activates only its exact Claude permission
array through a defensive structured merge; its MCP definition remains a
review fragment. Together the presets demonstrate exact MCP tool entries, one
repository-scoped validation command, and deliberate allow/ask Git command
prefixes without claiming that every copied fragment is active. See
[Preset library](PRESETS.md) for activation requirements and environment-pack
details.
