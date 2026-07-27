# Repository Instructions

## Scope

This repository owns the provider-neutral workflow definitions and the
`agentctl` installer.

## Development

- Use the Go standard library unless a dependency removes substantial
  complexity and has been reviewed.
- Write tests before implementation changes.
- Keep manifest and path handling defensive.
- Never add credentials, tokens, transcripts, runtime databases, or real user
  configuration values to fixtures.
- Use `apply_patch` for manual edits when working through Codex.

## Verification

Run:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go build ./cmd/agentctl
```

Also run:

```bash
go run ./cmd/agentctl doctor
go run ./cmd/agentctl render --target all --output ./build/rendered
```

## Safety

- `apply` must remain opt-in.
- Never weaken path, symlink, conflict, drift, backup, or confirmation checks to
  make a test pass.
- Provider home directories contain mixed runtime and declarative state. Do not
  introduce whole-directory synchronization.
