# /codex-review - Run Independent Review with Codex

This is a permanent explicit-runner alias for `/work review`.

Claude must invoke Codex in read-only review mode using the configured review
model. Pass the accepted contract, code and diff scope, verification evidence,
repository rules, and `$ARGUMENTS`. Do not pass persuasive builder reasoning as
proof.

Claude verifies critical and high findings against the code before acting.
