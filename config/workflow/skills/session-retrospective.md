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
3. `/work-improve` prepares and replays a minimal candidate only when evidence
   supports a change.
4. `/work-retrospective` coordinates the complete post-run package.

Keep measured values separate from estimates and unknowns. Prefer deterministic
repository evidence over model judgment. Treat transcripts as observable traces,
not hidden reasoning. Redact sensitive input before delegating review.

Repeated findings may justify a proposal. They do not authorize installation.
Never edit durable configuration, activate hooks, publish transcripts, or apply
an `agentctl` preset without explicit human approval.
