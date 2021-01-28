# Packages to build
PKGS := ./...
TEST_OPTS ?=

build:
	@echo "--- build ---"
	go build -v $(PKGS)

test:
	@echo "--- test ---"
	go test $(TEST_OPTS) $(PKGS)

LINTER := bin/golangci-lint
$(LINTER):
	wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v1.23.6

lint: $(LINTER) golangci.yml
	@echo "--- lint ---"
	$(LINTER) run --config golangci.yml

.PHONY: build test lint
