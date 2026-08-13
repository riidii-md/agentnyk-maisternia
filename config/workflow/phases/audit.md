---
name: work-audit
description: Audit a completed agent run for correctness, observable trajectory efficiency, process and safety, and cost using cited evidence.
version: 0.1.0
---

# /work-audit - Audit A Completed Run

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Audit one explicitly supplied run or the current completed task. Treat the
transcript as an observable event record, not as access to hidden reasoning.
Do not modify product code or durable agent configuration during the audit.

Input:

`$ARGUMENTS`

## Intake And Redaction

Identify the task contract, transcript or event export, final artifact, relevant
repository state, verification logs, provider telemetry, and current harness
profile. Process only explicitly selected inputs. Redact secrets before sending
content to another model or provider and record which evidence was unavailable.

## Audit Lanes

### Correctness

Extract delivered factual and completion claims at the smallest independently
verifiable unit. Evaluate code and edits by behavior rather than line-level
opinion. Prefer evidence in this order:

1. tests, type checks, linters, builds, and final environment state;
2. repository files, diffs, tool results, and pinned documentation;
3. external primary sources;
4. model judgment, clearly labeled as judgment.

Use `SUPPORTED`, `REFUTED`, `UNVERIFIABLE`, `SUPERSEDED`, or `NOT_APPLICABLE`.
Do not count unknown claims as correct.

### Observable Trajectory

Reconstruct only observable actions: messages, reads, searches, edits, failed
commands, retries, user corrections, tests, and reversals. Identify a detour
only when evidence available at that point supported a materially shorter path.
Label uncertain causal explanations as hypotheses.

### Process And Safety

Check instruction compliance, scope control, authority boundaries, sensitive
data handling, unrelated changes, destructive actions, and unsupported
completion claims. A transcript review reports observed failures; it does not
prove safety against unobserved adversarial cases.

### Cost

Record exact input, output, cache, and reasoning tokens, model calls, subagent
calls, tool calls, retries, permissions, and elapsed time when available. Keep
missing values `unknown`. Separate useful verification cost from avoidable
detour cost.

## Output

Write a self-contained report to:

```text
.agent-runs/retrospectives/<run-id>/audit.md
```

Update `record.json` in the same directory using the installed retrospective
record schema. Store structured metrics and findings there so later runs can
aggregate them without parsing Markdown.

Include:

- final task outcome and verification;
- separate correctness, trajectory, process/safety, and cost scorecards;
- finding table with ID, severity, evidence, confidence, and escaped status;
- user corrections and defects caught before delivery;
- measured values and explicit unknowns;
- repeated-pattern candidates, without changing configuration;
- evidence index and remaining uncertainty.

Do not emit one universal accuracy score. If model judges are used, record their
identity, blind generator identity where possible, randomize pair ordering, and
send disputed high-risk findings to a human.
