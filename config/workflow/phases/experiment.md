---
name: work-experiment
description: Run a long, scored improvement loop inside the current agent harness with protected evaluation and explicit safety limits.
version: 0.1.0
---

# /work-experiment - Scored Improvement Loop

Run a long, evidence-driven experiment loop inside the current agent harness.
`agentctl` configures this workflow but does not execute or supervise it.

Input:

`$ARGUMENTS`

## Establish The Contract

Before changing files, identify and report:

1. the objective and measurable target;
2. the scorer command and machine-readable metric;
3. whether the metric is minimized or maximized;
4. editable paths;
5. protected scorer, fixture, and evaluation paths;
6. per-attempt and total wall-clock limits;
7. no-improvement and scorer-failure limits;
8. the isolated worktree and evidence ledger.

Ask for missing contract values. Do not invent a success threshold, weaken the
scorer, or treat a model judgment as a substitute for a deterministic metric.

## Run The Loop

1. Verify the worktree is isolated and the protected paths are unchanged.
2. Run the scorer once to establish the baseline.
3. Record the baseline in `.agent-runs/experiment.jsonl`.
4. Make one focused change inside the editable paths.
5. Run the scorer with the configured attempt timeout.
6. Record the score, relevant diff metadata, result, and next hypothesis.
7. Keep the best verified state and continue through the harness's native goal,
   Stop-hook, or scheduled-loop mechanism.
8. Stop for review when the target or a safety limit is reached.

Do not use an `agentctl` runtime command to continue the loop. Codex, Claude
Code, Hermes, or Antigravity owns the live session and continuation behavior.

## Stop Conditions

Stop and present evidence when any of these is true:

- the target is reached;
- total time or attempt budget is exhausted;
- the no-improvement limit is reached;
- the scorer repeatedly fails or becomes nondeterministic;
- a protected path changed;
- progress requires broader file, network, secret, push, or deployment access;
- the user requests a stop.

## Final Output

Report:

- baseline, best score, and target;
- best diff or commit;
- concise attempt history;
- failed approaches worth remembering;
- exact reason the loop stopped;
- remaining risks and the next human decision.

Do not push, deploy, expose secrets, modify protected evaluation files, expand
scope, or perform destructive Git operations without explicit authorization.
