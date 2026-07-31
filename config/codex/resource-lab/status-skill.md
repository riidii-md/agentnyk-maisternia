---
name: agentctl-status
description: Use when the user asks what agentctl manages, what a preset would change, or why configuration is conflicting.
---

# Agentctl Status

Inspect agentctl configuration without changing provider files.

1. Run `agentctl version`.
2. Run `agentctl config show`.
3. List presets with `agentctl preset list`.
4. For a named preset, run `agentctl preset plan <preset>`.
5. Explain create, update, unchanged, ignored, and conflict states simply.
6. For conflicts, distinguish:
   - keep existing: preserve the customized target and remember its checksums;
   - replace: back up the target and install the preset source;
   - abort: make no changes.

Never apply configuration unless the user explicitly asks and confirms the
reviewed conflict policy.
