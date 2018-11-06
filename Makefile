# Packages to build
PKGS := $(shell go list -f '{{if .GoFiles}}{{.ImportPath}}{{end}}' ./...)
# Tests to run
TESTS := $(shell go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)

# Enable go modules when building, even if the repo is copied into a GOPATH.
export GO111MODULE=on

build:
	@echo "--- build ---"
	@go build -v $(PKGS)
	@go vet $(PKGS)

test:
	@echo "--- test ---"
	@go test $(TESTS)

LINTER := bin/golangci-lint
$(LINTER):
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.11.2

lint: $(LINTER) ./bin/.golangci.yml
	@echo "--- lint ---"
	@$(LINTER) run --config ./bin/.golangci.yml

.PHONY: build test lint
