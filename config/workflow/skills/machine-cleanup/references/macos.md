# macOS storage adapter

## Detection and inventory

Use `sw_vers` and a bounded native filesystem summary such as `df -k` for the
selected filesystem. `diskutil apfs list` may expose names and identifiers;
project only aggregate capacity/free data or obtain the narrow read first.
Record APFS container identity opaquely and count its free pool once. System
and Data volumes and firmlinked views can overlap; never add `/Users` and its
Data-volume view as separate storage. Native summaries may include purgeable
space differently from `df`; label both and explain the difference.

Start with System Settings → General → Storage. It offers native category
information and user controls. Before directory measurement, approve exact
roots/commands and privacy scope. Prefer shallow known application/tool roots
over whole-home recursion. On macOS, `du -x -k -d 1 <approved-root>` supplies
same-filesystem allocated summaries but is not a full physical-block oracle;
use no-follow defaults, verify components, and enforce supported timeout/output
budgets. Category names themselves may require aggregation before display.

Report applications, user data, developer projects, package caches, IDE output,
container backends, snapshots/backups, trash, and unknown system allocation.
Protected/denied directories stay unreadable; do not request Full Disk Access
or sudo merely to eliminate the unknown bucket.

## Cleanup choices

- Package/build candidates follow the package reference and native tool scope.
- Time Machine local snapshots and APFS snapshots are report-only; deleting
  them trades recovery history for space and belongs to native management.
- Xcode simulator/DerivedData, iOS backups, Photos libraries, Mail attachments,
  app containers, and downloads are not interchangeable caches. Classify exact
  tool-owned rebuildable outputs before any candidate; otherwise hand off.
- Trash emptying is irreversible, separate from cache cleanup. Prefer Finder's
  native review/empty action with exact scope chosen by the user; no raw fallback.
- Logs/system temp and broadly labelled “System Data” remain native handoff.
- Never recursively delete `/Library`, `/System`, `/private`, home, or unknown
  `~/Library` entries. Storage categories do not establish cache ownership.

OrbStack, Docker Desktop, and Colima/Lima data may live in sparse VM files.
Measure host allocated bytes, not virtual length; logical Docker reclamation
may precede physical host recovery. Read the Docker reference before a separate
backend compaction proposal. Do not manually modify VM disk files.

Verify native free space on the same APFS allocation domain and relevant tool
state. Do not reboot, restart apps, or drop filesystem caches to manufacture a
larger delta. Cache recovery means rebuild/download work, not an undo button.

Source: [Apple storage management](https://support.apple.com/en-nz/102624).
