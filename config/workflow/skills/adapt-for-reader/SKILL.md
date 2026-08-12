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
- Apply the clarification gate
- Choose and apply a reader mode
- Verify fidelity and reader success
- Deliver and optionally calibrate preferences

## Resolve the reader contract

Infer these fields from the current request and active context:

- audience and relevant prior knowledge;
- reader task: scan, decide, learn, operate, look up, or follow a narrative;
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

Ask a second question only if the first answer leaves a material time-budget or
precision tradeoff unresolved. Otherwise choose a reasonable mode, continue,
and state the assumption only when it helps the user evaluate the result.

Never ask merely to choose a style label, font category, heading count, or other
low-impact preference. An explicit `mode` bypasses mode clarification.

## Choose a mode and depth

Read [references/modes.md](references/modes.md) for every transformation. Select
one primary mode and optional plain-language or accessibility modifiers. Use the
profile's matching situation override when one exists.

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

Return the adapted text or artifact first. Mention the selected mode or a key
assumption only when useful. Do not add a self-congratulatory explanation of
the formatting work unless the user asks for it.
