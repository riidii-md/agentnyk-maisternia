---
name: work-improve
description: Turn repeated, evidence-backed run findings into minimal harness improvement proposals and validate them before human approval.
version: 0.1.0
---

# /work-improve - Propose A Harness Improvement

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Use completed profiles and audits to improve the harness over time. Optimize
task quality and cost together. Never rewrite durable configuration merely
because one run produced a finding.

Input:

`$ARGUMENTS`

## Gather Evidence

Read the current profile and audit, prior `record.json` files selected by the
retrospective policy, accepted and dismissed findings, and any replay suite
named by the user. Preserve provenance. Group equivalent findings across runs
and distinguish:

- isolated incidents;
- repeated defects;
- provider or model limitations;
- repository-specific needs;
- harness-wide problems.

## Select The Smallest Intervention

Prefer changes in this order:

1. deterministic test, validator, lint rule, or parser;
2. clearer repository discovery or documentation lookup;
3. correction to one command, prompt, or skill activation rule;
4. removal or lazy loading of duplicated context;
5. narrowly scoped workflow gate or hook;
6. provider or model routing change;
7. durable global instruction only when enforcement cannot be deterministic.

For every proposal state the root cause, supporting runs, affected providers,
expected quality impact, expected token or latency impact, risks, rollback, and
what evidence would disprove it.

## Replay Before Installation

Create an isolated candidate configuration. Compare baseline and candidate on
held-out representative tasks that were not used to invent the proposal. Use
deterministic scorers where possible and repeat nondeterministic runs. Report
quality, cost, safety, and reliability separately; do not select only the best
run.

Stop when the candidate regresses a critical metric, exceeds its budget, changes
protected evaluation files, or lacks enough evidence. A failed replay returns
to proposal design rather than weakening the scorer.

## Human Gate

Write the proposal and replay report under:

```text
.agent-runs/retrospectives/<run-id>/proposal.md
```

Update the proposal and human-decision fields in `record.json` using the
installed retrospective record schema.

Finish with one decision request: `accept`, `reject`, `revise`, or `collect more
evidence`. Do not install, push, publish, or mutate provider configuration until
the user explicitly accepts the reviewed proposal. After acceptance, use the
normal `maisternia` plan, conflict-resolution, backup, apply, and rollback flow.

Before requesting the decision, use the installed `session-retrospective`
skill's `Centralize Completed Packages` procedure to refresh the curated
proposal and record in the central store. This evidence copy does not authorize
installation.
