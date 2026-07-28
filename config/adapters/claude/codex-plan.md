# /codex-plan - Use Codex for Planning

Prepare a Codex-backed implementation plan for the current repository. This is
command-only orchestration and must not install hooks or edit files.

## Usage

`/codex-plan <task, ticket, accepted decision, or plan request>`

Treat the current conversation as input. Build a self-contained handoff with the
accepted task definition, constraints, decisions, proof requirements, relevant
files, and unresolved questions.

## Capture Context

```bash
git status --short
git branch --show-current
BASE_REF="${CODEX_BASE_REF:-$(git symbolic-ref --quiet refs/remotes/origin/HEAD | sed 's#^refs/remotes/##')}"
[ -n "$BASE_REF" ] || BASE_REF="origin/main"
git diff --stat "$BASE_REF"...HEAD 2>/dev/null || git diff --stat
```

## Run Codex

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-plan.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Create an implementation plan for the current repository.

Task and accepted context:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

First discover repository rules from files that exist, including:
- AGENTS.md, CLAUDE.md, CONTRIBUTING.md
- command and skill documentation
- package and build files
- CI workflows and hooks
- development, workflow, and operations documentation

Do not assume:
- a particular base branch or ticket format
- a specific code intelligence tool or MCP
- a Git hosting provider
- fixed test, commit, or PR conventions

Do not edit files.

Return:
- handoff summary used
- discovered repository rules and assumptions
- scope and exclusions
- existing files and patterns to inspect first
- ordered implementation steps
- risk and blast-radius checks
- migration or deployment concerns
- test and verification commands inferred from repository evidence
- commit, PR, and ticket conventions inferred from repository evidence
- stop conditions and unresolved user decisions
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

Summarize the plan for the user and call out unresolved decisions or conflicts
with repository rules.
