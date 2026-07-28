.PHONY: build clean doctor format install release-check release-snapshot test uninstall verify

GO ?= go
BINARY := agentctl
BUILDINFO := github.com/kagi-labs/agentctl/internal/buildinfo
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell git show -s --format=%cI HEAD 2>/dev/null || echo unknown)
LDFLAGS := -s -w \
	-X $(BUILDINFO).Version=$(VERSION) \
	-X $(BUILDINFO).Commit=$(COMMIT) \
	-X $(BUILDINFO).Date=$(BUILD_DATE)
INSTALL_BIN := $(shell $(GO) env GOBIN)
ifeq ($(INSTALL_BIN),)
INSTALL_BIN := $(shell $(GO) env GOPATH)/bin
endif

build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/agentctl

install:
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" ./cmd/agentctl
	@echo "installed $(BINARY) to $(INSTALL_BIN)/$(BINARY)"

uninstall:
	rm -f "$(INSTALL_BIN)/$(BINARY)"

clean:
	rm -f "$(BINARY)"
	rm -rf dist

doctor:
	$(GO) run ./cmd/agentctl doctor

format:
	gofmt -w cmd internal

test:
	$(GO) test ./...

release-check:
	goreleaser check

release-snapshot:
	goreleaser release --snapshot --clean

verify:
	gofmt -w cmd internal
	$(GO) vet ./...
	$(GO) test ./...
	$(GO) test -race ./...
	$(GO) test -coverprofile=coverage.out ./internal/...
	$(GO) build ./cmd/agentctl
