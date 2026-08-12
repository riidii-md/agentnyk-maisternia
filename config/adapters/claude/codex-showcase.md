# /codex-showcase - Present Findings and Plans with Codex

Create a Codex-backed showcase after planning, research, analysis,
implementation, review, or a long conversation.

## Usage

`/codex-showcase <plan, findings, conversation, file paths, or brief>`

When explicitly invoked, Claude must run Codex unless the user asks for a
Claude-only document.

Build a self-contained handoff containing:

- user goal;
- current phase and status;
- established facts and assumptions;
- relevant history;
- accepted decisions and rejected directions;
- plan or proposal;
- files, tickets, URLs, logs, and reports;
- risks and open questions;
- what the document should help the user decide.

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-showcase.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Create a standalone Markdown showcase from the supplied context.

Input:
$ARGUMENTS

Conversation handoff:
<CONVERSATION_HANDOFF>

Treat the conversation handoff as authoritative context and merge it with the
explicit input. The reader has not seen the original conversation.

Goal:
Help the reader understand what the work is about, where it stands, what was
learned or changed, what needs approval, and what should happen next.

Do:
- read referenced local files and reports when provided
- separate verified facts from assumptions and proposals
- explain technical findings in simple terms without losing constraints
- include architecture or workflow diagrams when useful
- use Mermaid when it materially improves understanding
- call out risks, unknowns, and review checkpoints
- include source paths and URLs

Do not:
- edit repository files
- commit, push, open a PR, or change external state
- expose secrets, tokens, private environment values, or sensitive logs
- dump raw logs unless short and necessary

Return complete Markdown with these sections when relevant:
- Title
- Executive Summary
- Current Status
- Problem or Goal
- What Happened So Far
- What We Found
- Decisions Already Made
- Proposed Direction or Plan
- Architecture or Workflow
- Risks and Unknowns
- Review Questions or Approval Needed
- Next Steps
- Sources and Relevant Files
PROMPT
WORKSPACE_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd -P)"
WORKSPACE_ROOT="$(cd "$WORKSPACE_ROOT" && pwd -P)"
ARTIFACT_DIR="$WORKSPACE_ROOT/.agent-runs/showcase"
mkdir -p "$ARTIFACT_DIR"
OUTPUT="$ARTIFACT_DIR/$(date -u +%Y%m%d-%H%M%S)-codex-showcase.md"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_REVIEW_MODEL" \
  -c 'model_reasoning_effort="medium"' \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex final message: %s\n' "$OUTPUT"

if command -v mdmaid-desk >/dev/null 2>&1; then
  DESK_WORKSPACE="${MDMAID_DESK_WORKSPACE:-}"
  if [ -z "$DESK_WORKSPACE" ]; then
    while IFS=$'\t' read -r candidate_id candidate_name candidate_root; do
      if [ "$candidate_root" = "$WORKSPACE_ROOT" ]; then
        DESK_WORKSPACE="$candidate_id"
        break
      fi
    done < <(mdmaid-desk workspace list)
  fi
  if [ -z "$DESK_WORKSPACE" ]; then
    workspace_slug="$(basename "$WORKSPACE_ROOT" | tr '[:upper:]_' '[:lower:]-' | tr -cd 'a-z0-9-' | cut -c1-48)"
    case "$workspace_slug" in
      ""|[0-9]*) workspace_slug="workspace-$workspace_slug" ;;
    esac
    workspace_hash="$(printf '%s' "$WORKSPACE_ROOT" | cksum | awk '{print $1}')"
    DESK_WORKSPACE="$workspace_slug-$workspace_hash"
    mdmaid-desk workspace add "$WORKSPACE_ROOT" \
      --id "$DESK_WORKSPACE" \
      --name "$(basename "$WORKSPACE_ROOT")"
  fi
  mdmaid-desk register "$OUTPUT" \
    --workspace "$DESK_WORKSPACE" \
    --kind showcase \
    --attention review
else
  printf 'mdmaid-desk unavailable; Markdown remains at %s\n' "$OUTPUT"
fi
```

Reply with the phase and status, Markdown path, mdmaid.desk registration status,
and the next user decision or command. Do not duplicate the complete Markdown.
