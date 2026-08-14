# Preset Collections

## Model

A collection is a source-scoped, tag-driven group of presets. Collections are
the general feature; a profession or role is one useful collection category.
The same model can represent workflows, team standards, or capability sets.

Preset tags are namespaced:

```json
{
  "id": "multi-lens-review",
  "tags": [
    "role/software-engineer",
    "capability/review"
  ]
}
```

Version 1 collection files live under:

```text
config/collections/<collection-id>.json
```

A collection declares stable display metadata and the tags every member must
contain:

```json
{
  "schema_version": 1,
  "id": "software-engineer",
  "name": "Software Engineer",
  "description": "A complete engineering workflow.",
  "match": {
    "all_tags": ["role/software-engineer"]
  }
}
```

Membership is resolved only from presets in the collection's own catalog
source. A built-in collection never absorbs matching presets from an external
source. External collections and their members use the same source-qualified
selectors as external presets.

## Discover and inspect

```bash
maisternia collection list
maisternia collection show software-engineer
maisternia collection validate all
```

The Admin Presets view uses `v` to switch between individual presets and
Collections. Collection details show the matching tags, exact members, common
providers, and union resource count. Search includes collection IDs, names,
descriptions, tags, members, and source metadata.

## Plan, apply, and uninstall

Collections use the same provider, user/project scope, conflict, plan, and
confirmation flow as presets:

```bash
maisternia collection plan \
  --scope user \
  --target codex \
  software-engineer

maisternia collection apply \
  --scope project \
  --project /path/to/repository \
  --target codex \
  --yes \
  software-engineer

maisternia collection uninstall \
  --scope user \
  --target codex \
  --yes \
  software-engineer
```

The supported provider list is the intersection declared by every member.
Unsupported providers are rejected explicitly; members are never silently
skipped.

Maisternia compiles the resolved members into one synthetic preset manifest and
builds one guarded configurator plan. Identical resources are deduplicated by
manifest resource ID, divergent target definitions retain the manifest's
existing validation and conflict checks, and the configuration files plus
install state are applied through the existing preflighted operation.

A collection has its own stable install owner. Direct preset ownership remains
separate, so uninstalling a collection cannot remove a shared target that the
user also installed directly. The stored resource ownership snapshot also
makes uninstall deterministic if tags change after installation. Reapplying a
collection reconciles its current resolved membership.

External source identity is retained after source removal, allowing a
previously installed external collection to be uninstalled safely by its
source-qualified selector.

## Current boundary

Version 1 collections accept configuration presets only. A collection that
matches an environment preset is rejected during validation. Host environment
installation has separate detection, confirmation, and lifecycle semantics;
it is not folded implicitly into a configuration collection apply.

## Authoring tags

Tags can be supplied when creating a preset or replaced when editing one:

```bash
maisternia preset create \
  --name "Team Review" \
  --tag role/software-engineer \
  --tag capability/review \
  team-review

maisternia preset edit \
  --tag role/security-engineer \
  --tag capability/review \
  team-review
```

Each `preset edit --tag` invocation replaces the preset's complete tag list
with the repeated values supplied by that command. Tags are validated,
deduplicated, displayed in preset details, and included in search.
