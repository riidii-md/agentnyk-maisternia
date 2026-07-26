# /codex-work-loop - Run the Implementation Phase with Codex

This is a permanent explicit-runner alias for `/work run`.

Require an approved handoff and understood working tree. Claude must invoke
Codex in an isolated writable workspace using the configured coding model and
pass the self-contained contract plus `$ARGUMENTS`.

Codex follows the canonical run phase. Commit, push, PR, destructive actions,
production data, and permission escalation remain explicit user checkpoints.
Claude verifies the resulting diff and checks before reporting completion.
