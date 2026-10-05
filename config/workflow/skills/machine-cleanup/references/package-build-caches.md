# Package and build caches

Load only rows for detected/requested tools. Discover the installed version,
resolved cache root, native selection semantics, active writers, and source
reproducibility before proposing anything. Do not execute project scripts or
download a CLI to discover its cache. Query individual non-secret settings;
never dump full package-manager configuration or credential-bearing URLs.

Directory names are evidence of location, not ownership or disposability.
Cached private/offline packages, local build outputs, virtual environments,
source checkouts, installed SDKs, and databases are not interchangeable.
Lockfiles do not guarantee that a vanished private source can be downloaded.
Unknown consumers/locks mean skip, not kill or force. Preserve offline/unique
artifacts unless deep-cleanup explicitly accepts their loss.

Every native GC/verify command that can write is a mutation, even if named
“verify,” “status,” or “maintenance.” Apply core approval, revalidation,
receipt, and verification to each invocation. Preview the manager's actual
selection, not just the parent directory size. A dry-run is useful only when
the installed version supports it, and is not a grant or a lock.

## Tool coverage

These are proposal directions, not commands to run automatically. Confirm
current installed help before choosing flags. Bind explicit cache roots where
supported; inherited environment/config overrides must not widen approved scope.
Windows remains native handoff only, including these rows.

| Tool | Inventory / native direction | Preservation and handoff boundary |
| --- | --- | --- |
| npm | Query only cache location; list metadata within approved scope. `npm cache verify` can GC and needs `host.cache.prune`. | Do not use force-required full clearing. Exact native npx cache removal only if this version supports it and each selected entry is reproducible. |
| pnpm | `pnpm store path`; native `pnpm store prune` for unused packages after checking consumers. | Shared store is not the sum of project `node_modules`. Running store server or incomplete dependency evidence means skip. |
| uv | `uv cache dir`; native `uv cache prune`, or `uv cache clean` for a disclosed exact package/cache selection. | Check active builds/installs and editable/local source origins; distinguish cached wheels from environments. |
| pip | Native `pip cache info` / metadata list; `pip cache remove` for explicit wheel patterns, or separately approved `pip cache purge`. | Purge includes HTTP and wheel cache. It does not authorize deleting environments or installed packages; protect private/offline wheels. |
| Homebrew | Query cache root; inspect version-supported cleanup dry-run and its full selection. | `brew cleanup` can include old installed versions as well as downloads. Cache-only approval cannot authorize keg/app removal. No autoremove/uninstall; mixed or unbounded selection means handoff. |
| Go | Query only `GOCACHE` / `GOMODCACHE`; cache-specific `go clean -cache` or `go clean -testcache` without package arguments. | Module cache clearing is separate, reproducibility-gated, not default. No project-output removal or Go environment dump. |
| NuGet | Native `dotnet nuget locals` for one named cache; separately approved `--clear` for that exact class. | Global packages may be an offline dependency store. No broad all-cache selection by default; skip locked/in-use entries. |
| Cargo | Inspect configured native automatic GC/version support and registry/git/build-cache classes. | Preserve source trees and unique downloaded crates. Do not raw-delete Cargo home or run `cargo clean` as generic host GC; project hooks/config and build outputs need separate scope. |
| Gradle | Inspect native cache retention policy and known inactive cache classes. | Native GC may run at daemon lifecycle boundaries; never start/stop a daemon for cleanup. No raw deletion of Gradle home, configuration, locks, or wrapper distributions in use. Otherwise report/handoff. |
| Poetry | Native cache list and version-supported clear for one named package/repository cache. | Cache is separate from environments, auth, config, and project source. No environment removal; unknown rebuild provenance means preserve. |
| Yarn | Detect Classic versus modern Yarn, global versus project cache, and native clean scope. | Project/zero-install caches may be tracked/offline assets. Keep them; do not apply one generation's command to another or run package-manager upgrades. |
| Maven | Report local repository storage and reproducibility evidence. | No raw `.m2` deletion or plugin execution to “clean” it. Credentials/config and locally installed artifacts are protected; handoff without a proven bounded native cache-only operation. |
| Playwright / Puppeteer | Report browser download storage and installed-tool native ownership/version selection. | Browsers may be used by multiple projects. Never run a downloading `npx` discovery command; no broad uninstall/all or raw browser-cache deletion. Exact native unused-version removal only with proven consumers/scope. |
| Xcode | Distinguish DerivedData/build products, archives, simulator runtimes, and device support. | Prefer native UI selection with exact scope; archives and runtime/device state are not disposable caches. Never delete simulator/user state as build GC. |
| Android | Distinguish Gradle caches, downloaded SDK packages, AVD images, and emulator writable state. | Installed SDKs are app/tool assets; AVD data is state. Native handoff for those; no SDK uninstall or emulator reset under cache authority. |
| CocoaPods | Native cache metadata and version-supported cache clean for exact reproducible pods. | Pods installed in projects and private/local pod sources are separate; preserve unknown/offline artifacts. No project cleanup. |
| Bazel | Report output/cache roots with installed native tooling, without building project targets. | `bazel clean` may remove project outputs or stop a server; do not use it as generic cache GC. Shared/remote cache and unknown native selection are report/handoff. |

For a tool without a proven bounded native cache operation, offer a specific
native-manager handoff with the measured category and preserved state. Do not
invent an arbitrary filesystem-delete fallback. Cache size is an upper bound,
not promised physical recovery. Verify native cache summary, dependent tool
health without rebuilding/downloading, and filesystem free delta after dispatch.

## Primary sources

- [npm cache](https://docs.npmjs.com/cli/v11/commands/npm-cache/)
- [pnpm store](https://pnpm.io/cli/store)
- [uv cache](https://docs.astral.sh/uv/concepts/cache/)
- [pip cache](https://pip.pypa.io/en/stable/cli/pip_cache/)
- [Homebrew manpage](https://docs.brew.sh/Manpage)
- [Go command reference](https://pkg.go.dev/cmd/go)
- [NuGet cache management](https://learn.microsoft.com/en-us/nuget/consume-packages/managing-the-global-packages-and-cache-folders)
- [Cargo cache configuration](https://doc.rust-lang.org/cargo/reference/config.html#cache)
- [Gradle directory layout / cleanup](https://docs.gradle.org/current/userguide/directory_layout.html)

For other rows, use installed native help or current primary documentation
before proposing a mutation; no unverified command syntax is implied here.
