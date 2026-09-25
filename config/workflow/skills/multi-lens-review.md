---
name: lens-review
description: Use for plan, design, decision-delta, diff, implementation, or delegated review that needs independent lenses, implementation repair or report-only disposition, behavior-preserving maintainability review, and adversarial verification.
version: 0.7.0
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

Resolve test mode as `review` unless an explicit tests-scope invocation selects
`authoring`, `audit`, or `campaign`. Full implementation review always embeds
`review`; it never silently expands into an audit. Authoring evaluates proposed
or newly changed tests, audit examines a bounded area, and campaign exhaustively
examines one explicitly named subsystem.

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

For maintainability reviews, complete four bounded inspection obligations before
returning `NO_FINDINGS` or passing the gate:

- `alternative-comparison` (`simplicity-dry`): compare a concrete earlier
  simplification rung, introduced concepts/state/branches/layers, forwarding
  wrappers, important consumers, and preserved safeguards;
- `contract-coherence` (`consistency`, `architecture`): inspect affected base,
  implementation, caller, and extension contracts; nullability and reachable
  absence; type-shape burden; and concrete behavior leaking into a generic
  boundary or its documentation;
- `error-and-validation-flow` (`correctness`, `architecture`): trace
  representative success and materially distinct failure paths across
  validation, conversion, adaptation, and error translation, including whether
  error types have distinct recovery behavior and coherent ownership;
- `runtime-invariant-placement` (`correctness`): distinguish caller-controlled
  failures from developer-only invariants using value origin and reachability,
  then inspect repeated operational-path checks and preserved failure semantics.

Record one `maintainability_inspections` entry per obligation with lenses,
evidence, rationale, linked finding IDs, and missing evidence. Its status is
`clear`, `candidate`, `unknown`, or `not-applicable`. `unknown` prevents a pass;
`not-applicable` needs a scoped reason. Under repair disposition, a `candidate`
inspection may pass only after every linked finding is refuted or its confirmed
fix is applied and verified. Under report-only disposition, it may pass when
every linked finding is independently resolved as refuted or confirmed and each
confirmed fix is recorded as `not-applicable`; this passes the review process,
not the implementation. `NO_FINDINGS` is valid only when each owned obligation
is `clear` or `not-applicable` with evidence. Existing internal structure is not
automatically an accepted behavior contract, and declared types never justify
removing runtime validation or safety checks without reachability and failure
evidence.

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

For maintainability review, bind base, head, worktree state, changed paths,
repository rules, and analyzer configuration into one immutable snapshot.
Discover the approved jscpd version and the existing repository-bounded
GitNexus index. Run each selected pass at most once per review snapshot, store
raw and normalized task-owned evidence, and share it with read-only lenses. A
relevant repair creates a new snapshot and reruns only affected passes.

Verify `jscpd --version` is exactly `5.3.2`; never install, upgrade, fetch, or
access the network during review. Prefer repository-owned configuration and
otherwise write fallback config only under the review evidence directory.
Respect confirmed generated, vendored, cache, build, minified, and snapshot
exclusions without categorically excluding tests. Collect bounded JSON for
exact, normalized, explicitly enabled near-miss, complexity, and supported
dead-code passes. Use `--baseline-from-ref` only for a locally available base.
Distinguish zero findings from zero applicable files, timeout, malformed output,
unsupported analysis, and failure. Threshold and health scores are never
findings or gates.

Reuse GitNexus rather than creating another graph. Record repository identity,
snapshot correspondence, index freshness, supported relationships, and missing
surfaces before using it. Query only bounded symbols, callers/callees,
dependencies, implementations, processes, tests, related code, and impact.
Preserve stale, partial, and ambiguous evidence; a missing edge is not proof of
absence.

Group overlaps into bounded candidate families. Keep similarity, repeated
responsibility, and safe to share as separate judgments. Feed candidates to the
applicable canonical lenses and independent refutation; analyzer output is a
signal, never a finding. Record producer state and coverage in
`analysis_tool_evidence`, candidate signals and counterevidence in
`maintainability_candidates`, and link confirmed candidates through canonical
finding IDs. Advisory absence is visible but non-blocking; required unavailable
or failed evidence makes affected inspections `unknown` and blocks a pass.

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

In authoring mode, record the protected observable contract, a credible
regression, why current evidence misses it, the primary owner boundary, and any
production seam the test would require. One contract has one primary owner at
the cheapest faithful boundary; repeated proof at another level needs a
distinct risk. A bug regression must fail on the pre-fix behavior for the
intended reason and pass after repair. If a safe control run is unavailable,
record the proof as blocked instead of inferring it from a green test.

In audit mode, read each candidate test and its production owner, callers,
sibling implementations, overlapping tests, CI routing, and relevant history.
Use concrete low-value patterns only for discovery: vacuous assertions,
self-derived expected values, copied inventories, private-call replays,
duplicate contract invocations, behavior manufactured by mocks, unrelated
negative controls, overstated test names, and test-only production seams. Do
not delete from pattern resemblance alone. Record the exact test, detected
failure, primary owner, non-test callers, surviving proof, history, unlocked
deletions, risk, validation, and `retain`, `fix`, `consolidate`, or `delete`
disposition before editing. Remove obsolete test-only exports, wrappers,
globals, hooks, and dead production paths with confirmed candidates.

Campaign mode pins a baseline for every owned test, partitions the subsystem by
production ownership, gives every declaration a disposition, names one keeper
per contract, cuts over by owner lane, and independently reviews preservation.
Prove each restored contract with a reversible deliberate mutation. Treat
persistent baseline failures as possible product defects. A campaign report
records `test_audit_candidates`, `test_campaign`, and production versus test
and support line counts; deletion count is never a success metric.

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
Implementation reports include `test_review_mode` and `test_evidence`;
authoring reports include `test_authoring_gates`, audit reports include
`test_audit_candidates`, campaign reports include `test_campaign`, and
maintainability reports also include `maintainability_inspections`,
`analysis_tool_evidence`, and `maintainability_candidates`.
