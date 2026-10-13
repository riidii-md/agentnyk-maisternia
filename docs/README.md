# Project Document Map

Use this map to find context for a task. Read the relevant sections and follow
concrete references. The map points to existing sources; it does not replace
repository instructions or require every document to be read.

| Topic | Existing sources | Use when |
|---|---|---|
| Product goals and scope | [Project overview](../README.md), [Configuration boundary](CONFIGURATION-BOUNDARY.md) | Defining user outcomes, supported behavior, or product exclusions. |
| Architecture | [Configurator architecture](CONFIGURATOR.md), [Workflow configuration](WORKFLOW.md), [Provider adapters](PROVIDERS.md) | Changing components, contracts, provider mappings, or configuration ownership. |
| Design and interaction | [Admin interface](ADMIN.md), [Approval policy](APPROVAL-POLICY.md), [Installation](INSTALLATION.md), existing code under `internal/admin` | Changing TUI navigation, CLI behavior, previews, conflicts, or confirmation flows. A separate visual design system is not defined. |
| Security | [Security model](../SECURITY.md), [Configuration boundary](CONFIGURATION-BOUNDARY.md), [Approval policy](APPROVAL-POLICY.md) | Changing paths, files, permissions, secrets, installation, or trust boundaries. |
| Code conventions | [Repository instructions](../AGENTS.md), [Contributing](../CONTRIBUTING.md), affected code and sibling tests | Implementing or reviewing Go code and configuration contracts. |
| Testing | [Repository checks](../AGENTS.md#verification), [Contributing](../CONTRIBUTING.md), [CI](../.github/workflows/ci.yml), [Test review](TEST-REVIEW.md) | Selecting checks, changing test evidence, or evaluating workflow contracts. |
| Task decisions and progress | The selected task's direction, plan, decision receipts, and `.agent-runs/work/<task-id>/progress.md` | Resuming a task or identifying approved scope and unfinished work. These local artifacts are not committed by default. |

Current implementation must be checked against code and test evidence. Documents
may contain implemented behavior, proposals, or historical decisions. Keep those
statuses distinct. [Architectural direction](ARCHITECTURAL-DIRECTION.md) describes
a workflow gate; it is not a replacement for the implemented architecture above.

If sources disagree, report the conflict and inspect the affected code and
accepted decisions. Do not treat implemented behavior as permission to weaken a
safety requirement. If evidence is missing or inaccessible, record the unknown.
Update affected sources during the approved task, then verify their descriptions,
commands, examples, and links against the resulting behavior.
