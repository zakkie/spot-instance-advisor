NAME := spot-instance-advisor
VERSION := 0.0.1
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X main.revision=$(REVISION)

export GO111MODULE=on

## Build binary
.PHONY: all
all: bin/$(NAME)

## Format code
.PHONY: fmt
fmt:
	go mod tidy
	go fmt ./...
	go vet ./...

# Build binaries ex. make bin/spot-instance-advisor
bin/%: cmd/%/main.go
	go build -ldflags "$(LDFLAGS)" -o $@ $<

## Build
.PHONY: build
build: bin/$(NAME)

## Run
.PHONY: run
run: build
	./bin/$(NAME)

## Run tests
.PHONY: test
test: deps
	go test ./...

## Clean
.PHONY: clean
clean:
	rm -rf ./bin/

## Install dependencies
.PHONY: deps
deps:
	echo "NOOP"
	# go get -v

## Setup development environment
.PHONY: devel-deps
devel-deps: deps
	go install \
	 github.com/Songmu/make2help/cmd/make2help@HEAD

## Help
.PHONY: help
help:
	@make2help $(MAKEFILE_LIST)
