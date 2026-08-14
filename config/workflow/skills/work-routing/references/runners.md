# Safe Harness Runners

Read this reference before dispatching a route outside the current harness.
Prefer the current harness's native subagent facility when it can select the
requested provider and enforce the required authority. Otherwise use a locally
available provider CLI only after the main routing contract passes.

Resolve the model source before launch: explicit model selector, saved model
preference, configured phase/role mapping, or provider default. Prefer a native
same-harness worker when it can enforce the selected model and authority. A
different model cannot replace the already-running coordinator session; it
requires a fresh lane. Verify model eligibility when the harness exposes that
evidence. Never silently substitute a model or claim a model was used without
runner evidence.

For current-harness execution, model-selected same-harness work requires a native subagent that accepts an explicit model. Keep the current session as
coordinator and give the subagent only the bounded phase packet. If native model
selection is unavailable, fail visibly; do not run the phase in the parent or
substitute a CLI while calling it a subagent. Named external harness routes
continue to use the isolated provider runner below.

## Check eligibility and authority

Determine the current harness and inspect each requested runner at execution
time. Use `maisternia provider inspect --json <harness>` or equivalent local
capability evidence when available. Do not run a provider's native doctor or
initialize configuration automatically.

Match the workflow's required authority:

- read-only workflows may use safe headless or native subagent execution;
- artifact writes stay with the coordinator unless exact artifact authority was
  approved;
- workspace-write delegation requires explicit human approval, a bounded task,
  and repository-appropriate isolation;
- external writes, commits, pushes, publication, deployment, production access,
  and destructive actions require separate approval and never follow merely
  from an `@harness` cue.

Use this conservative automated baseline:

| Harness | Baseline |
|---|---|
| Codex | Ephemeral; read-only sandbox; bounded workspace-write only after approval. |
| Claude | Fresh print/native agent; sterile plan/read-only execution; bounded edits only after approval. |
| Antigravity (`agy`) | Sterile capability evidence plus plan mode and sandbox; read-only only. |
| Hermes | Supervised interactive execution only; never unattended one-shot execution. |

Never widen authority because a preferred harness cannot satisfy the workflow.

## Prepare the delegation packet

Send only task-local context:

- workflow name and cleaned task;
- logical repository identity, approved sanitized staging root, and applicable
  repository instructions;
- accepted decisions, constraints, and exact referenced artifacts;
- required output and evidence;
- authority, disclosure, time/token budget, and stop conditions.

Remove credentials, personal/customer data, unrelated proprietary context, and
unnecessary conversation history. Naming a harness explicitly approves that
harness and the minimal disclosed task for this invocation; it does not approve
sensitive context or broader write authority. Treat every provider boundary as
a disclosure boundary even for the same model family.

## Common packet and output rules

- Create task and output files in a private temporary directory with user-only
  permissions.
- Do not expose the live workspace as an external provider's working root.
  Create a private sanitized staging tree and make it the runner's current
  directory/root. Copy only approved regular files or excerpts, preserving
  relative paths when they matter. Never follow symlinks; reject them or replace
  them with an explicit non-clickable description.
- A repository-wide task does not imply silent repository-wide disclosure.
  Preview the tracked/untracked scope and sensitive categories, then obtain
  explicit approval before staging a broad snapshot. Exclude ignored files,
  VCS metadata, credentials, local configuration, runtime data, and unrelated
  artifacts. Include worktree changes and applicable repository instructions
  deliberately instead of copying the directory wholesale.
- Put the complete, already-redacted delegation packet in the task file. Do not
  construct it through unquoted shell interpolation.
- Use one output path per lane and preserve provider/model attribution.
- Treat nonzero exit, missing output, timeout, parse failure, or an authority
  request as a failed lane. Do not replace its provider silently.
- Do not pass credentials through prompts, command arguments, or environment
  overrides.
- Never use approval-, permission-, sandbox-, rule-, or hook-bypass flags.

## Codex

For a read-only fresh lane, use the behavioral equivalent of:

```bash
codex exec \
  --ephemeral \
  --ignore-user-config \
  --skip-git-repo-check \
  -C "$STAGING_ROOT" \
  --sandbox read-only \
  --model "$MODEL" \
  -o "$OUTPUT" \
  - < "$PACKET"
```

Include `--model "$MODEL"` only when model resolution did not end at provider
default. Pass it as a separate process argument; never interpolate it into a
shell command.

An approved write lane may change only `--sandbox read-only` to
`--sandbox workspace-write`; it still needs the routing contract's explicit
workspace-write approval and bounded scope. Do not use danger-full-access or
approval bypasses.

Use the configured model-role mapping when present: strong reasoning for
analysis, research, decisions, and plans; balanced review for proof, readiness,
verification, review, and PR checks; fast coding only for an approved run.

## Claude

For a read-only fresh lane, use the behavioral equivalent of:

```bash
claude --print \
  --safe-mode \
  --disable-slash-commands \
  --strict-mcp-config \
  --mcp-config '{"mcpServers":{}}' \
  --permission-mode plan \
  --tools "Read,Grep,Glob" \
  --model "$MODEL" \
  --no-session-persistence \
  < "$PACKET" > "$OUTPUT"
```

Include `--model "$MODEL"` only for a resolved override. Claude aliases such as
`opus` and `sonnet` remain provider-owned aliases; do not rewrite them to a
guessed snapshot.

Launch the process with `$STAGING_ROOT` as its current directory. The lane is
unavailable when the installed Claude version cannot disable customizations and
external MCP configuration; do not weaken isolation to make it run.

Add a narrowly matched read-only tool only when the lane needs evidence that
the basic file tools cannot obtain. Do not enable edit tools or unrestricted
shell access for a read-only lane. A write-capable Claude lane requires separate
approval and a provider-native bounded permission mode.

## Antigravity (`agy`)

Antigravity's current automated contract is read-only. Use the behavioral
equivalent of:

```bash
agy --print "<contents of the redacted packet>" \
  --mode plan \
  --sandbox \
  --disable-slash-commands \
  --add-dir "$STAGING_ROOT" \
  --print-timeout 10m > "$OUTPUT"
```

Pass the packet as one argument using the host's safe process API where
possible. Do not grant accept-edits or permission bypasses. Treat the result as
text unless current capability inspection proves structured output support.
Launch from `$STAGING_ROOT`. Because the current AGY CLI has no general sterile
startup flag, automated use also requires capability evidence that no enabled
startup integration widens the plan/sandbox contract; otherwise report the lane
unavailable or use supervised execution.

## Hermes

Hermes is interactive only under the current contract. Its one-shot mode
bypasses dangerous-operation approvals and is not eligible for unattended
delegation. Open a supervised interactive Hermes session, present the routing
packet and boundaries, and keep the user present for approvals. If interactive
supervision is unavailable, report the lane unavailable instead of using
one-shot mode.

## Several harnesses

For one harness, use `single`. For several, resolve or state one of:

- `first-capable`: select the first eligible harness;
- `parallel-independent`: return independent results;
- `parallel-verify`: produce a primary result plus independent verification;
- `pipeline`: pass bounded artifacts through named stages.

The workflow's declared strategy overrides an incompatible saved strategy.
Otherwise research, analysis, decisions, and plans default to
`parallel-independent` plus local synthesis; readiness, proof, verification, PR
checks, and review default to `parallel-verify`.

For `/work-review @agy @codex @claude -- <target>`, distribute independent
read-only lenses, verify findings across origins, and keep the current harness
as coordinator and sole owner of fixes. Multi-harness implementation requires
an approved dependency plan, isolated writes, and explicit integration
ownership.

Run independent read-only lanes concurrently only when the coordinator can keep
their packets, output paths, budgets, and attribution separate. Do not implement
parallelism with shared temporary files. For `parallel-verify`, never ask a lane
to verify its own candidate finding when another eligible selected harness is
available.

Keep results separate until synthesis and preserve material disagreement. The
coordinator verifies returned work, performs coordinator-owned writes, and
delivers the final result. A fresh delegated run does not inherit the current
conversation, so its packet must be self-contained.

## Show the route

Before dispatch, emit a compact routing receipt:

```text
Route: Codex · delegated · read-only · explicit invocation
```

When a model is selected, include it and its source, for example:

```text
Route: Claude · sonnet · subagent · workspace-write · saved model preference
```

For several harnesses, include the strategy. Expand the receipt only when the
user must review files, sensitive categories, budget, or new authority. If a
target is unavailable or unsafe, ask whether to run locally, choose another
harness or model, inherit the provider default, or stop. Never silently
substitute a harness or model, claim delegation that did not occur, or discard
the user's route.
