---
name: work-delegated-review
description: Delegate independent review lenses and finding verification across explicitly selected agent providers while preserving read-only authority, evidence grounding, confidentiality, and coordinator-owned fixes.
version: 0.1.0
---

# /work-delegated-review - Cross-Provider Multi-Lens Review

Run `/work-review` or `/work-plan-review` semantics while assigning independent
lenses to selected providers.

Input:

```text
/work-delegated-review mode=<auto|plan|plan-delta|implementation> \
  providers=<codex,claude,antigravity,hermes> <target and focus>
```

Naming providers is explicit approval to send the disclosed, redacted review
packet to those providers for this run. Before dispatch, show the provider list,
files or excerpts to be shared, sensitive categories, and estimated concurrency
or budget. Require a new decision if the disclosure scope expands.

## Select Eligible Delegates

Inspect installed provider and runner capabilities at execution time. Under the
current conservative contracts:

| Provider | Automatic delegated review |
|---|---|
| `codex` | read-only headless or native subagents |
| `claude` | read-only headless or native subagents |
| `antigravity` (`agy`) | read-only headless or native subagents; text result |
| `hermes` | interactive only; never unattended one-shot execution |

Hermes one-shot mode bypasses approvals, so it is not an automatic delegate.
Do not broaden authority, bypass permissions, or silently substitute a provider.
If a selected provider is unavailable, record it and ask whether to continue
with the remaining providers or run the lane locally.

Use provider-native safe interfaces discovered from the installed CLI and the
checked-in adapter. Safe baseline intent is equivalent to:

- Codex: ephemeral execution with a read-only sandbox;
- Claude: print mode with plan permission and read-only tools;
- Antigravity: print mode with plan mode and sandbox enabled;
- Hermes: an explicitly supervised interactive session only.

Do not use dangerous bypass flags. Runtime invocation remains the harness's
responsibility; `maisternia` installs this contract but does not dispatch agents.

## Prepare Minimal Review Packets

Redact credentials, personal data, customer data, proprietary context not
needed by the lens, and unrelated conversation history before cross-provider
dispatch. Give each reviewer only:

- resolved mode, target, contract, and lens;
- repository instructions and exact base ref;
- relevant plan sections, diff, files, tests, and verification evidence;
- required output fields and report schema;
- read-only authority, token/time budget, and stop conditions.

Each reviewer must read the actual code relevant to its claim. A supplied diff
or summary is not sufficient evidence.

## Dispatch And Verify

Assign one reviewer per lens and parallelize independent lanes within budget.
For high-risk lenses, independent provider diversity is preferable to repeated
copies of the same model. For every candidate finding, assign a verifier that
did not originate the claim, preferably from another selected provider. The
verifier must try to refute the finding and return `is_real`, `grounded`,
rationale, and evidence.

Keep only `is_real && grounded`. Deduplicate across providers and preserve
material disagreements in the report. Provider consensus is not proof.

## Report And Apply

The coordinator reports provider/model attribution, failed or unavailable
lanes, confirmed findings, refuted findings and why, token/cost data when
available, and disclosure scope. The coordinator then applies every confirmed
fix under `/work-review` rules and runs focused plus final verification.

Provider delegates never edit the working tree, commit, push, publish, or widen
scope. Critical and High findings remain blocking until fixed or explicitly
blocked. Write schema-valid artifacts under `.agent-runs/reviews/<run-id>/`.
