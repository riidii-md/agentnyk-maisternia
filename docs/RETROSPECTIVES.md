# Session Retrospectives And Harness Improvement

## Goal

The retrospective presets make completed agent runs useful for improving the
harness over time. They address three related questions:

1. What is currently configured, and what does it cost?
2. What was correct, inefficient, unsafe, or unverifiable in this run?
3. Which repeated findings justify a measured configuration change?

`maisternia` installs the workflow into Codex, Claude, Antigravity, or Hermes. The
selected harness runs it. `maisternia` does not become a session observer, model
runner, or autonomous prompt optimizer.

Codex, Claude, and Hermes use the repository's existing provider mappings.
Antigravity currently renders into the legacy `.config/agy` compatibility tree;
as documented in [Provider adapters](PROVIDERS.md#antigravity-compatibility),
that staging target is not yet proof that Antigravity consumes the files.

## Included Presets

### `harness-profile`

Installs the canonical `work-profile` workflow, the shared
`session-retrospective` skill, and the
retrospective policy.

The command inventories active instructions, commands, prompts, skills, MCP
servers, hooks, settings, plugins, profiles, and model routing. It records
provenance, scope, size, duplication, conflicts, capability gaps, and exact
provider usage when telemetry exists. Missing token, cache, reasoning, latency,
or price data remains `unknown`.

This preset is read-only. Its DAG is:

```text
INVENTORY -> MEASURE -> REPORT
```

### `session-audit`

Installs the canonical `work-audit` and `work-session-analysis` workflows, the shared skill, and the
policy.

The audit reads one explicitly selected run and returns four independent lanes:

```text
CORRECTNESS
OBSERVABLE-TRAJECTORY EFFICIENCY
PROCESS AND SAFETY
COST
```

It does not produce a universal accuracy score. Claims require cited evidence,
unknowns stay unknown, and the transcript is treated as an observable trace
rather than hidden model reasoning.

### `harness-improvement`

Installs the complete command set:

```text
/work-profile
/work-audit
/work-session-analysis
/work-improve
/work-retrospective
```

Its lifecycle is:

```text
PROFILE -> AUDIT -> AGGREGATE -> PROPOSE -> REPLAY -> APPROVE
                                                       | accepted
                                                       v
                                                    INSTALL
                                                       |
                                                       v
                                                    MONITOR -> FINAL
```

Failed replay returns to proposal design. A monitored regression also returns
to proposal design. Only the human-accepted branch reaches installation.

## After Each Run

The `standard-work` delivery workflow offers `/work-session-analysis` as an
optional next step after a PR is successfully created. The user must accept the
offer; PR readiness alone, failed publication, or later PR updates do not start
analysis automatically. Publication remains successful even if the user
declines or the later analysis fails.

After installing `harness-improvement`, invoke `/work-retrospective` at the end
of a completed task. The shared skill also makes the workflow discoverable when
the harness selects skills from task context.

In Codex, invoke `$work-retrospective` from skill suggestions or use the
compatibility prompt `/prompts:work-retrospective`. In Claude Code, use
`/work-retrospective`. Restart the harness after applying the preset.

For a direct analysis of one completed session without running the wider
improvement lifecycle, invoke:

```text
Codex:  $work-session-analysis <explicit session export or current completed task>
Claude: /work-session-analysis <explicit session export or current completed task>
```

The command delegates seven independent review lanes when the provider supports
subagents:

| Reviewer | Bottlenecks it investigates |
|---|---|
| Token and time | Provider usage, retries, rework, delegation cost, avoidable detours |
| Repetitive work | Repeated reads, searches, explanations, attempts, and rediscovery |
| Skills | Broken activation, paths, frontmatter, duplication, conflicts, and stale procedures |
| User friction | Repeated context, corrections, rejected actions, unclear status, and unreadable output |
| Setup friction | Missing tools, permissions, stale config, MCP availability, and provider mismatches |
| Commands | Wrong flags, failed parsing, unsafe assumptions, missing validation, and weak fallbacks |
| Delegation | Duplicate subagents, weak scopes, unused results, context gaps, and merge overhead |

Each reviewer returns `BOTTLENECK`, `NO_BOTTLENECK`, or
`INSUFFICIENT_EVIDENCE` with cited evidence and one minimal, measurable
improvement when a bottleneck exists. Providers without subagents run the same
lanes sequentially and record that limitation.

When the invocation names external harnesses, for example
`/work-session-analysis @codex @claude -- <run>`, the shared `work-routing`
contract owns provider selection, redaction, authority, and fallback. The
retrospective workflow continues to own lane definitions and synthesis.

The first implementation deliberately does not install provider Stop hooks.
Automatically starting model review at every stop can recurse, spend tokens
without a useful task boundary, and disclose transcript content to an
unapproved provider. Provider-native automatic triggers should be added later
as explicit opt-in hook resources after structured hook merging exists.

Even when invocation becomes automatic, mutation remains proposal-only:

- each run may create a profile and audit;
- repeated findings may create an improvement proposal;
- a candidate must be replayed on held-out tasks;
- installation still requires explicit human approval and normal `maisternia`
  conflict handling.

## Run Artifacts

Each retrospective writes under:

```text
.agent-runs/retrospectives/<run-id>/
```

Expected files are:

```text
profile.md
audit.md
session-analysis.md
proposal.md       # only when evidence supports a change
index.md          # links the review package
record.json       # structured metrics, findings, proposals, and decisions
```

`record.json` follows the installed versioned schema. Improvement runs scan
these per-run records instead of scraping Markdown or appending to one shared
file, so concurrent sessions do not compete for a global history index.

Reports should preserve evidence references and exact commands without copying
unrelated source content. Projects that do not want these artifacts committed
should ignore `.agent-runs/`.

## Configuration Profiling

The profile separates static configuration footprint from measured runtime
usage.

Static evidence includes:

- file and resource counts;
- bytes by always-loaded and on-demand content;
- duplicate or contradictory instructions;
- skill activation overlap;
- MCP tool exposure and availability;
- hook count, trigger, and side effects;
- provider capability mismatches;
- managed, unmanaged, and shadowed files.

Runtime evidence includes input, output, cache, and reasoning tokens, model and
subagent calls, retries, tool calls, permissions, latency, and price when the
provider exposes them. Text-length estimates must not be presented as exact
provider token usage.

## Audit Evidence

The audit prefers deterministic and external evidence:

1. tests, type checks, linters, builds, and final environment state;
2. repository files, diffs, tool results, and pinned documentation;
3. external primary sources;
4. model judgment labeled as judgment.

High-risk disputed findings go to a human. Multiple instances of the same model
are not independent evidence. Pairwise judge order should be randomized, and
generator identity should be hidden where possible.

## Improvement Policy

The checked-in policy uses these defaults:

- proposal-only mutation;
- two equivalent findings before proposing a normal change;
- one critical escaped defect may create a proposal immediately;
- held-out replay is required;
- human approval is required;
- unknown metrics remain unknown.

The smallest enforceable intervention is preferred: deterministic checks first,
then discovery or documentation, one command or skill, context removal or lazy
loading, workflow gates, routing, and finally global instructions.

This order avoids turning every isolated failure into permanent prompt growth.

## Research Basis

The workflow is informed by, but is not directly validated as a complete
product by, the following research:

- [FActScore](https://aclanthology.org/2023.emnlp-main.741/) and
  [SAFE](https://arxiv.org/abs/2403.18802) motivate claim-level factual review.
- [AgentBoard](https://arxiv.org/abs/2401.13178) motivates separating final
  success from intermediate progress and trajectory evidence.
- [SWE-bench](https://arxiv.org/abs/2310.06770) motivates executable outcome
  checks for software tasks.
- [AI Agents That Matter](https://arxiv.org/abs/2407.01502) motivates joint
  quality and cost reporting, held-out evaluation, and reproducibility.
- [Judging LLM-as-a-Judge](https://arxiv.org/abs/2306.05685) and
  [LLM Evaluators Recognize and Favor Their Own Generations](https://arxiv.org/abs/2404.13076)
  motivate judge calibration, order controls, and identity blinding.
- [Large Language Models Cannot Self-Correct Reasoning Yet](https://arxiv.org/abs/2310.01798)
  motivates external evidence and human gates instead of unrestricted
  self-correction.
- [tau-bench](https://arxiv.org/abs/2406.12045) motivates repeated trials and
  policy-aware evaluation.

The trajectory labels, severity weights, token-efficiency thresholds, and
proposal promotion rules remain product hypotheses. Calibrate them against
human-reviewed runs before treating them as measurements.

## Try It

Inspect the complete preset:

```bash
maisternia preset show harness-improvement
maisternia preset plan --scope user --target codex harness-improvement
```

Apply it after reviewing the plan:

```bash
maisternia preset apply --scope user --target codex --yes harness-improvement
```

Then invoke this inside the configured harness after a completed task:

```text
/work-retrospective <explicit session export or current completed task>
```

Or run only the concrete session bottleneck analysis:

```text
/work-session-analysis <explicit session export or current completed task>
```
