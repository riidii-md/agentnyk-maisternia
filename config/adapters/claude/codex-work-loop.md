# /codex-work-loop - Execute an Approved Plan with Codex

Use Codex to carry out an already reviewed and approved implementation plan.
This is command-only orchestration and must not install hooks.

## Usage

`/codex-work-loop <task plus plan path or self-contained plan>`

## Preconditions

- The plan and acceptance contract exist and were approved.
- The working tree state is understood.
- Repository requirements come from repository evidence.
- Permission escalation, commits, pushes, PRs, destructive actions, and
  production access remain explicit user checkpoints.

## Capture Scope

```bash
git status --short
git branch --show-current
BASE_REF="${CODEX_BASE_REF:-$(git symbolic-ref --quiet refs/remotes/origin/HEAD | sed 's#^refs/remotes/##')}"
[ -n "$BASE_REF" ] || BASE_REF="origin/main"
git diff --stat "$BASE_REF"...HEAD 2>/dev/null || git diff --stat
```

## Run Codex

```bash
CODEX_FAST_MODEL="${CODEX_FAST_MODEL:-gpt-5.6-luna}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-work-loop.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Execute the approved plan in the current repository.

Task, plan, and acceptance contract:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Work loop:
1. Discover repository rules from instructions, command and skill docs, CI,
   hooks, package files, build files, and workflow docs.
2. Reconfirm the approved plan against current code and working tree state.
3. Select the first unfinished, unblocked task.
4. Implement only that task in a small restartable step.
5. Add or update focused tests where the repository or behavior requires them.
6. Run the task criteria and the smallest relevant deterministic checks.
7. Record files changed, checks, result, blockers, and next action.
8. Repeat until complete or blocked.

Use repository-required impact analysis only when the repository requires it
and the tool exists. Do not assume a specific MCP or code intelligence tool.

Stop after three consecutive failed verification attempts on the same issue
and report the blocker. Do not weaken acceptance criteria.

Do not commit, push, open a PR, access production, perform destructive actions,
or broaden scope unless the explicit handoff authorizes it.

Return:
- tasks completed and remaining
- files changed
- repository rules applied
- tests and commands run with results
- checks still needed
- risks or blockers
- exact next action
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_FAST_MODEL" \
  -c 'model_reasoning_effort="low"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox workspace-write \
  - < "$TASK"
printf 'Full Codex final message: %s\n' "$OUTPUT"
```

## Verify the Result

```bash
git status --short
git diff --stat
```

Run the checks required by the changed area and repository rules. If Codex
stopped on a blocker, include the blocker and current diff in the next handoff.
Stop and ask the user if the implementation expands beyond the approved plan.
