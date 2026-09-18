---
name: decision-capture
description: Capture a technical, product, process, risk, or business decision as durable Markdown without interrupting the user with an unsolicited browser page.
---

# Decision Capture

Use this skill when a decision and its rationale should survive the current
conversation.

## Capture The Decision

Record:

- the decision and current context;
- the chosen direction;
- options considered and why they were rejected;
- accepted risks and unresolved questions;
- follow-up actions and owners when known.

Inspect the repository for an ADR directory, decision log, or documentation
convention. Do not write into those project-owned locations without the user's
request or approval. Otherwise keep the durable Markdown under:

```text
.agent-runs/decisions/<timestamp>-<semantic-slug>.md
```

Use the installed `readable-output` skill when the user asks to publish the
decision to mdmaid.desk or an active workflow already requires desk delivery.
The Markdown file remains the canonical artifact.

## Keep Browser Presentation Opt-In

Do not generate HTML, temporary or otherwise, as a side effect of capturing a
decision. Do not invoke `codex-readable-doc`. Do not open a browser, local file,
preview, TUI, or persistent reader unless the user explicitly asks for that
presentation in the current request. A request to capture or persist a decision
does not imply a request to open it.

When the user explicitly asks for a browser preview, treat that as a separate
optional presentation step after the durable Markdown exists. Never make HTML
the only copy of the decision.

## Return

Return the key decision, durable Markdown path, suggested repository
persistence target when one exists, and whether approval is required before
updating project documentation. Do not duplicate the full decision in chat
unless requested.
