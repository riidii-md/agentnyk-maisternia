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

Use Mermaid only where it materially improves understanding. Render through the
configured readable-output tool when available, but preserve the Markdown path
when rendering is unavailable.

Do not edit repository files, change external state, include secrets or private
environment values, or dump sensitive raw logs.
