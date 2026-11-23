# Enable the Go format check in CI

The workflow pinned at `844bf31d4ec26df9d0aa531e79d8c3fccf9f0638` runs
`make fmt` when `run-format-check` is true. It reads Go 1.24.7 from
`go.mod`. The base commit `e63ea87` passed this command with Go 1.24.7.

## Format check plant

In a scratch copy of the Go sources and Makefile, changed `func main() {`
to `func main(){` in `main.go`. Ran the workflow's exact command, `make fmt`,
with the Go 1.24.7 toolchain first on `PATH`. The baseline exited 0.
The planted run printed `main.go` and Make reported recipe error 1.

plant unformatted-main: named assertion reached; red exit 2; restored;
green exit 0.

## Nit-round format check plants

Run from the worktree root. The block uses the Go sources from `HEAD` and the
working-tree Makefile. The malformed fixture must produce a gofmt syntax
diagnostic. The spacing plant must print `main.go`.

```bash
bash -euo pipefail <<'BASH'
function cleanup() {
  local plant_status=$?
  local cleanup_status=0
  rm -r -- "${scratch_dir}" || cleanup_status=$?
  if (( plant_status != 0 )); then
    exit "${plant_status}"
  fi
  exit "${cleanup_status}"
}

go_root=$(GOTOOLCHAIN=go1.24.7 go env GOROOT)
export PATH="${go_root}/bin:${PATH}"
scratch_dir=$(mktemp -d "${PWD}/.format-plant.XXXXXX")
trap cleanup EXIT
git archive HEAD | tar -x -C "${scratch_dir}"
cp Makefile "${scratch_dir}/Makefile"
cd "${scratch_dir}"
make fmt

mkdir testdata
printf 'package fixture\n\nfunc broken(\n' > testdata/format-plant.go
fmt_status=0
fmt_output=$(make fmt 2>&1) || fmt_status=$?
printf '%s\n' "${fmt_output}"
[[ "${fmt_status}" -eq 2 ]]
[[ "${fmt_output}" == *testdata/format-plant.go:* ]]
rm testdata/format-plant.go
make fmt

cp main.go main.go.original
sed 's/^func main() {$/func main(){/' main.go.original > main.go
fmt_status=0
fmt_output=$(make fmt 2>&1) || fmt_status=$?
printf '%s\n' "${fmt_output}"
[[ "${fmt_status}" -eq 2 ]]
[[ "${fmt_output}" == $'main.go\n'* ]]
mv main.go.original main.go
make fmt
BASH
```

plant malformed-testdata: named assertion reached; red exit 2; restored;
green exit 0.

plant unformatted-main: named assertion reached; red exit 2; restored;
green exit 0.

With the old `fmt` recipe restored, the malformed fixture printed the syntax
diagnostic and `make fmt` exited 0. The status assertion failed and the command
block exited 1. The trap removed the scratch directory. After restoring the
fixed recipe, the command block exited 0.

## Help recipe plant

Run from the worktree root. `Makefile/missing.mk` cannot be read because
`Makefile` is a regular file. The normal help listing must succeed.

```bash
bash -euo pipefail <<'BASH'
make help
help_status=0
help_output=$(make help MAKEFILE_LIST=Makefile/missing.mk 2>&1) || help_status=$?
printf '%s\n' "${help_output}"
[[ "${help_status}" -eq 2 ]]
[[ "${help_output}" == *Makefile/missing.mk* ]]
BASH
```

The old `help` recipe printed the read error and `make help` exited 0. The
command block exited 1 on the status assertion. With the fixed recipe, the
unreadable input made `make help` exit 2 and the command block exited 0.
The normal help output matched the old recipe's output.

## Failure propagation audit

The first two Codex review rounds each found one Important instance. The second
round stopped the review loop and prompted this class audit.

- Nit-round format check command block, changed. Strict mode stops on failed
  assertions. The trap removes the scratch directory and preserves the status.
- Help recipe command block, changed (new). Strict mode stops on a failed exit
  status or diagnostic assertion.
- `cover` recipe, unchanged. Make stops on a failed `go test` recipe line.
- `fmt` recipe, unchanged. The recipe propagates a failed `gofmt` and explicitly
  fails on unformatted files.
- `help` recipe, changed. A single `awk` command reports input read failures to
  Make.

The `all`, `build`, `test-all`, and `test-ci` targets only list prerequisites.
Make observes failures from the prerequisite recipes directly.
