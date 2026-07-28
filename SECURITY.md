# Security

## Supported Version

The project is pre-release. Security fixes are applied to the latest `main`
branch.

## Reporting

Report vulnerabilities privately to the Kagi Labs repository maintainers. Do
not open a public issue containing credentials, personal configuration, or an
exploit against a real home directory.

## Security Model

The configurator writes files under agent-specific home-directory roots. A
malicious or incorrect manifest could otherwise overwrite arbitrary user data.

Current controls:

- provider root allowlist;
- relative-path validation;
- traversal rejection;
- regular-file source requirement;
- managed file size limit;
- duplicate-destination rejection;
- destination symlink rejection;
- unmanaged-file conflicts;
- installed-checksum drift detection;
- explicit `--yes`;
- source and target revalidation before apply;
- backups before update;
- atomic writes;
- install-state permissions;
- strict event parsing with unknown-field rejection and size limits;
- unsupported-trigger, unsafe-path, credential-pattern, and URL-userinfo
  rejection;
- read-only trigger authority enforcement;
- source-event separation from generated runner context;
- event idempotency checksums;
- private task-state permissions, atomic replacement, and append-only logs;
- per-event and per-task writer leases with stale-process recovery on Unix.

## Secret Handling

Never commit:

- API keys;
- OAuth tokens;
- provider credentials;
- SSH keys;
- `.env` files;
- transcripts;
- auth or runtime databases.

Configuration should reference secrets by environment or keychain name rather
than storing values.

## Review Priorities

Changes to these areas require focused security review:

- path normalization;
- symlink behavior;
- manifest parsing;
- state and backup paths;
- atomic writes;
- structured settings merge;
- plugin installation;
- runner subprocess execution;
- permission routing;
- normalized event validation and idempotency;
- durable task state, indexes, and leases;
- runner capability resolution and dispatch.
