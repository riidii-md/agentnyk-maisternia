---
name: lens-review
description: Use for plan, design, decision-delta, diff, implementation, or delegated review that needs independent lenses, implementation repair or report-only disposition, evidence-grounded findings, and adversarial verification.
version: 0.5.0
---

# Lens Review

Resolve the target as `plan`, `plan-delta`, or `implementation`. An explicit
target wins. Do not refuse a plan review because no code diff exists. Review
plans against the actual repository, not against prose alone.

Resolve the profile as `standard` unless the user explicitly requests
`maintainability`. The maintainability profile applies only to implementation
review. If the target is a plan, do not silently reinterpret either the target
or profile.

Resolve the scope as `full` unless the user explicitly requests `tests` or
invokes `/work-test-review`. The `tests` scope applies only to implementation
review. Every full implementation review embeds the specialized test-review
bundle; the tests scope runs that bundle without unrelated implementation
lenses. Scope never widens authority.

For implementation review, resolve the disposition as `repair` unless the
caller explicitly requests `report-only`. Repair preserves the normal
coordinator-owned fix loop. Report-only is available only for implementation
review used as advisory or PR-publication feedback; reject it for plan,
plan-delta, or direction review instead of silently changing semantics. Every
report-only reviewer and the coordinator must not edit the reviewed
implementation, tests, or contributor branch. They may write only task-owned
review artifacts. The disposition changes mutation authority, not lenses,
evidence, severity, or independent verification.

For a PR, fork, patch, or other externally controlled implementation, treat
the reviewed code as untrusted executable content. Prefer provider CI evidence
and static inspection. Do not execute reviewed tests, builds, scripts, hooks,
package lifecycle steps, generators, or binaries on the host unless the user
explicitly authorizes that execution and it runs in a disposable sandbox with
no credentials or secrets, no host write access, no network by default,
isolated caches, and bounded resources. This boundary covers both dispositions
and every focused or final verification step.

Use the installed `work-routing` skill for every cross-provider selection. A
route such as `/work-review @agy @codex @claude -- <target>` selects independent
read-only reviewer pools and defaults to `parallel-verify`; it does not grant
write authority. Keep native subagent selection local when no cross-provider
route was resolved.

For the `maintainability` profile, run all implementation lenses and add the
`best-practices` lens. First state the behavior contract that must remain
unchanged. Deepen `correctness`, `consistency`, `architecture`,
`simplicity-dry`, and `test-review-bundle` to find repeated knowledge, avoidable
complexity, weak ownership boundaries, and repository-practice violations.
Distinguish repeated knowledge from incidental duplication. Propose an
abstraction only when it reduces concepts or coupling; reject speculative
helpers and indirection. Prove preserved behavior including failure paths.

Evaluate simplifications in order. Stop at the first behavior-preserving option
that fully satisfies the contract: apply YAGNI to skip or delete what is
unneeded, reuse existing repository code, use the standard library, use a
native platform capability, use an already-installed dependency, use
straightforward local code and direct control flow, and introduce an abstraction
only as the last option when it removes proven knowledge duplication or
coupling. Do not keep searching for a shorter option after an earlier rung fits,
and never use line count as the sole decision metric. When applicable, classify
the candidate as `delete`, `reuse`, `stdlib`, `native`, `dependency`, `yagni`,
or `shrink`; use `yagni` for speculative behavior or structure and `shrink` for
equivalent behavior expressed more directly. The tag does not replace severity,
evidence, or verification.

For comments, preserve irreducible rationale such as non-obvious decisions,
invariants and trust boundaries, compatibility or safety constraints, and
deliberate limitations. Prefer names, types, assertions, tests, or clearer
structure when they can express the same fact. Shorten useful prose to the
constraint and consequence; move cross-cutting decisions to durable documentation
and leave a local pointer when needed. Treat narration, stale or
contradictory prose, commented-out code, duplicated history, and speculative
guidance as candidates for `delete`, `shrink`, `reuse`, or `yagni` as
appropriate. Never remove security, validation, accessibility, data-loss, or
compatibility rationale without a durable equivalent, and never decide from
comment volume or line count alone.

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
for preserved behavior. Include its simplification kind when applicable. Ground
best-practices claims in repository rules, established neighboring code, or
authoritative documentation. Style preferences alone are not findings, and
`NO_FINDINGS` remains valid.

For implementation targets, establish expected behavior from accepted plans,
requirements, bugs, public contracts, repository rules, and user-visible
invariants rather than from the current implementation alone. Run the
specialized test lenses `intent-oracle`, `risk-edge-coverage`,
`level-fidelity`, and `economy-maintainability`. In a full review, dispatch them
in bounded waves when needed to preserve the configured reviewer limit.

Build a `test_evidence` matrix for every material changed behavior or failure
risk. Record its source, cheapest faithful test level, scenario, observable
oracle, evidence, distinct confidence or diagnostic value, residual risk, and
`covered`, `partial`, `missing`, or `accepted-risk` status. Missing or partial
evidence becomes a finding only when the risk is material, grounded, and in
scope; only an explicit decision may accept material residual risk.

Prefer public outcomes and state over private implementation interactions.
Require missing-test candidates to name a concrete behavior and distinguishing
scenario. Treat tests as redundant only when contract, scenario partition,
action, oracle, fidelity, and failure class coincide; preserve overlap that adds
a boundary, real-dependency check, ownership, or diagnosis. Share test mechanics
without hiding inputs or expected meaning. Coverage, mutation, test count, and
line count are contextual signals, never sufficient findings or universal
gates. Require synthetic fixtures with no credentials, tokens, transcripts,
runtime databases, or real user configuration.

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

Reviewers and verifiers remain read-only. Under repair disposition, the current
coordinating harness owns mutations and verification. Apply every confirmed fix
within the approved scope. Critical and High findings are blocking until
fixed or explicitly blocked by a user decision. For plans, edit the plan or
design artifact. For implementations, edit code and tests, run focused checks,
then run the repository-required final verification.

Under report-only disposition, do not apply fixes. Record every confirmed
finding in `applied_fixes` with status `not-applicable` and reason
`report-only disposition`; record `summary.applied` as zero. A passing gate
means that the review process completed, not that the implementation is
approved or merge-ready.

Write `review.md` and schema-valid `review.json` under
`.agent-runs/reviews/<run-id>/`. Report confirmed findings, applied changes,
refuted findings and rationale, checks, unresolved blockers, and gate status.
Record the selected `standard` or `maintainability` profile and `full` or
`tests` scope and selected `repair` or `report-only` disposition in the report.
Implementation reports include `test_evidence`.
