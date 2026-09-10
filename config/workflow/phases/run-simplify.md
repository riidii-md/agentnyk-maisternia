---
name: work-run-simplify
description: Execute an approved plan using the simplest concrete implementation that preserves its behavior and safeguards.
version: 0.2.0
---

# /work-run-simplify - Execute The Simplest Approved Implementation

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

This command is a thin alias for the canonical `work-run` workflow. Read and
follow the installed `work-run` definition in full, adding the simplicity
execution profile below. Do not create a second execution engine or weaken its
approval, testing, verification, retry, parking, or stop behavior.

Input:

`$ARGUMENTS`

Before implementing each selected task, establish the approved behavior,
safeguards, errors, compatibility, and verification that must remain unchanged.
Then inspect the affected code and stop at the first behavior-preserving option
that fully satisfies that contract:

1. apply YAGNI to unrequired behavior or structure;
2. reuse an established repository helper, abstraction, or pattern;
3. use the active language's standard library;
4. use a native platform, framework, database, operating-system, browser, or
   protocol capability;
5. use an already-installed dependency whose contract fits;
6. use straightforward local code and direct control flow;
7. introduce an abstraction only when it removes proven repeated knowledge,
   creates clear ownership, or materially reduces coupling.

Prefer the first safe fit even when a later option is shorter. Do not simplify
away approved behavior, validation, security, accessibility, compatibility,
data-loss prevention, error behavior, or required verification. Do not use
line count as the sole decision metric.

The canonical `work-run` comment-and-rationale policy remains required.
Simplification may remove narration or replace explanation with clearer code,
but it preserves irreducible local rationale unless an equally durable and
discoverable form replaces it.

Equivalent implementation details within the approved contract do not require
a new decision. When the simpler candidate changes approved behavior, scope, an
explicit architectural decision, risk, or verification, stop before that change,
show the concrete tradeoff, and ask the user. A material accepted change returns
to `plan-delta` review and readiness before execution resumes.

For each completed task, report the selected option, complexity avoided, files,
tests, checks, and any concrete condition that would justify more complexity.
This alias does not widen authority.
