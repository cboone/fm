#!/usr/bin/env bash
# Compile the synthetic draft error tests in the runner's private input copy.
set -euo pipefail

function main() {
  go test -c -o draft-error-tests ./internal/client
}

main "$@"
