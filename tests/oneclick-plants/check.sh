#!/usr/bin/env bash
set -euo pipefail
function main() {
  bash tests/test-cli-containment
  ./plant-unsubscribe.test -test.v -test.run '^Test(Target|PublicDial|Post|Authenticated|ProductionDestinationGuard)'
  ./plant-client.test -test.v -test.run '^Test(NewestSenderEmail|SenderLookup)'
  ./plant-cmd.test -test.v -test.run '^TestOneClick'
  bash tests/run-cli-tests tests/oneclick.md
  echo FM88_CHECK_COMPLETE
}
main "$@"
