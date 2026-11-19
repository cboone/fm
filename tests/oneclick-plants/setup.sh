#!/usr/bin/env bash
set -euo pipefail
function main() {
  go test -c -o plant-unsubscribe.test ./internal/unsubscribe
  go test -c -o plant-client.test ./internal/client
  go test -c -o plant-cmd.test ./cmd
  go build -o fm .
}
main "$@"
