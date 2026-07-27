.PHONY: build doctor format test verify

build:
	go build ./cmd/agentctl

doctor:
	go run ./cmd/agentctl doctor

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
	go build ./cmd/agentctl
