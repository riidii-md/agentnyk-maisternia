---
name: work-session-analysis
description: Analyze one completed session with dedicated reviewers for token cost, repetition, skills, user friction, setup, commands, and delegated subagents, then propose evidence-backed improvements.
version: 0.1.0
---

# /work-session-analysis - Find Session Bottlenecks

Analyze one completed session at the end of the run. Find concrete bottlenecks
and suggest the smallest measurable improvement for each one. Do not modify
product code or durable harness configuration during this command.

Input:

`$ARGUMENTS`

Use the current session history only when the harness exposes it to this command.
Otherwise require an explicit transcript or event-export path. Never search all
provider histories or infer missing events from memory.

## Prepare The Evidence Bundle

1. Record the task contract, provider, model, repository, final result, changed
   files, verification, and user corrections.
2. Collect the selected transcript, tool and subagent events, provider usage,
   timing, and the active harness profile.
3. Redact secrets before delegating content to another model or provider.
4. Separate measured values, estimates, and unknowns.
5. Create `.agent-runs/retrospectives/<run-id>/session-analysis.md` and update
   the run's schema-valid `record.json`.

## Delegate Seven Review Lanes

Launch one scoped reviewer per lane when subagents are available. Run independent
lanes in parallel within the configured token and concurrency budget. If the
harness cannot delegate, execute the same lanes sequentially and state that
limitation in the report.

Every reviewer must cite event IDs, transcript locations, file paths, commands,
or telemetry. It must return `BOTTLENECK`, `NO_BOTTLENECK`, or
`INSUFFICIENT_EVIDENCE`, followed by impact, confidence, a root-cause hypothesis,
and one minimal improvement. A hypothesis is not a verified fact.

### 1. Token And Time Cost Reviewer

Measure input, output, cache, and reasoning tokens, model and subagent calls,
tool calls, retries, permissions, latency, and price when provider telemetry
exists. Attribute cost to useful work, verification, rework, delegation, and
avoidable detours. Do not present text-length estimates as exact token usage.

### 2. Repetitive Work Reviewer

Find repeated file reads, searches, explanations, failed attempts, retries,
manual transformations, and rediscovery of already available facts. Distinguish
necessary verification from repetition that did not add evidence. Suggest a
cache, command, skill, deterministic tool, or workflow change only when it would
remove demonstrated repetition.

### 3. Skills Reviewer

Inspect skills that activated, should have activated, or interfered with the
task. Find broken paths, invalid frontmatter, vague activation descriptions,
duplicate skills, conflicting instructions, stale procedures, missing tools,
and skills whose use added cost without improving the result. Preserve useful
skills and recommend a testable edit, removal, or activation change.

### 4. User Friction Reviewer

Find places where the user had to repeat context, correct the agent, reject an
unexpected action, explain the same preference, ask for readable output, or
recover from unclear status. Describe interface or workflow friction without
blaming the user. Separate avoidable agent friction from genuinely missing task
requirements.

### 5. Setup Friction Reviewer

Find missing executables, stale configuration, broken paths, permission loops,
provider capability mismatches, unavailable MCP servers, environment problems,
incorrect repository discovery, and setup work repeated across sessions. Prefer
doctor checks, validation, or setup automation over additional prose.

### 6. Commands Reviewer

Inspect commands that failed, used wrong flags, assumed the wrong provider,
parsed output incorrectly, required repeated correction, or lacked a safe
fallback. Distinguish command defects from environment failures and user-owned
customization. Suggest exact command, validation, adapter, or error-message
improvements.

### 7. Delegation Reviewer

Inspect every delegated subagent task. Check whether its scope was precise,
evidence requirements were explicit, context was sufficient, work was duplicated,
results were actually used, and merge or arbitration cost exceeded the value
returned. Identify missing delegation where parallel independent work would have
materially reduced time or risk.

## Synthesize

Merge duplicate findings across lanes and preserve disagreements. Rank a finding
as a bottleneck only when evidence shows material impact on correctness, user
effort, token/time cost, safety, or repeated work. Do not invent an improvement
for a lane with no demonstrated bottleneck.

For every accepted bottleneck report:

```text
ID
Lane and category
Observed bottleneck
Evidence
Measured or unknown impact
Root-cause hypothesis
Minimal proposed improvement
Expected metric change
Validation or replay method
Risk and rollback
Confidence
```

End with:

1. the top three bottlenecks, if any;
2. quick deterministic fixes;
3. harness configuration proposals requiring replay;
4. findings that need more evidence;
5. explicit human decisions required.

The command produces analysis and proposals only. It must not automatically edit
prompts, skills, commands, MCP configuration, hooks, settings, provider routing,
or global instructions.
