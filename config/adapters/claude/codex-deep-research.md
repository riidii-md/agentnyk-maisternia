# /codex-deep-research - Run Independent Research Lanes

Run deeper solution research with Claude as orchestrator and available external
CLI agents as independent reviewers.

## Usage

`/codex-deep-research <analysis, constraints, evidence, and question>`

Use this for consequential, ambiguous, cross-system, or high-risk choices.
External agents do not share Claude's context, so every lane needs the same
self-contained handoff.

## Workflow

1. Claude records an independent position before reading other lanes.
2. Create one shared prompt with the task, facts, assumptions, constraints,
   rejected options, repository path, and desired decision.
3. Run available lanes independently, in parallel when practical.
4. Preserve disagreements and verify factual conflicts.
5. Synthesize a recommendation and the user decisions still needed.

## Codex Lane

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-deep-research.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Act as one independent lane in a multi-agent technical research process.

Context:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Inspect relevant repository files and current primary documentation.

Return:
- viable solution options
- tradeoffs in scope, correctness, security, testability, migration,
  deployment, user impact, reversibility, and complexity
- assumptions and unknowns
- recommended direction
- rejected options and why
- questions required before planning

Do not edit files, commit, push, open a PR, or change external state.
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

## Optional Additional Lanes

When Antigravity (`agy`) is installed, run an equivalent non-mutating research
prompt:

```bash
agy -p "$(cat "$TASK")" --mode plan --add-dir "$PWD" --print-timeout 10m
```

If a lane is unavailable or unauthenticated, record that and continue. Never
grant write or full-access authority for research.

## Synthesis

Return points of agreement, disagreements, facts requiring verification,
recommended direction, risks, decisions required from the user, and whether
the task is ready for a decision record.
