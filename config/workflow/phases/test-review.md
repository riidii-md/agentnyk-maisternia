---
name: work-test-review
description: Review test evidence for changed behavior through intent, risk, fidelity, and maintainability lenses using the canonical implementation-review engine.
version: 0.2.0
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
/work-test-review authoring <change or proposed tests>
/work-review implementation --scope tests --test-mode authoring <change or proposed tests>
/work-test-review audit <test area>
/work-review implementation --scope tests --test-mode audit <test area>
/work-test-review campaign <subsystem>
/work-review implementation --scope tests --test-mode campaign <subsystem>
```

Treat a leading `authoring`, `audit`, or `campaign` after any route delimiter as
the explicit test mode and remove it from the target. Otherwise use test mode
`review`, preserving the existing command behavior. Never infer `campaign` from
a large target; it requires an explicit invocation and one named subsystem.

Preserve an explicit `--disposition repair` or
`--disposition report-only`. `/work-start-pr-review` supplies report-only for
PR publication so this specialization cannot edit a contributor branch.
Preserve `--allow-degraded` only when the user explicitly supplied it; the
alias inherits the canonical blocked fallback and degraded reporting rules.

Preserve an optional leading route block and place the fixed mode and scope
after its delimiter:

```text
/work-test-review @agy @codex -- <target or focus>
/work-review @agy @codex -- implementation --scope tests <target or focus>
/work-test-review @agy @codex -- audit <test area>
/work-review @agy @codex -- implementation --scope tests --test-mode audit <test area>
```

Require an implementation, test-only change, or explicit implementation
contract under the same target-resolution rules as `work-review`. A standalone
test review runs the canonical four-lens test bundle and any inspection needed
to ground or refute its findings; it does not run unrelated implementation
lenses. The installed `work-review` definition owns the lens meanings, evidence
matrix, metric guardrails, fixture safety, severity, repair, verification, and
reporting rules. Do not copy or reinterpret those rules in this alias.

Record `mode: implementation`, `scope: tests`, the selected `test_review_mode`,
and the complete `test_evidence` matrix in the normal
`.agent-runs/reviews/<run-id>/review.md` and `review.json`. Authoring mode also
records `test_authoring_gates`; audit and campaign record
`test_audit_candidates`; campaign additionally records `test_campaign`.
This alias does not widen authority or add a human decision. Reviewers and
verifiers remain read-only. Under repair disposition, the coordinating harness
owns only independently confirmed fixes within the accepted scope.
Under report-only disposition, it must not apply fixes and records confirmed fixes as
`not-applicable` under the canonical review-process gate.
