# Coding Factory Feature Shape

## Status

Proposal for discussion. This document defines a candidate product boundary and
acceptance contract; it does not approve implementation or make the factory a
`maisternia` runtime.

## Intent

A coding factory accepts bounded engineering tasks and moves them toward
reviewable pull requests through a durable pool of interchangeable workers.
Each worker takes one user-verifiable vertical slice, records evidence, and
returns capacity to the pool. Work survives worker exits, provider limits,
machine restarts, and delayed verification because task state belongs to the
factory rather than to an agent session.

The factory should optimize for trustworthy throughput:

- small changes that a user can inspect and accept independently;
- progress that can resume from durable state;
- concurrency without overlapping write ownership;
- verification that does not keep a worker idle;
- bounded authority, cost, retries, and wall-clock time;
- proposals for process improvements without autonomous policy mutation.

## Product Boundary

Kaji is the candidate home for the live factory runtime. It would own:

- task intake, eligibility, prioritization, and queueing;
- atomic leases and recovery of abandoned work;
- worker launch, pause, resume, replacement, and shutdown;
- isolated workspaces and branch lifecycle;
- runtime budgets, concurrency, retries, and backpressure;
- verification receipts and state transitions;
- forge publication, status projection, and operational telemetry.

AgentnykMaisternia would own the declarative integration surface:

- a versioned, provider-neutral factory contract;
- installable worker, coordinator, verification, and review workflows;
- provider and Kaji target mappings;
- capability and approval policies;
- schemas for task packets, worker results, and redacted evidence;
- safe preview, conflict detection, apply, drift checks, and rollback.

The selected coding harness continues to own its sessions, tool calls, sandbox,
and native approval prompts. The source forge and CI system own their native
records and build execution. Maisternia may render configuration for each
participant but must not observe or control a live factory.

```text
maisternia
  define -> validate -> render -> install
                                  |
                                  v
kaji runtime -> task lease -> coding harness -> candidate change
      ^                                      |
      |                                      v
      +----------- durable evidence <- verification/forge
```

## Unit Of Work

The unit of scheduling is a task packet for one vertical slice. A packet must be
understandable without access to a previous worker's conversation and contain:

- stable factory, project, task, and slice identifiers;
- a pinned base revision and target branch;
- the user outcome and observable acceptance criteria;
- allowed write paths and protected paths;
- relevant repository instructions and trusted policy digests;
- required local and remote verification;
- the available capabilities and explicit authority envelope;
- model route, token, cost, attempt, and wall-time limits;
- dependency, priority, and blocking relationships;
- lease identity, expiry, and idempotency key.

Issue bodies, comments, build logs, and other external content are untrusted
inputs. They may describe desired behavior or supply evidence, but they cannot
grant authority, widen scope, change protected paths, or override trusted
project policy.

## Durable State Model

The authoritative state should be structured and machine-validatable. Markdown
may present a readable projection, but it should not act as a lock or require
agents to infer transitions from prose.

```text
candidate -> ready -> leased -> implementing -> verification_pending
   ^                    |              |                 |
   |                    |              v                 v
   +-- replan/clarify <-+----------> parked       review_ready
                         lease expiry                    |
                               +------> ready            v
                                                     accepted
```

Every mutation should append an immutable event and update the current task
record with compare-and-set semantics. Commands must accept idempotency keys so
that retries cannot create a second lease, branch, status transition, comment,
or pull request.

Minimum event data includes the task and slice IDs, prior and next state,
actor/service identity, lease generation, base and result revisions, timestamp,
policy digest, redacted evidence references, and stop reason. Raw prompts,
transcripts, secrets, and full environment captures are not factory records.

## Worker Cycle

One scheduling cycle is deliberately small:

1. Select an eligible task whose dependencies and authority requirements are
   satisfied.
2. Acquire an atomic, expiring lease and prepare an isolated workspace.
3. Give the worker a self-contained task packet.
4. Implement one user-visible slice and run focused local checks.
5. Publish a candidate revision and immutable evidence receipt.
6. Release the worker while remote verification runs asynchronously.
7. Consume the verification result when it arrives and either advance, retry,
   park, or request human judgment.
8. Open or update a reviewable pull request only when the publication policy
   permits it.

Any compatible worker may process the next transition. Correctness must not
depend on a specific agent retaining conversational memory.

## Scheduling And Recovery

The first scheduler can be simple and deterministic. It should select from
ready work by priority and age, skip tasks with unmet dependencies, and prevent
concurrent leases for overlapping write sets. Project and global concurrency
caps provide backpressure.

Leases need an owner, generation, acquisition time, heartbeat deadline, and
maximum duration. On expiry, the runtime inspects the workspace and forge
before requeueing. It preserves usable evidence, rejects stale publication, and
requires a fresh lease generation. A worker may never renew or complete a lease
it no longer owns.

Retries are bounded and classified. Deterministic product or verification
failures return to implementation only while the attempt budget remains.
Ambiguous requirements, policy conflicts, repeated failures, protected-path
changes, and disputed feedback move the task to `parked` for human review.

## Verification Contract

Verification is split into explicit receipts:

- **worker receipt:** focused formatting, linting, type checking, or tests that
  are fast and relevant to the slice;
- **integration receipt:** checks after combined changes, migrations, generated
  output, or shared contracts are integrated;
- **CI receipt:** the repository's complete required build and test gates;
- **acceptance receipt:** evidence for the user-visible outcome.

Receipts identify the exact revision, command or check name, result, duration,
artifact reference, and expiry policy. A result for an older revision cannot
advance a newer candidate. Missing or inconclusive evidence remains unknown;
it is never recorded as success.

## Authority And Safety

The initial factory should be pull-request-only. Workers may create scoped local
commits and, when explicitly authorized, publish task branches and pull-request
updates. They may not merge, deploy, rewrite history, approve their own work,
or change the authority policy.

Additional invariants:

- one isolated workspace and one active lease generation per write task;
- no overlapping write ownership across active tasks;
- repository, user, and organization denials remain in force for every worker;
- workflow, release, dependency, security, policy, and agent-instruction files
  are protected unless a task explicitly targets them;
- network, credentials, external writes, and destructive actions use narrowly
  scoped service identities and approval grants;
- unanswerable approval requests in unattended work resolve to deny or park;
- all queues have concurrency, token, cost, attempt, and wall-time ceilings;
- logs and dashboards expose redacted operational metadata, not transcripts or
  secret-bearing payloads;
- a user can pause one task, one project, or the entire factory.

Branch protection and required human review remain the final integration gate.

## Human Interaction

The factory should ask for judgment only when automation lacks legitimate
authority or product context. Human input is required for:

- approving a task or batch plan before implementation;
- resolving ambiguous acceptance criteria or disputed review feedback;
- expanding scope, write paths, capabilities, or budgets;
- changing protected configuration or policy;
- accepting the final pull request and any deployment action.

Feedback received through an authenticated project channel can update the task
packet after validation. The update creates a new packet revision and invalidates
leases based on superseded acceptance criteria.

## Improvement Proposals

Workers may identify recurring friction in task decomposition, prompts,
policies, provider mappings, or verification. They may only produce a proposal
with cited run evidence and a measurable expected outcome.

A factory improvement follows the existing controlled lifecycle:

```text
observe -> aggregate -> propose -> replay -> human approve -> install -> monitor
```

The proposer cannot edit active factory policy, its evaluator, protected
fixtures, or approval records. Failed replay and monitored regression return to
proposal design.

## Observability

The operational view is a projection of durable events. It should show:

- ready, leased, verifying, review-ready, parked, and failed task counts;
- worker availability and lease expiry;
- queue age and throughput by project;
- retry, recovery, and verification-failure rates;
- elapsed time, token usage, and cost when providers expose them;
- pull-request links and the latest trusted evidence receipt;
- budget consumption and pause reasons.

The event log is the source of truth; dashboards and Markdown summaries can be
rebuilt from it.

## Initial Delivery Slice

The first usable version should stay narrow:

1. Define the provider-neutral task packet, lease, event, evidence receipt, and
   worker-result schemas.
2. Implement a Kaji runtime adapter for one forge, one CI system, and one coding
   harness.
3. Support one repository per factory run, two workers, isolated branches,
   bounded retries, and pull-request-only publication.
4. Install the worker and review contracts through an opt-in Maisternia preset.
5. Run a supervised pilot on allowlisted tasks and retain only redacted
   operational records.
6. Evaluate recovery correctness, duplicate work, reviewability, throughput,
   verification failures, cost, and operator interventions before broadening
   scope.

Multi-host scheduling, automatic merge, deployment, arbitrary repository
intake, unconstrained model routing, and autonomous policy changes are outside
the initial slice.

## Acceptance Contract

An implementation is ready for a supervised pilot when it can demonstrate:

- two workers cannot hold valid leases for the same task or overlapping writes;
- a killed worker's task becomes safely claimable without duplicate
  publication;
- every transition can be reconstructed from validated, append-only events;
- a worker can complete its task from the packet and repository without prior
  conversation history;
- stale verification cannot advance a changed revision;
- local, integration, CI, and acceptance evidence remain distinguishable;
- retries and all resource budgets stop at configured limits;
- external content cannot grant authority or override protected policy;
- workers cannot merge, deploy, self-approve, or mutate factory safeguards;
- pause, resume, and recovery preserve the task and evidence trail;
- Maisternia can preview, install, verify, and remove the declarative factory
  configuration without owning live runtime state.

## Decisions Required Before Implementation

- Confirm Kaji as the runtime owner and define the repository handoff between
  Kaji and AgentnykMaisternia.
- Choose the transactional lease and event store, its backup policy, and its
  single-host consistency guarantees.
- Select the first forge, CI system, and coding harness adapters.
- Define trusted identities and approval semantics for task intake, feedback,
  branch publication, and pull-request updates.
- Set pilot concurrency, retry, retention, token, cost, and wall-time limits.
- Decide how write-set conflicts are declared and detected before leasing.
- Define the minimum evidence required for a slice to become review-ready.

Until these decisions are explicitly accepted, this document remains a feature
shape and not an executable implementation plan.
