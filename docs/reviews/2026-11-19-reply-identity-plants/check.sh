#!/usr/bin/env bash
set -euo pipefail

function main() {
  ./client.test -test.v -test.run '^TestCreateDraft_(ReplyIdentity|NonReplyNoIdentityLookup)'
  ./cmd.test -test.v -test.run '^TestDraftWarningsOutput$'
}

main "$@"
