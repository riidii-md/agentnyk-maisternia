# Reader Preference Resolution

Use preferences to avoid repeatedly asking the same low-value questions. Keep
them scoped, overridable, and explicit.

## Contents

- Precedence
- Preference dimensions and situation overrides
- View selection and delegation
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

## Delegation

The delegation preference has a policy, scope, and preferred target:

- policy `local`: shape the document in the current harness;
- policy `ask`: ask which harness should run it;
- policy `delegate`: delegate without asking when the preferred target exists;
- scope `explicit-command`: apply the gate to `/work-adapt-for-reader` only;
- scope `all-invocations`: also apply it when another workflow invokes the skill;
- preferred target `auto`, `current`, `codex`, `claude`, or `agy`.

When the whole delegation object is absent, explicit command use defaults to
`ask`, `explicit-command`, and `auto`; automatic or nested use defaults to
local. A missing scope also means `explicit-command`. This keeps old profiles
interactive at the deliberate command gate without interrupting workflows that
apply reader adaptation internally. Accept `codex-subagent` as a compatibility
alias for `codex` when reading an older profile.

Ask “Where should I run the adaptation?” and offer Here (current harness),
Codex, Claude, and AGY. `current` means no delegation. A named provider means a
fresh delegated run. Mark an available recommendation and make unavailable
choices visible. `auto` may select a native delegate, but it must not silently
cross a provider boundary.

Delegation changes who drafts or analyzes the content, not who owns delivery.
The coordinating harness verifies the result, writes the Markdown artifact, and
registers it with mdmaid.desk. If the target is unavailable, follow the explicit
request; otherwise fall back locally and disclose that fallback.

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
    },
    "delegation": {
      "policy": "ask",
      "scope": "explicit-command",
      "target": "auto"
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
