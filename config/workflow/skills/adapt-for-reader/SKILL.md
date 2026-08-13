---
name: adapt-for-reader
description: Adapt existing or generated text to a specific reader, purpose, time budget, and medium while preserving meaning and evidence. Use when a user asks to improve readability, rewrite or structure information for an audience, make output scannable, decision-ready, teachable, operational, or reference-friendly, invoke `$adapt-for-reader`, or apply saved readability preferences. Also use when choosing the right output form depends on what the reader must do; skip it for spelling-only or grammar-only edits where reader fit is not requested.
---

# Adapt for Reader

Optimize reader success, not word count or visual polish. Make it easier for the
intended reader to find, understand, evaluate, and use the information without
changing its meaning, evidence, uncertainty, or required detail.

## Contents

- Resolve the reader contract and preferences
- Apply the clarification and view-selection gates
- Use the shared work-routing result
- Choose and apply a reader view and depth
- Verify fidelity and reader success
- Deliver through mdmaid.desk and optionally calibrate preferences

## Resolve the reader contract

Infer these fields from the current request and active context:

- audience and relevant prior knowledge;
- desired result: big picture, decision, explanation, action brief, lookup, or
  story;
- conceptual depth: high-level, working, or deep;
- first-pass time budget;
- required decision, action, or understanding;
- medium and accessibility needs;
- fidelity and evidence requirements.

Do not turn this into a questionnaire when the context already answers it.

## Resolve preferences

Read [references/preferences.md](references/preferences.md) when user or project
preferences exist, conflict, or need calibration. Apply its precedence rules.
The explicit current request always wins.

Treat conversation patterns as evidence for the current response only.
Do not persist inferred preferences or sensitive personal attributes. Persist
a profile or instruction change only after the user approves its exact scope
and content.

## Apply the clarification gate

Ask one focused question only when an unknown audience, purpose, or fidelity
constraint would materially change the structure or content. Prefer:

> Who will use this text, and what should they be able to do after reading it?

Ask a second clarification question only if the first answer leaves a material
time-budget or precision tradeoff unresolved. Then apply the separately
configured view-selection gate. Combine the two interactions when one compact
question can resolve both.

Never ask merely to choose a style label, font category, heading count, or other
low-impact preference.

## Apply the view-selection gate

Resolve `view_selection` from the active preferences. An explicit `view` or
legacy `mode` always bypasses this gate.

- `infer`: select silently from the reader contract.
- `ask-when-ambiguous`: ask when plausible views materially change the result.
- `always-ask`: ask for every applicable invocation within its configured
  `explicit-command` or `all-invocations` scope.

When asking, offer the six plain-language outcomes from
[references/modes.md](references/modes.md), mark the inferred choice as
recommended, and do not expose unexplained internal names. For ambiguous
requests such as “help me understand,” prefer:

> Do you want the big picture, enough mechanics to work with it, or a deep explanation?

## Use shared workflow routing

When `/work-adapt-for-reader` invoked this skill, use the route already resolved
by the installed `work-routing` skill. Do not implement another harness picker,
availability check, disclosure rule, or fallback. Direct or nested skill use
stays in the current harness unless its caller explicitly resolved a general
all-invocations route.

If an explicit `/work-adapt-for-reader` run stayed local at its lazy gate and
preference resolution then finds a legacy reader-profile `delegation` object,
load `work-routing` and pass that object as compatibility input. This conditional
migration path is not a second picker. Without a general route or that legacy
object, keep adaptation local and do not ask where to run.

A delegated harness may draft, analyze, or independently structure the source.
The coordinating harness remains responsible for fidelity verification, the
final Markdown artifact, and mdmaid.desk registration.

## Choose a view and depth

Read [references/modes.md](references/modes.md) for every transformation. Select
one primary view, map it to the internal mode, and apply the requested conceptual
depth plus plain-language or accessibility modifiers. Use the profile's matching
situation override when one exists.

Do not force answer-first structure onto learning or narrative material when
context-first organization better serves the reader.

## Build the message hierarchy

Before formatting, identify:

1. governing message or orientation;
2. supporting claims and their relationships;
3. evidence, uncertainty, and alternatives;
4. required action or decision;
5. detail that belongs in a deeper layer.

Use claim-led headings where they improve scanning. Preserve coherent prose for
causal reasoning and qualifications. Use lists for actual sets or procedures,
tables for repeated-field comparisons, and diagrams only for relationships that
become easier to inspect spatially.

## Transform in layers

When length warrants it, make each layer add information rather than repeat it:

1. **Orient:** the answer, governing question, warning, or mental model.
2. **Scan:** the important claims, evidence, risks, and next action.
3. **Understand:** coherent explanation, examples, and boundaries.
4. **Verify:** sources, uncertainty, alternatives, and technical detail.

Keep related labels, evidence, and visuals close together. Remove decorative
detail that competes with the main message. Read
[references/principles.md](references/principles.md) when selecting a
representation, adjudicating a style rule, or reviewing quality.

## Preserve integrity

- Keep facts, constraints, caveats, and provenance intact.
- Distinguish verified facts, interpretation, proposal, and unknowns.
- Define unfamiliar terms; retain precise domain terms when simplification
  would distort them.
- Prefer active voice when responsibility matters, not as an absolute rule.
- Use readability formulas only as diagnostics, never as the quality target.
- Do not invent missing evidence or make uncertain claims sound settled.

## Verify the result

Check the transformed output against the reader contract:

- Can the reader locate the governing message quickly?
- Are important relationships and references explicit?
- Does the depth match the reader's knowledge and time?
- Does each table or diagram perform a real cognitive job?
- Can the reader identify evidence, uncertainty, and the next action?
- Was anything removed that changes meaning or usability?

## Deliver through mdmaid.desk

Always write the complete adapted result as a standalone Markdown artifact,
even when the terminal response could contain it. Use the current repository
root, or the current directory outside a repository, and create:

```text
.agent-runs/readability/<timestamp>-adapted.md
```

When `mdmaid-desk` is available:

1. Use `MDMAID_DESK_WORKSPACE` when explicitly configured.
2. Otherwise match the canonical current root in `mdmaid-desk workspace list`.
3. If it is absent, add the current root once with `mdmaid-desk workspace add`,
   using a stable collision-safe ID derived from the root.
4. Run `mdmaid-desk register <artifact.md> --workspace <id> --kind <kind>
   --attention review`. Prefer `decision`, `definition`, `progress`, or `brief`
   when the selected mode makes the kind clear.

Registration sends the document to the desk; it does not imply approval and
does not require starting the TUI or web client. If the CLI is unavailable or
registration fails, preserve the Markdown artifact and report its path, the
failure, and an exact retry command. Do not substitute a temporary-only file or
an HTML-only renderer.

Return a short terminal summary with the Markdown path and desk registration
status. Do not duplicate the complete document in chat unless the user asks.
Mention the selected mode or a key assumption only when useful.
