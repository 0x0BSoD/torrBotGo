
PROJECT_NAME := "torrBotGo"
PKG := "github.com/0x0BSoD/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all mod lint vet build clean

all: build

build: mod
	@go build -x -o build/main $(PKG)

mod:
	@go mod download

lint:
	@golint -set_exit_status ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

clean:
	@rm -rf ./build
