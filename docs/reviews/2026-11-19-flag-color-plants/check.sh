#!/usr/bin/env bash
set -euo pipefail

function main() {
  ./flag-color.test -test.v -test.run '^TestFlagColor'
  ./flag-color-cli.test -test.v -test.run '^TestFlagColorCLI'
  ./flag-color-text.test -test.v -test.run '^TestFlagColorTextWriteError$'
}

main "$@"
