---
description: Inspect a Maisternia preset plan without applying it
argument-hint: PRESET=<preset-id> [TARGET=codex]
---

Run `maisternia preset plan --target $TARGET $PRESET`.

Summarize:

- files that are unchanged, created, updated, kept, or conflicting;
- why every conflict exists;
- whether applying with `abort`, `keep`, or `replace` is appropriate;
- which existing files would be backed up by replacement.

Do not apply the preset. End with the exact reviewed apply command.
