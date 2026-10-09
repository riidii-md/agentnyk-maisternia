---
name: readable-output
description: Publish a response, plan, review, analysis, research report, or command output as durable Markdown in mdmaid.desk. Use when output would be uncomfortable to read in the terminal; when a decision, recommendation, migration, architecture choice, or tradeoff needs careful reading; or when the user asks to read, view, add, send, push, publish, or open a document in mdmaid-desk, mdmaid.desk, the desk, or the central reading hub. Do not trigger for a brief answer unless the user explicitly requests desk delivery.
---

# Deliver Readable Output

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

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

Use a semantic level-one heading that includes the grounded ticket ID when
available, the task name, and the document purpose, such as `TASK-123 — Worker
scaling: implementation plan`. Without a ticket, start with the task name; if
none was supplied, derive it from the grounded subject. Preserve evidence,
caveats, source links, and required detail. Never make a temporary HTML file
the only durable result.

Prefer a few durable checkpoints over one document per workflow phase.
Discovery may share one evolving brief; direction, implementation plan, and
change review are separate roles. When a calling workflow supplies a stable
task-and-role path, update it for a new revision instead of creating a
timestamped sibling. A new content hash never inherits earlier approval.

## Document-bound human checkpoints

Apply this prerequisite when an actual document is the basis of a requested
human response, or before reporting that document as waiting for a response.
Ordinary clarification without a document does not require creating one.

1. Finalize the exact current revision: complete authorized wording, screening,
   validation, and the local content hash before delivery.
2. Register or import that revision successfully with `--json`. Require exit 0
   and a valid `schemaVersion: 1` receipt; a local artifact alone is insufficient.
3. Verify and retain the artifact path, operation, workspace, returned document ID,
   document revision, receipt, and local content hash. Recompute the local hash
   after publication and stop if the source changed. A prior revision's receipt
   does not cover changed content.
4. Only after successful delivery, present the registered document and request
   the response or report waiting. Use the existing safe-link outcome below.

For candidate plan/direction AI-review choices and document-based discovery
questions, register passively with `--attention review`. Do not add `--expect`
or a decision waiter for those choices. Retain existing conversational response
rules for parallel plans, improvement proposals, and preferences. Exact
direction/plan/change decisions keep their authenticated transport, revision
binding, and foreground quiet-wait contracts.

If validation, registration/import, or receipt verification fails, preserve
the artifact and report **delivery blocked** with the exact recovery command.
Do not report waiting for review, approval, or a response. Successful delivery
without a safe clickable URL is different: report `safe-link-unavailable`,
retain the receipt, and give the existing diagnostic recovery and local artifact
link. Missing link availability never waives successful desk delivery.

Registration is not approval. Attention, opening, reading, archiving, and elapsed
time do not record a decision or authorize AI review. Internal artifacts,
copied evidence, machine-readable records, and archived packages are not
response-dependent automatically. Space setup questions remain prerequisites;
preserve their existing authority and fallback rules.

## Validate before delivery

Require mdmaid 0.1.17 or newer and check `mdmaid --version`. If it is missing or
older, preserve the Markdown artifact, do not deliver it, and report the exact
upgrade command:

```text
npm install --global mdmaid@0.1.17
```

Within one live session, reuse a successful version and decision-capability
preflight while the executable path and version remain unchanged. Recheck
after a command failure or executable change; do not repeat the same discovery
for every checkpoint document.

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

Read and follow `references/project-naming.md` before resolving the workspace or
building the registration command. Require mdmaid-desk 0.1.19 or newer. Ground
the repository and Jira ID, let the producing AI supply only the minimal feature
text, and never substitute a branch name. If no grounded Jira ID exists, omit
both project fields.

Resolve the intended workspace with `mdmaid-desk workspace list`:

1. Use `MDMAID_DESK_WORKSPACE` when explicitly configured.
2. Honor a workspace explicitly named by the user.
3. Otherwise match the canonical current repository root or current directory.
4. If the current root is absent and is the intended workspace, add it once
   with `mdmaid-desk workspace add`, using a stable collision-safe ID plus the
   credential-free `--repository` and grounded `--repository-name` from the
   project-naming contract.

Reuse a previously verified workspace ID while the canonical root is unchanged
in the same live session. Resolve again after a workspace error or root change.
Never create a new workspace merely because an earlier receipt is unavailable.

Before choosing the register or import command, settle the document kind, title,
project metadata, and up to three grounded subject tags using the rules below.
After resolving the workspace and finalizing the exact planned tags, follow the
installed readable-output `references/space-routing.md` contract before
registration and run its scoped postcondition after successful registration.
Space routing never changes the attention or decision semantics in this skill.

If the artifact is already inside the intended workspace or one of its allowed
artifact roots, run:

```text
mdmaid-desk register <artifact.md> --workspace <id> --kind <kind> --title "<title>" --attention review --task <jira-id> --feature-name "<minimal feature text>" --json
```

If the artifact is outside the intended workspace and the user wants it stored
there, run:

```text
mdmaid-desk import <artifact.md> --workspace <id> --kind <kind> --title "<title>" --attention review --task <jira-id> --feature-name "<minimal feature text>" --json
```

Use `--json` for every `mdmaid-desk register` or `import` publication. Retain
the structured receipt without echoing it wholesale, require
`schemaVersion: 1`, and read `document.route`. A missing or invalid route blocks
link creation; never derive a route from the document ID.

`review` is the default attention state. Honor an explicit workflow request for
`approval`, `failure`, or `changes_requested` when the artifact has that role.
Attention controls presentation priority only; even `approval` never records or
implies a human decision.

Prefer `decision`, `definition`, `progress`, `brief`, or `showcase` when the
document clearly matches one. Always pass `--title` for a generated document.
Use the first level-one heading without the Markdown marker when it names the
task and document purpose. If the supplied document has a generic heading,
derive the catalog title from its grounded task context and purpose without
rewriting the document. Never use only a filename, timestamp, or kind such as
`change-review` as the title. Add the paired `--task` and `--feature-name`
options only for an explicit stable task ID, following the project-naming
reference, and add up to three grounded subject tags when useful.

Registration is a presentation action, not approval. Do not start a persistent
server, daemon, TUI, or browser unless the user explicitly asks in the current
request.

## Return a full clickable document URL

After successful registration or import, capture `mdmaid-desk daemon status`
internally to resolve the active web origin. Do not print, log, persist, quote,
or relay its raw output because the reported web URL contains an authentication
token. Exit nonzero, more or fewer than one `mdmaid.desk web:` line, or malformed
output means that no safe active origin is available. Do not read private daemon
state as a substitute.

Parse the captured web URL with a URL parser. Accept only `http` or `https`,
reject user information, and reduce it to its origin: scheme, host, and explicit
port when present. Discard its path, query, and fragment, including the token.
Require `document.route` to be a single origin-relative document path: it must
start with `/d/`, must not start with `//`, and must not contain a scheme,
authority, user information, query, or fragment. Resolve it against the
sanitized origin and verify that the result has the same origin and the same
pathname before presenting it.

Return the full absolute HTTP(S) mdmaid.desk document URL as a clickable
Markdown link, for example
`[Open in mdmaid.desk](https://mdmaid.desk.localhost/d/...)`. Never substitute
a bare route, relative path, document ID, or bare URL. Do not include
credentials, tokens, or secret query parameters in the presented link.

If status capture or validation fails, report that registration succeeded but
no safe clickable URL is currently available. Never substitute the route or
raw status URL. Preserve the document ID and revision, report the diagnostic
retry command `mdmaid-desk daemon status`, and mention `mdmaid-desk web` only
as an explicit user-run way to start a foreground service. Never start it
automatically.

Do not generate a standalone HTML copy or invoke `codex-readable-doc` as an
implicit preview. Do not open a browser merely because Markdown was created,
validated, registered, or marked for review. Browser presentation is a separate
opt-in action that requires an explicit current request.

## Request a human decision only when explicitly required

Ordinary readable output remains passive. Do not add review expectations,
decision controls, or a wait merely because attention is `approval`. Enter
decision mode only when the calling workflow explicitly requires a human
direction or plan decision for the exact artifact revision.

Before decision-mode delivery, check `mdmaid-desk --help` for
`--expect plan-decision` and `review wait`. If either capability is absent,
preserve the validated artifact, stop the approval transition, and report that
the installed mdmaid-desk must be upgraded. Do not silently fall back to an
attention-only document or a chat-implied approval.

For a direction decision, use `kind=decision` with the existing authenticated
`plan-decision` transport and identify the semantic mode as `direction` in the
title, request message, and local receipt. Do not invent a dedicated desk
request kind.

Create a recoverable request first. Add these options to the normal `register`
or `import` command and retain its JSON receipt:

```text
--attention approval \
--expect plan-decision \
--request-message "<what the human should decide and any important focus>"
```

Read `reviewRequest.id` from the successful result, record it with the document
ID, revision, artifact path, and locally computed content hash, then start the
one long-lived foreground waiter in the current agent turn:

```bash
mdmaid-desk review wait <review-id> --json
```

Run it and keep the current agent turn open until that command returns the
durable result. The external command may sleep without model reasoning, but the
enclosing turn must remain active. Do not background or detach the waiter.
Do not return a final response while the review is pending; a
`waiting_for_approval` receipt is an intermediate update only.
The initial update must include the review request ID, exact document revision,
and the full absolute HTTP(S) mdmaid.desk document URL as a clickable Markdown
link when the safe origin contract above succeeds. Otherwise state that no
safe clickable URL is currently available; never print only the navigation
route.

If the execution tool yields a process or session ID instead of completed JSON,
resume that same process or session until it exits. Use the longest safe blocking interval allowed by active harness policy.
A yielded ID proves only that the waiter is still running;
it is not permission to finish the turn. Do not start a replacement waiter, create
nested wait commands, or poll only to demonstrate liveness. Report user-visible
updates only for material state changes, a waiter failure or required intervention,
an explicit user status request, or the final durable result.
Do not narrate unchanged pending checks. When the command exits,
surface the received outcome and human response text immediately, then apply
the outcome routing below. Do not wait for another user chat message to inspect
a completed waiter.

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
  direction or planning according to the recorded decision mode; publish the
  changed revision as a new request after review;
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
- the document ID and revision from the JSON receipt;
- the full absolute HTTP(S) mdmaid.desk document URL as a clickable Markdown
  link when the safe origin contract succeeds, or the explicit
  safe-link-unavailable result and diagnostic retry command when it does not.

Do not duplicate the full document in chat unless requested.
