# /codex-review - Use Codex For Multi-Lens Review

Run Codex as an independent, read-only reviewer of a plan, targeted plan delta,
diff, PR, or implementation. Claude remains the coordinator and applies fixes.
This is command-only orchestration and must not install hooks.

## Usage

```text
/codex-review auto <target or focus>
/codex-review plan <plan or design path>
/codex-review plan-delta <changed decision or section>
/codex-review implementation <contract, diff, PR, or focus>
```

An explicit Codex invocation is approval to share the stated review packet with
Codex for this run. Before dispatch, remove secrets, personal/customer data,
unrelated proprietary context, and unnecessary conversation history. State the
files or excerpts that will be shared if the scope is sensitive.

## Resolve Target

Explicit mode wins. In `auto`, prefer `implementation` when a diff or PR exists;
otherwise use `plan` when a plan or design artifact exists. Ask only when no
reviewable artifact can be established. Never refuse an explicit plan review
because there is no implementation diff.

Capture local repository context without assuming or fetching a remote:

```bash
git status --short
git branch --show-current
BASE_REF="${CODEX_BASE_REF:-$(git symbolic-ref --quiet refs/remotes/origin/HEAD | sed 's#^refs/remotes/##')}"
[ -n "$BASE_REF" ] || BASE_REF="origin/main"
git diff --stat "$BASE_REF"...HEAD 2>/dev/null || git diff --stat
```

Build a minimal, self-contained handoff with the mode, target, accepted
contract, relevant plan sections, base ref, repository rules, changed paths,
verification evidence, and unresolved decisions. Do not use builder reasoning
as proof.

## Run Codex

```bash
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.6-terra}"
CODEX_PROFILE_ARGS=()
[ -n "${CODEX_PROFILE:-}" ] && CODEX_PROFILE_ARGS=(--profile "$CODEX_PROFILE")
TASK="$(mktemp /tmp/codex-review.XXXX.md)"
cat > "$TASK" <<'PROMPT'
Run an independent multi-lens review with fresh context.

Mode, target, and focus:
$ARGUMENTS

Minimal conversation handoff:
<CONVERSATION_HANDOFF>

First discover repository instructions, the target artifact, actual relevant
code, tests, dependencies, migrations, CI, accepted decisions, and base state.
A diff or summary is not sufficient evidence. Do not edit files.

For implementation mode, run independent lenses for correctness, consistency,
completeness and edge cases, security, architecture, simplicity/DRY,
diff-analysis, dependency-currency, and tests/verification.

For plan or plan-delta mode, run independent lenses for correctness-vs-code,
internal consistency, completeness and edge cases, architecture/simplicity,
best practices, and acceptance/testability. A plan-delta review should stay
targeted unless the changed decision invalidates wider dependencies or scope.

Add domain lenses only when warranted: accessibility, privacy/PII/SOC 2
evidence, performance/scalability, migration safety, API compatibility, or data
integrity. Do not claim compliance without repository evidence.

Use one read-only subagent per lens when available; otherwise run the same
lenses sequentially and say so. Every candidate finding must include severity,
impact, proposed fix, and concrete file:line, short verbatim quote, test/command,
or authoritative-document evidence from the actual repository.

For every candidate, launch a separate verifier whose job is to refute it.
Require explicit is_real and grounded booleans, rationale, evidence, and
corrected severity. Keep only is_real && grounded. Deduplicate confirmed
findings and rank Critical, High, Medium, then Low.

Return:
- resolved mode, target, base ref, and lenses run;
- confirmed findings with grounding and proposed fixes;
- refuted findings and why they were rejected;
- unavailable or failed lanes;
- recommended focused and final verification;
- gate status, where unresolved Critical/High findings fail the gate.

Codex is a read-only reviewer in this invocation. Do not claim fixes were
applied.
PROMPT
OUTPUT="$(mktemp /tmp/codex-output.XXXX.md)"
codex exec "${CODEX_PROFILE_ARGS[@]}" \
  --model "$CODEX_REVIEW_MODEL" \
  -c 'model_reasoning_effort="high"' \
  --ephemeral \
  -o "$OUTPUT" \
  -C . \
  --sandbox read-only \
  - < "$TASK"
printf 'Full Codex review: %s\n' "$OUTPUT"
```

## Verify And Apply

Treat the Codex result as review input, not automatic truth. Claude verifies
the surviving findings against the repository, writes `review.md` and
schema-valid `review.json`, then applies every confirmed fix within approved
scope. For plans, edit the plan or design artifact. For implementations, edit
code and tests. Mark decision-dependent or unauthorized fixes blocked instead
of guessing.

Critical and High findings are blocking. Run focused checks after repairs and
the repository-required final verification before reporting gate status. Report
confirmed and applied fixes first, then refuted findings and rationale.
