# Makefile for torrbot (portable Linux binary from NixOS dev box)
# Usage:
#   make            # build
#   make run         # run locally
#   make test        # run tests
#   make fmt vet lint
#   make tidy
#   make clean

APP       := torrbot
PKG       := ./...
BIN_DIR   := bin
OUT       := $(BIN_DIR)/$(APP)

GOOS      ?= linux
GOARCH    ?= amd64
CGO_ENABLED ?= 0

# Version info (optional, but handy)
GIT_SHA   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "nogit")
GIT_TAG   := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_DESCR := $(shell git describe --tags 2>/dev/null || echo "v0.0.0-0-nogit")
BUILD_DATE:= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Flags
GOFLAGS   ?=
LDFLAGS   := -s -w \
	-X main.Version=$(GIT_TAG) \
	-X main.GitCommit=$(GIT_SHA) \
	-X main.BuildDate=$(BUILD_DATE)

.PHONY: all build run test fmt vet lint tidy clean tools info

all: build

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(GOFLAGS) -trimpath -ldflags='$(LDFLAGS)' -o $(OUT)$(if $(filter windows,$(GOOS)),.exe) .

run: build
	./$(OUT)

test:
	go test $(PKG)

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

# Installs tools to GOPATH/bin (or GOBIN) if missing
tools:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}

lint: tools
	@PATH="$$(go env GOPATH)/bin:$$PATH" golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

info:
	@echo "APP=$(APP)"
	@echo "OUT=$(OUT)"
	@echo "GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED)"
	@echo "GIT_SHA=$(GIT_SHA)"
	@echo "BUILD_DATE=$(BUILD_DATE)"
