# /codex-fleet - Run Planning, Review, and PR Lanes

Run a small read-only Codex fleet from Claude. This command does not install
hooks or modify the repository.

## Usage

`/codex-fleet <task, ticket, plan, or review focus>`

Use the configured Codex profile only when the repository requires one. The
plan lane uses the reasoning model; review and PR lanes use the review model.
Run lanes in parallel when supported and keep outputs separate.

## Plan Lane

```bash
CODEX_PLAN_MODEL="${CODEX_PLAN_MODEL:-gpt-5.6-sol}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-fleet-plan.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Plan lane for the current repository.

Task:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Discover repository rules. Return scope, assumptions, ordered steps, risks,
and inferred verification. Do not edit files or assume base branch, ticket
format, provider, MCPs, or fixed test commands.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_PLAN_MODEL" \
  -c 'model_reasoning_effort="high"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex plan message: %s\n' "$OUTPUT"
```

## Review Lane

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-fleet-review.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Review lane for the current repository.

Contract and focus:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Review the current branch and diff using discovered repository rules. Report
actionable bugs, security issues, contract failures, missing tests, migration
risks, and convention failures. Do not edit files.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_REVIEW_MODEL" \
  -c 'model_reasoning_effort="medium"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex review message: %s\n' "$OUTPUT"
```

## PR Lane

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-fleet-pr.XXXX.md)"
cat > "$TASK" <<'PROMPT'
PR readiness lane for the current repository.

Context:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Discover commit, ticket, PR title, branch, push, provider, and required-check
rules. Return a pass/fail checklist and exact fixes. Do not edit, commit, push,
or open a PR.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_REVIEW_MODEL" \
  -c 'model_reasoning_effort="medium"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex PR message: %s\n' "$OUTPUT"
```

After all lanes finish, merge duplicates, preserve disagreements, verify
critical and high findings locally, and present one concise synthesis.
