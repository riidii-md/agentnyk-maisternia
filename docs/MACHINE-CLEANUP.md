# Machine storage inventory and cleanup

`machine-cleanup` is an opt-in skill package for explaining local disk use and
proposing individually approved cleanup. It is not an automatic system cleaner,
RAM optimizer, background job, or Go cleanup engine. Existing `work-cleanup`
still handles only task-owned artifacts and finalization.

## Installation and invocation

Preview the installation for your chosen provider before applying it:

```bash
maisternia preset show machine-cleanup
maisternia preset plan --scope user --target codex machine-cleanup
# Only after reviewing the configuration plan:
maisternia preset apply --scope user --target codex --yes machine-cleanup
```

Use `--scope project` for project-local installation; choose `claude`,
`antigravity`, or `hermes` instead of `codex` when appropriate. The preset has
no pipeline and is not part of `standard-work`. Reload the provider's skill
discovery in a fresh interactive session. The CLI installs configuration;
it does not inventory or clean the host.

- Codex: `$machine-cleanup inventory`; the prompt wrapper also selects the skill.
- Claude: `/machine-cleanup inventory` invokes the native skill.
- Antigravity: compatibility prompt and skill tree; v1 inventory/handoff only.
- Hermes: interactive native skill selection; v1 inventory/handoff only.

`inventory` is the default and stops after reporting. `cleanup` proposes native
rebuildable-cache and reproducible-image candidates. `deep-cleanup` adds
stateful/privileged choices with additional recovery/loss and privilege decisions.
Even “clean everything” is not a deletion grant. Each category/batch gets a new
detailed preview and exact one-use approval immediately before dispatch.

## Where did the disk capacity go?

The report starts from native filesystem capacity/free values and keeps one
ledger per allocation domain. It explains personal files, apps, projects,
package/build caches, containers/VMs, OS, logs/temp, snapshots/backups, Trash,
shared blocks, and unclassified/unreadable space. Declined scans remain unknown;
they are not zero or presumed deletable.

An invented decimal-GB example for a 1 TB allocation domain:

| Allocation | GB |
| --- | ---: |
| Free | 180 |
| Personal data | 200 |
| Applications | 60 |
| Development projects | 110 |
| Package/build caches | 70 |
| VM host allocation | 160 |
| OS/logs/temp/Trash (not yet separated) | 120 |
| Shared/unclassified | 100 |
| Capacity | 1,000 |

A separate candidate overlay explains what might be reclaimed, dependencies,
rebuild/download cost, and uncertainty. Docker logical usage inside the 160 GB
VM is not added again. Neither is a sparse VM's apparent file length. APFS
shared free pools are counted once; hardlinks, clones, snapshots, layers, and
mount overlaps must not inflate the physical ledger. A negative residual is an
accounting error to investigate, not “negative unknown data.” GB and GiB are
distinct units.

## Approval and preservation

The preset includes shared wording and document-delivery references.
Inline inventory and cleanup previews remain document-free. If an authorized,
redacted report is the basis of a human response, register its exact revision
before asking. Registration does not approve an inventory read or cleanup batch.
Keep disclosure, destination approval, fresh one-use grants, drift checks,
write-ahead receipts, and protected data rules.


Before an outside-workspace traversal, the user sees exact roots, metadata,
commands, duration, and available bounds. Personal filenames/deep scans require
an additional narrower decision. Scans are shallow, no-follow, same-mount,
cancellable, and explicitly partial when permission or budgets prevent coverage.

Before each cleanup, the preview states exact targets/native selection, occupied
size, logical and physical recovery estimates, preserved objects, consumers and
locks, command/version, privileges/network needs, consequences, recovery or
irreversible loss, receipt destination, and verification. Revalidation drift
invalidates the decision. There is no process killing, service stopping, forced
removal, arbitrary host filesystem deletion, or automatic retry after an unknown
dispatch.

Docker offers dangling, older unused (30-day default), and all-unused image
choices. “All unused” waives age protection only. Running or stopped container
references, keep labels, irreproducible/local-only images, and uncertain shared
dependencies remain protected. Image creation is not last-use time. Build cache
uses installed-version native last-access/retention selection; active writers
are skipped. Volumes are excluded ordinarily: deep cleanup needs one exact
classified volume, zero consumers, and consistent backup/restore evidence or
explicit acceptance of its loss.

Docker's logical recovery and observed host free-space change are reported
separately. Backend compaction is a separate native operation/decision. If it
requires stopping services or unsupported tooling, the skill offers a handoff.
Package managers use native GC only within proven scope; unknown or mixed
cache/state selection remains report-only or native handoff.

## Enforcement, receipts, and rollback

Maisternia's portable policy is not live native permission enforcement. Codex
and Claude mutations are conditional on proof of exact, one-use native grants
and a writable durable receipt. Without either, they also return inventory and
handoff only. Antigravity, Hermes, and Windows mutations are handoff-only in v1.
No headless/scheduled/delegated destructive cleanup is supported.

Minimal write-ahead receipts live at an approved surviving root excluded from
cleanup. They use opaque/salted identity projections, not raw private paths,
filenames, image names, labels, environment, logs, or credentials. A richer
saved report needs its own destination/privacy decision. Neither belongs in
Maisternia's distributed catalog or fixtures.

The new preset co-owns the existing canonical `approval-policy` resource. Its
eight host operations add two critical ask rules; existing denials and workspace
cleanup rules remain unchanged. Reapplying any owning preset from this catalog
can update that shared policy vocabulary even without installing this skill.
That does not activate the skill or grant mutation authority. Uninstall removes
exclusive skill/reference/wrapper resources and releases its policy ownership;
the policy stays while another installed preset owns it. Uninstall does not
remove added policy vocabulary from a co-owned file or undo completed cleanup.
Path, conflict, user-drift, backup, and confirmation checks remain in force.
