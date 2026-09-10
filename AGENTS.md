# FM

## Overview

A safe, read-oriented CLI for Fastmail email via JMAP.

`fm` is machine-first by design: JSON output by default, structured errors, and strict guardrails
around anything destructive. It is built to be the execution layer an agent drives for inbox
analysis and triage, so its output contract matters as much as its behavior. See the "Agent
Runbook (Read First)" section of `README.md` and `skills/review-email/references/runbook.md`
before changing command output.

The repo is also a Claude Code plugin: `.claude-plugin/plugin.json` publishes `skills/` so the
runbook ships with the tool.

## Structure

```text
main.go                  entry point
cmd/                     cobra commands, one file per command
internal/client/         Fastmail session and transport
internal/jmap/           JMAP protocol types and requests
internal/output/         JSON and human output formatting
internal/types/          shared domain types
internal/unsubscribe/    unsubscribe link handling
skills/review-email/     bundled skill, including the agent runbook
tests/                   scrut CLI tests (arguments, errors, flags, help, live, sieve)
docs/plans/              plan files
docs/reviews/            review documents
```

Commands in `cmd/` cover reading and triage (`list`, `read`, `search`, `mailboxes`, `archive`,
`move`, `mark-read`, `flag`, `draft`) plus the full Sieve filter surface (`sieve_create`,
`sieve_show`, `sieve_list`, `sieve_validate`, `sieve_activate`, `sieve_deactivate`,
`sieve_delete`). `dryrun.go` and `filters.go` carry the guardrails; treat changes there as
behavior changes, not refactors.

## Development

Run `make help` for the full target list. The ones that matter:

| Target               | What it does                                              |
| -------------------- | --------------------------------------------------------- |
| `make build`         | Run unit tests, then build the binary                     |
| `make test`          | Unit tests only                                           |
| `make test-cli`      | Build, then run the scrut CLI tests in `tests/`           |
| `make test-all`      | Unit plus CLI tests                                       |
| `make test-ci`       | What CI runs: `vet`, `fmt`, `lint`, `test-all`            |
| `make lint`          | golangci-lint                                             |
| `make cover`         | Unit tests with coverage                                  |

CLI behavior is covered by scrut snapshot tests in `tests/*.md`. When you change output, flags, or
error text, update the matching snapshot in the same change; `docs_drift_test.go` also guards
against docs falling behind the command surface.

`make test-cli-live` runs the opt-in tests that talk to a real account. It requires
`FM_CREDENTIAL_COMMAND` and `FM_LIVE_TESTS=1`. Never hardcode a token to run these.

CI is `.github/workflows/ci.yml`, with `gitleaks.yml` and `release.yml` alongside.
