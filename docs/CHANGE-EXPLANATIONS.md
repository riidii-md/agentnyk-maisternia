# Change Narratives

`change-narrative` is an internal skill used by `/work-human-review` and
`/work-plan-review`. It writes the evidence-based explanation, selected code,
and diagrams. `/work-human-review` adds the complete native Git diff, source
fingerprint, mdmaid.desk publication, and revision-bound decision request in
one document. There is no separate explanation command.

The narrative distinguishes requested intent, verified behavior, and
inference. It covers the important behavior before and after the change,
changed abstractions, compatibility, operations, tests, uncertainty, and the
most useful source evidence. It selects only diagrams that explain a material
relationship or transition. Mermaid is the default; `animated-web` is an
explicit PR Lens option. Each diagram is grounded in inspected source and
validated before publication.

The skill can author Mermaid directly for architecture, class,
entity-relationship, state, sequence, requirement, and data-flow views. It
uses a traceability flowchart when a requirement diagram cannot render in the
selected terminal. It chooses exactly one presentation for a review and omits
diagrams that do not clarify the change.

The skill does not search for defects by itself. `/work-review` owns automated
finding review. `/work-human-review` includes the resulting findings when
available and always includes the complete diff, even when no diagram is
useful.

## Presentation preference

An explicit request takes precedence, followed by the project preference,
then the user preference. Store new choices at
`workflows.work-human-review.presentation`. Existing
`workflows.work-explain-change.presentation` values remain a fallback until
the user migrates them. `mermaid` is the default; `animated-web` selects PR
Lens, and `static-tui` is a legacy alias for `mermaid`. The preference affects
the diagram format, not the evidence or complete diff.

## Tools

The environment-only `change-explanation-tools` preset still owns the pinned
`@coldtea/pr-lens-cli` 0.2.0, mdmaid, and mdmaid-desk tools. It is separate
from provider command installation:

```bash
maisternia environment plan change-explanation
maisternia preset apply --yes change-explanation-tools
```

`/work-human-review` checks compatible versions, validates the final combined
Markdown, proves that its single fenced `diff` block matches the canonical
patch byte-for-byte, and registers that document in mdmaid.desk. Registration
creates a review request; it does not itself approve the change.
