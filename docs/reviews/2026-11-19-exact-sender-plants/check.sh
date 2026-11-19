#!/usr/bin/env bash
set -euo pipefail

./client.test -test.v -test.run '^TestExactSender'
./cmd.test -test.v -test.run '^TestExactSenderCLI'
