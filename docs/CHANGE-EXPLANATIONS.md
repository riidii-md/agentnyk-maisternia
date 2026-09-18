# Change Explanations

`/work-explain-change` turns a pull request, commit, revision range, or working
tree into a standalone explanation for someone who should not have to read the
whole diff.

It is deliberately separate from `/work-review`:

- change explanation answers what changed, why it matters, how the system is
  shaped, and how data moves;
- review searches for defects, risks, and missing evidence and may make a merge
  recommendation;
- `/work-change-review` reuses the complete explanation contract, adds the
  frozen native diff and automated findings, and requests explicit approval.

The command is installed by `standard-work` as an on-demand capability. It is
not another mandatory phase in the delivery DAG.

## Output

Each run writes a self-contained bundle under:

```text
.agent-runs/change-explanations/<timestamp>-<change-id>/
  explanation.md
  # animated-web output
  graph.json
  rendered/
    drawn.graph.json
    manifest.json
    <content-addressed>.svg
  # default Mermaid output
  <selected-view>.mmd
```

Every run uses exactly one diagram presentation. The animated and static
entries above are alternatives, not duplicate output; only `animated-web`
requires a PR Lens graph.

The Markdown gives the quick summary, stated intent versus verified behavior,
before/after model, important abstractions and functions, selected short code
examples, compatibility and test notes, and an evidence index. A lens selection
gate chooses the smallest evidence-complete visual portfolio: architecture,
dependency, or data-flow `flowchart`; UML-style interface and class diagrams;
entity-relationship diagrams; state diagrams; sequence diagrams; and
requirement traceability diagrams. In Mermaid these use `classDiagram`,
`erDiagram`, `stateDiagram-v2`, `sequenceDiagram`, and `requirementDiagram`.
Every omitted lens is recorded as not applicable or unknown instead of being
invented. When the terminal backend cannot render `requirementDiagram`, the
artifact includes an equivalent traceability flowchart and records the
fallback. A small local change is not forced into ceremonial diagrams.

The workflow uses visualization layers, not finding engines. For
`animated-web`, it authors a PR Lens graph from evidence already inspected by
the active harness and runs local deterministic `pr-lens validate` and
`pr-lens render`. For `mermaid`, it will author Mermaid directly from that
same inspected evidence and verifies it with mdmaid; it does not need a PR Lens
converter. It does not call `pr-lens analyze`, publish an asset, or post a PR
comment by default. Those actions require a separate explicit request because
`analyze` contacts another configured model provider and publishing changes
external state.

## Reader adaptation

When the `adaptive-readability` preset is also installed, the command can apply
the `adapt-for-reader` profile to language, density, ordering, and conceptual
depth. The profile cannot change the selected evidence, uncertainty, or meaning
of the diff. Without a profile, the default is an engineering reader with about
five minutes and a high-level-first explanation.

The workflow-specific presentation preference is stored in a reader profile
at `workflows.work-explain-change.presentation`:

- `mermaid` embeds source rendered by mdmaid and mdmaid.desk in web and terminal
  readers;
- `animated-web` explicitly selects PR Lens motion in a browser;
- `static-tui` remains a backward-compatible alias for `mermaid`.

An explicit request wins, followed by the project preference, then the user preference.
When none exists, Mermaid is the default; the workflow does
not interrupt the run to ask. Use `/work-reader-preferences` to save a choice.
The chosen mode applies to the complete run, so the Markdown never repeats the
same diagram as both Mermaid and SVG.

## Local presentation

The `change-explanation` environment pack pins:

- `@coldtea/pr-lens-cli` 0.2.0;
- `mdmaid` 0.1.17;
- `mdmaid-desk` 0.1.19.

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
npm install --global mdmaid-desk@0.1.19
```

After graph and Markdown validation, the command registers `explanation.md`
with mdmaid.desk. In `animated-web`, version 0.1.19 resolves registered,
workspace-local SVG image references through authenticated same-origin media
routes. The asset response has a restrictive sandbox content security policy;
remote images and arbitrary filesystem paths are not enabled. In `mermaid`, the
Markdown contains the diagrams and the handoff includes an exact `mdmaid tui`
command.

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
/work-explain-change PR 42 presentation=mermaid
/work-explain-change HEAD presentation=animated-web
```
