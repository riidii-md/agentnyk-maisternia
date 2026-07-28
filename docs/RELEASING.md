# Releasing agentctl

## Release Outputs

Pushing a `v*` tag runs `.github/workflows/release.yml`. GoReleaser creates:

- macOS binaries for `amd64` and `arm64`;
- Linux binaries for `amd64` and `arm64`;
- Windows binaries for `amd64` and `arm64`;
- `.tar.gz` archives for macOS and Linux;
- `.zip` archives for Windows;
- `checksums.txt`;
- a private GitHub release;
- a Homebrew cask when tap publishing is configured.

Release binaries embed the version, commit, and commit date.

## One-Time Homebrew Setup

Homebrew uses a separate repository:

```text
kagi-labs/homebrew-tap
```

Create it as a private repository with a `main` branch and a `Casks/`
directory.

Create a fine-grained GitHub token that has repository contents write access to
`kagi-labs/homebrew-tap`. Add it to the `kagi-labs/agentctl` Actions secrets as:

```text
HOMEBREW_TAP_TOKEN
```

The default GitHub Actions token cannot publish to another repository.

When the secret is absent, GoReleaser still creates GitHub release archives and
checksums but skips tap publication. This prevents Homebrew setup from breaking
the primary release.

Because release assets are private, the generated cask uses an authenticated
download strategy. Users provide `HOMEBREW_GITHUB_API_TOKEN` when installing or
upgrading.

## Validate Locally

Install GoReleaser v2, then run:

```bash
make release-check
make release-snapshot
```

Inspect `dist/` and verify a generated binary:

```bash
./dist/agentctl_darwin_arm64_v8.0/agentctl --version
```

The exact internal build directory may vary between GoReleaser versions. The
release archives use stable names declared in `.goreleaser.yml`.

## Create A Release

Start from a clean, verified `main` branch:

```bash
make verify
git status --short
```

Create and push an annotated semantic-version tag:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

The tag push is the release authority boundary. The Makefile intentionally does
not create or push tags.

## Post-Release Checks

1. Confirm all six platform archives and `checksums.txt` exist.
2. Download an archive and run `agentctl --version`.
3. Confirm `Casks/agentctl.rb` was updated in the tap.
4. Test a Homebrew install using a GitHub token with read access.
5. Run `agentctl doctor` from the installed binary.
6. Confirm no configuration was applied during installation.
