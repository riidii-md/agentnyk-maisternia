# /codex-pr-check - Check PR Readiness with Codex

Use Codex to inspect commit, push, and pull request readiness for the current
repository. This command does not commit, push, or open a PR.

## Usage

`/codex-pr-check <optional PR, ticket, branch, or notes>`

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-pr-check.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Check the current repository branch for pull request readiness.

Context:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input.

Discover conventions from repository instructions, contribution docs, workflow
docs, CI, hooks, command and skill docs, and package or build files.

Do not assume:
- a ticket format
- Conventional Commits
- that PR title rules match commit rules
- GitHub, Gitea, GitLab, or another provider
- availability of gh, tea, glab, or an MCP

Verify:
- staged, unstaged, and untracked files
- accidental secrets and local-only configuration
- commit messages against discovered rules
- branch tracking and push status
- PR title and body requirements
- likely provider command
- claimed tests and documentation against actual evidence
- required checks inferred from CI, hooks, and repository docs

Return a concise pass/fail checklist with evidence and exact fixes.

Do not commit, push, open, or update a PR.
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

Present the failed checks, exact fixes, and approvals still required.
