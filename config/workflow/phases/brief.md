# /work-brief - Quick Task Refresher

Give a concise reminder of what the active task is about and where it stands.

Input:

`$ARGUMENTS`

Use the current conversation, durable task state, event history, current
repository, branch, Git status, recent relevant commits, and explicitly
referenced artifacts. Use read-only checks only.

Return:

```text
Task: <ticket or short name>
What this is about: <2-4 simple sentences>
Current status: <phase and why>
What happened so far:
- <3-6 short events>
Important files and places:
- <only the most relevant items>
Next step: <one action>
Open questions: <none or 1-3 questions>
```

Do not edit files, change external state, expose secrets, or dump raw logs.
