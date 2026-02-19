BINARY := fm

.PHONY: all build binary test test-cli test-cli-live test-all test-ci cover vet fmt clean help

all: build ## Build the binary (default)

build: test binary ## Run unit tests and build the binary

binary: ## Build the binary
	go build -o $(BINARY) .

test: ## Run unit tests
	go test ./...

test-cli: binary ## Run scrut CLI integration tests
	scrut test tests/errors.md tests/flags.md tests/arguments.md tests/help.md tests/sieve.md

test-cli-live: binary ## Run opt-in live CLI integration tests (requires FM_TOKEN and FM_LIVE_TESTS=1)
	scrut test tests/live.md

test-all: test test-cli ## Run all tests (unit + CLI)

test-ci: vet fmt test-all ## Run CI test suite

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
