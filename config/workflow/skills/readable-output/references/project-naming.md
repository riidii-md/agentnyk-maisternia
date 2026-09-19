# mdmaid.desk Project Naming

Use this contract whenever a workflow registers or imports a document into
mdmaid.desk. The visible shape is:

```text
Repository / JIRA-ID (minimal AI feature text)
```

The repository and Jira ID are grounded identity. The producing AI supplies
only the concise feature text. mdmaid.desk owns the project record and assembles
the final label.

## Capability gate

Require mdmaid-desk 0.1.19 or newer and check `mdmaid-desk --version`. If it is
missing or older, preserve the Markdown artifact, do not register it with
project metadata, and report:

```text
npm install --global mdmaid-desk@0.1.19
```

Do not silently retry without project metadata. An older daemon may accept an
older request shape while discarding the intended identity.

## Ground the repository

Resolve the canonical current repository root without contacting a remote.
Inspect the configured origin URL locally when available and reduce it to a
credential-free identity such as `github.com/owner/repository`. Never pass or
persist a username, password, token, query, or fragment. Use the repository's
stable name as `<repository-name>`; do not use a worktree directory or branch
label as the repository name.

When adding a new workspace, include:

```text
mdmaid-desk workspace add <canonical-root> \
  --id <stable-workspace-id> \
  --repository <credential-free-identity> \
  --repository-name <repository-name>
```

If an existing workspace matches the canonical root, keep using it. Do not
re-add it merely to attach repository metadata because that can change its
configured artifact roots. Legacy metadata reconciliation is a separate,
explicit operation.

## Ground the task and generate the feature text

Use a Jira-style task ID only when it is explicit in the user's task, trusted
issue context, or another authoritative source. Normalize it to uppercase.
Branch names are not project names and are not sufficient evidence by
themselves for either the task ID or feature text.

When a stable Jira ID exists, generate a short feature phrase that distinguishes
the work inside the repository:

- use a compact noun phrase, normally two to six words;
- describe the feature, not the document kind or workflow phase;
- omit the repository name, Jira ID, provider name, branch, and status;
- avoid generic labels such as `Work`, `Task`, `Changes`, or `Implementation`;
- keep the wording stable throughout the task.

Pass both fields on every register or import for that task:

```text
--task <jira-id> --feature-name "<minimal feature text>"
```

The first accepted feature text wins. Later wording variants do not rename or
duplicate the project. If no grounded Jira ID exists, omit both `--task` and
`--feature-name`; do not invent an ID or turn a branch name into visible
identity.

## Receipt

Use the project ID and project name returned by mdmaid.desk as the delivery
receipt. Report that assembled name rather than reconstructing it locally.
