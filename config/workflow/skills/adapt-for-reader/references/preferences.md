# Reader Preference Resolution

Use preferences to avoid repeatedly asking the same low-value questions. Keep
them scoped, overridable, and explicit.

## Contents

- Precedence
- Preference dimensions and situation overrides
- View selection and shared work routing
- Persistence gate and profile locations
- Update procedure

## Precedence

Resolve conflicts from highest to lowest priority:

1. **Explicit current request** and explicit view or mode arguments.
2. Current task constraints, including required format and fidelity.
3. Matching project situation override.
4. Matching user situation override.
5. Project defaults or project instructions.
6. User defaults or user instructions.
7. Skill defaults.

A specific situation override beats a general default at the same scope.
Project preferences beat user preferences because they are more local. Never
let a profile weaken a safety, accuracy, accessibility, or repository rule.

## Preference dimensions

Store only communication preferences that change output:

- primary plain-language `view` or `auto`;
- legacy/internal `mode` only when maintaining an existing profile;
- conceptual depth: `high-level`, `working`, or `deep`;
- `density`;
- `answer_position`;
- `terminology`;
- `visuals`;
- `evidence` placement;
- `interaction`: ask when material, infer and state, or never ask;
- `layering`;
- optional first-pass time budget.

Do not store both `view` and `mode` at the same scope. Translate a legacy mode
to its plain-language view when revising a profile.

## View selection

The `view_selection` preference controls the selection gate separately from the
reader-contract `interaction` preference:

- `infer`: select from the reader contract without asking;
- `ask-when-ambiguous`: ask only when plausible views materially differ;
- `always-ask`: ask on every applicable invocation unless it already names a
  view or mode.

Its scope is `explicit-command` or `all-invocations`. With `explicit-command`,
always-ask applies to `/work-adapt-for-reader` while automatic skill use can
continue without interruption. When asking, show the six plain-language views
with one-line outcomes and mark the inferred choice as recommended.

## Shared work routing

Harness choice is not a readability preference. Resolve it through the
installed `work-routing` skill and the provider-neutral work-routing profile.
Use `/work-routing-preferences` for `local`, `ask`, or `delegate` policy,
ordered harnesses, and per-workflow overrides.

Treat a reader profile's legacy `delegation` object as migration input only.
For `work-adapt-for-reader`, the shared router may honor it when no new route
exists, normalize `codex-subagent` to `codex`, and suggest moving it to the
general routing profile. Never maintain both copies or remove the legacy value
without an approved migration diff.

Do not store inferred identity, disability, protected traits, medical details,
or other sensitive personal attributes. Record a functional requirement such
as “plain terminology” or “screen-reader-compatible structure” only when the
user explicitly requests it.

## Situation overrides

Use situation overrides for preferences such as:

- decision material: `decision`, recommendation first, evidence near claims;
- big-picture orientation: `big-picture`, high-level depth;
- technical teaching: `explanation`, deep depth, contextual orientation;
- progress update: `action-brief`, compact, current state first;
- policy or API material: `lookup`, domain-native terms, stable sections;
- rationale or retrospective: `story`, contextual answer position.

Match only explicit task or medium fields. Do not infer a situation from a
stereotype about the reader.

## Supported sources

Read active natural-language instructions first. When present, also support a
schema-valid `reader-profile.json` in the project Maisternia directory or the
user-scoped Maisternia configuration directory. The installed
`reader-profile.schema.json` defines the portable shape.

Do not require a JSON profile. A concise preference block in existing user or
project instructions is sufficient and often easier to maintain.

## Calibration and persistence

During calibration:

1. establish user or project scope;
2. identify high-frequency situations;
3. ask only about dimensions that materially change those outputs;
4. show proposed defaults and overrides;
5. show the exact destination and diff;
6. wait for explicit approval before writing.

Do not persist preferences inferred from prior conversations. Do not overwrite
an existing profile or instruction block without inspecting it and presenting
the change. If the user declines persistence, apply the preferences only to the
current task or session.

## Minimal profile example

```json
{
  "schema_version": 1,
  "defaults": {
    "view": "auto",
    "depth": "working",
    "density": "balanced",
    "answer_position": "auto",
    "terminology": "dual-label",
    "visuals": "auto",
    "evidence": "near-claim",
    "interaction": "ask-when-material",
    "layering": true,
    "view_selection": {
      "policy": "ask-when-ambiguous",
      "scope": "explicit-command"
    }
  },
  "situations": [
    {
      "id": "progress-updates",
      "when": {"task": "operate"},
      "overrides": {"density": "compact", "answer_position": "first"}
    }
  ]
}
```
