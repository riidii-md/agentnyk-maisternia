# /codex-brief - Ask Codex for a Quick Task Refresher

Create a short Codex-backed reminder of what this task or session is about,
where it stands, and the important history in simple terms.

## Usage

`/codex-brief <optional ticket, notes, file paths, branch, PR, or question>`

Treat the current conversation as input. Before invoking Codex, create a short
self-contained handoff containing:

- user goal;
- likely task or ticket identifier;
- current phase;
- established facts;
- changes attempted or completed;
- relevant files, plans, PRs, URLs, and commands;
- blockers and open questions;
- what the user probably needs to remember.

Do not replace an explicitly requested Codex run with a Claude-only summary
unless the user asks.

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-brief.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Create a brief refresher for the user from the supplied context.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Goal:
Help the user remember what this task is about, where it stands, and the
important history in simple terms.

Do:
- inspect the current directory, branch, git status, and recent commits if useful
- infer a task identifier from branch, path, arguments, or conversation
- read explicitly referenced local artifacts when useful
- separate facts from assumptions
- keep the response short and practical

Do not:
- edit files
- commit, push, open a PR, or change external state
- dump raw logs or long command output
- expose secrets or private environment values

Return exactly:

Ticket: <ticket, inferred short name, or "unknown">
What this is about: <2-4 simple sentences>
Current status: <phase and why>
What happened so far:
- <3-6 short events>
Important files/places:
- <only the most relevant paths, PRs, plans, or commands>
Next step: <one concrete action>
Open questions: <none, or 1-3 questions>

Keep the answer readable in under one minute.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_REVIEW_MODEL" \
  -c 'model_reasoning_effort="medium"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex final message: %s\n' "$OUTPUT"
```

Reply with the brief and no additional long explanation.
