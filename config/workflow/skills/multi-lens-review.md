---
name: lens-review
description: Use for plan, design, decision-delta, diff, implementation, or delegated review that needs independent lenses, evidence-grounded findings, behavior-preserving maintainability review, adversarial verification, and applied fixes.
version: 0.3.0
---

# Lens Review

Resolve the target as `plan`, `plan-delta`, or `implementation`. An explicit
target wins. Do not refuse a plan review because no code diff exists. Review
plans against the actual repository, not against prose alone.

Resolve the profile as `standard` unless the user explicitly requests
`maintainability`. The maintainability profile applies only to implementation
review. If the target is a plan, do not silently reinterpret either the target
or profile.

Use the installed `work-routing` skill for every cross-provider selection. A
route such as `/work-review @agy @codex @claude -- <target>` selects independent
read-only reviewer pools and defaults to `parallel-verify`; it does not grant
write authority. Keep native subagent selection local when no cross-provider
route was resolved.

For the `maintainability` profile, run all implementation lenses and add the
`best-practices` lens. First state the behavior contract that must remain
unchanged. Deepen `consistency`, `architecture`, `simplicity-dry`, and
`tests-verification` to find repeated knowledge, avoidable complexity, weak
ownership boundaries, and repository-practice violations. Distinguish repeated
knowledge from incidental duplication. Propose an abstraction only when it
reduces concepts or coupling; reject speculative helpers and indirection.

Discover languages, frameworks, build systems, and generated surfaces before
choosing practices or checks. Discovery must be language-agnostic,
evidence-led, and confidence-aware rather than a fixed language-to-tool table.
Use repository instructions, CI and hooks, manifests, lockfiles, source
markers, and changed paths; a file extension alone is insufficient. Record
`detected`, `mixed`, or `unknown`, the evidence and confidence, and cover each
affected surface in a multi-language repository. Prefer repository-owned
commands. Use only relevant, already-available supplementary tools, never
install or enable one without approval, and preserve uncertainty rather than
inventing a required gate.

Every maintainability candidate must include concrete evidence, the minimal
fix, the expected net simplification, regression risk, and a verification plan
for preserved behavior. Ground best-practices claims in repository rules,
established neighboring code, or authoritative documentation. Style
preferences alone are not findings, and `NO_FINDINGS` remains valid.

Run one read-only reviewer per required lens, in parallel when supported. Add
domain lenses only when the affected surface warrants them. Every candidate
finding needs concrete grounding such as `file:line`, a short verbatim quote,
a reproducible command or test, or an authoritative document.

With several routed harnesses, distribute lenses before dispatch and preserve
provider/model attribution. Prefer a verifier from a different selected harness
than the finding origin. A model committee is evidence diversity, not proof.

For every candidate, launch an independent verifier whose job is to refute it.
The verifier must read the relevant code or plan and return explicit `is_real`
and `grounded` booleans. Keep a finding only when both are true. Record what was
refuted and why, then deduplicate and rank confirmed findings by severity.

Reviewers and verifiers remain read-only. The current coordinating harness owns mutations and
verification. Apply every confirmed fix within the approved scope. Critical and
High findings are blocking until fixed or explicitly blocked by a user decision.
For plans, edit the plan or design artifact. For implementations, edit code and
tests, run focused checks, then run the repository-required final verification.

Write `review.md` and schema-valid `review.json` under
`.agent-runs/reviews/<run-id>/`. Report confirmed findings, applied changes,
refuted findings and rationale, checks, unresolved blockers, and gate status.
Record the selected `standard` or `maintainability` profile in the report.
