# /codex-analyze - Define the Task with Codex

Use Codex to turn gathered scout context into a crisp bug, feature, refactor,
investigation, or operations definition.

## Usage

`/codex-analyze <scout report and task context>`

Treat the current conversation as input. Build a self-contained handoff with
the user goal, scout facts, evidence, labeled assumptions, unknowns, and the
analysis output needed now.

## Run Codex

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-analyze.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Analyze the gathered scout context.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Goal:
Produce a clear task definition, not an implementation plan.

Do:
- separate facts, assumptions, inferences, and unknowns
- classify the task as bug, feature, refactor, investigation, operations, or mixed
- for bugs, define observed behavior, expected behavior, reproduction evidence,
  impact, and suspected root-cause areas
- for features, define the user goal, current and desired behavior, constraints,
  scope exclusions, and draft acceptance criteria
- identify repository rules discovered from evidence
- identify decisions still requiring user confirmation

Do not:
- edit files
- choose a final solution prematurely
- assume repository conventions without evidence

Return:
- handoff summary used
- concise problem or feature statement
- evidence summary
- scope and exclusions
- acceptance criteria draft
- constraints and risks
- open questions
- recommended next phase
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

Discuss the result with the user and resolve material ambiguity before
research or planning.
