# Docker images, build cache, volumes, and backend

## Bind identity before inventory

Resolve effective context/host selection, canonical endpoint/socket, daemon ID,
builder, and VM/backend. Read only projected IDs/states, versions, mount edges,
and selected keep/ownership labels; do not print full inspection, environment,
logs, TLS material, credentials, or interpolated Compose configuration. A local
context name or Unix socket alone is insufficient if it forwards to another
engine. Classify `local-non-production | remote | production | unknown` from
trusted operator evidence and endpoint/daemon identity. Unknown is retained.

Never inventory a remote daemon without separate read authority; never mutate
remote, production, unknown, or changed identity. Explicitly bind every Docker
invocation to the proven endpoint/context and every buildx invocation to the
proven builder; do not rely on changing CLI defaults.

Use `docker system df` (project verbose evidence only with approval) for logical
sizes/shared/reclaimable data. Buildx cache lives at the selected builder, which
may differ from the engine; identify it independently. Containers and mounts are
dependency evidence, not cleanup candidates in this v1 adapter. Never stop or
remove containers to make images/volumes appear unused.

## Image choice and protected set

Ask the user to choose, showing counts and logical/physical estimates:

| Policy | Candidate rule | Retention |
| --- | --- | --- |
| Dangling only | Untagged, unreferenced images | All protected rules below still apply |
| Old unused (recommended) | Unreferenced and creation time older than 30 days | Recent, protected, or unknown images retained |
| All unused | Unreferenced without creation-age limit | Only the age protection is explicitly waived; all other protections remain |

Creation time describes image age, not last use. Explain that Docker does not
provide a generally reliable image last-use timestamp. Protect every image
referenced by running or stopped containers, keep-labelled images, local-only
or irreproducible builds, and shared bases whose dependency safety is unclear.
Attest reproducibility through a retained registry digest or build definition
and source; a familiar repository/tag is not proof. A dangling local build can
be valuable. An exact loss decision for irreproducible data belongs to
deep-cleanup, never the ordinary all-unused choice.

Enumerate immutable IDs/tags and protected consumers into the preview. Prefer
an exact removal such as `docker --context <bound-context> image rm --no-prune <approved-ref-or-id>`
after revalidation, no force. Reject unintended tag/parent deletion; one tag
removal may recover no bytes if other tags/layers retain allocation. Maximum
1,000 exact targets per approved command; partial success is recorded per item.
Do not derive executable shell from an image name or label.

Native image-prune confirmation does not enumerate the exact deletion set.
Default to exact removal. A filter is permitted only if the preview can prove
its bounded membership and protected exclusions at dispatch; negative label
filters and creation-age filters alone do not satisfy that proof. If the harness
cannot enforce it, hand off. Map exact image removal to
`host.docker.image.remove`; do not reuse task-owned Docker operations.

## Build cache

Inventory the proven builder with supported buildx disk-usage output, including
record ID, size, shared/private status, in-use state, and last access where
available. In-use and unknown consumers remain protected. Cache age refers to
last access where the installed tool documents it, unlike image creation age.

Check installed `docker buildx prune --help` for `--filter`, `--max-used-space`,
`--min-free-space`, and `--reserved-space`. Preview exact IDs where supported,
or a bounded last-access/storage-budget policy with protected exclusions and
uncertainty. Selection rules must not admit newly created/in-use/protected
entries between approval and dispatch; if that bound cannot be enforced,
remain inventory/handoff only.

A supported candidate template is
`docker --context <bound-context> buildx --builder <bound-builder> prune --filter until=<approved-age> --reserved-space <approved-reserve>`.
Show the concrete argv before approval; no unbounded prune, `--all`, or force.
Do not imply dry-run exists or that Docker's confirmation is the user's exact
approval. Map to `host.docker.build_cache.prune`. Account for shared cache/image
layers once and verify builder usage plus actual host free-space change.

## Volumes

Volumes are excluded from ordinary Docker cleanup and all unused image policies.
An unreferenced volume may contain the only copy of database, object-store,
Redis, NATS, model, or development data. Named and anonymous volumes require
the same dependency/ownership scrutiny. Never use aggregate volume prune or
container removal with implicit volume deletion.

Deep-cleanup may propose one exact zero-consumer volume after type/owner/data
classification. Model/dependency stores can be rebuildable; database, queue,
object storage, and unknown content are stateful. Establish application-aware
consistent backup and restore proof, or obtain the user's explicit acceptance
of exact irreversible data loss. No generic live copy counts as a valid database
backup. No verified recovery/loss decision means retain.

The exact volume ID/name is transient in the UI. Re-enumerate all consumers and
recreation controllers before dispatch; locks or unknown writers block it. Use
one exact `docker --context <bound-context> volume rm <approved-volume>` with no
force and map to `host.docker.volume.remove`. A command may only run under the
core exact-enforcement and durable-receipt contract.

## Physical backend recovery

Detect Docker Desktop, OrbStack, Colima/Lima, WSL, or native storage. Report
logical objects, host allocated VM-file bytes, virtual/apparent length, and
free space separately. Avoid duplicate accounting of layers and the VM file
that contains them. Never sum daemon totals into host disk categories.

Backend compaction/reclaim is separate from object deletion and maps to
`host.docker.backend.compact`. Use only a detected version's documented native
reclaim function; no raw disk-file edits, truncation, forced shutdown, or guessed
command. If compaction needs restart/stop, this v1 workflow hands off because
it does not authorize stopping services. Verify unchanged engine/container
health and the original host filesystem delta after any supported action.

Sources: [image prune](https://docs.docker.com/reference/cli/docker/image/prune/),
[system df](https://docs.docker.com/reference/cli/docker/system/df/),
[buildx prune](https://docs.docker.com/reference/cli/docker/buildx/prune/),
[volumes](https://docs.docker.com/engine/storage/volumes/).
