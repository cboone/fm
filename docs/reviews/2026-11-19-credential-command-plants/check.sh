#!/usr/bin/env bash
set -euo pipefail

function main() {
  cd cmd
  ../cmd.test -test.v -test.run '^TestCredentialCommandConfig$'
}

main "$@"
