# Environment Requirements

Presets may depend on tools that live outside a coding-agent configuration
directory. A workflow can need a terminal multiplexer, workspace manager,
agent multiplexer, or host plugin before its installed
commands are useful.

`maisternia` represents those dependencies as environment packs. Dedicated
environment-only presets keep machine setup separate from provider files and
from workflow presets.

## Model

Version 1 environment packs live under:

```text
config/environments/<pack-id>.json
```

A preset references packs through `environment_packs`. Each pack declares:

- required or optional capabilities;
- dependencies between requirements;
- a bounded command-presence check;
- supported platforms;
- typed installation choices and pinned versions or refs where applicable.

The provider-neutral `terminal-orchestration` preset references the same-named
environment pack. It installs or verifies:

- Zellij;
- Tatami;
- Herdr;
- [Mdmaid](https://github.com/OleksandrBesan/mdmaid) `0.1.14` through npm;
- [mdmaid.desk](https://github.com/riidii-md/mdmaid.desk) `0.1.0` through npm;
- [Herdr Automatic Rename](https://github.com/qu8n/herdr-automatic-rename);
- [Herdr Bar](https://github.com/jeffarese/herdr-bar).

The two plugin sources are pinned to immutable Git commit SHAs. Installation
does not invent their configuration: Herdr Bar still needs a chosen key binding,
and Automatic Rename's optional shell hook remains an explicit shell
configuration choice.

## Inspect And Plan

List and inspect the library:

```bash
maisternia environment list
maisternia environment show terminal-orchestration
maisternia environment validate all
```

Build a plan for the current machine:

```bash
maisternia environment plan terminal-orchestration
maisternia preset plan terminal-orchestration
```

Planning only calls the operating system's executable lookup. It does not run
the detected tools, execute an installer, fetch remote content, or write files.
Each requirement is reported as:

- `satisfied`: its declared command is on `PATH`;
- `missing`: it is absent and at least one typed installer supports the host;
- `blocked`: a declared dependency is not satisfied;
- `inspect-required`: plugin presence will be checked through Herdr only after
  installation is explicitly confirmed;
- `unsupported`: no installer is declared for the host platform.

Environment results do not silently expand a configuration-preset apply.
Provider configuration and machine setup remain separately reviewed
operations. Applying an environment-only preset opens or prints its environment
plan instead of asking for a provider or project scope.

## Install

After reviewing the plan, install the missing requirements explicitly:

```bash
maisternia environment plan terminal-orchestration
maisternia preset apply --yes terminal-orchestration
# Equivalent direct pack command:
maisternia environment install --yes terminal-orchestration
```

Without `--yes`, `install` prints the exact plan and exits without running a
command. With confirmation, maisternia:

1. checks each requirement and its declared dependencies;
2. selects the first platform-compatible typed installer whose executable is
   present;
3. inspects Herdr's structured plugin registry for plugin requirements;
4. executes only the displayed argument arrays;
5. verifies the command or plugin is present afterward.

Already-satisfied requirements are skipped. Failure stops the remaining work
and reports the requirement and command; package-manager changes already made
before a failure are not transactionally rolled back. In Admin, `i` opens the
same environment review and `y` explicitly confirms installation. A
configuration-preset apply never installs packages implicitly.

Environment requirements do not currently participate in preset ownership or
automatic reconciliation. Maisternia cannot safely infer that a Homebrew, npm,
Cargo, Go, or host-plugin installation is unused merely because one preset no
longer references it. Removing or changing an environment requirement therefore
does not upgrade or uninstall an already-present host tool. Use the owning
package manager or plugin host explicitly. Managed provider files are different:
`preset apply` and `preset uninstall` reconcile their recorded ownership as
documented in [Preset library](PRESETS.md).

## Typed Installer Choices

Environment definitions cannot contain arbitrary shell. The schema currently
allows only these typed choices:

- Homebrew package, with an optional tap;
- pinned `go install` module;
- pinned `cargo binstall` crate;
- pinned global npm package;
- pinned Herdr plugin source;
- an HTTPS link to official manual instructions.

Generated commands are represented as argument arrays, never as shell source.
All manifest text and command components are validated, unknown JSON fields are
rejected, files are size-limited, and symlinked library paths are rejected.
Installing a Herdr plugin can run build steps declared by that pinned external
repository, which is why plugin execution happens only after the exact source
and revision are displayed and `--yes` is supplied.

## Plugin Boundary

Herdr has a documented plugin installation interface, so the model can
represent those operations without guessing a filesystem layout.
Tatami does not currently expose an equivalent stable plugin-install contract.
Tatami layouts or other declarative files should remain normal managed
resources; executable Tatami plugins should be added only after Tatami defines
the host API, target identity, versioning, and verification behavior.

## Safety Boundary

The implemented flow is:

```text
detect -> plan -> choose exact installer -> confirm -> execute -> verify
```

Execution remains opt-in, uses typed argument arrays, displays exact commands,
and verifies each requirement afterward. Remote pipe-to-shell scripts are not
an environment-pack extension mechanism. A future environment install-state
layer may add drift and uninstall support, but current truth comes from command
presence and the Herdr plugin registry.
