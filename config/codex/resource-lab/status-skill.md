---
name: maisternia-status
description: Use when the user asks what maisternia manages, what a preset would change, or why configuration is conflicting.
---

# AgentnykMaisternia Status

Inspect maisternia configuration without changing provider files.

1. Run `maisternia version`.
2. Run `maisternia config show`.
3. List presets with `maisternia preset list`.
4. For a named preset, run `maisternia preset plan <preset>`.
5. Explain create, update, unchanged, ignored, and conflict states simply.
6. For conflicts, distinguish:
   - keep existing: preserve the customized target and remember its checksums;
   - replace: back up the target and install the preset source;
   - abort: make no changes.

Never apply configuration unless the user explicitly asks and confirms the
reviewed conflict policy.
