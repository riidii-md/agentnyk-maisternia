# Human Collaboration in Workflows

Standard work and idea shaping include collaboration after analysis, after any
research, and when presenting direction and a plan. The harness enters
`/work-collaborate` automatically. The human can discuss the work even when the
agent has no unanswered questions.

```text
SCOUT -> ANALYZE -> COLLABORATE -> [RESEARCH -> COLLABORATE]
    -> [DIRECTION -> COLLABORATE + AI REVIEW CHOICE -> DIRECTION DECISION]
    -> PLAN -> [PROVE] -> COLLABORATE + AI REVIEW CHOICE -> PLAN DECISION
```

Research, direction, and expanded proof retain their existing conditions. Idea
shaping also compares and challenges options before direction or an option
choice. Each completed analysis or research pass reaches its own collaboration
checkpoint before the flow advances. Internal evidence gathering may return to
scout before analysis or direction is complete.

## Conversation Modes

| Mode | Use |
|---|---|
| `discuss` | Ask questions, correct assumptions, compare findings, and develop ideas together. This is the default. |
| `grill` | Resolve an important human constraint with one focused question at a time, using `/work-grill`. |
| `human-draft` | Describe the desired solution in rough notes or conversation. The agent helps develop it and preserves it as attributed input. |

After research, the agent offers human-draft mode before direction. If research
is unnecessary, this opportunity belongs to analysis collaboration. A draft
can describe behavior, user experience, architecture, boundaries, or compromises.
It does not require a separate file and does not approve a solution.

Answering a grill question keeps collaboration open. The agent incorporates
the answer and handles further discussion until the human explicitly continues.
"Continue" or "go to direction" can close discovery discussion. A reply can
both answer and continue. Reading a document, silence, or the agent having enough
information cannot close the checkpoint. No further command is required.

A request for more research authorizes that bounded work and returns to research
collaboration. Scope changes return to analysis. Reuse a completed checkpoint
for unchanged evidence when explicit continuation is already recorded. Material
revisions reopen the affected checkpoint.

## Documents and Decisions

Use one evolving discovery brief for analysis, research, and human input. Keep
direction and plan as separate stable artifacts. Before inviting a document-based
response, validate and register the exact revision in mdmaid.desk under the
`readable-output` receipt and safe-link contract. Discovery discussion uses
`--attention review` without a decision request. Failed delivery blocks that
transition and preserves the local artifact for recovery.

Direction and plan collaboration includes the existing explicit choice to run
AI review, skip it, or review later. Selecting or skipping review can also express
continuation. Exact-revision human approval is still required through the existing
authenticated direction or plan decision. A changed artifact cannot inherit
approval from an earlier revision.

Replies use the active harness conversation today. Document chat in mdmaid.desk
is a future integration. Maisternia installs the phase contracts and graph;
it does not own live conversation, rooms, task execution, or a question queue.
Collaboration has artifact-only write authority and cannot modify target-project
files or perform unrelated external writes.
