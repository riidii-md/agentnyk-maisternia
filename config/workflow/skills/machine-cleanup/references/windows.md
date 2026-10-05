# Windows storage adapter

Use native drive totals via Settings → System → Storage or bounded projected
volume data from available Windows tooling. Specify volume identity, bytes,
observation time, and NTFS/ReFS allocation semantics. Do not equate logical
directory size, compressed/sparse allocation, hardlinks, recovery partitions,
shadow copies, reserved storage, and WSL virtual disks.

Directory/category inspection still requires exact-root read approval and a
separate personal-name/deep inspection choice. Unknown or denied volumes remain
unclassified. Do not recursively enumerate every drive or follow reparse
points/junctions to create a seemingly complete report.

V1 direct Windows mutation is native handoff only. Explain the proposed
category, size, recovery/disruption, and native Settings action so the user
can decide; do not silently execute PowerShell deletion, DISM, WSL compaction,
registry edits, or scheduled cleanup. Storage Sense settings can affect future
Downloads/Recycle Bin/cloud-file behavior; show these consequences and leave
scheduling unchanged unless separately requested.

WSL/Docker virtual disk apparent capacity and host allocated bytes differ.
Deleting Linux/Docker objects does not prove the VHDX shrank. Any supported
compaction is a separate explicit backend operation with shutdown/recovery
consequences; in v1 return the native handoff.

Verify the user's native action only after they report completion, using the
same volume totals and category observations. Package-manager inventory may be
reported, but the core Windows handoff limit overrides tool mutation examples.

Source: [Windows Storage Sense](https://support.microsoft.com/en-us/windows/experience/storage-filemanagement/manage-drive-space-with-storage-sense).
