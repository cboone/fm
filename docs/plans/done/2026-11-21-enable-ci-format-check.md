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
