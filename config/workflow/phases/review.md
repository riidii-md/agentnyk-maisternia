# /work-review - Independent Contract and Code Review

Review the accepted contract and actual implementation with fresh context.

Input:

`$ARGUMENTS`

Prefer a runner or model different from the builder. Review the contract, code,
diff, progress claims, and verification evidence. Do not trust the builder's
reasoning transcript as proof.

First verify contract compliance and observable behavior. Then review code for
bugs, regressions, security issues, missing tests, migration risk, and repository
convention failures.

Lead with actionable findings ordered by severity. Every contract finding must
include evidence. If there are no findings, say so and identify residual risk.
