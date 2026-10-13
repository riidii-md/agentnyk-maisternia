# Task Context

Load this reference when saving or resuming task progress. It defines a compact
local summary and its safeguards. It does not grant workspace-write authority,
change approval gates, or require automatic publication in mdmaid.desk. Direct
read-only phases return evidence to a writer with existing artifact authority.

## Persistent task summary

The coordinator maintains one task-local Markdown file so work can continue in
another session. Use an explicit progress artifact path when supplied. Otherwise
use `.agent-runs/work/<task-id>/progress.md` in the dedicated task workspace.
Choose a stable task ID from the ticket or task branch. Use only letters,
digits, hyphens, and underscores in the directory name. Verify the file's task
identity before reuse and confirm that it belongs to the same task. If the
identity differs or is ambiguous, preserve the existing file and select a distinct
task path; ask only if task selection is
unresolved. Report the chosen path so the human can supply it in a new session.

After Git preflight, create or refresh the file from observed evidence. Update
it after each completed phase, after each completed implementation step, before
pausing, and before context compaction when the harness gives advance notice.
Keep the latest complete summary rather than an appended transcript. A sudden
interruption may leave it behind current work; recovery must inspect the actual
repository. The coordinator writes summaries of read-only phase results outside
those phases. `/work-brief` and `/work-scout` retain read-only authority.

Keep this compact field set; use links instead of copying plans or source bodies:

```text
Task identity: task ID, repository, base, branch, worktree, updated UTC time
Goal and scope: requested outcome and exclusions
Current phase: status and why
Project document map: relevant sources, sections, status, and unresolved conflicts
Decision references: artifact paths, revisions, content hashes, and decision receipts when available
Constraints: safety rules and permitted executor discretion
Completed work: observed results and changed files
Remaining work: unfinished steps and blockers
Verification evidence: commands, results, evaluated revision or working snapshot, and safe log paths
Failed approaches: concise causes worth avoiding
Open questions: unresolved facts and human decisions
Next action: one concrete step or the exact response needed
```

Record unavailable evidence as `unknown`. A summary is not approval or a new
execution contract. Preserve references to the approved plan, any required
direction, and exact-revision decisions. Do not invent receipts or reproduce
credentials, tokens, personal data, private configuration, raw logs, transcripts,
or sensitive source bodies. Screen and redact before writing. Keep the file out
of commits by default; use the repository's local ignore mechanism within the
available write authority. Do not commit it or publish it automatically.
Maisternia project-scoped configuration apply installs a project ignore rule
for `.agent-runs/`. User-scoped installations do not cover project ignore rules;
check the selected repository before saving. Git ignore rules do not untrack
previously committed files.

Validate the destination within the permitted task workspace. Reject path
traversal, symlink destinations or ancestors, and unrelated task files. Use a
same-directory temporary file and atomic replacement for updates. Recheck the
existing file before replacing it and stop on concurrent changes; atomic writing
alone does not prevent lost updates. One coordinator owns the file. Workers
return evidence to that coordinator instead of writing competing summaries.
If writing fails, preserve the previous file and report that saved progress is
outdated; do not claim a successful checkpoint. This file is a harness-authored
task artifact, not a Maisternia runtime database or approval queue.

On resume, read the selected summary and revalidate task identity, worktree,
branch, current files, repository instructions, relevant documents, verification
evidence, and referenced artifacts and decisions. Follow the existing Git
preflight and readiness rules. If evidence is missing or stale, inspect the
affected work or return to the unmet gate. Do not infer approval or completed
work from the summary. Do not replay recorded commands or external writes
merely because they appear in the file. Continue from the first unmet valid
step, preserve unrelated edits, and refresh the summary from verified evidence.
