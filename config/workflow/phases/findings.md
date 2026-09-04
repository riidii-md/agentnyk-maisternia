---
name: work-findings
description: Centralize explicitly selected retrospective packages, aggregate repeated harness findings, and prepare controlled Maisternia improvement proposals.
metadata:
  version: 0.1.0
---

# /work-findings - Centralize And Analyze Harness Findings

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible
explicit route, an active session route exists, or the exact
`.maisternia/work-routing.json` or
`${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise
continue locally without loading it. After loading, continue only with its
cleaned task.

Use this command in either import or analysis mode:

```text
/work-findings import <retrospective-directory> [<retrospective-directory> ...]
/work-findings analyze
```

Import only explicitly named retrospective directories. Never search all
provider histories. Follow the `Centralize Completed Packages` procedure in the
installed `session-retrospective` skill and the installed retrospective policy.
The central root is
`${XDG_STATE_HOME:-$HOME/.local/state}/maisternia/findings`; each current run is
stored under `runs/<provider>/<run-id>/` with a schema-valid `source.json`.

Never copy transcripts, provider histories, event streams, runtime databases,
credentials, tokens, or files outside the policy allowlist. Preserve a legacy
`record.json` byte-for-byte and label it `legacy` in `source.json`; do not invent
missing evidence or silently upgrade it to the current retrospective schema.

## Analyze The Central Store

Run analysis from the AgentnykMaisternia repository. Read only current
top-level run packages and count a `run_id` once. Verify `source.json` and the
listed SHA-256 checksums before using a package. Exclude corrupt, unresolved
collision, or unprovenanced packages and report why. Legacy packages may inform
hypotheses, but do not count them as schema-valid repeated evidence.

Group semantically equivalent findings across providers and repositories while
preserving run-level provenance, disagreements, accepted decisions, dismissed
findings, and unknown metrics. A normal improvement requires the configured
minimum number of independent, schema-valid runs. A single critical finding may
advance only under the policy's explicit exception.

For every candidate, identify the smallest owned Maisternia resource or preset
change, expected metric movement, affected providers, rollback, and a held-out
replay that was not used to create the proposal. The output is proposal only.
Do not waive the held-out replay requirement for a normal improvement. Do not
create or change a preset, skill, prompt, command, hook, MCP setting, or global
instruction without explicit human approval.

After approval, use `$work-run` in the AgentnykMaisternia repository, follow its
test-first and verification contract, and keep preset installation as a
separate opt-in action. Never treat approval to analyze as approval to edit,
commit, publish, or apply configuration.
