---
name: work-ready
description: Derive whether the reviewed and explicitly approved plan is safe to execute without inventing missing decisions.
version: 0.2.0
---

# /work-ready - Implementation Readiness Gate

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Evaluate implementation readiness after the human decision. This is a derived
gate over existing evidence, not a phase that creates or approves missing work.

Input:

`$ARGUMENTS`

Check:

- the final reviewed plan exists;
- its scope, exclusions, tasks, and acceptance criteria are complete;
- required plan review passed;
- the approved plan content hash matches the current plan;
- expanded proof exists when risk requires it;
- repository rules are known or explicitly unknown;
- important risks are resolved or explicitly accepted;
- execution authority and approval boundaries are known;
- no critical blocker remains.

Return pass, conditional pass, or fail with exact missing inputs and the next
phase. A passing result routes directly to `run` in a continuous session or to
`handoff` for a fresh executor. Do not use readiness to approve a plan, fill in
a missing human decision, or proceed past unresolved critical ambiguity.
