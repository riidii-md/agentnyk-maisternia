# Linux storage adapter

## Detection and inventory

Use bounded native totals (`df -B1 <selected-filesystem>` where supported),
mount/filesystem metadata, and manager summaries. Project mount labels/paths
when they are personal. Distinguish user quotas, reserved blocks, tmpfs, bind
mounts, overlays, and container mounts; do not add alternate views of the same
allocation. tmpfs size is not SSD allocation.

After exact-root outside-workspace read approval, shallow
`du -x -B1 --max-depth=1 <approved-root>` may classify known roots. Check GNU
feature support first, reject symlinked components, and exclude nested mounts;
bind mounts on the same device need an explicit mount-boundary check beyond
`du -x`. Native directory totals do not attribute shared COW blocks exactly.

Report personal/project data, package-manager caches, container/VM allocation,
applications, system, logs/temp, snapshots/backups, trash, and unknown space.
Do not scan network/FUSE mounts or `/proc`, `/sys`, `/dev` to find disk candidates.
Permission failures become unreadable scope; root escalation is separate.

## Native maintenance choices

| Manager/category | Read-only evidence | Possible separately approved action | Preserve/verify |
| --- | --- | --- | --- |
| APT download cache | Approved cache allocated bytes; local `apt-get --help` | `apt-get autoclean` for obsolete archives or `apt-get clean` for the exact manager cache; `host.cache.prune`/`clear` | Privilege separately approved; no package autoremove; verify cache/free space and unchanged packages |
| DNF download cache | Manager cache root/size and installed version | A version-supported `dnf clean packages`; `host.cache.clear` | Keep installed packages/repository configuration; separate root authority |
| systemd journal | `journalctl --disk-usage` in approved read scope | Version-supported `journalctl --vacuum-time=<approved-retention>` or `--vacuum-size=<approved-budget>`; `host.log.prune` | Removes archived diagnostic history; retention selected by user; active files may remain; verify journal bytes/free space |
| Other package managers | Native help/status and exact cache scope | Native cache GC only after its command/scope are verified | Unknown semantics or missing dry-run becomes handoff |
| Temp/system storage | OS-managed aggregate information | Native retention/settings handoff | No generic directory age deletion or broad `/tmp` cleanup |
| Trash/snapshots/backups | Aggregate native information | Desktop/storage-manager handoff | User recovery data; report-only default |

Never combine sudo/privilege approval with the destructive category grant.
Package removal, system cleanup scripts, log truncation, service restarts,
filesystem repair, and snapshot deletion are outside ordinary cache cleanup.
Do not clear journal/system caches to suppress evidence of ongoing failures.

Verify the original filesystem's free bytes and category/native manager state;
active mounts/services must remain healthy. Concurrent writes and deleted-open
files can delay recovery. Detect open handles without printing process argv,
and retain rather than killing a process to reclaim them.

Sources: [APT manual](https://manpages.debian.org/stable/apt/apt-get.8.en.html),
[journalctl](https://www.freedesktop.org/software/systemd/man/latest/journalctl.html).
