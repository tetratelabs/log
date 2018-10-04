# Location of the 'dep' binary, to install it if missing
DEP := $(shell which dep)

# Packages to build
PKGS := $(shell go list -f '{{if .GoFiles}}{{.ImportPath}}{{end}}' ./...)
# Tests to run
TESTS := $(shell go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)


all: build

build: vendor
	@echo "--- build ---"
	@go build -v -i $(PKGS)
	@go vet $(PKGS)

test:
	@echo "--- test ---"
	@go test $(TESTS)

lint: $(GOLINT)
	@echo "--- lint ---"
	@bin/gometalinter.sh

$(DEP):
	@go get -v github.com/golang/dep/cmd/dep

vendor: $(DEP)
	@echo "--- update dependencies ---"
	@dep ensure -v

.PHONY: all build test lint vendor
