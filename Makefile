BINDIR	:= $(CURDIR)/bin
BINNAME	?= goweathertui
INSTALL_PATH ?= /usr/local/bin

SHELL      = /usr/bin/env bash

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

# go option
PKG        := ./...
TAGS       :=
TESTS      := .
TESTFLAGS  :=
LDFLAGS    := -w -s
GOFLAGS    :=

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

.PHONY: all
all: build

.PHONY: build
build: lint $(BINDIR)/$(BINNAME)

# Rebuild the binary if any of these files change
SRC := $(shell find . -type f -name '*.go' -print) go.mod go.sum

$(BINDIR)/$(BINNAME): $(SRC)
	GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)'/$(BINNAME) ./cmd/goweathertui

.PHONY: install
install: build
	@install "$(BINDIR)/$(BINNAME)" "$(INSTALL_PATH)/$(BINNAME)"

.PHONY: test
test: lint build
test: TESTFLAGS += -race -v
test: test-unit

.PHONY: test-unit
test-unit:
	@echo "==> Running unit tests <=="
	GO111MODULE=on go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: coverage
coverage:
	@echo "==> Running coverage tests <=="
	GO111MODULE=on go test $(GOFLAGS) -run $(TESTS) $(PKG) -coverprofile=coverage.out --covermode=atomic

.PHONY: lint
lint: fmt vet

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm $(BINDIR)/$(BINNAME)
	rmdir $(BINDIR)
