---
name: work-review
description: Run evidence-grounded multi-lens review of a plan, plan delta, diff, PR, or implementation, with implementation repair or report-only disposition, an optional behavior-preserving maintainability profile, and independent refutation.
version: 0.7.0
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
/work-review implementation --scope tests --test-mode authoring <change or proposed tests>
/work-review implementation --scope tests --test-mode audit <test area>
/work-review implementation --scope tests --test-mode campaign <subsystem>
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

The default test mode is `review`. Every full implementation review uses that
mode. `--test-mode authoring`, `--test-mode audit`, and `--test-mode campaign`
require `implementation --scope tests`; never infer them during a normal full
review. Authoring gates proposed or newly changed tests. Audit examines a
bounded test area for low-value assurance and test-owned production seams.
Campaign is an explicit exhaustive audit of one subsystem and must name that
subsystem. Test mode changes the depth and required evidence, not authority,
disposition, severity, independent verification, or final verification.

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

### Gate Test Authoring

In `authoring` mode, and whenever the coordinator adds or materially changes a
test during repair, record one `test_authoring_gates` entry per test before
accepting it. Establish all of the following:

1. the observable behavior, invariant, or independent contract it protects;
2. a credible regression that makes the test fail;
3. why existing evidence does not already detect that regression;
4. whether the test requires an export, flag, wrapper, global, reset hook,
   injection point, or other production seam with no non-test caller.

A missing answer fails the authoring gate. Place each contract at one primary
owner: the cheapest faithful boundary that observes it without reproducing the
implementation. Another level may repeat a scenario only when it owns a
distinct transport, integration, lifecycle, compatibility, security, or
diagnostic risk. If a test needs a test-only production seam, move the proof to
the owning public boundary instead of weakening production design.

A bug regression test must demonstrably fail on the pre-fix implementation for
the intended reason and pass after the repair. Use an isolated reversible
control such as the known pre-fix revision or a deliberate local mutation when
repository policy and execution safety allow it. If the control cannot be run
safely, record it as blocked; a green post-fix test alone is not proof that it
detects the bug.

### Audit Existing Tests

`audit` and `campaign` modes inspect complete tests, their production owners,
entry points, callers, callees, sibling implementations, overlapping evidence,
CI routing, and relevant history. When a claim depends on a library or service,
inspect its repository-pinned source, types, or authoritative contract when
available. Discovery remains read-only until candidates are independently
verified.

Look for concrete low-value patterns, including:

- assertion-free probes, self-comparisons, or expected values produced by the
  same helper under test;
- copied inventories, manifests, exports, or exact source/import/string checks
  that do not independently protect a durable contract;
- private predicate, call-shape, or provider-local replays of behavior already
  proven at the owning boundary;
- repeated invocations of one contract without a distinct failure class;
- mocks or fixtures that manufacture the behavior or ordering being asserted;
- capability checks that restate declarations without exercising their
  promised result;
- negative controls that pass because of an unrelated guard or unreachable
  production path;
- names and fixtures that promise behavior the assertions do not exercise;
- tests whose sole purpose is preserving test-only production seams or dead
  production code.

These are discovery signals, not automatic deletion rules. Retain independently
meaningful API, protocol, configuration, migration, storage, security,
platform, default, generated cross-language, package, release, or architecture
contracts. Retain observable ordering and credible regressions. Static
inspection can be the cheapest faithful guard when it protects a durable byte,
key, path, or generated boundary and survives irrelevant refactoring. A slow
test or one that resembles implementation is not deletable without proof.
Treat a retained baseline failure as a possible product defect: reproduce the
behavior and repair the production owner separately rather than deleting its
only signal.

Before editing an audit candidate, add a `test_audit_candidates` entry with the
exact test and location, `retain`, `fix`, `consolidate`, or `delete`
disposition, the failure it can detect, its primary owner, non-test callers of
covered support seams, stronger surviving proof or why none is required,
relevant history, production or test-support deletion unlocked, risk, and a
focused validation command. Missing evidence means the candidate is not ready
to change. Prefer a few high-confidence candidates in ordinary audit mode.

Under repair disposition, change one coherent owner-boundary batch at a time.
Move retained regressions to their canonical owner, consolidate only after the
keeper absorbs the distinct assertion, and delete obsolete test-only exports,
globals, wrappers, hooks, and dead production paths rather than preserving
aliases. Under report-only disposition, record the same proposed edits without
changing the reviewed tree.

### Run A Test Campaign

`campaign` mode is an explicit exhaustive pruning pass over one subsystem. It
uses the authoring and audit rules above and completes these ordered steps:

1. pin a baseline revision and record the result of every owned test file,
   keeping baseline failures separate;
2. partition every test file and relevant QA or live-proof scenario into
   exactly one lane based on production ownership;
3. record every test declaration in the audit ledger, splitting parameterized
   rows only when they need different dispositions;
4. perform a second read-only layer pass that names the keeper for each
   contract, assertions to move, retired files, and seams unlocked;
5. under repair disposition, cut over lane by lane while one coordinator owns
   shared harness and support edits;
6. independently compare removed coverage with keepers and restore any lost
   contract; for each restored contract, use a deliberate reversible mutation
   to prove the keeper detects it;
7. classify persistent baseline failures as product defects, prove repairs with
   failing controls, and leave unrelated defects as follow-ups;
8. reconcile upstream changes, rerun the complete subsystem proof, and report
   production lines separately from test and test-support lines.

Record the subsystem, pinned baseline, baseline checks, ownership lanes,
preservation evidence, product defects, and before/after line counts in
`test_campaign`. A campaign cannot pass with unassigned tests, an incomplete
declaration ledger, unresolved preservation gaps, or unverified restored
contracts. Optimize for retained confidence and simpler ownership, never for a
deletion or line-count target.

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

### Collect Supplementary Maintainability Evidence

For the maintainability profile, bind the resolved target, base revision, head
revision, worktree state, changed paths, repository rules, and analyzer
configuration into one immutable review snapshot. Discover approved jscpd and
the existing repository-bounded GitNexus capability. Run each selected external
analysis pass at most once per review snapshot and share its immutable evidence
with the read-only lenses; do not let parallel reviewers rerun it. A repair that
changes relevant source creates a new snapshot and invalidates only the affected
passes.

The default requirement for both analyzers is `advisory`. Repository-owned
policy may set a tool to `disabled` or `required`. Record a complete
`analysis_tool_evidence` entry for jscpd and GitNexus even when a tool is
disabled, unavailable, partial, failed, or stale. When an advisory tool is
unavailable, continue with repository evidence and preserve the limitation. A
required unavailable or failed tool makes every affected inspection `unknown`
and blocks a maintainability pass. Distinguish a valid scan with zero candidates
from a failed scan or one that analyzed zero applicable files.

#### jscpd adapter

Use only the approved `jscpd` executable at version `5.3.2`. Resolve it from the
current environment, run `jscpd --version`, and record the executable, exact
version, and configuration identity before analysis. A missing or different
version is unavailable evidence under the default policy; never install,
upgrade, or replace it during review.

Prefer a committed `.jscpd.json` or another repository-owned supported config.
Otherwise create invocation configuration only beneath
`.agent-runs/reviews/<run-id>/evidence/jscpd/`. Respect `.gitignore`, repository
rules, and confirmed generated, vendored, cache, build, minified, and snapshot
surfaces. Do not exclude tests merely because they repeat code. Do not follow
symlinks unless repository rules and explicit authorization make that safe.

Collect bounded JSON evidence for the selected passes: exact clones;
identifier/literal-normalized renamed clones; explicitly enabled gap or AST
near-miss similarity where the installed version and language support it;
complexity; and language-supported dead-code candidates. Near-miss passes are
off by default. Use compact AI output only for bounded diagnostics; JSON is the
deterministic ingestion source. Use `--baseline-from-ref` only when the resolved
base exists locally, and never fetch during review. Preserve native clone kind,
method, score, source locations, exclusions, warnings, and counts. A project
duplication threshold or health score is never a finding or gate.

Apply a bounded timeout and evidence budget to each pass. Treat timeout,
malformed or missing JSON, non-zero execution, unsupported capability, and zero
applicable files as distinct states. Store raw and normalized task-owned
artifacts under the evidence directory and bound any copied source excerpts.
Parsing reviewed source does not authorize executing repository tests, builds,
scripts, hooks, generators, package lifecycle steps, or binaries.

#### GitNexus enrichment

Reuse the existing GitNexus installation, MCP/CLI policy, repository allowlist,
and index; never create a second graph. Before querying, record repository
identity, snapshot correspondence, index freshness, supported relationships,
and excluded, unresolved, or failed surfaces. A stale index is `stale` evidence,
not a clean result. Do not refresh it by running untrusted code, downloading
dependencies, accessing the network, or writing outside authorized index
locations.

For each bounded candidate, request only the enclosing and related symbols,
material direct or bounded transitive callers and callees, imports,
dependencies, inheritance, implementations, processes or routes, related tests,
related implementations, and change impact needed to judge the proposed
treatment. Prefer stable symbol identities. When resolution is ambiguous,
record the candidate identities and missing evidence instead of guessing. A
missing edge never proves that no caller, implementation, side effect, or
runtime path exists.

#### Correlate candidates

Group overlapping analyzer matches into stable snapshot-local candidate
families and cap both family count and locations per family. Preserve omitted
counts and the reason for truncation. Prioritize families that intersect changed
code, match an unchanged implementation, cross an ownership or layer boundary,
contain material validation, authorization, persistence, serialization, error,
or decision logic, reveal an existing reusable helper, or combine several weak
signals into a concrete risk.

For comparison candidates, record similarity, repeated responsibility, and
safe to share as independent judgments. Exact or normalized similarity does not
prove repeated knowledge; repeated responsibility does not prove that sharing
preserves behavior, errors, ordering, ownership, security, compatibility, or
coupling. Complexity and dead-code candidates use `not-applicable` for judgments
that do not apply and never invent a comparison target.

Write every reviewed family to `maintainability_candidates` with its signals,
relationships, counterevidence, missing evidence, treatment, and links to any
canonical finding IDs. Supply bounded packets to `simplicity-dry`,
`architecture`, `consistency`, `correctness`, `diff-analysis`, and
`test-review-bundle` as applicable. Tool candidates remain signals until the
normal lens evidence standard and independent refutation confirm a finding.

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

### Complete The Maintainability Inspection

Every maintainability review must complete four bounded inspection obligations
before it may return `NO_FINDINGS` or pass the gate. Inspect the relationships
that are material to the changed surface, including necessary unchanged code;
do not turn an inapplicable relationship type into ceremonial work. Existing
internal structure is evidence, not automatically a behavior contract. Preserve
public and extension contracts, relied-upon errors, trust boundaries, and
compatibility behavior.

- `alternative-comparison` is owned by `simplicity-dry`. Compare the current
  design with a concrete earlier rung of the simplification ladder. Inspect
  introduced concepts, state, branches, layers, forwarding wrappers, and their
  important consumers. Explain why the smaller shape does or does not preserve
  required behavior and safeguards; line count alone is insufficient.
- `contract-coherence` is owned by `consistency` and `architecture`. Compare
  affected base declarations, implementations, callers, and supported extension
  points. Inspect nullability and reachable absence, nested or repeated type
  shapes, and whether a generic boundary or its documentation encodes behavior
  owned only by a concrete implementation.
- `error-and-validation-flow` is owned by `correctness` and `architecture`.
  Trace at least one representative success path and each materially distinct
  failure path through validation, conversion, adaptation, and error
  translation. Inspect whether error types correspond to distinct recovery or
  compatibility behavior and whether their ownership matches repository rules.
- `runtime-invariant-placement` is owned by `correctness`. Distinguish failures
  caused by runtime or caller-controlled data from developer-only invariants.
  Inspect value origin, reachability, repeated checks on operational paths, and
  the failure semantics that would remain after simplification. Never remove a
  validation or safety check merely because the type declaration appears to
  promise the invariant.

For each obligation, add one `maintainability_inspections` entry to `review.json`
with its owning lenses, concrete evidence, rationale, linked candidate finding
IDs, and missing evidence. Record the status as `clear`, `candidate`, `unknown`, or `not-applicable`.
`unknown` names the missing evidence and prevents a pass;
`not-applicable` explains why the relationship does not exist in the reviewed
surface. Under repair disposition, a `candidate` inspection may pass only after
every linked finding is refuted or its confirmed fix is applied and verified.
Under report-only disposition, it may pass when every linked finding is
independently resolved as refuted or confirmed and each confirmed fix is
recorded as `not-applicable`; this passes the review process, not the
implementation. `NO_FINDINGS` is valid only when every owned obligation is
`clear` or `not-applicable` with supporting evidence.

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
a fix materially changes behavior. For the maintainability profile, also rerun
the affected analyzer passes against the repaired snapshot and preserve both
the original and replacement evidence identities. Gate status is `pass` only
when confirmed fixes are applied and checks pass; otherwise return `fail` or
`blocked`.

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
(`repair` or `report-only`), selected `test_review_mode`, and the test-evidence
matrix for implementation reviews. Include `test_authoring_gates`,
`test_audit_candidates`, and `test_campaign` when their modes require them, the
maintainability-inspection matrix when that profile is selected, the
`analysis_tool_evidence` and `maintainability_candidates` collections, and gate
status. Link confirmed candidates to canonical findings, fixes through finding
IDs, test evidence, inspections, and independent verification without allowing
raw analyzer output to bypass the canonical report.
