# /codex-showcase - Run the Showcase Phase with Codex

This is a permanent explicit-runner alias for `/work showcase`.

Claude must invoke Codex in read-only mode with a self-contained handoff and
`$ARGUMENTS`. Codex follows the canonical showcase phase and returns complete
Markdown. Claude renders it with the configured readable-output tool and opens
the result when requested.
