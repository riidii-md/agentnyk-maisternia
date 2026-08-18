# Event Validation

`maisternia event validate` checks an untrusted event envelope against the
declarative workflow policy in the selected catalog.

```bash
maisternia event validate --repo /path/to/catalog event.json
```

Validation checks:

- the strict event JSON shape and schema version;
- bounded text, normalized artifact paths, and safe HTTP(S) URLs;
- common credential and private-key signatures;
- whether the event type exists in `config/workflow/triggers.json`;
- whether the initial phase has matching capability and routing policy;
- whether the external trigger remains read-only.

Successful output reports the event ID, configured initial phase, and
authority. The command does not:

- ingest or retain the event;
- create a task or context directory;
- choose or start a harness;
- transition a phase;
- grant approval or write authority.

Event validation is a configuration diagnostic. A downstream harness or
collaboration service may consume an event under its own security and storage
contract.
