.PHONY: build doctor format test verify

build:
	go build ./cmd/agent-config

doctor:
	go run ./cmd/agent-config doctor

format:
	gofmt -w cmd internal

test:
	go test ./...

verify:
	gofmt -w cmd internal
	go vet ./...
	go test ./...
	go test -race ./...
	go test -coverprofile=coverage.out ./...
	go build ./cmd/agent-config
