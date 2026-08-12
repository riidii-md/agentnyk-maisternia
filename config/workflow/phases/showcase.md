# /work-showcase - Create a Standalone Review Document

Produce a standalone Markdown report after substantial analysis, research,
planning, implementation, review, or a long conversation. The reader must not
need access to the original chat.

Input:

`$ARGUMENTS`

Read referenced local files and reports when provided. Separate verified facts
from assumptions and proposals. Explain technical findings in simple terms
without losing important constraints.

Include when relevant:

- executive summary;
- current status;
- problem and goal;
- important history;
- findings;
- decisions already made;
- proposed direction or plan;
- architecture or flow;
- risks and unknowns;
- review questions and approvals needed;
- next steps;
- sources and local files.

Use Mermaid only where it materially improves understanding.

Always write the complete report to a durable artifact at
`.agent-runs/showcase/<timestamp>-showcase.md`. Resolve the mdmaid.desk workspace
from `MDMAID_DESK_WORKSPACE` or by matching the canonical current project root
in `mdmaid-desk workspace list`. If the root is not registered, add it once with
a stable collision-safe workspace ID. Then run:

```text
mdmaid-desk register <artifact.md> --workspace <id> --kind showcase --attention review
```

Registration sends the document to mdmaid.desk but does not imply approval. Do
not replace the durable Markdown with temporary output or HTML. If the desk CLI
is unavailable or registration fails, preserve the Markdown artifact and report
its path, the failure, and an exact retry command. Return a short summary plus
the Markdown path and registration status instead of duplicating the full
showcase in the terminal.

Do not edit repository source or configuration files, change unrelated external
state, include secrets or private environment values, or dump sensitive raw
logs. Writing and registering the requested artifact is allowed.
