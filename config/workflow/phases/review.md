---
name: work-review
description: Run evidence-grounded multi-lens review of a plan, plan delta, diff, PR, or implementation, with an optional behavior-preserving maintainability profile, independent refutation, and applied fixes.
version: 0.4.0
---

# /work-review - Multi-Lens Review And Repair

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Review the requested artifact with fresh context. Reviewers are read-only; the
coordinator applies confirmed fixes and verifies the result.

Input:

`$ARGUMENTS`

Accepted modes:

```text
/work-review auto <target or focus>
/work-review plan <plan or design>
/work-review plan-delta <changed decision or section>
/work-review implementation <diff, branch, PR, contract, or focus>
/work-review implementation --profile maintainability <diff, branch, PR, or focus>
/work-review @agy @codex @claude -- implementation <target or focus>
```

Explicit mode and target win. In `auto`, select `implementation` when a diff or
PR exists, otherwise select `plan` when a plan/design artifact exists. Ask for
the target only when neither can be established. Never reject an explicit plan
because no implementation exists. Plan modes follow `/work-plan-review`.

The default profile is `standard`. `--profile maintainability` is available
only for implementation review. It keeps all implementation lenses, adds a
grounded `best-practices` lens, and deepens the correctness, consistency,
architecture, `simplicity-dry`, and `tests-verification` lenses. In `auto`, use
this profile only when the resolved target is an implementation; otherwise ask
for an implementation target instead of silently changing the profile.

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
- `tests-verification`: missing assertions, wrong test level, weak evidence,
  flaky behavior, and untested failure paths.

For dependency currency, use lockfiles plus official registries or primary
project sources when network access is available. Do not label a package stale
from memory, and do not recommend an upgrade without compatibility evidence.

Add domain lenses when warranted: accessibility, privacy/PII/SOC 2 evidence,
performance/scalability, migration safety, API compatibility, data integrity,
or operational observability. Do not claim compliance from generic practices.

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
- `correctness` and `tests-verification`: prove that a proposed simplification
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

Write the initial report, then apply every confirmed fix within approved scope:

- for a plan or plan-delta, edit the plan/design artifact;
- for an implementation, edit code and tests using repository conventions;
- when a fix needs a product decision, broader authority, migration approval,
  or unrelated scope, mark it blocked rather than guessing.

Critical and High findings are blocking. Run focused checks after each repair
group, then repository-required final verification. Re-run affected lenses when
a fix materially changes behavior. Gate status is `pass` only when confirmed
fixes are applied and checks pass; otherwise return `fail` or `blocked`.

Write `review.md` and schema-valid `review.json` under
`.agent-runs/reviews/<run-id>/`. Report confirmed findings and applied fixes
first, followed by refuted findings and rationale, checks, residual risk,
provider/model attribution, selected profile (`standard` or
`maintainability`), and gate status.
