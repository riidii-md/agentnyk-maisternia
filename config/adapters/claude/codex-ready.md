# /codex-ready - Run the Readiness Gate with Codex

Check whether the task has enough clarity to move to planning or
implementation.

## Usage

`/codex-ready <scout, analysis, research, decision, and approval context>`

Claude must run the Codex gate unless the user asks for a Claude-only check.

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-ready.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Evaluate readiness for planning or implementation.

Context:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Return a checklist:
- problem or feature statement exists
- facts and assumptions are separated
- scope and exclusions are clear
- acceptance criteria exist
- solution direction is decided
- repository rules are identified or explicitly unknown
- key unknowns are resolved or accepted as risks
- required user approval exists
- result: pass, conditional pass, or fail
- exact missing inputs
- recommended next phase

Do not edit files. Do not pass unresolved critical ambiguity.
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

Do not proceed past unresolved critical or high ambiguity unless the user
explicitly accepts the risk.
