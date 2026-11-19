#!/usr/bin/env bash
set -euo pipefail

function main() {
  local judge_exit
  ./parser.test -test.run '^Test(Parse|EqualFoldASCII|EqualSenderAddressASCII|ProtocolMarkerASCII|Authenticated)' -test.v
  ./client.test -test.run 'ASCII|^TestConvertDetail' -test.v
  ./command.test -test.run '^TestOneClick(FromASCII|Senders)$' -test.v
  if FM_PARSER_TEST_WRONG_URL=1 ./parser.test \
    -test.run '^TestParseBareAndOneClick/bare_HTTPS$' -test.v >judge.log 2>&1; then
    judge_exit=0
  else
    judge_exit=$?
  fi
  if [[ "${judge_exit}" != 1 ]] || ! rg -q 'wrong.invalid/judge' judge.log; then
    cat judge.log
    echo 'URL_JUDGE_MISSED'
    return 1
  fi
  echo 'URL_JUDGE_REJECTED'
}

main "$@"
