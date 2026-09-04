---
name: session-retrospective
description: Use after a completed agent task to profile the active harness, audit the run with evidence, and prepare controlled improvement proposals without automatically changing durable configuration.
---

# Session Retrospective

Use this skill when a completed run should become evidence for improving the
agent harness. It applies to correctness, token and latency cost, commands,
prompts, skills, MCP servers, hooks, settings, plugins, and model routing.

Follow the installed retrospective policy and compose these procedures:

1. `/work-profile` establishes the current configuration baseline.
2. `/work-audit` evaluates one explicitly selected run.
3. `/work-session-analysis` delegates focused token, repetition, skills, user
   friction, setup, command, and subagent reviews.
4. `/work-improve` prepares and replays a minimal candidate only when evidence
   supports a change.
5. `/work-retrospective` coordinates the complete post-run package.

Use the installed `work-routing` skill whenever a retrospective lane crosses a
provider boundary. Keep same-harness subagent scheduling local, but reuse the
shared router for harness selection, disclosure, authority, unavailable-target
handling, and the routing receipt.

Keep measured values separate from estimates and unknowns. Prefer deterministic
repository evidence over model judgment. Treat transcripts as observable traces,
not hidden reasoning. Redact sensitive input before delegating review. Routed
reviewers remain read-only and receive only the selected run evidence.

## Centralize Completed Packages

After a retrospective-producing command completes its local package, copy its
curated artifacts into the central store defined by the installed retrospective
policy. Resolve the store to
`${XDG_STATE_HOME:-$HOME/.local/state}/maisternia/findings` when no XDG state
override is set. Use `runs/<provider>/<run-id>/` and create `source.json` from
the installed retrospective source schema.

Accept only an explicitly selected local retrospective directory. Resolve and
inspect the source and destination before writing. Reject symlinks, non-regular
files, path traversal, oversized files, unsafe provider or run identifiers, and
artifacts outside the policy allowlist. Never copy transcripts, provider
histories, event streams, runtime databases, credentials, or tokens. Use private
directory and file modes, stage a complete import beside the destination, and
rename it into place only after validation and SHA-256 calculation succeeds.

Classify a current-schema, validated `record.json` as `valid`. Preserve an older
or structurally different record byte-for-byte and classify it as `legacy`; do
not manufacture fields. An identical import is a no-op. A same-source refresh
may replace the current curated package atomically after retaining its previous
version under `history/`. A matching run ID from a different source is a
collision: abort and request a human decision. Aggregation counts only the
current top-level package for each run.

Repeated findings may justify a proposal. They do not authorize installation.
Never edit durable configuration, activate hooks, publish transcripts, or apply
an `maisternia` preset without explicit human approval.
