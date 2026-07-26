# /codex-brief - Run the Brief Phase with Codex

This is a permanent explicit-runner alias for `/work brief`.

Claude must invoke Codex in read-only mode, pass the current conversation
handoff, `$ARGUMENTS`, repository context, and the canonical brief output shape,
then return Codex's concise result.

Use the configured Codex review model and profile. Do not replace the requested
Codex run with a Claude-only summary unless the user explicitly asks.
