SHELL := /usr/bin/env bash -e
.DEFAULT_GOAL := build

.PHONY: default
default: build ;

.PHONY: build
build: require-go format ## build the common-inventory binary
	go mod tidy
	go build -o ./bin/common-inventory main.go

.PHONY: test
test: WHAT ?= ./...
test: build require-go
	go test -v $(WHAT)

.PHONY: format
format: ## format go code in the project
	gofmt -w .

.PHONY: clean
clean: ## clean out the binaries
	@rm -rf ./bin/*

.PHONY: require-%
require-%:
	@if ! command -v $* 1> /dev/null 2>&1; then echo "$* not found in \$$PATH"; exit 1; fi

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
