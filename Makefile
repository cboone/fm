BINARY := jm

.PHONY: all build test test-cli test-cli-live test-all cover vet fmt clean help

all: build ## Build the binary (default)

build: ## Build the binary
	go build -o $(BINARY) .

test: ## Run unit tests
	go test ./...

test-cli: build ## Run scrut CLI integration tests
	scrut test tests/errors.md tests/flags.md tests/arguments.md tests/help.md

test-cli-live: build ## Run opt-in live CLI integration tests (requires JMAP_TOKEN and JMAP_LIVE_TESTS=1)
	scrut test tests/live.md

test-all: test test-cli ## Run all tests (unit + CLI)

cover: ## Run unit tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet: ## Run go vet
	go vet ./...

fmt: ## Check formatting (exits non-zero if files need formatting)
	@test -z "$$(gofmt -l .)" || { gofmt -l . && exit 1; }

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
