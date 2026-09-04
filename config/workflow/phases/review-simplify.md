---
name: work-review-simplify
description: Run the canonical implementation review with the behavior-preserving maintainability profile.
version: 0.1.0
---

# /work-review-simplify - Maintainability Review And Repair

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

This command is a thin alias for the canonical `work-review` workflow. Read and
follow the installed `work-review` definition in full, using `implementation`
mode and the `maintainability` profile. Do not create a second review engine or
copy its lens definitions here.

Input:

`$ARGUMENTS`

Resolve an invocation without explicit routing as:

```text
/work-review implementation --profile maintainability <target or focus>
```

Preserve an optional leading route block and place the fixed mode and profile
after its delimiter:

```text
/work-review-simplify @agy @codex -- <target or focus>
/work-review @agy @codex -- implementation --profile maintainability <target or focus>
```

Require an implementation target under the same resolution rules as
`work-review`. Do not reinterpret a plan as implementation work.

This alias does not widen authority. Reviewers and verifiers remain read-only;
the coordinating harness applies only independently confirmed fixes and runs
the canonical verification and reporting gates.
