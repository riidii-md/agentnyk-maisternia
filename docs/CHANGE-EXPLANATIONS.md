# Change Explanations

`/work-explain-change` turns a pull request, commit, revision range, or working
tree into a standalone explanation for someone who should not have to read the
whole diff.

It is deliberately separate from `/work-review`:

- change explanation answers what changed, why it matters, how the system is
  shaped, and how data moves;
- review searches for defects, risks, and missing evidence and may make a merge
  recommendation;
- human review gates request an explicit approval and remain a different
  mdmaid.desk operation.

The command is installed by `standard-work` as an on-demand capability. It is
not another mandatory phase in the delivery DAG.

## Output

Each run writes a self-contained bundle under:

```text
.agent-runs/change-explanations/<timestamp>-<change-id>/
  explanation.md
  graph.json
  rendered/
    drawn.graph.json
    manifest.json
    <content-addressed>.svg
```

The Markdown gives the quick summary, stated intent versus verified behavior,
before/after model, important abstractions and functions, selected short code
examples, compatibility and test notes, and an evidence index. A representation
gate chooses the smallest useful visual: pseudocode, call/component/file trees,
a diff-shaped sketch, selected code, or PR Lens. PR Lens adds an architecture
diagram and, only when order is meaningful, an animated data-flow diagram. A
small local change is not forced into an architecture graph. The prose remains
complete for readers who cannot see motion.

PR Lens is a visualization layer here, not a finding engine. The workflow
authors the graph from evidence already inspected by the active harness and
runs only local deterministic `pr-lens validate` and `pr-lens render` commands.
It does not call `pr-lens analyze`, publish an asset, or post a PR comment by
default. Those actions require a separate explicit request because `analyze`
contacts another configured model provider and publishing changes external
state.

## Reader adaptation

When the `adaptive-readability` preset is also installed, the command can apply
the `adapt-for-reader` profile to language, density, ordering, and conceptual
depth. The profile cannot change the selected evidence, uncertainty, or meaning
of the diff. Without a profile, the default is an engineering reader with about
five minutes and a high-level-first explanation.

## Local presentation

The `change-explanation` environment pack pins:

- `@coldtea/pr-lens-cli` 0.2.0;
- `mdmaid` 0.1.17;
- `mdmaid-desk` 0.1.12.

The environment-only `change-explanation-tools` preset owns this pack. Review
and install it separately from provider configuration:

```bash
maisternia environment plan change-explanation
maisternia preset apply --yes change-explanation-tools
```

Environment detection is command-presence based. An older installed command is
reported as present, so `/work-explain-change` also checks versions before use.
Upgrade explicitly when needed:

```bash
npm install --global @coldtea/pr-lens-cli@0.2.0
npm install --global mdmaid@0.1.17
npm install --global mdmaid-desk@0.1.12
```

After graph and Markdown validation, the command registers `explanation.md`
with mdmaid.desk. Version 0.1.12 is the minimum because it resolves registered,
workspace-local SVG image references through authenticated same-origin media
routes. The asset response has a restrictive sandbox content security policy;
remote images and arbitrary filesystem paths are not enabled.

Registration is presentation, not approval. If validation, rendering, version
checks, or registration fail, the workflow preserves the local bundle and
reports the exact blocker instead of claiming successful animated delivery.

## Future extensions

The visual vocabulary can later grow to responsive UI comparisons or focused
interactive HTML. That should not reuse the SVG route accidentally: HTML needs
its own sandbox, navigation, authentication, content-type, and lifecycle
design. Theme-aware light/dark PR Lens asset selection and a reduced-motion
alternative are also useful reader controls once mdmaid.desk exposes a safe
document-media contract for them.

## Examples

```text
/work-explain-change PR 42 for a product engineer with five minutes
/work-explain-change 8c0ffee -- focus on the new service boundary
/work-explain-change origin/main...HEAD -- reader: support lead
/work-explain-change working tree -- deep explanation
```
