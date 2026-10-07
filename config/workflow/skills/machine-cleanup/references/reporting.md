# Storage accounting, scan limits, and receipts

## Two views of storage

Use native filesystem totals as the authoritative baseline, with observation
time and allocation domain. A marketed 1 TB device is decimal; `1 GB = 10^9`
bytes while `1 GiB = 2^30` bytes. Show bytes and one consistent display unit;
do not call the TB-to-TiB difference missing files.

The exclusive accounting ledger uses disjoint allocated categories:

| Category | Typical ownership | Cleanup disposition |
| --- | --- | --- |
| Personal data | Documents, media, user-selected data roots | Report-only |
| Applications | Installed apps and required resources | Native uninstall handoff |
| Development projects | Source, environments, working data | Report-only; build caches can be split once |
| Package/build caches | Native-manager download/build stores | Native GC candidate after active-use checks |
| Containers/VMs | Backend files, container runtime allocation | Docker/backend reference |
| OS/system | System allocation, protected files | Native OS handoff |
| Logs/temp | Proven manager-owned retention scopes | Separate native retention candidate |
| Snapshots/backups | Local snapshots and backup stores | Report-only; native handoff |
| Trash | Files awaiting irreversible deletion | Separate exact approval |
| Shared/unassigned | Blocks whose single-category owner cannot be established | No cleanup attribution |
| Unclassified/unreadable | Unscanned, denied, cancelled, or unassignable allocation | Explain uncertainty |
| Free/system reserved | Native free/reserved values | Report separately |

For one allocation domain, within the native rounding tolerance:

`capacity = native free + reserved + disjoint allocated categories + shared/unassigned + unclassified/unreadable + reconciliation drift`.

Residual is never negative. If measured children exceed the parent's used
allocation, stop reconciliation and flag overlap, unit mismatch, time drift,
or incompatible sources; do not subtract unknown data or scale children to fit.
APFS volumes may share a container's free pool; do not add their reported free
values. Distinct filesystems/quotas require distinct baselines.

Example in decimal GB, entirely invented: capacity 1,000; free 180; personal
200; apps 60; projects 110; caches 70; VM host allocation 160; OS/logs/trash
120; shared/unknown 100. This balances to 1,000. Docker's 90 GB logical
reclaim estimate is a non-additive overlay within the VM category, not another
90 GB on disk. No promise that deleting it yields 90 GB of host free space.

Count a child once, either in its parent or as a separated category. Deduplicate
hardlinks by file identity within each allocation domain. Apparent sparse file
length, clone/COW ownership, snapshot-retained blocks, Docker shared layers,
and VM backend allocation cannot be added interchangeably. If tools cannot
attribute physical blocks reliably, show directory totals as a non-additive
diagnostic view and put the physical difference in shared/unknown allocation.

## Measurement fields

For each observation: opaque domain/host/backend identity, category, source
tool/version, timestamp, approved scope, measure kind, byte count or range,
confidence (`measured`, `estimated`, `unknown`), and overlap caveat. Measure
kinds are allocated, apparent, native free/reserved, purgeable/shared,
logical-reclaimable, and estimated physical-reclaimable. Never sum incompatible
measure kinds. Report a before/after physical delta as observed, not necessarily
caused entirely by cleanup; concurrent writes and delayed reclamation matter.

The candidate overlay gives size, logical/physical estimates, age semantics,
dependency/lock/consumer state, recoverability, and retain/block reason.
Native “reclaimable” or “unused” is evidence, not a deletion decision.

## Scan budgets and privacy

Start with non-enumerating OS and tool summaries. Before approved traversal,
propose 120 seconds per collector, ten minutes per pass, one traversal at a
time, 250,000 entries per approved root, and at most 1 MiB of retained redacted
evidence per collector. Report progress at least every 15 seconds. Native
tools may need smaller limits; increases need a new scoped read decision.
Declare available I/O/rate limiting or explicitly say it is unavailable.

Use supported timeouts/cancellation and a count-limited walker where available.
An aggregate `du` command cannot enforce an entry cap by itself: disclose the
supported bounds rather than claiming all ceilings are enforced. If the user
requires a bound the harness cannot enforce, hand off or reduce scope. Budget
or cancellation returns partial totals and unknown scope; it never creates a
candidate whose safety evidence is incomplete. A destructive batch is limited
to 1,000 exact targets; larger manifests need separate batches and approvals.

No-follow, same-mount traversal is mandatory. Reject symlinked components and
nested mounts. Avoid filenames by default in both retained output and display;
suppression does not authorize collecting them. Exact-root outside-workspace
read approval precedes traversal; a deeper/personal-name read is separate.

## Preview and durable receipt

The transient approval UI contains exact targets, native filter/argv, protected
objects, version, amount/range, side effects, recovery cost, and verification.
Target equality covers command, identity, candidate/protected sets, policy, and
safety state. Approval expiry/use count is enforced by the live harness.

Use an approved surviving workspace artifact directory such as
`.agent-runs/machine-cleanup/<session-id>/receipt.jsonl`, or an exact approved
alternative. Prove non-overlap with every target, check all path components
without following symlinks, and use restricted local permissions. Creation,
append, and persistence/flush must be available before cleanup. No writable
durable destination means inventory/handoff only.

The minimal receipt contains: session ID; random per-receipt salt; timestamp;
policy/preview/safety digests; category/adapter/operation; machine, mount,
backend, and target fingerprint; exact native opaque object ID when safe;
approval ID/expiry; dispatch/consumption marker; per-target result; and bounded
verification. A salted fingerprint must be reconstructible from a fresh
inventory and the saved salt, not by persisting raw identity material.

Do not persist usernames, raw paths/home roots, filenames, device serials,
volume labels, image/repository names, process arguments, logs, environment,
auth/configuration payloads, or secrets. Hashes minimize disclosure but do not
make guessable identifiers secret; protect the receipt. The receipt itself is
not a full replay instruction. Keep raw exact identities transient in the UI.
An approval record alone does not prove a call was dispatched.

After interruption, read the dispatch boundary and fresh authoritative state.
An unknown dispatch remains unknown until reconciled; never replay it simply
because the UI or tool timed out. Filter-based GC may not identify every removed
member after a crash: record that limitation and do not invent per-item success.

A richer Markdown report is optional, with separate destination/disclosure
approval. Safe category totals can be included in it; persistent target names
need an explicit privacy choice. Never place receipts, transcripts, or real
machine observations in the distributed skill package or fixture catalog.
