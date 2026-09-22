---
name: work-review
description: Run evidence-grounded multi-lens review of a plan, plan delta, diff, PR, or implementation, with implementation repair or report-only disposition, optional maintainability focus, and independent refutation.
version: 0.5.0
---

# /work-review - Multi-Lens Review And Repair

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Review the requested artifact with fresh context. Reviewers are read-only. The
selected review disposition determines whether the coordinator repairs the
artifact or only reports grounded findings.

Input:

`$ARGUMENTS`

Accepted modes:

```text
/work-review auto <target or focus>
/work-review plan <plan or design>
/work-review plan-delta <changed decision or section>
/work-review implementation <diff, branch, PR, contract, or focus>
/work-review implementation --scope tests <diff, branch, PR, contract, or focus>
/work-review implementation --profile maintainability <diff, branch, PR, or focus>
/work-review implementation --disposition report-only <PR or diff>
/work-review @agy @codex @claude -- implementation <target or focus>
```

Explicit mode and target win. In `auto`, select `implementation` when a diff or
PR exists, otherwise select `plan` when a plan/design artifact exists. Ask for
the target only when neither can be established. Never reject an explicit plan
because no implementation exists. Plan modes follow `/work-plan-review`.

The default profile is `standard`. `--profile maintainability` is available
only for implementation review. It keeps all implementation lenses, adds a
grounded `best-practices` lens, and deepens the correctness, consistency,
architecture, `simplicity-dry`, and `test-review-bundle` lenses. In `auto`, use
this profile only when the resolved target is an implementation; otherwise ask
for an implementation target instead of silently changing the profile.

The default scope is `full`. `--scope tests` is available only for
implementation review and is the canonical expansion of `/work-test-review`.
It runs the specialized test-review bundle and the inspection needed to ground
or refute its findings, but omits unrelated implementation lenses. Every full
implementation review embeds the same bundle in place of a shallow generic
test lens. The scope changes review focus, not reviewer authority, repair rules,
or final verification.

The default review disposition is `--disposition repair`, preserving the
normal review-and-repair workflow. `--disposition report-only` is available for
implementation review when the output will be advisory or published as review
feedback. Under report-only disposition every reviewer and the coordinator
must not edit the reviewed implementation, its tests, plan, or contributor
branch. They may write only task-owned review artifacts. A report-only run
still verifies and deduplicates every candidate, runs safe read-only checks,
and records concrete proposed fixes; it never marks a finding applied.

For a PR, fork, patch, or other externally controlled implementation, treat
the reviewed code as untrusted executable content. Prefer provider CI evidence
and static inspection. A read-only disposition does not authorize executing
review-controlled tests, builds, scripts, hooks, package lifecycle steps,
generators, or binaries on the host. Such execution requires explicit user
authorization and a disposable credential-free sandbox with no host writes,
no network by default, isolated caches, and bounded resources.

An existing PR chosen for PR publication by `/work-start-pr-review` must use
report-only. An internal repair review uses repair unless the user explicitly
asks for advice without edits. The disposition changes mutation authority, not
the selected lenses, evidence standard, severity, or routing strategy.

When `work-routing` resolves several harnesses, use `parallel-verify`: distribute
independent read-only lenses across them, prefer a verifier from a different
harness than each finding origin, and preserve attribution and disagreement.
The current harness remains coordinator and owns every confirmed fix.
Any subset of `@codex`, `@claude`, `@agy`, and `@hermes` may be requested;
`work-routing` filters it through the current safe-runner contract and never
silently substitutes an unavailable harness.

## Establish Evidence

Discover repository rules, accepted contract, base ref, changed files, actual
code around every changed path, tests, generated-file rules, CI, migrations,
and verification evidence. A diff identifies changed behavior but is not enough
context by itself. Do not trust builder summaries as proof.

## Run Independent Implementation Lenses

Launch one read-only reviewer per applicable base lens:

- `correctness`: observable behavior, state transitions, errors, and invariants;
- `consistency`: repository conventions, sibling behavior, naming, and contracts;
- `completeness-edge-cases`: missing paths, boundaries, failures, concurrency,
  cleanup, and recovery;
- `security`: authorization, injection, secrets, privacy, unsafe defaults, and
  data exposure;
- `architecture`: ownership boundaries, coupling, layering, API compatibility,
  and system-wide consequences;
- `simplicity-dry`: avoid duplication and needless complexity without forcing
  premature abstraction;
- `diff-analysis`: unintended changes, stale assumptions, generated output,
  migrations, and behavior outside the stated scope;
- `dependency-currency`: newly added direct dependencies, non-latest selections,
  stale sibling dependencies in the touched area, advisories, and compatibility;
- `test-review-bundle`: the specialized intent, risk, fidelity, economy, and
  maintainability review defined below.

For dependency currency, use lockfiles plus official registries or primary
project sources when network access is available. Do not label a package stale
from memory, and do not recommend an upgrade without compatibility evidence.

Add domain lenses when warranted: accessibility, privacy/PII/SOC 2 evidence,
performance/scalability, migration safety, API compatibility, data integrity,
or operational observability. Do not claim compliance from generic practices.

## Run The Specialized Test Review

Every full implementation review runs this bundle. `--scope tests` runs only
this bundle plus any code, contract, or runtime inspection required to ground
and refute its findings. In a full review, schedule its lenses in bounded waves
so the configured reviewer limit remains effective.

Establish the accepted behavior from plans, requirements, bug reports, public
contracts, repository rules, and user-visible invariants before judging the
tests. Treat the current implementation as evidence, not as the sole source of
expected behavior. Read the implementation diff, test diff, affected existing
tests, repository-owned test commands, and available verification results.

Run one read-only reviewer per specialized lens:

- `intent-oracle`: requirement and risk grounding, meaningful observable
  assertions, tautologies, and accidental coupling to private implementation;
- `risk-edge-coverage`: relevant normal cases, equivalence classes,
  boundaries, condition combinations, state transitions, invalid inputs,
  permissions, dependency failures, partial effects, cancellation, timeouts,
  retries, concurrency, cleanup, rollback, and recovery;
- `level-fidelity`: whether unit, contract, integration, end-to-end, fuzz,
  property, static, or manual evidence is the cheapest faithful level, and
  whether mocks or fakes remove the behavior that creates the risk;
- `economy-maintainability`: redundant assurance, hidden test logic, brittle
  setup, unsafe fixtures, nondeterminism, isolation failures, weak diagnostics,
  and disproportionate runtime or continuing maintenance cost.

For each material changed behavior or failure risk, add a `test_evidence` entry
to `review.json` with its accepted source, selected test level, scenario,
observable oracle, current or proposed evidence, distinct confidence or
diagnostic value, residual risk, and `covered`, `partial`, `missing`, or
`accepted-risk` status. Only an explicit user or repository-owned decision can
accept material residual risk.

A test is justified when it provides a credible observable signal for a
material behavior or risk at the cheapest faithful level. Prefer public
behavior, returned values, resulting state, errors, durable side effects, and
user-visible outcomes. Interaction assertions are appropriate only when the
interaction itself is an accepted contract. Do not expose private structure
solely to test it.

Keep tests complete, concise, deterministic, and locally readable. Helpers may
share mechanics, but must not hide scenario inputs, expected behavior, or
important assertions. Repeated setup is acceptable when it preserves clarity.
Treat tests as consolidation candidates only when they protect the same
contract with the same scenario partition, action, oracle, fidelity, and
failure class. Similar-looking tests can remain valuable when they add a
different boundary, real-dependency check, contract owner, or diagnostic
signal.

A missing-test candidate names the unprotected behavior or risk and gives a
concrete scenario that distinguishes correct from incorrect behavior. Do not
request tests for trivial lines, getters, generated output, or unreachable
states without a product-risk explanation. An evidence entry marked `missing`
or `partial` becomes a finding only when the uncovered risk is material,
grounded, and within scope.

Use line, branch, and diff coverage to locate unexamined changed code. Use
fuzzing or mutation testing only when discovered repository tooling and risk
warrant them. Coverage, mutants, test count, and line count are contextual
signals and never sufficient findings or universal numeric gates. Do not
install tools or enable services without approval.

Fixtures must be synthetic and contain no credentials, tokens, transcripts,
runtime databases, or real user configuration values.

## Run The Maintainability Profile

For `--profile maintainability`, first establish the observable behavior that
must remain unchanged: public and internal contracts, outputs, side effects,
errors, ordering, compatibility, and relevant tests. Run every normal
implementation lens plus `best-practices`; do not trade correctness, security,
or edge-case coverage for a smaller diff.

Before selecting practices or checks, discover the target's active languages,
frameworks, build systems, and generated surfaces. Keep this discovery
language-agnostic, evidence-led and fallible; it is not a deterministic
language-to-command lookup. Inspect repository instructions, CI and hooks,
build and package manifests, lockfiles, source and generated-file markers, and
the changed paths. Record the result as `detected`, `mixed`, or `unknown`, with
supporting evidence, confidence, and unresolved ambiguity. Do not infer an
ecosystem from a file extension alone. In a multi-language repository, resolve
the context and checks for every materially affected surface.

Prefer repository-owned commands declared by instructions, CI, hooks, or build
configuration. When those are incomplete, consider only already-available
tools relevant to the discovered context and label supplementary checks as
such. Do not install or enable tooling, access the network, or turn a remembered
ecosystem convention into a required gate without approval and authoritative
support. Preserve `unknown` instead of inventing certainty. Record every
selected command, why it applies, and its result so execution remains
reproducible even when context discovery is uncertain.

After establishing the behavior contract, evaluate simplifications in this
order and stop at the first behavior-preserving option that fully satisfies it:

1. apply YAGNI: skip or delete behavior and structure that no accepted
   requirement, caller, or invariant needs;
2. reuse an established repository abstraction, helper, or pattern;
3. use the active language's standard library;
4. use a native platform capability from the browser, database, operating
   system, framework, or protocol;
5. use an already-installed dependency when its reviewed contract fits;
6. use straightforward local code and direct control flow;
7. introduce or retain an abstraction only when it removes proven repeated
   knowledge, creates clear ownership, or materially reduces coupling.

Do not continue down the ladder merely because a later option is shorter. The
first fit must still preserve security, validation, accessibility, data-loss
prevention, error behavior, compatibility, and required verification. For an
applicable simplification candidate, record one kind: `delete`, `reuse`,
`stdlib`, `native`, `dependency`, `yagni`, or `shrink`. Use `yagni` for
speculative behavior or structure and `shrink` for the same behavior expressed
more directly. The kind aids triage; it does not replace severity, evidence, or
independent refutation. A line count may support the net-simplification estimate
but is never sufficient evidence or the sole decision metric.

Treat comments as maintainability assets, not free documentation or a line-count
problem. Preserve irreducible rationale: non-obvious local decisions,
invariants and trust boundaries, compatibility or safety constraints, and
deliberate limitations or upgrade paths that the implementation cannot express.
Before retaining explanatory prose, prefer names, types, assertions, tests, or
clearer structure when they can make the same fact executable or self-evident.
Shorten useful but verbose comments to the constraint and consequence. Move
cross-cutting tradeoffs or decision history to durable documentation and leave
a short local pointer when discoverability matters.

Flag code narration, stale or contradictory comments, commented-out code,
speculative future guidance, and duplicated decision history only when there is
concrete evidence of a maintainability cost. Classify redundant or stale prose
as `delete`, useful but verbose rationale as `shrink`, centralized rationale as
`reuse`, and speculative guidance as `yagni`. Do not remove security,
validation, accessibility, data-loss, or compatibility rationale without an
equally durable equivalent. Comment volume and line count are never sufficient
evidence by themselves.

Deepen the focused lenses as follows:

- `consistency` and `best-practices`: enforce repository rules and established
  local patterns. Ground external practices in authoritative documentation.
  Reject taste-only or cargo-cult findings.
- `simplicity-dry`: identify knowledge duplication, repeated decision logic,
  and unnecessary branches, state, indirection, or configuration. Distinguish
  those from incidental duplication that merely looks similar.
- `architecture`: propose a better abstraction only when it creates clear
  ownership, reduces coupling, or removes repeated knowledge. Reject a
  speculative abstraction, single-use generic helper, or extra layer that
  increases the number of concepts.
- `correctness` and `test-review-bundle`: prove that a proposed simplification
  preserves the established behavior contract, including failure paths.

Every candidate from this profile must identify the preserved behavior and
concrete evidence of the duplication or complexity. It must state the minimal
fix and net simplification, regression risk, and the exact verification plan.
Prefer deletion, direct control flow, and existing abstractions when they solve
the problem. `NO_FINDINGS` is correct when no evidence-grounded simplification
improves the code.

Every reviewer reads the actual code and returns candidates with severity,
claim, impact, proposed fix, and concrete `file:line`, short verbatim quote,
command/test, or authoritative-document evidence. `NO_FINDINGS` is valid.

## Verify, Rank, And Deduplicate

For each candidate, spawn an independent verifier that tries to refute it
against code, tests, docs, and runtime evidence. Require explicit `is_real` and
`grounded` booleans. Keep only `is_real && grounded`; record everything refuted
and why. Merge duplicates and rank confirmed findings Critical, High, Medium,
then Low.

## Report And Apply

Write the initial report. Under repair disposition, apply every confirmed fix
within approved scope:

- for a plan or plan-delta, edit the plan/design artifact;
- for an implementation, edit code and tests using repository conventions;
- when a fix needs a product decision, broader authority, migration approval,
  or unrelated scope, mark it blocked rather than guessing.

Critical and High findings are blocking. Run focused checks after each repair
group, then repository-required final verification. Re-run affected lenses when
a fix materially changes behavior. Gate status is `pass` only when confirmed
fixes are applied and checks pass; otherwise return `fail` or `blocked`.

Under report-only disposition, do not apply fixes. Mark every confirmed
finding's `applied_fixes` entry `not-applicable` with `report-only disposition`
as the reason. Gate status describes completion of the review process: `pass`
when the requested lenses, independent verification, deduplication, and safe
checks completed, even when the implementation has confirmed findings;
`fail` when review evidence or checks are invalid; and `blocked` when required
evidence could not be obtained. A report-only `pass` is not approval of the
implementation and must never be presented as merge readiness.

Write `review.md` and schema-valid `review.json` under
`.agent-runs/reviews/<run-id>/`. Report confirmed findings and applied fixes
first, followed by refuted findings and rationale, checks, residual risk,
provider/model attribution, selected profile (`standard` or
`maintainability`), selected scope (`full` or `tests`), review disposition
(`repair` or `report-only`), the test-evidence matrix for implementation
reviews, and gate status.
