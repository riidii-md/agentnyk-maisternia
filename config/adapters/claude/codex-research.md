# /codex-research - Explore Solution Options with Codex

Research possible solutions for an accepted task definition before detailed
planning.

## Usage

`/codex-research <analysis, evidence, constraints, and research question>`

Treat the current conversation as input. Build a self-contained handoff with:

- the accepted problem or feature definition;
- evidence and source locations;
- user constraints and preferences;
- facts, assumptions, and unknowns;
- rejected directions;
- open decisions;
- the output the user needs from research.

Do not replace an explicitly requested Codex run with Claude-only research
unless the user asks.

## Run Codex

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-research.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Research solution options for the current repository.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Goal:
Find and compare viable solution directions. Do not create a detailed
implementation plan yet.

Do:
- inspect repository code and documentation for existing patterns
- use current primary external documentation when library or API behavior matters
- propose 2-4 options when there is meaningful choice
- compare scope, risk, testability, security, migration, deployment, user impact, and complexity
- recommend one direction and explain why
- identify decisions required from the user
- infer verification gates from repository evidence

Do not:
- edit files
- commit, push, or open a PR
- assume conventions not supported by repository evidence
- expand scope without calling it out

Return:
- handoff summary used
- options considered
- recommendation
- rejected options and why
- user decisions needed
- rough implementation shape
- verification implications
- readiness for planning
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

Discuss the recommendation with the user. Preserve uncertainty and dissent. If
the direction is accepted, continue through decision and readiness before
planning.
