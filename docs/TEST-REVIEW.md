# Specialized Test Review

## Status

Accepted and implemented, 2026-09-10. Expanded with authoring, audit, and
campaign modes, 2026-09-24.

## Context

Standard work already verified and reviewed implementation changes, but the
former general `tests-verification` lens was too compact to judge AI-generated
tests consistently. Passing tests and increased coverage can still hide weak
assertions, implementation-coupled expectations, missing failure paths, and
several tests that provide the same assurance.

The desired outcome is a minimal sufficient test portfolio. Each important
changed behavior and material failure risk should have credible evidence, and
each test should have a defensible reason to exist. Test count, test lines, code
coverage, and mutation score are supporting signals rather than goals.

## Decision

Embed a specialized test-review bundle in the existing implementation-review
gate. Expose the same capability through a standalone `/work-test-review`
entrypoint for test-heavy and test-only changes.

Do not add another human approval gate. The specialized result feeds the
existing implementation review and `review.json`. Reuse the existing candidate
verification, deduplication, repair, and reporting machinery instead of
creating a parallel review system.

Confirmed Critical and High assurance gaps block completion. Lower-severity
maintainability findings are advisory unless they demonstrate false assurance,
flakiness, security exposure, or significant ongoing maintenance cost.

## Test Review Modes

`review` remains the default mode embedded in every implementation review. It
maps changed behavior and risks to sufficient evidence without inventorying an
unchanged test surface.

Authoring mode gates every proposed, new, or materially changed test. Before a
test is accepted, it must name the observable contract it protects, a credible
regression, the gap in existing coverage, its primary owner boundary, and any
production seam it requires. A bug regression must fail on the pre-fix behavior
for the intended reason and pass after the repair. A test that needs a test-only
production export, wrapper, global, reset hook, or injection point moves to the
real owning boundary instead.

Confirmed cleanup removes obsolete test-only production seams together with
the tests that kept them alive; it does not preserve compatibility aliases for
code with no production caller.

Audit mode inspects a bounded test area for false assurance, implementation
coupling, duplicated contracts, and obsolete test support. Each candidate is
marked `retain`, `fix`, `consolidate`, or `delete` only after its actual failure
signal, production owner, non-test callers, surviving proof, history, unlocked
deletions, risk, and focused validation are known.

Campaign mode applies the same evidence bar exhaustively to one explicitly
named subsystem. It pins a baseline, partitions tests by production owner,
records every declaration, names contract keepers, performs lane-by-lane
cutover, and runs an independent preservation review. Restored coverage must
catch a reversible deliberate mutation. Persistent baseline failures are
investigated as possible product defects instead of being discarded as stale
tests.

The focused commands are:

```text
/work-test-review <target or focus>
/work-test-review authoring <change or proposed tests>
/work-test-review audit <test area>
/work-test-review campaign <subsystem>
```

The non-default modes are never inferred by a normal implementation review.
They reuse the existing repair/report-only dispositions, independent candidate
refutation, and final verification.

## Test Standard

A test is justified when it provides a credible observable signal for a
material behavior or risk at the cheapest faithful test level.

Review each test against these questions:

1. Which accepted requirement, public contract, invariant, bug, or product risk
   does it protect?
2. Why is unit, contract, integration, end-to-end, fuzz, or another test level
   the smallest level that can exercise that risk faithfully?
3. Which returned value, resulting state, error, durable side effect, or other
   user-visible outcome is the oracle?
4. Which normal case, equivalence class, boundary, condition combination, state
   transition, invalid input, dependency failure, timeout, cancellation,
   concurrency path, cleanup, or recovery behavior does it cover?
5. Is its purpose and failure reason clear from its name, setup, action,
   assertion, and diagnostic output?
6. What confidence would disappear if the test were removed?

Prefer assertions against public behavior and resulting state. Interaction
assertions are appropriate only when the interaction itself is an accepted
contract. Do not make private structure observable merely to test it.

Keep tests complete, concise, deterministic, and locally readable. Test helpers
may share mechanics, but should not hide scenario inputs, expected behavior, or
important assertions. Some repeated setup is acceptable when it preserves test
clarity.

Two tests are candidates for consolidation only when they protect the same
contract with the same scenario partition, action, observable, fidelity, and
failure class. Similar-looking tests remain justified when they provide a
different boundary, real-dependency evidence, contract owner, or diagnostic
signal.

One contract has one primary owner at the cheapest faithful boundary. A second
layer needs a distinct transport, integration, lifecycle, compatibility,
security, or diagnostic risk; merely traversing the same call path does not
justify another copy.

Audit discovery explicitly checks for vacuous assertions, self-derived
expected values, copied inventories, source greps without an independent
contract, private-call replays, duplicate contract invocations, mocks that
manufacture the asserted behavior, unrelated negative controls, overstated
test names, and production code used only by tests. These are candidate signals
rather than automatic deletion rules. Static checks, observable ordering,
cross-system contracts, and credible regressions remain when they independently
protect behavior.

## Review Contract

The specialized review contains four evidence-grounded lenses:

- `intent-oracle`: requirement grounding, behavioral assertions, and accidental
  implementation coupling;
- `risk-edge-coverage`: behavior and risk coverage, boundaries, invalid paths,
  state combinations, failures, cleanup, and recovery;
- `level-fidelity`: the selected test level, dependency fidelity, doubles, and
  whether a smaller or more realistic test would give better evidence;
- `economy-maintainability`: redundant assurance, test logic and abstraction,
  determinism, runtime cost, isolation, diagnostics, and fixture safety.

Each candidate finding must include concrete repository evidence, the affected
contract or risk, impact, the smallest proposed correction, and an executable
verification method. An independent verifier attempts to refute it before it
is confirmed. `NO_FINDINGS` is valid.

The review produces a behavior/risk-to-evidence matrix with:

| Field | Meaning |
|---|---|
| Contract or risk | The behavior or failure that matters |
| Source | Plan, requirement, bug, API, invariant, or repository rule |
| Test level | The selected scope and why it is credible |
| Scenario | Normal, boundary, invalid, transition, failure, or recovery case |
| Oracle | The observable result that proves the behavior |
| Evidence | Existing or proposed tests and verification commands |
| Distinct value | Confidence lost if this evidence is removed |
| Residual risk | Material behavior intentionally left untested and why |

Coverage should locate relevant unexercised changed code, not justify tests by
percentage alone. Fuzzing and mutation testing may probe input spaces and
assertion strength when repository tooling and risk warrant them; neither is a
universal numeric gate.

Authoring reports add `test_authoring_gates`. Audit and campaign reports add
`test_audit_candidates`, including the exact test, disposition, detected
failure, primary owner, non-test callers, surviving proof, history, unlocked
deletions, risk, and validation. Campaign reports also add `test_campaign` with
its subsystem, baseline, lanes, preservation evidence, product defects, and
production versus test-support line counts. All implementation reports record
their `test_review_mode`.

## Options Considered

### Extend the existing lens only

Rejected because a short checklist remains too shallow for consistent analysis
of test purpose, scenario selection, fidelity, and redundant assurance.

### Add a separate mandatory phase and human gate

Rejected for the initial design because it would duplicate repair loops and
artifacts, increase latency, and add approval burden without adding authority.

### Use numeric quality thresholds

Rejected as the primary acceptance model because coverage, mutation score,
test count, and line count are context-sensitive and can be optimized without
improving product assurance.

## Accepted Risks And Safeguards

- The reviewer may invent requirements. Findings must cite an accepted source;
  ambiguity and residual risk remain explicit.
- The reviewer may encourage test bloat. Every proposed test states what
  confidence would otherwise be absent.
- The reviewer may remove useful overlap. Different fidelity, ownership,
  boundaries, and diagnostics justify intentional overlap.
- Generated fixtures may expose sensitive state. Fixtures remain synthetic and
  contain no credentials, tokens, transcripts, runtime databases, or real user
  configuration.
- Review cost may grow. Start with changed behavior and affected tests, and
  expand only across demonstrated contract boundaries.

## Implementation

The provider-neutral sources are:

- `config/workflow/phases/review.md` for the canonical embedded bundle;
- `config/workflow/phases/test-review.md` for the thin standalone command;
- `config/workflow/skills/multi-lens-review.md` for shared execution rules;
- `config/workflow/review-policy.json` for lenses and evidence fields;
- `config/schema/review-report.schema.json` for report validation;
- `config/manifest.json`, `config/presets/standard-work.json`, and
  `config/presets/multi-lens-review.json` for provider rendering and
  installation.
