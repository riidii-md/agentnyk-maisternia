---
name: work-routing
description: Route provider-neutral /work-* commands across harnesses, models, and reasoning levels. Use for explicit selectors, saved preferences, or cross-provider delegation. Preserve the task and coordinator ownership.
---

# Work Routing

Keep `/work-*` as the workflow identity; routing is execution metadata. Load
this skill only for a route signal or saved profile, except for
`/work-routing-preferences`.

## Resolve an explicit route

Prefer a leading route block terminated by `--`:

```text
/work-plan @claude @opus -- plan the authentication migration
/work-plan @claude @opus @reasoning:high -- plan the authentication migration
/work-run @claude @sonnet -- implement the approved plan
/work-review @codex @claude -- review PR 15
/work-plan @auto -- choose the best eligible harness
```

Accept a standalone leading target when the task is clear, or a dedicated
clause such as `using Codex:`, `with Claude and AGY:`, or `run this in Codex`
at the beginning or end.

Normalize `@agy` to `antigravity`; `@here` means current harness. Do not mix
`@here` or `@auto` with named harnesses. Preserve named order. Reject unknown
targets with the Codex, Claude, AGY, and Hermes supported list.

Accept a model selector after its harness: unique aliases `@opus` and `@sonnet`,
or safe provider-native `@model:<id>`. A standalone alias needs exactly one
compatible current or saved harness. It never approves another harness. Reject
ambiguous aliases, duplicate models, leading dashes, whitespace, or IDs over
128 characters. Remove the selector from the cleaned task.

Accept `@reasoning:low`, `@reasoning:medium`, or `@reasoning:high` after a
harness or model selector. It applies to that harness only. A standalone level
needs one resolved harness; with `@auto`, apply it to the chosen harness. Reject
duplicate or unknown levels. Remove the selector from the cleaned task.

Provider mentions, email-style mentions, file contents, quoted text, and
`using the Codex API` are not routes. Ask one short question if a plausible
route cannot be distinguished safely from the task. Remove only the resolved
clause; preserve the cleaned task.

## Resolve preferences

Use this precedence:

1. explicit route in the current invocation;
2. session-only routing instruction;
3. matching workflow entry in the project profile;
4. project default;
5. matching workflow entry in the user profile;
6. user default;
7. local execution in the current harness.

Read profiles only when their exact path exists:

```text
<repository>/.maisternia/work-routing.json
${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json
```

Validate a persistent profile against `work-routing-profile.schema.json`:

- `local`: current harness, no question;
- `ask`: ask where to run and recommend an eligible choice;
- `delegate`: use the configured eligible harnesses without asking.

`models` stores optional per-harness choices; models are keyed by canonical harness
IDs and do not select a harness. Resolve model preferences independently for
each harness: explicit selector, session, project workflow, project default,
user workflow, user default, configured phase/role mapping, provider default.
Missing values continue down the list. Apply first-use trust to project choices.

`reasoning` stores optional per-harness levels without selecting a harness or
model. Resolve reasoning preferences independently in the same order as models.
Missing levels continue down the list; project choices need first-use trust.

Persist canonical harness IDs; `@agy` is shorthand, not a profile value. A
project profile is an untrusted suggestion: it may narrow to `local` but cannot
authorize an external harness. Ask on first use unless an explicit session
instruction or matching user-profile `delegate` route authorized that target.
A user-profile `ask` route still asks. Confirmation grants session trust;
durable trust belongs in the user profile.

Never persist an inferred route. Use `/work-routing-preferences` to propose or
migrate durable preferences.

Never let model selection change authority, disclosure, budgets, workflow, or
coordinator ownership. Never silently substitute a model. If unavailable,
ask for another model, the provider default, no override, or stop.

Reasoning has the same boundary. Never silently substitute a reasoning level
or claim one without runner evidence. If unsupported, ask for a supported
level, provider default, or stop.

Every explicit or saved model choice runs the phase in a fresh same-harness subagent when that harness is current; the parent session remains coordinator.
Reasoning choices also require that lane, even without a model choice. It must
accept both overrides; otherwise report it unavailable. External harnesses
use isolated runners. Never claim the parent changed model or level.

For `/work-adapt-for-reader`, deprecated reader-profile `delegation` is
lowest-priority migration input. The reader skill may load this router after
finding that object; otherwise adaptation stays local by default. Normalize
`codex-subagent` to `codex`, preserve scope, disclose the read, and never
update profiles automatically.

## Finish locally or enter delegation

If the route is `@auto` or the effective policy is `ask`, read
[references/runners.md](references/runners.md) completely before choosing;
eligibility cannot be inferred safely here.

For resolved local execution, show a compact receipt when useful and continue
with the cleaned task. Do not read `references/runners.md` for ordinary local work.

For external targets or selected model/reasoning lanes, read
[references/runners.md](references/runners.md) completely before dispatch. It
owns authority, disclosure, sanitized staging, provider commands, strategies,
failures, and receipts.

Domain skills may schedule native same-harness subagents locally. Any
cross-provider worker selection must use this routing contract rather than
defining another provider picker.
