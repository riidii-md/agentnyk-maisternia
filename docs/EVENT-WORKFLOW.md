# Event Workflow Compatibility Note

Maisternia's experimental event-ingestion and task-state workflow was removed.
This path remains only so older links explain the migration instead of leading
to obsolete runtime instructions.

Use:

- [Event validation](EVENT-VALIDATION.md) for the remaining read-only event
  envelope diagnostic;
- [Runtime-boundary migration](RUNTIME-BOUNDARY-MIGRATION.md) for removed
  commands, legacy local data, and preset upgrade guidance;
- [Configuration boundary](CONFIGURATION-BOUNDARY.md) for current ownership.

Maisternia does not ingest events, create tasks, transition phases, dispatch
agents, or own collaboration history.
