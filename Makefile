VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")

LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: test build build-no-modules modules

all: test build

test:
	go test ./...

build:
	go build -o bin/yobot -v -ldflags "$(LDFLAGS)" cmd/yobot/main.go

build-no-modules:
	CGO_ENABLED=0 go build -o bin/yobot -v -ldflags "$(LDFLAGS)" cmd/yobot/main.go

modules: SHELL:=/bin/bash
modules:
	@echo "Building modules"
	@rm -f ./modules/*.so
	@find modules/* -type d -print0 | while IFS= read -r -d $$'\0' m; do \
		cd "$$m"; \
		echo "Building $$(basename $$m).so"; \
		go build -buildmode=plugin -o "../$$(basename $$m).so" .; \
		cd ../..; \
	done
