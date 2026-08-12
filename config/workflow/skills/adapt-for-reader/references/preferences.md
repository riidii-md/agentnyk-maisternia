# Reader Preference Resolution

Use preferences to avoid repeatedly asking the same low-value questions. Keep
them scoped, overridable, and explicit.

## Contents

- Precedence
- Preference dimensions and situation overrides
- Persistence gate and profile locations
- Update procedure

## Precedence

Resolve conflicts from highest to lowest priority:

1. **Explicit current request** and explicit mode arguments.
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

- primary `mode` or `auto`;
- `density`;
- `answer_position`;
- `terminology`;
- `visuals`;
- `evidence` placement;
- `interaction`: ask when material, infer and state, or never ask;
- `layering`;
- optional first-pass time budget.

Do not store inferred identity, disability, protected traits, medical details,
or other sensitive personal attributes. Record a functional requirement such
as “plain terminology” or “screen-reader-compatible structure” only when the
user explicitly requests it.

## Situation overrides

Use situation overrides for preferences such as:

- decision material: `decide`, recommendation first, evidence near claims;
- learning material: `learn`, detailed, contextual orientation;
- progress update: `operate`, compact, current state first;
- policy or API material: `reference`, domain-native terms, stable sections;
- rationale or retrospective: `narrative`, contextual answer position.

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
    "mode": "auto",
    "density": "balanced",
    "answer_position": "auto",
    "terminology": "dual-label",
    "visuals": "auto",
    "evidence": "near-claim",
    "interaction": "ask-when-material",
    "layering": true
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
