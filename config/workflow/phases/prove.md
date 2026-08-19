---
name: work-prove
description: Expand a candidate plan's acceptance contract when risk requires more detailed observable proof.
version: 0.2.0
---

# /work-prove - Define the Acceptance Contract

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Treat `/work-prove` as an optional expansion of the plan's acceptance contract,
not a mandatory artifact for every task. Use it when integration, migration,
security, operational, or end-to-end risk needs evidence that would make the
main plan hard to review. Work from the candidate plan before human approval.

Input:

`$ARGUMENTS`

For every task, define:

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
proof is included and continue to plan review without duplicating it. Do not
implement code or approve the candidate plan.
