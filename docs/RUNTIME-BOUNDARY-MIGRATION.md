# Runtime-Boundary Migration

Maisternia no longer exposes the experimental workflow-runtime commands that
created and managed local task state.

## Removed Commands

The removed surfaces are:

```text
maisternia event ingest
maisternia task ...
maisternia work next ...
maisternia pipeline ...
maisternia source ...
maisternia grill ...
```

`maisternia event validate` remains read-only. Preset, collection, provider,
environment, hook, approval, render, plan, apply, and Admin configuration
features are unaffected.

## Existing Local Data

Older versions may have written task data under:

```text
~/.agent-workflow/
```

Upgrading does not inspect, migrate, modify, or delete that directory. It may
contain private notes, source locations, questions, decisions, or generated
artifacts, so automatic cleanup would be unsafe.

Before removing it manually:

1. inspect the directory for documents or notes you want to retain;
2. copy retained human-facing documents into an appropriate project or private
   archive;
3. confirm that no older Maisternia binary or skill still depends on it;
4. delete or archive it using your normal private-data process.

The directory is not configuration input for current Maisternia releases.

## Updating Installed Presets

Installed provider resources do not change until the updated catalog is
applied. When the presets were installed directly, preview each affected
preset:

```bash
maisternia preset plan --scope user idea-shaping
maisternia preset plan --scope user standard-work
```

When those resources are owned by a collection, update the owning collection
instead of applying a member preset independently:

```bash
maisternia collection plan \
  --scope user \
  --target codex \
  software-engineer

maisternia collection apply \
  --scope user \
  --target codex \
  --conflicts replace \
  --yes \
  software-engineer
```

A collection plan can report changed shared targets as conflicts because their
installed content differs from the new member resources. Confirm that every
replacement is one of the expected workflow files before selecting `replace`.
Do not select `keep`: that intentionally preserves the older runtime-coupled
skill. Maisternia backs up replaced files and preserves its normal ownership,
drift, and confirmation safeguards.

The updated `work-shape`, `work-source`, `work-grill`, `work-brainstorm`,
`work`, `work-run`, and `work-brief` resources use the current harness session
and explicitly supplied artifacts instead of Maisternia runtime state.

## Future Collaboration

Cross-session rooms, shared context, agent-to-agent discussion, live human
steering, and collaboration history belong to a dedicated collaboration
service. Maisternia may install connection settings or client skills for that
service when it exists, but it remains the configurator rather than the room or
task runtime.
