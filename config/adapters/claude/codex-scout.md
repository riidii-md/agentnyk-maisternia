# /codex-scout - Gather Facts with Codex

Collect source material before analysis or planning by invoking Codex from
Claude.

## Usage

`/codex-scout <ticket, URLs, logs, files, screenshots, or brief>`

Treat the current conversation as input. Build a self-contained handoff with:

- user goal or question;
- facts already established;
- evidence, logs, screenshots, ticket text, URLs, and file paths;
- labeled inferences and assumptions;
- unknowns and questions;
- the phase requested now.

Do not replace an explicitly requested Codex run with Claude-only scouting
unless the user asks.

## Local Context

```bash
git status --short 2>/dev/null || true
git branch --show-current 2>/dev/null || true
pwd
```

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-scout.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Scout the task before analysis or planning.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Goal:
Gather facts and identify missing evidence. Do not solve or implement yet.

Do:
- read repository instructions, contribution docs, command and skill docs,
  CI, hooks, package and build files, and workflow documentation
- inspect user-provided files, logs, errors, URLs, and ticket text
- use web research only for supplied URLs or relevant public documentation
- separate facts from inferences
- identify likely boundaries, affected subsystems, ambiguities, and missing data
- infer repository-specific rules only from evidence

Do not:
- edit files
- commit, push, or open a PR
- assume ticket format, base branch, MCPs, provider, or test commands
- jump to implementation

Return:
- handoff summary used
- facts with source paths or URLs
- unknowns and questions
- likely scope boundaries
- likely affected files or subsystems
- evidence still needed
- recommended next phase
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

Summarize the report and ask only blocking questions before analysis.
