# Interactive Pipeline Studio Proposal

## Status

This document proposes a behavior-changing interactive editor for the workflow
pipeline DAGs stored in Maisternia presets. It is an implementation plan, not an
implementation. Merging this proposal does not enable editing, change rendered
provider configuration, or apply configuration to a user or project scope.

The intended product sequence is:

```text
edit graph
  -> validate draft
  -> review source and generated-configuration changes
  -> confirm catalog save
  -> ask whether to apply the updated preset
  -> choose one or more harnesses
  -> choose user-global or project scope
  -> review plan and conflicts
  -> confirm apply
  -> use the updated workflow inside the selected harness
```

Saving and applying remain separate decisions. Saving a pipeline must never
silently write provider configuration.

## Problem

Maisternia already stores pipeline topology as `entry_phases`, `phases`, and
`edges`, validates normal-edge acyclicity and explicit loops, and displays a
text summary in Admin. Structured DAG editing is still file-based.

There are two gaps:

1. The current Admin presentation renders the ordered `phases` list as a chain
   and lists conditional and loop edges separately. It does not provide a true
   interactive view of branches, joins, multiple entries, or reconnectable
   edges.
2. Preset selection currently chooses static manifest resources, and rendering
   copies their source files. Changing only the JSON DAG therefore does not
   change the provider-native command or skill that a harness executes.

An editor that changes only the diagram would imply behavior it does not
deliver. The graph must participate in provider rendering before editing can be
presented as behavior-changing configuration.

## Discovered Repository Rules

Implementation must preserve these repository contracts:

- The Go standard library is preferred unless a reviewed dependency removes
  substantial complexity.
- Tests precede implementation changes.
- Manifest and path handling remain defensive.
- Preset and provider fixtures must not contain credentials, tokens,
  transcripts, runtime databases, or real user configuration.
- `apply` remains opt-in.
- Path, symlink, conflict, drift, backup, stale-plan, and confirmation checks
  must not be weakened.
- Provider homes contain mixed runtime and declarative state; Maisternia manages
  individual declared files, never whole provider directories.
- Maisternia defines, renders, validates, previews, and installs configuration.
  Harnesses retain ownership of execution, live phase state, approvals,
  histories, and runtime loops.
- Pull requests describe behavior, tests, commands run, security implications,
  migration impact, and remaining limitations.

Required verification remains:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go build ./cmd/maisternia
go run ./cmd/maisternia doctor
go run ./cmd/maisternia render --target all --output ./build/rendered
```

## Goals

The completed feature should let a catalog author:

1. Open the real topology of a selected preset pipeline in Admin.
2. Navigate the graph entirely by keyboard, with mouse interaction when the
   terminal supports it.
3. Reconnect phases, add or remove edges, edit conditions, mark explicit loops,
   and change entry phases in an in-memory draft.
4. See validation failures immediately and understand which graph rule failed.
5. Undo and redo draft mutations.
6. Review the exact preset change and generated controller change before save.
7. Save only to an eligible source catalog with stale-source detection,
   recoverable backup, and atomic replacement.
8. Choose whether to apply the saved preset immediately or later.
9. Reuse the existing target, scope, project, planning, conflict-resolution, and
   confirmation stages when applying immediately.
10. Cause an accepted topology change to alter the controller instructions
    installed for every selected supported harness.

## Explicit Exclusions

The first release will not:

- execute or supervise a pipeline;
- display live nodes, phase progress, or harness session state;
- infer or dispatch the next runtime phase outside generated instructions;
- silently save or apply after a drag gesture;
- directly edit provider files;
- synchronize whole provider directories;
- edit immutable embedded catalogs or cached external-source snapshots;
- persist canvas positions inside provider-neutral preset JSON;
- add, remove, or rename behavioral phases;
- edit the detailed instructions for a phase;
- guarantee deterministic runtime transition evaluation by a language model;
- introduce a Rust sidecar, a second release binary, or a Maisternia workflow
  daemon.

Phase creation is deferred because the current schema names a phase but does not
associate that name with complete executable behavior. The first slice edits
topology among an already-defined phase set. A later structured phase-content
model can safely unlock node creation and renaming.

## Product Decisions

### The Graph Becomes A Render Input

Each behavior-changing pipeline should identify one managed controller resource,
for example the command or skill that coordinates the pipeline. A proposed
backward-compatible field is:

```json
{
  "id": "shape",
  "name": "Shape",
  "controller_resource": "work-shape",
  "entry_phases": ["intake"],
  "phases": ["intake", "research", "grill", "brainstorm", "challenge", "decide", "plan", "final"],
  "edges": []
}
```

Pipelines without `controller_resource` remain valid legacy declarative graphs,
but Admin must not claim that editing them changes harness behavior. The shipped
behavior-changing presets should be migrated before their editors are enabled.

The controller resource must:

- be included in the preset's managed command or skill contents;
- have targets for every provider declared by the preset;
- be uniquely owned as a pipeline controller so different presets cannot
  require different generated bytes at the same managed destination;
- contain exactly one validated pipeline-contract insertion marker;
- remain within the existing manifest size, path, target-root, and symlink
  boundaries.

### Compile Instead Of Duplicating Source

The preset JSON remains the source of truth. Saving the editor must not rewrite
the same adjacency data into a checked-in Markdown controller file.

During preset selection, planning, rendering, and apply, a deterministic
compiler should:

1. load the selected static manifest resources;
2. validate the pipeline and its controller binding;
3. render one normalized workflow-contract block from the pipeline;
4. replace the controller's exact insertion marker in memory;
5. retain all other controller bytes;
6. checksum and plan the compiled content;
7. pass those exact bytes to staging render or guarded apply.

The generated block controls sequencing. The surrounding controller source
continues to define what each existing phase does.

Conceptually, the generated block is:

```text
GENERATED WORKFLOW CONTRACT

Entry phases:
- intake

Transitions:
- intake -> research
- research -> grill
- grill -> brainstorm
- grill -> research when "evidence gap" [explicit loop]

Rules:
- The graph is authoritative for sequencing.
- Existing phase definitions are authoritative for phase behavior.
- Follow a conditional edge only when its condition applies.
- A loop marker describes a permitted return edge; it does not grant unbounded
  continuation, additional authority, or permission to exceed a human budget.
- If no valid transition applies, stop and ask the human.
```

Conditions are rendered as bounded data labels. Validation should reject
control characters, unsafe formatting characters already rejected elsewhere,
empty normalized conditions when a condition is present, and size violations.
The generated preview remains the final defense against unintended instruction
changes.

### Compile Into A Managed Resource Bundle

The existing configurator plans from source paths. Generated controller bytes
need a first-class, bounded representation rather than temporary files with
unclear lifetimes.

A refactor should introduce an internal desired-resource bundle containing:

- resource identity;
- original source description;
- selected provider targets;
- validated content bytes;
- content checksum;
- whether content was copied or generated.

Static resources enter the bundle unchanged. A pipeline controller enters with
the compiled bytes. Render and plan/apply consume the same bundle so previewed
and installed bytes cannot diverge.

Generated content remains subject to the same maximum size, target-root,
symlink, checksum, conflict, backup, ownership, and stale-plan checks as static
content. The representation is internal runtime data and is not written to a
temporary source tree.

### Save Then Offer Standard Apply

The editor has two explicit transitions:

```text
Draft --confirm save--> Source catalog
Source catalog --confirm apply--> Selected provider targets
```

After a successful source save, Admin asks:

```text
Preset saved. Apply the updated configuration now?

[Review apply] [Later]
```

`Review apply` opens the existing guarded installer:

1. choose one, several, or all supported harnesses;
2. choose user-global or project scope;
3. enter or confirm the project root when needed;
4. build the scoped plan from the exact saved preset digest;
5. review create, update, unchanged, removal, release, and conflict actions;
6. choose keep or replace only for explicit conflicts;
7. confirm apply.

`Later`, cancellation, or apply failure leaves the source save intact and
provider configuration untouched or unchanged by that failed attempt. Admin
must display `saved, not applied` rather than implying rollback or success.

## Interaction Design

### Entry And Modes

The Presets view should distinguish these actions:

- `Enter`: inspect managed resource source, as today;
- `g`: open the selected pipeline graph;
- `e`: enter graph edit mode when the selected pipeline and source are eligible;
- `i`: open the standard preset install flow, as today.

The graph view starts read-only. Edit mode is visually unmistakable and shows
one of `clean`, `modified`, `invalid`, `saving`, `saved`, or `save failed`.

### Canvas

The canvas should provide:

- deterministic layered layout based on non-loop edges;
- multiple entry nodes;
- branch and join rendering;
- visually distinct normal, conditional, and explicit-loop edges;
- pan, fit-to-view, and center-on-selection;
- keyboard node and edge traversal;
- mouse selection and dragging where supported;
- a textual topology fallback at the existing minimum terminal size;
- stable ordering derived from existing phase and edge order to avoid visual
  churn.

Loop edges should not participate in topological ranking. They should route in a
visually distinct lane and retain their condition label.

Node dragging changes only the current canvas. The initial release recalculates
layout when the view opens and does not persist coordinates.

### Editing

The topology editor should support focused transactions:

- connect selected source and target phases;
- reconnect one selected edge;
- delete one selected edge;
- edit an edge condition;
- toggle explicit-loop status;
- toggle an existing phase as an entry;
- undo or redo one completed mutation;
- discard the complete draft after explicit confirmation when modified.

The data model permits multiple edges between the same phases when condition or
loop attributes differ. Editor state therefore needs ephemeral stable edge IDs;
it must not identify an edge only by its endpoints.

The draft may temporarily be invalid while a transaction is incomplete, but
save remains disabled. A normal edge that introduces a cycle should be rejected
or offered as an explicit loop rather than silently changing its meaning.

### Review And Save

Save review shows:

- a semantic summary of entry and edge changes;
- the exact preset JSON change;
- the generated controller block change;
- affected managed controller targets;
- validation status;
- the source path and original digest.

The save operation must compare the current source digest with the digest loaded
when editing began. A mismatch stops the save and offers reload; it must not
merge or overwrite unseen changes.

Before replacement, preserve the previous regular file in Maisternia's private
backup area rather than adding backup artifacts to the source repository. Then
write a synchronized temporary file, preserve the intended mode, and rename it
within the validated preset directory. Existing symlink and containment checks
remain mandatory.

## Source Eligibility

Graph viewing is available for primary, embedded, directory-source, and GitHub
source presets.

Editing is initially available only for a primary source catalog selected from
a real checkout or explicit repository path. It is unavailable for:

- the content-addressed embedded catalog;
- cached snapshots of local external sources;
- cached GitHub sources;
- collections, because their pipelines are synthetic;
- paths that fail regular-file, directory, containment, permission, or symlink
  checks.

The UI should explain the reason and show how to relaunch with `--repo` for a
source checkout. A later `Copy to local catalog` workflow can make external
presets authorable without weakening snapshot immutability.

## Options Considered

### 1. Save The DAG And Reuse Apply Unchanged

Rejected. The current selected manifest contains only static resource sources,
so provider output would not change when an edge changes.

### 2. Rewrite Controller Markdown During Source Save

Rejected. It creates two checked-in representations of the graph and introduces
a multi-file consistency problem between preset JSON and controller Markdown.

### 3. Compile The Pipeline Into Its Controller

Selected. It keeps one graph source of truth, makes provider plans observable,
uses existing managed targets, and retains the standard apply boundary.

### 4. Generate Different Per-Phase Wrappers

Rejected for the initial slice. Rewriting every phase creates noisy plans,
duplicates graph instructions, and risks ownership conflicts when phase
resources are shared by presets.

### 5. Add A Runtime Engine Or Rust Sidecar

Rejected. A runtime engine would cross the product boundary, while a Rust
sidecar would add a second toolchain and release artifact to a self-contained Go
binary. The `rataflow` project remains a useful interaction reference and
prototype target, not a production dependency decision.

## Ordered Implementation Plan

Each task below should leave the repository runnable and should begin with tests
that fail for the intended reason.

### Task 1: Prove One Thin End-To-End Compiled Slice

Start with one shipped pipeline, preferably `idea-shaping/shape`, without TUI
editing.

1. Add a controller binding to the fixture and an exact marker to its controller
   source.
2. Compile its current graph into deterministic controller bytes.
3. Build a preset plan from those bytes.
4. Prove that changing one edge in a test changes the controller checksum and
   produces an update action for a selected provider.
5. Prove that the static controller source and unrelated resources remain
   unchanged.

This slice validates the central behavior claim before graph-rendering effort is
committed.

Likely files and patterns:

- `internal/presets/types.go`
- `internal/presets/validate.go`
- `internal/presets/manifest.go`
- `internal/configurator/types.go`
- `internal/configurator/plan.go`
- `internal/configurator/render.go`
- `config/schema/preset.schema.json`
- `config/presets/idea-shaping.json`
- `config/workflow/phases/shape.md`

### Task 2: Define Controller Binding And Catalog Validation

1. Add the optional `controller_resource` pipeline field to Go types and JSON
   schema.
2. Require the controller to be a managed command or skill in the same preset.
3. Require selected targets to cover every preset provider.
4. Reject missing, repeated, unsupported, or multiply owned controller bindings.
5. Require exactly one insertion marker in the base controller source.
6. Strengthen condition validation for generated-instruction safety.
7. Add catalog-wide tests for shipped and external legacy presets.

Legacy pipelines without a controller continue to load and render as
declarative-only pipelines. Behavior editing remains disabled for them.

### Task 3: Introduce The Desired-Resource Bundle

1. Separate manifest metadata from the bytes that plan/render/apply consume.
2. Read static resource bytes once into bounded desired resources.
3. Carry generated controller bytes in the same representation.
4. Preserve source descriptions for plans and diagnostics.
5. Refactor checksum, plan, render, and apply paths without altering current
   action classification.
6. Add regression tests for create, unchanged, update, ignored, conflict,
   release, and removal actions.
7. Re-run symlink, path escape, non-regular file, file-size, stale plan, backup,
   and shared-ownership tests against both static and generated content.

No generated content should be materialized in a repository or provider home
before confirmed render/apply.

### Task 4: Complete Deterministic Pipeline Compilation

1. Define a stable generated-contract format.
2. Preserve phase, entry, and edge order as stable tie-breakers.
3. Encode multiple entries, normal transitions, conditions, and loops.
4. State sequencing, behavior, authority, budget, and stop semantics explicitly.
5. Reject absent or repeated markers and oversized generated resources.
6. Ensure compilation is byte-identical for identical inputs.
7. Add golden tests without user data or provider secrets.
8. Expose compiled content through existing CLI render and Admin resource
   preview paths.

### Task 5: Migrate Shipped Behavior-Changing Pipelines

Audit each shipped pipeline and select one unique existing controller resource.
Candidate mappings require verification against actual source behavior:

| Preset | Pipeline | Candidate controller |
| --- | --- | --- |
| `standard-work` | `delivery` | `work-conductor` |
| `idea-shaping` | `shape` | `work-shape` |
| `parallel-work` | `speed-loop` | `work-speed-loop` |
| `scored-experiment` | `improve` | `work-experiment` |
| `adaptive-readability` | `reader-adaptation` | `work-adapt-for-reader` |
| `multi-lens-review` | `review-loop` | `multi-lens-review-skill` or a unique command |
| `harness-profile` | `profile` | `work-profile` |
| `session-audit` | `audit` | `work-audit` |
| `harness-improvement` | `improve-harness` | `work-retrospective` |

For every mapping:

1. verify that the base controller defines behavior for the existing phase set;
2. add one marker;
3. prove the resource is not a conflicting controller for another preset;
4. validate provider target coverage;
5. render all supported providers and review the exact generated files.

If a unique controller does not exist, leave that pipeline declarative-only and
stop its editor enablement rather than inventing unsafe behavior.

### Task 6: Build A Pure Graph Projection And Layout Package

1. Project `presets.Pipeline` into view nodes and ephemeral edge IDs.
2. Rank nodes using non-loop edges.
3. Order each layer deterministically with existing schema order as a stable
   fallback;
4. Route normal, conditional, and loop edges distinctly;
5. Clip to a bounded terminal canvas;
6. Support pan, fit, selection, and resize;
7. Render a textual fallback for insufficient dimensions.

Keep layout independent of Bubble Tea messages so topology and rendering can be
unit tested without a terminal.

Likely files and patterns:

- new focused package under `internal/admin/` or `internal/pipelineui/`;
- `internal/admin/render.go`;
- `internal/admin/model.go`;
- existing width, truncation, crop, and style helpers.

Do not add a layout dependency until a focused prototype demonstrates that a
small standard-library layered layout is insufficient and the dependency has
been reviewed.

### Task 7: Add The Read-Only Graph View

1. Open graph mode from a selected preset and pipeline.
2. Render actual edges instead of deriving a chain from the phase list.
3. Add keyboard node/edge traversal and an inspector.
4. Add optional mouse selection, dragging, pan, and zoom without making mouse
   input mandatory.
5. Preserve existing resource preview and install key paths.
6. Show source eligibility and behavior-binding status.
7. Test small terminals, resizing, multiple pipelines, empty pipelines,
   branches, joins, multiple entries, loops, and duplicate endpoint pairs.

This task is useful independently and provides a safe UI integration point
before writes are introduced.

### Task 8: Add The In-Memory Topology Editor

1. Model editor state separately from the loaded snapshot.
2. Implement focused semantic mutations for entries and edges.
3. Assign ephemeral stable IDs to edges;
4. Validate after every completed mutation;
5. Keep invalid intermediate state visible but unsavable;
6. Add bounded undo and redo snapshots;
7. Confirm discard when leaving a modified draft;
8. Keep drag-only positions outside semantic undo history.

Tests should exercise every mutation through both direct model methods and
Bubble Tea messages. No editor action writes a file.

### Task 9: Add Safe Source Persistence

1. Make authoring eligibility explicit in the Admin snapshot.
2. Capture the exact source digest when editing starts.
3. Produce semantic, JSON, and compiled-controller previews;
4. Compare the current source digest before save;
5. Refuse stale, missing, symlinked, non-regular, oversized, or escaped paths;
6. Back up the previous source file in a private recoverable location;
7. Write and synchronize a temporary file in the validated directory;
8. Atomically replace the original and refresh the snapshot;
9. Report `saved` only after the refreshed source validates and compiles.

Persistence should be exposed through a typed callback in `admin.RunOptions`,
parallel to existing plan/apply callbacks, so the UI remains testable.

### Task 10: Connect Save To Standard Guarded Apply

1. After save, present `Review apply` and `Later`.
2. On review, reuse target selection, user/project scope, project input, plan,
   conflict selection, final confirmation, and completion states.
3. Bind the plan request to the saved preset and compiled bundle digest.
4. Rebuild and compare before apply; reject a stale saved revision.
5. Keep the saved catalog edit when apply is canceled or fails.
6. Refresh provider and plan state after successful apply.
7. Distinguish `saved, not applied`, `saved, apply failed`, and
   `saved and applied` in UI and tests.

No new shortcut may bypass the existing final confirmation.

### Task 11: Documentation, Migration, And Release Verification

1. Document graph/view/edit keys and source eligibility in `docs/ADMIN.md`.
2. Document controller binding and compilation in `docs/PRESETS.md` and
   `docs/CONFIGURATOR.md`.
3. Update `docs/WORKFLOW.md` and `docs/CONFIGURATION-BOUNDARY.md` to explain that
   graphs compile into harness instructions but do not create runtime state.
4. Add schema examples and external-source compatibility notes.
5. Render all shipped presets for all supported providers and review changed
   controller files.
6. Run the complete required verification suite.
7. Perform an independent safety review of generated instruction content,
   provider target coverage, path handling, save/apply separation, and
   migration behavior.

## Test Plan

### Unit Tests

- schema decoding and encoding with and without `controller_resource`;
- controller membership, uniqueness, target coverage, and marker validation;
- control-character and maximum-length rejection;
- deterministic contract generation;
- multiple entries and branches;
- same endpoints with different conditions;
- explicit loop rendering and normal-cycle rejection;
- missing, repeated, and oversized insertion markers;
- static and generated resource checksum parity;
- graph projection, ranking, routing, clipping, and stable ordering;
- editor mutations, validation, undo, redo, and discard;
- stale source digest and atomic persistence errors.

### Integration Tests

- a graph edge change produces a provider controller update plan;
- unrelated resources remain unchanged;
- render output contains the reviewed generated contract;
- apply writes only selected provider targets;
- user and project scopes retain independent ownership state;
- apply-later leaves provider targets unchanged;
- canceled and failed apply preserve the saved source edit;
- conflict abort, keep, and replace retain current behavior;
- stale source and stale plan both fail closed;
- embedded and external cached sources remain read-only;
- multiple selected providers receive one aggregated reviewed plan.

### TUI Model Tests

- graph mode does not collide with resource-preview or install keys;
- keyboard-only navigation reaches every node, edge, form control, and action;
- mouse messages are ignored safely outside graph mode;
- window resize keeps selection and produces a valid fallback;
- invalid drafts cannot enter save confirmation;
- save success opens the apply invitation, not apply execution;
- `Later` closes cleanly with a visible unapplied state;
- apply invitation reuses the existing target/scope stages.

### Regression Tests

- all existing preset fixtures without controller bindings still load;
- all current static resource plan states remain unchanged;
- catalog, provider, environment, hook, approval, collection, and routing tests
  remain green;
- release builds remain one CGO-disabled Go binary for macOS, Linux, and
  Windows.

## Risk And Blast-Radius Review

### Generated Content Changes Plan Semantics

The configurator currently treats source files as desired bytes. Adding
generated bytes touches checksum, render, plan, apply, ownership, backup, and
staleness code. This is the largest blast radius and is why the first thin slice
must prove one controller update before TUI work.

### Model-Interpreted Conditions Are Not A Runtime Engine

Generated graph instructions change harness behavior, but natural-language
conditions remain model interpreted. Documentation and UI must not claim
deterministic transition enforcement or live pipeline control.

### Controller Resource Collisions

If two presets compile different graphs into one shared target, ownership and
checksum conflict. Catalog validation must reject ambiguous controller bindings
before render or apply.

### Phase Behavior Drift

Reconnecting existing phases is meaningful only if the controller defines those
phases. Adding new phases is excluded until phase behavior becomes structured
and independently validatable.

### Prompt And Formatting Injection

Conditions become part of installed instructions. Bound, normalize, quote, and
preview them. Treat external preset text as untrusted during inspection and do
not let it expand installation authority.

### Stale Source Between Save And Apply

The source may change after graph save. Bind apply review to the saved preset
and compiled bundle digest, then fail closed when either changes.

### Terminal Portability

Mouse protocols and glyph widths vary. Keyboard behavior and a text fallback are
requirements, not degraded afterthoughts. Keep layout math independent from
styled string widths and test Unicode clipping.

### Apply Failure After Save

Source authoring and provider installation are deliberately separate. Do not
roll back a valid source save merely because one selected provider could not be
updated. Report the exact partial or failed installation state using existing
plan/apply evidence.

## Migration And Rollout

1. Add optional schema support without changing existing preset behavior.
2. Land compiler and desired-resource refactor behind pipelines that explicitly
   declare a controller.
3. Migrate and verify one shipped preset.
4. Enable its read-only graph view and compare compiled render output in CI.
5. Migrate remaining presets only after controller uniqueness and phase
   behavior are verified.
6. Enable topology editing only for compiled, source-eligible pipelines.
7. Keep legacy and external declarative pipelines view-only.
8. Add node/content authoring in a separate schema proposal.

Existing installations are not mutated by migration. Their next explicit
preset plan should show controller updates caused by generated graph contracts.
Users retain the normal choice to apply now or later and to select the same or a
different supported scope.

## Verification Gates

### Gate 1: Compiler Proof

- one edge change produces a deterministic controller byte change;
- selected provider plan reports the intended update;
- no unrelated resource changes.

### Gate 2: Configurator Safety

- static and generated resources pass identical path, size, conflict, backup,
  ownership, and staleness tests;
- full existing configurator tests remain green.

### Gate 3: Read-Only Graph UX

- real topology is visible and keyboard navigable;
- no write callback exists in graph-view mode;
- small terminal fallback works.

### Gate 4: Draft Editing

- all mutations are in memory;
- invalid graphs cannot save;
- undo, redo, discard, and reload behave deterministically.

### Gate 5: Save And Apply Boundary

- source save requires its own confirmation;
- apply invitation does not execute apply;
- target, scope, plan, conflicts, and final confirmation are preserved;
- cancel and failure states are accurate.

### Gate 6: Release Readiness

- all shipped preset/controller mappings are reviewed;
- provider render outputs are reviewed for all supported targets;
- required repository verification passes;
- security review confirms no widened authority or whole-directory ownership.

## Stop Conditions

Stop and return to design if:

- a graph cannot change provider behavior without a runtime daemon;
- a provider requires bypassing its managed target or native configuration
  contract;
- a controller resource is shared in a way that cannot preserve ownership;
- generated content cannot pass the same plan, conflict, backup, and staleness
  protections as static resources;
- the implementation requires whole-directory synchronization;
- safe editing requires creating phases whose behavior is not represented;
- a save or apply path would need implicit confirmation;
- applying the reviewed bytes cannot be bound to an exact digest;
- a dependency is needed but its complexity, maintenance, license, and release
  implications have not been reviewed;
- existing external preset compatibility would be broken without an explicit
  migration strategy.

## Inputs For `/work-prove`

A proof contract should map these requirements to observable evidence:

| Requirement | Required evidence |
| --- | --- |
| Graph changes behavior | Edge mutation changes compiled controller bytes and provider plan |
| One source of truth | Controller source contains a marker, not duplicated adjacency data |
| No silent apply | Saving alone leaves provider checksum and ownership state unchanged |
| Standard apply flow | Model tests traverse harness, scope, plan, conflict, and confirmation stages |
| Safe invalid state | Invalid draft is visible but cannot save or apply |
| Exact revision | Stale source and stale compiled bundle are rejected |
| Defensive paths | Generated and static resources pass the same escape and symlink tests |
| Recoverability | Source save and provider replacement both retain recoverable prior bytes |
| Provider coverage | Every declared target receives a reviewed compiled controller or is rejected |
| Runtime boundary | No task state, phase progress, dispatch, or session observation is introduced |
| Accessibility | Keyboard-only graph editing and small-terminal fallback are tested |
| Compatibility | Legacy declarative pipelines remain loadable and view-only |

`/work-prove` should also define the exact golden fixtures, expected provider
target paths, negative security cases, and acceptable generated-output review
process before implementation begins.

## Remaining Deliberate Limitation

The initial editor changes topology among existing behavioral phases. Full node
creation requires a separate design for structured phase instructions and their
provider-neutral compilation. Keeping that boundary explicit prevents the UI
from creating nodes that look executable but have no defined behavior.
