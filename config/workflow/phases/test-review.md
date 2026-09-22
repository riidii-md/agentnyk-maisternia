---
name: work-test-review
description: Review test evidence for changed behavior through intent, risk, fidelity, and maintainability lenses using the canonical implementation-review engine.
version: 0.1.0
---

# /work-test-review - Specialized Test Review And Repair

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

This command is a thin specialization of the canonical `work-review` workflow.
Read and follow the installed `work-review` definition in full, using
`implementation` mode and the `tests` scope. Reuse its evidence gathering,
read-only reviewers, independent candidate refutation, coordinator-owned
repairs, verification, and report artifacts. Do not create a second review
engine or another human approval gate.

Input:

`$ARGUMENTS`

Resolve an invocation without explicit routing as:

```text
/work-review implementation --scope tests <diff, branch, PR, contract, or focus>
```

Preserve an explicit `--disposition repair` or
`--disposition report-only`. `/work-start-pr-review` supplies report-only for
PR publication so this specialization cannot edit a contributor branch.

Preserve an optional leading route block and place the fixed mode and scope
after its delimiter:

```text
/work-test-review @agy @codex -- <target or focus>
/work-review @agy @codex -- implementation --scope tests <target or focus>
```

Require an implementation, test-only change, or explicit implementation
contract under the same target-resolution rules as `work-review`. A standalone
test review runs the canonical four-lens test bundle and any inspection needed
to ground or refute its findings; it does not run unrelated implementation
lenses. The installed `work-review` definition owns the lens meanings, evidence
matrix, metric guardrails, fixture safety, severity, repair, verification, and
reporting rules. Do not copy or reinterpret those rules in this alias.

Record `mode: implementation`, `scope: tests`, and the complete `test_evidence`
matrix in the normal `.agent-runs/reviews/<run-id>/review.md` and `review.json`.
This alias does not widen authority or add a human decision. Reviewers and
verifiers remain read-only, and the coordinating harness owns only independently
confirmed fixes within the accepted scope.
