# /codex-review - Use Codex as an Independent Reviewer

Run Codex as a separate review pass on the current repository. This is
command-only orchestration and must not install hooks.

## Usage

`/codex-review <optional contract, diff, PR, or review focus>`

Build a self-contained handoff with the accepted contract, implementation
scope, verification evidence, known risks, and current conversation context.
Do not use persuasive builder reasoning as proof.

## Resolve Scope

```bash
git fetch origin --quiet 2>/dev/null || true
BASE_REF="${CODEX_BASE_REF:-$(git symbolic-ref --quiet refs/remotes/origin/HEAD | sed 's#^refs/remotes/##')}"
[ -n "$BASE_REF" ] || BASE_REF="origin/main"
git diff --name-only "$BASE_REF"...HEAD 2>/dev/null || git diff --name-only
```

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-review.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Review the current repository branch with fresh context.

Contract and focus:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

First discover repository review expectations from files that exist, including
instructions, contribution docs, commands, skills, workflows, CI, hooks, and
package or build files.

Do not assume a ticket format, base branch, PR provider, test command, or MCP.

First verify contract compliance and observable behavior. Then find actionable:
- bugs and behavioral regressions
- security, authorization, secrets, and data exposure risks
- missing tests for changed behavior
- migration, deployment, and CI risks
- repository convention failures

Lead with findings ordered by severity. Include file paths, evidence, impact,
and a concrete fix. If there are no findings, say so and identify residual
test risk.
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

Treat the result as review input, not automatic truth. Verify critical and high
findings against code and tests before changing anything.
