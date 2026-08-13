---
name: work-review
description: Run evidence-grounded multi-lens review of a plan, plan delta, diff, PR, or implementation, independently refute every candidate finding, and apply confirmed fixes.
version: 0.2.0
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
/work-review @agy @codex @claude -- implementation <target or focus>
```

Explicit mode and target win. In `auto`, select `implementation` when a diff or
PR exists, otherwise select `plan` when a plan/design artifact exists. Ask for
the target only when neither can be established. Never reject an explicit plan
because no implementation exists. Plan modes follow `/work-plan-review`.

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
provider/model attribution, and gate status.
