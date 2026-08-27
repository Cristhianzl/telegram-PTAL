# PullAlerts - your GitHub pull requests on Telegram
BINARY  := pullalerts
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0-dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
GO      ?= go

.DEFAULT_GOAL := help

## help: show these commands
help:
	@echo "PullAlerts $(VERSION)"
	@echo
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | awk -F': ' '{printf "  \033[1m%-12s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Starting from scratch:  make setup && make install"

## build: compile the binary for this system
build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/pullalerts

## setup: connect Telegram and discover your chat
setup: build
	./$(BINARY) setup

## doctor: diagnose token, chat, connectivity and service
doctor: build
	./$(BINARY) doctor

## once: run one cycle and show what it found
once: build
	./$(BINARY) once

## panel: send the current state to Telegram
panel: build
	./$(BINARY) panel

## run: stay in the foreground (Ctrl+C to stop)
run: build
	./$(BINARY) run

## install: register to start with the computer
install: build
	./$(BINARY) install

## uninstall: remove the registration
uninstall:
	./$(BINARY) uninstall

## status: show the service and the last sync
status: build
	./$(BINARY) status

## test: run all tests
test:
	$(GO) test ./...

## check: vet plus tests with the race detector
check:
	$(GO) vet ./...
	$(GO) test -race ./...

## ci: what continuous integration runs
ci: check
	$(GO) build ./...

## dist: build for Linux, macOS and Windows
dist:
	@mkdir -p dist
	GOOS=linux   GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64   ./cmd/pullalerts
	GOOS=linux   GOARCH=arm64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64   ./cmd/pullalerts
	GOOS=darwin  GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64  ./cmd/pullalerts
	GOOS=darwin  GOARCH=arm64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64  ./cmd/pullalerts
	GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS) -H=windowsgui" -o dist/$(BINARY)-windows-amd64.exe ./cmd/pullalerts
	@ls -lh dist/

## clean: remove binaries and local state
clean:
	rm -rf dist $(BINARY) $(BINARY).exe state.json state.json.tmp

.PHONY: help build setup doctor once panel run install uninstall status test check ci dist clean
