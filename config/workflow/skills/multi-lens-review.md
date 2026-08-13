---
name: lens-review
description: Use for plan, design, decision-delta, diff, implementation, or delegated review that needs independent lenses, evidence-grounded findings, adversarial verification, and applied fixes.
---

# Lens Review

Resolve the target as `plan`, `plan-delta`, or `implementation`. An explicit
target wins. Do not refuse a plan review because no code diff exists. Review
plans against the actual repository, not against prose alone.

Use the installed `work-routing` skill for every cross-provider selection. A
route such as `/work-review @agy @codex @claude -- <target>` selects independent
read-only reviewer pools and defaults to `parallel-verify`; it does not grant
write authority. Keep native subagent selection local when no cross-provider
route was resolved.

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
