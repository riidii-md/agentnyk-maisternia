# Contributing

## Workflow

1. Define observable behavior.
2. Add or update tests first.
3. Confirm the tests fail for the expected reason.
4. Implement the smallest safe change.
5. Run formatting, vet, tests, race detection, coverage, and build.
6. Update documentation and sample configuration.
7. Review security boundaries before opening a PR.

## Commands

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go build ./cmd/maisternia
```

## Pull Requests

PR descriptions should include:

- behavior changed;
- tests added;
- commands run;
- security implications;
- migration impact;
- remaining limitations.
