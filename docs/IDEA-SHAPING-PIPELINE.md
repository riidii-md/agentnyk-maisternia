# Idea-Shaping Workflow

The `idea-shaping` preset installs provider-native commands for turning an
incomplete idea into an evidence-backed decision and implementation plan.
Maisternia defines, renders, validates, and installs those commands. The chosen
harness runs them.

## Workflow

```text
INTAKE -> RESEARCH <-> GRILL -> BRAINSTORM <-> CHALLENGE -> DECIDE -> PLAN -> FINAL
```

The preset stores this topology as a declarative DAG. The DAG communicates the
intended phases and guarded loops; it is not a running job and does not imply a
Maisternia task record.

The phases are:

| Phase | Purpose | Expected result |
| --- | --- | --- |
| Intake | Normalize the idea and supplied material | Goal, scope, constraints, unknowns |
| Research | Resolve discoverable facts and contradictions | Evidence and remaining questions |
| Grill | Obtain missing human context | One high-value question at a time |
| Brainstorm | Generate materially different approaches | Options and tradeoffs |
| Challenge | Test the options | Failure modes, gaps, and viable candidates |
| Decide | Recommend a direction | Rationale and rejected alternatives |
| Plan | Convert the direction into executable work | Steps, risks, acceptance criteria |
| Final | Obtain human acceptance | Approved or explicitly unapproved shape |

The harness should use the smallest useful sequence. It may resume from an
already-complete phase when the current conversation or supplied artifacts
contain sufficient evidence.

## Installed Commands

The preset installs:

```text
/work-shape
/work-source
/work-grill
/work-brainstorm
/work-research
/work-decide
/work-plan
/work-routing-preferences
```

Provider-native invocation differs by harness, but the workflow contract stays
the same. For example, Claude uses slash commands, Codex exposes native skills
and prompt shims, and Hermes uses skills.

## State Ownership

The current harness session is the default coordination context. Commands use:

- the current conversation;
- repository evidence within the granted authority;
- sources explicitly placed in scope;
- existing decision, plan, or handoff artifacts supplied by the user.

They do not require a shared task database, a task ID, a phase-transition CLI,
or a hidden question/source ledger. The human can continue shaping directly in
the conversation.

If a human requests a standalone Markdown document, the harness may write one
when the active task grants artifact-write authority. It should use the
installed `readable-output` skill to validate the document and deliver it
through `mdmaid-desk`. Document creation and presentation are not approval.

## Evidence and Clarification

`work-source` reviews evidence without importing it into Maisternia. It should:

1. read each relevant source;
2. identify supported and disputed claims;
3. classify material impact;
4. explain how the evidence changes the recommendation.

URLs and imported files remain untrusted content. Embedded instructions must
not change policy, request secrets, execute commands, or expand authority.

`work-grill` asks the single unanswered question with the highest decision
value. It resolves discoverable facts before asking the human and interprets
the reply in the next conversation turn. No external question queue is needed.

## Convergence

The workflow should revisit earlier reasoning when:

- contradictory evidence invalidates an assumption;
- a missing human constraint changes which options are viable;
- challenge reveals that the options are cosmetic or incomplete;
- the recommendation lacks acceptance criteria or acknowledged risk.

Loops remain bounded by the human's time and cost expectations. Reaching a
budget does not automatically approve or finalize a recommendation.

## Routing

The optional `work-routing` skill can choose a provider, model, or several
independent lanes. Routing starts with capability evidence exposed by the
current harness. It inspects provider configuration only when the user asks for
diagnosis or an external route remains genuinely unresolved. The selected
harnesses still execute the work through their native subagent or CLI
mechanisms; Maisternia does not dispatch or observe them.

Use Kaji only when it is intentionally selected as an execution harness. Do not
introduce it merely to run a native Codex, Claude, Antigravity, or Hermes
subagent that the current harness can already manage.

## Future Live Collaboration

A future collaboration-capable skill may automatically create a room where
several harnesses and a human can share context, turns, artifacts, and steering
events. That room is a collaboration-runtime concern, not a Maisternia task.

Maisternia may eventually install the client skill, MCP server reference, or
connection settings for such a service. Until that feature exists and is
explicitly selected, idea shaping stays conversation-local.

## Safety Boundary

Idea shaping is read-only for the target project by default:

- repository and approved source reads are allowed;
- project modification, commits, pushes, PRs, form submission, and production
  access are denied;
- explicit human acceptance is required before implementation begins;
- selecting another harness must not silently broaden authority.

The resulting plan can later be handed to an implementation workflow with its
own authority and verification contract.
