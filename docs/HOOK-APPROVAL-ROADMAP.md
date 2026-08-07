# Hook And Approval Roadmap

## Goal

Agentctl should let a human author one understandable safety and automation
configuration, preview how it maps to Codex, Claude Code, Antigravity, and
Hermes, and install only provider-native behavior that preserves the declared
guarantees.

The current implementation establishes the configuration layer:

- six strict provider-neutral hook packs and scoped hook presets;
- a strict allow/ask/deny approval policy and an `approval-standard` preset;
- user-global and repository-local planning and installation;
- CLI inspection, validation, explanation, and conflict-safe apply;
- managed definitions kept separate from native activation.

The remaining work is enforcement, native compilation, simulation, and
ergonomic editing. It should remain configuration work: agentctl does not need
to become a long-running workflow controller.

## Research Conclusions

1. Authorization must execute at the tool-call boundary. Prompts and model
   self-review are useful context, but they are not an authorization mechanism.
   Progent and ClawGuard both support deterministic, programmable enforcement
   around tool use rather than relying on model compliance.
2. The default must be the least privilege that can complete the immediate
   operation. ToolPrivBench reports systematic privilege over-escalation,
   especially after failed attempts; a retry must not automatically widen
   filesystem, network, or tool authority.
3. External content is untrusted input. AgentDojo demonstrates that tool-using
   agents remain vulnerable to indirect prompt injection, so fetched content
   cannot grant permission, change policy, or authorize a later tool call.
4. Human approval must bind the exact action. Operation, target, repository,
   task, worktree, parameters or preview digest, policy version, expiry, and use
   count belong in the decision record. Changed input requires a new decision.
5. Policy composition must be monotonic: `deny > ask > allow`; project policy
   may tighten a user policy but may not weaken it. Unknown operations ask, and
   an unavailable human means deny.
6. Enforcement claims must match provider behavior. If a provider can bypass a
   hook, ignores its exit status, or lacks a matching lifecycle event, agentctl
   must report degraded or reject activation instead of calling it fail-closed.
7. Audit data must be useful without becoming a transcript archive. Record
   decisions, operation classes, timing, result, rule IDs, hashes, and bounded
   error categories; omit prompts, output bodies, credentials, and raw
   environment values.

These conclusions also align with the OWASP guidance for least privilege,
human approval of high-impact actions, constrained tools, budgets, validation,
and monitoring, and with NIST's least-privilege principle.

## Proposed Architecture

```text
canonical hook packs + approval policy
                 |
                 v
        policy/capability compiler
        | validate  | explain gaps
        v           v
 native settings   enforcement report
        |
        v
 provider invokes short-lived agentctl hook dispatcher
        |
        +--> deterministic rule and requirement evaluation
        +--> bounded handler execution
        +--> allow / ask / deny result
        +--> redacted audit record
```

The dispatcher should be a short-lived provider-invoked process, not a daemon.
It reads a versioned JSON event from stdin, resolves user and project policy,
validates the event, evaluates deterministic rules, executes only declared
handlers, returns provider-native output, and exits within a strict timeout.

Decision and execution should remain separate packages. A handler cannot alter
its own authority, mint an approval, suppress an audit record, or recursively
invoke the dispatcher.

## Delivery Plan

### P0: Structured Settings Engine

Build round-trip-preserving JSON, TOML, and YAML mergers with:

- field-level ownership instead of whole-file replacement;
- provider and scope-specific merge rules;
- preview of the exact native diff;
- unmanaged-value preservation;
- conflict decisions, backups, atomic writes, rollback, and drift detection;
- strict symlink, path, permission, and file-size checks.

This is the prerequisite for safely registering hooks or permission policy in
existing provider files.

### P0: Capability Compiler

Compile each hook event and approval operation against versioned provider
capabilities. Produce one of:

- `enforceable`: native behavior preserves the policy;
- `degraded`: advisory behavior only, with the exact missing guarantee;
- `unsupported`: no meaningful mapping;
- `rejected`: activating it would weaken a required deny or approval gate.

Compilation must include event availability, matcher semantics, rule
precedence, exit-code behavior, timeout behavior, configuration precedence,
trust requirements, approval modes, and known bypasses.

### P0: Decision Engine And Grants

Implement deterministic operation classification, requirement evaluation, and
grant verification:

- exact operation and target matching;
- normalized paths and repository/worktree identity;
- preview/parameter and policy digests;
- monotonic user/project policy merge;
- expiry and atomic use counters;
- deny-on-invalid, expired, replayed, or unverifiable grants;
- append-only, redacted decision records with tamper-evident hash chaining.

Use the operating-system credential store for any future signing key. Do not
put signing secrets in managed presets or provider configuration.

### P0: Hook Dispatcher And Initial Handlers

Add `agentctl hook run` with strict stdin/stdout schemas, no ambient network,
bounded input and output, recursion protection, deterministic timeouts, and
provider adapters. Initial handlers should cover:

- destructive-operation and policy-bypass guards;
- sensitive-path classification;
- approved repository-check discovery;
- continuity checkpoint metadata;
- delegation contract validation;
- redacted timing and outcome metrics.

Model-backed analysis remains optional and advisory. It cannot return allow or
create an approval grant.

### P1: Native Provider Renderers

Implement and test renderers independently for Codex, Claude Code,
Antigravity, and Hermes. Each renderer needs golden fixtures for user and
project scope, existing unmanaged configuration, provider-version changes,
rollback, and unsupported mappings.

Native activation is complete only when `doctor` proves that the expected
provider file references the expected dispatcher and policy digest.

### P1: Simulation And Adversarial Replay

Add a no-side-effect simulator:

```bash
agentctl approval simulate --event fixture.json --target claude
agentctl hook test safety --event fixture.json --target codex
```

Test prompt-injected content, encoded sensitive paths, symlink traversal,
renames between preview and execution, shell indirection, repeated grants,
provider timeouts, malformed events, nested agents, transient tool failure,
scope changes, and policy updates. Include a provider-version compatibility
matrix in CI.

### P1: TUI Authoring And Resolution

Extend admin with dedicated Hooks and Approvals views:

- inspect each pack, rule, trigger, handler prompt, authority, and provider map;
- edit preset contents and DAGs through structured forms;
- choose user or project scope;
- compare canonical and generated native configuration;
- resolve each conflict with keep, replace, import, or cancel;
- simulate an event before activation;
- show installed, compiled, active, degraded, drifted, or unenforceable status;
- show pending human decisions with exact action previews.

The TUI remains an administrator for configuration and bounded approvals. It
does not dispatch pipelines or observe live agent reasoning.

### P2: Managed Policy And Supply Chain

Add optional organization policy layers that repositories can tighten but not
weaken. Sign release artifacts and handler bundles, verify provenance before
activation, pin provider capability data to tested versions, and expose stale
capability warnings. Add import and classification for existing native config
without claiming ownership of sessions, transcripts, caches, or history.

## Acceptance Criteria

Native activation should not ship until all of these are true:

1. Existing provider settings survive a round trip byte-equivalent where not
   touched and semantically equivalent where edited.
2. Every active rule has an enforcement report and tested provider mapping.
3. A repository cannot weaken a user or managed deny.
4. A model or delegated agent cannot grant, widen, or replay approval.
5. Changed operation input invalidates the reviewed grant.
6. Unknown operations and failed requirements cannot execute automatically.
7. Doctor detects removed hooks, changed policy, stale provider capabilities,
   unsafe approval modes, and bypass settings.
8. Rollback restores both native configuration and agentctl ownership state.
9. Audit fixtures prove that prompts, outputs, secrets, and raw environment
   values are not retained.
10. CI covers all four providers, both scopes, conflict modes, hostile events,
    and provider-version fixtures.

## Sources

- [Progent: Programmable Privilege Control for Language Model Agents](https://arxiv.org/abs/2504.11703)
- [When Lower Privileges Suffice: A Benchmark of Agent Permission Use](https://arxiv.org/abs/2606.20023)
- [AgentDojo: A Dynamic Environment to Evaluate Prompt Injection Attacks and Defenses](https://arxiv.org/abs/2406.13352)
- [ClawGuard: Deterministic Security for Tool-Using Agents](https://arxiv.org/abs/2604.11790)
- [OWASP AI Agent Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/AI_Agent_Security_Cheat_Sheet.html)
- [OWASP Excessive Agency](https://genai.owasp.org/llmrisk/llm062025-excessive-agency/)
- [NIST SP 800-171 Rev. 3](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/800-171r3/NIST.SP.800-171r3.html)
- [Codex configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference)
- [Claude Code permissions](https://code.claude.com/docs/en/permissions)
- [Antigravity permissions](https://antigravity.google/docs/cli/permissions)
- [Hermes security](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/security.md)
