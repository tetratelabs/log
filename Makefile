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

LINTER := bin/golangci-lint
$(LINTER):
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.10.2

lint: $(LINTER) ./bin/.golangci.yml
	@echo "--- lint ---"
	@$(LINTER) run --config ./bin/.golangci.yml

.PHONY: all build test lint vendor
