# /codex-research - Run the Research Phase with Codex

This is a permanent explicit-runner alias for `/work research`.

Claude must invoke Codex in read-only mode using the configured reasoning model.
Pass a self-contained handoff with facts, accepted analysis, constraints,
unknowns, rejected directions, `$ARGUMENTS`, and relevant repository paths.

Codex must follow the canonical research phase. Claude may synthesize afterward
but must preserve disagreements and uncertainty.
