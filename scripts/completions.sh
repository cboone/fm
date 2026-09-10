#!/usr/bin/env bash
# completions.sh -- Generate shell completion scripts for fm
# Called by goreleaser before builds to bundle completions in archives.

set -euo pipefail

readonly BINARY="fm"
readonly OUTPUT_DIR="completions"
readonly SHELLS=(bash zsh fish)

function main() {
  rm -rf "${OUTPUT_DIR}"
  mkdir "${OUTPUT_DIR}"

  go build -o "${BINARY}" .

  local shell
  for shell in "${SHELLS[@]}"; do
    "./${BINARY}" completion "${shell}" > "${OUTPUT_DIR}/${BINARY}.${shell}"
  done

  rm "${BINARY}"
}

main "${@}"
