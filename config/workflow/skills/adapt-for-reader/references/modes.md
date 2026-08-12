# Reader Modes

Choose one primary mode from the reader's task. Treat plain language,
accessibility, density, and evidence placement as modifiers rather than separate
modes.

## Contents

- Mode selection
- `scan`, `decide`, `learn`, `operate`, `reference`, and `narrative`
- Cross-mode modifiers
- Choosing a structured notation

## Mode selection

| Mode | Choose when the reader must | Governing orientation |
|---|---|---|
| `scan` | find the important point quickly | answer or key state first |
| `decide` | compare options and make a choice | recommendation first |
| `learn` | build a reliable mental model | question, premise, or model first |
| `operate` | coordinate work or respond to change | current state first |
| `reference` | retrieve exact information repeatedly | stable navigation first |
| `narrative` | follow context, rationale, or events | coherent progression first |

When tasks overlap, choose the mode for the reader's immediate action and use a
deeper layer for the secondary task. For example, a decision document can put
the recommendation first and place a learning-oriented explanation below it.

## `scan`

Use for executive summaries, handoffs, overviews, and time-limited reading.

Shape:

1. governing message in one to three sentences;
2. three to five claim-led sections or fields;
3. evidence, risk, and next action where relevant;
4. optional deeper detail.

Prefer short paragraphs and strong information scent. Do not remove caveats
that change the governing message.

## `decide`

Use for approvals, recommendations, tradeoffs, and risk acceptance.

Shape:

1. recommendation and decision needed;
2. reasons and decisive evidence;
3. options compared on consistent criteria;
4. risks, uncertainty, and reversibility;
5. next action, owner, or approval.

Use a table when at least three options or repeated criteria need comparison.
Keep evidence near the claim it supports.

## `learn`

Use for explanations, onboarding, tutorials, and unfamiliar concepts.

Shape:

1. motivating question or governing mental model;
2. prerequisites and definitions;
3. components and relationships in a coherent sequence;
4. worked example;
5. boundary, counterexample, or comprehension check;
6. sources and deeper detail.

Supply more connective explanation for novices. Remove redundant scaffolding
for experts. Do not mistake ease of reading for durable learning.

## `operate`

Use for status, incidents, process updates, and coordination. Borrow the
fixed-field function of operational reports without importing their jargon.

Shape:

```text
Current state
What changed
Evidence
Risk or blocker
Next action
Decision needed
```

Omit empty fields. Keep owners and deadlines only when known. Do not invent an
estimate to make the report look complete.

## `reference`

Use for policies, specifications, glossaries, API references, and procedures
that readers revisit.

Shape:

1. scope and navigation;
2. stable terminology and definitions;
3. predictable repeated sections;
4. exact tables or procedures;
5. exceptions and cross-references near the rule;
6. change or version information when relevant.

Optimize retrieval and consistency rather than narrative flow. Avoid hiding
critical rules behind progressive disclosure.

## `narrative`

Use for rationale, history, retrospectives, persuasion, and event sequences
where context changes the meaning of the conclusion.

Shape:

1. establish the situation and stakes;
2. preserve causal or chronological continuity;
3. introduce evidence where it becomes meaningful;
4. end sections with the implication or transition;
5. summarize the conclusion and action when relevant.

Use fewer headings than `scan` mode when fragmentation would break flow.
Answer-first is optional.

## Cross-cutting modifiers

- `density`: `compact`, `balanced`, or `detailed`.
- `terminology`: `plain`, `dual-label`, or `domain-native`.
- `visuals`: `auto`, `prefer`, or `avoid`.
- `evidence`: `inline`, `near-claim`, or `appendix`.
- `answer_position`: `auto`, `first`, or `contextual`.
- `layering`: enabled or disabled.

Accessibility is always a constraint, never an aesthetic variant. Preserve
semantic headings, meaningful link text, text alternatives, contrast, resize,
reflow, and a plain-text fallback for essential diagrams.

## Choosing a structured notation

Military-style templates are useful when omission, coordination, or confirmation
costs matter. Their value comes from a shared schema and predictable fields,
not from acronyms or a commanding tone.

- Use a BLUF-like opening when a decision, result, or warning must be found
  immediately.
- Use SITREP-like fixed fields for current state, change, evidence, risk, and
  next action.
- Use a SMEAC-like sequence for coordinated execution: situation, objective,
  execution, support, and communication or control.
- Add readback, backbrief, or an explicit confirmation field when mutual
  understanding matters more than passive comprehension.

Use the domain's actual notation only when the reader already shares it or asks
for it. Otherwise translate the function into plain labels, omit irrelevant
fields, and define unavoidable acronyms. Never use a military template merely
to make ordinary prose appear decisive.
