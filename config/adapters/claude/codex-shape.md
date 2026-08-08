# /codex-shape - Use Codex to Shape an Idea

Run the canonical `/work-shape` workflow through Codex while preserving the
same durable maisternia task state and read-only target-project authority.

## Input

`$ARGUMENTS`

Build a self-contained handoff containing the current shape task ID, phase,
source ledger summary, open grill questions, existing artifacts, repository
rules, and the user's request.

Run Codex in the current repository with a read-only sandbox. Ask it to perform
only the current shape phase, use maisternia for durable source, question, and
transition state, and return the generated Markdown artifact or next human
question.

Do not let the runner modify target project files, commit, push, open a PR,
submit forms, or finalize a revision without explicit human confirmation.

