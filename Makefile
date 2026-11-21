BINARY := fm

.PHONY: all build binary test lint test-cli test-cli-live test-all test-ci cover vet fmt clean help

all: build ## Build the binary (default)

build: test binary ## Run unit tests and build the binary

binary: ## Build the binary
	go build -o $(BINARY) .

test: ## Run unit tests
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

test-cli: binary ## Run scrut CLI integration tests
	bash tests/run-cli-tests

test-cli-live: binary ## Run opt-in live CLI integration tests (requires FM_CREDENTIAL_COMMAND and FM_LIVE_TESTS=1)
	scrut test tests/live.md

test-all: test test-cli ## Run all tests (unit + CLI)

test-ci: vet fmt lint test-all ## Run CI test suite

cover: ## Run unit tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet: ## Run go vet
	go vet ./...

fmt: ## Check formatting (exits non-zero on errors or unformatted files)
	@unformatted=$$(gofmt -l .) || exit $$?; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
