# Installing AgentnykMaisternia

AgentnykMaisternia is pre-release. Building from the public source repository is
the supported installation method today.

Homebrew, GitHub release archives, and `go install ...@latest` are not available
yet. The project has no tagged release or public Homebrew tap, and its Go module
and release metadata still need to move from the former repository namespace.
Older commands that reference `kagi-labs` are obsolete.

## Requirements

- Git;
- Go 1.25.8 or newer;
- at least one supported coding-agent CLI if you want provider inspection or
  configuration: Codex, Claude Code, Antigravity, or Hermes.

The agent CLIs are not required to build Maisternia or browse its catalog.

## Build from source

Clone the current public repository and install the CLI:

```bash
git clone https://github.com/riidii-md/agentnyk-maisternia.git
cd agentnyk-maisternia
make install
```

`make install` embeds the current Git version, commit, commit date, and complete
configuration catalog. It installs `maisternia` under `GOBIN`, or under
`$(go env GOPATH)/bin` when `GOBIN` is empty.

If the command is not found after installation, add the Go binary directory to
your `PATH`.

## Verify the installation

```bash
maisternia --version
maisternia doctor
maisternia provider doctor all
```

`provider doctor` is read-only. It inspects known provider executables and
configuration roots without running a provider's own doctor command.

## First launch

Open the Admin terminal interface:

```bash
maisternia
```

`maisternia admin` is the explicit equivalent.

The first command that needs the catalog materializes the embedded catalog
under:

```text
~/.config/maisternia/catalogs/<content-sha256>/
```

The catalog is private, content-addressed, and installed atomically. The
installed binary does not depend on the source checkout.

Installation and first launch do not change coding-agent configuration. In
Admin, every preset installation asks for user-global or project scope, shows
the exact plan, and requires confirmation. When launched inside a Git
repository, Maisternia recommends that repository for project scope.

## Plan before applying

For non-interactive use, choose the scope and target explicitly:

```bash
maisternia preset plan \
  --scope project \
  --project /path/to/repository \
  --target codex \
  standard-work
```

Apply only after reviewing the plan:

```bash
maisternia preset apply \
  --scope project \
  --project /path/to/repository \
  --target codex \
  --yes \
  standard-work
```

Applying the same preset again updates changed resources and reconciles
resources removed from the preset. Exclusive unchanged targets are backed up
and removed, shared targets are retained, and local drift becomes a conflict.

`apply` aborts on conflicts by default. An explicit `--conflicts keep` or
`--conflicts replace` decision is required to continue. Replacement backs up
the existing target before writing.

For Codex workflow presets, restart Codex after apply so newly installed skills
appear in suggestions.

## Optional maintainability analyzers

The standard-work and multi-lens-review presets install review definitions but
never install external analyzers. To add pinned jscpd evidence for canonical
maintainability implementation review, inspect and explicitly confirm its
separate environment-only preset:

```bash
maisternia environment plan maintainability-review
maisternia preset apply --yes maintainability-review-tools
```

The plan shows `npm install --global jscpd@5.3.2`. Without `--yes`, no package
command runs. Review execution verifies the exact version with
`jscpd --version`; it does not install or upgrade the command. Existing
repository-bounded GitNexus evidence continues to come from the separately
installed `developer-context` pack.

## User-global installation

To make a preset available across projects for one provider:

```bash
maisternia preset plan --scope user --target codex standard-work
maisternia preset apply --scope user --target codex --yes standard-work
```

Provider home directories contain both declarative and runtime state.
Maisternia manages only the individual targets declared by a preset; it never
synchronizes a provider directory as a whole.

## Uninstall a preset

Use the same scope and target used for installation:

```bash
maisternia preset uninstall \
  --scope user \
  --target codex \
  --yes \
  standard-work
```

Uninstall uses recorded ownership, so it still works if the preset definition
has been removed from the active catalog. Shared targets are preserved, and
locally changed targets are handled as conflicts.

Environment packs describe host tools and are not automatically removed through
package managers or plugin hosts.

## Upgrade a source installation

From the checkout:

```bash
git pull --ff-only
make install
maisternia --version
maisternia doctor
```

The upgraded binary installs its matching content-addressed catalog without
overwriting older catalog versions.

## Remove the CLI

From the source checkout:

```bash
make uninstall
```

This removes the executable from the Go binary directory. It does not remove
managed provider configuration, install records, backups, or materialized
catalogs. Uninstall managed presets before removing the CLI if you want those
targets reconciled safely.

## Developer catalog override

Contributors editing catalog definitions can point one command at a checkout:

```bash
maisternia admin --repo /path/to/agentnyk-maisternia
```

Or save that developer override:

```bash
maisternia config set-repository /path/to/agentnyk-maisternia
```

Restore embedded-catalog discovery with:

```bash
maisternia config clear-repository
```

## Package availability

Do not substitute `riidii-md` into old Homebrew or `go install` examples. Those
commands become valid only after all of the following are complete:

1. the Go module and internal import paths use the current repository namespace;
2. release metadata targets the current repository;
3. a tagged release publishes archives and checksums;
4. a public Homebrew tap exists and receives a working cask.

The [project status](../README.md#project-status) is the user-facing source of
truth for availability.
