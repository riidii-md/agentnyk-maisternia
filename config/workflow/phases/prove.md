---
name: work-prove
description: Expand a candidate plan's acceptance contract when risk requires more detailed observable proof.
version: 0.3.0
---

# /work-prove - Define the Acceptance Contract

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Treat `/work-prove` as an optional expansion of the plan's acceptance contract,
not a mandatory artifact for every task. Use it when integration, migration,
security, operational, or end-to-end risk needs evidence that would make the
main plan hard to review. Work from the candidate plan before human approval.

Input:

`$ARGUMENTS`

Expand selected high-risk entries in the candidate plan's same consolidated test plan;
do not create a second portfolio or mandatory proof artifact. For each selected
task, define:

- setup or initial state;
- action or command;
- observable result;
- most likely error path;
- evidence source;
- verification tier;
- expected test location or command when discoverable.

Use risk-based tiers:

- offline deterministic;
- gated live integration;
- end-to-end smoke;
- non-blocking external canary.

Also define a concise component-level contract for the most important end-to-end
behavior. Do not use arbitrary task or assertion counts. Repository conventions
and risk determine depth.

If the plan already contains sufficient acceptance evidence, report that the
proof is included and return to the plan's explicit AI review choice without
duplicating it. Do not automatically invoke `/work-plan-review`. Do not
implement code or approve the candidate plan.
