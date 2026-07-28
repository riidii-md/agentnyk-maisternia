# /codex-cleanup - Review Codex Temporary Files

List temporary prompts, captured Codex outputs, deep-research inputs, and
readable Markdown or HTML created by the personal workflow.

## Usage

`/codex-cleanup`

Resolve the cleanup helper portably:

```bash
CODEX_CLEANUP_TEMP="${CODEX_CLEANUP_TEMP:-${CODEX_HOME:-$HOME/.codex}/bin/codex-cleanup-temp}"
```

List candidates first:

```bash
"$CODEX_CLEANUP_TEMP" --list
```

Show the complete candidate list and ask for explicit approval. Delete only
after approval:

```bash
"$CODEX_CLEANUP_TEMP" --delete
```

If the helper is unavailable, report its expected path and do not improvise a
delete command.

Never delete automatically. Never include repository files, credentials,
environment files, durable task state, decisions, plans, contracts, review
records, or files outside the helper's declared temporary patterns.
