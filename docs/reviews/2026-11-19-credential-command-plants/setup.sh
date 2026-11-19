#!/usr/bin/env bash
set -euo pipefail

function main() {
  go test -trimpath -c -o cmd.test ./cmd
}

main "$@"
