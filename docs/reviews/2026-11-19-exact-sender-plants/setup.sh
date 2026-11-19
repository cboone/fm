#!/usr/bin/env bash
set -euo pipefail

go test -trimpath -c -o client.test ./internal/client
go test -trimpath -c -o cmd.test ./cmd
