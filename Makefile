GO ?= go
GO_ENV := GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache
BIN_DIR := bin
BIN_NAME := rainbow-backend

.PHONY: run build test

run:
	$(GO_ENV) $(GO) run ./cmd/server

build:
	mkdir -p $(BIN_DIR)
	$(GO_ENV) $(GO) build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/server

test:
	$(GO_ENV) $(GO) test ./...
