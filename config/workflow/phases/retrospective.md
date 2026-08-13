---
name: work-retrospective
description: Run the controlled post-task profile, audit, and harness-improvement workflow and produce a human review package.
version: 0.1.0
---

# /work-retrospective - Post-Run Improvement Workflow

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Run a controlled retrospective after a completed task. `maisternia` configures
this workflow; the current Codex, Claude, Antigravity, or Hermes harness executes
it.

Input:

`$ARGUMENTS`

## Workflow

1. Create a run ID, `.agent-runs/retrospectives/<run-id>/`, and a `record.json`
   that conforms to the installed retrospective record schema.
2. Run the `/work-profile` procedure against the active harness configuration.
3. Run `/work-session-analysis` against the explicitly selected completed run;
   use `/work-audit` for its correctness, trajectory, process/safety, and cost
   evidence contract.
4. Compare the result with earlier accepted and dismissed findings when they
   are available.
5. Run the `/work-improve` procedure only for evidence-backed repeated patterns
   or a high-severity escaped defect.
6. Produce one review index linking the profile, audit, evidence, and any
   proposal.
7. Stop at the human decision gate.

The four report lanes remain separate:

```text
CORRECTNESS
OBSERVABLE-TRAJECTORY EFFICIENCY
PROCESS AND SAFETY
COST
```

## Boundaries

- Do not crawl unrelated provider session history.
- Do not expose secrets or private transcript content to an unapproved model.
- Do not infer precise token or monetary cost from text length.
- Do not claim access to hidden reasoning.
- Do not turn model agreement into ground truth.
- Do not automatically edit durable prompts, skills, MCP configuration, hooks,
  settings, provider routing, or global instructions.
- Do not run an endless self-review loop.

The run is complete when the report package is readable and the next human
decision is explicit. Configuration installation is a separate approved action.
