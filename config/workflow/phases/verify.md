---
name: work-verify
description: Discover and run repository-specific verification gates, preserve evidence, and report readiness accurately.
version: 0.2.0
---

# /work-verify - Run Repository-Specific Checks

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Verify the current change using the repository's actual commands.

## Go Preflight

When Go checks are required, inspect only GOCACHE, GOMODCACHE, and required
toolchain/module access with `go env GOCACHE GOMODCACHE GOTOOLCHAIN`. Do not dump
the full environment. Verify writable build-cache access once in the current
sandbox, along with module access needed by the repository's actual checks.
Reuse the result in the coordinator and worker worktrees until settings change.

Run test, vet, build, and scoped formatting as direct commands. Keep edits,
scripts, environment setup, and log redirection separate; use the runner's output
capture rather than changing an approved command into an opaque shell wrapper.
Check loaded native routine-development approvals once, including their scope
and inheritance. Existing approvals do not remove a filesystem or network limit.

Do not retry the same permission failure for each check. Use a task-owned writable
cache configured in the session environment when supported, or obtain the bounded
cache access once. Keep dependency/toolchain network access separate from offline
checks. Do not persist global Go settings, grant blanket shell execution, or
disable a provider sandbox. Report an unsupported native permission mapping;
never invent an allowlist for a harness that cannot enforce it. Continue checks
that do not depend on the blocker.

Input:

`$ARGUMENTS`

Discover build, format, lint, typecheck, unit, integration, end-to-end,
security, migration, and review gates from repository instructions, scripts,
CI, and hooks.

Run the smallest relevant checks first, then broaden with risk. Preserve full
logs. Classify failures as regression, environment, flaky, pre-existing, or
unknown.

Check affected documents against the accepted behavior and current implementation.
Reconcile the plan's document updates or no-document-change rationale. Verify
changed commands, paths, links, examples, and interface or security descriptions
at the smallest faithful level. Label proposals and historical decisions so
they cannot be mistaken for current behavior. Report unresolved contradictions
with their sources and return required corrections to execution. Do not claim
consistency when required evidence is unavailable. Return this evidence to the
coordinator for the selected progress summary.

Return commands, results, failure classification, remaining risk, log paths, and
whether independent review can begin.
