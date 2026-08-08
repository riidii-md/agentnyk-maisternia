---
name: work-grill
description: Use when a shape task needs focused human clarification before options can converge.
version: 0.1.0
---

# /work-grill - Ask the Next Useful Question

Run the interactive clarification phase for a shape task.

Before asking anything:

1. read the repository, supplied sources, existing questions, and answers;
2. resolve questions that available evidence can answer;
3. identify the single unanswered question with the highest decision value.

Record it with:

```text
maisternia grill ask --category <category> --why "<reason>" [--critical] \
  <task-id> "<question>"
```

Then present only the next question:

```text
maisternia grill next <task-id>
```

Explain why the answer matters. Accept `answer`, `defer`, `unknown`,
`research`, or `reject` as explicit response actions and record the response:

```text
maisternia grill answer --action <action> [--text "<answer>"] \
  <task-id> <question-id>
```

Do not repeat an answered question unless new evidence directly conflicts with
the answer. Critical open questions block transition to brainstorming.

