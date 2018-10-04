# Packages to build
PKGS := $(shell go list -f '{{if .GoFiles}}{{.ImportPath}}{{end}}' ./...)
# Tests to run
TESTS := $(shell go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)

all: build

build: vendor
	@echo "--- build ---"
	@go build -v $(PKGS)
	@go vet $(PKGS)

test:
	@echo "--- test ---"
	@go test $(TESTS)

lint: $(GOLINT)
	@echo "--- lint ---"
	@bin/gometalinter.sh

.PHONY: all build test lint vendor
