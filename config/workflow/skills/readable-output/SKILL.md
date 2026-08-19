---
name: readable-output
description: Publish a response, plan, review, analysis, research report, or command output as durable Markdown in mdmaid.desk. Use when output would be uncomfortable to read in the terminal; when a decision, recommendation, migration, architecture choice, or tradeoff needs careful reading; or when the user asks to read, view, add, send, push, publish, or open a document in mdmaid-desk, mdmaid.desk, the desk, or the central reading hub. Do not trigger for a brief answer unless the user explicitly requests desk delivery.
---

# Deliver Readable Output

Treat mdmaid.desk as the canonical reading hub. `mdmaid` validates or renders a
document; `mdmaid-desk` catalogs it. Rendering Markdown to HTML, opening a local
file, or running a helper that only calls `mdmaid` is not desk delivery.

## Preserve the complete document

When the user supplies an existing Markdown artifact, deliver that artifact
without rewriting it unless adaptation was requested. Otherwise write the
complete standalone result under the current repository root, or the current
directory outside a repository, at:

```text
.agent-runs/readable-output/<timestamp>-<semantic-slug>.md
```

Use a semantic level-one heading and preserve evidence, caveats, source links,
and required detail. Never make a temporary HTML file the only durable result.

## Validate before delivery

Require mdmaid 0.1.17 or newer and check `mdmaid --version`. If it is missing or
older, preserve the Markdown artifact, do not deliver it, and report the exact
upgrade command:

```text
npm install --global mdmaid@0.1.17
```

With a compatible version, run:

```text
mdmaid validate <artifact.md> --json
```

Treat validation as a hard gate:

- Exit 0 permits delivery.
- Exit 1 means invalid content. Fix the source-located diagnostics and repeat
  validation until it succeeds.
- Exit 2 means the validation runtime is unavailable. Preserve the artifact,
  stop delivery, and report the blocker and exact retry command.

## Deliver to mdmaid.desk

Only `mdmaid-desk register` or `mdmaid-desk import` with exit 0 proves delivery.
Do not treat `mdmaid`, `codex-readable-doc`, a local browser open, temporary HTML,
or a likely URL as equivalent.

Resolve the intended workspace with `mdmaid-desk workspace list`:

1. Use `MDMAID_DESK_WORKSPACE` when explicitly configured.
2. Honor a workspace explicitly named by the user.
3. Otherwise match the canonical current repository root or current directory.
4. If the current root is absent and is the intended workspace, add it once
   with `mdmaid-desk workspace add`, using a stable collision-safe ID.

If the artifact is already inside the intended workspace or one of its allowed
artifact roots, run:

```text
mdmaid-desk register <artifact.md> --workspace <id> --kind <kind> --title "<title>" --attention review
```

If the artifact is outside the intended workspace and the user wants it stored
there, run:

```text
mdmaid-desk import <artifact.md> --workspace <id> --kind <kind> --title "<title>" --attention review
```

`review` is the default attention state. Honor an explicit workflow request for
`approval`, `failure`, or `changes_requested` when the artifact has that role.
Attention controls presentation priority only; even `approval` never records or
implies a human decision.

Prefer `decision`, `definition`, `progress`, `brief`, or `showcase` when the
document clearly matches one. Use the first level-one heading as the title,
without the Markdown marker. Add `--task <id>` only for an explicit stable task
ID and add up to three grounded subject tags when useful.

Registration is a presentation action, not approval. Do not start a persistent
server, daemon, TUI, or browser unless the user asks or an existing workflow
already requires it.

## Recover instead of substituting

If the desk command fails, preserve the artifact and diagnose the actual
failure. Check the executable, runtime version, workspace mapping, artifact
roots, permissions, and command syntax. A renderer is not a fallback for a
broken desk CLI.

For a runtime or native-module mismatch, identify the package's matching
runtime or the repository's documented rebuild/install command. For a sandbox
or state-directory denial, request the narrow permission needed for the same
desk command. Do not rebuild, reinstall, or change a workspace mapping without
the authority implied by the request.

If delivery still cannot complete, return the artifact path, concise failure,
and exact repair and retry commands. Do not claim the document was added,
published, sent, imported, registered, or opened in mdmaid.desk.

## Return a verifiable receipt

After exit 0, return a short terminal summary containing:

- the durable Markdown path;
- the selected desk workspace;
- whether `register` or `import` succeeded;
- the document ID or desk URL when the CLI provides one.

Do not duplicate the full document in chat unless requested.
