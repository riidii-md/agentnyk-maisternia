# /work-ready - Readiness Gate

Evaluate whether the task can safely move to planning or implementation.

Input:

`$ARGUMENTS`

Check:

- task statement exists;
- facts and assumptions are separated;
- scope and exclusions are clear;
- acceptance criteria exist;
- solution direction is decided;
- repository rules are known or explicitly unknown;
- important risks are resolved or accepted;
- required user approval exists.

Return pass, conditional pass, or fail with exact missing inputs and the next
phase. Do not proceed past unresolved critical ambiguity.
