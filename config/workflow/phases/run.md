---
name: work-run
description: Execute an approved implementation contract in small verified steps and report only genuinely completed outcomes.
version: 0.2.0
---

# /work-run - Execute the Approved Plan

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Implement an approved handoff in small, restartable steps.

Input:

`$ARGUMENTS`

For each pass:

1. Read the approved plan or handoff plus the current session's progress and
   select the first unfinished unblocked task.
2. Reconfirm the task against current code and repository rules.
3. Implement only that task.
4. Add or update focused tests.
5. Run the task criteria and smallest relevant checks.
6. Report result, files, checks, attempt count, and next action to the
   coordinating session.
7. Repeat.

Prefer non-login shell execution for routine commands so startup scripts do not
add unrelated output or latency.

## Place Rationale Deliberately

Prefer code that explains itself through precise names, narrow responsibilities,
types, assertions, tests, and direct control flow. Before adding a comment,
check whether simpler structure or clearer code can remove the need. Many
comments are a reason to reconsider the design, but not proof that the design is
wrong.

Keep a local comment only when the code cannot express why the constraint exists
and what can break if it is ignored. Examples include non-obvious invariants or
trust boundaries, external protocol behavior, compatibility and safety
constraints, and deliberate limitations with a concrete upgrade trigger. Keep
the comment to the constraint and consequence. Do not add line-by-line narration,
restate syntax, leave commented-out code, or speculate about future features.

Place broader rationale where it remains useful:

- use durable Markdown or an ADR for cross-cutting design, alternatives, and
  tradeoffs, leaving a short local pointer when discovery at the code matters;
- use a ticket such as Jira for product context, acceptance scope, and business
  decisions, but do not make a mutable or access-restricted ticket the only source
  of a code-critical invariant;
- use a commit or pull-request message for why the change is being made now,
  reviewer context, and migration history, but not as the only source of
  rationale needed to modify the code safely.

Create or update documentation, tickets, commits, and pull requests only when
they are within the approved task and authority. Do not manufacture an external
artifact merely to avoid one useful local comment.

When the approved task asks to build or install an executable, completion also
requires an installation verification gate:

1. Run the repository's relevant tests and build command.
2. Resolve and report the exact installation destination before writing outside
   the workspace, and obtain any required approval once for that bounded target.
3. After installation, use `command -v` to confirm which executable will run.
4. Confirm its identity or version, then run the smallest safe smoke check.
5. Do not report the task complete until those checks pass or a concrete blocker
   is recorded.

Park a task after three failed verification attempts. Stop when three tasks are
parked, all remaining work is blocked, or the configured run cap is reached.

Do not commit, push, open a PR, weaken criteria, touch production data, or
perform destructive actions without explicit authorization.

Do not create or update a Maisternia runtime task. Use a progress artifact only
when the approved plan explicitly supplies one.
