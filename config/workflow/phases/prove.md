---
name: work-prove
description: Define a concrete acceptance contract that maps requirements and risks to observable proof.
version: 0.1.0
---

# /work-prove - Define the Acceptance Contract

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Turn the accepted plan into observable proof before implementation.

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

Do not implement code.
