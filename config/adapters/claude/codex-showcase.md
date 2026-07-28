# /codex-showcase - Present Findings and Plans with Codex

Create a Codex-backed showcase after planning, research, analysis,
implementation, review, or a long conversation.

## Usage

`/codex-showcase <plan, findings, conversation, file paths, or brief>`

When explicitly invoked, Claude must run Codex unless the user asks for a
Claude-only document.

Build a self-contained handoff containing:

- user goal;
- current phase and status;
- established facts and assumptions;
- relevant history;
- accepted decisions and rejected directions;
- plan or proposal;
- files, tickets, URLs, logs, and reports;
- risks and open questions;
- what the document should help the user decide.

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-showcase.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Create a standalone Markdown showcase from the supplied context.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input. The reader has not seen the original conversation.

Goal:
Help the reader understand what the work is about, where it stands, what was
learned or changed, what needs approval, and what should happen next.

Do:
- read referenced local files and reports when provided
- separate verified facts from assumptions and proposals
- explain technical findings in simple terms without losing constraints
- include architecture or workflow diagrams when useful
- use Mermaid when it materially improves understanding
- call out risks, unknowns, and review checkpoints
- include source paths and URLs

Do not:
- edit repository files
- commit, push, open a PR, or change external state
- expose secrets, tokens, private environment values, or sensitive logs
- dump raw logs unless short and necessary

Return complete Markdown with these sections when relevant:
- Title
- Executive Summary
- Current Status
- Problem or Goal
- What Happened So Far
- What We Found
- Decisions Already Made
- Proposed Direction or Plan
- Architecture or Workflow
- Risks and Unknowns
- Review Questions or Approval Needed
- Next Steps
- Sources and Relevant Files
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

CODEX_READABLE_DOC="${CODEX_READABLE_DOC:-${CODEX_HOME:-$HOME/.codex}/bin/codex-readable-doc}"
if [ -x "$CODEX_READABLE_DOC" ]; then
  "$CODEX_READABLE_DOC" --open codex-showcase < "$OUTPUT"
else
  printf 'Readable-output helper unavailable; Markdown remains at %s\n' "$OUTPUT"
fi
```

Reply with the phase and status, Markdown and rendered paths when available,
and the next user decision or command.
