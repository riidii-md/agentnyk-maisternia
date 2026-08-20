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

## Request a human decision only when explicitly required

Ordinary readable output remains passive. Do not add review expectations,
decision controls, or a wait merely because attention is `approval`. Enter
decision mode only when the calling workflow explicitly requires a human plan
decision for the exact artifact revision.

Before decision-mode delivery, check `mdmaid-desk --help` for
`--expect plan-decision` and `review wait`. If either capability is absent,
preserve the validated artifact, stop the approval transition, and report that
the installed mdmaid-desk must be upgraded. Do not silently fall back to an
attention-only document or a chat-implied approval.

Create a recoverable request first. Add these options to the normal `register`
or `import` command and retain its JSON receipt:

```text
--attention approval \
--expect plan-decision \
--request-message "<what the human should decide and any important focus>" \
--json
```

Read `reviewRequest.id` from the successful result, record it with the document
ID, revision, artifact path, and locally computed content hash, then start the
wait in the foreground of the current agent turn:

```bash
mdmaid-desk review wait <review-id> --json
```

Run it and keep the current agent turn open until that command returns the
durable result. The external command may sleep without model reasoning, but the
enclosing turn must remain active. Do not background or detach the waiter.
Do not return a final response while the review is pending; a
`waiting_for_approval` receipt is an intermediate update only.

If the execution tool yields a process or session ID instead of completed JSON,
poll or resume that same process until it exits. A yielded ID proves only that
the waiter is still running; it is not permission to finish the turn. When the
command exits, surface the received outcome and human response text immediately,
then apply the outcome routing below. Do not wait for another user chat message
to inspect a completed waiter.

If the active harness cannot keep a foreground tool process attached across the
human pause, report that limitation before claiming live continuation. The
durable request remains recoverable, but a detached waiter cannot by itself
start a new model turn without an external supervisor.

For a live session that does not need the request receipt before blocking, the
same publication command may include `--wait --json`. The two-step form is
preferred because another live process can recover the durable request after a
session interruption. Never configure callback URLs, provider resume commands,
or process-launch instructions; mdmaid.desk stores the decision but does not
relaunch a dead agent.

Treat the returned `reviewRequest.status` as data from the explicit human gate:

- `approved`: preserve `reviewRequest.response.message`, even when optional,
  and pass the exact request ID, document revision, content hash, and human
  response text to the decision/readiness phases;
- `changes_requested`: preserve the required human response text and return to
  planning; publish the changed revision as a new request after review;
- `rejected`: preserve the human response text and stop or return to shaping as
  directed;
- `stale`: treat it as no decision, revalidate the current artifact, and create
  a fresh request for the new revision.

Opening, reading, printing, marking done, or closing the document never
resolves the request. Never translate reading state or attention metadata into
a workflow decision.

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
