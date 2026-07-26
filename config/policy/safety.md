# Safety Policy

The workflow never treats model or provider selection as authorization.

Explicit approval is required before:

- implementation without an approved handoff;
- permission escalation;
- commit;
- push;
- PR creation or update;
- destructive cleanup;
- secret access;
- production operations;
- destructive migration;
- real external side effects.

Managed configuration must not contain credentials, tokens, transcripts,
runtime databases, caches, logs, or raw environment values.

Configuration apply must show a plan, refuse conflicts, create backups, and
write atomically.
