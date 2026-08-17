# Installing AgentnykMaisternia

## Requirements

AgentnykMaisternia is currently hosted in a private GitHub organization. Every
installation method requires access to `kagi-labs/agentnyk-maisternia`.

Verify GitHub CLI authentication:

```bash
gh auth status
gh auth setup-git
```

## Availability

Building from source is available immediately. Homebrew, `go install
...@latest`, and GitHub release archives require at least one tagged release.
Homebrew also requires the one-time private tap setup in
[Releasing AgentnykMaisternia](RELEASING.md).

Every installation method produces the same self-contained executable. The
versioned configuration catalog is embedded in the binary; a separate source
checkout is not required at runtime.

## Homebrew

The Homebrew cask installs release binaries from the private
`kagi-labs/agentnyk-maisternia` repository. These commands work after the first tagged
release has published `Casks/maisternia.rb` to `kagi-labs/homebrew-tap`.

Add the tap:

```bash
brew tap kagi-labs/tap
```

Install:

```bash
HOMEBREW_GITHUB_API_TOKEN="$(gh auth token)" \
  brew install --cask kagi-labs/tap/maisternia
```

Upgrade:

```bash
brew update
HOMEBREW_GITHUB_API_TOKEN="$(gh auth token)" \
  brew upgrade --cask maisternia
```

Uninstall:

```bash
brew uninstall --cask maisternia
brew untap kagi-labs/tap
```

The token is passed only to the `brew` process in these examples. Do not commit
tokens or put them in repository configuration.

Homebrew 5.1.14 and later hide sensitive environment variables while loading a
cask. The generated cask uses a download strategy that retrieves
`HOMEBREW_GITHUB_API_TOKEN` only when the private release asset is downloaded.

## Go Install

Configure Git authentication, bypass the public Go proxy for the private
organization, and install the latest tagged release:

```bash
gh auth setup-git
GOPRIVATE=github.com/kagi-labs/* \
  go install github.com/kagi-labs/agentnyk-maisternia/cmd/maisternia@latest
```

The binary is installed under `GOBIN`, or under `$(go env GOPATH)/bin` when
`GOBIN` is empty. Ensure that directory is in `PATH`.

Upgrade by running the same command again. Uninstall:

```bash
GOBIN="$(go env GOBIN)"
if [ -z "$GOBIN" ]; then
  GOBIN="$(go env GOPATH)/bin"
fi
rm -f "$GOBIN/maisternia"
```

## GitHub Release Binary

Download the archive matching the operating system and architecture:

```bash
gh release download \
  --repo kagi-labs/agentnyk-maisternia \
  --pattern "maisternia_*_darwin_arm64.tar.gz" \
  --pattern checksums.txt
```

Available release targets:

| Operating system | Architectures | Archive |
|---|---|---|
| macOS | `amd64`, `arm64` | `.tar.gz` |
| Linux | `amd64`, `arm64` | `.tar.gz` |
| Windows | `amd64`, `arm64` | `.zip` |

Verify the selected archive against `checksums.txt`, extract it, and install the
binary in a directory on `PATH`.

## Build From Source

```bash
git clone git@github.com:kagi-labs/agentnyk-maisternia.git
cd agentnyk-maisternia
make install
```

`make install` embeds the current Git version, commit, and commit date, then
installs to the Go binary directory.

Other useful targets:

```bash
make build
make test
make verify
make uninstall
```

## Verification

```bash
maisternia --version
maisternia doctor
maisternia provider doctor all
```

Open the Admin interface:

```bash
maisternia
```

`maisternia admin` is the explicit equivalent. The first command that needs the
catalog materializes it under:

```text
~/.config/maisternia/catalogs/<content-sha256>/
```

The catalog directory is private, content-addressed, and installed atomically.
Upgraded binaries can therefore install their matching catalog without
overwriting an older version or depending on the location of a Git checkout.

Installation and first launch do not apply managed agent configuration. In
Admin, every configuration preset asks for user-global or project scope, shows
the exact plan, and requires confirmation. When launched inside a Git repository,
Maisternia recommends and prefills that repository for project scope.

For non-interactive use, configuration preset planning and apply require an
explicit scope:

```bash
maisternia preset plan --scope user --target codex standard-work
maisternia preset apply --scope project --project /path/to/repo \
  --target codex --yes standard-work
maisternia preset uninstall --scope user --target codex --yes standard-work
```

Applying the same configuration preset again updates changed resources and
reconciles resources removed from the preset. Exclusive, unchanged targets are
backed up and removed; shared targets are retained; local drift becomes a
conflict. `preset uninstall` uses recorded ownership and therefore still works
when the preset definition has already been removed from the catalog.

For Codex workflow presets, reapplying migrates managed workflow files from the
unsupported `.codex/commands/` layout to native `.codex/skills/` entries plus
`.codex/prompts/` compatibility shims. Restart Codex after apply. Provider
inspection reports remaining legacy `work*.md` command files as degraded so
unmanaged leftovers can be reviewed explicitly.

Environment-only presets remain local-machine scoped because they install host
tools rather than provider configuration.

## Optional Source Catalog Override

Developers editing catalog definitions can point Maisternia at a checkout for
one command:

```bash
maisternia admin --repo /path/to/agentnyk-maisternia
```

Or save that developer override:

```bash
maisternia config set-repository /path/to/agentnyk-maisternia
```

This setting is optional. `maisternia config clear-repository` restores
automatic source-checkout and embedded-catalog discovery.
