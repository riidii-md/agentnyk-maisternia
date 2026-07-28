# /codex-decision - Record the Chosen Direction with Codex

Create a compact decision record after research and before detailed planning.

## Usage

`/codex-decision <analysis, research, user choice, and constraints>`

Claude must ask Codex to draft the decision unless the user requests otherwise.
An unconfirmed recommendation is not an approved decision.

## Run Codex

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-decision.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Create a decision record from the supplied context.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Return:
- decision title
- chosen approach
- why this approach was chosen
- rejected options and why
- accepted user constraints
- assumptions
- accepted risks
- open questions
- ready-for-planning status: yes or no

Do not edit repository files or reinterpret an unconfirmed recommendation as
user approval.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_PLAN_MODEL" \
  -c 'model_reasoning_effort="high"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex final message: %s\n' "$OUTPUT"
```

If the result is not ready for planning, discuss the blockers or return to
research. Store a personal decision artifact only when useful; do not commit it
unless the user requests repository persistence.
