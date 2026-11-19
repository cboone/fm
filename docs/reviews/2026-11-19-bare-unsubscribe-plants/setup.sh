#!/usr/bin/env bash
set -euo pipefail

function main() {
  go test -c -o parser.test ./internal/unsubscribe
  go test -c -o client.test ./internal/client
  go test -c -o command.test ./cmd
}

main "$@"
