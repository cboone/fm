#!/usr/bin/env bash
set -euo pipefail

function main() {
  go test -c -o flag-color.test ./internal/client
  go test -c -o flag-color-cli.test ./cmd
  go test -c -o flag-color-text.test ./internal/output
}

main "$@"
